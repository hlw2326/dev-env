package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"devenv/internal/domain/deployment"
	"devenv/internal/domain/project"
	"devenv/internal/ports"
	appErrors "devenv/internal/shared/errors"
	"github.com/google/uuid"
)

const DEPLOYMENT_TIMEOUT = 30 * time.Minute
const DEPLOYMENT_PLAN_TTL = 15 * time.Minute

type deploymentPlan struct {
	plan    deployment.Plan
	config  deployment.Config
	root    string
	expires time.Time
}
type DeploymentService struct {
	projects   ports.ProjectRepository
	repository ports.DeploymentRepository
	settings   ports.SettingRepository
	providers  map[string]ports.DeploymentProvider
	mu         sync.Mutex
	plans      map[int64]deploymentPlan
	tasks      map[int64]deployment.Task
	cancel     map[int64]context.CancelFunc
	targets    map[string]bool
	closed     bool
	group      sync.WaitGroup
}

func NewDeploymentService(projects ports.ProjectRepository, repository ports.DeploymentRepository, providers map[string]ports.DeploymentProvider, settings ...ports.SettingRepository) *DeploymentService {
	var settingRepository ports.SettingRepository
	if len(settings) > 0 {
		settingRepository = settings[0]
	}
	return &DeploymentService{projects: projects, repository: repository, settings: settingRepository, providers: providers, plans: make(map[int64]deploymentPlan), tasks: make(map[int64]deployment.Task), cancel: make(map[int64]context.CancelFunc), targets: make(map[string]bool)}
}

func (s *DeploymentService) project(ctx context.Context, id int64) (*project.Project, error) {
	item, err := s.projects.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, appErrors.New(appErrors.NOT_FOUND, "项目不存在")
	}
	if deployment.ProviderFor(item.Kind) == "" {
		return nil, appErrors.New(appErrors.INVALID_ARGUMENT, "此项目不支持同步或微信上传")
	}
	return item, nil
}
func (s *DeploymentService) Config(ctx context.Context, id int64) (deployment.Config, error) {
	item, err := s.project(ctx, id)
	if err != nil {
		return deployment.Config{}, err
	}
	stored, err := s.repository.Load(ctx, id)
	if err != nil {
		return deployment.Config{}, appErrors.Wrap(err, appErrors.DATABASE_ERROR, "读取发布配置失败")
	}
	config := deployment.Default(id, item.Kind)
	if stored != nil {
		config = *stored
	}
	deployment.NormalizeAppIDs(&config)
	if config.Provider == deployment.PROVIDER_WECHAT {
		config.Exclude = ""
	}
	if item.Kind == project.KIND_MINIPROGRAM && s.settings != nil {
		global, settingsErr := s.settings.Load(ctx)
		if settingsErr != nil {
			return deployment.Config{}, appErrors.Wrap(settingsErr, appErrors.DATABASE_ERROR, "读取微信开发者工具设置失败")
		}
		if global.DeveloperToolsCLI == "" && config.DeveloperToolsCLI != "" {
			global.DeveloperToolsCLI = config.DeveloperToolsCLI
			_ = s.settings.Save(ctx, global)
		}
		config.DeveloperToolsCLI = global.DeveloperToolsCLI
	}
	if stored == nil {
		return config, nil
	}
	if item.Kind == project.KIND_MINIPROGRAM && (strings.TrimSpace(config.Source) == "" || config.Source == "." || config.Source == "dist/build/mp-weixin") {
		compiled := filepath.Join(item.Root, filepath.FromSlash(deployment.DEFAULT_WECHAT_SOURCE))
		if info, statErr := os.Stat(compiled); statErr == nil && info.IsDir() {
			config.Source = deployment.DEFAULT_WECHAT_SOURCE
		}
	}
	return config, nil
}
func (s *DeploymentService) Save(ctx context.Context, config deployment.Config) error {
	item, err := s.project(ctx, config.ProjectID)
	if err != nil {
		return err
	}
	deployment.NormalizeAppIDs(&config)
	if config.Provider == deployment.PROVIDER_WECHAT {
		config.Exclude = ""
		if s.settings != nil {
			global, settingsErr := s.settings.Load(ctx)
			if settingsErr != nil {
				return appErrors.Wrap(settingsErr, appErrors.DATABASE_ERROR, "读取微信开发者工具设置失败")
			}
			config.DeveloperToolsCLI = global.DeveloperToolsCLI
		}
	}
	if err := deployment.Validate(config, item.Kind); err != nil {
		return appErrors.Wrap(err, appErrors.INVALID_ARGUMENT, "发布配置无效")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.cancel[config.ProjectID] != nil {
		return appErrors.New(appErrors.DEPLOYMENT_FAILED, "当前有发布任务运行或应用正在退出，不能修改配置")
	}
	if config.Provider == deployment.PROVIDER_WECHAT {
		config.DeveloperToolsCLI = ""
	}
	if err := s.repository.Save(ctx, config); err != nil {
		return appErrors.Wrap(err, appErrors.DATABASE_ERROR, "保存发布配置失败")
	}
	delete(s.plans, config.ProjectID)
	return nil
}
func (s *DeploymentService) prepared(ctx context.Context, id int64) (deployment.Config, *project.Project, ports.DeploymentProvider, error) {
	item, err := s.project(ctx, id)
	if err != nil {
		return deployment.Config{}, nil, nil, err
	}
	config, err := s.Config(ctx, id)
	if err != nil {
		return config, nil, nil, err
	}
	if err := deployment.Validate(config, item.Kind); err != nil {
		return config, nil, nil, appErrors.Wrap(err, appErrors.INVALID_ARGUMENT, "请先保存有效的发布配置")
	}
	provider, ok := s.providers[config.Provider]
	if !ok {
		return config, nil, nil, appErrors.New(appErrors.INVALID_ARGUMENT, "发布提供器不存在")
	}
	return config, item, provider, nil
}
func (s *DeploymentService) Preview(ctx context.Context, id int64) (deployment.Plan, error) {
	config, item, provider, err := s.prepared(ctx, id)
	if err != nil {
		return deployment.Plan{}, err
	}
	plan, err := provider.Preview(ctx, config, item.Root)
	if err != nil {
		return plan, appErrors.Wrap(err, appErrors.DEPLOYMENT_FAILED, "生成发布预览失败")
	}
	plan.Token = uuid.NewString()
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.cancel[id] != nil {
		return deployment.Plan{}, appErrors.New(appErrors.DEPLOYMENT_FAILED, "发布任务正在运行或应用正在退出")
	}
	s.plans[id] = deploymentPlan{plan: plan, config: config, root: item.Root, expires: time.Now().Add(DEPLOYMENT_PLAN_TTL)}
	return plan, nil
}
func (s *DeploymentService) Test(ctx context.Context, id int64, credentials deployment.Credentials) error {
	config, item, provider, err := s.prepared(ctx, id)
	if err != nil {
		return err
	}
	testContext, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	if err := provider.Test(testContext, config, item.Root, credentials); err != nil {
		return appErrors.New(appErrors.DEPLOYMENT_FAILED, "发布连接或环境检查失败: "+redactDeployment(err.Error(), credentials))
	}
	return nil
}
func (s *DeploymentService) Start(ctx context.Context, id int64, token string, credentials deployment.Credentials) error {
	config, item, provider, err := s.prepared(ctx, id)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.cancel[id] != nil {
		return appErrors.New(appErrors.DEPLOYMENT_FAILED, "当前项目已有发布任务或应用正在退出")
	}
	plan, ok := s.plans[id]
	if !ok || token == "" || plan.plan.Token != token || time.Now().After(plan.expires) || !reflect.DeepEqual(plan.config, config) || plan.root != item.Root {
		return appErrors.New(appErrors.INVALID_ARGUMENT, "预览已失效、配置已改变或已被使用，请重新预览并确认")
	}
	targetKeys := []string{config.Provider + ":" + strings.ToLower(config.Host) + fmt.Sprintf(":%d", config.Port) + config.Remote}
	if config.Provider == deployment.PROVIDER_WECHAT {
		targetKeys = make([]string, 0, len(config.AppIDValues()))
		for _, appID := range config.AppIDValues() {
			targetKeys = append(targetKeys, config.Provider+":"+appID)
		}
	}
	for _, targetKey := range targetKeys {
		if s.targets[targetKey] {
			return appErrors.New(appErrors.DEPLOYMENT_FAILED, "同一个发布目标已有任务，请等待完成")
		}
	}
	total := len(plan.plan.Files)
	if config.Provider == deployment.PROVIDER_WECHAT {
		total *= len(config.AppIDValues())
	}
	task := deployment.Task{ID: uuid.NewString(), ProjectID: id, Provider: config.Provider, State: deployment.STATE_RUNNING, Message: "准备开始手动发布", Target: plan.plan.Target, Total: total, Created: time.Now().UTC().Format(time.RFC3339Nano)}
	if err := s.repository.Record(ctx, task); err != nil {
		return appErrors.Wrap(err, appErrors.DATABASE_ERROR, "保存任务记录失败，未开始上传")
	}
	workerContext, cancel := context.WithTimeout(context.Background(), DEPLOYMENT_TIMEOUT)
	s.cancel[id] = cancel
	for _, targetKey := range targetKeys {
		s.targets[targetKey] = true
	}
	s.tasks[id] = task
	delete(s.plans, id)
	s.group.Add(1)
	go s.run(workerContext, provider, plan, credentials, task, targetKeys)
	return nil
}
func (s *DeploymentService) run(ctx context.Context, provider ports.DeploymentProvider, plan deploymentPlan, credentials deployment.Credentials, task deployment.Task, targetKeys []string) {
	defer s.group.Done()
	report := func(completed int, message string) {
		s.mu.Lock()
		task.Completed = completed
		task.Message = redactDeployment(message, credentials)
		s.tasks[task.ProjectID] = task
		s.mu.Unlock()
	}
	err := provider.Run(ctx, plan.config, plan.root, plan.plan, credentials, report)
	s.mu.Lock()
	defer s.mu.Unlock()
	task.State = deployment.STATE_SUCCESS
	task.Message = "手动同步已完成；未删除远端文件、未执行依赖安装或服务重启"
	if task.Provider == deployment.PROVIDER_WECHAT {
		task.Message = "代码已上传微信开发版本，请在微信后台核对并手动提审、发布"
	}
	if err != nil {
		task.State = deployment.STATE_FAILED
		task.Message = redactDeployment(err.Error(), credentials)
		if ctx.Err() != nil {
			task.State = deployment.STATE_CANCELLED
			task.Message = "任务已取消或超时，可能已经完成部分上传；请核对目标状态，不会自动重试。" + task.Message
		}
	}
	task.Finished = time.Now().UTC().Format(time.RFC3339Nano)
	if err := s.repository.Record(context.Background(), task); err != nil {
		task.Message += "；写入发布历史失败: " + err.Error()
	}
	s.tasks[task.ProjectID] = task
	s.cancel[task.ProjectID]()
	delete(s.cancel, task.ProjectID)
	for _, targetKey := range targetKeys {
		delete(s.targets, targetKey)
	}
}
func redactDeployment(message string, credentials deployment.Credentials) string {
	for _, secret := range []string{credentials.Password, credentials.Passphrase} {
		if secret != "" {
			message = strings.ReplaceAll(message, secret, "[已隐藏]")
		}
	}
	return message
}
func (s *DeploymentService) History(ctx context.Context, id int64) ([]deployment.Task, error) {
	if _, err := s.project(ctx, id); err != nil {
		return nil, err
	}
	items, err := s.repository.History(ctx, id)
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.DATABASE_ERROR, "读取发布记录失败")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if latest, ok := s.tasks[id]; ok {
		for index, item := range items {
			if item.ID == latest.ID {
				items[index] = latest
			}
		}
	}
	return items, nil
}
func (s *DeploymentService) Cancel(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if cancel, ok := s.cancel[id]; ok {
		cancel()
		return nil
	}
	return appErrors.New(appErrors.NOT_FOUND, "没有正在执行的发布任务")
}
func (s *DeploymentService) Close() {
	s.mu.Lock()
	s.closed = true
	for _, cancel := range s.cancel {
		cancel()
	}
	s.mu.Unlock()
	s.group.Wait()
}

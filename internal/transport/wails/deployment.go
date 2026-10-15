package wails

import "devenv/internal/domain/deployment"

// DTO names and fields match the deployment domain; secrets exist only on an explicit operation input.
type DeploymentConfigDTO deployment.Config
type DeploymentFileDTO deployment.File
type DeploymentPlanDTO struct {
	Token     string              `json:"token"`
	ProjectID int64               `json:"projectId"`
	Provider  string              `json:"provider"`
	Source    string              `json:"source"`
	Target    string              `json:"target"`
	Files     []DeploymentFileDTO `json:"files"`
	Bytes     int64               `json:"bytes"`
	Excluded  int                 `json:"excluded"`
}
type DeploymentTaskDTO deployment.Task
type DeploymentInput struct {
	ProjectID  int64  `json:"projectId"`
	Token      string `json:"token"`
	Password   string `json:"password"`
	Passphrase string `json:"passphrase"`
}

func (a *ProjectAPI) DeploymentConfig(id int64) (DeploymentConfigDTO, error) {
	config, err := a.deployment.Config(a.ctx, id)
	return DeploymentConfigDTO(config), err
}
func (a *ProjectAPI) SaveDeployment(config DeploymentConfigDTO) Result {
	if err := a.deployment.Save(a.ctx, deployment.Config(config)); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "发布配置已保存，不会自动上传"}
}
func (a *ProjectAPI) PreviewDeployment(id int64) (DeploymentPlanDTO, error) {
	plan, err := a.deployment.Preview(a.ctx, id)
	if err != nil {
		return DeploymentPlanDTO{}, err
	}
	files := make([]DeploymentFileDTO, 0, len(plan.Files))
	for _, file := range plan.Files {
		files = append(files, DeploymentFileDTO(file))
	}
	return DeploymentPlanDTO{Token: plan.Token, ProjectID: plan.ProjectID, Provider: plan.Provider, Source: plan.Source, Target: plan.Target, Files: files, Bytes: plan.Bytes, Excluded: plan.Excluded}, nil
}
func (a *ProjectAPI) TestDeployment(input DeploymentInput) Result {
	if err := a.deployment.Test(a.ctx, input.ProjectID, deployment.Credentials{Password: input.Password, Passphrase: input.Passphrase}); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "检查通过，未上传任何文件（微信仅检查本地工具与配置）"}
}
func (a *ProjectAPI) StartDeployment(input DeploymentInput) Result {
	if err := a.deployment.Start(a.ctx, input.ProjectID, input.Token, deployment.Credentials{Password: input.Password, Passphrase: input.Passphrase}); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "手动发布任务已启动"}
}
func (a *ProjectAPI) CancelDeployment(id int64) Result {
	if err := a.deployment.Cancel(id); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "已请求取消，请核对目标状态，已上传部分不会回滚"}
}
func (a *ProjectAPI) DeploymentHistory(id int64) ([]DeploymentTaskDTO, error) {
	items, err := a.deployment.History(a.ctx, id)
	if err != nil {
		return nil, err
	}
	result := make([]DeploymentTaskDTO, 0, len(items))
	for _, item := range items {
		result = append(result, DeploymentTaskDTO(item))
	}
	return result, nil
}

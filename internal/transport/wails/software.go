package wails

import (
	"context"

	"devenv/internal/app"
	domainmonitor "devenv/internal/domain/monitor"
	domainredis "devenv/internal/domain/redis"
	"devenv/internal/domain/software"
)

type SoftwareAPI struct {
	ctx     context.Context
	service *app.SoftwareService
}

func NewSoftwareAPI(service *app.SoftwareService) *SoftwareAPI {
	return &SoftwareAPI{
		ctx:     context.Background(),
		service: service,
	}
}

func (a *SoftwareAPI) SetContext(ctx context.Context) {
	a.ctx = ctx
}

func (a *SoftwareAPI) Market() ([]ItemDTO, error) {
	items, err := a.service.Market(a.ctx)
	if err != nil {
		return nil, err
	}
	result := make([]ItemDTO, 0, len(items))
	for _, item := range items {
		result = append(result, ItemDTO{
			Kind:        item.Kind,
			Name:        item.Name,
			Description: item.Description,
			Category:    item.Category,
			Count:       item.Count,
			Recommended: item.Recommended,
		})
	}
	return result, nil
}

func (a *SoftwareAPI) List() ([]SoftwareDTO, error) {
	list, err := a.service.List(a.ctx)
	if err != nil {
		return nil, err
	}
	result := make([]SoftwareDTO, 0, len(list))
	for _, item := range list {
		result = append(result, SoftwareDTO{
			ID:        item.ID,
			Kind:      item.Kind,
			Name:      software.Name(item.Kind),
			Version:   item.Version,
			Path:      item.Path,
			State:     item.State,
			Progress:  item.Progress,
			Error:     item.Error,
			Main:      item.Main,
			Port:      item.Port,
			Autostart: item.Autostart,
			Running:   item.Running,
			PID:       item.PID,
			Metrics:   convertProcessMetrics(item.Metrics),
			Daemon:    software.IsDaemon(item.Kind),
			Service:   item.State == software.STATE_INSTALLED && software.IsService(item.Kind) && (software.IsDaemon(item.Kind) || item.Main),
		})
	}
	return result, nil
}

func convertProcessMetrics(item domainmonitor.ProcessMetrics) ProcessMetricsDTO {
	return ProcessMetricsDTO{
		CPUPercent: item.CPUPercent, MemoryBytes: item.MemoryBytes,
		ReadBytes: item.ReadBytes, WriteBytes: item.WriteBytes,
		UptimeSeconds: item.UptimeSeconds,
	}
}

func (a *SoftwareAPI) Versions(kind string) ([]VersionDTO, error) {
	list, err := a.service.Versions(a.ctx, kind)
	if err != nil {
		return nil, err
	}
	result := make([]VersionDTO, 0, len(list))
	for _, item := range list {
		result = append(result, VersionDTO{
			Version:     item.Version,
			Recommended: item.Recommended,
			Installed:   item.Installed,
			Note:        item.Note,
		})
	}
	return result, nil
}

func (a *SoftwareAPI) Config(id int64) (ConfigDTO, error) {
	config, err := a.service.Config(a.ctx, id)
	if err != nil {
		return ConfigDTO{Items: make([]ConfigItem, 0)}, err
	}
	items := make([]ConfigItem, 0, len(config.Items))
	for _, item := range config.Items {
		items = append(items, ConfigItem{
			Key:         item.Key,
			Label:       item.Label,
			Value:       item.Value,
			Description: item.Description,
		})
	}
	return ConfigDTO{
		Items: items,
		File:  config.File,
	}, nil
}

func (a *SoftwareAPI) SaveConfig(in ConfigInput) ConfigResult {
	restart, err := a.service.SaveConfig(a.ctx, in.ID, in.Items)
	if err != nil {
		return ConfigResult{OK: false, Message: err.Error(), Restart: false}
	}
	message := "配置已保存并立即生效"
	if restart {
		message = "配置已保存，重启服务后生效"
	}
	return ConfigResult{OK: true, Message: message, Restart: restart}
}

func (a *SoftwareAPI) ReadConfig(id int64) (string, error) {
	return a.service.ReadConfig(a.ctx, id)
}

func (a *SoftwareAPI) WriteConfig(id int64, text string) Result {
	if err := a.service.WriteConfig(a.ctx, id, text); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "配置文件已保存"}
}

func (a *SoftwareAPI) Extensions(id int64) ([]ExtensionDTO, error) {
	extensions, err := a.service.Extensions(a.ctx, id)
	if err != nil {
		return nil, err
	}
	result := make([]ExtensionDTO, 0, len(extensions))
	for _, item := range extensions {
		result = append(result, ExtensionDTO{
			Name:    item.Name,
			Enabled: item.Enabled,
		})
	}
	return result, nil
}

func (a *SoftwareAPI) SetExtension(id int64, name string, on bool) Result {
	if err := a.service.SetExtension(a.ctx, id, name, on); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "扩展状态已更新"}
}

func (a *SoftwareAPI) OpenFolder(id int64) Result {
	if err := a.service.OpenFolder(a.ctx, id); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "已打开目录"}
}

func (a *SoftwareAPI) Install(in InstallInput) Result {
	if err := a.service.Install(a.ctx, in.Kind, in.Version); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "已开始安装"}
}

func (a *SoftwareAPI) Cancel(kind, version string) Result {
	if err := a.service.Cancel(a.ctx, kind, version); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "已取消安装"}
}

func (a *SoftwareAPI) Remove(id int64) Result {
	if err := a.service.Remove(a.ctx, id); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "软件已卸载"}
}

func (a *SoftwareAPI) SetMain(id int64) Result {
	if err := a.service.SetMain(a.ctx, id); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "已设为默认版本"}
}

func (a *SoftwareAPI) Start(id int64) Result {
	if err := a.service.StartAsync(a.ctx, id); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "服务启动中"}
}

func (a *SoftwareAPI) Stop(id int64) Result {
	if err := a.service.StopAsync(a.ctx, id); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "服务停止中"}
}

func (a *SoftwareAPI) Restart(id int64) Result {
	if err := a.service.RestartAsync(a.ctx, id); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "服务重启中"}
}

func (a *SoftwareAPI) SetAutostart(id int64, on bool) Result {
	if err := a.service.SetAutostart(a.ctx, id, on); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "自启动设置已更新"}
}

func (a *SoftwareAPI) Logs(id int64, limit int) ([]string, error) {
	return a.service.Logs(a.ctx, id, limit)
}

func (a *SoftwareAPI) RedisDatabases(id int64) ([]RedisDatabaseDTO, error) {
	items, err := a.service.RedisDatabases(a.ctx, id)
	if err != nil {
		return make([]RedisDatabaseDTO, 0), err
	}
	result := make([]RedisDatabaseDTO, 0, len(items))
	for _, item := range items {
		result = append(result, RedisDatabaseDTO{Index: item.Index, Keys: item.Keys, Expires: item.Expires})
	}
	return result, nil
}

func (a *SoftwareAPI) RedisKeys(in RedisKeysInput) (RedisKeyPageDTO, error) {
	result, err := a.service.RedisKeys(a.ctx, in.ID, domainredis.KeyQuery{
		Database: in.Database,
		Pattern:  in.Pattern,
		Cursor:   in.Cursor,
		Limit:    in.Limit,
	})
	if err != nil {
		return RedisKeyPageDTO{Keys: make([]RedisKeyDTO, 0)}, err
	}
	keys := make([]RedisKeyDTO, 0, len(result.Keys))
	for _, item := range result.Keys {
		keys = append(keys, RedisKeyDTO{Name: item.Name, Type: item.Type, TTL: item.TTL})
	}
	return RedisKeyPageDTO{Cursor: result.Cursor, Keys: keys}, nil
}

func (a *SoftwareAPI) RedisValue(in RedisValueInput) (RedisValueDTO, error) {
	result, err := a.service.RedisValue(a.ctx, in.ID, domainredis.ValueQuery{
		Database: in.Database,
		Key:      in.Key,
		Cursor:   in.Cursor,
		Limit:    in.Limit,
	})
	if err != nil {
		return RedisValueDTO{Entries: make([]RedisEntryDTO, 0)}, err
	}
	entries := make([]RedisEntryDTO, 0, len(result.Entries))
	for _, item := range result.Entries {
		entries = append(entries, RedisEntryDTO{Field: item.Field, Value: item.Value, Score: item.Score})
	}
	return RedisValueDTO{
		Type:      result.Type,
		TTL:       result.TTL,
		Length:    result.Length,
		Text:      result.Text,
		Encoding:  result.Encoding,
		Truncated: result.Truncated,
		Cursor:    result.Cursor,
		Entries:   entries,
	}, nil
}

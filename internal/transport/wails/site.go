package wails

import (
	"context"

	"devenv/internal/app"
	"devenv/internal/domain/site"
)

type SiteAPI struct {
	ctx     context.Context
	service *app.SiteService
}

func NewSiteAPI(service *app.SiteService) *SiteAPI {
	return &SiteAPI{
		ctx:     context.Background(),
		service: service,
	}
}

func (a *SiteAPI) SetContext(ctx context.Context) {
	a.ctx = ctx
}

func (a *SiteAPI) List() ([]SiteDTO, error) {
	list, err := a.service.List(a.ctx)
	if err != nil {
		return nil, err
	}
	result := make([]SiteDTO, 0, len(list))
	for _, item := range list {
		result = append(result, convertSite(item))
	}
	return result, nil
}

func (a *SiteAPI) Get(id int64) (*SiteDTO, error) {
	item, err := a.service.Get(a.ctx, id)
	if err != nil {
		return nil, err
	}
	dto := convertSite(*item)
	return &dto, nil
}

func (a *SiteAPI) Create(in CreateInput) (*SiteDTO, error) {
	item, err := a.service.Create(a.ctx, site.CreateInput{
		Kind:    in.Kind,
		Name:    in.Name,
		Host:    in.Host,
		Root:    in.Root,
		Run:     in.Run,
		Command: in.Command,
	})
	if err != nil {
		return nil, err
	}
	dto := convertSite(*item)
	return &dto, nil
}

func (a *SiteAPI) Update(in UpdateInput) (*SiteDTO, error) {
	item, err := a.service.Update(a.ctx, site.UpdateInput{
		ID:        in.ID,
		Hosts:     in.Hosts,
		Root:      in.Root,
		Run:       in.Run,
		Software:  in.Software,
		Command:   in.Command,
		Nginx:     in.Nginx,
		Autostart: in.Autostart,
	})
	if err != nil {
		return nil, err
	}
	dto := convertSite(*item)
	return &dto, nil
}

func (a *SiteAPI) FindRun(root, kind string) string {
	return a.service.FindRun(root, kind)
}

func (a *SiteAPI) Start(id int64) Result {
	if err := a.service.Start(a.ctx, id); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "站点已启动"}
}

func (a *SiteAPI) Stop(id int64) Result {
	if err := a.service.Stop(a.ctx, id); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "站点已停止"}
}

func (a *SiteAPI) Restart(id int64) Result {
	if err := a.service.Restart(a.ctx, id); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "站点已重启"}
}

func (a *SiteAPI) Delete(id int64) Result {
	if err := a.service.Delete(a.ctx, id); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: "站点记录已删除"}
}

func (a *SiteAPI) Logs(id int64, kind string, limit int) ([]string, error) {
	return a.service.Logs(a.ctx, id, kind, limit)
}

func (a *SiteAPI) OpenNginxConfig(id int64) Result {
	if err := a.service.OpenNginxConfig(a.ctx, id); err != nil {
		return Result{OK: false, Message: err.Error()}
	}
	return Result{OK: true, Message: ""}
}

func convertSite(view app.SiteView) SiteDTO {
	hosts := view.Hosts
	if hosts == nil {
		hosts = make([]string, 0)
	}
	return SiteDTO{
		ID:        view.ID,
		Name:      view.Name,
		Kind:      view.Kind,
		Hosts:     hosts,
		Root:      view.Root,
		Run:       view.Run,
		Path:      view.Path,
		Software:  view.Software,
		Version:   view.Version,
		Command:   view.Command,
		Nginx:     view.Nginx,
		Port:      view.Port,
		Autostart: view.Autostart,
		State:     view.State,
		URL:       view.URL,
		Need:      view.Need,
		Created:   view.Created,
		Updated:   view.Updated,
	}
}

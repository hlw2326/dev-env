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
	res := make([]SiteDTO, 0, len(list))
	for _, item := range list {
		res = append(res, toSiteDTO(item))
	}
	return res, nil
}

func (a *SiteAPI) Get(id int64) (*SiteDTO, error) {
	item, err := a.service.Get(a.ctx, id)
	if err != nil {
		return nil, err
	}
	dto := toSiteDTO(*item)
	return &dto, nil
}

func (a *SiteAPI) Create(in CreateSiteInput) (*SiteDTO, error) {
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
	dto := toSiteDTO(*item)
	return &dto, nil
}

func (a *SiteAPI) Update(in UpdateSiteInput) (*SiteDTO, error) {
	item, err := a.service.Update(a.ctx, site.UpdateInput{
		ID:        in.ID,
		Hosts:     in.Hosts,
		Root:      in.Root,
		Run:       in.Run,
		Software:  in.Software,
		Command:   in.Command,
		Autostart: in.Autostart,
	})
	if err != nil {
		return nil, err
	}
	dto := toSiteDTO(*item)
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

func toSiteDTO(v app.SiteView) SiteDTO {
	return SiteDTO{
		ID:        v.ID,
		Name:      v.Name,
		Kind:      v.Kind,
		Hosts:     v.Hosts,
		Root:      v.Root,
		Run:       v.Run,
		Path:      v.Path,
		Software:  v.Software,
		Version:   v.Version,
		Command:   v.Command,
		Port:      v.Port,
		Autostart: v.Autostart,
		State:     v.State,
		URL:       v.URL,
		Need:      v.Need,
		Created:   v.Created,
		Updated:   v.Updated,
	}
}

package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"devenv/internal/domain/runtime"
	"devenv/internal/domain/service"
	"devenv/internal/domain/site"
	"devenv/internal/ports"
	appErrors "devenv/internal/shared/errors"
)

type SiteService struct {
	sites     ports.SiteRepository
	runtimes  ports.RuntimeRepository
	instances ports.ServiceRepository
	services  *ServiceService
	nginx     ports.NginxManager
	git       ports.GitManager
	probe     ports.PortChecker
	bus       ports.EventPublisher
}

func NewSiteService(
	sites ports.SiteRepository,
	runtimes ports.RuntimeRepository,
	instances ports.ServiceRepository,
	services *ServiceService,
	nginx ports.NginxManager,
	git ports.GitManager,
	probe ports.PortChecker,
	bus ports.EventPublisher,
) *SiteService {
	return &SiteService{
		sites:     sites,
		runtimes:  runtimes,
		instances: instances,
		services:  services,
		nginx:     nginx,
		git:       git,
		probe:     probe,
		bus:       bus,
	}
}

func (s *SiteService) List(ctx context.Context) ([]site.Site, error) {
	return s.sites.List(ctx)
}

func (s *SiteService) Get(ctx context.Context, id int64) (*site.Site, error) {
	cur, err := s.sites.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if cur == nil {
		return nil, appErrors.New(appErrors.CodeNotFound, "网站不存在")
	}
	return cur, nil
}

func (s *SiteService) Create(ctx context.Context, in site.CreateInput) (*site.Site, error) {
	if err := in.Check(); err != nil {
		return nil, appErrors.New(appErrors.CodeInvalidArgument, err.Error())
	}

	for _, domain := range in.Domains {
		d := strings.ToLower(strings.TrimSpace(domain))
		if d == "localhost" || d == "127.0.0.1" {
			continue
		}
		cur, err := s.sites.GetByDomain(ctx, domain)
		if err != nil {
			return nil, err
		}
		if cur != nil {
			return nil, appErrors.New(appErrors.CodeAlreadyExists, fmt.Sprintf("域名 %s 已被网站「%s」占用", domain, cur.Name))
		}
	}

	root := filepath.Clean(strings.TrimSpace(in.Root))

	if in.Git && strings.TrimSpace(in.Remote) != "" {
		if err := s.git.Clone(ctx, strings.TrimSpace(in.Remote), root, in.Branch); err != nil {
			return nil, appErrors.Wrap(err, appErrors.CodeInvalidArgument, "Git 克隆失败")
		}
	} else {
		if err := os.MkdirAll(root, 0755); err != nil {
			return nil, appErrors.Wrap(err, appErrors.CodePathNotAccessible, "无法创建网站根目录")
		}
	}

	if _, err := os.Stat(filepath.Join(root, ".git")); err == nil {
		in.Git = true
	}

	preset := site.Presets[in.Kind]
	upstream := in.Upstream
	if upstream == nil && preset.Port > 0 {
		p := preset.Port
		upstream = &p
	}

	var serviceID *int64
	if in.AutoService && in.RuntimeID != nil && *in.RuntimeID > 0 {
		rt, err := s.runtimes.GetByID(ctx, *in.RuntimeID)
		if err != nil {
			return nil, err
		}
		now := time.Now().UTC()
		sid, err := s.instances.Create(ctx, service.Instance{
			Name:        fmt.Sprintf("%s (%s %s)", in.Name, strings.ToUpper(string(rt.RuntimeType)), rt.Version),
			ServiceType: rt.RuntimeType,
			RuntimeID:   rt.ID,
			Port:        upstream,
			AutoStart:   in.AutoStart,
			Enabled:     true,
			CreatedAt:   now,
			UpdatedAt:   now,
		})
		if err != nil {
			return nil, err
		}
		serviceID = &sid
	} else if in.ServiceID != nil && *in.ServiceID > 0 {
		serviceID = in.ServiceID
		srv, err := s.instances.GetByID(ctx, *in.ServiceID)
		if err != nil {
			return nil, err
		}
		if srv.Port != nil {
			upstream = srv.Port
		}
	}

	domain := in.Domains[0]
	meta := site.Meta{
		Entry:     in.Entry,
		Cmd:       in.Cmd,
		AutoStart: in.AutoStart,
		Env:       in.Env,
		Custom:    in.Custom,
		AccessLog: fmt.Sprintf("logs/%s-access.log", domain),
		ErrorLog:  fmt.Sprintf("logs/%s-error.log", domain),
		History:   make([]site.Snapshot, 0),
	}
	if meta.Entry == "" {
		meta.Entry = preset.Entry
	}
	if meta.Cmd == "" {
		meta.Cmd = preset.Cmd
	}

	now := time.Now().UTC()
	cur := &site.Site{
		Name:      strings.TrimSpace(in.Name),
		Kind:      in.Kind,
		Domains:   in.Domains,
		Root:      root,
		RuntimeID: in.RuntimeID,
		ServiceID: serviceID,
		Git:       in.Git,
		Remote:    strings.TrimSpace(in.Remote),
		Branch:    in.Branch,
		Port:      in.Port,
		TLS:       in.TLS,
		Issuer:    in.Issuer,
		Cert:      in.Cert,
		Key:       in.Key,
		Proxy:     preset.Proxy,
		Upstream:  upstream,
		Meta:      meta,
		Status:    site.Stopped,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if cur.Git {
		st, err := s.git.Status(ctx, root)
		if err != nil {
			return nil, err
		}
		cur.Branch = st.Branch
		cur.Commit = st.Commit
		cur.PulledAt = now.Format(time.RFC3339)
	}

	conf := s.nginx.Generate(cur)
	cur.Meta.Push("初始创建配置", conf)

	if err := s.sites.Create(ctx, cur); err != nil {
		return nil, err
	}

	path, err := s.nginx.Write(ctx, cur.ID, domain, conf)
	if err != nil {
		return nil, err
	}
	cur.Conf = path
	if err := s.sites.Update(ctx, cur); err != nil {
		return nil, err
	}

	if err := s.writeWelcome(cur); err != nil {
		return nil, err
	}

	_ = s.syncHosts(cur.Domains, nil)

	if in.StartNow {
		if err := s.Start(ctx, cur.ID); err != nil {
			return nil, err
		}
		updated, err := s.sites.Get(ctx, cur.ID)
		if err != nil {
			return nil, err
		}
		cur = updated
	}

	s.bus.Emit("site:created", cur.ID)
	return cur, nil
}

func (s *SiteService) writeWelcome(cur *site.Site) error {
	if cur.Meta.Entry == "" {
		return nil
	}
	target := filepath.Join(cur.Root, cur.Meta.Entry)
	if _, err := os.Stat(target); err == nil {
		return nil
	}
	switch cur.Kind {
	case site.Static:
		return os.WriteFile(target, []byte(fmt.Sprintf("<!DOCTYPE html>\n<html>\n<head><meta charset=\"utf-8\"><title>%s</title></head>\n<body><h1>%s</h1></body>\n</html>\n", cur.Name, cur.Name)), 0644)
	case site.PHP:
		return os.WriteFile(target, []byte(fmt.Sprintf("<?php\necho '<h1>%s</h1>';\nphpinfo();\n", cur.Name)), 0644)
	}
	return nil
}

func (s *SiteService) Update(ctx context.Context, in site.UpdateInput) (*site.Site, error) {
	cur, err := s.sites.Get(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	if cur == nil {
		return nil, appErrors.New(appErrors.CodeNotFound, "网站不存在")
	}

	oldDomains := append([]string(nil), cur.Domains...)

	if in.Name != nil {
		name := strings.TrimSpace(*in.Name)
		if name == "" {
			return nil, appErrors.New(appErrors.CodeInvalidArgument, "网站名称不能为空")
		}
		cur.Name = name
	}
	if len(in.Domains) > 0 {
		clean := make([]string, 0, len(in.Domains))
		for _, d := range in.Domains {
			trimmed := strings.ToLower(strings.TrimSpace(d))
			if trimmed != "" {
				clean = append(clean, trimmed)
			}
		}
		if len(clean) == 0 {
			return nil, appErrors.New(appErrors.CodeInvalidArgument, "至少需要绑定一个域名")
		}
		cur.Domains = clean
	}
	if in.Root != nil {
		root := filepath.Clean(strings.TrimSpace(*in.Root))
		if root == "" {
			return nil, appErrors.New(appErrors.CodeInvalidArgument, "网站根目录不能为空")
		}
		cur.Root = root
	}
	if in.RuntimeID != nil {
		cur.RuntimeID = in.RuntimeID
	}
	if in.ServiceID != nil {
		cur.ServiceID = in.ServiceID
	}
	if in.Port != nil {
		if *in.Port <= 0 || *in.Port > 65535 {
			return nil, appErrors.New(appErrors.CodeInvalidArgument, "监听端口必须在 1 到 65535 之间")
		}
		cur.Port = *in.Port
	}
	if in.TLS != nil {
		cur.TLS = *in.TLS
	}
	if in.Issuer != nil {
		cur.Issuer = *in.Issuer
	}
	if in.Cert != nil {
		cur.Cert = *in.Cert
	}
	if in.Key != nil {
		cur.Key = *in.Key
	}
	if in.Upstream != nil {
		cur.Upstream = in.Upstream
	}
	if in.Entry != nil {
		cur.Meta.Entry = *in.Entry
	}
	if in.Cmd != nil {
		cur.Meta.Cmd = *in.Cmd
	}
	if in.Custom != nil {
		cur.Meta.Custom = *in.Custom
	}
	if in.AutoStart != nil {
		cur.Meta.AutoStart = *in.AutoStart
	}

	conf := s.nginx.Generate(cur)
	cur.Meta.Push("更新站点基础配置", conf)

	path, err := s.nginx.Write(ctx, cur.ID, cur.Domains[0], conf)
	if err != nil {
		return nil, err
	}
	cur.Conf = path

	if err := s.sites.Update(ctx, cur); err != nil {
		return nil, err
	}

	_ = s.syncHosts(cur.Domains, diffRemovedDomains(oldDomains, cur.Domains))

	s.bus.Emit("site:updated", cur.ID)
	return cur, nil
}

func (s *SiteService) Start(ctx context.Context, id int64) error {
	cur, err := s.sites.Get(ctx, id)
	if err != nil {
		return err
	}
	if cur == nil {
		return appErrors.New(appErrors.CodeNotFound, "网站不存在")
	}

	if cur.ServiceID != nil && *cur.ServiceID > 0 {
		if err := s.services.Start(ctx, *cur.ServiceID); err != nil {
			return err
		}
	}

	conf, err := s.ReadConf(ctx, id)
	if err != nil {
		return err
	}
	ok, msg := s.nginx.Check(ctx, conf)
	if !ok {
		return appErrors.New(appErrors.CodeConfigInvalid, "Nginx 配置测试失败: "+msg)
	}

	if err := s.nginx.Reload(ctx); err != nil {
		return err
	}

	cur.Status = site.Running
	if err := s.sites.Update(ctx, cur); err != nil {
		return err
	}

	s.bus.Emit("site:started", id)
	return nil
}

func (s *SiteService) Stop(ctx context.Context, id int64) error {
	cur, err := s.sites.Get(ctx, id)
	if err != nil {
		return err
	}
	if cur == nil {
		return appErrors.New(appErrors.CodeNotFound, "网站不存在")
	}

	cur.Status = site.Stopped
	if err := s.sites.Update(ctx, cur); err != nil {
		return err
	}

	if err := s.nginx.Reload(ctx); err != nil {
		return err
	}

	s.bus.Emit("site:stopped", id)
	return nil
}

func (s *SiteService) Restart(ctx context.Context, id int64) error {
	if err := s.Stop(ctx, id); err != nil {
		return err
	}
	return s.Start(ctx, id)
}

func (s *SiteService) Delete(ctx context.Context, id int64, purge bool) error {
	cur, err := s.sites.Get(ctx, id)
	if err != nil {
		return err
	}
	if cur == nil {
		return appErrors.New(appErrors.CodeNotFound, "网站不存在")
	}

	if cur.Status == site.Running {
		if err := s.Stop(ctx, id); err != nil {
			return err
		}
	}

	if cur.Conf != "" {
		if err := s.nginx.Delete(ctx, cur.Conf); err != nil {
			return err
		}
	}

	if err := s.sites.Delete(ctx, id); err != nil {
		return err
	}

	_ = s.syncHosts(nil, cur.Domains)

	if purge && cur.Root != "" {
		if err := os.RemoveAll(cur.Root); err != nil {
			return err
		}
	}

	s.bus.Emit("site:deleted", id)
	return nil
}

func (s *SiteService) ReadConf(ctx context.Context, id int64) (string, error) {
	cur, err := s.sites.Get(ctx, id)
	if err != nil {
		return "", err
	}
	if cur == nil {
		return "", appErrors.New(appErrors.CodeNotFound, "网站不存在")
	}

	if cur.Conf != "" {
		content, err := s.nginx.Read(ctx, cur.Conf)
		if err == nil && content != "" {
			return content, nil
		}
	}
	return s.nginx.Generate(cur), nil
}

func (s *SiteService) RenderConf(ctx context.Context, id int64) (string, error) {
	cur, err := s.sites.Get(ctx, id)
	if err != nil {
		return "", err
	}
	if cur == nil {
		return "", appErrors.New(appErrors.CodeNotFound, "网站不存在")
	}
	return s.nginx.Generate(cur), nil
}

func (s *SiteService) WriteConf(ctx context.Context, id int64, content, summary string) error {
	cur, err := s.sites.Get(ctx, id)
	if err != nil {
		return err
	}
	if cur == nil {
		return appErrors.New(appErrors.CodeNotFound, "网站不存在")
	}

	ok, msg := s.nginx.Check(ctx, content)
	if !ok {
		return appErrors.New(appErrors.CodeConfigInvalid, msg)
	}

	path, err := s.nginx.Write(ctx, cur.ID, cur.Domains[0], content)
	if err != nil {
		return err
	}
	cur.Conf = path
	cur.Meta.Push(summary, content)

	if err := s.sites.Update(ctx, cur); err != nil {
		return err
	}
	return s.nginx.Reload(ctx)
}

func (s *SiteService) CheckConf(ctx context.Context, content string) (bool, string) {
	return s.nginx.Check(ctx, content)
}

func (s *SiteService) Logs(ctx context.Context, id int64, kind string, limit int) ([]string, error) {
	cur, err := s.sites.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if cur == nil {
		return nil, appErrors.New(appErrors.CodeNotFound, "网站不存在")
	}

	logPath := cur.Meta.AccessLog
	if kind == "error" {
		logPath = cur.Meta.ErrorLog
	}
	if logPath == "" {
		return make([]string, 0), nil
	}
	if !filepath.IsAbs(logPath) {
		dataDir := os.Getenv("LOCALAPPDATA")
		if dataDir == "" {
			dataDir = "."
		}
		logPath = filepath.Join(dataDir, "DevEnv", logPath)
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return make([]string, 0), nil
		}
		return nil, err
	}

	rawLines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	lines := make([]string, 0, len(rawLines))
	for _, line := range rawLines {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}
	if limit > 0 && len(lines) > limit {
		lines = lines[len(lines)-limit:]
	}
	return lines, nil
}

func diffRemovedDomains(oldList, newList []string) []string {
	keep := make(map[string]bool, len(newList))
	for _, d := range newList {
		keep[strings.ToLower(strings.TrimSpace(d))] = true
	}
	var removed []string
	for _, d := range oldList {
		clean := strings.ToLower(strings.TrimSpace(d))
		if clean != "" && !keep[clean] {
			removed = append(removed, clean)
		}
	}
	return removed
}

func isLocalhostDomain(d string) bool {
	return d == "" || d == "localhost" || d == "127.0.0.1" || d == "::1"
}

func (s *SiteService) syncHosts(addDomains, removeDomains []string) error {
	sysRoot := os.Getenv("SystemRoot")
	if sysRoot == "" {
		sysRoot = `C:\Windows`
	}
	hostsPath := filepath.Join(sysRoot, "System32", "drivers", "etc", "hosts")

	raw, err := os.ReadFile(hostsPath)
	if err != nil {
		return err
	}

	removeSet := make(map[string]bool, len(removeDomains))
	for _, d := range removeDomains {
		clean := strings.ToLower(strings.TrimSpace(d))
		if !isLocalhostDomain(clean) {
			removeSet[clean] = true
		}
	}

	existing := make(map[string]bool)
	lines := strings.Split(strings.ReplaceAll(string(raw), "\r\n", "\n"), "\n")
	kept := make([]string, 0, len(lines))

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			fields := strings.Fields(trimmed)
			if len(fields) >= 2 && fields[0] == "127.0.0.1" {
				host := strings.ToLower(fields[1])
				if removeSet[host] {
					continue
				}
				for _, f := range fields[1:] {
					existing[strings.ToLower(f)] = true
				}
			}
		}
		kept = append(kept, line)
	}

	changed := len(kept) != len(lines)
	for _, d := range addDomains {
		clean := strings.ToLower(strings.TrimSpace(d))
		if isLocalhostDomain(clean) || existing[clean] {
			continue
		}
		kept = append(kept, fmt.Sprintf("127.0.0.1      %s", clean))
		existing[clean] = true
		changed = true
	}

	if !changed {
		return nil
	}

	out := strings.Join(kept, "\r\n")
	if !strings.HasSuffix(out, "\r\n") {
		out += "\r\n"
	}
	return os.WriteFile(hostsPath, []byte(out), 0644)
}

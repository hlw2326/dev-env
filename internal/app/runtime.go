package app

import (
	"archive/zip"
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"devenv/internal/domain/runtime"
	"devenv/internal/ports"
	appErrors "devenv/internal/shared/errors"
)

type RuntimeService struct {
	runtimes    ports.RuntimeRepository
	sites       ports.SiteRepository
	services    ports.ServiceRepository
	settings    ports.SettingsRepository
	provisioner ports.ServiceProvisioner
	bus         ports.EventPublisher
}

func NewRuntimeService(
	runtimes ports.RuntimeRepository,
	sites ports.SiteRepository,
	services ports.ServiceRepository,
	settings ports.SettingsRepository,
	provisioner ports.ServiceProvisioner,
	bus ports.EventPublisher,
) *RuntimeService {
	return &RuntimeService{
		runtimes:    runtimes,
		sites:       sites,
		services:    services,
		settings:    settings,
		provisioner: provisioner,
		bus:         bus,
	}
}

type runtimeMeta struct {
	Default bool   `json:"isDefault"`
	Variant string `json:"variant,omitempty"`
}

func (s *RuntimeService) List(ctx context.Context) ([]runtime.RuntimeView, error) {
	list, err := s.runtimes.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	sites, err := s.sites.List(ctx)
	if err != nil {
		return nil, err
	}
	srvs, err := s.services.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	srvRuntimeMap := make(map[int64]int64, len(srvs))
	srvCountMap := make(map[int64]int)
	for _, srv := range srvs {
		srvRuntimeMap[srv.ID] = srv.RuntimeID
		srvCountMap[srv.RuntimeID]++
	}

	siteCountMap := make(map[int64]int)
	for _, st := range sites {
		if st.RuntimeID != nil && *st.RuntimeID > 0 {
			siteCountMap[*st.RuntimeID]++
		} else if st.ServiceID != nil && *st.ServiceID > 0 {
			if rtID, ok := srvRuntimeMap[*st.ServiceID]; ok {
				siteCountMap[rtID]++
			}
		}
	}

	res := make([]runtime.RuntimeView, 0, len(list))
	for _, item := range list {
		var meta runtimeMeta
		if item.MetadataJSON != "" {
			_ = json.Unmarshal([]byte(item.MetadataJSON), &meta)
		}
		variant := meta.Variant
		if variant == "" && item.RuntimeType == runtime.TypePHP {
			variant = "nts"
		}
		res = append(res, runtime.RuntimeView{
			ID:           item.ID,
			RuntimeType:  item.RuntimeType,
			Version:      item.Version,
			Variant:      variant,
			Architecture: item.Architecture,
			InstallPath:  item.InstallPath,
			Status:       item.Status,
			IsDefault:    meta.Default,
			SiteCount:    siteCountMap[item.ID],
			ServiceCount: srvCountMap[item.ID],
		})
	}
	return res, nil
}

func (s *RuntimeService) Register(ctx context.Context, rType, version, variant, arch, path string) (*runtime.RuntimeView, error) {
	rType = strings.ToLower(strings.TrimSpace(rType))
	if rType == "" {
		rType = string(runtime.TypePHP)
	}
	version = strings.TrimSpace(version)
	variant = strings.ToLower(strings.TrimSpace(variant))
	arch = strings.TrimSpace(arch)
	path = filepath.Clean(strings.TrimSpace(path))

	if version == "" || path == "" {
		return nil, appErrors.New(appErrors.CodeInvalidArgument, "版本号和安装路径不能为空")
	}
	if arch == "" {
		arch = "x64"
	}

	metaBytes, _ := json.Marshal(runtimeMeta{Variant: variant})
	now := time.Now().UTC()

	// 若同类型同版本同架构已存在则更新，避免 UNIQUE 约束冲突
	existing, _ := s.runtimes.ListByType(ctx, runtime.RuntimeType(rType))
	for _, ex := range existing {
		if ex.Version == version && ex.Architecture == arch {
			ex.InstallPath = path
			ex.Status = runtime.StatusInstalled
			ex.MetadataJSON = string(metaBytes)
			ex.UpdatedAt = now
			_ = s.runtimes.Update(ctx, ex)
			if s.provisioner != nil {
				_, _, _ = s.provisioner.EnsureForRuntime(ctx, ex)
			}
			return &runtime.RuntimeView{
				ID:           ex.ID,
				RuntimeType:  ex.RuntimeType,
				Version:      ex.Version,
				Variant:      variant,
				Architecture: ex.Architecture,
				InstallPath:  ex.InstallPath,
				Status:       ex.Status,
			}, nil
		}
	}

	cur := runtime.RuntimeVersion{
		RuntimeType:  runtime.RuntimeType(rType),
		Version:      version,
		Architecture: arch,
		InstallPath:  path,
		Status:       runtime.StatusInstalled,
		MetadataJSON: string(metaBytes),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	id, err := s.runtimes.Create(ctx, cur)
	if err != nil {
		return nil, err
	}
	cur.ID = id

	srvCount := 0
	if s.provisioner != nil {
		if srv, _, err := s.provisioner.EnsureForRuntime(ctx, cur); err == nil && srv != nil && srv.RuntimeID == id {
			srvCount = 1
		}
	}

	s.bus.Emit("runtime:registered", id)
	return &runtime.RuntimeView{
		ID:           cur.ID,
		RuntimeType:  cur.RuntimeType,
		Version:      cur.Version,
		Variant:      variant,
		Architecture: cur.Architecture,
		InstallPath:  cur.InstallPath,
		Status:       cur.Status,
		ServiceCount: srvCount,
	}, nil
}

func (s *RuntimeService) Install(ctx context.Context, in runtime.InstallInput) (*runtime.InstallResult, error) {
	rType := strings.ToLower(strings.TrimSpace(string(in.RuntimeType)))
	version := strings.TrimSpace(in.Version)
	arch := strings.TrimSpace(in.Architecture)
	if arch == "" {
		arch = "x64"
	}
	if rType == "" || version == "" {
		return nil, appErrors.New(appErrors.CodeInvalidArgument, "环境类型与版本号不能为空")
	}

	installPath := strings.TrimSpace(in.InstallPath)
	if installPath == "" {
		if in.Variant != "" {
			installPath = filepath.Join("runtimes", rType, version+"-"+in.Variant+"-"+arch)
		} else {
			installPath = filepath.Join("runtimes", rType, version+"-"+arch)
		}
	}
	installPath = filepath.Clean(installPath)
	_ = os.MkdirAll(installPath, 0755)

	s.bus.Emit("task:progress", map[string]any{"progress": 25, "message": "正在下载软件包..."})
	if strings.HasPrefix(strings.ToLower(in.DownloadURL), "http") {
		_ = s.downloadAndExtractZip(ctx, in.DownloadURL, installPath, rType)
	}
	s.bus.Emit("task:progress", map[string]any{"progress": 80, "message": "正在校验并配置环境目录..."})
	if rType == string(runtime.TypePHP) {
		ensurePHPIniConfigured(installPath)
	}

	metaBytes, _ := json.Marshal(runtimeMeta{Variant: strings.ToLower(strings.TrimSpace(in.Variant))})
	now := time.Now().UTC()

	var rt runtime.RuntimeVersion
	existing, _ := s.runtimes.ListByType(ctx, runtime.RuntimeType(rType))
	found := false
	for _, ex := range existing {
		if ex.Version == version && ex.Architecture == arch {
			ex.InstallPath = installPath
			ex.Status = runtime.StatusInstalled
			ex.MetadataJSON = string(metaBytes)
			ex.UpdatedAt = now
			if err := s.runtimes.Update(ctx, ex); err != nil {
				return nil, err
			}
			rt = ex
			found = true
			break
		}
	}

	if !found {
		cur := runtime.RuntimeVersion{
			RuntimeType:  runtime.RuntimeType(rType),
			Version:      version,
			Architecture: arch,
			InstallPath:  installPath,
			Status:       runtime.StatusInstalled,
			MetadataJSON: string(metaBytes),
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		id, err := s.runtimes.Create(ctx, cur)
		if err != nil {
			return nil, err
		}
		cur.ID = id
		rt = cur
	}

	var srvID *int64
	srvCreated := false
	if s.provisioner != nil {
		srv, created, err := s.provisioner.EnsureForRuntime(ctx, rt)
		if err != nil {
			return nil, err
		}
		if srv != nil {
			srvID = &srv.ID
			srvCreated = created
		}
	}

	s.bus.Emit("task:progress", map[string]any{"progress": 100, "message": "安装完成"})
	s.bus.Emit("task:completed", rt.ID)
	s.bus.Emit("runtime:registered", rt.ID)

	return &runtime.InstallResult{
		RuntimeID:      rt.ID,
		ServiceID:      srvID,
		ServiceCreated: srvCreated,
	}, nil
}

func (s *RuntimeService) downloadAndExtractZip(ctx context.Context, url, destDir, rType string) error {
	if rType == "php" {
		if _, err := os.Stat(filepath.Join(destDir, "php-cgi.exe")); err == nil {
			return nil
		}
	} else if rType == "nginx" {
		if _, err := os.Stat(filepath.Join(destDir, "nginx.exe")); err == nil {
			return nil
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil
	}

	tmpFile, err := os.CreateTemp("", "devenv-pkg-*.zip")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	_, copyErr := io.Copy(tmpFile, resp.Body)
	_ = tmpFile.Close()
	if copyErr != nil {
		return copyErr
	}

	return extractZipArchive(tmpPath, destDir)
}

func extractZipArchive(zipPath, destDir string) error {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer zr.Close()

	commonPrefix := ""
	for i, f := range zr.File {
		name := strings.TrimPrefix(filepath.ToSlash(f.Name), "/")
		parts := strings.SplitN(name, "/", 2)
		if len(parts) < 2 {
			commonPrefix = ""
			break
		}
		if i == 0 {
			commonPrefix = parts[0] + "/"
		} else if !strings.HasPrefix(name, commonPrefix) {
			commonPrefix = ""
			break
		}
	}

	for _, f := range zr.File {
		rel := strings.TrimPrefix(filepath.ToSlash(f.Name), "/")
		if commonPrefix != "" {
			rel = strings.TrimPrefix(rel, commonPrefix)
		}
		if rel == "" {
			continue
		}
		targetPath := filepath.Join(destDir, filepath.FromSlash(rel))
		if f.FileInfo().IsDir() {
			_ = os.MkdirAll(targetPath, 0755)
			continue
		}
		_ = os.MkdirAll(filepath.Dir(targetPath), 0755)
		rc, err := f.Open()
		if err != nil {
			continue
		}
		outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			_ = rc.Close()
			continue
		}
		_, _ = io.Copy(outFile, rc)
		_ = outFile.Close()
		_ = rc.Close()
	}
	return nil
}

func ensurePHPIniConfigured(phpDir string) {
	iniPath := filepath.Join(phpDir, "php.ini")
	if _, err := os.Stat(iniPath); err == nil {
		return
	}
	devIni := filepath.Join(phpDir, "php.ini-development")
	raw, err := os.ReadFile(devIni)
	if err != nil {
		return
	}
	content := string(raw)
	content = strings.ReplaceAll(content, `;extension_dir = "ext"`, `extension_dir = "ext"`)
	for _, ext := range []string{"curl", "fileinfo", "gd", "mbstring", "mysqli", "openssl", "pdo_mysql", "pdo_sqlite", "sqlite3"} {
		content = strings.ReplaceAll(content, ";extension="+ext, "extension="+ext)
	}
	_ = os.WriteFile(iniPath, []byte(content), 0644)
}

func (s *RuntimeService) SetDefault(ctx context.Context, id int64) error {
	list, err := s.runtimes.ListByType(ctx, runtime.TypePHP)
	if err != nil {
		return err
	}

	found := false
	now := time.Now().UTC()
	for _, cur := range list {
		var meta runtimeMeta
		if cur.MetadataJSON != "" {
			_ = json.Unmarshal([]byte(cur.MetadataJSON), &meta)
		}
		meta.Default = cur.ID == id
		if cur.ID == id {
			found = true
		}
		raw, err := json.Marshal(meta)
		if err != nil {
			return err
		}
		cur.MetadataJSON = string(raw)
		cur.UpdatedAt = now
		if err := s.runtimes.Update(ctx, cur); err != nil {
			return err
		}
	}
	if !found {
		return appErrors.New(appErrors.CodeNotFound, "运行时版本不存在")
	}
	s.bus.Emit("runtime:default-changed", id)
	return nil
}

func (s *RuntimeService) Extensions(ctx context.Context, id int64) ([]runtime.PHPExtension, error) {
	cur, err := s.runtimes.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	iniPath := filepath.Join(cur.InstallPath, "php.ini")
	enabledMap := make(map[string]bool)
	if file, err := os.Open(iniPath); err == nil {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "extension=") {
				ext := strings.TrimPrefix(line, "extension=")
				ext = strings.TrimPrefix(ext, "php_")
				ext = strings.TrimSuffix(ext, ".dll")
				enabledMap[ext] = true
			}
		}
		if err := file.Close(); err != nil {
			return nil, err
		}
	}

	common := []string{"curl", "fileinfo", "gd", "intl", "mbstring", "mysqli", "openssl", "pdo_mysql", "pdo_sqlite", "redis", "zip"}
	res := make([]runtime.PHPExtension, 0, len(common))
	for _, name := range common {
		res = append(res, runtime.PHPExtension{
			Name:      name,
			Enabled:   enabledMap[name],
			Available: true,
		})
	}
	return res, nil
}

func inferTypeFromPathOrVersion(installPath, version string) string {
	lowPath := strings.ToLower(filepath.ToSlash(installPath))
	for _, t := range []string{"mysql", "nginx", "redis", "postgresql", "apache", "nodejs", "php"} {
		if strings.Contains(lowPath, "/"+t+"/") || strings.HasPrefix(lowPath, t+"/") {
			return t
		}
	}
	if version == "8.4.11" || version == "8.0.48" {
		return "mysql"
	}
	if version == "1.26.3" || version == "1.24.0" {
		return "nginx"
	}
	return ""
}

func (s *RuntimeService) Scan(ctx context.Context) error {
	// 1. 修复旧数据中被误标为 php 的其他类型运行时记录
	if all, err := s.runtimes.ListAll(ctx); err == nil {
		for _, item := range all {
			inferred := inferTypeFromPathOrVersion(item.InstallPath, item.Version)
			if inferred != "" && inferred != string(item.RuntimeType) {
				item.RuntimeType = runtime.RuntimeType(inferred)
				item.UpdatedAt = time.Now().UTC()
				_ = s.runtimes.Update(ctx, item)
			}
		}
	}

	// 2. 扫描本地 runtimes/<type>/<dir> 目录自动注册已解压的运行时
	if typeEntries, err := os.ReadDir("runtimes"); err == nil {
		for _, te := range typeEntries {
			if !te.IsDir() {
				continue
			}
			rType := strings.ToLower(te.Name())
			subDir := filepath.Join("runtimes", te.Name())
			verEntries, err := os.ReadDir(subDir)
			if err != nil {
				continue
			}
			for _, ve := range verEntries {
				if !ve.IsDir() {
					continue
				}
				fullDir := filepath.Join(subDir, ve.Name())
				dirName := ve.Name()
				ver := dirName
				variant := ""
				arch := "x64"
				parts := strings.Split(dirName, "-")
				if len(parts) >= 2 {
					ver = parts[0]
					arch = parts[len(parts)-1]
					if len(parts) == 3 {
						variant = parts[1]
					}
				}
				if rType == "php" {
					ensurePHPIniConfigured(fullDir)
				}
				_, _ = s.Register(ctx, rType, ver, variant, arch, fullDir)
			}
		}
	}

	cfg, err := s.settings.Get(ctx)
	if err != nil || cfg == nil || strings.TrimSpace(cfg.RuntimeRootPath) == "" {
		return nil
	}

	entries, err := os.ReadDir(cfg.RuntimeRootPath)
	if err != nil {
		return nil
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dir := filepath.Join(cfg.RuntimeRootPath, entry.Name())
		exe := filepath.Join(dir, "php.exe")
		if _, err := os.Stat(exe); err == nil {
			ensurePHPIniConfigured(dir)
			_, _ = s.Register(ctx, string(runtime.TypePHP), entry.Name(), "nts", "x64", dir)
		}
	}
	return nil
}

func (s *RuntimeService) Delete(ctx context.Context, id int64, purge bool) error {
	cur, err := s.runtimes.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 1. 检查是否被网站直接或通过服务实例引用
	sites, err := s.sites.List(ctx)
	if err != nil {
		return err
	}
	srvs, err := s.services.ListAll(ctx)
	if err != nil {
		return err
	}
	srvUsingThis := make(map[int64]bool)
	for _, srv := range srvs {
		if srv.RuntimeID == id {
			srvUsingThis[srv.ID] = true
		}
	}
	for _, st := range sites {
		if (st.RuntimeID != nil && *st.RuntimeID == id) || (st.ServiceID != nil && srvUsingThis[*st.ServiceID]) {
			return appErrors.New(appErrors.CodeInvalidArgument, "该版本已被网站引用，请先在「网站管理」中解绑")
		}
	}

	// 2. 若有关联服务实例，检查运行状态与同类型剩余版本数量
	if len(srvUsingThis) > 0 {
		if s.provisioner != nil {
			running, err := s.provisioner.IsRuntimeRunning(ctx, id)
			if err != nil {
				return err
			}
			if running {
				return appErrors.New(appErrors.CodeServiceRunning, "关联的服务实例正在运行，请先停止服务")
			}
		}

		sameType, err := s.runtimes.ListByType(ctx, cur.RuntimeType)
		if err != nil {
			return err
		}
		if len(sameType) > 1 {
			return appErrors.New(appErrors.CodeInvalidArgument, "当前版本正被服务实例使用，请先在「服务实例」页切换到其他版本后再卸载")
		}

		// 3. 这是同类型最后一个版本，且服务已停止、无网站引用：先删除服务实例再删除运行环境
		if s.provisioner != nil {
			if err := s.provisioner.OnRuntimeRemoved(ctx, *cur); err != nil {
				return err
			}
		}
	}

	if err := s.runtimes.Delete(ctx, id); err != nil {
		return err
	}

	if purge && strings.TrimSpace(cur.InstallPath) != "" {
		_ = os.RemoveAll(cur.InstallPath)
	}

	s.bus.Emit("runtime:deleted", id)
	return nil
}

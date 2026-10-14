package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"devenv/internal/domain/host"
	"devenv/internal/ports"
	appErrors "devenv/internal/shared/errors"
)

type HostService struct {
	repo        ports.HostRepository
	projectRepo ports.ProjectRepository
	publisher   ports.EventPublisher
}

func NewHostService(repo ports.HostRepository, projectRepo ports.ProjectRepository, publisher ports.EventPublisher) *HostService {
	return &HostService{
		repo:        repo,
		projectRepo: projectRepo,
		publisher:   publisher,
	}
}

func (s *HostService) ValidateVirtualHost(ctx context.Context, in host.CreateInput) host.ValidationResult {
	res := host.ValidationResult{
		Errors:   make([]string, 0),
		Warnings: make([]string, 0),
	}

	domain := strings.TrimSpace(in.Domain)
	if domain == "" {
		res.Errors = append(res.Errors, "域名不能为空")
	} else if strings.Contains(domain, " ") || !strings.Contains(domain, ".") {
		res.Warnings = append(res.Warnings, "建议使用有效的本地测试域名，如 app.local 或 test.test")
	}

	if in.Port <= 0 || in.Port > 65535 {
		res.Errors = append(res.Errors, "端口必须在 1 到 65535 之间")
	}

	docRoot := filepath.Clean(strings.TrimSpace(in.DocumentRoot))
	if docRoot == "" {
		res.Errors = append(res.Errors, "网站根目录不能为空")
	} else {
		if info, err := os.Stat(docRoot); err != nil || !info.IsDir() {
			res.Errors = append(res.Errors, "网站根目录不存在或不是有效目录")
		}
	}

	res.Valid = len(res.Errors) == 0
	return res
}

func (s *HostService) ListVirtualHosts(ctx context.Context) ([]host.VirtualHost, error) {
	return s.repo.ListAll(ctx)
}

func (s *HostService) GetVirtualHost(ctx context.Context, id int64) (*host.VirtualHost, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *HostService) CreateVirtualHost(ctx context.Context, in host.CreateInput) (*host.VirtualHost, error) {
	val := s.ValidateVirtualHost(ctx, in)
	if !val.Valid {
		return nil, appErrors.New(appErrors.CodeValidationFailed, strings.Join(val.Errors, "; "))
	}

	// Verify project exists
	p, err := s.projectRepo.GetByID(ctx, in.ProjectID)
	if err != nil || p == nil {
		return nil, appErrors.New(appErrors.CodeNotFound, "关联的项目不存在")
	}

	domain := strings.ToLower(strings.TrimSpace(in.Domain))
	existing, err := s.repo.GetByDomain(ctx, domain)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, appErrors.New(appErrors.CodeAlreadyExists, fmt.Sprintf("域名 %s 已被占用", domain))
	}

	now := time.Now().UTC()
	vh := host.VirtualHost{
		ProjectID:     in.ProjectID,
		Domain:        domain,
		DocumentRoot:  filepath.Clean(in.DocumentRoot),
		WebServerType: in.WebServerType,
		Port:          in.Port,
		Enabled:       true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	id, err := s.repo.Create(ctx, vh)
	if err != nil {
		return nil, err
	}
	vh.ID = id
	return &vh, nil
}

func (s *HostService) UpdateVirtualHost(ctx context.Context, item host.VirtualHost) (*host.VirtualHost, error) {
	existing, err := s.repo.GetByID(ctx, item.ID)
	if err != nil || existing == nil {
		return nil, appErrors.New(appErrors.CodeNotFound, "虚拟主机不存在")
	}

	existing.Domain = strings.ToLower(strings.TrimSpace(item.Domain))
	existing.DocumentRoot = filepath.Clean(item.DocumentRoot)
	existing.Port = item.Port
	existing.WebServerType = item.WebServerType
	existing.Enabled = item.Enabled
	existing.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, *existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *HostService) DeleteVirtualHost(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *HostService) EnableVirtualHost(ctx context.Context, id int64) error {
	vh, err := s.repo.GetByID(ctx, id)
	if err != nil || vh == nil {
		return appErrors.New(appErrors.CodeNotFound, "虚拟主机不存在")
	}
	vh.Enabled = true
	vh.UpdatedAt = time.Now().UTC()
	return s.repo.Update(ctx, *vh)
}

func (s *HostService) DisableVirtualHost(ctx context.Context, id int64) error {
	vh, err := s.repo.GetByID(ctx, id)
	if err != nil || vh == nil {
		return appErrors.New(appErrors.CodeNotFound, "虚拟主机不存在")
	}
	vh.Enabled = false
	vh.UpdatedAt = time.Now().UTC()
	return s.repo.Update(ctx, *vh)
}

package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"devenv/internal/domain/project"
	"devenv/internal/ports"
	appErrors "devenv/internal/shared/errors"
)

type ProjectService struct {
	repo        ports.ProjectRepository
	runtimeRepo ports.RuntimeRepository
	publisher   ports.EventPublisher
}

func NewProjectService(repo ports.ProjectRepository, runtimeRepo ports.RuntimeRepository, publisher ports.EventPublisher) *ProjectService {
	return &ProjectService{
		repo:        repo,
		runtimeRepo: runtimeRepo,
		publisher:   publisher,
	}
}

func (s *ProjectService) ValidateProjectPath(ctx context.Context, rawPath string) project.PathValidation {
	rawPath = strings.TrimSpace(rawPath)
	res := project.PathValidation{
		Errors: make([]string, 0),
	}
	if rawPath == "" {
		res.Valid = false
		res.Errors = append(res.Errors, "路径不能为空")
		return res
	}

	cleanPath := filepath.Clean(rawPath)
	res.NormalizedPath = cleanPath

	info, err := os.Stat(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			res.Exists = false
			res.Writable = false
			// Path does not exist yet
			res.Valid = true
			return res
		}
		res.Valid = false
		res.Errors = append(res.Errors, "无法访问路径: "+err.Error())
		return res
	}

	res.Exists = true
	if !info.IsDir() {
		res.Valid = false
		res.Errors = append(res.Errors, "目标路径不是文件夹")
		return res
	}

	// Test writable
	testFile := filepath.Join(cleanPath, ".devenv_write_test")
	if f, err := os.Create(testFile); err == nil {
		_ = f.Close()
		_ = os.Remove(testFile)
		res.Writable = true
	} else {
		res.Writable = false
		res.Errors = append(res.Errors, "目录无写权限")
	}

	res.Valid = len(res.Errors) == 0
	return res
}

func (s *ProjectService) CreateProject(ctx context.Context, in project.CreateInput) (*project.Project, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, appErrors.New(appErrors.CodeInvalidArgument, "项目名称不能为空")
	}

	val := s.ValidateProjectPath(ctx, in.Path)
	if !val.Valid {
		return nil, appErrors.New(appErrors.CodePathInvalid, strings.Join(val.Errors, "; "))
	}

	// Make directory if not exists
	if !val.Exists {
		if err := os.MkdirAll(val.NormalizedPath, 0755); err != nil {
			return nil, appErrors.Wrap(err, appErrors.CodePathNotAccessible, "无法创建项目目录")
		}
	}

	pathKey := strings.ToLower(val.NormalizedPath)
	existing, err := s.repo.GetByPathKey(ctx, pathKey)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, appErrors.New(appErrors.CodeAlreadyExists, "该路径已被其他项目登记")
	}

	if in.PHPVersionID != nil {
		rt, err := s.runtimeRepo.GetByID(ctx, *in.PHPVersionID)
		if err != nil || rt == nil {
			return nil, appErrors.New(appErrors.CodeNotFound, "指定的 PHP 版本不存在")
		}
	}

	now := time.Now().UTC()
	item := project.Project{
		Name:         name,
		Path:         val.NormalizedPath,
		PathKey:      pathKey,
		PHPVersionID: in.PHPVersionID,
		Status:       project.StatusReady,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	id, err := s.repo.Create(ctx, item)
	if err != nil {
		return nil, err
	}
	item.ID = id

	if s.publisher != nil {
		s.publisher.Emit("project:changed", item)
	}

	return &item, nil
}

func (s *ProjectService) ImportProject(ctx context.Context, path string) (*project.Project, error) {
	val := s.ValidateProjectPath(ctx, path)
	if !val.Valid || !val.Exists {
		return nil, appErrors.New(appErrors.CodePathInvalid, "项目目录不存在或不可访问")
	}

	name := filepath.Base(val.NormalizedPath)
	return s.CreateProject(ctx, project.CreateInput{
		Name: name,
		Path: val.NormalizedPath,
	})
}

func (s *ProjectService) ListProjects(ctx context.Context) ([]project.Project, error) {
	list, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	// Update status dynamically if missing
	for i := range list {
		if _, err := os.Stat(list[i].Path); err != nil {
			list[i].Status = project.StatusMissing
		} else {
			list[i].Status = project.StatusReady
		}
	}
	return list, nil
}

func (s *ProjectService) GetProject(ctx context.Context, id int64) (*project.Project, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(p.Path); err != nil {
		p.Status = project.StatusMissing
	} else {
		p.Status = project.StatusReady
	}
	return p, nil
}

func (s *ProjectService) UpdateProject(ctx context.Context, in project.UpdateInput) (*project.Project, error) {
	existing, err := s.repo.GetByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}

	if in.Name != nil {
		name := strings.TrimSpace(*in.Name)
		if name == "" {
			return nil, appErrors.New(appErrors.CodeInvalidArgument, "项目名称不能为空")
		}
		existing.Name = name
	}

	if in.PHPVersionID != nil {
		rt, err := s.runtimeRepo.GetByID(ctx, *in.PHPVersionID)
		if err != nil || rt == nil {
			return nil, appErrors.New(appErrors.CodeNotFound, "指定的 PHP 版本不存在")
		}
		existing.PHPVersionID = in.PHPVersionID
	}

	existing.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, *existing); err != nil {
		return nil, err
	}

	if s.publisher != nil {
		s.publisher.Emit("project:changed", existing)
	}

	return existing, nil
}

func (s *ProjectService) DeleteProject(ctx context.Context, id int64) error {
	// Only delete metadata records, never delete physical folder
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	if s.publisher != nil {
		s.publisher.Emit("project:changed", map[string]interface{}{"deletedId": id})
	}
	return nil
}

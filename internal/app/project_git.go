package app

import (
	"context"
	"sync"

	"devenv/internal/domain/project"
	"devenv/internal/ports"
	appErrors "devenv/internal/shared/errors"
)

const GIT_STATUS_WORKERS = 4

func (s *ProjectService) GitStatus(ctx context.Context, id int64) (ports.ProjectGitStatus, error) {
	item, err := s.projects.Get(ctx, id)
	if err != nil {
		return ports.ProjectGitStatus{}, err
	}
	if item == nil {
		return ports.ProjectGitStatus{}, appErrors.New(appErrors.NOT_FOUND, "项目不存在")
	}
	return s.projectGitStatus(ctx, *item), nil
}

func (s *ProjectService) GitStatuses(ctx context.Context) ([]ports.ProjectGitStatus, error) {
	items, err := s.projects.List(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]ports.ProjectGitStatus, len(items))
	workers := GIT_STATUS_WORKERS
	if len(items) < workers {
		workers = len(items)
	}
	if workers == 0 {
		return result, nil
	}
	queue := make(chan int)
	var group sync.WaitGroup
	for worker := 0; worker < workers; worker++ {
		group.Add(1)
		go func() {
			defer group.Done()
			for index := range queue {
				result[index] = s.projectGitStatus(ctx, items[index])
			}
		}()
	}
	for index := range items {
		queue <- index
	}
	close(queue)
	group.Wait()
	return result, nil
}

func (s *ProjectService) InitializeGit(ctx context.Context, id int64) error {
	item, err := s.projects.Get(ctx, id)
	if err != nil {
		return err
	}
	if item == nil {
		return appErrors.New(appErrors.NOT_FOUND, "项目不存在")
	}
	return s.git.Initialize(ctx, item.Root)
}

func (s *ProjectService) CommitGit(ctx context.Context, id int64, message string, bumpVersion bool) error {
	item, err := s.projects.Get(ctx, id)
	if err != nil {
		return err
	}
	if item == nil {
		return appErrors.New(appErrors.NOT_FOUND, "项目不存在")
	}
	return s.git.Commit(ctx, item.Root, item.Kind, message, bumpVersion)
}

func (s *ProjectService) projectGitStatus(ctx context.Context, item project.Project) ports.ProjectGitStatus {
	status, err := s.git.Status(ctx, item.Root, item.Kind)
	if err != nil {
		status.Error = err.Error()
		status.Files = make([]ports.GitFileStatus, 0)
	}
	return ports.ProjectGitStatus{
		ProjectID:   item.ID,
		ProjectName: item.Name,
		GitStatus:   status,
	}
}

func (s *ProjectService) SaveGitRemote(ctx context.Context, id int64, remote ports.GitRemote) error {
	return s.withGitProject(ctx, id, func(root string) error { return s.git.SaveRemote(ctx, root, remote) })
}

func (s *ProjectService) RemoveGitRemote(ctx context.Context, id int64, name string) error {
	return s.withGitProject(ctx, id, func(root string) error { return s.git.RemoveRemote(ctx, root, name) })
}

func (s *ProjectService) FetchGit(ctx context.Context, id int64, name string) error {
	return s.withGitProject(ctx, id, func(root string) error { return s.git.Fetch(ctx, root, name) })
}

func (s *ProjectService) PushGit(ctx context.Context, id int64, name string) error {
	return s.withGitProject(ctx, id, func(root string) error { return s.git.Push(ctx, root, name) })
}

func (s *ProjectService) withGitProject(ctx context.Context, id int64, action func(string) error) error {
	item, err := s.projects.Get(ctx, id)
	if err != nil {
		return err
	}
	if item == nil {
		return appErrors.New(appErrors.NOT_FOUND, "项目不存在")
	}
	return action(item.Root)
}

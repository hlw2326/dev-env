package app

import (
	"context"
	"time"

	"devenv/internal/domain/task"
	"devenv/internal/ports"
	appErrors "devenv/internal/shared/errors"
)

type TaskService struct {
	repo      ports.TaskRepository
	publisher ports.EventPublisher
}

func NewTaskService(repo ports.TaskRepository, publisher ports.EventPublisher) *TaskService {
	return &TaskService{
		repo:      repo,
		publisher: publisher,
	}
}

func (s *TaskService) GetTask(ctx context.Context, id int64) (*task.Task, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *TaskService) ListTasks(ctx context.Context) ([]task.Task, error) {
	return s.repo.ListAll(ctx)
}

func (s *TaskService) CancelTask(ctx context.Context, id int64) error {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if t.Status == task.StatusSucceeded || t.Status == task.StatusFailed || t.Status == task.StatusCancelled {
		return appErrors.New(appErrors.CodeTaskNotCancellable, "任务已结束，无法取消")
	}

	now := time.Now().UTC()
	t.Status = task.StatusCancelled
	t.FinishedAt = &now
	t.UpdatedAt = now
	t.Message = "任务已由用户取消"

	if err := s.repo.Update(ctx, *t); err != nil {
		return err
	}

	if s.publisher != nil {
		s.publisher.Emit("task:cancelled", t)
	}

	return nil
}

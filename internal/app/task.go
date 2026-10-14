package app

import (
	"context"
	"time"

	"devenv/internal/domain/task"
	"devenv/internal/ports"
)

type TaskService struct {
	tasks ports.TaskRepository
	bus   ports.EventPublisher
}

func NewTaskService(
	tasks ports.TaskRepository,
	bus ports.EventPublisher,
) *TaskService {
	return &TaskService{
		tasks: tasks,
		bus:   bus,
	}
}

func (s *TaskService) List(ctx context.Context) ([]task.Task, error) {
	return s.tasks.ListAll(ctx)
}

func (s *TaskService) Get(ctx context.Context, id int64) (*task.Task, error) {
	return s.tasks.GetByID(ctx, id)
}

func (s *TaskService) Cancel(ctx context.Context, id int64) error {
	cur, err := s.tasks.GetByID(ctx, id)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	cur.Status = task.StatusCancelled
	cur.Message = "Cancelled by user"
	cur.UpdatedAt = now
	cur.FinishedAt = &now
	if err := s.tasks.Update(ctx, *cur); err != nil {
		return err
	}
	s.bus.Emit("task:cancelled", id)
	return nil
}

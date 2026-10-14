package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"devenv/internal/domain/task"
	appErrors "devenv/internal/shared/errors"
)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(ctx context.Context, item task.Task) (int64, error) {
	query := `
		INSERT INTO system_tasks (task_type, status, progress, message, error_code, error_message, payload_json, created_at, updated_at, finished_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	createdAt := item.CreatedAt.UTC().Format(time.RFC3339)
	updatedAt := item.UpdatedAt.UTC().Format(time.RFC3339)
	var finishedAt *string
	if item.FinishedAt != nil {
		str := item.FinishedAt.UTC().Format(time.RFC3339)
		finishedAt = &str
	}

	res, err := r.db.ExecContext(ctx, query,
		item.TaskType, string(item.Status), item.Progress, item.Message, item.ErrorCode, item.ErrorMessage, item.PayloadJSON, createdAt, updatedAt, finishedAt,
	)
	if err != nil {
		return 0, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to insert task")
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to get task insert id")
	}
	return id, nil
}

func (r *TaskRepository) GetByID(ctx context.Context, id int64) (*task.Task, error) {
	query := `
		SELECT id, task_type, status, progress, message, error_code, error_message, payload_json, created_at, updated_at, finished_at
		FROM system_tasks WHERE id = ?
	`
	var t task.Task
	var statusStr, msg, payload, createdStr, updatedStr string
	var errCode, errMsg, finishedStr sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID, &t.TaskType, &statusStr, &t.Progress, &msg, &errCode, &errMsg, &payload, &createdStr, &updatedStr, &finishedStr,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, appErrors.New(appErrors.CodeNotFound, "task not found")
	}
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to query task")
	}

	t.Status = task.Status(statusStr)
	t.Message = msg
	t.PayloadJSON = payload
	if errCode.Valid {
		t.ErrorCode = &errCode.String
	}
	if errMsg.Valid {
		t.ErrorMessage = &errMsg.String
	}
	createdAt, err := time.Parse(time.RFC3339, createdStr)
	if err != nil {
		return nil, err
	}
	updatedAt, err := time.Parse(time.RFC3339, updatedStr)
	if err != nil {
		return nil, err
	}
	t.CreatedAt = createdAt
	t.UpdatedAt = updatedAt
	if finishedStr.Valid {
		fTime, err := time.Parse(time.RFC3339, finishedStr.String)
		if err == nil {
			t.FinishedAt = &fTime
		}
	}
	return &t, nil
}

func (r *TaskRepository) ListAll(ctx context.Context) ([]task.Task, error) {
	query := `
		SELECT id, task_type, status, progress, message, error_code, error_message, payload_json, created_at, updated_at, finished_at
		FROM system_tasks ORDER BY created_at DESC LIMIT 50
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to list tasks")
	}
	defer rows.Close()

	list := make([]task.Task, 0)
	for rows.Next() {
		var t task.Task
		var statusStr, msg, payload, createdStr, updatedStr string
		var errCode, errMsg, finishedStr sql.NullString

		if err := rows.Scan(
			&t.ID, &t.TaskType, &statusStr, &t.Progress, &msg, &errCode, &errMsg, &payload, &createdStr, &updatedStr, &finishedStr,
		); err != nil {
			return nil, appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to scan task")
		}

		t.Status = task.Status(statusStr)
		t.Message = msg
		t.PayloadJSON = payload
		if errCode.Valid {
			t.ErrorCode = &errCode.String
		}
		if errMsg.Valid {
			t.ErrorMessage = &errMsg.String
		}
		createdAt, err := time.Parse(time.RFC3339, createdStr)
		if err != nil {
			return nil, err
		}
		updatedAt, err := time.Parse(time.RFC3339, updatedStr)
		if err != nil {
			return nil, err
		}
		t.CreatedAt = createdAt
		t.UpdatedAt = updatedAt
		if finishedStr.Valid {
			fTime, err := time.Parse(time.RFC3339, finishedStr.String)
			if err == nil {
				t.FinishedAt = &fTime
			}
		}
		list = append(list, t)
	}
	return list, nil
}

func (r *TaskRepository) Update(ctx context.Context, item task.Task) error {
	query := `
		UPDATE system_tasks
		SET status = ?, progress = ?, message = ?, error_code = ?, error_message = ?, updated_at = ?, finished_at = ?
		WHERE id = ?
	`
	updatedAt := item.UpdatedAt.UTC().Format(time.RFC3339)
	var finishedAt *string
	if item.FinishedAt != nil {
		str := item.FinishedAt.UTC().Format(time.RFC3339)
		finishedAt = &str
	}

	res, err := r.db.ExecContext(ctx, query,
		string(item.Status), item.Progress, item.Message, item.ErrorCode, item.ErrorMessage, updatedAt, finishedAt, item.ID,
	)
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to update task")
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeDatabaseError, "failed to get rows affected")
	}
	if affected == 0 {
		return appErrors.New(appErrors.CodeNotFound, "task not found for update")
	}
	return nil
}

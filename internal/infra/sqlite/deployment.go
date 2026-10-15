package sqlite

import (
	"context"
	"database/sql"
	"devenv/internal/domain/deployment"
	"encoding/json"
	"errors"
	"time"
)

type DeploymentRepository struct{ db *sql.DB }

func NewDeploymentRepository(db *sql.DB) *DeploymentRepository { return &DeploymentRepository{db: db} }
func (r *DeploymentRepository) Load(ctx context.Context, id int64) (*deployment.Config, error) {
	var raw string
	err := r.db.QueryRowContext(ctx, `SELECT config FROM system_project_deployment WHERE project=?`, id).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var config deployment.Config
	if err := json.Unmarshal([]byte(raw), &config); err != nil {
		return nil, err
	}
	return &config, nil
}
func (r *DeploymentRepository) Save(ctx context.Context, config deployment.Config) error {
	raw, err := json.Marshal(config)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO system_project_deployment(project,config,updated) VALUES(?,?,?) ON CONFLICT(project) DO UPDATE SET config=excluded.config,updated=excluded.updated`, config.ProjectID, string(raw), time.Now().UTC().Format(time.RFC3339))
	return err
}
func (r *DeploymentRepository) Record(ctx context.Context, item deployment.Task) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO system_project_deployment_log(id,project,provider,state,message,target,completed,total,created,finished) VALUES(?,?,?,?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET state=excluded.state,message=excluded.message,completed=excluded.completed,finished=excluded.finished`, item.ID, item.ProjectID, item.Provider, item.State, item.Message, item.Target, item.Completed, item.Total, item.Created, item.Finished)
	return err
}
func (r *DeploymentRepository) History(ctx context.Context, id int64) ([]deployment.Task, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,project,provider,state,message,target,completed,total,created,finished FROM system_project_deployment_log WHERE project=? ORDER BY created DESC,rowid DESC LIMIT 20`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]deployment.Task, 0)
	for rows.Next() {
		var item deployment.Task
		if err := rows.Scan(&item.ID, &item.ProjectID, &item.Provider, &item.State, &item.Message, &item.Target, &item.Completed, &item.Total, &item.Created, &item.Finished); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
func (r *DeploymentRepository) Interrupt(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `UPDATE system_project_deployment_log SET state=?,message=?,finished=? WHERE state=?`, deployment.STATE_INTERRUPTED, "上次发布随应用退出而中断，结果未确认，请核对服务器或微信后台；不会自动重试", time.Now().UTC().Format(time.RFC3339), deployment.STATE_RUNNING)
	return err
}

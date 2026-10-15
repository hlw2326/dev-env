package ports

import (
	"context"
	"devenv/internal/domain/deployment"
)

type DeploymentRepository interface {
	Load(context.Context, int64) (*deployment.Config, error)
	Save(context.Context, deployment.Config) error
	Record(context.Context, deployment.Task) error
	History(context.Context, int64) ([]deployment.Task, error)
	Interrupt(context.Context) error
}

// Providers are explicitly invoked; no deployment operation participates in application startup.
type DeploymentProvider interface {
	Preview(context.Context, deployment.Config, string) (deployment.Plan, error)
	Test(context.Context, deployment.Config, string, deployment.Credentials) error
	Run(context.Context, deployment.Config, string, deployment.Plan, deployment.Credentials, func(int, string)) error
}

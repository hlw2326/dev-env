package project

import "time"

type Status string

const (
	StatusReady   Status = "ready"
	StatusMissing Status = "missing"
	StatusInvalid Status = "invalid"
)

type Project struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	PathKey      string    `json:"pathKey"`
	PHPVersionID *int64    `json:"phpVersionId"`
	Status       Status    `json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type CreateInput struct {
	Name         string `json:"name"`
	Path         string `json:"path"`
	PHPVersionID *int64 `json:"phpVersionId"`
}

type UpdateInput struct {
	ID           int64   `json:"id"`
	Name         *string `json:"name"`
	PHPVersionID *int64  `json:"phpVersionId"`
}

type PathValidation struct {
	Valid          bool     `json:"valid"`
	NormalizedPath string   `json:"normalizedPath"`
	Exists         bool     `json:"exists"`
	Writable       bool     `json:"writable"`
	Errors         []string `json:"errors"`
}

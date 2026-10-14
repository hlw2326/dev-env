package host

import "time"

type VirtualHost struct {
	ID            int64     `json:"id"`
	ProjectID     int64     `json:"projectId"`
	Domain        string    `json:"domain"`
	DocumentRoot  string    `json:"documentRoot"`
	WebServerType string    `json:"webServerType"`
	Port          int       `json:"port"`
	Enabled       bool      `json:"enabled"`
	ConfigPath    *string   `json:"configPath"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type CreateInput struct {
	ProjectID     int64  `json:"projectId"`
	Domain        string `json:"domain"`
	DocumentRoot  string `json:"documentRoot"`
	WebServerType string `json:"webServerType"`
	Port          int    `json:"port"`
}

type ValidationResult struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
}

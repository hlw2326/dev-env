package site

import (
	"errors"
	"strings"
	"time"
)

type Kind string

const (
	Static Kind = "static"
	PHP    Kind = "php"
	Node   Kind = "nodejs"
	Go     Kind = "go"
	Python Kind = "python"
)

type Status string

const (
	Stopped Status = "stopped"
	Running Status = "running"
	Errored Status = "error"
)

type Preset struct {
	Port  int
	Entry string
	Cmd   string
	Proxy bool
}

var Presets = map[Kind]Preset{
	Static: {Port: 0, Entry: "index.html", Cmd: "", Proxy: false},
	PHP:    {Port: 9000, Entry: "index.php", Cmd: "", Proxy: false},
	Node:   {Port: 3000, Entry: "", Cmd: "npm start", Proxy: true},
	Go:     {Port: 8080, Entry: "", Cmd: "./main.exe", Proxy: true},
	Python: {Port: 8000, Entry: "", Cmd: "app:application", Proxy: true},
}

type Snapshot struct {
	Time    string `json:"time"`
	Summary string `json:"summary"`
	Content string `json:"content"`
}

type Meta struct {
	Entry     string            `json:"entry"`
	Cmd       string            `json:"cmd"`
	AutoStart bool              `json:"autoStart"`
	Env       map[string]string `json:"env"`
	Custom    string            `json:"custom"`
	AccessLog string            `json:"accessLog"`
	ErrorLog  string            `json:"errorLog"`
	History   []Snapshot        `json:"history"`
}

func (m *Meta) Push(summary, content string) {
	snap := Snapshot{
		Time:    time.Now().Local().Format("2006-01-02 15:04:05"),
		Summary: summary,
		Content: content,
	}
	m.History = append([]Snapshot{snap}, m.History...)
	if len(m.History) > 15 {
		m.History = m.History[:15]
	}
}

type Site struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Kind      Kind      `json:"kind"`
	Domains   []string  `json:"domains"`
	Root      string    `json:"root"`
	RuntimeID *int64    `json:"runtimeId"`
	ServiceID *int64    `json:"serviceId"`
	Git       bool      `json:"git"`
	Remote    string    `json:"remote"`
	Branch    string    `json:"branch"`
	Commit    string    `json:"commit"`
	PulledAt  string    `json:"pulledAt"`
	Conf      string    `json:"conf"`
	Port      int       `json:"port"`
	TLS       bool      `json:"tls"`
	Cert      string    `json:"cert"`
	Key       string    `json:"key"`
	Issuer    string    `json:"issuer"`
	Proxy     bool      `json:"proxy"`
	Upstream  *int      `json:"upstream"`
	Meta      Meta      `json:"meta"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type CreateInput struct {
	Name        string            `json:"name"`
	Kind        Kind              `json:"kind"`
	Domains     []string          `json:"domains"`
	Root        string            `json:"root"`
	RuntimeID   *int64            `json:"runtimeId"`
	ServiceID   *int64            `json:"serviceId"`
	AutoService bool              `json:"autoService"`
	Git         bool              `json:"git"`
	Remote      string            `json:"remote"`
	Branch      string            `json:"branch"`
	Port        int               `json:"port"`
	TLS         bool              `json:"tls"`
	Issuer      string            `json:"issuer"`
	Cert        string            `json:"cert"`
	Key         string            `json:"key"`
	Upstream    *int              `json:"upstream"`
	Entry       string            `json:"entry"`
	Cmd         string            `json:"cmd"`
	AutoStart   bool              `json:"autoStart"`
	StartNow    bool              `json:"startNow"`
	Custom      string            `json:"custom"`
	Env         map[string]string `json:"env"`
}

func (in *CreateInput) Check() error {
	if strings.TrimSpace(in.Name) == "" {
		return errors.New("网站名称不能为空")
	}
	if _, ok := Presets[in.Kind]; !ok {
		return errors.New("不支持的网站类型")
	}
	clean := make([]string, 0, len(in.Domains))
	for _, d := range in.Domains {
		trimmed := strings.ToLower(strings.TrimSpace(d))
		if trimmed != "" {
			clean = append(clean, trimmed)
		}
	}
	if len(clean) == 0 {
		return errors.New("至少需要绑定一个域名")
	}
	in.Domains = clean
	if strings.TrimSpace(in.Root) == "" {
		return errors.New("网站根目录不能为空")
	}
	if in.Port <= 0 || in.Port > 65535 {
		return errors.New("监听端口必须在 1 到 65535 之间")
	}
	return nil
}

type UpdateInput struct {
	ID        int64    `json:"id"`
	Name      *string  `json:"name,omitempty"`
	Domains   []string `json:"domains,omitempty"`
	Root      *string  `json:"root,omitempty"`
	RuntimeID *int64   `json:"runtimeId,omitempty"`
	ServiceID *int64   `json:"serviceId,omitempty"`
	Port      *int     `json:"port,omitempty"`
	TLS       *bool    `json:"tls,omitempty"`
	Issuer    *string  `json:"issuer,omitempty"`
	Cert      *string  `json:"cert,omitempty"`
	Key       *string  `json:"key,omitempty"`
	Upstream  *int     `json:"upstream,omitempty"`
	Entry     *string  `json:"entry,omitempty"`
	Cmd       *string  `json:"cmd,omitempty"`
	Custom    *string  `json:"custom,omitempty"`
	AutoStart *bool    `json:"autoStart,omitempty"`
}

type GitState struct {
	Branch    string   `json:"branch"`
	Commit    string   `json:"commit"`
	Message   string   `json:"message"`
	Clean     bool     `json:"clean"`
	Ahead     int      `json:"ahead"`
	Behind    int      `json:"behind"`
	Modified  []string `json:"modified"`
	Untracked []string `json:"untracked"`
	CheckedAt string   `json:"checkedAt"`
}

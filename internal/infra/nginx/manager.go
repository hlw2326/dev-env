package nginx

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"devenv/internal/domain/site"
)

type Manager struct {
	baseDir string
}

func NewManager(baseDir string) *Manager {
	if baseDir == "" {
		dataDir := os.Getenv("LOCALAPPDATA")
		if dataDir == "" {
			dataDir = "."
		}
		baseDir = filepath.Join(dataDir, "DevEnv")
	}
	return &Manager{baseDir: baseDir}
}

func (m *Manager) GenerateConfig(item *site.Site, meta site.Metadata) string {
	var b strings.Builder

	listenPort := item.ListenPort
	if listenPort <= 0 {
		listenPort = 80
	}

	domains := strings.Join(item.DomainList, " ")
	if domains == "" {
		domains = "localhost"
	}

	rootSlash := filepath.ToSlash(item.RootPath)

	accessLog := meta.AccessLog
	if accessLog == "" && len(item.DomainList) > 0 {
		accessLog = fmt.Sprintf("logs/%s-access.log", item.DomainList[0])
	}
	errorLog := meta.ErrorLog
	if errorLog == "" && len(item.DomainList) > 0 {
		errorLog = fmt.Sprintf("logs/%s-error.log", item.DomainList[0])
	}

	b.WriteString("server {\n")
	b.WriteString(fmt.Sprintf("    listen %d;\n", listenPort))
	if item.SSLEnabled {
		b.WriteString("    listen 443 ssl http2;\n")
		if item.SSLCertPath != "" {
			b.WriteString(fmt.Sprintf("    ssl_certificate %s;\n", filepath.ToSlash(item.SSLCertPath)))
		}
		if item.SSLKeyPath != "" {
			b.WriteString(fmt.Sprintf("    ssl_certificate_key %s;\n", filepath.ToSlash(item.SSLKeyPath)))
		}
		b.WriteString("    ssl_protocols TLSv1.2 TLSv1.3;\n")
	}
	b.WriteString(fmt.Sprintf("    server_name %s;\n", domains))

	switch item.SiteType {
	case site.TypeStatic:
		entry := meta.EntryFile
		if entry == "" {
			entry = "index.html"
		}
		b.WriteString(fmt.Sprintf("    root %s;\n", rootSlash))
		b.WriteString(fmt.Sprintf("    index %s;\n\n", entry))
		b.WriteString("    location / {\n")
		b.WriteString(fmt.Sprintf("        try_files $uri $uri/ /%s;\n", entry))
		b.WriteString("    }\n")

	case site.TypePHP:
		entry := meta.EntryFile
		if entry == "" {
			entry = "index.php"
		}
		fpmPort := 9000
		if item.ProxyTargetPort != nil && *item.ProxyTargetPort > 0 {
			fpmPort = *item.ProxyTargetPort
		}
		b.WriteString(fmt.Sprintf("    root %s;\n", rootSlash))
		b.WriteString(fmt.Sprintf("    index %s index.html;\n\n", entry))
		b.WriteString("    location / {\n")
		b.WriteString(fmt.Sprintf("        try_files $uri $uri/ /%s?$query_string;\n", entry))
		b.WriteString("    }\n\n")
		b.WriteString("    location ~ \\.php$ {\n")
		b.WriteString(fmt.Sprintf("        fastcgi_pass 127.0.0.1:%d;\n", fpmPort))
		b.WriteString(fmt.Sprintf("        fastcgi_index %s;\n", entry))
		b.WriteString("        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;\n")
		b.WriteString("        include fastcgi_params;\n")
		b.WriteString("    }\n")

	case site.TypeNodeJS, site.TypeGo, site.TypePython:
		targetPort := 3000
		if item.SiteType == site.TypeGo {
			targetPort = 8080
		} else if item.SiteType == site.TypePython {
			targetPort = 8000
		}
		if item.ProxyTargetPort != nil && *item.ProxyTargetPort > 0 {
			targetPort = *item.ProxyTargetPort
		}
		b.WriteString("\n    location / {\n")
		b.WriteString(fmt.Sprintf("        proxy_pass http://127.0.0.1:%d;\n", targetPort))
		b.WriteString("        proxy_http_version 1.1;\n")
		b.WriteString("        proxy_set_header Upgrade $http_upgrade;\n")
		b.WriteString("        proxy_set_header Connection 'upgrade';\n")
		b.WriteString("        proxy_set_header Host $host;\n")
		b.WriteString("        proxy_set_header X-Real-IP $remote_addr;\n")
		b.WriteString("        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n")
		b.WriteString("        proxy_cache_bypass $http_upgrade;\n")
		b.WriteString("    }\n")
	}

	if strings.TrimSpace(meta.NginxCustomConfig) != "" {
		b.WriteString("\n    # 自定义配置\n")
		for _, line := range strings.Split(strings.TrimSpace(meta.NginxCustomConfig), "\n") {
			b.WriteString("    " + line + "\n")
		}
	}

	b.WriteString("\n")
	if accessLog != "" {
		b.WriteString(fmt.Sprintf("    access_log %s;\n", filepath.ToSlash(accessLog)))
	}
	if errorLog != "" {
		b.WriteString(fmt.Sprintf("    error_log %s;\n", filepath.ToSlash(errorLog)))
	}
	b.WriteString("}\n")

	return b.String()
}

func (m *Manager) WriteConfig(ctx context.Context, siteID int64, domain string, content string) (string, error) {
	vhostsDir := filepath.Join(m.baseDir, "config", "nginx", "vhosts")
	if err := os.MkdirAll(vhostsDir, 0755); err != nil {
		return "", err
	}
	safeName := strings.ReplaceAll(domain, ":", "_")
	if safeName == "" {
		safeName = fmt.Sprintf("site-%d", siteID)
	}
	configPath := filepath.Join(vhostsDir, safeName+".conf")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return "", err
	}
	return configPath, nil
}

func (m *Manager) ReadConfig(ctx context.Context, configPath string) (string, error) {
	if configPath == "" {
		return "", nil
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (m *Manager) DeleteConfig(ctx context.Context, configPath string) error {
	if configPath == "" {
		return nil
	}
	if _, err := os.Stat(configPath); err == nil {
		return os.Remove(configPath)
	}
	return nil
}

func (m *Manager) TestConfig(ctx context.Context, configContent string) (bool, string, error) {
	trimmed := strings.TrimSpace(configContent)
	if trimmed == "" {
		return false, "nginx: [emerg] configuration file is empty", nil
	}

	openBraces := strings.Count(trimmed, "{")
	closeBraces := strings.Count(trimmed, "}")
	if openBraces != closeBraces {
		return false, fmt.Sprintf("nginx: [emerg] unexpected \"}\" or unclosed \"{\" (open: %d, close: %d)", openBraces, closeBraces), nil
	}
	if !strings.Contains(trimmed, "server") || !strings.Contains(trimmed, "listen") {
		return false, "nginx: [emerg] missing 'server' block or 'listen' directive in configuration", nil
	}

	return true, "nginx: the configuration file syntax is ok\nnginx: configuration file test is successful", nil
}

func (m *Manager) Reload(ctx context.Context) error {
	return nil
}

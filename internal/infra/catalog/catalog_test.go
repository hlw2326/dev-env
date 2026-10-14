package catalog_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"devenv/internal/domain/software"
	"devenv/internal/infra/catalog"
)

func TestCatalogPHPParsingAndError(t *testing.T) {
	phpJSON := `{
		"8.3": {
			"version": "8.3.35",
			"nts-vs16-x64": {
				"zip": {
					"path": "php-8.3.35-nts-Win32-vs16-x64.zip",
					"size": "32.16MB",
					"sha256": "deadbeef"
				}
			}
		},
		"8.2": {
			"version": "8.2.28",
			"nts-vs16-x64": {
				"zip": {
					"path": "php-8.2.28-nts-Win32-vs16-x64.zip",
					"size": "30.00MB",
					"sha256": "cafebabe"
				}
			}
		}
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(phpJSON))
	}))
	defer srv.Close()

	client := catalog.NewClientWithEndpoints(srv.Client(), catalog.Endpoints{
		PHPReleases: srv.URL + "/releases.json",
		PHPBase:     "https://downloads.php.net/~windows/releases/",
	})

	list, err := client.Versions(context.Background(), software.KindPHP)
	if err != nil {
		t.Fatalf("Versions failed: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(list))
	}
	if list[0].Version != "8.3.35" || list[0].Checksum != "deadbeef" {
		t.Fatalf("unexpected first PHP version: %+v", list[0])
	}
}

func TestCatalogNodeParsingAndOfflineFailure(t *testing.T) {
	nodeJSON := `[
		{"version": "v22.14.0", "date": "2025-02-11", "files": ["win-x64-zip"], "lts": "Jod"},
		{"version": "v20.18.3", "date": "2025-02-10", "files": ["win-x64-zip"], "lts": "Iron"}
	]`

	mux := http.NewServeMux()
	mux.HandleFunc("/index.json", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(nodeJSON))
	})
	mux.HandleFunc("/v22.14.0/SHASUMS256.txt", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("11223344  node-v22.14.0-win-x64.zip\n"))
	})
	mux.HandleFunc("/v20.18.3/SHASUMS256.txt", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("55667788  node-v20.18.3-win-x64.zip\n"))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := catalog.NewClientWithEndpoints(srv.Client(), catalog.Endpoints{
		NodeIndex: srv.URL + "/index.json",
		NodeDist:  srv.URL,
	})

	list, err := client.Versions(context.Background(), software.KindNode)
	if err != nil {
		t.Fatalf("Node Versions failed: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 node versions, got %d", len(list))
	}
	if list[0].Version != "22.14.0" || list[0].Checksum != "11223344" || !list[0].Recommended {
		t.Fatalf("unexpected first node version: %+v", list[0])
	}

	// Verify failure returns error with no built-in fallback
	errSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer errSrv.Close()

	errClient := catalog.NewClientWithEndpoints(errSrv.Client(), catalog.Endpoints{
		PHPReleases: errSrv.URL,
	})
	if fallbackList, err := errClient.Versions(context.Background(), software.KindPHP); err == nil || len(fallbackList) != 0 {
		t.Fatalf("expected error and zero fallback versions on HTTP failure, got list=%v err=%v", fallbackList, err)
	}
}

func TestCatalogMySQLAndNginxProbe(t *testing.T) {
	client := catalog.NewClient()
	mysqlList, err := client.Versions(context.Background(), software.KindMySQL)
	if err != nil {
		t.Fatalf("MySQL Versions failed: %v", err)
	}
	if len(mysqlList) == 0 {
		t.Fatalf("expected non-empty MySQL versions")
	}
	nginxList, err := client.Versions(context.Background(), software.KindNginx)
	if err != nil {
		t.Fatalf("Nginx Versions failed: %v", err)
	}
	if len(nginxList) == 0 {
		t.Fatalf("expected non-empty Nginx versions")
	}
	apacheList, err := client.Versions(context.Background(), software.KindApache)
	if err != nil {
		t.Fatalf("Apache Versions failed: %v", err)
	}
	if len(apacheList) == 0 {
		t.Fatalf("expected non-empty Apache versions")
	}
}


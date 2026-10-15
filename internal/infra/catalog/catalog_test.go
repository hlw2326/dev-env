package catalog_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"devenv/internal/domain/software"
	"devenv/internal/infra/catalog"
)

func TestPHPParsing(t *testing.T) {
	payload := `{
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

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		if _, err := writer.Write([]byte(payload)); err != nil {
			t.Error(err)
		}
	}))
	defer server.Close()

	client := catalog.NewWithEndpoints(server.Client(), catalog.Endpoints{
		PHPReleases: server.URL + "/releases.json",
		PHPBase:     "https://downloads.php.net/~windows/releases/",
	})

	list, err := client.Versions(context.Background(), software.KIND_PHP)
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

func TestNodeParsing(t *testing.T) {
	payload := `[
		{"version": "v22.14.0", "date": "2025-02-11", "files": ["win-x64-zip"], "lts": "Jod"},
		{"version": "v20.18.3", "date": "2025-02-10", "files": ["win-x64-zip"], "lts": "Iron"}
	]`

	mux := http.NewServeMux()
	mux.HandleFunc("/index.json", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		if _, err := writer.Write([]byte(payload)); err != nil {
			t.Error(err)
		}
	})
	mux.HandleFunc("/v22.14.0/SHASUMS256.txt", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		if _, err := writer.Write([]byte("11223344  node-v22.14.0-win-x64.zip\n")); err != nil {
			t.Error(err)
		}
	})
	mux.HandleFunc("/v20.18.3/SHASUMS256.txt", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		if _, err := writer.Write([]byte("55667788  node-v20.18.3-win-x64.zip\n")); err != nil {
			t.Error(err)
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := catalog.NewWithEndpoints(server.Client(), catalog.Endpoints{
		NodeIndex: server.URL + "/index.json",
		NodeDist:  server.URL,
	})

	list, err := client.Versions(context.Background(), software.KIND_NODE)
	if err != nil {
		t.Fatalf("Node Versions failed: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 node versions, got %d", len(list))
	}
	if list[0].Version != "22.14.0" || list[0].Checksum != "11223344" || !list[0].Recommended {
		t.Fatalf("unexpected first node version: %+v", list[0])
	}

	failServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusBadGateway)
	}))
	defer failServer.Close()

	failClient := catalog.NewWithEndpoints(failServer.Client(), catalog.Endpoints{
		PHPReleases: failServer.URL,
	})
	if fallback, err := failClient.Versions(context.Background(), software.KIND_PHP); err == nil || len(fallback) != 0 {
		t.Fatalf("expected error and zero fallback versions on HTTP failure, got list=%v err=%v", fallback, err)
	}
}

func TestMySQLAndRedis(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/MySQL-8.4/mysql-8.4.11-winx64.zip", func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Range") != "bytes=0-0" {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		writer.WriteHeader(http.StatusPartialContent)
	})
	mux.HandleFunc("/MySQL-8.0/mysql-8.0.46-winx64.zip", func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Range") != "bytes=0-0" {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		writer.WriteHeader(http.StatusPartialContent)
	})
	mux.HandleFunc("/redis", func(writer http.ResponseWriter, request *http.Request) {
		payload := `[{"tag_name":"v7.4.2","assets":[{"name":"Redis-7.4.2-Windows-x64.zip","browser_download_url":"https://example.com/redis.zip"}]}]`
		writer.WriteHeader(http.StatusOK)
		if _, err := writer.Write([]byte(payload)); err != nil {
			t.Error(err)
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := catalog.NewWithEndpoints(server.Client(), catalog.Endpoints{
		MySQLBase: server.URL,
		RedisAPI:  server.URL + "/redis",
	})

	mysqlList, err := client.Versions(context.Background(), software.KIND_MYSQL)
	if err != nil {
		t.Fatalf("MySQL Versions failed: %v", err)
	}
	if len(mysqlList) != 2 || mysqlList[0].Version != "8.4.11" || !mysqlList[0].Recommended {
		t.Fatalf("unexpected MySQL versions: %+v", mysqlList)
	}

	redisList, err := client.Versions(context.Background(), software.KIND_REDIS)
	if err != nil {
		t.Fatalf("Redis Versions failed: %v", err)
	}
	if len(redisList) != 1 || redisList[0].Version != "7.4.2" {
		t.Fatalf("unexpected Redis versions: %+v", redisList)
	}
}

func TestExtendedCatalogs(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/go", func(writer http.ResponseWriter, request *http.Request) {
		payload := `[{"version":"go1.24.1","stable":true,"files":[{"filename":"go1.24.1.windows-amd64.zip","os":"windows","arch":"amd64","kind":"archive","sha256":"aabbccdd"}]}]`
		writer.WriteHeader(http.StatusOK)
		if _, err := writer.Write([]byte(payload)); err != nil {
			t.Error(err)
		}
	})
	mux.HandleFunc("/cycles", func(writer http.ResponseWriter, request *http.Request) {
		payload := `[{"cycle":"3.13","latest":"3.13.2","lts":false},{"cycle":"3.12","latest":"3.12.9","lts":true}]`
		writer.WriteHeader(http.StatusOK)
		if _, err := writer.Write([]byte(payload)); err != nil {
			t.Error(err)
		}
	})
	mux.HandleFunc("/composer", func(writer http.ResponseWriter, request *http.Request) {
		payload := `{"stable":[{"version":"2.8.6","path":"/download/2.8.6/composer.phar","lts":false}]}`
		writer.WriteHeader(http.StatusOK)
		if _, err := writer.Write([]byte(payload)); err != nil {
			t.Error(err)
		}
	})
	mux.HandleFunc("/memcached", func(writer http.ResponseWriter, request *http.Request) {
		payload := "/gcc-15.2.0/libevent-2.1.12/x64/memcached.exe;PE32+;x86-64;15.2.0;20250810;1.6.45\n"
		writer.WriteHeader(http.StatusOK)
		if _, err := writer.Write([]byte(payload)); err != nil {
			t.Error(err)
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := catalog.NewWithEndpoints(server.Client(), catalog.Endpoints{
		GoList:       server.URL + "/go",
		GoBase:       server.URL,
		PythonList:   server.URL + "/cycles",
		PythonBase:   server.URL,
		ComposerList: server.URL + "/composer",
		ComposerBase: server.URL,
		MariaList:    server.URL + "/cycles",
		MariaBase:    server.URL,
		PostgresList: server.URL + "/cycles",
		PostgresBase: server.URL,
		MongoList:    server.URL + "/cycles",
		MongoBase:    server.URL,
		MemcacheList: server.URL + "/memcached",
		MemcacheZip:  server.URL + "/memcached.zip",
	})

	for _, kind := range []string{
		software.KIND_GO,
		software.KIND_PYTHON,
		software.KIND_COMPOSER,
		software.KIND_MARIADB,
		software.KIND_POSTGRESQL,
		software.KIND_MONGODB,
		software.KIND_MEMCACHED,
	} {
		list, err := client.Versions(context.Background(), kind)
		if err != nil || len(list) == 0 {
			t.Fatalf("expected versions for %s, got list=%v err=%v", kind, list, err)
		}
	}
}

func TestPortableDeveloperCatalogs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasPrefix(request.URL.Path, "/java/"):
			writeResponse(t, writer, `[{"release_name":"jdk-25.0.1+8","binaries":[{"package":{"link":"https://example.com/jdk.zip","checksum":"aabb"}}]}]`)
		case request.URL.Path == "/dotnet-index":
			writeResponse(t, writer, `{"releases-index":[{"channel-version":"10.0","latest-sdk":"10.0.100","support-phase":"active","release-type":"lts","releases.json":"`+"http://"+request.Host+`/dotnet-releases"}]}`)
		case request.URL.Path == "/dotnet-releases":
			writeResponse(t, writer, `{"releases":[{"sdk":{"version":"10.0.100","files":[{"name":"dotnet-sdk-10.0.100-win-x64.zip","rid":"win-x64","url":"https://example.com/dotnet.zip","hash":"ccdd"}]}}]}`)
		case request.URL.Path == "/deno":
			writeResponse(t, writer, `[{"tag_name":"v2.5.0","assets":[{"name":"deno-x86_64-pc-windows-msvc.zip","browser_download_url":"https://example.com/deno.zip","digest":"sha256:1122"}]}]`)
		case request.URL.Path == "/bun":
			writeResponse(t, writer, `[{"tag_name":"bun-v1.3.0","assets":[{"name":"bun-windows-x64.zip","browser_download_url":"https://example.com/bun.zip"}]}]`)
		case request.URL.Path == "/caddy":
			writeResponse(t, writer, `[{"tag_name":"v2.10.0","assets":[{"name":"caddy_2.10.0_windows_amd64.zip","browser_download_url":"https://example.com/caddy.zip"}]}]`)
		case request.URL.Path == "/nginx":
			writer.Header().Set("Content-Type", "text/html")
			writeResponse(t, writer, `<a href="/download/nginx-1.29.0.zip">nginx-1.29.0.zip</a><a href="/download/nginx-1.28.0.zip">nginx-1.28.0.zip</a>`)
		case request.URL.Path == "/sqlite":
			writer.Header().Set("Content-Type", "text/html")
			writeResponse(t, writer, `PRODUCT,3.51.0,2026/sqlite-tools-win-x64-3510000.zip,123`)
		case request.URL.Path == "/gradle":
			writeResponse(t, writer, `[{"version":"9.1.0","downloadUrl":"https://example.com/gradle-9.1.0-bin.zip"}]`)
		case request.URL.Path == "/maven":
			writer.Header().Set("Content-Type", "application/xml")
			writeResponse(t, writer, `<metadata><versioning><versions><version>3.9.10</version><version>3.9.11</version></versions></versioning></metadata>`)
		default:
			writer.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := catalog.NewWithEndpoints(server.Client(), catalog.Endpoints{
		JavaAPI: server.URL + "/java", DotnetIndex: server.URL + "/dotnet-index",
		DenoAPI: server.URL + "/deno", BunAPI: server.URL + "/bun",
		NginxList: server.URL + "/nginx", NginxBase: server.URL,
		CaddyAPI: server.URL + "/caddy", SQLiteList: server.URL + "/sqlite",
		GradleList: server.URL + "/gradle", MavenMeta: server.URL + "/maven", MavenBase: server.URL,
	})

	for _, kind := range []string{
		software.KIND_JAVA, software.KIND_DOTNET, software.KIND_DENO,
		software.KIND_BUN, software.KIND_NGINX, software.KIND_CADDY,
		software.KIND_SQLITE, software.KIND_GRADLE, software.KIND_MAVEN,
	} {
		list, err := client.Versions(context.Background(), kind)
		if err != nil || len(list) == 0 || !hasRecommendedVersion(list) {
			t.Fatalf("expected recommended version for %s, got list=%v err=%v", kind, list, err)
		}
	}
}

func writeResponse(t *testing.T, writer http.ResponseWriter, body string) {
	t.Helper()
	if _, err := writer.Write([]byte(body)); err != nil {
		t.Errorf("write catalog response: %v", err)
	}
}

func hasRecommendedVersion(list []software.Version) bool {
	for _, item := range list {
		if item.Recommended {
			return true
		}
	}
	return false
}

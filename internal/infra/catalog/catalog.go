package catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"devenv/internal/domain/software"
	appErrors "devenv/internal/shared/errors"
)

const (
	HTTP_TIMEOUT = 20 * time.Second
	CACHE_TTL    = time.Hour
	MAX_LTS      = 3
	MAX_REDIS    = 3
)

type Endpoints struct {
	PHPReleases  string
	PHPBase      string
	NodeIndex    string
	NodeDist     string
	GoList       string
	GoBase       string
	PythonList   string
	PythonBase   string
	ComposerList string
	ComposerBase string
	MySQLBase    string
	MariaList    string
	MariaBase    string
	PostgresList string
	PostgresBase string
	MongoList    string
	MongoBase    string
	RedisAPI     string
	MemcacheList string
	MemcacheZip  string
	JavaAPI      string
	DotnetIndex  string
	DenoAPI      string
	BunAPI       string
	NginxList    string
	NginxBase    string
	CaddyAPI     string
	SQLiteList   string
	GradleList   string
	MavenMeta    string
	MavenBase    string
}

func DefaultEndpoints() Endpoints {
	return Endpoints{
		PHPReleases:  "https://downloads.php.net/~windows/releases/releases.json",
		PHPBase:      "https://downloads.php.net/~windows/releases/",
		NodeIndex:    "https://nodejs.org/dist/index.json",
		NodeDist:     "https://nodejs.org/dist",
		GoList:       "https://go.dev/dl/?mode=json",
		GoBase:       "https://dl.google.com/go",
		PythonList:   "https://endoflife.date/api/python.json",
		PythonBase:   "https://www.python.org/ftp/python",
		ComposerList: "https://getcomposer.org/versions",
		ComposerBase: "https://getcomposer.org",
		MySQLBase:    "https://dev.mysql.com/get/Downloads",
		MariaList:    "https://endoflife.date/api/mariadb.json",
		MariaBase:    "https://archive.mariadb.org",
		PostgresList: "https://endoflife.date/api/postgresql.json",
		PostgresBase: "https://get.enterprisedb.com/postgresql",
		MongoList:    "https://endoflife.date/api/mongodb.json",
		MongoBase:    "https://fastdl.mongodb.org/windows",
		RedisAPI:     "https://api.github.com/repos/redis-windows/redis-windows/releases?per_page=5",
		MemcacheList: "https://raw.githubusercontent.com/nono303/memcached/master/bininfo.csv",
		MemcacheZip:  "https://github.com/nono303/memcached/archive/refs/heads/master.zip",
		JavaAPI:      "https://api.adoptium.net/v3/assets/feature_releases",
		DotnetIndex:  "https://builds.dotnet.microsoft.com/dotnet/release-metadata/releases-index.json",
		DenoAPI:      "https://api.github.com/repos/denoland/deno/releases?per_page=5",
		BunAPI:       "https://api.github.com/repos/oven-sh/bun/releases?per_page=5",
		NginxList:    "https://nginx.org/en/download.html",
		NginxBase:    "https://nginx.org/download",
		CaddyAPI:     "https://api.github.com/repos/caddyserver/caddy/releases?per_page=5",
		SQLiteList:   "https://www.sqlite.org/download.html",
		GradleList:   "https://services.gradle.org/versions/all",
		MavenMeta:    "https://repo.maven.apache.org/maven2/org/apache/maven/apache-maven/maven-metadata.xml",
		MavenBase:    "https://archive.apache.org/dist/maven/maven-3",
	}
}

type cacheEntry struct {
	versions []software.Version
	expires  time.Time
}

type Client struct {
	http      *http.Client
	endpoints Endpoints
	mu        sync.RWMutex
	cache     map[string]cacheEntry
	locks     map[string]*sync.Mutex
}

func NewClient() *Client {
	return NewWithEndpoints(
		&http.Client{Timeout: HTTP_TIMEOUT},
		DefaultEndpoints(),
	)
}

func NewWithEndpoints(client *http.Client, endpoints Endpoints) *Client {
	return &Client{
		http:      client,
		endpoints: endpoints,
		cache:     make(map[string]cacheEntry),
		locks:     make(map[string]*sync.Mutex),
	}
}

func (c *Client) Cached(kind string) []software.Version {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if item, ok := c.cache[kind]; ok && time.Now().Before(item.expires) {
		cloned := make([]software.Version, len(item.versions))
		copy(cloned, item.versions)
		return cloned
	}
	return nil
}

func (c *Client) kindLock(kind string) *sync.Mutex {
	c.mu.Lock()
	defer c.mu.Unlock()
	lock, ok := c.locks[kind]
	if !ok {
		lock = &sync.Mutex{}
		c.locks[kind] = lock
	}
	return lock
}

func (c *Client) Versions(ctx context.Context, kind string) ([]software.Version, error) {
	if cached := c.Cached(kind); len(cached) > 0 {
		return cached, nil
	}

	lock := c.kindLock(kind)
	lock.Lock()
	defer lock.Unlock()

	if cached := c.Cached(kind); len(cached) > 0 {
		return cached, nil
	}

	var versions []software.Version
	var err error

	switch kind {
	case software.KIND_PHP:
		versions, err = c.fetchPHP(ctx)
	case software.KIND_NODE:
		versions, err = c.fetchNode(ctx)
	case software.KIND_GO:
		versions, err = c.fetchGo(ctx)
	case software.KIND_PYTHON:
		versions, err = c.fetchPython(ctx)
	case software.KIND_COMPOSER:
		versions, err = c.fetchComposer(ctx)
	case software.KIND_MYSQL:
		versions, err = c.fetchMySQL(ctx)
	case software.KIND_MARIADB:
		versions, err = c.fetchMariaDB(ctx)
	case software.KIND_POSTGRESQL:
		versions, err = c.fetchPostgres(ctx)
	case software.KIND_MONGODB:
		versions, err = c.fetchMongoDB(ctx)
	case software.KIND_REDIS:
		versions, err = c.fetchRedis(ctx)
	case software.KIND_MEMCACHED:
		versions, err = c.fetchMemcached(ctx)
	case software.KIND_JAVA:
		versions, err = c.fetchJava(ctx)
	case software.KIND_DOTNET:
		versions, err = c.fetchDotnet(ctx)
	case software.KIND_DENO:
		versions, err = c.fetchGitHubArchives(ctx, c.endpoints.DenoAPI, "Deno", "deno-x86_64-pc-windows-msvc.zip")
	case software.KIND_BUN:
		versions, err = c.fetchGitHubArchives(ctx, c.endpoints.BunAPI, "Bun", "bun-windows-x64.zip")
	case software.KIND_NGINX:
		versions, err = c.fetchNginx(ctx)
	case software.KIND_CADDY:
		versions, err = c.fetchGitHubArchives(ctx, c.endpoints.CaddyAPI, "Caddy", "caddy_*_windows_amd64.zip")
	case software.KIND_SQLITE:
		versions, err = c.fetchSQLite(ctx)
	case software.KIND_GRADLE:
		versions, err = c.fetchGradle(ctx)
	case software.KIND_MAVEN:
		versions, err = c.fetchMaven(ctx)
	default:
		return nil, appErrors.New(appErrors.INVALID_ARGUMENT, "不支持的软件类型: "+kind)
	}

	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.cache[kind] = cacheEntry{
		versions: versions,
		expires:  time.Now().Add(CACHE_TTL),
	}
	c.mu.Unlock()

	cloned := make([]software.Version, len(versions))
	copy(cloned, versions)
	return cloned, nil
}

func (c *Client) fetchPHP(ctx context.Context) ([]software.Version, error) {
	body, err := c.getBytes(ctx, c.endpoints.PHPReleases)
	if err != nil {
		return nil, err
	}

	var raw map[string]map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, appErrors.Wrap(err, appErrors.DOWNLOAD_FAILED, "解析 PHP 版本清单失败")
	}

	type zipFile struct {
		Path   string `json:"path"`
		SHA256 string `json:"sha256"`
	}
	type zipWrap struct {
		Zip zipFile `json:"zip"`
	}

	var list []software.Version
	for branch, branchMap := range raw {
		var version string
		if rawVersion, ok := branchMap["version"]; ok {
			if err := json.Unmarshal(rawVersion, &version); err != nil {
				return nil, appErrors.Wrap(err, appErrors.DOWNLOAD_FAILED, "解析 PHP 版本号失败")
			}
		}
		if version == "" {
			continue
		}

		var chosen zipFile
		for key, rawVal := range branchMap {
			if strings.HasPrefix(key, "nts-") && strings.HasSuffix(key, "-x64") {
				var wrap zipWrap
				if err := json.Unmarshal(rawVal, &wrap); err != nil {
					return nil, appErrors.Wrap(err, appErrors.DOWNLOAD_FAILED, "解析 PHP 安装包信息失败")
				}
				if wrap.Zip.Path != "" {
					chosen = wrap.Zip
					break
				}
			}
		}
		if chosen.Path == "" {
			continue
		}

		note := branch + " NTS"
		recommended := branch == "8.3"
		list = append(list, software.Version{
			Version:     version,
			Recommended: recommended,
			Note:        note,
			URL:         strings.TrimRight(c.endpoints.PHPBase, "/") + "/" + strings.TrimLeft(chosen.Path, "/"),
			Checksum:    strings.ToLower(strings.TrimSpace(chosen.SHA256)),
		})
	}

	if len(list) == 0 {
		return nil, appErrors.New(appErrors.DOWNLOAD_FAILED, "未找到可用的 PHP x64 NTS 版本")
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].Version > list[j].Version
	})
	return list, nil
}

func (c *Client) fetchNode(ctx context.Context) ([]software.Version, error) {
	body, err := c.getBytes(ctx, c.endpoints.NodeIndex)
	if err != nil {
		return nil, err
	}

	type releaseItem struct {
		Version string      `json:"version"`
		LTS     interface{} `json:"lts"`
		Files   []string    `json:"files"`
	}
	var releases []releaseItem
	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, appErrors.Wrap(err, appErrors.DOWNLOAD_FAILED, "解析 Node.js 版本清单失败")
	}

	checkZip := func(files []string) bool {
		for _, item := range files {
			if item == "win-x64-zip" {
				return true
			}
		}
		return false
	}

	type nodeCandidate struct {
		rawVersion string
		item       software.Version
	}

	var candidates []nodeCandidate
	seen := make(map[string]bool)
	count := 0
	hasLatest := false

	for _, release := range releases {
		if !checkZip(release.Files) {
			continue
		}
		version := strings.TrimPrefix(release.Version, "v")
		parts := strings.SplitN(version, ".", 2)
		major := parts[0]

		ltsName, isLTS := release.LTS.(string)
		if !hasLatest {
			hasLatest = true
			seen[major] = true
			note := "最新版"
			if isLTS && ltsName != "" {
				note = "LTS (" + ltsName + ")"
				count++
			}
			candidates = append(candidates, nodeCandidate{
				rawVersion: release.Version,
				item: software.Version{
					Version:     version,
					Recommended: isLTS && count == 1,
					Note:        note,
					URL:         fmt.Sprintf("%s/%s/node-%s-win-x64.zip", strings.TrimRight(c.endpoints.NodeDist, "/"), release.Version, release.Version),
				},
			})
			continue
		}

		if isLTS && ltsName != "" && !seen[major] && count < MAX_LTS {
			seen[major] = true
			count++
			candidates = append(candidates, nodeCandidate{
				rawVersion: release.Version,
				item: software.Version{
					Version:     version,
					Recommended: count == 1,
					Note:        "LTS (" + ltsName + ")",
					URL:         fmt.Sprintf("%s/%s/node-%s-win-x64.zip", strings.TrimRight(c.endpoints.NodeDist, "/"), release.Version, release.Version),
				},
			})
		}
		if hasLatest && count >= MAX_LTS {
			break
		}
	}

	if len(candidates) == 0 {
		return nil, appErrors.New(appErrors.DOWNLOAD_FAILED, "未找到可用的 Node.js 版本")
	}

	selected := make([]software.Version, len(candidates))
	errs := make([]error, len(candidates))
	var group sync.WaitGroup
	for i := range candidates {
		group.Add(1)
		go func(idx int) {
			defer group.Done()
			shasum, sumErr := c.fetchChecksum(ctx, candidates[idx].rawVersion)
			if sumErr != nil {
				errs[idx] = sumErr
				return
			}
			entry := candidates[idx].item
			entry.Checksum = shasum
			selected[idx] = entry
		}(i)
	}
	group.Wait()

	for _, sumErr := range errs {
		if sumErr != nil {
			return nil, sumErr
		}
	}
	return selected, nil
}

func (c *Client) fetchChecksum(ctx context.Context, rawVersion string) (string, error) {
	url := fmt.Sprintf("%s/%s/SHASUMS256.txt", strings.TrimRight(c.endpoints.NodeDist, "/"), rawVersion)
	body, err := c.getBytes(ctx, url)
	if err != nil {
		return "", err
	}
	target := fmt.Sprintf("node-%s-win-x64.zip", rawVersion)
	lines := strings.Split(strings.ReplaceAll(string(body), "\r\n", "\n"), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 && strings.TrimPrefix(fields[1], "*") == target {
			return strings.ToLower(fields[0]), nil
		}
	}
	return "", appErrors.New(appErrors.DOWNLOAD_FAILED, "SHASUMS256.txt 中未找到 "+target+" 的校验值")
}

func (c *Client) fetchMySQL(ctx context.Context) ([]software.Version, error) {
	groups := []struct {
		series      string
		versions    []string
		recommended bool
		note        string
	}{
		{series: "8.4", versions: []string{"8.4.11", "8.4.10", "8.4.9"}, recommended: true, note: "8.4 LTS"},
		{series: "8.0", versions: []string{"8.0.46", "8.0.45", "8.0.44"}, recommended: false, note: "8.0 稳定版"},
	}

	found := make([]*software.Version, len(groups))
	errs := make([]error, len(groups))
	base := strings.TrimRight(c.endpoints.MySQLBase, "/")
	var wait sync.WaitGroup

	for i, group := range groups {
		wait.Add(1)
		go func(idx int, grp struct {
			series      string
			versions    []string
			recommended bool
			note        string
		}) {
			defer wait.Done()
			for _, version := range grp.versions {
				url := fmt.Sprintf("%s/MySQL-%s/mysql-%s-winx64.zip", base, grp.series, version)
				if err := c.probeRange(ctx, url); err != nil {
					errs[idx] = err
					continue
				}
				found[idx] = &software.Version{
					Version:     version,
					Recommended: grp.recommended,
					Note:        grp.note,
					URL:         url,
				}
				return
			}
		}(i, group)
	}
	wait.Wait()

	var list []software.Version
	var lastErr error
	for i, item := range found {
		if item != nil {
			list = append(list, *item)
		} else if errs[i] != nil {
			lastErr = errs[i]
		}
	}
	if len(list) == 0 {
		if lastErr != nil {
			return nil, lastErr
		}
		return nil, appErrors.New(appErrors.DOWNLOAD_FAILED, "未找到可用的 MySQL 版本")
	}
	return list, nil
}

func (c *Client) fetchRedis(ctx context.Context) ([]software.Version, error) {
	body, err := c.getBytes(ctx, c.endpoints.RedisAPI)
	if err != nil {
		return nil, err
	}

	type assetItem struct {
		Name        string `json:"name"`
		DownloadURL string `json:"browser_download_url"`
	}
	type releaseItem struct {
		TagName string      `json:"tag_name"`
		Assets  []assetItem `json:"assets"`
	}
	var releases []releaseItem
	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, appErrors.Wrap(err, appErrors.DOWNLOAD_FAILED, "解析 Redis 版本清单失败")
	}

	var list []software.Version
	for _, release := range releases {
		version := strings.TrimPrefix(release.TagName, "v")
		var download string
		for _, asset := range release.Assets {
			lower := strings.ToLower(asset.Name)
			if strings.HasPrefix(lower, "redis-") && strings.Contains(lower, "windows-x64") && strings.HasSuffix(lower, ".zip") {
				download = asset.DownloadURL
				break
			}
		}
		if download == "" {
			continue
		}
		recommended := len(list) == 0
		note := "稳定版"
		if recommended {
			note = "最新稳定版"
		}
		list = append(list, software.Version{
			Version:     version,
			Recommended: recommended,
			Note:        note,
			URL:         download,
		})
		if len(list) >= MAX_REDIS {
			break
		}
	}

	if len(list) == 0 {
		return nil, appErrors.New(appErrors.DOWNLOAD_FAILED, "未找到可用的 Redis Windows x64 版本")
	}
	return list, nil
}

func (c *Client) fetchGo(ctx context.Context) ([]software.Version, error) {
	body, err := c.getBytes(ctx, c.endpoints.GoList)
	if err != nil {
		return nil, err
	}

	type goFile struct {
		Filename string `json:"filename"`
		OS       string `json:"os"`
		Arch     string `json:"arch"`
		Kind     string `json:"kind"`
		SHA256   string `json:"sha256"`
	}
	type goRelease struct {
		Version string   `json:"version"`
		Stable  bool     `json:"stable"`
		Files   []goFile `json:"files"`
	}
	var releases []goRelease
	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, appErrors.Wrap(err, appErrors.DOWNLOAD_FAILED, "解析 Go 版本清单失败")
	}

	base := strings.TrimRight(c.endpoints.GoBase, "/")
	var list []software.Version
	for _, release := range releases {
		if !release.Stable {
			continue
		}
		version := strings.TrimPrefix(release.Version, "go")
		if version == "" {
			continue
		}
		var chosen *goFile
		for i := range release.Files {
			item := &release.Files[i]
			if item.OS == "windows" && item.Arch == "amd64" && item.Kind == "archive" && strings.HasSuffix(item.Filename, ".zip") {
				chosen = item
				break
			}
		}
		if chosen == nil {
			continue
		}
		recommended := len(list) == 0
		note := "稳定版"
		if recommended {
			note = "最新稳定版"
		}
		list = append(list, software.Version{
			Version:     version,
			Recommended: recommended,
			Note:        note,
			URL:         base + "/" + strings.TrimLeft(chosen.Filename, "/"),
			Checksum:    strings.ToLower(strings.TrimSpace(chosen.SHA256)),
		})
		if len(list) >= MAX_LTS {
			break
		}
	}

	if len(list) == 0 {
		return nil, appErrors.New(appErrors.DOWNLOAD_FAILED, "未找到可用的 Go Windows amd64 版本")
	}
	return list, nil
}

func (c *Client) fetchPython(ctx context.Context) ([]software.Version, error) {
	base := strings.TrimRight(c.endpoints.PythonBase, "/")
	return c.fetchCycles(ctx, c.endpoints.PythonList, "Python", func(version string) string {
		return fmt.Sprintf("%s/%s/python-%s-embed-amd64.zip", base, version, version)
	})
}

func (c *Client) fetchComposer(ctx context.Context) ([]software.Version, error) {
	body, err := c.getBytes(ctx, c.endpoints.ComposerList)
	if err != nil {
		return nil, err
	}

	type composerItem struct {
		Version string `json:"version"`
		Path    string `json:"path"`
		LTS     bool   `json:"lts"`
	}
	type composerPayload struct {
		Stable []composerItem `json:"stable"`
	}
	var payload composerPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, appErrors.Wrap(err, appErrors.DOWNLOAD_FAILED, "解析 Composer 版本清单失败")
	}

	base := strings.TrimRight(c.endpoints.ComposerBase, "/")
	var list []software.Version
	for _, item := range payload.Stable {
		if item.Version == "" || item.Path == "" {
			continue
		}
		recommended := len(list) == 0
		note := "稳定版"
		if recommended {
			note = "最新稳定版"
		} else if item.LTS {
			note = "LTS"
		}
		list = append(list, software.Version{
			Version:     item.Version,
			Recommended: recommended,
			Note:        note,
			URL:         base + "/" + strings.TrimLeft(item.Path, "/"),
		})
		if len(list) >= MAX_LTS {
			break
		}
	}

	if len(list) == 0 {
		return nil, appErrors.New(appErrors.DOWNLOAD_FAILED, "未找到可用的 Composer 版本")
	}
	return list, nil
}

func (c *Client) fetchMariaDB(ctx context.Context) ([]software.Version, error) {
	base := strings.TrimRight(c.endpoints.MariaBase, "/")
	return c.fetchCycles(ctx, c.endpoints.MariaList, "MariaDB", func(version string) string {
		return fmt.Sprintf("%s/mariadb-%s/winx64-packages/mariadb-%s-winx64.zip", base, version, version)
	})
}

func (c *Client) fetchPostgres(ctx context.Context) ([]software.Version, error) {
	base := strings.TrimRight(c.endpoints.PostgresBase, "/")
	return c.fetchCycles(ctx, c.endpoints.PostgresList, "PostgreSQL", func(version string) string {
		return fmt.Sprintf("%s/postgresql-%s-1-windows-x64-binaries.zip", base, version)
	})
}

func (c *Client) fetchMongoDB(ctx context.Context) ([]software.Version, error) {
	base := strings.TrimRight(c.endpoints.MongoBase, "/")
	return c.fetchCycles(ctx, c.endpoints.MongoList, "MongoDB", func(version string) string {
		return fmt.Sprintf("%s/mongodb-windows-x86_64-%s.zip", base, version)
	})
}

func (c *Client) fetchMemcached(ctx context.Context) ([]software.Version, error) {
	body, err := c.getBytes(ctx, c.endpoints.MemcacheList)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.ReplaceAll(string(body), "\r\n", "\n"), "\n")
	seen := make(map[string]bool)
	var list []software.Version
	for _, line := range lines {
		fields := strings.Split(line, ";")
		if len(fields) < 6 {
			continue
		}
		path := strings.TrimSpace(fields[0])
		version := strings.TrimSpace(fields[5])
		if !strings.HasSuffix(path, "/x64/memcached.exe") || version == "" || seen[version] {
			continue
		}
		seen[version] = true
		list = append(list, software.Version{
			Version: version,
			Note:    "稳定版",
			URL:     c.endpoints.MemcacheZip,
		})
	}

	if len(list) == 0 {
		return nil, appErrors.New(appErrors.DOWNLOAD_FAILED, "未找到可用的 Memcached Windows x64 版本")
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Version > list[j].Version
	})
	if len(list) > MAX_LTS {
		list = list[:MAX_LTS]
	}
	list[0].Recommended = true
	list[0].Note = "最新稳定版"
	return list, nil
}

func (c *Client) fetchCycles(ctx context.Context, endpoint, label string, formatURL func(version string) string) ([]software.Version, error) {
	body, err := c.getBytes(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	type cycleItem struct {
		Cycle  string      `json:"cycle"`
		Latest string      `json:"latest"`
		LTS    interface{} `json:"lts"`
	}
	var cycles []cycleItem
	if err := json.Unmarshal(body, &cycles); err != nil {
		return nil, appErrors.Wrap(err, appErrors.DOWNLOAD_FAILED, "解析 "+label+" 版本清单失败")
	}

	var list []software.Version
	hasRecommended := false
	for _, item := range cycles {
		version := strings.TrimSpace(item.Latest)
		if version == "" {
			continue
		}
		isLTS := checkLTS(item.LTS)
		note := item.Cycle + " 稳定版"
		if isLTS {
			note = item.Cycle + " LTS"
		} else if len(list) == 0 {
			note = item.Cycle + " 最新版"
		}
		recommended := false
		if isLTS && !hasRecommended {
			recommended = true
			hasRecommended = true
		}
		list = append(list, software.Version{
			Version:     version,
			Recommended: recommended,
			Note:        note,
			URL:         formatURL(version),
		})
		if len(list) >= MAX_LTS {
			break
		}
	}

	if len(list) == 0 {
		return nil, appErrors.New(appErrors.DOWNLOAD_FAILED, "未找到可用的 "+label+" 版本")
	}
	if !hasRecommended {
		list[0].Recommended = true
	}
	return list, nil
}

func checkLTS(value interface{}) bool {
	if flag, ok := value.(bool); ok {
		return flag
	}
	if text, ok := value.(string); ok {
		return strings.TrimSpace(text) != ""
	}
	return false
}

func (c *Client) fetchJava(ctx context.Context) ([]software.Version, error) {
	type javaPackage struct {
		Link     string `json:"link"`
		Checksum string `json:"checksum"`
	}
	type javaBinary struct {
		Package javaPackage `json:"package"`
	}
	type javaRelease struct {
		ReleaseName string       `json:"release_name"`
		Binaries    []javaBinary `json:"binaries"`
	}

	var list []software.Version
	for _, feature := range []int{25, 21, 17} {
		endpoint := fmt.Sprintf("%s/%d/ga?architecture=x64&heap_size=normal&image_type=jdk&jvm_impl=hotspot&os=windows&page_size=1&project=jdk&sort_method=DATE&sort_order=DESC&vendor=eclipse", strings.TrimRight(c.endpoints.JavaAPI, "/"), feature)
		body, err := c.getBytes(ctx, endpoint)
		if err != nil {
			continue
		}
		var releases []javaRelease
		if json.Unmarshal(body, &releases) != nil || len(releases) == 0 || len(releases[0].Binaries) == 0 {
			continue
		}
		pkg := releases[0].Binaries[0].Package
		if pkg.Link == "" {
			continue
		}
		version := strings.TrimPrefix(releases[0].ReleaseName, "jdk-")
		list = append(list, software.Version{
			Version: version, Recommended: len(list) == 0, Note: fmt.Sprintf("Java %d LTS", feature),
			URL: pkg.Link, Checksum: strings.ToLower(pkg.Checksum),
		})
	}
	if len(list) == 0 {
		return nil, appErrors.New(appErrors.DOWNLOAD_FAILED, "未找到可用的 OpenJDK Windows x64 版本")
	}
	if !hasRecommended(list) {
		list[0].Recommended = true
	}
	return list, nil
}

func (c *Client) fetchDotnet(ctx context.Context) ([]software.Version, error) {
	type channelItem struct {
		Channel      string `json:"channel-version"`
		LatestSDK    string `json:"latest-sdk"`
		SupportPhase string `json:"support-phase"`
		ReleaseType  string `json:"release-type"`
		ReleasesURL  string `json:"releases.json"`
	}
	var index struct {
		Channels []channelItem `json:"releases-index"`
	}
	body, err := c.getBytes(ctx, c.endpoints.DotnetIndex)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, &index); err != nil {
		return nil, appErrors.Wrap(err, appErrors.DOWNLOAD_FAILED, "解析 .NET 版本索引失败")
	}
	type sdkFile struct {
		Name string `json:"name"`
		RID  string `json:"rid"`
		URL  string `json:"url"`
		Hash string `json:"hash"`
	}
	type release struct {
		SDK struct {
			Version string    `json:"version"`
			Files   []sdkFile `json:"files"`
		} `json:"sdk"`
	}
	var list []software.Version
	for _, channel := range index.Channels {
		if len(list) >= MAX_LTS || (channel.SupportPhase != "active" && channel.SupportPhase != "maintenance") || channel.LatestSDK == "" || channel.ReleasesURL == "" {
			continue
		}
		releaseBody, fetchErr := c.getBytes(ctx, channel.ReleasesURL)
		if fetchErr != nil {
			continue
		}
		var payload struct {
			Releases []release `json:"releases"`
		}
		if json.Unmarshal(releaseBody, &payload) != nil {
			continue
		}
		for _, entry := range payload.Releases {
			if entry.SDK.Version != channel.LatestSDK {
				continue
			}
			for _, file := range entry.SDK.Files {
				if file.RID == "win-x64" && strings.HasSuffix(strings.ToLower(file.Name), ".zip") {
					isLTS := strings.EqualFold(channel.ReleaseType, "lts")
					list = append(list, software.Version{Version: entry.SDK.Version, Recommended: isLTS && !hasRecommended(list), Note: channel.Channel + map[bool]string{true: " LTS", false: " STS"}[isLTS], URL: file.URL, Checksum: strings.ToLower(file.Hash)})
					break
				}
			}
			break
		}
	}
	if len(list) == 0 {
		return nil, appErrors.New(appErrors.DOWNLOAD_FAILED, "未找到可用的 .NET SDK Windows x64 版本")
	}
	if !hasRecommended(list) {
		list[0].Recommended = true
	}
	return list, nil
}

func hasRecommended(list []software.Version) bool {
	for _, item := range list {
		if item.Recommended {
			return true
		}
	}
	return false
}

func (c *Client) fetchGitHubArchives(ctx context.Context, endpoint, label, assetPattern string) ([]software.Version, error) {
	type asset struct {
		Name   string `json:"name"`
		URL    string `json:"browser_download_url"`
		Digest string `json:"digest"`
	}
	type release struct {
		Tag        string  `json:"tag_name"`
		Prerelease bool    `json:"prerelease"`
		Draft      bool    `json:"draft"`
		Assets     []asset `json:"assets"`
	}
	body, err := c.getBytes(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	var releases []release
	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, appErrors.Wrap(err, appErrors.DOWNLOAD_FAILED, "解析 "+label+" 版本列表失败")
	}
	matchAsset := func(name string) bool {
		if !strings.Contains(assetPattern, "*") {
			return strings.EqualFold(name, assetPattern)
		}
		parts := strings.SplitN(strings.ToLower(assetPattern), "*", 2)
		lower := strings.ToLower(name)
		return strings.HasPrefix(lower, parts[0]) && strings.HasSuffix(lower, parts[1])
	}
	var list []software.Version
	for _, entry := range releases {
		if entry.Draft || entry.Prerelease {
			continue
		}
		for _, file := range entry.Assets {
			if !matchAsset(file.Name) || file.URL == "" {
				continue
			}
			checksum := strings.TrimPrefix(file.Digest, "sha256:")
			list = append(list, software.Version{Version: strings.TrimPrefix(strings.TrimPrefix(entry.Tag, "bun-v"), "v"), Recommended: len(list) == 0, Note: map[bool]string{true: "最新稳定版", false: "稳定版"}[len(list) == 0], URL: file.URL, Checksum: checksum})
			break
		}
		if len(list) >= MAX_LTS {
			break
		}
	}
	if len(list) == 0 {
		return nil, appErrors.New(appErrors.DOWNLOAD_FAILED, "未找到可用的 "+label+" Windows x64 版本")
	}
	return list, nil
}

func (c *Client) fetchNginx(ctx context.Context) ([]software.Version, error) {
	body, err := c.getBytes(ctx, c.endpoints.NginxList)
	if err != nil {
		return nil, err
	}
	matches := regexp.MustCompile(`nginx-([0-9]+\.[0-9]+\.[0-9]+)\.zip`).FindAllStringSubmatch(string(body), -1)
	seen := make(map[string]bool)
	var list []software.Version
	for _, match := range matches {
		version := match[1]
		if seen[version] {
			continue
		}
		seen[version] = true
		position := len(list)
		note := "Legacy"
		if position == 0 {
			note = "Mainline"
		} else if position == 1 {
			note = "Stable"
		}
		list = append(list, software.Version{Version: version, Recommended: position == 1, Note: note, URL: strings.TrimRight(c.endpoints.NginxBase, "/") + "/nginx-" + version + ".zip"})
		if len(list) >= MAX_LTS {
			break
		}
	}
	if len(list) == 0 {
		return nil, appErrors.New(appErrors.DOWNLOAD_FAILED, "未找到可用的 Nginx Windows 版本")
	}
	if !hasRecommended(list) {
		list[0].Recommended = true
	}
	return list, nil
}

func (c *Client) fetchSQLite(ctx context.Context) ([]software.Version, error) {
	body, err := c.getBytes(ctx, c.endpoints.SQLiteList)
	if err != nil {
		return nil, err
	}
	re := regexp.MustCompile(`PRODUCT,([0-9.]+),([0-9]{4}/sqlite-tools-win-x64-[0-9]+\.zip)`)
	match := re.FindStringSubmatch(string(body))
	if len(match) == 0 {
		return nil, appErrors.New(appErrors.DOWNLOAD_FAILED, "未找到可用的 SQLite Windows x64 工具包")
	}
	return []software.Version{{Version: match[1], Recommended: true, Note: "最新稳定版", URL: "https://www.sqlite.org/" + match[2]}}, nil
}

func (c *Client) fetchGradle(ctx context.Context) ([]software.Version, error) {
	type gradleRelease struct {
		Version     string `json:"version"`
		DownloadURL string `json:"downloadUrl"`
		Snapshot    bool   `json:"snapshot"`
		Nightly     bool   `json:"nightly"`
		Broken      bool   `json:"broken"`
		RC          bool   `json:"activeRc"`
	}
	body, err := c.getBytes(ctx, c.endpoints.GradleList)
	if err != nil {
		return nil, err
	}
	var releases []gradleRelease
	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, appErrors.Wrap(err, appErrors.DOWNLOAD_FAILED, "解析 Gradle 版本列表失败")
	}
	var list []software.Version
	for _, entry := range releases {
		if entry.Version == "" || entry.DownloadURL == "" || entry.Snapshot || entry.Nightly || entry.Broken || entry.RC || !strings.HasSuffix(entry.DownloadURL, "-bin.zip") {
			continue
		}
		list = append(list, software.Version{Version: entry.Version, Recommended: len(list) == 0, Note: map[bool]string{true: "最新稳定版", false: "稳定版"}[len(list) == 0], URL: entry.DownloadURL})
		if len(list) >= MAX_LTS {
			break
		}
	}
	if len(list) == 0 {
		return nil, appErrors.New(appErrors.DOWNLOAD_FAILED, "未找到可用的 Gradle 版本")
	}
	return list, nil
}

func (c *Client) fetchMaven(ctx context.Context) ([]software.Version, error) {
	body, err := c.getBytes(ctx, c.endpoints.MavenMeta)
	if err != nil {
		return nil, err
	}
	matches := regexp.MustCompile(`<version>(3\.[0-9]+\.[0-9]+)</version>`).FindAllStringSubmatch(string(body), -1)
	var list []software.Version
	for i := len(matches) - 1; i >= 0 && len(list) < MAX_LTS; i-- {
		version := matches[i][1]
		list = append(list, software.Version{Version: version, Recommended: len(list) == 0, Note: map[bool]string{true: "最新稳定版", false: "稳定版"}[len(list) == 0], URL: fmt.Sprintf("%s/%s/binaries/apache-maven-%s-bin.zip", strings.TrimRight(c.endpoints.MavenBase, "/"), version, version)})
	}
	if len(list) == 0 {
		return nil, appErrors.New(appErrors.DOWNLOAD_FAILED, "未找到可用的 Maven 版本")
	}
	return list, nil
}

func (c *Client) getBytes(ctx context.Context, url string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.DOWNLOAD_FAILED, "创建请求失败")
	}
	request.Header.Set("User-Agent", "curl/8.13.0")
	request.Header.Set("Accept", "application/json, text/plain, */*")

	response, err := c.http.Do(request)
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.DOWNLOAD_FAILED, "拉取版本清单失败")
	}
	defer func() {
		if closeErr := response.Body.Close(); closeErr != nil {
			return
		}
	}()

	if response.StatusCode != http.StatusOK {
		return nil, appErrors.New(appErrors.DOWNLOAD_FAILED, fmt.Sprintf("拉取版本清单返回状态码 %d", response.StatusCode))
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.DOWNLOAD_FAILED, "读取版本清单响应失败")
	}
	return data, nil
}

func (c *Client) probeRange(ctx context.Context, url string) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return appErrors.Wrap(err, appErrors.DOWNLOAD_FAILED, "创建探测请求失败")
	}
	request.Header.Set("User-Agent", "curl/8.13.0")
	request.Header.Set("Range", "bytes=0-0")

	response, err := c.http.Do(request)
	if err != nil {
		return appErrors.Wrap(err, appErrors.DOWNLOAD_FAILED, "探测下载地址失败")
	}
	defer func() {
		if closeErr := response.Body.Close(); closeErr != nil {
			return
		}
	}()

	if response.StatusCode != http.StatusPartialContent && response.StatusCode != http.StatusOK {
		return appErrors.New(appErrors.DOWNLOAD_FAILED, fmt.Sprintf("探测下载地址返回状态码 %d", response.StatusCode))
	}
	return nil
}

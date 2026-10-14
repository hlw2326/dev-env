package catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"devenv/internal/domain/software"
	appErrors "devenv/internal/shared/errors"
)

type Endpoints struct {
	PHPReleases string
	PHPBase     string
	NodeIndex   string
	NodeDist    string
	MySQLBase   string
	RedisAPI    string
	NginxBase   string
}

func DefaultEndpoints() Endpoints {
	return Endpoints{
		PHPReleases: "https://downloads.php.net/~windows/releases/releases.json",
		PHPBase:     "https://downloads.php.net/~windows/releases/",
		NodeIndex:   "https://nodejs.org/dist/index.json",
		NodeDist:    "https://nodejs.org/dist",
		MySQLBase:   "https://cdn.mysql.com/Downloads",
		RedisAPI:    "https://api.github.com/repos/redis-windows/redis-windows/releases?per_page=5",
		NginxBase:   "https://nginx.org/download",
	}
}

type cacheEntry struct {
	versions  []software.Version
	expiresAt time.Time
}

type Client struct {
	http      *http.Client
	endpoints Endpoints
	mu        sync.RWMutex
	cache     map[string]cacheEntry
}

func NewClient() *Client {
	return NewClientWithEndpoints(
		&http.Client{Timeout: 20 * time.Second},
		DefaultEndpoints(),
	)
}

func NewClientWithEndpoints(httpClient *http.Client, endpoints Endpoints) *Client {
	return &Client{
		http:      httpClient,
		endpoints: endpoints,
		cache:     make(map[string]cacheEntry),
	}
}

func (c *Client) Versions(ctx context.Context, kind string) ([]software.Version, error) {
	c.mu.RLock()
	if ent, ok := c.cache[kind]; ok && time.Now().Before(ent.expiresAt) {
		cloned := make([]software.Version, len(ent.versions))
		copy(cloned, ent.versions)
		c.mu.RUnlock()
		return cloned, nil
	}
	c.mu.RUnlock()

	var versions []software.Version
	var err error

	switch kind {
	case software.KindPHP:
		versions, err = c.fetchPHP(ctx)
	case software.KindNode:
		versions, err = c.fetchNode(ctx)
	case software.KindMySQL:
		versions, err = c.fetchMySQL(ctx)
	case software.KindRedis:
		versions, err = c.fetchRedis(ctx)
	case software.KindNginx:
		versions, err = c.fetchNginx(ctx)
	default:
		return nil, appErrors.New(appErrors.CodeInvalidArgument, "不支持的软件类型: "+kind)
	}

	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.cache[kind] = cacheEntry{
		versions:  versions,
		expiresAt: time.Now().Add(time.Hour),
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
		return nil, appErrors.Wrap(err, appErrors.CodeDownloadFailed, "解析 PHP 版本清单失败")
	}

	type zipFile struct {
		Path   string `json:"path"`
		SHA256 string `json:"sha256"`
	}
	type zipWrapper struct {
		Zip zipFile `json:"zip"`
	}

	var list []software.Version
	for branch, branchMap := range raw {
		var versionStr string
		if verRaw, ok := branchMap["version"]; ok {
			_ = json.Unmarshal(verRaw, &versionStr)
		}
		if versionStr == "" {
			continue
		}

		var chosen zipFile
		for key, valRaw := range branchMap {
			if strings.HasPrefix(key, "nts-") && strings.HasSuffix(key, "-x64") {
				var wrap zipWrapper
				if json.Unmarshal(valRaw, &wrap) == nil && wrap.Zip.Path != "" {
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
			Version:     versionStr,
			Recommended: recommended,
			Note:        note,
			URL:         strings.TrimRight(c.endpoints.PHPBase, "/") + "/" + strings.TrimLeft(chosen.Path, "/"),
			Checksum:    strings.ToLower(strings.TrimSpace(chosen.SHA256)),
		})
	}

	if len(list) == 0 {
		return nil, appErrors.New(appErrors.CodeDownloadFailed, "未找到可用的 PHP x64 NTS 版本")
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

	type nodeRelease struct {
		Version string      `json:"version"`
		LTS     interface{} `json:"lts"`
		Files   []string    `json:"files"`
	}
	var releases []nodeRelease
	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodeDownloadFailed, "解析 Node.js 版本清单失败")
	}

	hasWinZip := func(files []string) bool {
		for _, f := range files {
			if f == "win-x64-zip" {
				return true
			}
		}
		return false
	}

	var selected []software.Version
	seenMajor := make(map[string]bool)
	ltsCount := 0
	addedLatest := false

	for _, rel := range releases {
		if !hasWinZip(rel.Files) {
			continue
		}
		ver := strings.TrimPrefix(rel.Version, "v")
		parts := strings.SplitN(ver, ".", 2)
		major := parts[0]

		ltsName, isLTS := rel.LTS.(string)
		if !addedLatest {
			addedLatest = true
			seenMajor[major] = true
			note := "最新版"
			if isLTS && ltsName != "" {
				note = "LTS (" + ltsName + ")"
				ltsCount++
			}
			shasum, sumErr := c.fetchNodeChecksum(ctx, rel.Version)
			if sumErr != nil {
				return nil, sumErr
			}
			selected = append(selected, software.Version{
				Version:     ver,
				Recommended: isLTS && ltsCount == 1,
				Note:        note,
				URL:         fmt.Sprintf("%s/%s/node-%s-win-x64.zip", strings.TrimRight(c.endpoints.NodeDist, "/"), rel.Version, rel.Version),
				Checksum:    shasum,
			})
			continue
		}

		if isLTS && ltsName != "" && !seenMajor[major] && ltsCount < 3 {
			seenMajor[major] = true
			ltsCount++
			shasum, sumErr := c.fetchNodeChecksum(ctx, rel.Version)
			if sumErr != nil {
				return nil, sumErr
			}
			selected = append(selected, software.Version{
				Version:     ver,
				Recommended: ltsCount == 1,
				Note:        "LTS (" + ltsName + ")",
				URL:         fmt.Sprintf("%s/%s/node-%s-win-x64.zip", strings.TrimRight(c.endpoints.NodeDist, "/"), rel.Version, rel.Version),
				Checksum:    shasum,
			})
		}
		if addedLatest && ltsCount >= 3 {
			break
		}
	}

	if len(selected) == 0 {
		return nil, appErrors.New(appErrors.CodeDownloadFailed, "未找到可用的 Node.js 版本")
	}
	return selected, nil
}

func (c *Client) fetchNodeChecksum(ctx context.Context, rawVersion string) (string, error) {
	url := fmt.Sprintf("%s/%s/SHASUMS256.txt", strings.TrimRight(c.endpoints.NodeDist, "/"), rawVersion)
	body, err := c.getBytes(ctx, url)
	if err != nil {
		return "", err
	}
	targetFile := fmt.Sprintf("node-%s-win-x64.zip", rawVersion)
	lines := strings.Split(strings.ReplaceAll(string(body), "\r\n", "\n"), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 && strings.TrimPrefix(fields[1], "*") == targetFile {
			return strings.ToLower(fields[0]), nil
		}
	}
	return "", appErrors.New(appErrors.CodeDownloadFailed, "SHASUMS256.txt 中未找到 "+targetFile+" 的校验值")
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

	var list []software.Version
	var lastErr error
	base := strings.TrimRight(c.endpoints.MySQLBase, "/")
	for _, grp := range groups {
		for _, ver := range grp.versions {
			url := fmt.Sprintf("%s/MySQL-%s/mysql-%s-winx64.zip", base, grp.series, ver)
			if err := c.probeRange(ctx, url); err != nil {
				lastErr = err
				continue
			}
			list = append(list, software.Version{
				Version:     ver,
				Recommended: grp.recommended,
				Note:        grp.note,
				URL:         url,
			})
			break
		}
	}
	if len(list) == 0 {
		if lastErr != nil {
			return nil, lastErr
		}
		return nil, appErrors.New(appErrors.CodeDownloadFailed, "未找到可用的 MySQL 版本")
	}
	return list, nil
}

func (c *Client) fetchRedis(ctx context.Context) ([]software.Version, error) {
	body, err := c.getBytes(ctx, c.endpoints.RedisAPI)
	if err != nil {
		return nil, err
	}

	type ghAsset struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	}
	type ghRelease struct {
		TagName string    `json:"tag_name"`
		Assets  []ghAsset `json:"assets"`
	}
	var releases []ghRelease
	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodeDownloadFailed, "解析 Redis 版本清单失败")
	}

	var list []software.Version
	for _, rel := range releases {
		ver := strings.TrimPrefix(rel.TagName, "v")
		var assetURL string
		for _, a := range rel.Assets {
			lower := strings.ToLower(a.Name)
			if strings.HasPrefix(lower, "redis-") && strings.Contains(lower, "windows-x64") && strings.HasSuffix(lower, ".zip") {
				assetURL = a.BrowserDownloadURL
				break
			}
		}
		if assetURL == "" {
			continue
		}
		recommended := len(list) == 0
		note := "稳定版"
		if recommended {
			note = "最新稳定版"
		}
		list = append(list, software.Version{
			Version:     ver,
			Recommended: recommended,
			Note:        note,
			URL:         assetURL,
		})
		if len(list) >= 3 {
			break
		}
	}

	if len(list) == 0 {
		return nil, appErrors.New(appErrors.CodeDownloadFailed, "未找到可用的 Redis Windows x64 版本")
	}
	return list, nil
}

func (c *Client) fetchNginx(ctx context.Context) ([]software.Version, error) {
	candidates := []struct {
		version     string
		recommended bool
		note        string
	}{
		{version: "1.26.3", recommended: true, note: "1.26 稳定版"},
		{version: "1.24.0", recommended: false, note: "1.24 稳定版"},
	}

	var list []software.Version
	for _, cand := range candidates {
		url := fmt.Sprintf("%s/nginx-%s.zip", strings.TrimRight(c.endpoints.NginxBase, "/"), cand.version)
		if err := c.probeRange(ctx, url); err != nil {
			return nil, err
		}
		list = append(list, software.Version{
			Version:     cand.version,
			Recommended: cand.recommended,
			Note:        cand.note,
			URL:         url,
		})
	}
	return list, nil
}

func (c *Client) getBytes(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodeDownloadFailed, "创建请求失败")
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) DevEnv/2.0")
	req.Header.Set("Accept", "application/json, text/plain, */*")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodeDownloadFailed, "拉取版本清单失败")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, appErrors.New(appErrors.CodeDownloadFailed, fmt.Sprintf("拉取版本清单返回状态码 %d", resp.StatusCode))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.CodeDownloadFailed, "读取版本清单响应失败")
	}
	return data, nil
}

func (c *Client) probeRange(ctx context.Context, url string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeDownloadFailed, "创建探测请求失败")
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) DevEnv/2.0")
	req.Header.Set("Range", "bytes=0-0")

	resp, err := c.http.Do(req)
	if err != nil {
		return appErrors.Wrap(err, appErrors.CodeDownloadFailed, "探测下载地址失败")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
		return appErrors.New(appErrors.CodeDownloadFailed, fmt.Sprintf("探测下载地址返回状态码 %d", resp.StatusCode))
	}
	return nil
}

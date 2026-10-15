package deployment

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	model "devenv/internal/domain/deployment"
	"devenv/internal/shared/paths"
)

const MAX_FILES = 100000

var protectedPatterns = []string{".git", ".svn", ".hg", ".env*", ".ssh", ".npmrc", ".yarnrc*", ".pnp.*", ".pypirc", ".netrc", ".pgpass", ".my.cnf", ".git-credentials", "node_modules", "*.pem", "*.key", "*.p12", "*.pfx", "id_rsa*", "id_ed25519*", "auth.json", "*.db", "*.db-*", "*.sqlite*", "*.sql", "*.sql.gz", ".DS_Store", ".devenv-*"}
var serverPatterns = []string{"storage", "runtime", "logs", "uploads", "*.log"}

func patterns(config model.Config) ([]string, error) {
	items := append([]string{}, protectedPatterns...)
	if config.Provider == model.PROVIDER_SFTP {
		items = append(items, serverPatterns...)
	}
	if config.Provider != model.PROVIDER_SFTP {
		return items, nil
	}
	for _, line := range strings.Split(config.Exclude, "\n") {
		pattern := strings.TrimSpace(line)
		if pattern == "" {
			continue
		}
		if strings.Contains(pattern, "\\") || strings.Contains(pattern, "..") || strings.Contains(pattern, "**") || strings.HasPrefix(pattern, "/") {
			return nil, fmt.Errorf("排除规则 %q 无效；使用相对路径、目录名或 *.ext，不使用 ** 或 ..", pattern)
		}
		pattern = strings.TrimSuffix(pattern, "/")
		if _, err := path.Match(pattern, "file"); err != nil {
			return nil, fmt.Errorf("排除规则无效: %w", err)
		}
		items = append(items, pattern)
	}
	return items, nil
}

func excluded(relative string, items []string) bool {
	for _, pattern := range items {
		if strings.Contains(pattern, "/") {
			match, _ := path.Match(pattern, relative)
			if match || strings.HasPrefix(relative, pattern+"/") {
				return true
			}
			continue
		}
		for _, part := range strings.Split(relative, "/") {
			match, _ := path.Match(strings.ToLower(pattern), strings.ToLower(part))
			if match {
				return true
			}
		}
	}
	return false
}

func scan(ctx context.Context, config model.Config, root string) (model.Plan, error) {
	plan := model.Plan{ProjectID: config.ProjectID, Provider: config.Provider, Files: make([]model.File, 0)}
	source, err := paths.DeploymentSource(root, config.Source)
	if err != nil {
		return plan, err
	}
	plan.Source = source
	items, err := patterns(config)
	if err != nil {
		return plan, err
	}
	err = filepath.WalkDir(source, func(file string, entry fs.DirEntry, walkError error) error {
		if walkError != nil {
			return walkError
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if file == source {
			return nil
		}
		relative, err := filepath.Rel(source, file)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if excluded(relative, items) || sameFilePath(file, config.PrivateKey) {
			plan.Excluded++
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("发布目录含符号链接 %s，请排除或复制为实际文件", relative)
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("不支持的文件类型: %s", relative)
		}
		if len(plan.Files) >= MAX_FILES {
			return fmt.Errorf("发布文件超过 %d 个，请缩小目录或增加排除规则", MAX_FILES)
		}
		digest, err := fileDigest(ctx, file)
		if err != nil {
			return err
		}
		plan.Files = append(plan.Files, model.File{Path: relative, Size: info.Size(), Digest: digest})
		plan.Bytes += info.Size()
		return nil
	})
	if err == nil && len(plan.Files) == 0 {
		err = fmt.Errorf("发布目录中没有可上传文件")
	}
	return plan, err
}

func sameFilePath(left, right string) bool {
	return right != "" && strings.EqualFold(filepath.Clean(left), filepath.Clean(right))
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r contextReader) Read(buffer []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(buffer)
}
func fileDigest(ctx context.Context, file string) (string, error) {
	reader, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer reader.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, contextReader{ctx, reader}); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// Stage an immutable, filtered snapshot before any external writes.
func stage(ctx context.Context, base string, config model.Config, root string, plan model.Plan) (string, func(), error) {
	current, err := scan(ctx, config, root)
	if err != nil {
		return "", nil, err
	}
	if len(current.Files) != len(plan.Files) {
		return "", nil, fmt.Errorf("本地文件已变化，请重新预览")
	}
	for index, item := range current.Files {
		if item != plan.Files[index] {
			return "", nil, fmt.Errorf("文件 %s 已变化，请重新预览", item.Path)
		}
	}
	cache := paths.DeploymentCache(base)
	if err := os.MkdirAll(cache, 0o700); err != nil {
		return "", nil, err
	}
	directory, err := os.MkdirTemp(cache, "snapshot-")
	if err != nil {
		return "", nil, err
	}
	cleanup := func() { _ = os.RemoveAll(directory) }
	for _, item := range plan.Files {
		source := filepath.Join(current.Source, filepath.FromSlash(item.Path))
		destination := filepath.Join(directory, filepath.FromSlash(item.Path))
		if err := copySnapshot(ctx, source, destination, item.Digest); err != nil {
			cleanup()
			return "", nil, err
		}
	}
	return directory, cleanup, nil
}

func copySnapshot(ctx context.Context, source, destination, digest string) error {
	resolved, err := filepath.EvalSymlinks(source)
	if err != nil {
		return err
	}
	if !sameFilePath(resolved, source) {
		return fmt.Errorf("文件路径已变为符号链接，请重新预览")
	}
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	info, err := input.Stat()
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("发布文件不再是普通文件")
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o700); err != nil {
		return err
	}
	output, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, info.Mode().Perm())
	if err != nil {
		return err
	}
	hash := sha256.New()
	_, copyError := io.Copy(output, io.TeeReader(contextReader{ctx, input}, hash))
	closeError := output.Close()
	if copyError != nil {
		return copyError
	}
	if closeError != nil {
		return closeError
	}
	if hex.EncodeToString(hash.Sum(nil)) != digest {
		return fmt.Errorf("复制期间文件发生变化，请重新预览")
	}
	return nil
}

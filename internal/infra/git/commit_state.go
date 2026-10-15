package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type indexSnapshot struct {
	path    string
	content []byte
	exists  bool
	head    string
}

func normalizeLineEndings(content []byte) []byte {
	return bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n"))
}

func captureIndex(ctx context.Context, binary, root string) (indexSnapshot, error) {
	output, err := run(ctx, binary, root, STATUS_TIMEOUT, "rev-parse", "--git-path", "index")
	if err != nil {
		return indexSnapshot{}, err
	}
	path := strings.TrimSpace(string(output))
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	content, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return indexSnapshot{}, err
	}
	head, _ := runRaw(ctx, binary, root, STATUS_TIMEOUT, "rev-parse", "--verify", "HEAD")
	return indexSnapshot{path: path, content: content, exists: err == nil, head: string(head)}, nil
}

func restoreCommitState(binary, root string, index indexSnapshot, path string, original, updated []byte) error {
	head, _ := runRaw(context.Background(), binary, root, STATUS_TIMEOUT, "rev-parse", "--verify", "HEAD")
	if string(head) != index.head {
		return fmt.Errorf("HEAD 已变化，不自动覆盖当前版本文件或暂存区")
	}
	var versionError error
	if path != "" && original != nil {
		current, err := os.ReadFile(path)
		if err != nil {
			versionError = err
		} else if !bytes.Equal(current, updated) {
			versionError = fmt.Errorf("版本文件被其他操作修改，已保留现有内容")
		} else {
			info, err := os.Stat(path)
			if err != nil {
				versionError = err
			} else {
				versionError = os.WriteFile(path, original, info.Mode().Perm())
			}
		}
	}
	// Use Git's index.lock convention to avoid overwriting another Git writer.
	lock, err := os.OpenFile(index.path+".lock", os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return errors.Join(versionError, err)
	}
	defer os.Remove(index.path + ".lock")
	if index.exists {
		_, err = lock.Write(index.content)
		closeError := lock.Close()
		if err == nil {
			err = closeError
		}
		if err == nil {
			err = os.Rename(index.path+".lock", index.path)
		}
	} else {
		err = lock.Close()
		if err == nil {
			err = os.Remove(index.path)
			if os.IsNotExist(err) {
				err = nil
			}
		}
	}
	return errors.Join(versionError, err)
}

package manager

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func (m *Manager) syncData() error {
	includes := parseDirList(m.Cfg.CopyDirs)
	excludes := map[string]struct{}{}
	for _, name := range parseDirList(m.Cfg.SkipDirs) {
		excludes[name] = struct{}{}
	}

	var dirs []string
	for _, name := range includes {
		if _, skip := excludes[name]; skip {
			continue
		}
		dirs = append(dirs, name)
	}
	if len(dirs) == 0 {
		return fmt.Errorf("no directories to copy after applying SKIP_DIRS")
	}

	var existing []string
	for _, name := range dirs {
		path := filepath.Join(m.Cfg.RepoPath, name)
		if _, err := os.Stat(path); err != nil {
			log.Printf("warning: expected directory missing: %s: %v", path, err)
			continue
		}
		existing = append(existing, name)
	}
	if len(existing) == 0 {
		return fmt.Errorf("no configured directories found in repo")
	}

	if err := os.MkdirAll(m.Cfg.DataDir, 0o755); err != nil {
		return fmt.Errorf("ensure data dir: %w", err)
	}

	var copied []string
	for _, name := range existing {
		src := filepath.Join(m.Cfg.RepoPath, name)
		dst := filepath.Join(m.Cfg.DataDir, name)
		if err := copyDir(src, dst); err != nil {
			return fmt.Errorf("copy %s: %w", name, err)
		}
		copied = append(copied, dst)
	}

	if m.Cfg.PluginsUID != 0 && runtime.GOOS == "linux" {
		for _, dst := range copied {
			if err := chownRecursive(dst, m.Cfg.PluginsUID); err != nil {
				return fmt.Errorf("chown %s: %w", dst, err)
			}
		}
	}

	return nil
}

func parseDirList(raw string) []string {
	var out []string
	seen := map[string]struct{}{}
	for _, part := range strings.Split(raw, ",") {
		name := strings.TrimSpace(part)
		if name == "" || name == "." {
			continue
		}
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out
}

func copyDir(src, dst string) error {
	if err := os.RemoveAll(dst); err != nil {
		return fmt.Errorf("clean destination: %w", err)
	}
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		target := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}

		if err := copyFile(path, target, info.Mode()); err != nil {
			return err
		}
		return nil
	})
}

func copyFile(src, dst string, perm os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	return out.Sync()
}

func chownRecursive(root string, uid int) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if err := os.Chown(path, uid, uid); err != nil {
			return fmt.Errorf("chown %s: %w", path, err)
		}
		return nil
	})
}

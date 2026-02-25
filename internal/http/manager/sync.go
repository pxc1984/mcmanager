package manager

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

func (m *Manager) syncData() error {
	pluginsSrc := filepath.Join(m.Cfg.RepoPath, "plugins")
	worldsSrc := filepath.Join(m.Cfg.RepoPath, "bedwars_worlds")

	pluginsDst := filepath.Join(m.Cfg.DataDir, "plugins")
	worldsDst := filepath.Join(m.Cfg.DataDir, "bedwars_worlds")

	for _, path := range []string{pluginsSrc, worldsSrc} {
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("expected directory missing: %s: %w", path, err)
		}
	}

	if err := os.MkdirAll(m.Cfg.DataDir, 0o755); err != nil {
		return fmt.Errorf("ensure data dir: %w", err)
	}

	if err := copyDir(pluginsSrc, pluginsDst); err != nil {
		return fmt.Errorf("copy plugins: %w", err)
	}

	if err := copyDir(worldsSrc, worldsDst); err != nil {
		return fmt.Errorf("copy bedwars worlds: %w", err)
	}

	if m.Cfg.PluginsUID != 0 && runtime.GOOS == "linux" {
		if err := chownRecursive(pluginsDst, m.Cfg.PluginsUID); err != nil {
			return fmt.Errorf("chown plugins: %w", err)
		}
		if err := chownRecursive(worldsDst, m.Cfg.PluginsUID); err != nil {
			return fmt.Errorf("chown worlds: %w", err)
		}
	}

	return nil
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

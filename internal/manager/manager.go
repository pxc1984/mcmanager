package manager

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"mcmanager/internal/config"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/gorcon/rcon"
)

type Manager struct {
	cfg config.Config
	mu  sync.Mutex
	msg Messages
}

type Messages struct {
	RestartSoon string
	Countdown   string
}

func selectMessages(locale string) Messages {
	switch locale {
	case "ru":
		return Messages{
			RestartSoon: "Обновление получено, перезапуск через 60 секунд",
			Countdown:   "Перезапуск через %d",
		}
	default:
		return Messages{
			RestartSoon: "Update received, restarting in 60 seconds",
			Countdown:   "Restarting in %d",
		}
	}
}

func New(cfg config.Config) *Manager {
	return &Manager{
		cfg: cfg,
		msg: selectMessages(cfg.Locale),
	}
}

func (m *Manager) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	log.Printf("received update request from %s", r.RemoteAddr)

	if err := m.syncRepo(); err != nil {
		log.Printf("repo sync failed: %v", err)
		http.Error(w, fmt.Sprintf("repo sync failed: %v", err), http.StatusInternalServerError)
		return
	}

	if err := m.runPluginDownload(); err != nil {
		log.Printf("plugin download failed: %v", err)
		http.Error(w, fmt.Sprintf("plugin download failed: %v", err), http.StatusInternalServerError)
		return
	}

	if err := m.syncData(); err != nil {
		log.Printf("data sync failed: %v", err)
		http.Error(w, fmt.Sprintf("data sync failed: %v", err), http.StatusInternalServerError)
		return
	}

	go func() {
		if err := m.restartWithCountdown(); err != nil {
			log.Printf("restart failed: %v", err)
		}
	}()

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("update applied; restart scheduled"))
}

func (m *Manager) syncRepo() error {
	if _, err := os.Stat(m.cfg.RepoPath); errors.Is(err, os.ErrNotExist) {
		log.Printf("cloning repo %s into %s (branch %s)", m.cfg.RepoURL, m.cfg.RepoPath, m.cfg.RepoBranch)
		_, err := git.PlainClone(m.cfg.RepoPath, false, &git.CloneOptions{
			URL:           m.cfg.RepoURL,
			ReferenceName: plumbing.NewBranchReferenceName(m.cfg.RepoBranch),
			SingleBranch:  true,
			Depth:         1,
		})
		return err
	}

	repo, err := git.PlainOpen(m.cfg.RepoPath)
	if err != nil {
		return fmt.Errorf("open repo: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("open worktree: %w", err)
	}

	log.Printf("pulling latest changes (branch %s)", m.cfg.RepoBranch)
	err = worktree.Pull(&git.PullOptions{
		RemoteName:    "origin",
		ReferenceName: plumbing.NewBranchReferenceName(m.cfg.RepoBranch),
		SingleBranch:  true,
		Force:         true,
	})
	if errors.Is(err, git.NoErrAlreadyUpToDate) {
		return nil
	}

	return err
}

func (m *Manager) runPluginDownload() error {
	scriptPath := filepath.Join(m.cfg.RepoPath, "plugins", "download.sh")
	if _, err := os.Stat(scriptPath); err != nil {
		return fmt.Errorf("plugins/download.sh missing: %w", err)
	}

	cmd := exec.Command("bash", scriptPath)
	cmd.Dir = filepath.Dir(scriptPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Printf("running %s", scriptPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run plugins/download.sh: %w", err)
	}

	return nil
}

func (m *Manager) syncData() error {
	pluginsSrc := filepath.Join(m.cfg.RepoPath, "plugins")
	worldsSrc := filepath.Join(m.cfg.RepoPath, "bedwars_worlds")

	pluginsDst := filepath.Join(m.cfg.DataDir, "plugins")
	worldsDst := filepath.Join(m.cfg.DataDir, "bedwars_worlds")

	for _, path := range []string{pluginsSrc, worldsSrc} {
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("expected directory missing: %s: %w", path, err)
		}
	}

	if err := os.MkdirAll(m.cfg.DataDir, 0o755); err != nil {
		return fmt.Errorf("ensure data dir: %w", err)
	}

	if err := copyDir(pluginsSrc, pluginsDst); err != nil {
		return fmt.Errorf("copy plugins: %w", err)
	}

	if err := copyDir(worldsSrc, worldsDst); err != nil {
		return fmt.Errorf("copy bedwars worlds: %w", err)
	}

	if m.cfg.HasPluginsUID && runtime.GOOS == "linux" {
		if err := chownRecursive(pluginsDst, m.cfg.PluginsUID); err != nil {
			return fmt.Errorf("chown plugins: %w", err)
		}
		if err := chownRecursive(worldsDst, m.cfg.PluginsUID); err != nil {
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

func (m *Manager) restartWithCountdown() error {
	address := fmt.Sprintf("%s:%d", m.cfg.RconHost, m.cfg.RconPort)
	client, err := rcon.Dial(address, m.cfg.RconPassword)
	if err != nil {
		return fmt.Errorf("connect to rcon: %w", err)
	}
	defer client.Close()

	announce := func(msg string) error {
		_, err := client.Execute(fmt.Sprintf("say %s", msg))
		return err
	}

	if err := announce(m.msg.RestartSoon); err != nil {
		return fmt.Errorf("announce restart: %w", err)
	}

	time.Sleep(m.cfg.CountdownWait)

	for i := 10; i >= 1; i-- {
		if err := announce(fmt.Sprintf(m.msg.Countdown, i)); err != nil {
			return fmt.Errorf("countdown announce: %w", err)
		}
		time.Sleep(1 * time.Second)
	}

	if _, err := client.Execute(m.cfg.RestartCmd); err != nil {
		return fmt.Errorf("send restart command: %w", err)
	}

	return nil
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

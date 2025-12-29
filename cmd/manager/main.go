package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/gorcon/rcon"
)

type config struct {
	repoURL       string
	repoBranch    string
	repoPath      string
	dataDir       string
	httpAddr      string
	rconHost      string
	rconPort      int
	rconPassword  string
	restartCmd    string
	countdownWait time.Duration
}

type manager struct {
	cfg config
	mu  sync.Mutex
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	mgr := &manager{cfg: cfg}

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	http.HandleFunc("/update", mgr.updateHandler)

	log.Printf("listening on %s", cfg.httpAddr)
	if err := http.ListenAndServe(cfg.httpAddr, nil); err != nil {
		log.Fatalf("http server failed: %v", err)
	}
}

func loadConfig() (config, error) {
	repoURL := os.Getenv("REPO_URL")
	if repoURL == "" {
		return config{}, errors.New("REPO_URL is required")
	}

	repoBranch := os.Getenv("REPO_BRANCH")
	if repoBranch == "" {
		repoBranch = "main"
	}

	repoPath := os.Getenv("REPO_PATH")
	if repoPath == "" {
		repoPath = "/tmp/plugin-repo"
	}

	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	rconHost := os.Getenv("RCON_HOST")
	if rconHost == "" {
		return config{}, errors.New("RCON_HOST is required")
	}

	rconPortStr := os.Getenv("RCON_PORT")
	if rconPortStr == "" {
		return config{}, errors.New("RCON_PORT is required")
	}
	rconPort, err := strconv.Atoi(rconPortStr)
	if err != nil {
		return config{}, fmt.Errorf("invalid RCON_PORT: %w", err)
	}

	rconPassword := os.Getenv("RCON_PASSWORD")
	if rconPassword == "" {
		return config{}, errors.New("RCON_PASSWORD is required")
	}

	restartCmd := os.Getenv("RCON_RESTART_COMMAND")
	if restartCmd == "" {
		restartCmd = "restart"
	}

	return config{
		repoURL:       repoURL,
		repoBranch:    repoBranch,
		repoPath:      repoPath,
		dataDir:       dataDir,
		httpAddr:      ":" + port,
		rconHost:      rconHost,
		rconPort:      rconPort,
		rconPassword:  rconPassword,
		restartCmd:    restartCmd,
		countdownWait: 50 * time.Second,
	}, nil
}

func (m *manager) updateHandler(w http.ResponseWriter, r *http.Request) {
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

func (m *manager) syncRepo() error {
	if _, err := os.Stat(m.cfg.repoPath); errors.Is(err, os.ErrNotExist) {
		log.Printf("cloning repo %s into %s (branch %s)", m.cfg.repoURL, m.cfg.repoPath, m.cfg.repoBranch)
		_, err := git.PlainClone(m.cfg.repoPath, false, &git.CloneOptions{
			URL:           m.cfg.repoURL,
			ReferenceName: plumbing.NewBranchReferenceName(m.cfg.repoBranch),
			SingleBranch:  true,
			Depth:         1,
		})
		return err
	}

	repo, err := git.PlainOpen(m.cfg.repoPath)
	if err != nil {
		return fmt.Errorf("open repo: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("open worktree: %w", err)
	}

	log.Printf("pulling latest changes (branch %s)", m.cfg.repoBranch)
	err = worktree.Pull(&git.PullOptions{
		RemoteName:    "origin",
		ReferenceName: plumbing.NewBranchReferenceName(m.cfg.repoBranch),
		SingleBranch:  true,
		Force:         true,
	})
	if errors.Is(err, git.NoErrAlreadyUpToDate) {
		return nil
	}

	return err
}

func (m *manager) syncData() error {
	pluginsSrc := filepath.Join(m.cfg.repoPath, "plugins")
	worldsSrc := filepath.Join(m.cfg.repoPath, "bedwars_worlds")

	pluginsDst := filepath.Join(m.cfg.dataDir, "plugins")
	worldsDst := filepath.Join(m.cfg.dataDir, "bedwars_worlds")

	for _, path := range []string{pluginsSrc, worldsSrc} {
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("expected directory missing: %s: %w", path, err)
		}
	}

	if err := os.MkdirAll(m.cfg.dataDir, 0o755); err != nil {
		return fmt.Errorf("ensure data dir: %w", err)
	}

	if err := copyDir(pluginsSrc, pluginsDst); err != nil {
		return fmt.Errorf("copy plugins: %w", err)
	}

	if err := copyDir(worldsSrc, worldsDst); err != nil {
		return fmt.Errorf("copy bedwars worlds: %w", err)
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

func (m *manager) restartWithCountdown() error {
	address := fmt.Sprintf("%s:%d", m.cfg.rconHost, m.cfg.rconPort)
	client, err := rcon.Dial(address, m.cfg.rconPassword)
	if err != nil {
		return fmt.Errorf("connect to rcon: %w", err)
	}
	defer client.Close()

	announce := func(msg string) error {
		_, err := client.Execute(fmt.Sprintf("say %s", msg))
		return err
	}

	if err := announce("Restarting in 60 seconds"); err != nil {
		return fmt.Errorf("announce restart: %w", err)
	}

	time.Sleep(m.cfg.countdownWait)

	for i := 10; i >= 1; i-- {
		if err := announce(fmt.Sprintf("Restarting in %d", i)); err != nil {
			return fmt.Errorf("countdown announce: %w", err)
		}
		time.Sleep(1 * time.Second)
	}

	if _, err := client.Execute(m.cfg.restartCmd); err != nil {
		return fmt.Errorf("send restart command: %w", err)
	}

	return nil
}

package manager

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func (m *Manager) runPluginDownload() error {
	scriptPath := filepath.Join(m.Cfg.RepoPath, "plugins", "download.sh")
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

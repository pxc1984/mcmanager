package manager

import (
	"os"
	"path/filepath"
	"testing"

	"mcmanager/internal/config"
)

func TestRunPluginDownloadDisabled(t *testing.T) {
	mgr := New(config.Config{
		RepoPath:        t.TempDir(),
		PluginsDownload: false,
	})

	if err := mgr.runPluginDownload(); err != nil {
		t.Fatalf("expected nil error when disabled, got %v", err)
	}
}

func TestRunPluginDownloadMissingScript(t *testing.T) {
	mgr := New(config.Config{
		RepoPath:        t.TempDir(),
		PluginsDownload: true,
	})

	if err := mgr.runPluginDownload(); err == nil {
		t.Fatalf("expected error when script missing")
	}
}

func TestRunPluginDownloadRuns(t *testing.T) {
	repo := t.TempDir()
	plugins := filepath.Join(repo, "plugins")
	if err := os.MkdirAll(plugins, 0o755); err != nil {
		t.Fatalf("mkdir plugins: %v", err)
	}

	script := filepath.Join(plugins, "download.sh")
	content := "#!/usr/bin/env bash\n" +
		"echo ok > ran.txt\n"
	if err := os.WriteFile(script, []byte(content), 0o755); err != nil {
		t.Fatalf("write script: %v", err)
	}

	mgr := New(config.Config{
		RepoPath:        repo,
		PluginsDownload: true,
	})

	if err := mgr.runPluginDownload(); err != nil {
		t.Fatalf("run plugin download: %v", err)
	}

	if _, err := os.Stat(filepath.Join(plugins, "ran.txt")); err != nil {
		t.Fatalf("expected script to run: %v", err)
	}
}

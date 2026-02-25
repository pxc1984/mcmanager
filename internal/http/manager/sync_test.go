package manager

import (
	"os"
	"path/filepath"
	"testing"

	"mcmanager/internal/config"
)

func TestParseDirList(t *testing.T) {
	got := parseDirList("a, b , ,.,a")
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("unexpected parse result: %#v", got)
	}
}

func TestSyncDataNoDirs(t *testing.T) {
	mgr := New(config.Config{
		CopyDirs: "",
		SkipDirs: "",
		RepoPath: t.TempDir(),
		DataDir:  t.TempDir(),
	})

	if err := mgr.syncData(); err == nil {
		t.Fatalf("expected error when no directories are configured")
	}
}

func TestSyncDataCopies(t *testing.T) {
	repo := t.TempDir()
	data := t.TempDir()

	worldDir := filepath.Join(repo, "world")
	if err := os.MkdirAll(worldDir, 0o755); err != nil {
		t.Fatalf("mkdir world: %v", err)
	}
	if err := os.WriteFile(filepath.Join(worldDir, "level.dat"), []byte("ok"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	mgr := New(config.Config{
		RepoPath: repo,
		DataDir:  data,
		CopyDirs: "world,plugins",
		SkipDirs: "plugins",
	})

	if err := mgr.syncData(); err != nil {
		t.Fatalf("sync data: %v", err)
	}

	if _, err := os.Stat(filepath.Join(data, "world", "level.dat")); err != nil {
		t.Fatalf("expected file copied: %v", err)
	}
}

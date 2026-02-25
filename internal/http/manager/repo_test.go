package manager

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"mcmanager/internal/config"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func TestSyncRepoCloneAndPull(t *testing.T) {
	origin := t.TempDir()

	repo, err := git.PlainInit(origin, false)
	if err != nil {
		t.Fatalf("init repo: %v", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		t.Fatalf("worktree: %v", err)
	}

	filePath := filepath.Join(origin, "README.md")
	if err := os.WriteFile(filePath, []byte("hello"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	if _, err := worktree.Add("README.md"); err != nil {
		t.Fatalf("add file: %v", err)
	}

	_, err = worktree.Commit("initial", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "tester",
			Email: "tester@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		t.Fatalf("commit: %v", err)
	}

	clonePath := filepath.Join(t.TempDir(), "clone")
	mgr := New(config.Config{
		RepoURL:    origin,
		RepoPath:   clonePath,
		RepoBranch: "master",
	})

	if err := mgr.syncRepo(); err != nil {
		t.Fatalf("clone: %v", err)
	}

	if err := mgr.syncRepo(); err != nil {
		t.Fatalf("pull: %v", err)
	}
}

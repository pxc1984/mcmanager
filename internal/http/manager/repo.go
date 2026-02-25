package manager

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func (m *Manager) syncRepo() error {
	if _, err := os.Stat(m.Cfg.RepoPath); errors.Is(err, os.ErrNotExist) {
		log.Printf("cloning repo %s into %s (branch %s)", m.Cfg.RepoURL, m.Cfg.RepoPath, m.Cfg.RepoBranch)
		_, err := git.PlainClone(m.Cfg.RepoPath, false, &git.CloneOptions{
			URL:           m.Cfg.RepoURL,
			ReferenceName: plumbing.NewBranchReferenceName(m.Cfg.RepoBranch),
			SingleBranch:  true,
			Depth:         1,
		})
		return err
	}

	repo, err := git.PlainOpen(m.Cfg.RepoPath)
	if err != nil {
		return fmt.Errorf("open repo: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("open worktree: %w", err)
	}

	log.Printf("pulling latest changes (branch %s)", m.Cfg.RepoBranch)
	err = worktree.Pull(&git.PullOptions{
		RemoteName:    "origin",
		ReferenceName: plumbing.NewBranchReferenceName(m.Cfg.RepoBranch),
		SingleBranch:  true,
		Force:         true,
	})
	if errors.Is(err, git.NoErrAlreadyUpToDate) {
		return nil
	}

	return err
}

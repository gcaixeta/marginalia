package storage

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/gcaixeta/marginalia/internal/config"
)

type GitSync struct {
	dataDir string
	repo    string
	remote  string
	branch  string
	pullWg  sync.WaitGroup
	pullErr error
}

func NewGitSync(cfg *config.BackupConfig) (*GitSync, error) {
	if cfg.Provider != "git" || cfg.Git.Repo == "" {
		return nil, nil
	}

	dataDir, err := DataDir()
	if err != nil {
		return nil, err
	}

	remote := cfg.Git.Remote
	if remote == "" {
		remote = "origin"
	}

	branch := cfg.Git.Branch
	if branch == "" {
		branch = "main"
	}

	return &GitSync{
		dataDir: dataDir,
		repo:    cfg.Git.Repo,
		remote:  remote,
		branch:  branch,
	}, nil
}

func (g *GitSync) run(args ...string) error {
	cmdArgs := append([]string{"-C", g.dataDir}, args...)
	cmd := exec.Command("git", cmdArgs...)
	return cmd.Run()
}

func (g *GitSync) runSilent(args ...string) error {
	cmdArgs := append([]string{"-C", g.dataDir}, args...)
	cmd := exec.Command("git", cmdArgs...)
	cmd.Stdin = nil
	cmd.Env = append(os.Environ(),
		"GIT_TERMINAL_PROMPT=0",
		"GIT_SSH_COMMAND=ssh -o BatchMode=yes",
	)
	return cmd.Run()
}

func (g *GitSync) Synchronize() error {
	g.pullWg.Go(func() {
		if _, err := os.Stat(g.dataDir + "/.git"); os.IsNotExist(err) {
			if err := g.runSilent("init"); err != nil {
				g.pullErr = fmt.Errorf("git init: %w", err)
				return
			}
			if err := g.runSilent("remote", "add", g.remote, g.repo); err != nil {
				g.pullErr = fmt.Errorf("git remote add: %w", err)
				return
			}
		}

		_ = g.runSilent("pull", g.remote, g.branch)
	})
	return nil
}

func (g *GitSync) CommitAndPush(message string) error {
	g.pullWg.Wait()
	if g.pullErr != nil {
		fmt.Printf("Warning: git pull failed: %v\n", g.pullErr)
	}

	if err := g.run("add", "-A"); err != nil {
		return fmt.Errorf("git add: %w", err)
	}

	cmd := exec.Command("git", "-C", g.dataDir, "status", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("git status: %w", err)
	}
	if len(bytes.TrimSpace(out)) == 0 {
		return nil
	}

	if err := g.run("commit", "-m", message); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}

	if err := g.run("push", g.remote, g.branch); err != nil {
		return fmt.Errorf("git push: %w", err)
	}

	fmt.Println("↑ synced")
	return nil
}

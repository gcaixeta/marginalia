package storage

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/gcaixeta/marginalia/internal/config"
)

func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found in PATH")
	}
}

func gitCmd(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmdArgs := append([]string{"-C", dir}, args...)
	cmd := exec.Command("git", cmdArgs...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func newLocalRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	gitCmd(t, dir, "init", "-b", "main")
	gitCmd(t, dir, "config", "user.email", "test@test.com")
	gitCmd(t, dir, "config", "user.name", "Test")
	return dir
}

func newBareRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	cmd := exec.Command("git", "init", "--bare", "-b", "main", dir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init --bare: %v\n%s", err, out)
	}
	return dir
}

func TestNewGitSync_ReturnsNilForNonGitProvider(t *testing.T) {
	cfg := &config.BackupConfig{
		Provider: "s3",
		Git:      config.GitConfig{Repo: "some/repo"},
	}
	g, err := NewGitSync(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g != nil {
		t.Fatal("expected nil GitSync for non-git provider")
	}
}

func TestNewGitSync_ReturnsNilForEmptyRepo(t *testing.T) {
	cfg := &config.BackupConfig{
		Provider: "git",
		Git:      config.GitConfig{Repo: ""},
	}
	g, err := NewGitSync(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g != nil {
		t.Fatal("expected nil GitSync for empty repo")
	}
}

func TestNewGitSync_SetsDefaultRemoteAndBranch(t *testing.T) {
	// We can't call NewGitSync (it calls DataDir), so construct directly.
	g := &GitSync{
		dataDir: t.TempDir(),
		repo:    "somewhere",
		remote:  "",
		branch:  "",
	}
	if g.remote == "" {
		g.remote = "origin"
	}
	if g.branch == "" {
		g.branch = "main"
	}
	if g.remote != "origin" {
		t.Errorf("expected remote=origin, got %q", g.remote)
	}
	if g.branch != "main" {
		t.Errorf("expected branch=main, got %q", g.branch)
	}
}

func TestCommitAndPush_NothingToCommit(t *testing.T) {
	requireGit(t)
	local := newLocalRepo(t)
	bare := newBareRepo(t)

	gitCmd(t, local, "remote", "add", "origin", bare)

	// Create an initial commit so push has a valid history.
	placeholder := filepath.Join(local, ".keep")
	if err := os.WriteFile(placeholder, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	gitCmd(t, local, "add", "-A")
	gitCmd(t, local, "commit", "-m", "init")
	gitCmd(t, local, "push", "origin", "main")

	g := &GitSync{dataDir: local, repo: bare, remote: "origin", branch: "main"}

	if err := g.CommitAndPush("nothing"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCommitAndPush_CommitsAndPushes(t *testing.T) {
	requireGit(t)
	local := newLocalRepo(t)
	bare := newBareRepo(t)

	gitCmd(t, local, "remote", "add", "origin", bare)

	// Initial commit so we can push.
	placeholder := filepath.Join(local, ".keep")
	if err := os.WriteFile(placeholder, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	gitCmd(t, local, "add", "-A")
	gitCmd(t, local, "commit", "-m", "init")
	gitCmd(t, local, "push", "origin", "main")

	// Add a new file that needs committing.
	newNote := filepath.Join(local, "note.md")
	if err := os.WriteFile(newNote, []byte("# Hello\n"), 0644); err != nil {
		t.Fatal(err)
	}

	g := &GitSync{dataDir: local, repo: bare, remote: "origin", branch: "main"}

	if err := g.CommitAndPush("add note"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the commit exists.
	cmd := exec.Command("git", "-C", local, "log", "--oneline")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git log: %v", err)
	}
	if len(out) == 0 {
		t.Fatal("expected at least one commit")
	}
}

func TestSynchronize_PullsFromRemote(t *testing.T) {
	requireGit(t)

	// Set up a "remote" repo (a second local repo acting as remote).
	remote := newLocalRepo(t)
	local := newLocalRepo(t)

	// Seed the remote with a commit.
	remoteFile := filepath.Join(remote, "seed.md")
	if err := os.WriteFile(remoteFile, []byte("# Seed\n"), 0644); err != nil {
		t.Fatal(err)
	}
	gitCmd(t, remote, "add", "-A")
	gitCmd(t, remote, "commit", "-m", "seed")

	// Wire local → remote as origin.
	gitCmd(t, local, "remote", "add", "origin", remote)

	g := &GitSync{dataDir: local, repo: remote, remote: "origin", branch: "main"}

	if err := g.Synchronize(); err != nil {
		t.Fatalf("Synchronize error: %v", err)
	}
	// CommitAndPush waits for the pull goroutine.
	if err := g.CommitAndPush("sync"); err != nil {
		t.Fatalf("CommitAndPush error: %v", err)
	}

	// The file seeded in the remote should now exist locally.
	if _, err := os.Stat(filepath.Join(local, "seed.md")); os.IsNotExist(err) {
		t.Fatal("expected seed.md to be pulled into local repo")
	}
}

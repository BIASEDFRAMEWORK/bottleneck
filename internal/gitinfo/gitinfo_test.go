package gitinfo

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestSnapshotDoesNotFailOutsideGitRepo(t *testing.T) {
	info := Detect(t.TempDir())
	if info.Commit != "" || info.Branch != "" || info.Dirty != nil {
		t.Fatalf("expected empty git info outside repo, got %#v", info)
	}
}

func TestSnapshotIncludesGitMetadataWhenAvailable(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	root := t.TempDir()
	runGitCommand(t, root, "init")
	runGitCommand(t, root, "config", "user.email", "bottleneck@example.test")
	runGitCommand(t, root, "config", "user.name", "Bottleneck Test")
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# test\n"), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}
	runGitCommand(t, root, "add", "README.md")
	runGitCommand(t, root, "commit", "-m", "initial")

	info := Detect(root)
	if info.Commit == "" {
		t.Fatalf("expected git commit metadata, got %#v", info)
	}
	if info.Branch == "" {
		t.Fatalf("expected git branch metadata, got %#v", info)
	}
	if info.Dirty == nil || *info.Dirty {
		t.Fatalf("expected clean git dirty metadata, got %#v", info)
	}

	if err := os.WriteFile(filepath.Join(root, "dirty.txt"), []byte("dirty\n"), 0o644); err != nil {
		t.Fatalf("write dirty file: %v", err)
	}
	dirtyInfo := Detect(root)
	if dirtyInfo.Dirty == nil || !*dirtyInfo.Dirty {
		t.Fatalf("expected dirty git metadata after untracked file, got %#v", dirtyInfo)
	}
}

func runGitCommand(t *testing.T, root string, args ...string) {
	t.Helper()
	command := exec.Command("git", args...)
	command.Dir = root
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, string(output))
	}
}

package gitinfo

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

type Info struct {
	Commit string `json:"commit"`
	Branch string `json:"branch"`
	Dirty  *bool  `json:"dirty"`
}

func Detect(root string) Info {
	if root == "" {
		root = "."
	}
	if _, ok := runGit(root, "rev-parse", "--is-inside-work-tree"); !ok {
		return Info{}
	}

	info := Info{}
	if commit, ok := runGit(root, "rev-parse", "--short", "HEAD"); ok {
		info.Commit = commit
	}
	if branch, ok := runGit(root, "rev-parse", "--abbrev-ref", "HEAD"); ok {
		info.Branch = branch
	}
	if status, ok := runGit(root, "status", "--porcelain"); ok {
		dirty := strings.TrimSpace(status) != ""
		info.Dirty = &dirty
	}
	return info
}

func runGit(root string, args ...string) (string, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	commandArgs := append([]string{"-C", root}, args...)
	command := exec.CommandContext(ctx, "git", commandArgs...)
	output, err := command.Output()
	if err != nil {
		return "", false
	}
	return strings.TrimSpace(string(output)), true
}

package prrisk

import (
	"path/filepath"
	"strconv"
	"strings"

	"bottleneck/internal/githubactions"
)

const (
	LevelWarning = "WARNING"
	LevelUnknown = "UNKNOWN"

	LargeChangedFilesThreshold = 25
	LargeDiffThreshold         = 1000
)

type Signal struct {
	ID       string `json:"id"`
	Level    string `json:"level"`
	Message  string `json:"message"`
	Evidence string `json:"evidence,omitempty"`
}

func Assess(metadata githubactions.Metadata) []Signal {
	if metadata.PullRequest == nil {
		return nil
	}

	pr := metadata.PullRequest
	var signals []Signal

	if pr.ChangedFiles != nil && *pr.ChangedFiles > LargeChangedFilesThreshold {
		signals = append(signals, Signal{
			ID:       "large_changed_file_count",
			Level:    LevelWarning,
			Message:  "Large pull request by changed file count.",
			Evidence: pluralCount(*pr.ChangedFiles, "changed file"),
		})
	}

	if pr.Additions != nil && pr.Deletions != nil {
		total := *pr.Additions + *pr.Deletions
		if total > LargeDiffThreshold {
			signals = append(signals, Signal{
				ID:       "large_diff_size",
				Level:    LevelWarning,
				Message:  "Large pull request by additions plus deletions.",
				Evidence: pluralCount(total, "changed line"),
			})
		}
	}

	if pr.Draft != nil && *pr.Draft {
		signals = append(signals, Signal{
			ID:      "draft_pull_request",
			Level:   LevelWarning,
			Message: "Draft pull request should not be treated as release-ready.",
		})
	}

	if len(pr.RequestedReviewers) == 0 {
		signals = append(signals, Signal{
			ID:      "missing_requested_reviewers",
			Level:   LevelWarning,
			Message: "No requested reviewers were found for this pull request.",
		})
	}

	for _, label := range pr.Labels {
		if isAIAssistedLabel(label) {
			signals = append(signals, Signal{
				ID:       "ai_assisted_label",
				Level:    LevelWarning,
				Message:  "Pull request is labeled as AI-generated or AI-assisted.",
				Evidence: label,
			})
			break
		}
	}

	if len(pr.ChangedFileNames) > 0 {
		if hasReleaseRelevantSourceChange(pr.ChangedFileNames) && !hasArtifactChange(pr.ChangedFileNames) {
			signals = append(signals, Signal{
				ID:      "source_without_evidence_artifacts",
				Level:   LevelWarning,
				Message: "Release-relevant source files changed without matching bottleneck evidence artifact changes.",
			})
		}
	} else {
		signals = append(signals, Signal{
			ID:      "changed_files_unavailable",
			Level:   LevelUnknown,
			Message: "Changed file names were not available; source-to-evidence artifact risk was not assessed.",
		})
	}

	if pr.ApprovalCount != nil && *pr.ApprovalCount == 0 && !(pr.Draft != nil && *pr.Draft) {
		signals = append(signals, Signal{
			ID:      "missing_approval",
			Level:   LevelWarning,
			Message: "No approvals were found for this non-draft pull request.",
		})
	}

	if len(pr.PendingReviewers) > 0 {
		signals = append(signals, Signal{
			ID:       "pending_reviewers",
			Level:    LevelWarning,
			Message:  "Requested reviewers are still pending.",
			Evidence: strings.Join(pr.PendingReviewers, ", "),
		})
	}

	if len(pr.FailedChecks) > 0 {
		signals = append(signals, Signal{
			ID:       "failed_check_runs",
			Level:    LevelWarning,
			Message:  "One or more GitHub check runs failed.",
			Evidence: strings.Join(pr.FailedChecks, ", "),
		})
	}

	return signals
}

func isAIAssistedLabel(label string) bool {
	normalized := strings.ToLower(label)
	keywords := []string{"ai-generated", "ai-assisted", "copilot", "codex"}
	for _, keyword := range keywords {
		if strings.Contains(normalized, keyword) {
			return true
		}
	}
	return false
}

func hasReleaseRelevantSourceChange(paths []string) bool {
	for _, path := range paths {
		normalized := filepath.ToSlash(path)
		if strings.HasPrefix(normalized, "cmd/") ||
			strings.HasPrefix(normalized, "internal/") ||
			strings.HasPrefix(normalized, "pkg/") ||
			strings.HasPrefix(normalized, "src/") ||
			strings.HasPrefix(normalized, "app/") ||
			strings.HasPrefix(normalized, "services/") {
			return true
		}

		switch strings.ToLower(filepath.Ext(normalized)) {
		case ".go", ".ts", ".tsx", ".js", ".jsx", ".py":
			return true
		}
	}
	return false
}

func hasArtifactChange(paths []string) bool {
	for _, path := range paths {
		if strings.HasPrefix(filepath.ToSlash(path), "bottleneck/") {
			return true
		}
	}
	return false
}

func pluralCount(count int, label string) string {
	if count == 1 {
		return "1 " + label
	}
	return strconv.Itoa(count) + " " + label + "s"
}

package prrisk

import (
	"testing"

	"bottleneck/internal/githubactions"
)

func TestAssessFlagsLargeChangedFileCount(t *testing.T) {
	count := 26
	signals := Assess(metadataWithPR(githubactions.PullRequestMetadata{
		ChangedFiles: &count,
	}))

	if !hasSignal(signals, "large_changed_file_count") {
		t.Fatalf("expected large file count signal, got %#v", signals)
	}
}

func TestAssessFlagsLargeDiffSize(t *testing.T) {
	additions := 900
	deletions := 101
	signals := Assess(metadataWithPR(githubactions.PullRequestMetadata{
		Additions: &additions,
		Deletions: &deletions,
	}))

	if !hasSignal(signals, "large_diff_size") {
		t.Fatalf("expected large diff signal, got %#v", signals)
	}
}

func TestAssessFlagsAIAssistedLabels(t *testing.T) {
	signals := Assess(metadataWithPR(githubactions.PullRequestMetadata{
		Labels: []string{"codex", "release"},
	}))

	if !hasSignal(signals, "ai_assisted_label") {
		t.Fatalf("expected AI-assisted label signal, got %#v", signals)
	}
}

func TestAssessFlagsSourceChangesWithoutEvidenceArtifacts(t *testing.T) {
	signals := Assess(metadataWithPR(githubactions.PullRequestMetadata{
		ChangedFileNames: []string{"cmd/scorecard.go", "internal/validator/engine.go"},
	}))

	if !hasSignal(signals, "source_without_evidence_artifacts") {
		t.Fatalf("expected source without artifact signal, got %#v", signals)
	}
}

func TestAssessDoesNotFlagSourceChangesWhenEvidenceArtifactsChange(t *testing.T) {
	signals := Assess(metadataWithPR(githubactions.PullRequestMetadata{
		ChangedFileNames: []string{"cmd/scorecard.go", "bottleneck/intent/intent.md"},
	}))

	if hasSignal(signals, "source_without_evidence_artifacts") {
		t.Fatalf("did not expect source without artifact signal, got %#v", signals)
	}
}

func TestAssessFlagsReviewAndCheckSignals(t *testing.T) {
	draft := false
	approvals := 0
	signals := Assess(metadataWithPR(githubactions.PullRequestMetadata{
		Draft:              &draft,
		RequestedReviewers: []string{"octocat"},
		ApprovalCount:      &approvals,
		PendingReviewers:   []string{"octocat"},
		FailedChecks:       []string{"unit"},
	}))

	for _, id := range []string{"missing_approval", "pending_reviewers", "failed_check_runs"} {
		if !hasSignal(signals, id) {
			t.Fatalf("expected %s signal, got %#v", id, signals)
		}
	}
}

func metadataWithPR(pr githubactions.PullRequestMetadata) githubactions.Metadata {
	return githubactions.Metadata{
		Detected:    true,
		Repository:  "acme/widgets",
		PullRequest: &pr,
	}
}

func hasSignal(signals []Signal, id string) bool {
	for _, signal := range signals {
		if signal.ID == id {
			return true
		}
	}
	return false
}

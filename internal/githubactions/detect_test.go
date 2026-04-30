package githubactions

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestDetectReturnsFalseOutsideGitHubActions(t *testing.T) {
	metadata := Detect(map[string]string{})

	if metadata.Detected {
		t.Fatal("expected GitHub Actions detection to be false")
	}
	if metadata.PullRequest != nil {
		t.Fatalf("expected no pull request metadata, got %#v", metadata.PullRequest)
	}
}

func TestDetectParsesPullRequestEventPayload(t *testing.T) {
	eventPath := writeEventPayload(t, `{
  "pull_request": {
    "number": 42,
    "title": "Add release gate",
    "html_url": "https://github.com/acme/widgets/pull/42",
    "user": {"login": "octocat"},
    "base": {"ref": "main"},
    "head": {"ref": "feature/release-gate"},
    "labels": [{"name": "ai-assisted"}],
    "requested_reviewers": [{"login": "hubot"}],
    "changed_files": 18,
    "additions": 900,
    "deletions": 120,
    "draft": false,
    "files": [{"filename": "cmd/scorecard.go"}]
  }
}`)

	metadata := Detect(map[string]string{
		"GITHUB_ACTIONS":    "true",
		"GITHUB_EVENT_NAME": "pull_request",
		"GITHUB_EVENT_PATH": eventPath,
		"GITHUB_REPOSITORY": "acme/widgets",
		"GITHUB_SHA":        "abc123",
		"GITHUB_RUN_ID":     "123456",
		"GITHUB_SERVER_URL": "https://github.com",
	})

	if !metadata.Detected {
		t.Fatal("expected GitHub Actions detection")
	}
	if metadata.PullRequest == nil {
		t.Fatal("expected pull request metadata")
	}
	if metadata.PullRequest.Number != 42 {
		t.Fatalf("expected PR number 42, got %d", metadata.PullRequest.Number)
	}
	if metadata.PullRequest.Title != "Add release gate" {
		t.Fatalf("unexpected title %q", metadata.PullRequest.Title)
	}
	if metadata.PullRequest.Author != "octocat" {
		t.Fatalf("unexpected author %q", metadata.PullRequest.Author)
	}
	if len(metadata.PullRequest.Labels) != 1 || metadata.PullRequest.Labels[0] != "ai-assisted" {
		t.Fatalf("unexpected labels %#v", metadata.PullRequest.Labels)
	}
	if len(metadata.PullRequest.RequestedReviewers) != 1 || metadata.PullRequest.RequestedReviewers[0] != "hubot" {
		t.Fatalf("unexpected reviewers %#v", metadata.PullRequest.RequestedReviewers)
	}
	if metadata.PullRequest.ChangedFiles == nil || *metadata.PullRequest.ChangedFiles != 18 {
		t.Fatalf("expected changed file count 18, got %#v", metadata.PullRequest.ChangedFiles)
	}
	if metadata.PullRequest.Draft == nil || *metadata.PullRequest.Draft {
		t.Fatalf("expected non-draft PR, got %#v", metadata.PullRequest.Draft)
	}
	if len(metadata.PullRequest.ChangedFileNames) != 1 || metadata.PullRequest.ChangedFileNames[0] != "cmd/scorecard.go" {
		t.Fatalf("unexpected changed file names %#v", metadata.PullRequest.ChangedFileNames)
	}
}

func TestDetectHandlesMalformedEventPayloadGracefully(t *testing.T) {
	eventPath := writeEventPayload(t, `{this is not json`)

	metadata := Detect(map[string]string{
		"GITHUB_ACTIONS":    "true",
		"GITHUB_EVENT_NAME": "pull_request",
		"GITHUB_EVENT_PATH": eventPath,
	})

	if !metadata.Detected {
		t.Fatal("expected GitHub Actions detection")
	}
	if metadata.PullRequest != nil {
		t.Fatalf("expected no parsed PR metadata, got %#v", metadata.PullRequest)
	}
	if len(metadata.Warnings) == 0 {
		t.Fatal("expected parse warning")
	}
}

func TestEnrichUsesClientData(t *testing.T) {
	approvalCount := 0
	metadata := Metadata{
		Detected:   true,
		Repository: "acme/widgets",
		SHA:        "abc123",
		PullRequest: &PullRequestMetadata{
			Number:             42,
			RequestedReviewers: []string{"octocat", "hubot"},
			ApprovalCount:      &approvalCount,
		},
	}

	Enrich(context.Background(), &metadata, fakeClient{
		files: []ChangedFile{
			{Filename: "cmd/scorecard.go", Additions: 10, Deletions: 2},
			{Filename: "bottleneck/intent/intent.md", Additions: 3, Deletions: 0},
		},
		reviews: []Review{
			{User: "octocat", State: "APPROVED"},
		},
		checks: []CheckRun{
			{Name: "unit", Conclusion: "success"},
			{Name: "lint", Conclusion: "failure"},
		},
	})

	if metadata.PullRequest.ChangedFiles == nil || *metadata.PullRequest.ChangedFiles != 2 {
		t.Fatalf("expected changed file count 2, got %#v", metadata.PullRequest.ChangedFiles)
	}
	if metadata.PullRequest.Additions == nil || *metadata.PullRequest.Additions != 13 {
		t.Fatalf("expected additions 13, got %#v", metadata.PullRequest.Additions)
	}
	if metadata.PullRequest.Deletions == nil || *metadata.PullRequest.Deletions != 2 {
		t.Fatalf("expected deletions 2, got %#v", metadata.PullRequest.Deletions)
	}
	if metadata.PullRequest.ApprovalCount == nil || *metadata.PullRequest.ApprovalCount != 1 {
		t.Fatalf("expected approval count 1, got %#v", metadata.PullRequest.ApprovalCount)
	}
	if len(metadata.PullRequest.PendingReviewers) != 1 || metadata.PullRequest.PendingReviewers[0] != "hubot" {
		t.Fatalf("unexpected pending reviewers %#v", metadata.PullRequest.PendingReviewers)
	}
	if len(metadata.PullRequest.FailedChecks) != 1 || metadata.PullRequest.FailedChecks[0] != "lint" {
		t.Fatalf("unexpected failed checks %#v", metadata.PullRequest.FailedChecks)
	}
}

type fakeClient struct {
	files   []ChangedFile
	reviews []Review
	checks  []CheckRun
}

func (f fakeClient) ChangedFiles(context.Context, string, int) ([]ChangedFile, error) {
	return f.files, nil
}

func (f fakeClient) Reviews(context.Context, string, int) ([]Review, error) {
	return f.reviews, nil
}

func (f fakeClient) CheckRuns(context.Context, string, string) ([]CheckRun, error) {
	return f.checks, nil
}

func writeEventPayload(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "event.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write event payload: %v", err)
	}
	return path
}

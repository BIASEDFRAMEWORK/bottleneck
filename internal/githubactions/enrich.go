package githubactions

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type EnrichmentClient interface {
	ChangedFiles(ctx context.Context, repository string, prNumber int) ([]ChangedFile, error)
	Reviews(ctx context.Context, repository string, prNumber int) ([]Review, error)
	CheckRuns(ctx context.Context, repository string, headSHA string) ([]CheckRun, error)
}

type ChangedFile struct {
	Filename  string
	Additions int
	Deletions int
}

type Review struct {
	User  string
	State string
}

type CheckRun struct {
	Name       string
	Status     string
	Conclusion string
}

func Enrich(ctx context.Context, metadata *Metadata, client EnrichmentClient) {
	if metadata == nil || client == nil || metadata.PullRequest == nil || metadata.Repository == "" {
		return
	}

	pr := metadata.PullRequest
	if files, err := client.ChangedFiles(ctx, metadata.Repository, pr.Number); err == nil {
		applyChangedFiles(pr, files)
	} else {
		metadata.Warnings = append(metadata.Warnings, fmt.Sprintf("could not enrich changed files: %v", err))
	}

	if reviews, err := client.Reviews(ctx, metadata.Repository, pr.Number); err == nil {
		applyReviews(pr, reviews)
	} else {
		metadata.Warnings = append(metadata.Warnings, fmt.Sprintf("could not enrich reviews: %v", err))
	}

	if metadata.SHA != "" {
		if checks, err := client.CheckRuns(ctx, metadata.Repository, metadata.SHA); err == nil {
			applyCheckRuns(pr, checks)
		} else {
			metadata.Warnings = append(metadata.Warnings, fmt.Sprintf("could not enrich check runs: %v", err))
		}
	}
}

func EnrichFromEnv(ctx context.Context, metadata *Metadata, token string) {
	if metadata == nil || !metadata.Detected || token == "" || metadata.PullRequest == nil {
		return
	}

	client := NewRESTClient(token, metadata.ServerURL)
	Enrich(ctx, metadata, client)
}

type RESTClient struct {
	Token      string
	APIBaseURL string
	HTTPClient *http.Client
}

func NewRESTClient(token string, serverURL string) *RESTClient {
	return &RESTClient{
		Token:      token,
		APIBaseURL: apiBaseURL(serverURL),
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *RESTClient) ChangedFiles(ctx context.Context, repository string, prNumber int) ([]ChangedFile, error) {
	var payload []struct {
		Filename  string `json:"filename"`
		Additions int    `json:"additions"`
		Deletions int    `json:"deletions"`
	}
	if err := c.get(ctx, fmt.Sprintf("/repos/%s/pulls/%d/files?per_page=100", repository, prNumber), &payload); err != nil {
		return nil, err
	}

	files := make([]ChangedFile, 0, len(payload))
	for _, file := range payload {
		files = append(files, ChangedFile{
			Filename:  file.Filename,
			Additions: file.Additions,
			Deletions: file.Deletions,
		})
	}
	return files, nil
}

func (c *RESTClient) Reviews(ctx context.Context, repository string, prNumber int) ([]Review, error) {
	var payload []struct {
		User  userPayload `json:"user"`
		State string      `json:"state"`
	}
	if err := c.get(ctx, fmt.Sprintf("/repos/%s/pulls/%d/reviews?per_page=100", repository, prNumber), &payload); err != nil {
		return nil, err
	}

	reviews := make([]Review, 0, len(payload))
	for _, review := range payload {
		reviews = append(reviews, Review{
			User:  review.User.Login,
			State: review.State,
		})
	}
	return reviews, nil
}

func (c *RESTClient) CheckRuns(ctx context.Context, repository string, headSHA string) ([]CheckRun, error) {
	var payload struct {
		CheckRuns []struct {
			Name       string `json:"name"`
			Status     string `json:"status"`
			Conclusion string `json:"conclusion"`
		} `json:"check_runs"`
	}
	if err := c.get(ctx, fmt.Sprintf("/repos/%s/commits/%s/check-runs?per_page=100", repository, headSHA), &payload); err != nil {
		return nil, err
	}

	checks := make([]CheckRun, 0, len(payload.CheckRuns))
	for _, check := range payload.CheckRuns {
		checks = append(checks, CheckRun{
			Name:       check.Name,
			Status:     check.Status,
			Conclusion: check.Conclusion,
		})
	}
	return checks, nil
}

func (c *RESTClient) get(ctx context.Context, path string, target any) error {
	requestURL := strings.TrimRight(c.APIBaseURL, "/") + path
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if c.Token != "" {
		request.Header.Set("Authorization", "Bearer "+c.Token)
	}

	response, err := c.HTTPClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusForbidden || response.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("GitHub API permission denied: %s", response.Status)
	}
	if response.StatusCode == http.StatusTooManyRequests || response.Header.Get("X-RateLimit-Remaining") == "0" {
		return fmt.Errorf("GitHub API rate limit reached")
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("GitHub API returned %s", response.Status)
	}

	return json.NewDecoder(response.Body).Decode(target)
}

func applyChangedFiles(pr *PullRequestMetadata, files []ChangedFile) {
	names := make([]string, 0, len(files))
	additions := 0
	deletions := 0
	for _, file := range files {
		if file.Filename != "" {
			names = append(names, file.Filename)
		}
		additions += file.Additions
		deletions += file.Deletions
	}

	changedFiles := len(files)
	pr.ChangedFiles = &changedFiles
	pr.Additions = &additions
	pr.Deletions = &deletions
	pr.ChangedFileNames = names
}

func applyReviews(pr *PullRequestMetadata, reviews []Review) {
	approvedBy := map[string]bool{}
	changesRequestedBy := map[string]bool{}
	for _, review := range reviews {
		user := strings.TrimSpace(review.User)
		if user == "" {
			continue
		}
		switch strings.ToUpper(review.State) {
		case "APPROVED":
			approvedBy[user] = true
			delete(changesRequestedBy, user)
		case "CHANGES_REQUESTED":
			changesRequestedBy[user] = true
			delete(approvedBy, user)
		}
	}

	approvalCount := len(approvedBy)
	pr.ApprovalCount = &approvalCount

	var pending []string
	for _, reviewer := range pr.RequestedReviewers {
		if !approvedBy[reviewer] && !changesRequestedBy[reviewer] {
			pending = append(pending, reviewer)
		}
	}
	pr.PendingReviewers = pending
}

func applyCheckRuns(pr *PullRequestMetadata, checks []CheckRun) {
	var failed []string
	for _, check := range checks {
		conclusion := strings.ToLower(check.Conclusion)
		if conclusion == "failure" || conclusion == "timed_out" || conclusion == "cancelled" || conclusion == "action_required" {
			failed = append(failed, check.Name)
		}
	}
	pr.FailedChecks = failed
}

func apiBaseURL(serverURL string) string {
	if serverURL == "" {
		return "https://api.github.com"
	}

	parsed, err := url.Parse(serverURL)
	if err != nil {
		return "https://api.github.com"
	}
	if strings.EqualFold(parsed.Host, "github.com") {
		return "https://api.github.com"
	}

	return strings.TrimRight(serverURL, "/") + "/api/v3"
}

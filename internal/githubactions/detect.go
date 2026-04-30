package githubactions

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Metadata struct {
	Detected    bool                 `json:"detected"`
	EventName   string               `json:"event_name,omitempty"`
	Repository  string               `json:"repository,omitempty"`
	SHA         string               `json:"sha,omitempty"`
	Ref         string               `json:"ref,omitempty"`
	HeadRef     string               `json:"head_ref,omitempty"`
	BaseRef     string               `json:"base_ref,omitempty"`
	RunID       string               `json:"run_id,omitempty"`
	ServerURL   string               `json:"server_url,omitempty"`
	PullRequest *PullRequestMetadata `json:"pull_request,omitempty"`
	Warnings    []string             `json:"warnings,omitempty"`
}

type PullRequestMetadata struct {
	Number             int      `json:"number,omitempty"`
	Title              string   `json:"title,omitempty"`
	URL                string   `json:"url,omitempty"`
	BaseRef            string   `json:"base_ref,omitempty"`
	HeadRef            string   `json:"head_ref,omitempty"`
	Author             string   `json:"author,omitempty"`
	Labels             []string `json:"labels,omitempty"`
	RequestedReviewers []string `json:"requested_reviewers,omitempty"`
	ChangedFiles       *int     `json:"changed_files,omitempty"`
	Additions          *int     `json:"additions,omitempty"`
	Deletions          *int     `json:"deletions,omitempty"`
	Draft              *bool    `json:"draft,omitempty"`
	ChangedFileNames   []string `json:"changed_file_names,omitempty"`
	ApprovalCount      *int     `json:"approval_count,omitempty"`
	PendingReviewers   []string `json:"pending_reviewers,omitempty"`
	FailedChecks       []string `json:"failed_checks,omitempty"`
}

func DetectFromEnv() Metadata {
	return Detect(environmentMap(os.Environ()))
}

func Detect(env map[string]string) Metadata {
	detected := strings.EqualFold(env["GITHUB_ACTIONS"], "true")
	metadata := Metadata{
		Detected:   detected,
		EventName:  env["GITHUB_EVENT_NAME"],
		Repository: env["GITHUB_REPOSITORY"],
		SHA:        env["GITHUB_SHA"],
		Ref:        env["GITHUB_REF"],
		HeadRef:    env["GITHUB_HEAD_REF"],
		BaseRef:    env["GITHUB_BASE_REF"],
		RunID:      env["GITHUB_RUN_ID"],
		ServerURL:  env["GITHUB_SERVER_URL"],
	}

	if !detected {
		return Metadata{Detected: false}
	}
	if metadata.ServerURL == "" {
		metadata.ServerURL = "https://github.com"
	}

	eventPath := env["GITHUB_EVENT_PATH"]
	if eventPath == "" {
		return metadata
	}

	content, err := os.ReadFile(eventPath)
	if err != nil {
		metadata.Warnings = append(metadata.Warnings, fmt.Sprintf("could not read GitHub event payload: %v", err))
		return metadata
	}

	pr, err := parsePullRequestPayload(content)
	if err != nil {
		metadata.Warnings = append(metadata.Warnings, fmt.Sprintf("could not parse GitHub event payload: %v", err))
		return metadata
	}
	if pr != nil {
		metadata.PullRequest = pr
		if metadata.BaseRef == "" {
			metadata.BaseRef = pr.BaseRef
		}
		if metadata.HeadRef == "" {
			metadata.HeadRef = pr.HeadRef
		}
	}

	return metadata
}

type eventPayload struct {
	PullRequest *pullRequestPayload `json:"pull_request"`
}

type pullRequestPayload struct {
	Number             int                  `json:"number"`
	Title              string               `json:"title"`
	HTMLURL            string               `json:"html_url"`
	User               userPayload          `json:"user"`
	Base               refPayload           `json:"base"`
	Head               refPayload           `json:"head"`
	Labels             []labelPayload       `json:"labels"`
	RequestedReviewers []userPayload        `json:"requested_reviewers"`
	ChangedFiles       *int                 `json:"changed_files"`
	Additions          *int                 `json:"additions"`
	Deletions          *int                 `json:"deletions"`
	Draft              *bool                `json:"draft"`
	Files              []changedFilePayload `json:"files"`
}

type userPayload struct {
	Login string `json:"login"`
}

type refPayload struct {
	Ref string `json:"ref"`
}

type labelPayload struct {
	Name string `json:"name"`
}

type changedFilePayload struct {
	Filename  string `json:"filename"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
}

func parsePullRequestPayload(content []byte) (*PullRequestMetadata, error) {
	var payload eventPayload
	if err := json.Unmarshal(content, &payload); err != nil {
		return nil, err
	}
	if payload.PullRequest == nil {
		return nil, nil
	}

	prPayload := payload.PullRequest
	pr := &PullRequestMetadata{
		Number:             prPayload.Number,
		Title:              prPayload.Title,
		URL:                prPayload.HTMLURL,
		BaseRef:            prPayload.Base.Ref,
		HeadRef:            prPayload.Head.Ref,
		Author:             prPayload.User.Login,
		Labels:             labelNames(prPayload.Labels),
		RequestedReviewers: userLogins(prPayload.RequestedReviewers),
		ChangedFiles:       prPayload.ChangedFiles,
		Additions:          prPayload.Additions,
		Deletions:          prPayload.Deletions,
		Draft:              prPayload.Draft,
	}

	for _, file := range prPayload.Files {
		if file.Filename == "" {
			continue
		}
		pr.ChangedFileNames = append(pr.ChangedFileNames, file.Filename)
	}

	return pr, nil
}

func labelNames(labels []labelPayload) []string {
	names := make([]string, 0, len(labels))
	for _, label := range labels {
		if label.Name != "" {
			names = append(names, label.Name)
		}
	}
	return names
}

func userLogins(users []userPayload) []string {
	logins := make([]string, 0, len(users))
	for _, user := range users {
		if user.Login != "" {
			logins = append(logins, user.Login)
		}
	}
	return logins
}

func environmentMap(values []string) map[string]string {
	env := make(map[string]string, len(values))
	for _, value := range values {
		key, val, ok := strings.Cut(value, "=")
		if ok {
			env[key] = val
		}
	}
	return env
}

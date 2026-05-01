# AI Implementation Prompt: GitHub Actions and Pull Request Integration

You are working in the `bottleneck` Go CLI repository.

Implement feature 4 from the roadmap: **GitHub Actions and Pull Request Integration**.

## Product Goal

Place bottleneck in the delivery workflow where release decisions happen.

Developers should be able to copy a workflow into `.github/workflows/`, run bottleneck during pull requests, publish a readable scorecard to GitHub Actions Step Summary, optionally maintain a PR comment, surface annotations on weak or missing artifacts, and block PRs based on configured evidence thresholds.

## Current Architecture To Respect

Use the existing CLI commands and validation engine. Do not create a second validation path just for CI.

Relevant files:

- `cmd/validate.go`
  - Runs validation and exits non-zero when system status is `FAIL`.
- `cmd/scorecard.go`
  - Renders scorecards. If feature 2 is implemented, it already supports `--format=markdown`, `--view`, and richer JSON.
- `internal/validator/engine.go`
  - Produces `models.EngineResult`.
- `internal/models/result.go`
  - Defines validation results and statuses.
- `internal/scorecard/scorecard.go`
  - Renders text, JSON, and possibly Markdown scorecards.
- `internal/explainer/explainer.go`
  - Reference for evidence-oriented output.
- `readme.md`
  - Update with GitHub Actions usage.
- `bottleneck/docs/validation.md`
  - Update if command behavior or workflow guidance changes.

There is currently no `.github/workflows/` directory in the repository. Add copyable workflow examples without assuming users want those exact workflows active in this repo unless explicitly intended.

## Required Behavior

### 1. Add Example GitHub Actions Workflows

Add copyable workflow examples for:

- `bottleneck validate`
- `bottleneck scorecard`
- PR gating with Step Summary, annotations, and optional PR comment

Recommended location:

```text
examples/github-actions/bottleneck-validate.yml
examples/github-actions/bottleneck-scorecard.yml
examples/github-actions/bottleneck-pr-gate.yml
```

Each workflow should be copyable into:

```text
.github/workflows/
```

Workflow requirements:

- Trigger on `pull_request`.
- Trigger manually with `workflow_dispatch`.
- Check out the repository.
- Build or install the local bottleneck CLI.
- Run validation with a configurable environment, defaulting to `production` for PR gates.
- Run scorecard with Markdown output.
- Append Markdown output to `$GITHUB_STEP_SUMMARY`.
- Fail the job when bottleneck returns non-zero because evidence thresholds failed.
- Include permissions needed for PR comments only in the workflow that comments.

Example shell pattern:

```sh
go build -o bottleneck .
./bottleneck validate --env=production --strict --github-annotations
./bottleneck scorecard --env=production --view=governance --format=markdown > bottleneck-scorecard.md
cat bottleneck-scorecard.md >> "$GITHUB_STEP_SUMMARY"
```

If the CLI does not yet support `--format=markdown`, implement it or clearly wire the workflow to the existing Markdown renderer from feature 2.

### 2. Output Markdown To GitHub Step Summary

Ensure `bottleneck scorecard --format=markdown` produces clean GitHub-flavored Markdown that works in:

- GitHub Actions Step Summary
- Pull request comments
- Release notes

The workflow examples should append that Markdown to:

```sh
$GITHUB_STEP_SUMMARY
```

Do not make Step Summary output require network access or GitHub API calls.

### 3. Support PR Comment Output

Support PR comment publishing through a workflow.

Recommended implementation:

- Have bottleneck generate stable Markdown scorecard output.
- Include a workflow using `actions/github-script` or `gh pr comment` to upsert a PR comment.
- Use a stable hidden marker so repeated runs update the same comment instead of spamming the PR.

Example marker:

```markdown
<!-- bottleneck-scorecard -->
```

The PR comment workflow should:

- Read `bottleneck-scorecard.md`.
- Find an existing bot comment containing the marker.
- Update it when present.
- Create it when absent.

Keep the CLI independent of GitHub authentication unless implementing an optional GitHub client is necessary for metadata enrichment.

### 4. Detect PR Metadata In GitHub Actions

Add a GitHub Actions metadata detector.

Recommended package:

```text
internal/githubactions
```

It should detect:

- `GITHUB_ACTIONS`
- `GITHUB_EVENT_NAME`
- `GITHUB_EVENT_PATH`
- `GITHUB_REPOSITORY`
- `GITHUB_SHA`
- `GITHUB_REF`
- `GITHUB_HEAD_REF`
- `GITHUB_BASE_REF`
- `GITHUB_RUN_ID`
- `GITHUB_SERVER_URL`

When `GITHUB_EVENT_PATH` points to a pull request event JSON file, parse the payload and extract:

- PR number
- PR title
- PR URL
- base branch
- head branch
- author
- labels
- requested reviewers
- changed file count if present
- additions/deletions if present
- draft status if present

Expose this metadata to scorecard output when running in GitHub Actions.

Recommended scorecard fields:

```json
"github": {
  "detected": true,
  "event_name": "pull_request",
  "repository": "owner/repo",
  "run_id": "123456",
  "pull_request": {
    "number": 42,
    "title": "Add release gate",
    "url": "https://github.com/owner/repo/pull/42",
    "base_ref": "main",
    "head_ref": "feature/release-gate",
    "labels": ["ai-assisted"],
    "requested_reviewers": ["octocat"],
    "changed_files": 18,
    "additions": 900,
    "deletions": 120,
    "draft": false
  }
}
```

If not running in GitHub Actions, scorecard should continue to work normally.

### 5. Include PR Risk Signals

Add deterministic PR risk detection based on available metadata.

Minimum risk signals:

- Large PR by changed file count.
- Large PR by additions plus deletions.
- Draft PR.
- Missing requested reviewers or no reviewers.
- AI-generated or AI-assisted labels.
- Missing artifact changes when source code changes appear release-relevant.

Recommended defaults:

- Warn when `changed_files > 25`.
- Warn when `additions + deletions > 1000`.
- Warn when label includes any of:
  - `ai-generated`
  - `ai-assisted`
  - `copilot`
  - `codex`
- Warn when a PR changes likely source paths but does not change any `bottleneck/` artifacts.

Suggested release-relevant source path patterns:

```text
cmd/**
internal/**
pkg/**
src/**
app/**
services/**
*.go
*.ts
*.tsx
*.js
*.jsx
*.py
```

Suggested artifact path patterns:

```text
bottleneck/**
bottleneck/**
```

If changed file names are unavailable from the event payload, support optional enrichment from the GitHub API when `GITHUB_TOKEN` is present.

Do not fail solely because metadata enrichment is unavailable. Report it as `UNKNOWN` or a warning depending on existing scorecard status support.

### 6. Optional GitHub API Enrichment

Approvals and checks are not fully available in every pull request event payload. Add optional GitHub API enrichment only when credentials are available.

Use `GITHUB_TOKEN` from the workflow and GitHub REST API to fetch:

- Changed files for the PR.
- Reviews and approval count.
- Check runs or combined status for the PR head SHA.

Implementation requirements:

- Keep network calls optional.
- Use a small interface so tests can use a fake client.
- Time out requests.
- Handle rate limits and permission errors gracefully.
- Do not make local scorecard usage depend on GitHub API access.

Risk signals from enrichment:

- No approval on non-draft PR.
- Requested reviewers still pending.
- Failed check runs.
- Changed file list confirms source changes without matching `bottleneck/` evidence artifacts.

### 7. Support GitHub Workflow Annotations

Add support for GitHub Actions annotations for missing or weak artifacts.

Recommended CLI flag:

```sh
bottleneck validate --github-annotations
bottleneck scorecard --github-annotations
```

When enabled in GitHub Actions, emit workflow commands:

```text
::warning file=bottleneck/intent/intent.md::Intent missing measurable success criteria
::error file=bottleneck/assurance/results.json::Assurance accuracy below production threshold
```

Rules:

- Emit `::error` for failing validation results.
- Emit `::warning` for warning validation results.
- Include `file=` when the finding can be tied to a file.
- Include `line=` when the finding can be tied to a line. Line numbers are optional.
- Escape workflow command characters correctly:
  - `%` -> `%25`
  - `\r` -> `%0D`
  - `\n` -> `%0A`
  - `:` and `,` in properties as needed.

If current `ValidationResult` only has unstructured `Details`, add a small structured finding type without breaking existing renderers:

```go
type ValidationFinding struct {
    Level   string
    Message string
    Path    string
    Line    int
}
```

Then preserve current `Message` and `Details` for terminal output.

Annotations should link to affected files where possible:

- `bottleneck/intent/intent.md`
- `bottleneck/behavior/behavior-spec.md`
- `bottleneck/design/architecture.md`
- `bottleneck/assurance/results.json`
- `bottleneck/security/guardrails.json`
- `bottleneck/execution/telemetry.json`

### 8. Pull Request Blocking Behavior

Pull requests should be blockable based on existing evidence thresholds.

Required behavior:

- `bottleneck validate --env=production --strict` exits non-zero on failures.
- `bottleneck scorecard --env=production` exits non-zero on system `FAIL`.
- GitHub workflows must rely on these exit codes to block PRs.
- Warnings should not fail the job unless strict mode or production rules intentionally promote them to failures.

Do not add a separate blocking mechanism if existing exit-code behavior is enough.

### 9. Documentation

Update docs with:

- How to copy workflows into `.github/workflows/`.
- How to publish Step Summary output.
- How PR comments are updated.
- What permissions are needed.
- How PR risk signals are calculated.
- How annotations appear in GitHub checks.
- How to choose `--env=dev`, `--env=stage`, or `--env=production`.
- How `--strict` affects PR gates.

Keep docs clear that bottleneck integrates with GitHub but does not replace GitHub Actions, CodeQL, review rules, or branch protection.

## Backward Compatibility

- Local `validate`, `explain`, and `scorecard` commands must keep working outside GitHub Actions.
- GitHub metadata must be optional.
- GitHub API enrichment must be optional.
- Existing scorecard text and JSON output should remain usable.
- Existing exit-code behavior should remain intact.
- Workflow examples should not require secrets beyond `GITHUB_TOKEN`.

## Testing Requirements

Add focused tests using local JSON fixtures and fake clients.

Required test cases:

1. GitHub Actions detector returns `detected=false` outside GitHub Actions.
2. Detector parses a pull request event payload from `GITHUB_EVENT_PATH`.
3. Detector handles missing or malformed event payloads gracefully.
4. PR risk flags large changed file count.
5. PR risk flags large additions plus deletions.
6. PR risk flags AI-assisted labels.
7. PR risk flags source changes without matching `bottleneck/` artifact changes when file list is available.
8. Optional GitHub API enrichment uses a fake client in tests.
9. Annotation rendering emits valid GitHub workflow commands.
10. Annotation rendering escapes special characters.
11. Scorecard JSON includes GitHub metadata when detected.
12. Markdown scorecard remains readable when GitHub metadata is present.
13. Workflow example files contain `pull_request`, `$GITHUB_STEP_SUMMARY`, and the expected bottleneck commands.

Run:

```sh
go test ./...
```

## Implementation Guidance

Recommended approach:

1. Add `internal/githubactions` for environment and event payload parsing.
2. Add `internal/prrisk` or a small scorecard helper for deterministic PR risk findings.
3. Add optional GitHub API enrichment behind a small interface.
4. Add annotation rendering in a focused package such as `internal/githubannotations`.
5. Add `--github-annotations` to `validate` and `scorecard`.
6. Add GitHub metadata and PR risk output to scorecard structs.
7. Add workflow examples under `examples/github-actions/`.
8. Update docs.
9. Add tests before or alongside implementation.

Keep the implementation small and deterministic. Prefer local event payload parsing first. Use GitHub API calls only for data that cannot be obtained from the event payload, and make those calls optional.

## Acceptance Criteria

- A user can copy workflow examples into `.github/workflows/`.
- Workflows run `bottleneck validate` and `bottleneck scorecard`.
- Markdown scorecard output is appended to GitHub Actions Step Summary.
- PR comment workflow updates a stable bottleneck comment instead of creating duplicates.
- Pull requests can be blocked based on configured evidence thresholds and strict mode.
- bottleneck detects PR metadata when running in GitHub Actions.
- Scorecard includes PR risk signals for large PRs, reviewer/approval gaps, checks, AI-assisted labels, and source changes without evidence artifacts when data is available.
- GitHub annotations appear for missing or weak artifacts and link to affected files where possible.
- Local usage outside GitHub Actions remains unchanged.
- `go test ./...` passes.

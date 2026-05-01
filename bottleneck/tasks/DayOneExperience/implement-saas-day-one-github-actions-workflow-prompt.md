# AI Implementation Prompt: Add A GitHub Actions Day-One Workflow

You are working in the `bottleneck` Go CLI repository.

Implement the sixth slice of the **SaaS Team Day-One Success** milestone.

## Milestone Goal

Make Bottleneck usable by a SaaS engineering team in the first 10 minutes.

A developer should be able to initialize Bottleneck, run it in GitHub Actions, see a scorecard in the workflow summary, understand what blocks release readiness, and see what only warns.

This implementation prompt is scoped to:

- Epic 6: Add A GitHub Actions Day-One Workflow
- Task 6.1: Create a copy/paste SaaS workflow
- Task 6.2: Document PR comment and step summary behavior
- Task 6.3: Add workflow validation tests

## Definition Of Done For This Slice

A SaaS developer can copy a workflow from `examples/github-actions/`, commit it to `.github/workflows/`, and get:

- Bottleneck validation in CI.
- A Markdown scorecard written to `$GITHUB_STEP_SUMMARY`.
- Release gate diagnosis using `bottleneck diagnose --gate=release`.
- GitHub annotation output when the CLI supports it.
- Clear docs explaining what appears in a pull request and what blocks release readiness.

The workflow must be safe to copy into a repository without requiring unavailable secrets or hardcoded local paths.

## Current Architecture To Respect

Inspect the repository before changing files. Follow existing patterns for:

- example workflows under `examples/` or `.github/workflows/`
- GitHub Actions helpers under `internal/githubactions/`
- GitHub annotation helpers under `internal/githubannotations/`
- diagnosis output formats under `cmd/diagnose.go`
- scorecard Markdown output under `cmd/scorecard.go` or `internal/scorecard/`
- docs under `readme.md` and `docs/`
- existing YAML parsing or workflow validation tests

Preserve backwards compatibility:

- Do not rename CLI commands or flags.
- Do not change existing workflow examples unless needed to keep them correct.
- Do not require secrets for the Day-One workflow.
- Do not assume the Bottleneck binary is preinstalled unless the workflow clearly builds or installs it.
- Do not document `bottleneck diagnose --format=github` as supported unless the command actually supports it.
- Do not weaken release gate behavior to make the example pass.

Prefer adding a copy/paste example and tests. Only change CLI behavior if the workflow exposes a clear bug in existing supported behavior.

## Primary Source Of Truth

Read:

- `tasks/DayOneExperience/saas-team-day-one-success-task-list.md`
- `readme.md`
- `docs/quickstart-saas.md`, if it exists
- existing files under `examples/`
- existing `.github/workflows/` files
- `cmd/diagnose.go`
- `cmd/scorecard.go`
- `internal/githubactions/*`
- `internal/githubannotations/*`
- existing workflow or docs tests

Keep the workflow consistent with the Day-One SaaS flow:

```sh
bottleneck validate
bottleneck scorecard --format=markdown
bottleneck diagnose --gate=release
```

## Epic 6: Add A GitHub Actions Day-One Workflow

### Task 6.1 - Create A Copy/Paste SaaS Workflow

Goal: Make Bottleneck easy to try in CI.

Create:

```text
examples/github-actions/bottleneck-saas-scorecard.yml
```

Workflow requirements:

- Runs `bottleneck validate`.
- Runs `bottleneck scorecard --format=markdown`.
- Runs `bottleneck diagnose --gate=release`.
- Writes scorecard output to `$GITHUB_STEP_SUMMARY`.
- Supports annotations if available through `bottleneck diagnose --format=github`.

Implementation guidance:

- Use standard GitHub Actions YAML.
- Trigger on `pull_request` and optionally `push`.
- Include a clear job name such as `bottleneck`.
- Use `actions/checkout`.
- If this repository is a Go CLI and no published action exists, build the CLI from source in the workflow:

```yaml
- uses: actions/setup-go@v5
  with:
    go-version-file: go.mod

- name: Build Bottleneck
  run: go build -o ./bin/bottleneck .
```

- Run commands through the built binary:

```yaml
- name: Validate Bottleneck evidence
  run: ./bin/bottleneck validate

- name: Write Bottleneck scorecard summary
  run: ./bin/bottleneck scorecard --format=markdown >> "$GITHUB_STEP_SUMMARY"

- name: Check release readiness
  run: ./bin/bottleneck diagnose --gate=release
```

- If GitHub annotation output is supported, add a step such as:

```yaml
- name: Emit Bottleneck annotations
  if: always()
  run: ./bin/bottleneck diagnose --format=github
```

- If annotation output is not supported, do not include that step. Instead document that annotations are an optional future enhancement or only include the step after implementing and testing the supported format.
- Keep the workflow independent from unavailable secrets.
- Avoid hardcoded local paths.
- Use repo-relative paths only.
- Do not write generated artifacts outside the workspace.

### Task 6.2 - Document PR Comment And Step Summary Behavior

Goal: Show how this fits into pull requests.

Docs must explain:

- How Bottleneck appears in GitHub Actions.
- What developers see in a PR.
- What blocks a release.
- What only warns.
- How to tune behavior by environment.

Recommended doc locations:

- `docs/quickstart-saas.md`, if it exists
- README CI section
- a short workflow-specific doc only if the repository already has one

Documentation requirements:

- Link to `examples/github-actions/bottleneck-saas-scorecard.yml`.
- Explain that `bottleneck scorecard --format=markdown` writes a human-readable release readiness summary to `$GITHUB_STEP_SUMMARY`.
- Explain that `bottleneck diagnose --gate=release` controls the blocking CI result.
- Explain that warnings can appear in the scorecard without necessarily failing the workflow, depending on environment and gate thresholds.
- Explain that release blockers include critical evidence gaps such as missing required assurance, broken traceability, critical security findings, or production gate failures where implemented.
- Explain how to tune environment behavior, for example:

```sh
bottleneck scorecard --env=stage --format=markdown
bottleneck diagnose --env=production --gate=release
```

- If GitHub annotations are supported, explain what developers see in PR checks and how annotations map to warnings/errors.
- If PR comments are not implemented, do not imply the workflow posts comments. Use "step summary" and "checks" language instead.

Recommended wording:

```text
In a pull request, Bottleneck writes the scorecard to the GitHub Actions step summary. Developers see the release recommendation, primary bottleneck, category results, and next action. The release gate step fails the job only when the configured environment treats the issue as blocking.
```

### Task 6.3 - Add Workflow Validation Tests

Goal: Prevent broken example workflows.

Tests must verify:

- Workflow YAML parses.
- Commands reference valid Bottleneck commands.
- Workflow does not require unavailable secrets.
- Workflow uses temporary or repo-safe paths.
- Workflow includes scorecard and diagnosis steps.

Implementation guidance:

- Add tests in the most appropriate existing package.
- If no workflow test package exists, add a small test file such as `github_actions_examples_test.go` at the repository root or in a relevant internal package.
- Use a YAML parser if already available in the repo. If no YAML parser dependency exists and adding one would be excessive, keep the validation lightweight and deterministic:
  - read the YAML file
  - assert key sections and commands exist
  - assert indentation-sensitive parse only if the repo already has YAML tooling
- Do not add a heavy dependency solely for this test unless the project already uses YAML parsing.

Minimum test coverage:

- `examples/github-actions/bottleneck-saas-scorecard.yml` exists.
- Workflow includes `actions/checkout`.
- Workflow builds or installs Bottleneck in a way that can work in CI.
- Workflow runs `bottleneck validate`.
- Workflow runs `bottleneck scorecard --format=markdown`.
- Workflow writes to `$GITHUB_STEP_SUMMARY`.
- Workflow runs `bottleneck diagnose --gate=release`.
- Workflow includes `bottleneck diagnose --format=github` only if that format is supported by the CLI.
- Workflow does not contain `secrets.`.
- Workflow does not contain absolute local paths such as `/Users/`, `/tmp/`, or machine-specific workspace paths.
- Workflow does not write generated evidence outside the repository workspace.

Command validity tests:

- Validate the workflow references commands that exist in the CLI.
- If tests can execute help safely, verify:

```sh
bottleneck validate --help
bottleneck scorecard --help
bottleneck diagnose --help
```

- If command execution is not practical, assert against known command registration in code or existing command test helpers.

## UX Requirements

The workflow should feel like a practical copy/paste starting point.

Prefer:

- simple CI steps
- clear step names
- Markdown scorecard in step summary
- release gate as the blocking step
- explicit environment examples in docs
- no secrets
- no hidden external services

Avoid:

- workflows that only work inside this repository
- workflows that rely on unpublished actions without explanation
- claiming PR comments exist when only step summary exists
- noisy shell scripts embedded in YAML
- changing CLI output just to satisfy the workflow example

## Tests To Add Or Update

Add tests in the most appropriate package:

- workflow example tests near other example validation tests
- docs tests if the repository already checks README or docs content
- command help tests in `cmd/*_test.go` only if help text is changed
- GitHub annotations tests in `internal/githubannotations/*_test.go` only if annotation behavior is implemented or changed

Minimum test coverage:

- Workflow file exists and is readable.
- Workflow YAML is structurally valid or passes lightweight validation.
- Workflow contains required Bottleneck commands.
- Workflow writes Markdown scorecard output to `$GITHUB_STEP_SUMMARY`.
- Workflow does not require unavailable secrets.
- Workflow uses repo-safe paths.
- Docs mention the workflow path.
- Docs explain step summary, release blocking, warning behavior, and environment tuning.

If `bottleneck diagnose --format=github` is not supported, add a test that the workflow does not reference it and report annotation support as deferred.

## Documentation Updates

Update docs only where needed:

- README CI section should link to the workflow example.
- `docs/quickstart-saas.md`, if it exists, should include the GitHub Actions step.
- Mention the copy target:

```sh
mkdir -p .github/workflows
cp examples/github-actions/bottleneck-saas-scorecard.yml .github/workflows/bottleneck.yml
```

- Explain what a developer sees:
  - step summary scorecard
  - release gate pass/fail
  - optional annotations if supported
- Explain how to tune environment:

```yaml
run: ./bin/bottleneck scorecard --env=production --format=markdown >> "$GITHUB_STEP_SUMMARY"
```

## Verification Commands

Run:

```sh
go test ./...
```

If feasible, manually inspect or validate the example workflow:

```sh
sed -n '1,220p' examples/github-actions/bottleneck-saas-scorecard.yml
```

If the repository has a workflow linter or YAML validation command, run it. Do not introduce a network-dependent linter just for this slice.

## Final Response Requirements

When finished, report:

1. Workflow file added or changed.
2. Docs updated for PR and step summary behavior.
3. Workflow validation tests added or changed.
4. Whether GitHub annotation output is supported or deferred.
5. Exact commands run and results.
6. Any acceptance criteria intentionally deferred and why.


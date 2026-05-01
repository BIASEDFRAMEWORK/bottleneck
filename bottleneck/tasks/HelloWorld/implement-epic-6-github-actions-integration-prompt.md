# AI Implementation Prompt: Epic 6 GitHub Actions Integration

You are working in the `bottleneck` Go CLI repository.

Implement **Epic 6: Strengthen GitHub Actions Integration**.

This epic covers:

- Task 6.1: Add PR comment output mode
- Task 6.2: Add GitHub annotation output
- Task 6.3: Add release gate mode

## Product Goal

Make Bottleneck useful inside pull requests and CI release gates.

Developers should be able to run Bottleneck in GitHub Actions, publish a PR-friendly diagnosis, emit annotations that point at weak evidence, and fail a build when diagnosis severity violates release gate thresholds.

## Current Architecture To Respect

Use existing diagnosis, scorecard, validation, and annotation packages. Do not duplicate rendering or validation logic.

Relevant files:

- `cmd/diagnose.go`
  - Existing `diagnose` command with `--env`, `--format`, and `--strict`.
- `cmd/scorecard.go`
  - Existing scorecard command and GitHub annotation support.
- `internal/diagnosis/diagnosis.go`
  - Existing diagnosis model, category scores, findings, confidence, and render behavior.
- `internal/scorecard/scorecard.go`
  - Existing scorecard Markdown and JSON output.
- `internal/githubannotations/annotations.go`
  - Existing GitHub Actions annotation renderer.
- `internal/config/config.go`
  - Existing environment threshold configuration.
- `internal/models/result.go`
  - Existing validation result, finding, and evidence-quality models.
- `examples/github-actions/`
  - Existing or future workflow examples.
- `readme.md`
  - Update CI and PR usage.

Some of this functionality may already exist. If so, harden it, complete missing behavior, and add tests rather than replacing it.

## Required Behavior

### 1. Add Or Refine PR Comment Output Mode

Make this command produce PR-friendly Markdown:

```sh
bottleneck diagnose --format markdown
```

If `diagnose --format markdown` already exists, verify it includes all required sections and improve it if needed.

Markdown output must include:

- Primary bottleneck
- Category scores
- Top findings
- Recommended next action

Recommended Markdown shape:

```markdown
## Bottleneck Diagnosis

| Field | Value |
| --- | --- |
| Primary Bottleneck | Assurance |
| Confidence | Low |
| Recommended Action | Map BEHAVIOR-001 to a passing BDD or evaluation result. |

### Category Scores

| Category | Score | Status |
| --- | ---: | --- |
| Assurance | 20 | FAIL |
| Intent | 80 | PASS |

### Top Findings

1. No assurance result references BEHAVIOR-001.
2. Assurance accuracy is below the selected environment threshold.
```

Requirements:

- No ANSI color codes.
- No terminal-only glyphs.
- Stable headings.
- Stable table columns.
- Safe for GitHub PR comments and GitHub Step Summary.

Add snapshot-style tests if the repository already uses snapshots. Otherwise use focused string assertions.

### 2. Add GitHub Annotation Output

Add GitHub annotation output mode:

```sh
bottleneck diagnose --format github
```

If it fits the current CLI better, also support:

```sh
bottleneck validate --format github
bottleneck scorecard --format github
```

But the minimum required command for this epic is:

```sh
bottleneck diagnose --format github
```

Output should use GitHub Actions workflow commands:

```text
::warning file=bottleneck/intent/intent.md,line=1::Intent evidence contains placeholder content
::error file=bottleneck/security/guardrails.json::Security evidence fails release policy
```

Rules:

- Emit `::warning` for warning findings.
- Emit `::error` for failing findings.
- Include file path when known.
- Include line number when known.
- Support warning vs error severity.
- Escape workflow command values correctly.

Escaping requirements:

- Message: `%` to `%25`, `\r` to `%0D`, `\n` to `%0A`.
- Properties: message escaping plus `:` to `%3A` and `,` to `%2C`.

Use existing `internal/githubannotations` where possible.

### 3. Improve Annotation Finding Quality

Annotations should point to real evidence artifacts where possible.

Expected default paths:

- Behavior: `bottleneck/behavior/behavior-spec.md`
- Intent: `bottleneck/intent/intent.md`
- Design: `bottleneck/design/architecture.md`
- Assurance: `bottleneck/assurance/results.json`
- Security: `bottleneck/security/guardrails.json`
- Execution: `bottleneck/execution/telemetry.json`
- Config: `bottleneck/config.yaml`

If a `ValidationFinding` includes a path and line, preserve them.

If details contain a path, extract it.

If neither exists, fall back to the default path by capability.

### 4. Add Release Gate Mode

Add:

```sh
bottleneck diagnose --gate release
```

Release gate mode should fail the build when diagnosis severity violates configured thresholds.

Required failure conditions:

- Primary bottleneck score is below threshold.
- Required category is missing.
- Traceability is broken.
- Security evidence fails.
- Governance evidence fails.

If governance artifacts are not implemented yet, handle that explicitly:

- Do not invent governance evidence.
- If release gate config requires governance, fail when missing.
- If no governance requirement exists, report governance as not assessed.

### 5. Add Configurable Release Gate Thresholds

Extend `config.yaml` with release gate settings while preserving backward compatibility.

Recommended schema:

```yaml
environments:
  default:
    gate:
      release:
        min_primary_score: 60
        required_categories:
          - Intent
          - Behavior
          - Assurance
          - Security
          - Execution
        require_traceability: true
        require_governance: false
```

Production can override:

```yaml
  production:
    gate:
      release:
        min_primary_score: 75
        require_traceability: true
        require_governance: true
```

Rules:

- Missing gate config should use safe defaults.
- Existing config files must still parse.
- Environment overrides should inherit from `default`, consistent with existing config behavior.

### 6. Release Gate Output

When `--gate release` is used, output should include:

- Gate name: `release`
- Gate result: `PASS` or `FAIL`
- Primary bottleneck
- Primary score
- Threshold
- Gate reasons

Example:

```text
Release Gate: FAIL

Primary Bottleneck: Assurance
Primary Score: 20
Required Score: 75

Reasons:
1. Primary bottleneck score is below release threshold.
2. Traceability is broken.

Recommended next action:
Map BEHAVIOR-001 to a passing BDD or evaluation result.
```

Exit behavior:

- Exit `0` when gate passes.
- Exit `1` when gate fails.
- Existing non-gated diagnose behavior should remain unchanged.

### 7. GitHub Actions Workflow Guidance

Add or update workflow examples.

Recommended command:

```sh
go build -o bottleneck .
./bottleneck diagnose --env=production --strict --gate release --format markdown > bottleneck-diagnosis.md
cat bottleneck-diagnosis.md >> "$GITHUB_STEP_SUMMARY"
./bottleneck diagnose --env=production --strict --format github
```

If a stable PR comment workflow exists, update it to use:

```sh
bottleneck diagnose --format markdown
```

Use a hidden marker for PR comments:

```markdown
<!-- bottleneck-diagnosis -->
```

## Backward Compatibility

- Existing `diagnose --format text|json|markdown` behavior should keep working.
- Existing `scorecard --format markdown` should keep working.
- Existing `validate` and `scorecard` exit behavior should remain unchanged.
- `--gate release` should only affect diagnose when explicitly provided.
- Existing config files should still load.
- GitHub annotations should not require network access.

## Testing Requirements

Add focused tests.

Required test cases:

1. `diagnose --format markdown` includes primary bottleneck.
2. Markdown output includes category scores.
3. Markdown output includes top findings.
4. Markdown output includes recommended next action.
5. `diagnose --format github` emits warning annotations.
6. `diagnose --format github` emits error annotations.
7. GitHub annotations include file path when known.
8. GitHub annotations include line number when known.
9. GitHub annotations escape `%`, newlines, colons, and commas correctly.
10. `diagnose --gate release` passes when thresholds are met.
11. `diagnose --gate release` fails when primary bottleneck score is below threshold.
12. Release gate fails when a required category is missing.
13. Release gate fails when traceability is broken.
14. Release gate fails when Security fails.
15. Release gate fails when governance is required and missing.
16. Existing config without gate settings still loads.
17. Environment-specific gate settings inherit from default.

Run:

```sh
go test ./...
```

## Implementation Guidance

Recommended approach:

1. Verify existing diagnosis Markdown output and add missing required sections.
2. Add `github` as a supported diagnosis output format.
3. Reuse `internal/githubannotations.Render` for annotation output.
4. Add gate config structs under `internal/config`.
5. Add a small release gate evaluator, for example `internal/gate`.
6. Add `--gate` to `cmd/diagnose.go`.
7. Keep gate evaluation based on existing diagnosis and validation results.
8. Update docs and workflow examples.
9. Add tests before or alongside implementation.

Avoid creating a separate CI-only diagnosis model. GitHub output and release gates should be renderers/evaluators over the same diagnosis data used locally.

## Acceptance Criteria

- `bottleneck diagnose --format markdown` produces PR-friendly Markdown.
- Markdown includes primary bottleneck, category scores, top findings, and recommended next action.
- `bottleneck diagnose --format github` emits valid GitHub Actions annotations.
- Annotations support warning vs error severity.
- Annotations include file and line when known.
- `bottleneck diagnose --gate release` exists.
- Release gate thresholds are configurable in `config.yaml`.
- Release gate fails when primary bottleneck score is below threshold.
- Release gate fails when a required category is missing.
- Release gate fails when traceability is broken.
- Release gate fails when Security fails.
- Release gate fails when governance evidence is required and missing.
- Existing local CLI behavior remains backward compatible.
- Tests cover Markdown, annotations, and release gates.
- `go test ./...` passes.

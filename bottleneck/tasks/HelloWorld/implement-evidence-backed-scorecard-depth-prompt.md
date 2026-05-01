# AI Implementation Prompt: Evidence-Backed Scorecard Depth

You are working in the `bottleneck` Go CLI repository.

Implement feature 2 from the roadmap: **Evidence-Backed Scorecard Depth**.

## Product Goal

Make `bottleneck scorecard` the primary product surface instead of a thin validation summary.

The scorecard should help operators, engineers, and governance reviewers understand release readiness from evidence, thresholds, missing signals, and recommended release posture. It must remain deterministic and suitable for CI/CD automation.

## Current Architecture To Respect

Use the existing validation engine and scorecard renderer. Do not create a separate validation path just for scorecards.

Relevant files:

- `cmd/scorecard.go`
  - Defines the `bottleneck scorecard` command and current `--env` / `--format` flags.
- `internal/scorecard/scorecard.go`
  - Builds and renders the current scorecard.
- `internal/scorecard/scorecard_test.go`
  - Existing scorecard tests.
- `internal/validator/engine.go`
  - Produces `models.EngineResult`.
- `internal/models/result.go`
  - Defines `ValidationResult`, `EngineResult`, and status values.
- `internal/config/config.go`
  - Loads and resolves environment thresholds.
- `internal/validator/assurance.go`
  - Uses effective assurance thresholds.
- `internal/validator/execution.go`
  - Uses effective execution thresholds.
- `internal/explainer/explainer.go`
  - Useful reference for evidence-oriented rendering.
- `readme.md`
  - Update command documentation.
- `bottleneck/docs/validation.md` or generated validation docs if command behavior changes.

## Required Behavior

### 1. Add Evidence Counts and Missing Evidence Per Capability

Extend the scorecard model so each capability reports:

- `evidence_count`
- `missing_evidence`
- `evidence`
- `reason`
- `recommended_action`

Use existing validation `Message` and `Details` as evidence inputs.

Minimum expected behavior:

- Passing validation with supporting detail should count as evidence.
- Missing files, missing required sections, placeholder warnings, failed thresholds, or empty details should produce missing-evidence entries.
- The scorecard must explain what evidence is missing in a way that is actionable.

Example JSON shape:

```json
{
  "capability": "Assurance",
  "status": "FAIL",
  "evidence_count": 2,
  "missing_evidence": [],
  "reason": "accuracy below threshold",
  "recommended_action": "Fix failing scenarios or regenerate external BDD results.",
  "evidence": [
    "accuracy: 0.90 (threshold: 0.95)",
    "scenarios_failed: 1 (allowed: 0)"
  ]
}
```

Keep field names stable and snake_case for JSON.

### 2. Add Status Levels: PASS, WARN, FAIL, UNKNOWN

Normalize scorecard display statuses to:

- `PASS`
- `WARN`
- `FAIL`
- `UNKNOWN`

The internal model currently has `WARNING`. The scorecard may display `WARN`, but JSON should consistently use the same displayed status values unless there is a strong compatibility reason.

Expected mapping:

- `models.StatusPass` -> `PASS`
- `models.StatusWarning` -> `WARN`
- `models.StatusFail` -> `FAIL`
- Missing, skipped, or unavailable evidence -> `UNKNOWN`

Do not make normal missing required files `UNKNOWN`; those should remain `FAIL`. Use `UNKNOWN` for cases where a category cannot be assessed because evidence is explicitly unavailable or validation was not run.

### 3. Add Release Recommendation

Add a top-level `release_recommendation` field with one of:

- `Proceed`
- `Conditional`
- `Block`
- `Unknown`

Recommended logic:

- `Block` when any capability is `FAIL`.
- `Conditional` when there are no failures but at least one capability is `WARN`.
- `Proceed` when all required capabilities are `PASS`.
- `Unknown` when required scorecard evidence is unavailable or system status is unknown.

Text output should show the recommendation near the top.

JSON output should include:

```json
"release_recommendation": "Conditional"
```

### 4. Display Effective Environment Thresholds

The scorecard must show the effective thresholds used for the selected environment.

At minimum include:

- `assurance.min_accuracy`
- `assurance.max_failures`
- `execution.max_error_rate`
- `execution.min_adoption`

These thresholds should reflect inherited config after resolving the selected environment.

Implementation options:

- Add effective threshold fields to `models.EngineResult`.
- Or have the scorecard command load resolved config once and pass it into the scorecard builder.

Prefer an approach that avoids duplicate config resolution bugs.

Text output example:

```text
Effective Thresholds:
  assurance.min_accuracy: 0.95
  assurance.max_failures: 0
  execution.max_error_rate: 0.05
  execution.min_adoption: 0.50
```

JSON output example:

```json
"effective_thresholds": {
  "assurance": {
    "min_accuracy": 0.95,
    "max_failures": 0
  },
  "execution": {
    "max_error_rate": 0.05,
    "min_adoption": 0.5
  }
}
```

### 5. Add Scorecard Views

Add:

- `bottleneck scorecard --view executive`
- `bottleneck scorecard --view engineering`
- `bottleneck scorecard --view governance`

Default view should remain useful for terminal use. Choose one of:

- Keep default as current broad scorecard behavior.
- Or make `engineering` the default if that fits the existing CLI style better.

Required view behavior:

#### Executive View

Short, decision-oriented.

Include:

- Environment
- System status
- Release recommendation
- Primary bottleneck
- Capability status summary
- One concise bottom line

Avoid long evidence details.

#### Engineering View

Detailed and remediation-oriented.

Include:

- Environment
- System status
- Release recommendation
- Primary bottleneck
- Effective thresholds
- Capability rows
- Evidence counts
- Missing evidence
- Reasons
- Recommended actions
- Detailed evidence

#### Governance View

Policy and approval oriented.

Include:

- Environment
- System status
- Release recommendation
- Primary bottleneck
- Effective thresholds
- Security status
- Assurance status
- Execution status
- Any missing evidence that should block or condition release
- A release decision summary suitable for review

Do not invent approval state unless a governance artifact exists. If governance evidence does not exist yet, call it out as missing or not assessed.

### 6. Support Markdown Output

Add:

```sh
bottleneck scorecard --format=markdown
```

Existing formats:

- `text`
- `json`

New format:

- `markdown`

Markdown output must be readable in:

- GitHub Actions Step Summary
- Pull request comments
- Release notes

Use simple GitHub-flavored Markdown:

- A short heading
- A summary table
- Capability table
- Bullet lists for missing evidence and recommended actions where needed

Avoid terminal-only formatting in Markdown.

Example:

```markdown
# bottleneck Scorecard

| Field | Value |
| --- | --- |
| Environment | production |
| System Status | WARN |
| Release Recommendation | Conditional |
| Primary Bottleneck | Behavior |

## Capabilities

| Capability | Status | Evidence | Missing Evidence | Recommendation |
| --- | --- | ---: | --- | --- |
| Behavior | WARN | 1 | Expected Behavior needs real evidence | Update behavior-spec.md |
```

### 7. Stable JSON For CI/CD

The JSON output must be stable enough for automation.

Requirements:

- Use explicit structs rather than `map[string]any` for primary output.
- Use snake_case JSON field names.
- Preserve deterministic capability ordering.
- Include all top-level fields even when empty where practical.
- Include `schema_version`, starting with `"scorecard.v1"`.

Recommended top-level JSON fields:

- `schema_version`
- `environment`
- `system_status`
- `release_recommendation`
- `primary_bottleneck`
- `effective_thresholds`
- `capabilities`
- `bottom_line`

## Backward Compatibility

Do not remove existing `text` and `json` scorecard support.

It is acceptable for the text output to become richer, but existing core strings should remain understandable:

- `Environment`
- `System Status`
- `Primary Bottleneck`
- Capability names
- Bottom line

`scorecard` should continue to exit non-zero when the system is failing. It should not exit non-zero only because the release recommendation is `Conditional`.

## Testing Requirements

Add focused Go tests for scorecard construction and rendering.

Required test cases:

1. JSON includes `schema_version`, `release_recommendation`, `effective_thresholds`, and capabilities.
2. Release recommendation is `Proceed` when all capabilities pass.
3. Release recommendation is `Conditional` when any capability warns and none fail.
4. Release recommendation is `Block` when any capability fails.
5. Capability entries include evidence count, missing evidence, reason, and recommended action.
6. Text output includes effective thresholds and release recommendation.
7. Markdown output renders a GitHub-readable table.
8. Executive view omits detailed evidence but includes release decision fields.
9. Engineering view includes detailed evidence and missing evidence.
10. Governance view calls out security, assurance, execution, and missing governance evidence without inventing approval state.
11. Unsupported `--format` or `--view` values return useful errors.

Run:

```sh
go test ./...
```

## Implementation Guidance

Recommended approach:

1. Extend scorecard structs in `internal/scorecard/scorecard.go`.
2. Add explicit threshold structs either in `models.EngineResult` or scorecard-specific input.
3. Add release recommendation calculation.
4. Add capability evidence summarization helpers.
5. Add renderer support for `text`, `json`, and `markdown`.
6. Add view-specific rendering behavior.
7. Update `cmd/scorecard.go` with `--view` and `--format=markdown`.
8. Update docs and examples.
9. Add tests before or alongside implementation.

Keep this feature deterministic. Do not use LLM calls, network calls, or external services.

## Acceptance Criteria

- `bottleneck scorecard` explains why every category passed, warned, failed, or is unknown.
- Scorecard includes evidence counts and missing evidence per capability.
- Scorecard includes release recommendation: `Proceed`, `Conditional`, `Block`, or `Unknown`.
- Scorecard displays effective environment thresholds.
- `--view executive`, `--view engineering`, and `--view governance` work.
- `--format=markdown` produces readable GitHub-flavored Markdown.
- `--format=json` is stable and automation-friendly.
- Existing validation behavior remains intact.
- `go test ./...` passes.

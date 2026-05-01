# AI Implementation Prompt: Epic 3 Scorecard Gauge And Diagnose Command

You are working in the `bottleneck` Go CLI repository.

Implement **Epic 3: Make The Scorecard Feel Like A Gauge**.

This epic covers:

- Task 3.1: Redesign terminal scorecard output
- Task 3.2: Add `bottleneck diagnose`
- Task 3.3: Add confidence level to diagnosis

## Product Goal

Make the CLI immediately show where the system is weakest.

`bottleneck scorecard` should feel like a gauge: quick to scan, sorted or highlighted by weakness, and useful in a plain terminal or CI log. `bottleneck diagnose` should give the shortest path to understanding the primary bottleneck, top contributing findings, recommended next action, and confidence level.

Target scorecard shape:

```text
Bottleneck Scorecard

Primary Bottleneck: Assurance

Intent      [########--] 80
Behavior    [######----] 60
Design      [#####-----] 50
Assurance   [##--------] 20
Security    [#######---] 70
Execution   [######----] 60

Next action:
Map BEHAVIOR-001 to a passing BDD or evaluation result.
```

Target diagnose shape:

```text
Primary Bottleneck: Behavior

Contributing findings:
1. No unacceptable behaviors defined.
2. Behavior spec does not reference INTENT-001.
3. No evaluation evidence found for expected behavior.

Recommended next action:
Define expected and unacceptable behavior for each intent statement.

Diagnosis Confidence: Low

Reason:
Only 2 of 6 evidence categories contain meaningful content.
```

## Current Architecture To Respect

Use the existing scorecard and diagnosis models. Do not add a separate diagnosis implementation that disagrees with scorecard output.

Relevant files:

- `cmd/scorecard.go`
  - Existing scorecard command, flags, GitHub metadata, and annotation behavior.
- `internal/scorecard/scorecard.go`
  - Existing scorecard rendering, capability scores, diagnosis, Markdown, JSON, and views.
- `internal/scorecard/scorecard_test.go`
  - Extend for gauge-style text output.
- `internal/diagnosis/diagnosis.go`
  - Existing diagnosis analysis and scoring, if present.
- `internal/diagnosis/diagnosis_test.go`
  - Extend for confidence and contributing findings.
- `internal/models/result.go`
  - Existing validation result, evidence quality, and score-impact structures.
- `internal/validator/engine.go`
  - Source of validation results.
- `cmd/root.go`
  - Register the new `diagnose` command.
- `readme.md`
  - Update command examples.

If Epic 1 or Epic 2 is already implemented, build on those models. If pieces are missing, implement the smallest shared model needed so `scorecard` and `diagnose` use the same diagnosis result.

## Required Behavior

### 1. Redesign Terminal Scorecard Output

Update text scorecard output so the terminal view is scan-first.

Required content:

- `Bottleneck Scorecard` heading.
- Primary bottleneck section.
- Overall diagnosis.
- Next action section.
- Category gauges or simple visual indicators.
- Capability scores.
- Plain terminal readability.
- CI log readability.

Gauge requirements:

- Use ASCII by default for portability: `[########--] 80`.
- Avoid Unicode blocks unless there is an explicit existing style or flag.
- Keep labels aligned.
- Clamp scores from `0` to `100`.
- Render missing or unknown scores clearly.

Recommended gauge helper:

```go
func gauge(score int, width int) string
```

Example:

```text
Intent      [########--] 80
Behavior    [######----] 60
Design      [#####-----] 50
Assurance   [##--------] 20
Security    [#######---] 70
Execution   [######----] 60
```

### 2. Sort Or Highlight Weakest Category

The output should immediately show the weakest category.

Acceptable approaches:

- Sort categories by weakest first.
- Keep BIASED order but visibly mark the weakest category.

Recommended approach:

- Keep BIASED order for familiarity.
- Add a marker to the weakest category, such as `<-- primary bottleneck`.

Example:

```text
Assurance   [##--------] 20  <-- primary bottleneck
```

If multiple categories tie, mark each tied category or show a tied bottlenecks line.

### 3. Add Overall Diagnosis

Add a short diagnosis section near the top.

Required fields:

- Primary bottleneck.
- Why it matters.
- Recommended next action.

Example:

```text
Primary Bottleneck: Assurance

Why:
Your system has defined intent and behavior, but no evidence proves that behavior was tested.

Next action:
Add assurance evidence that maps test or evaluation results to BEHAVIOR-001.
```

Use the same diagnosis source as JSON and Markdown output.

### 4. Add `bottleneck diagnose`

Add a new command:

```sh
bottleneck diagnose
```

Command purpose:

- Focus on diagnosis, not full validation detail.
- Reuse validation and scorecard diagnosis logic.
- Provide the fastest explanation of the current bottleneck.

Required flags:

- `--env`, default `default`
- `--strict`, consistent with `validate` and `scorecard`
- `--format text|json|markdown`, default `text`

Optional if existing patterns make it simple:

- `--github-annotations`
- `--gate release`

Required text output:

- Primary bottleneck.
- Top 3 contributing findings.
- Recommended next action.
- Confidence level.
- Confidence reason.

Example:

```text
Primary Bottleneck: Behavior

Contributing findings:
1. No unacceptable behaviors defined.
2. Behavior spec does not reference INTENT-001.
3. No evaluation evidence found for expected behavior.

Recommended next action:
Define expected and unacceptable behavior for each intent statement.

Diagnosis Confidence: Low

Reason:
Only 2 of 6 evidence categories contain meaningful content.
```

Exit behavior:

- Exit non-zero when system status is `FAIL`.
- Do not exit non-zero for warning-only diagnosis unless strict mode promotes warnings to failures.

### 5. Add Top 3 Contributing Findings

Diagnosis should show the top 3 findings that explain the bottleneck.

Finding sources:

- `ValidationResult.Message`
- `ValidationResult.Details`
- `ValidationResult.Findings`
- evidence quality missing items
- score impacts
- traceability findings, if present

Selection rules:

- Prefer findings from the primary bottleneck category.
- Prefer failures over warnings.
- Prefer score-impact reasons when available.
- Limit to 3 findings.
- Use clear human wording.

Fallback:

- If no specific finding exists, use the category reason from the diagnosis metadata.

### 6. Add Confidence Level

Add confidence levels:

- `High`
- `Medium`
- `Low`

Base confidence on:

- Number of evidence files present.
- Traceability completeness.
- Amount of meaningful content.
- Evidence recency when timestamps exist.

Recommended deterministic calculation:

- Start from `Low`.
- Move to `Medium` when at least 4 of 6 BIASED categories have meaningful evidence.
- Move to `High` when all 6 categories have meaningful evidence and traceability has no warnings or failures.
- Reduce one level when traceability is broken.
- Reduce one level when evidence is stale, if timestamps exist.

Confidence reason examples:

```text
Only 2 of 6 evidence categories contain meaningful content.
```

```text
All 6 evidence categories are present, but traceability has broken references.
```

```text
All 6 evidence categories contain meaningful, connected evidence.
```

Do not infer confidence from vague prose. Use explicit validation, evidence counts, score impacts, and timestamps.

### 7. Support JSON And Markdown Diagnose Output

`bottleneck diagnose --format=json` should return stable JSON.

Recommended shape:

```json
{
  "schema_version": "diagnose.v1",
  "environment": "production",
  "system_status": "FAIL",
  "primary_bottleneck": "Assurance",
  "tied_bottlenecks": [],
  "why_it_matters": "Assurance proves the system behaves as intended.",
  "contributing_findings": [
    "No assurance result references BEHAVIOR-001"
  ],
  "recommended_action": "Map BEHAVIOR-001 to a passing BDD or evaluation result.",
  "confidence": "Low",
  "confidence_reason": "Only 2 of 6 evidence categories contain meaningful content."
}
```

`bottleneck diagnose --format=markdown` should be PR-comment friendly:

```markdown
## Bottleneck Diagnosis

| Field | Value |
| --- | --- |
| Primary Bottleneck | Assurance |
| Confidence | Low |

### Contributing Findings

1. No assurance result references BEHAVIOR-001.

### Next Action

Map BEHAVIOR-001 to a passing BDD or evaluation result.
```

### 8. Keep Output Readable In Plain Terminals And CI Logs

Requirements:

- No color-only meaning.
- No wide tables that wrap badly.
- No required Unicode glyphs.
- Clear labels.
- Deterministic ordering.
- Useful output when redirected to a file.

## Backward Compatibility

- Existing `bottleneck scorecard` flags should keep working.
- Existing text, JSON, and Markdown scorecard output can become richer but should preserve core summary fields.
- Existing validation and exit behavior should remain intact.
- New `diagnose` command should reuse existing validation logic and not change `validate`.
- No network calls or GitHub dependencies should be required for local diagnosis.

## Testing Requirements

Add focused Go tests.

Required test cases:

1. Gauge helper renders `0`, `20`, `60`, `80`, and `100` correctly.
2. Gauge helper clamps values below `0` and above `100`.
3. Scorecard text output includes category gauges.
4. Scorecard text output highlights or sorts the weakest category.
5. Scorecard text output includes primary bottleneck, why, and next action.
6. `bottleneck diagnose` text output includes primary bottleneck.
7. `bottleneck diagnose` text output includes top 3 contributing findings.
8. `bottleneck diagnose` text output includes recommended next action.
9. Diagnosis confidence is `Low` for sparse evidence.
10. Diagnosis confidence is `Medium` for partial meaningful evidence.
11. Diagnosis confidence is `High` for complete connected evidence.
12. Broken traceability lowers confidence.
13. Diagnose JSON output includes stable schema fields.
14. Diagnose Markdown output is PR-comment friendly.
15. Command exit behavior matches system status.

Run:

```sh
go test ./...
```

## Implementation Guidance

Recommended approach:

1. Add or extend `internal/diagnosis` so it owns:
   - primary bottleneck
   - tied bottlenecks
   - contributing findings
   - recommended action
   - confidence
   - confidence reason
2. Update `internal/scorecard` text rendering to include gauge output.
3. Add small rendering helpers for gauges and diagnosis sections.
4. Add `cmd/diagnose.go`.
5. Add `internal/diagnose` renderer only if keeping it separate from `internal/diagnosis` improves clarity.
6. Reuse `validator.NewEngine(".", env, validator.WithStrictMode(strict))`.
7. Keep JSON structs explicit and versioned.
8. Add tests before or alongside implementation.

Avoid duplicating category-scoring rules in multiple packages. `scorecard` and `diagnose` should agree on the primary bottleneck.

## Acceptance Criteria

- Terminal scorecard shows category gauges or equivalent visual indicators.
- Scorecard immediately shows the weakest category.
- Scorecard includes overall diagnosis, primary bottleneck, and next action.
- `bottleneck diagnose` exists.
- `bottleneck diagnose` outputs primary bottleneck, top 3 contributing findings, recommended next action, and confidence level.
- Confidence can be `High`, `Medium`, or `Low`.
- Confidence is based on evidence presence, traceability completeness, meaningful content, and recency when timestamps exist.
- Output remains readable in plain terminals and CI logs.
- Tests cover rendering, command behavior, and confidence calculation.
- `go test ./...` passes.

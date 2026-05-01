# AI Implementation Prompt: Epic 1 Primary Bottleneck Diagnosis

You are working in the `bottleneck` Go CLI repository.

Implement **Epic 1: Make The CLI Diagnose The Primary Bottleneck**.

This epic covers:

- Task 1.1: Add primary bottleneck detection
- Task 1.2: Add why-this-matters explanations
- Task 1.3: Add recommended next action per diagnosis

## Product Goal

Make Bottleneck diagnose the weakest delivery category instead of only reporting validation status.

When a developer runs `bottleneck scorecard` or `bottleneck explain`, the CLI should identify the weakest category, explain why that bottleneck matters in plain language, and recommend one top next action.

Target user experience:

```text
Primary Bottleneck: Assurance

Why:
Your system has defined intent and behavior, but no evidence proves that behavior was tested.

Next action:
Add assurance evidence that maps test or evaluation results to BEHAVIOR-001.
```

## Current Architecture To Respect

Use the existing validation engine, scorecard model, and explainer model. Do not create a separate diagnosis system that duplicates existing capability results.

Relevant files:

- `internal/scorecard/scorecard.go`
  - Already builds a richer scorecard with capability metadata, evidence, missing evidence, reasons, and recommended actions.
- `internal/scorecard/scorecard_test.go`
  - Extend for diagnosis behavior.
- `internal/explainer/explainer.go`
  - Already has capability metadata and recommended next actions.
- `internal/explainer/explainer_test.go`
  - Extend for why-this-matters and diagnosis behavior.
- `internal/models/result.go`
  - Defines `EngineResult` and `ValidationResult`.
- `internal/validator/engine.go`
  - Produces capability validation results and current system bottleneck.
- `cmd/scorecard.go`
  - Renders scorecard output.
- `cmd/explain.go`
  - Renders explanation output.
- `readme.md`
  - Update command behavior documentation if output meaning changes.

Current code already includes:

- capability metadata in scorecard
- capability metadata in explainer
- `PrimaryBottleneck` in `EngineResult`
- scorecard release recommendation and evidence details

This epic should improve that into a real diagnosis model rather than replacing it.

## Required Behavior

### 1. Add Scoring Logic For Each BIASED Category

Add deterministic category scoring for:

- Behavior
- Intent
- Design
- Assurance
- Security
- Execution

Recommended score range:

- `0` to `100`

You may include `Traceability` or `Config` in the internal model if already present in the codebase, but the primary diagnosis for this epic must focus on the six BIASED categories above.

Scoring should be derived from existing validation output, not invented independently.

Minimum expectations:

- `PASS` should score high.
- `WARNING` should score mid-range.
- `FAIL` should score low.
- Missing or weak evidence should reduce scores further where details are available.

Recommended baseline:

- `PASS` => `85`
- `WARNING` => `60`
- `FAIL` => `20`

Then adjust based on evidence shape:

- Missing evidence lowers score.
- Broken traceability lowers score if traceability data exists.
- Placeholder or weak evidence lowers score if content-quality data exists.
- Strong supporting evidence can raise a passing category toward `100`.

Keep the scoring deterministic and explainable. Do not add AI-based scoring.

### 2. Calculate The Primary Bottleneck

Primary bottleneck should be the weakest of the six BIASED categories.

Required behavior:

- Identify the lowest-scoring category.
- Return a clear `Primary Bottleneck` message.
- Handle ties when multiple categories are equally weak.

Recommended tie behavior:

- Add `PrimaryBottleneck` as the first bottleneck in priority order for concise CLI output.
- Also expose all tied bottlenecks in a new field, for example:

```go
PrimaryBottlenecks []string
```

Suggested priority order when scores tie:

1. Assurance
2. Security
3. Behavior
4. Intent
5. Execution
6. Design

Reasoning:

- Assurance and Security are most release-critical.
- Behavior and Intent define correctness.
- Execution and Design matter materially but usually after correctness and risk.

If all six categories are strong, report:

```text
Primary Bottleneck: None
```

or another equally explicit no-bottleneck signal.

### 3. Add Why-This-Matters Explanations

Each bottleneck must explain the delivery risk in plain language.

Add or refine a category explanation map for:

- Behavior
- Intent
- Design
- Assurance
- Security
- Execution

Current explainer metadata already contains `whyItMatters`. Reuse or extend that idea so the scorecard and explainer share the same explanation source where practical.

Content requirements:

- Developer-friendly language.
- Plain English.
- Avoid heavy framework wording.
- Explain downstream delivery risk, not abstract theory.

Example:

```text
Intent bottleneck:
The team has not clearly defined what good looks like. This creates downstream ambiguity in design, testing, security, and release decisions.
```

### 4. Add One Top Recommended Next Action

Every diagnosis should produce one prioritized next action.

Create recommendation rules for:

- Missing evidence
- Weak evidence
- Stale evidence
- Disconnected evidence

Use one top recommendation per category instead of listing every possible follow-up.

Current scorecard and explainer code already have recommended action metadata. Refine it so the recommended action is diagnosis-driven and specific to the observed weakness.

Examples:

- Intent missing or placeholder:
  `Replace placeholder intent statements with 1-3 measurable outcomes.`
- Assurance weak:
  `Add assurance evidence that maps test or evaluation results to BEHAVIOR-001.`
- Behavior disconnected:
  `Link BEHAVIOR-001 to its supporting INTENT and ASSURANCE evidence.`

### 5. Surface Diagnosis In Scorecard

Update `internal/scorecard/scorecard.go` so the rendered scorecard includes:

- primary bottleneck
- why it matters
- one recommended next action
- category scores

Recommended additions to the scorecard model:

```go
type Diagnosis struct {
    PrimaryBottleneck  string   `json:"primary_bottleneck"`
    TiedBottlenecks    []string `json:"tied_bottlenecks,omitempty"`
    WhyItMatters       string   `json:"why_it_matters"`
    RecommendedAction  string   `json:"recommended_action"`
}
```

Or equivalent fields at the scorecard top level if that fits the current schema better.

Requirements:

- Text output should show the diagnosis near the top.
- Markdown output should include the diagnosis in a PR-friendly form if Markdown exists.
- JSON output should expose the diagnosis in a stable schema.

### 6. Surface Diagnosis In Explain

Update `internal/explainer/explainer.go` so `bottleneck explain` includes diagnosis framing:

- weakest category
- why it matters
- recommended next action

If `--capability` is used, preserve current capability-specific behavior. When no capability filter is given, the top of the output should include the primary diagnosis.

### 7. Missing Evidence Files Must Affect Diagnosis

Tests and logic must cover missing evidence files.

Required behavior:

- Missing assurance evidence should strongly bias diagnosis toward Assurance.
- Missing security evidence should strongly bias diagnosis toward Security.
- Missing execution evidence should reduce confidence in execution and lower its score.

If a file is missing and that validator already returns `FAIL`, the diagnosis should reflect that rather than inventing a separate missing-file rule.

## Backward Compatibility

- Existing `validate` behavior should remain intact.
- Existing capability validation should remain the source of truth.
- Existing scorecard and explain output may become richer, but should not lose the current summary fields.
- JSON output must remain stable or be versioned if the schema changes.
- Do not remove existing capability metadata, evidence, or recommendation fields unless replacing them with clearly better structured equivalents.

## Testing Requirements

Add focused Go tests.

Required test cases:

1. Single weakest category selects the correct primary bottleneck.
2. Multiple tied bottlenecks are handled deterministically.
3. All categories passing yields no primary bottleneck or an explicit healthy state.
4. Missing evidence files produce the expected weakest-category diagnosis.
5. Scorecard includes why-this-matters text for the primary bottleneck.
6. Scorecard includes one recommended next action for the primary bottleneck.
7. Explainer includes diagnosis framing when no capability filter is used.
8. Recommendation selection changes appropriately for missing, weak, stale, and disconnected evidence where those states exist.
9. JSON output includes diagnosis fields in a stable schema.

Run:

```sh
go test ./...
```

## Implementation Guidance

Recommended approach:

1. Add a diagnosis helper package, for example `internal/diagnosis`, or keep it inside `internal/scorecard` if the scope stays small.
2. Define explicit structs for category scoring and diagnosis.
3. Build category scores from `ValidationResult` status plus evidence shape:
   - status
   - message
   - details
   - missing evidence
   - evidence count
4. Reuse one shared metadata source for:
   - bottleneck label
   - why-it-matters explanation
   - default recommended action
5. Refine actions based on actual validation state rather than only static category metadata.
6. Feed diagnosis output into both scorecard and explainer.
7. Add or update tests before or alongside implementation.

Keep the logic deterministic, small, and auditable. The diagnosis should be explainable from existing validation evidence, not from opaque heuristics.

## Acceptance Criteria

- The CLI assigns a deterministic score to each BIASED category.
- The CLI identifies the weakest category as the primary bottleneck.
- Ties are handled deterministically.
- Scorecard output explains why the primary bottleneck matters.
- Scorecard output includes one top recommended next action.
- Explain output includes the same diagnosis framing when no single capability is requested.
- Missing evidence files are reflected correctly in diagnosis.
- Tests cover single weakest category, tied bottlenecks, all categories passing, and missing evidence files.
- `go test ./...` passes.

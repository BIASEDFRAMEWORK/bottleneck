# AI Implementation Prompt: Epic 2 Evidence Quality Scoring

You are working in the `bottleneck` Go CLI repository.

Implement **Epic 2: Improve Scoring Beyond File Existence**.

This epic covers:

- Task 2.1: Detect placeholder content
- Task 2.2: Add thin-evidence scoring
- Task 2.3: Add traceability scoring

## Product Goal

Make Bottleneck score evidence quality, not just artifact existence.

A project should not look healthy because `intent.md`, `behavior-spec.md`, and `results.json` exist. The CLI should penalize placeholder-heavy content, header-only documents, missing evidence IDs, weak measurable language, and disconnected evidence relationships.

Target output:

```text
Intent: Weak

Reason:
intent.md exists, but still contains starter placeholder content.
```

```text
Traceability Gap:
BEHAVIOR-001 exists, but no assurance result references it.
```

## Current Architecture To Respect

Use existing validators, scorecard, explainer, and diagnosis logic. Do not create a separate scoring path that bypasses validation evidence.

Relevant files:

- `internal/validator/content_quality.go`
  - Already contains basic placeholder and thin-content detection.
- `internal/validator/behavior.go`
  - Validates behavior structure and content quality.
- `internal/validator/intent.go`
  - Validates intent structure and content quality.
- `internal/validator/design.go`
  - Validates design structure and content quality.
- `internal/validator/traceability.go`
  - If present, reuse existing ID/reference parsing and findings.
- `internal/validator/engine.go`
  - Registers validators and returns `models.EngineResult`.
- `internal/models/result.go`
  - Defines `ValidationResult`, `ValidationFinding`, and status values.
- `internal/scorecard/scorecard.go`
  - Builds capability scorecard output.
- `internal/explainer/explainer.go`
  - Renders evidence and recommendations.
- `tasks/implement-epic-1-primary-bottleneck-diagnosis-prompt.md`
  - If Epic 1 is implemented, integrate these quality signals into category scoring and diagnosis.

Current code may already detect some placeholders and traceability gaps. This epic should deepen that into explicit evidence-quality scoring and score impact.

## Required Behavior

### 1. Detect Placeholder Content

Detect placeholder-heavy evidence in Markdown artifacts.

Required placeholder phrases:

- `Describe required outcomes`
- `Describe system constraints`
- `TODO`
- `TBD`
- `Add measurable success criteria`

Also keep support for existing starter phrases if already implemented:

- `Describe intended system behavior.`
- `Describe behavior the system must prevent.`
- `Describe measurable success criteria.`
- `Describe system architecture.`

Behavior requirements:

- Placeholder content should produce a warning by default.
- Placeholder content should produce a failure in strict mode.
- Files that contain mostly placeholder content should receive a score penalty.
- Output should identify the affected file and section when possible.

Example:

```text
Intent: Weak

Reason:
bottleneck/intent/intent.md exists, but still contains starter placeholder content.
```

### 2. Penalize Placeholder-Heavy Files

Add a deterministic placeholder-density calculation.

Recommended approach:

- Count meaningful non-placeholder tokens.
- Count placeholder tokens or placeholder lines.
- Mark a file as placeholder-heavy when placeholder content is the dominant meaningful content.
- Keep section-specific messages when the placeholder appears inside a known required section.

Do not use AI scoring or semantic inference.

Example details:

```text
bottleneck/intent/intent.md section "Outcomes" still contains placeholder content
bottleneck/intent/intent.md is placeholder-heavy
```

### 3. Add Thin-Evidence Scoring

A file should not pass simply because it exists.

Add minimum evidence requirements per artifact type:

- Intent should include required sections, at least one `INTENT-*` ID, and measurable outcome language.
- Behavior should include expected and unacceptable behavior, at least one `BEHAVIOR-*` ID, and concrete behavior statements.
- Design should include a `DESIGN-*` ID or enough concrete architecture content to be reviewable.
- Assurance should include scenario or test evidence and preferably `ASSURANCE-*` IDs or behavior refs.
- Security should include guardrail or scan evidence and preferably `SECURITY-*` IDs or refs.
- Execution should include telemetry values and preferably `EXECUTION-*` IDs or refs.

Minimum scoring factors:

- Empty files score lowest.
- Header-only files score low.
- Placeholder files score low.
- Required sections with insufficient body text score low.
- Meaningful files with IDs, refs, and measurable statements score higher.

Expected tests:

- Empty files.
- Header-only files.
- Placeholder files.
- Meaningful files.

### 4. Check Expected Sections

Preserve existing structural validation, then add score penalties.

Required Markdown checks:

- `intent/intent.md` includes `## Outcomes`, `## Constraints`, and `## Success Criteria`.
- `behavior/behavior-spec.md` includes `## Expected Behavior` and `## Unacceptable Behavior`.
- `design/architecture.md` includes at least one Markdown section header.

Missing required sections should remain a validation failure.

Weak section bodies should affect evidence quality and scoring.

### 5. Check Evidence IDs

Detect evidence IDs in Markdown and JSON.

Supported ID format:

```text
^(INTENT|BEHAVIOR|DESIGN|ASSURANCE|SECURITY|EXECUTION)-[0-9]{3,}$
```

Score impact:

- Missing expected IDs should reduce the score.
- Duplicate IDs should fail or strongly penalize traceability.
- Invalid ID syntax should fail or strongly penalize traceability.

Expected ID locations:

- Markdown headings such as `### INTENT-001: Reduce release risk`.
- Markdown references under `Refs:` or `References:`.
- JSON evidence arrays with `id` and `refs`.

### 6. Check Measurable Language

Intent and execution-adjacent evidence should be penalized when outcomes are vague.

Use deterministic heuristics only.

Examples of measurable signals:

- numbers or percentages
- explicit thresholds
- dates or time windows
- words such as `at least`, `no more than`, `below`, `above`, `within`, `less than`, `greater than`

Examples of weak language:

- `better`
- `improve`
- `fast`
- `easy`
- `robust`
- `user-friendly`

Do not fail solely because measurable language is missing. Penalize score and explain the weakness.

Example:

```text
bottleneck/intent/intent.md section "Success Criteria" does not include measurable criteria
```

### 7. Add Traceability Scoring

Bottleneck should detect when intent, behavior, assurance, security, and execution are disconnected.

Required graph checks:

- Parse evidence IDs from Markdown.
- Parse evidence IDs from JSON.
- Validate references across files.
- Detect orphaned intent statements.
- Detect behavior specs with no assurance evidence.
- Detect security guardrails not mapped to behavior or intent.
- Detect execution metrics not tied to release readiness.

Required scoring penalties:

- Broken references produce a strong penalty.
- Orphaned behavior produces a strong penalty.
- Behavior without assurance evidence produces a strong penalty.
- Security evidence with no mapping produces a moderate penalty.
- Execution telemetry with no mapping produces a moderate penalty.

Example:

```text
Traceability Gap:
BEHAVIOR-001 exists, but no assurance result references it.
```

### 8. Feed Quality Signals Into Diagnosis And Scorecard

If Epic 1 diagnosis scoring exists, integrate these signals into category scores.

Expected score impacts:

- Placeholder-heavy intent lowers Intent.
- Missing behavior IDs lowers Behavior.
- Critical behavior without assurance lowers Assurance and Traceability.
- Broken refs lower Traceability and the affected categories.
- Security evidence not mapped to behavior or intent lowers Security.
- Execution metrics not tied to release readiness lowers Execution.

Scorecard and explain output should show:

- evidence found
- evidence missing
- reason
- score impact when available
- recommended next action

Do not hide these as generic `WARNING`; make the weakness visible.

## Backward Compatibility

- Existing `validate` behavior must remain intact.
- Missing required files and sections should still fail as before.
- Basic artifacts without IDs should not crash parsing.
- Existing simple JSON artifacts should remain valid.
- Strict mode should keep turning content-quality warnings into failures.
- If score fields are added to JSON output, keep the scorecard schema explicit and versioned.

## Testing Requirements

Add focused Go tests with temp directories and small fixtures.

Required test cases:

1. Placeholder phrases are detected:
   `Describe required outcomes`, `Describe system constraints`, `TODO`, `TBD`, and `Add measurable success criteria`.
2. Placeholder-heavy files receive a score penalty and warning detail.
3. Empty files score low.
4. Header-only files score low.
5. Placeholder files score low.
6. Meaningful files score higher.
7. Expected sections are checked before quality scoring.
8. Missing `INTENT-001`, `BEHAVIOR-001`, or related evidence IDs lowers score.
9. Measurable intent language raises or preserves score.
10. Vague intent language lowers score.
11. Markdown evidence IDs parse correctly.
12. JSON evidence IDs parse correctly.
13. Broken references are reported.
14. Orphaned intent is reported.
15. Behavior without assurance evidence is reported.
16. Security guardrails not mapped to behavior or intent are reported.
17. Execution metrics not tied to release readiness are reported.
18. Valid relationships avoid traceability penalties.

Run:

```sh
go test ./...
```

## Implementation Guidance

Recommended approach:

1. Extend existing content-quality helpers instead of creating duplicate placeholder scanners.
2. Add an evidence-quality model, for example:

```go
type EvidenceQuality struct {
    Score         int
    Findings      []models.ValidationFinding
    Details       []string
    Missing       []string
    ScoreImpacts  []ScoreImpact
}
```

3. Keep score impact deterministic and small:

```go
type ScoreImpact struct {
    Reason string
    Delta  int
}
```

4. Reuse existing traceability parsing if present.
5. Add missing helpers only where the repo does not already have them.
6. Feed quality output into `ValidationResult.Details` and `Findings`.
7. Feed quality output into scorecard and diagnosis scoring when those models exist.
8. Update docs only where command behavior visibly changes.

Prefer small, composable helpers:

- placeholder detection
- thin-content detection
- ID extraction
- measurable-language detection
- traceability scoring

Avoid large, untestable scoring functions.

## Acceptance Criteria

- Starter placeholder content no longer scores as meaningful evidence.
- Placeholder-heavy files produce clear warning output.
- Empty, header-only, placeholder, and meaningful files receive different scores.
- Expected sections, evidence IDs, and measurable language affect evidence quality.
- Broken traceability lowers scores and is visible in scorecard or explain output.
- `BEHAVIOR-001` with no assurance evidence is reported as a traceability gap.
- Tests cover placeholder-heavy files, thin evidence, missing IDs, measurable language, missing relationships, and valid relationships.
- Existing validation behavior remains backward compatible unless strict mode is enabled.
- `go test ./...` passes.

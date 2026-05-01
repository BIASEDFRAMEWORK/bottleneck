# Codex Prompt 03: Implement `bottleneck assess`

## Objective

Add a new day-one command that gives developers a maturity-oriented SDLC assessment with minimal onboarding burden.

`bottleneck assess` should become the primary command for new users.

## Required behavior

Add:

```bash
bottleneck assess
```

The command should orchestrate existing capabilities instead of duplicating logic:

1. Run discovery.
2. Optionally run auto-ingest when supported evidence is found.
3. Run scorecard logic.
4. Run diagnosis logic.
5. Render a concise maturity-oriented summary.

## Flags

Add these flags:

```bash
bottleneck assess --no-ingest
bottleneck assess --format text
bottleneck assess --format json
bottleneck assess --strict
bottleneck assess --environment dev
bottleneck assess --environment stage
bottleneck assess --environment production
```

Behavior:

- Default format is `text`.
- Default behavior should run auto-ingest if Prompt 02 functionality exists.
- `--no-ingest` means discover and assess existing evidence only.
- `--strict` should reuse existing strict validation behavior if present.
- `--environment` should pass through to existing config/scoring logic if supported; otherwise store in output and avoid breaking.

## Target text output

Example:

```text
Bottleneck SDLC Maturity Assessment

Overall Maturity: Level 2 - Managed
AI Readiness: Limited
Release Friction: Medium
Primary Bottleneck: Assurance
Score Confidence: Medium
Release Recommendation: Conditional

What Bottleneck Found:
  ✓ GitHub Actions workflow detected
  ✓ Cucumber test report detected and ingested
  ✓ SARIF security report detected and ingested
  ⚠ No production telemetry freshness signal detected
  ⚠ 1 behavior has no mapped passing assurance evidence

Why This Matters:
  The team has automated evidence, but release confidence still depends on incomplete behavior-to-test traceability.

Next Action:
  Add or ingest assurance evidence mapped to BEHAVIOR-003.

Useful Commands:
  bottleneck trace BEHAVIOR-003
  bottleneck explain-score
  bottleneck report
```

Do not make the output too verbose. `explain-score` will be the detailed command.

## JSON output shape

Return a stable JSON object suitable for CI:

```json
{
  "overall_maturity": {
    "level": 2,
    "label": "Managed",
    "summary": "Automated evidence exists, but traceability is incomplete."
  },
  "ai_readiness": "limited",
  "release_friction": "medium",
  "primary_bottleneck": "Assurance",
  "score_confidence": "medium",
  "release_recommendation": "conditional",
  "found": [],
  "warnings": [],
  "next_action": "Add or ingest assurance evidence mapped to BEHAVIOR-003.",
  "useful_commands": [
    "bottleneck trace BEHAVIOR-003",
    "bottleneck explain-score",
    "bottleneck report"
  ]
}
```

Use existing names/statuses where possible.

## Maturity-level placeholder rules

If Prompt 04 has not yet implemented maturity scoring, implement a simple placeholder mapping using existing scorecard/validation results.

Suggested v1 mapping:

```text
Level 0 - Ad Hoc
  Most required evidence is missing.

Level 1 - Documented
  Intent/behavior/design artifacts exist, but automated validation evidence is thin or missing.

Level 2 - Managed
  Automated test/security/CI evidence exists, but traceability or telemetry is incomplete.

Level 3 - Measured
  Automated evidence is present, traceable, and fresh enough for release decisions.

Level 4 - Optimized
  Evidence is automated, trendable, traceable, and consistently improving over snapshots.
```

Do not over-engineer maturity scoring here. Prompt 04 will strengthen scoring trust.

## AI readiness rules

Add simple derived labels:

```text
Blocked
  Critical security or assurance failure, or overall maturity Level 0.

Limited
  Some automated evidence exists, but traceability, telemetry, or confidence is incomplete.

Ready With Guardrails
  Automated evidence exists, no critical blockers, traceability is mostly complete.

Strong
  Level 4 maturity, trend history exists, and no release-blocking findings.
```

This should be deterministic and testable.

## Files likely to change

- new `cmd/assess.go`
- new `cmd/assess_test.go`
- possibly new `internal/assess/*`
- `cmd/root.go` command list and long help text
- reuse `internal/discover`, `internal/ingest`, `internal/scorecard`, `internal/validator`, or current packages as much as possible

## Implementation constraints

- Do not duplicate all scorecard or diagnose logic. Wrap or reuse existing functions.
- Do not break `scorecard`; it remains the detailed release-readiness surface.
- Do not make `assess` fail just because no evidence is found. It should return a low maturity assessment with next actions.
- `--format json` must emit valid JSON only, with no extra text.
- Text output should be stable enough for tests but friendly for developers.

## Tests to add

Required tests:

1. `assess` works in an empty temp repo and returns Level 0 or equivalent low maturity with next action.
2. `assess` detects and/or ingests Cucumber and SARIF evidence when present.
3. `assess --no-ingest` does not create Bottleneck evidence files.
4. `assess --format json` emits valid JSON with maturity, AI readiness, primary bottleneck, score confidence, release recommendation, and next action.
5. `assess` includes useful commands in text output.
6. Existing command tests still pass.

## Acceptance criteria

- `go test ./...` passes.
- `bottleneck assess` works after `bottleneck init --template saas`.
- New users can understand the main bottleneck without knowing the entire BIASED model.
- Existing scorecard/diagnose behavior remains intact.

## Commit message suggestion

```text
Add maturity assessment command
```

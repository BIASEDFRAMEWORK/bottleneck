# Codex Prompt 04: Implement score trust, maturity levels, and `bottleneck explain-score`

## Objective

Make Bottleneck's scoring more trustworthy by separating maturity, release recommendation, primary bottleneck, score confidence, and AI readiness.

Add a detailed command that explains exactly how scores were calculated.

## Required behavior

Add:

```bash
bottleneck explain-score
```

This command should explain scoring in a deterministic, developer-trustworthy way.

Also update score/assessment models so `bottleneck assess` and/or `bottleneck scorecard` can include:

- SDLC maturity level
- Release recommendation
- Primary bottleneck
- Score confidence
- AI readiness
- Score rationale

## Conceptual distinction

Do not collapse these into one number:

```text
SDLC maturity level = how developed the delivery system is.
Release recommendation = whether current evidence supports release.
Primary bottleneck = the weakest constraint right now.
Score confidence = how trustworthy the assessment is.
AI readiness = whether the SDLC can safely absorb AI-generated change.
```

## Maturity levels

Implement these levels:

```text
Level 0 - Ad Hoc
Evidence is mostly missing or manually asserted.

Level 1 - Documented
Intent, behavior, and design evidence exist, but validation is weak or disconnected.

Level 2 - Managed
Tests, security scans, and CI evidence exist, but traceability is incomplete.

Level 3 - Measured
Evidence is automated, traceable, fresh, and usable for release decisions.

Level 4 - Optimized
The SDLC detects bottlenecks early, trends improve over time, and AI-assisted development can operate within guardrails.
```

Implement an overall maturity result and category-level maturity where practical:

- Behavior
- Intent
- Design
- Assurance
- Security
- Execution

## Score confidence

Implement score confidence as:

```text
High
  Most evidence is tool-generated, traceable, and fresh.

Medium
  Some evidence is tool-generated, but traceability or freshness is incomplete.

Low
  Evidence is mostly missing, manual, stale, or inferred.
```

Inputs to confidence should include:

- presence of provenance fields
- `provenance` values: `tool-generated`, `manual`, `inferred`
- `confidence` values on evidence items
- missing refs
- stale evidence if timestamps exist
- missing generated evidence

Do not overcomplicate freshness initially. A simple timestamp check is acceptable if existing data supports it.

Suggested freshness defaults:

```text
assurance evidence: stale after 14 days
security evidence: stale after 14 days
execution/telemetry evidence: stale after 7 days
```

Make thresholds configurable later if not already easy.

## Release recommendation

Keep or align with existing release recommendations, but make sure they remain distinct from maturity.

Suggested values:

```text
Ready
Conditional
Blocked
Unknown
```

Rules:

- Critical security or assurance failure should produce `Blocked`.
- Missing important evidence should usually produce `Conditional`, not automatically `Blocked`, unless strict/release gate mode requires it.
- Empty or uninitialized repo should produce `Unknown` or `Blocked` depending on current scoring conventions. Prefer `Unknown` for assessment, `Blocked` for release gate.

## AI readiness

Add deterministic labels:

```text
Blocked
Limited
Ready With Guardrails
Strong
```

Suggested rules:

- `Blocked`: Level 0 or release recommendation is Blocked.
- `Limited`: Level 1 or Level 2 with Medium/Low confidence.
- `Ready With Guardrails`: Level 2/3 with Medium/High confidence and no blockers.
- `Strong`: Level 4 with High confidence and trend/history evidence.

## `bottleneck explain-score` output

Example:

```text
Bottleneck score explanation

Overall SDLC Maturity: Level 2 - Managed
Score Confidence: Medium
Release Recommendation: Conditional
AI Readiness: Limited
Primary Bottleneck: Assurance

How this was calculated:

Behavior: Level 2 - Managed
  Score: 80
  Evidence:
    ✓ behavior-spec.md found
    ✓ 3 behavior IDs found
    ⚠ 1 behavior lacks mapped assurance evidence
  Score impact:
    -20 incomplete behavior-to-test traceability

Assurance: Level 2 - Managed
  Score: 65
  Evidence:
    ✓ Cucumber evidence found: reports/cucumber.json
    ✓ 5 scenarios passed
    ⚠ BEHAVIOR-003 has no mapped passing test
  Score impact:
    -25 unmapped behavior evidence
  Confidence: Medium
    Some evidence is tool-generated, but traceability is incomplete.

Security: Level 3 - Measured
  Score: 90
  Evidence:
    ✓ SARIF evidence found
    ✓ 0 critical findings
    ⚠ 1 low finding
  Score impact:
    -5 low finding present
  Confidence: High
    Evidence is tool-generated and current.

Next action:
  Add or ingest assurance evidence mapped to BEHAVIOR-003.
```

## JSON output

Support:

```bash
bottleneck explain-score --format json
```

Emit structured JSON only.

Suggested shape:

```json
{
  "overall_maturity": {"level": 2, "label": "Managed", "reason": "..."},
  "score_confidence": "medium",
  "release_recommendation": "conditional",
  "ai_readiness": "limited",
  "primary_bottleneck": "Assurance",
  "categories": [
    {
      "name": "Assurance",
      "score": 65,
      "maturity": {"level": 2, "label": "Managed"},
      "confidence": "medium",
      "evidence": ["Cucumber evidence found"],
      "score_impacts": ["-25 unmapped behavior evidence"],
      "next_actions": ["Add or ingest assurance evidence mapped to BEHAVIOR-003"]
    }
  ],
  "next_action": "Add or ingest assurance evidence mapped to BEHAVIOR-003."
}
```

## Files likely to change

- new `cmd/explain_score.go`
- new `cmd/explain_score_test.go`
- new or existing `internal/scorecard/*`
- existing `internal/explainer/*` if present
- `cmd/assess.go`
- `cmd/scorecard.go` if adding optional score confidence output
- `internal/models/*`
- `internal/validator/json_quality.go`

## Implementation constraints

- Do not make scoring random or non-deterministic.
- Do not silently change existing scorecard output in a breaking way unless tests are updated deliberately.
- Additive output is preferred.
- All score impacts should be explainable as strings or structured details.
- Avoid heavy statistical scoring. Simple deterministic rules are better for v1 trust.

## Tests to add

Required tests:

1. `explain-score` emits text with maturity, score confidence, release recommendation, AI readiness, and primary bottleneck.
2. `explain-score --format json` emits valid JSON.
3. Tool-generated evidence increases confidence compared to manual/inferred evidence.
4. Missing refs lower confidence or category maturity.
5. Critical security/assurance findings can produce blocked release recommendation.
6. Empty repo produces low maturity and low confidence.
7. `assess` output includes new score confidence and maturity values.
8. Existing `scorecard`, `diagnose`, and `explain` tests still pass.
9. `go test ./...` passes.

## Acceptance criteria

- Developers can run `bottleneck explain-score` and understand why the score exists.
- Maturity and release recommendation are separate.
- Confidence is visible and based on evidence quality/provenance.
- Existing CLI workflows remain backward compatible.

## Commit message suggestion

```text
Add explainable maturity scoring
```

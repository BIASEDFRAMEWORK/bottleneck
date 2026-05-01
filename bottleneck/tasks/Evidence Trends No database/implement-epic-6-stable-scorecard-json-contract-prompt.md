# AI Implementation Prompt: Stable Scorecard JSON Contract

You are working in the Bottleneck Go CLI codebase.

Implement **Implementation Epic 6: Improve scorecard JSON stability** from the `Enterprise SDLC Evidence Package` milestone.

## Milestone Context

Snapshot, trends, report, CI, and future integrations need a stable machine-readable scorecard contract. Text output can evolve, but JSON output should be treated as an integration contract.

This epic stabilizes:

```sh
bottleneck scorecard --format=json
```

## Recommended Implementation Order Note

The task list recommends implementing this epic before snapshot, trends, and report because those features need a stable JSON contract. If some of those features already exist, update them to consume the stabilized DTO without breaking existing behavior.

## Current Code To Inspect

Read before changing code:

- `cmd/scorecard.go`
- `cmd/scorecard_test.go` or related command tests
- `internal/scorecard/*`
- `internal/gate/*`
- `internal/diagnosis/*`
- any snapshot/trends/report packages if already implemented
- existing JSON output tests
- README or docs examples of scorecard JSON

## Target JSON Shape

Formalize a scorecard DTO similar to:

```json
{
  "schema_version": "scorecard.v1",
  "environment": "default",
  "system_status": "WARN",
  "release_recommendation": "Conditional",
  "primary_bottleneck": "Assurance",
  "categories": [
    {
      "name": "Intent",
      "status": "PASS",
      "score": 90,
      "summary": "Intent evidence exists and includes measurable outcomes.",
      "evidence_found": [],
      "evidence_missing": [],
      "recommendations": []
    }
  ]
}
```

Use existing status values and release recommendation values unless the code has already standardized them. Do not change score math or gate behavior just to fill fields.

## Contract Rules

- JSON field names use `snake_case`.
- `schema_version` is explicit.
- Unknown fields may be added later.
- Required top-level fields should be stable.
- Existing JSON output should not be broken unless tests are updated intentionally and compatibility has been considered.
- Text and Markdown scorecard output must not be broken.
- Snapshot should embed this stable scorecard structure.
- Trends should parse this stable structure.
- Report should consume this stable structure.

## Required Top-Level Fields

`bottleneck scorecard --format=json` must include:

- `schema_version`
- `environment`
- `system_status`
- `release_recommendation`
- `primary_bottleneck`
- `categories`

## Required Category Fields

Each category object should include, where available:

- `name`
- `status`
- `score`
- `summary`
- `evidence_found`
- `evidence_missing`
- `recommendations`

If a value is not available, prefer an empty array or omitted optional field according to existing JSON conventions. Be consistent and test the decision.

## Suggested Implementation

Create or formalize DTOs under `internal/scorecard`, for example:

```go
type JSONScorecard struct {
    SchemaVersion        string         `json:"schema_version"`
    Environment          string         `json:"environment"`
    SystemStatus         string         `json:"system_status"`
    ReleaseRecommendation string        `json:"release_recommendation"`
    PrimaryBottleneck    string         `json:"primary_bottleneck"`
    Categories           []JSONCategory `json:"categories"`
}

type JSONCategory struct {
    Name            string   `json:"name"`
    Status          string   `json:"status"`
    Score           *float64 `json:"score,omitempty"`
    Summary         string   `json:"summary,omitempty"`
    EvidenceFound   []string `json:"evidence_found,omitempty"`
    EvidenceMissing []string `json:"evidence_missing,omitempty"`
    Recommendations []string `json:"recommendations,omitempty"`
}
```

Use exact types that fit the codebase. If scores are integers in existing model, use integer scores.

## Backwards Compatibility

Before changing output, inspect existing tests and examples. If older fields exist, consider retaining them while adding the stable contract fields.

Do not:

- rename command flags
- remove text or Markdown behavior
- change scorecard status calculation
- change release gate semantics
- reorder user-facing categories unpredictably

## Acceptance Criteria

- Scorecard JSON includes `schema_version`.
- Scorecard JSON includes `environment`.
- Scorecard JSON includes `system_status`.
- Scorecard JSON includes `release_recommendation`.
- Scorecard JSON includes `primary_bottleneck`.
- Scorecard JSON includes category objects.
- Each category includes `name`, `status`, `score` if available, `summary`, `evidence_found`, `evidence_missing`, and `recommendations` if available.
- Snapshot uses this stable scorecard structure.
- Trends uses this stable scorecard structure.
- Tests protect this JSON contract.

## Tests To Add

Add or update tests:

- `TestScorecardJSONIncludesSchemaVersion`
- `TestScorecardJSONIncludesEnvironment`
- `TestScorecardJSONIncludesPrimaryBottleneck`
- `TestScorecardJSONIncludesCategories`
- `TestScorecardJSONCategoryFields`
- `TestScorecardJSONBackwardsCompatibleForSnapshot`

Also consider:

- JSON output is valid JSON.
- Category ordering is deterministic.
- Markdown/text scorecard tests still pass.
- Snapshot/trends/report tests parse the same DTO if those features exist.

Prefer structural JSON assertions over full string snapshots.

## Implementation Constraints

- Keep the contract local and deterministic.
- Do not add external dependencies unless the repo already uses them.
- Do not introduce a database or service.
- Do not break existing commands.
- Do not change public field names after adding tests.

## Verification

Run:

```sh
go test ./...
```

If feasible, manually verify:

```sh
bottleneck scorecard --format=json
bottleneck scorecard --format=text
bottleneck scorecard --format=markdown
```

If snapshot/trends/report are implemented, also verify:

```sh
bottleneck snapshot
bottleneck trends --format=json
bottleneck report --format=json
```

## Final Response Requirements

When finished, report:

1. Stable JSON DTO or contract changes.
2. Compatibility decisions.
3. Snapshot/trends/report integration updates, if applicable.
4. Tests added or changed.
5. Exact commands run and results.
6. Any acceptance criteria intentionally deferred and why.


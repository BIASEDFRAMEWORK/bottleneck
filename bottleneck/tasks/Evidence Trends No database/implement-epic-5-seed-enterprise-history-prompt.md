# AI Implementation Prompt: Seed Enterprise History For Demos And Tests

You are working in the Bottleneck Go CLI codebase.

Implement **Implementation Epic 5: Seed enterprise history for demos and tests** from the `Enterprise SDLC Evidence Package` milestone.

## Milestone Context

Bottleneck should show SDLC evidence trends over time using local files and Git, not a database or external metrics service. This epic creates realistic historical snapshot data so teams can demo, test, and understand trend analysis without waiting weeks to accumulate history.

This epic adds:

```sh
bottleneck seed-history
```

## Scope

Add a new command:

```sh
bottleneck seed-history
```

Supported flags:

```sh
bottleneck seed-history --scenario=saas-day-one
bottleneck seed-history --env=default
bottleneck seed-history --snapshots=6
bottleneck seed-history --out=bottleneck/history/scorecards
bottleneck seed-history --overwrite
```

First implementation should support one scenario:

```text
saas-day-one
```

Future scenarios such as `ai-product`, `regulated-enterprise`, `security-regression`, and `execution-drift` should not be implemented unless trivial. Keep this slice focused.

## Current Code To Inspect

Read before changing code:

- `cmd/root.go`
- `cmd/snapshot.go`, if implemented
- `cmd/trends.go`, if implemented
- `internal/snapshot/*`
- `internal/trends/*`
- `internal/scorecard/*`
- existing testdata fixtures
- existing CLI command tests

Seed snapshots should use the same schema as real snapshots from `bottleneck snapshot`.

## Scenario: SaaS Day-One Success

Generate 6 snapshots showing SDLC evidence maturity over time.

### Snapshot 1: Fast demo, weak evidence

```text
Status: FAIL
Primary bottleneck: Intent
Intent: FAIL
Behavior: FAIL
Design: WARN
Assurance: FAIL
Security: WARN
Execution: FAIL
```

Narrative:

```text
The team has code, but cannot clearly explain what outcome the system is designed to produce.
```

### Snapshot 2: Intent clarified

```text
Status: FAIL
Primary bottleneck: Behavior
Intent: PASS
Behavior: FAIL
Design: WARN
Assurance: FAIL
Security: WARN
Execution: FAIL
```

Narrative:

```text
Intent improved, but expected and unacceptable behaviors are not measurable.
```

### Snapshot 3: Behavior documented, assurance weak

```text
Status: WARN
Primary bottleneck: Assurance
Intent: PASS
Behavior: PASS
Design: WARN
Assurance: FAIL
Security: WARN
Execution: FAIL
```

Narrative:

```text
Behavior is documented, but tests are not mapped to behavior evidence.
```

### Snapshot 4: Tests added, security regresses

```text
Status: WARN
Primary bottleneck: Security
Intent: PASS
Behavior: PASS
Design: PASS
Assurance: WARN
Security: FAIL
Execution: WARN
```

Narrative:

```text
Validation improved, but a high-severity security issue was introduced.
```

### Snapshot 5: Security recovered, execution weak

```text
Status: WARN
Primary bottleneck: Execution
Intent: PASS
Behavior: PASS
Design: PASS
Assurance: PASS
Security: PASS
Execution: WARN
```

Narrative:

```text
Pre-release evidence improved, but production/adoption evidence is incomplete.
```

### Snapshot 6: Stable release candidate

```text
Status: PASS
Primary bottleneck: None
Intent: PASS
Behavior: PASS
Design: PASS
Assurance: PASS
Security: PASS
Execution: PASS
```

Narrative:

```text
The delivery system can now prove intent, behavior, assurance, security, and execution evidence together.
```

## Output Behavior

Default output directory:

```text
bottleneck/history/scorecards
```

Use deterministic timestamps spaced across delivery cycles. For example, one snapshot per week ending at current time, or fixed relative offsets. Tests should inject or fix time so results are deterministic.

Generated snapshots must include:

- `schema_version`
- `snapshot` metadata
- `scorecard` object compatible with real snapshot output
- category statuses and scores or values parseable by `bottleneck trends`
- primary bottleneck
- narrative or summary field if the schema supports it

Do not overwrite existing files unless `--overwrite` is provided.

## CLI Output

Recommended success output:

```text
Bottleneck seed history created

Scenario: saas-day-one
Environment: default
Snapshots: 6
Output: bottleneck/history/scorecards

Next:
Run `bottleneck trends` to see SDLC evidence direction over time.
```

If files already exist and `--overwrite` is false, return a useful error:

```text
Seed history already exists in bottleneck/history/scorecards.
Next action: use --overwrite or choose a different --out path.
```

## Acceptance Criteria

- `bottleneck seed-history` creates realistic snapshot files.
- Seed snapshots use the same schema as real snapshots.
- Seed snapshots work with `bottleneck trends`.
- Seed snapshots work with `bottleneck report`.
- Default scenario is `saas-day-one`.
- Existing snapshot files are not overwritten unless `--overwrite` is passed.
- Command does not require a database or external service.

## Tests To Add

Add tests in command and internal packages as appropriate:

- `TestSeedHistoryCreatesSnapshots`
- `TestSeedHistoryUsesSnapshotSchema`
- `TestSeedHistoryDoesNotOverwriteByDefault`
- `TestSeedHistoryOverwriteFlag`
- `TestSeedHistoryWorksWithTrends`
- `TestSeedHistoryScenarioSaaSDayOne`

Also test:

- Unsupported scenario returns useful error.
- `--snapshots` less than 1 fails or is handled explicitly.
- `--env=production` writes production metadata.
- Output files are sorted correctly by timestamp.

Use temporary directories. Do not write seed snapshots into the repository root during tests.

## Implementation Constraints

- Do not add a database.
- Do not call external services.
- Do not require Git.
- Do not overwrite files without `--overwrite`.
- Do not implement unrelated seed scenarios unless they are explicitly requested.
- Reuse snapshot schema and helpers where possible.

## Verification

Run:

```sh
go test ./...
```

If feasible, manually verify:

```sh
bottleneck seed-history
bottleneck trends
bottleneck report
```

Use a temporary project directory for manual verification.

## Final Response Requirements

When finished, report:

1. Seed-history command behavior.
2. Scenario generated.
3. Snapshot schema compatibility.
4. Overwrite behavior.
5. Trends/report compatibility.
6. Tests added or changed.
7. Exact commands run and results.
8. Any acceptance criteria intentionally deferred and why.


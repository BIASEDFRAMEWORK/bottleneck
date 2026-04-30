# AI Implementation Prompt: Evidence Ingestion for Tests, Security, and Telemetry

You are working in the `bottleneck` Go CLI repository.

Implement feature 5 from the roadmap: **Evidence Ingestion for Tests, Security, and Telemetry**.

## Product Goal

Reduce manual artifact editing and ground the bottleneck scorecard in real tool outputs.

Teams should be able to ingest BDD results, security scan output, generic test summaries, and telemetry snapshots into normalized bottleneck evidence artifacts. bottleneck should not replace those tools. It should read their output, normalize it, and use it in validation, scorecard, explain, diagnose, and trace flows.

## Current Architecture To Respect

Use the existing CLI and artifact model. Do not add a database, network service, or background daemon.

Relevant files:

- `cmd/root.go`
  - Registers Cobra commands.
- `cmd/validate.go`
  - Runs validation and exits non-zero on `FAIL`.
- `cmd/scorecard.go`
  - Renders scorecards.
- `cmd/explain.go`
  - Explains validation state.
- `cmd/init.go`
  - Creates default framework artifacts.
- `internal/validator/assurance.go`
  - Validates `biased/assurance/results.json`.
- `internal/validator/security.go`
  - Validates `biased/security/guardrails.json`.
- `internal/validator/execution.go`
  - Validates `biased/execution/telemetry.json`.
- `internal/config/config.go`
  - Resolves environment thresholds.
- `internal/models/result.go`
  - Defines validation status and result types.
- `readme.md`
  - Update with ingestion examples.
- `bottleneck/docs/validation.md`
  - Update generated validation guidance if command behavior changes.

If prior features added score depth, traceability, strict mode, or diagnose commands, integrate ingestion output with those paths without duplicating logic.

## Required Commands

Add an `ingest` command group:

```sh
bottleneck ingest cucumber --file <path>
bottleneck ingest codeql --file <path>
bottleneck ingest test-summary --file <path>
bottleneck ingest telemetry --file <path>
```

Recommended shared flags:

- `--file <path>`: required input file.
- `--out <path>`: optional explicit output artifact path.
- `--dry-run`: parse and print normalized evidence without writing files.
- `--format json|text`: output format for dry-run or command summary.
- `--merge`: merge with existing artifact evidence instead of replacing normalized evidence.

Default output artifacts:

- Cucumber and generic tests write to `biased/assurance/results.json`.
- CodeQL writes to `biased/security/guardrails.json`.
- Telemetry writes to `biased/execution/telemetry.json`.

## Required Behavior

### 1. Ingest Cucumber BDD Results

Add:

```sh
bottleneck ingest cucumber --file <path>
```

Support common Cucumber JSON shape:

- feature array
- elements or scenarios
- steps
- result status
- tags

Behavior requirements:

- Count total scenarios.
- Count passed scenarios.
- Count failed scenarios.
- Capture failed scenario names.
- Map scenarios to behavior IDs through tags such as `@BEHAVIOR-001`.
- Include unmapped scenario warnings.
- Preserve enough source metadata to explain where evidence came from.

Normalized assurance artifact example:

```json
{
  "scenarios_total": 3,
  "scenarios_passed": 2,
  "scenarios_failed": 1,
  "failures": [
    "Ambiguous risk clause is summarized as fact"
  ],
  "evidence": [
    {
      "id": "ASSURANCE-001",
      "type": "cucumber",
      "source": "reports/cucumber.json",
      "refs": ["BEHAVIOR-001"],
      "status": "fail",
      "summary": "Ambiguous risk clause is summarized as fact"
    }
  ]
}
```

Acceptance behavior:

- Failed scenarios reduce Assurance score through existing assurance validation.
- Behavior specs without matching scenario evidence are visible to scorecard, explain, diagnose, or trace if traceability exists.

### 2. Ingest CodeQL SARIF Security Results

Add:

```sh
bottleneck ingest codeql --file <path>
```

Support SARIF 2.1.0.

Behavior requirements:

- Parse `runs[].results[]`.
- Count findings by severity.
- Support severity from SARIF properties where available:
  - `security-severity`
  - `severity`
  - `level`
- Preserve rule ID, message, file path, and line where available.
- Write normalized security evidence to `biased/security/guardrails.json`.

Recommended normalized security artifact:

```json
{
  "violations": 1,
  "findings": {
    "critical": 0,
    "high": 1,
    "medium": 0,
    "low": 0,
    "note": 0
  },
  "evidence": [
    {
      "id": "SECURITY-001",
      "type": "codeql",
      "source": "results/codeql.sarif",
      "severity": "high",
      "rule_id": "go/sql-injection",
      "path": "internal/db/query.go",
      "line": 42,
      "status": "fail",
      "summary": "Database query built from user-controlled input"
    }
  ]
}
```

Acceptance behavior:

- CodeQL findings affect Security score based on severity and selected environment.
- Existing `violations > 0` behavior remains valid.
- If config adds severity thresholds, validation should use them without breaking current simple schemas.

### 3. Ingest Generic Test Summary

Add:

```sh
bottleneck ingest test-summary --file <path>
```

Support a deterministic generic JSON schema:

```json
{
  "tests_total": 100,
  "tests_passed": 96,
  "tests_failed": 2,
  "tests_skipped": 2,
  "coverage": 0.84,
  "source": "go test ./...",
  "evidence": [
    {
      "id": "ASSURANCE-002",
      "refs": ["BEHAVIOR-001"],
      "summary": "Unit test summary"
    }
  ]
}
```

Behavior requirements:

- Validate required numeric fields.
- Convert failed tests into Assurance evidence.
- Include skipped tests and coverage as supporting evidence.
- Merge with Cucumber assurance evidence when `--merge` is set.
- Do not pretend unit test coverage proves behavior unless a behavior ref is explicitly present.

Acceptance behavior:

- Failed tests reduce Assurance score.
- Coverage can warn or fail based on environment thresholds if configured.

### 4. Ingest Telemetry Evidence

Add:

```sh
bottleneck ingest telemetry --file <path>
```

Support a deterministic telemetry JSON schema:

```json
{
  "source_environment": "production",
  "window_start": "2026-04-01T00:00:00Z",
  "window_end": "2026-04-30T00:00:00Z",
  "adoption_rate": 0.62,
  "error_rate": 0.04,
  "latency_p95_ms": 850,
  "rollback_rate": 0.02,
  "user_override_rate": 0.12,
  "cost": {
    "total_usd": 120.50,
    "unit_cost_usd": 0.04
  },
  "evidence": [
    {
      "id": "EXECUTION-001",
      "refs": ["BEHAVIOR-001", "ASSURANCE-001"],
      "summary": "Production telemetry snapshot"
    }
  ]
}
```

Behavior requirements:

- Validate `adoption_rate` and `error_rate` because current execution validation depends on them.
- Preserve optional latency, rollback, override, cost, and source environment fields.
- Mark telemetry source as production, staging, synthetic, or unknown.
- Warn or fail stale telemetry when timestamps exist and config defines freshness thresholds.

Acceptance behavior:

- Production telemetry can warn or fail Execution even when tests pass.
- Execution validation remains backward compatible with simple `adoption_rate` and `error_rate` artifacts.

### 5. Store Normalized Evidence Without Replacing Source Tools

Ingestion commands should write normalized bottleneck artifacts while preserving source metadata.

Rules:

- Store source file path.
- Store ingestion timestamp.
- Store source tool type.
- Store references to evidence IDs when available.
- Keep schemas simple and deterministic.
- Do not delete existing manually entered evidence unless explicitly replacing the normalized section.
- `--dry-run` must not write files.

### 6. Configuration and Thresholds

Extend `config.yaml` only where needed and keep defaults backward compatible.

Potential threshold additions:

```yaml
environments:
  default:
    assurance:
      min_accuracy: 0.90
      max_failures: 0
      min_coverage: 0.80
    security:
      max_critical: 0
      max_high: 0
      max_medium: 5
    execution:
      max_error_rate: 0.05
      min_adoption: 0.5
      max_latency_p95_ms: 1000
      max_rollback_rate: 0.05
      max_user_override_rate: 0.20
      max_telemetry_age_days: 30
```

If config structs are extended, ensure missing values inherit safely and existing config files still load.

## Backward Compatibility

- Existing simple artifacts must still validate:
  - `biased/assurance/results.json`
  - `biased/security/guardrails.json`
  - `biased/execution/telemetry.json`
- Existing `validate`, `explain`, and `scorecard` behavior should remain intact.
- Ingestion should add richer evidence without requiring every user to adopt every schema.
- No command should require network access.

## Testing Requirements

Add focused tests with fixtures.

Required test cases:

1. Cucumber parser counts total, passed, and failed scenarios.
2. Cucumber parser maps `@BEHAVIOR-001` tags to assurance evidence refs.
3. Cucumber failed scenarios reduce Assurance score.
4. Cucumber unmapped scenarios produce warnings or missing-evidence findings.
5. CodeQL SARIF parser counts findings by severity.
6. CodeQL high or critical findings affect Security based on environment thresholds.
7. Generic test summary parser validates totals, failures, skipped tests, and coverage.
8. Generic test failures reduce Assurance score.
9. Telemetry parser validates adoption and error rates.
10. Production telemetry can warn or fail Execution even when Assurance passes.
11. `--dry-run` prints normalized evidence and writes no files.
12. `--merge` preserves existing evidence and appends new normalized evidence.
13. Invalid input files return useful errors.
14. Existing simple artifacts still validate.

Run:

```sh
go test ./...
```

## Implementation Guidance

Recommended approach:

1. Add `cmd/ingest.go` with subcommands for `cucumber`, `codeql`, `test-summary`, and `telemetry`.
2. Add `internal/ingest` with focused parsers and normalizers.
3. Define explicit normalized evidence structs.
4. Write artifacts using `encoding/json` with stable indentation.
5. Add threshold support carefully in config structs if needed.
6. Reuse existing validators so ingested evidence changes existing scores naturally.
7. Integrate evidence IDs with traceability if that feature exists.
8. Update docs and add fixtures under `testdata/` or package-local `testdata/`.
9. Add tests before or alongside implementation.

Keep ingestion deterministic. Do not infer relationships from prose. Use explicit tags, refs, and fields.

## Acceptance Criteria

- `bottleneck ingest cucumber --file <path>` creates or updates Assurance evidence from BDD results.
- `bottleneck ingest codeql --file <path>` creates or updates Security evidence from SARIF results.
- `bottleneck ingest test-summary --file <path>` creates or updates Assurance evidence from generic test output.
- `bottleneck ingest telemetry --file <path>` creates or updates Execution evidence from telemetry snapshots.
- Failed tests reduce the Assurance score.
- CodeQL findings affect Security based on severity and environment.
- Production telemetry can warn or fail Execution even when tests pass.
- Ingestion commands are covered by parsing and threshold tests.
- Existing `validate`, `explain`, and `scorecard` behavior remains backward compatible unless strict mode is enabled.
- Every new score or warning links back to an artifact, threshold, or ingested evidence source.
- `go test ./...` passes.

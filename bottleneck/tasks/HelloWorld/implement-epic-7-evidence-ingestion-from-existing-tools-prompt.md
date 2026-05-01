# AI Implementation Prompt: Epic 7 Evidence Ingestion From Existing Tools

You are working in the `bottleneck` Go CLI repository.

Implement **Epic 7: Add Evidence Ingestion From Existing Tools**.

This epic covers:

- Task 7.1: Support BDD/Cucumber result ingestion
- Task 7.2: Support CodeQL or security scan ingestion
- Task 7.3: Support telemetry evidence ingestion

## Product Goal

Bottleneck scores should be grounded in real evidence produced by existing tools.

Assurance scores should reflect actual BDD/Cucumber test output. Security scores should reflect real SARIF scan results such as CodeQL. Execution scores should use deployment, reliability, adoption, override, and cost telemetry where available.

Preserve existing behavior when evidence files are missing unless the selected validation or scoring mode explicitly requires evidence.

## Current Architecture To Respect

Inspect the repository before changing code. Use existing scoring, validation, configuration, fixture, and CLI patterns rather than inventing parallel systems.

Relevant artifacts in the current framework include:

- `behavior/behavior-spec.md`
- `assurance/results.json`
- `security/guardrails.json`
- `execution/telemetry.json`
- `config.yaml`
- `docs/validation.md`

Likely implementation areas may include:

- CLI commands under `cmd/`
- validation or scoring packages under `internal/`
- models under `internal/models/`
- config loading under `internal/config/`
- tests and fixtures following the existing test layout

Do not add a database, service, daemon, or network dependency. Evidence ingestion should be deterministic file parsing.

## Required Behavior

### Task 7.1: Support BDD/Cucumber Result Ingestion

Goal: Assurance scores come from actual test output.

Support Cucumber JSON result files.

Requirements:

- Parse common Cucumber JSON output.
- Support features containing scenarios or elements.
- Support scenario tags, step results, names, and statuses.
- Map scenarios to behavior IDs through tags such as `@BEHAVIOR-001`.
- Treat a scenario as passing only when all relevant steps pass.
- Treat failed, undefined, skipped, ambiguous, errored, or missing step status as non-passing unless existing project behavior defines a clearer rule.
- Update the Assurance score based on matched scenario pass/fail results.
- Detect behavior specs with no matching scenario evidence.
- Report unmatched behavior specs through the existing validation, scorecard, explain, or diagnosis style.
- Add a sample Cucumber result fixture.
- Add tests.

Example Cucumber scenario tag:

```gherkin
@BEHAVIOR-001
Scenario: Ambiguous risk clause is flagged
```

Recommended normalized evidence shape, adjusted to match existing models:

```json
{
  "scenarios_total": 2,
  "scenarios_passed": 1,
  "scenarios_failed": 1,
  "evidence": [
    {
      "id": "ASSURANCE-001",
      "type": "cucumber",
      "source": "fixtures/cucumber-results.json",
      "refs": ["BEHAVIOR-001"],
      "status": "pass",
      "summary": "Ambiguous risk clause is flagged"
    }
  ]
}
```

Test coverage must include:

- Behavior ID tag mapping.
- Passing scenario improves or preserves Assurance score.
- Failing scenario lowers Assurance score.
- Scenario with no behavior tag is reported or ignored according to existing conventions.
- Behavior spec with no matching scenario is detected.
- Empty, missing, or malformed Cucumber result files fail gracefully.

### Task 7.2: Support CodeQL Or Security Scan Ingestion

Goal: Security scores reflect real scan evidence.

Support SARIF input.

Requirements:

- Parse SARIF 2.1.0 files.
- Parse `runs[].results[]`.
- Count findings by severity.
- Preserve useful evidence metadata when available:
  - rule ID
  - rule name
  - result message
  - file path
  - start line
  - SARIF result level
  - source SARIF file
- Support CodeQL-style severity extraction from fields such as:
  - `result.level`
  - `result.properties.security-severity`
  - `result.properties.problem.severity`
  - `result.properties.severity`
  - `rule.properties.security-severity`
  - `rule.properties.problem.severity`
  - `rule.defaultConfiguration.level`
- Normalize severities into stable buckets such as `critical`, `high`, `medium`, `low`, and `note`.
- Map findings to the Security score.
- Add configuration for severity thresholds.
- Keep parser logic separate from scoring policy where practical.
- Add tests with sample SARIF files.

Recommended severity policy:

- Numeric `security-severity` values map approximately as:
  - `>= 9.0`: critical
  - `>= 7.0`: high
  - `>= 4.0`: medium
  - `> 0.0`: low
- SARIF `error` maps to high when no better severity exists.
- SARIF `warning` maps to medium when no better severity exists.
- SARIF `note` maps to note or low according to existing project conventions.
- Missing severity should be deterministic and visible in output.

Recommended configurable threshold shape, adjusted to match existing `config.yaml` patterns:

```yaml
environments:
  default:
    security:
      sarif:
        max_critical: 0
        max_high: 0
        max_medium: 5
        max_low: 20
        fail_on_unknown_severity: false
```

Recommended normalized evidence shape, adjusted to match existing models:

```json
{
  "violations": 1,
  "findings": {
    "critical": 0,
    "high": 1,
    "medium": 0,
    "low": 0,
    "note": 0,
    "unknown": 0
  },
  "evidence": [
    {
      "id": "SECURITY-001",
      "type": "sarif",
      "source": "fixtures/codeql.sarif",
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

Test coverage must include:

- No findings.
- Low, medium, high, and critical findings.
- Multiple SARIF runs.
- CodeQL severity fields.
- Missing severity fallback behavior.
- Configured severity threshold behavior.
- Malformed SARIF input.

### Task 7.3: Support Telemetry Evidence Ingestion

Goal: Execution scores become more meaningful.

Define a basic telemetry schema and use it to score Execution.

Required telemetry signals:

- Deployment frequency.
- Change failure rate.
- Error rate.
- User override rate.
- Adoption rate.
- Cost signals.

Requirements:

- Define a stable telemetry JSON schema.
- Support deployment frequency as a count or rate over a time window.
- Support change failure rate as a percentage or ratio.
- Support error rate as a percentage, ratio, or per-request rate.
- Support user override rate as a percentage or ratio.
- Support adoption rate as a percentage or ratio.
- Support cost signals such as total cost, cost per request, budget variance, or cost trend.
- Penalize missing telemetry.
- Penalize stale telemetry.
- Make telemetry freshness configurable if existing config patterns support it.
- Produce clear validation findings for missing, stale, invalid, or poor telemetry.
- Add tests for telemetry scoring.

Recommended telemetry schema, adjusted to match existing models:

```json
{
  "generated_at": "2026-04-30T12:00:00Z",
  "window": {
    "start": "2026-04-23T00:00:00Z",
    "end": "2026-04-30T00:00:00Z"
  },
  "deployment_frequency": {
    "deployments": 7,
    "period_days": 7
  },
  "change_failure_rate": 0.05,
  "error_rate": 0.01,
  "user_override_rate": 0.03,
  "adoption_rate": 0.72,
  "cost": {
    "total": 120.5,
    "currency": "USD",
    "budget": 150,
    "trend": "stable"
  }
}
```

Recommended configurable threshold shape, adjusted to match existing `config.yaml` patterns:

```yaml
environments:
  default:
    execution:
      telemetry:
        max_age_hours: 168
        min_deployments_per_week: 1
        max_change_failure_rate: 0.15
        max_error_rate: 0.05
        max_user_override_rate: 0.10
        min_adoption_rate: 0.50
        max_budget_variance: 0.20
```

Test coverage must include:

- Healthy telemetry.
- Poor telemetry.
- Missing telemetry.
- Stale telemetry.
- Partial telemetry.
- Invalid telemetry schema.
- Cost signal penalties.
- Configured freshness threshold behavior.

## CLI And Integration Expectations

If the repository already has an ingestion command pattern, extend it. If not, add a minimal command group that follows existing Cobra command style.

Recommended commands:

```sh
bottleneck ingest cucumber --file <path>
bottleneck ingest sarif --file <path>
bottleneck ingest telemetry --file <path>
```

Recommended shared flags:

- `--file <path>`: required input file.
- `--out <path>`: optional output artifact path.
- `--dry-run`: parse and print normalized evidence without writing.
- `--format json|text`: output format for dry-run or summary output.
- `--merge`: merge with existing evidence where existing models support it.

Default output artifacts:

- Cucumber writes to `assurance/results.json`.
- SARIF writes to `security/guardrails.json`.
- Telemetry writes to `execution/telemetry.json`.

Scoring integration requirements:

- Existing scorecard, validation, explain, and diagnosis flows should consume the normalized evidence.
- Existing artifact schemas should remain backward compatible.
- Existing users without new evidence files should not see unexpected crashes.
- Invalid evidence should produce clear errors or validation findings.
- Missing required evidence should be penalized where this epic explicitly requires it.

## Fixture Expectations

Add representative fixtures in the repository's existing test fixture location. If no fixture location exists, create one that matches Go test conventions.

Required fixtures:

- Passing Cucumber result.
- Failing Cucumber result.
- Cucumber result with an unmapped behavior tag.
- SARIF with no findings.
- SARIF with mixed severity findings.
- SARIF with CodeQL-style severity metadata.
- Healthy telemetry.
- Stale telemetry.
- Poor telemetry.
- Invalid telemetry.

## Implementation Constraints

- Prefer structured JSON parsing over ad hoc string matching.
- Keep parsers deterministic and side-effect free where possible.
- Keep file writing and CLI behavior separate from parser functions.
- Keep scoring policy separate from raw evidence parsing where practical.
- Use existing model names and status values when available.
- Avoid unrelated refactors.
- Do not rewrite existing artifacts unless the command explicitly ingests into them.
- Add concise comments only where the logic is not obvious.

## Acceptance Criteria

The implementation is complete when:

- Cucumber JSON can be parsed and mapped to behavior IDs through tags like `@BEHAVIOR-001`.
- Assurance scoring reflects Cucumber pass/fail evidence.
- Behavior specs with no matching scenario are detected.
- SARIF can be parsed and findings are counted by severity.
- Security scoring reflects configured SARIF severity thresholds.
- Telemetry schema is documented or represented in code.
- Execution scoring reflects telemetry health, missing telemetry, and stale telemetry.
- Sample fixtures are committed.
- Tests cover all new ingestion and scoring paths.
- Existing tests still pass.

## Verification

Run the relevant test suite before finishing. Prefer the full Go test suite:

```sh
go test ./...
```

Also manually exercise the new ingestion commands with sample fixtures, for example:

```sh
go run . ingest cucumber --file <fixture>
go run . ingest sarif --file <fixture>
go run . ingest telemetry --file <fixture>
```

## Final Response Requirements

When finished, summarize:

- Files changed.
- Commands added or changed.
- Scoring behavior added.
- Fixtures added.
- Tests run and results.
- Any intentional compatibility decisions or remaining limitations.

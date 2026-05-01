# bottleneck

BIASED is the evidence model.
Bottleneck is the CLI that diagnoses delivery risk using that model.

Bottleneck evaluates a repo or release using local evidence artifacts. It diagnoses release readiness, hidden delivery risk, and evidence gaps before a team treats a change as ready to ship.

BIASED stands for:

- Behavior
- Intent
- Design
- Assurance
- Security
- Execution

BIASED defines what evidence matters. Bottleneck diagnoses delivery risk from local evidence artifacts today and can grow into broader scorecard surfaces as more tool signals are connected.

## What Bottleneck Evaluates

Bottleneck evaluates one repo or release at a time. Useful scopes include:

- single application repo
- service repo
- AI feature repo
- platform repo

Team and organization-level views can come later by aggregating repo scorecards.

## AI And Non-AI Systems

Bottleneck works for any software system, but it is especially useful for AI-enabled systems where behavior, drift, evaluation, and governance cannot be inferred from code alone.

Examples:

- An AI PDF Risk Summarizer needs evidence that ambiguous financial risk language is flagged instead of summarized as fact.
- A payments service needs evidence that checkout behavior, test results, security checks, and production telemetry are connected before release.

## Product And Evidence Model

BIASED is the evidence model for release decisions. It helps teams prove that software is aligned to business intent, validated against expected behavior, designed coherently, governed by policy, secured before release, and measured in production.

Bottleneck is the CLI implementation of that model. It is not a replacement for GitHub, Jira, security scanners, observability platforms, product analytics, or CI/CD. It connects local evidence back to BIASED categories and presents the result as a delivery-risk diagnosis and release-readiness scorecard.

The current implementation is a Go CLI that diagnoses delivery risk from Git-based local artifacts. Future product direction includes broader scorecard surfaces that can ingest:

- codebase structure and repository changes
- BDD and test results
- CI/CD outputs
- security, dependency, and compliance findings
- operational telemetry
- product analytics and adoption signals
- planning and product-management context
- agent-generated artifacts and review evidence

## Core Question

bottleneck answers one question:

> Where is the delivery system blocked, and what evidence proves it?

The CLI does not run tests, execute BDD frameworks, call external services, or store data in a database yet. It reads the artifacts and outcomes produced by the team and by external tools, then validates them as evidence for a release decision.

## Core Model

```text
Behavior -> Intent -> Design -> Assurance -> Security -> Execution -> Behavior
```

Execution reveals truth. Truth updates Behavior and Intent.

## CLI Commands

### `bottleneck init`

Creates the local evidence artifact structure with a realistic starter sample: `AI PDF Risk Summarizer`.

```text
bottleneck/
  config.yaml
  behavior/behavior-spec.md
  intent/intent.md
  design/architecture.md
  assurance/features/sample.feature
  assurance/results.json
  security/guardrails.json
  execution/telemetry.json
  docs/validation.md
```

First run:

```sh
bottleneck init
bottleneck diagnose
bottleneck scorecard
```

The generated sample intentionally leaves Assurance weak: one evaluation fails because ambiguous financial risk language was summarized as fact. This gives `diagnose` a real primary bottleneck to explain before you replace the starter evidence with your own system context.

Start by editing `bottleneck/intent/intent.md`, then update behavior, assurance, security, and execution artifacts with project-specific evidence.

### `bottleneck validate`

Validates the local BIASED evidence categories and prints system status.

Select thresholds with an environment:

```sh
bottleneck validate --env=production
bottleneck validate --env=production --strict --github-annotations
```

If `--env` is omitted, `default` is used. Environment-specific config inherits missing values from `default`.

Example:

```text
Behavior: PASS
Intent: PASS
Design: PASS
Assurance: FAIL (scenarios_failed above threshold)
  failure: Ambiguous risk clause was summarized as confirmed exposure
  accuracy: 0.50 (threshold: 0.90)
  scenarios_failed: 1 (allowed: 0)
Security: PASS
Execution: PASS

System Status: FAIL
Primary Bottleneck: Assurance
Environment: default
```

Exit behavior:

- Exit `0` when all evidence categories are `PASS` or `WARNING`
- Exit `1` when any evidence category is `FAIL`

### `bottleneck status`

Checks whether the local `bottleneck/` artifact directory exists.

Output:

```text
System initialized
```

or:

```text
System incomplete
```

### `bottleneck ingest`

Ingest external tool output into normalized bottleneck artifacts.

```sh
bottleneck ingest cucumber --file reports/cucumber.json
bottleneck ingest sarif --file results/codeql.sarif
bottleneck ingest codeql --file results/codeql.sarif
bottleneck ingest test-summary --file results/test-summary.json
bottleneck ingest telemetry --file results/telemetry.json
```

Use `--dry-run` to parse and inspect normalized evidence without writing artifact files, and `--out` to override the default output path. Cucumber ingestion writes assurance evidence, SARIF/CodeQL ingestion writes security evidence, and telemetry ingestion writes execution evidence.

### `bottleneck explain`

Explains the current validation state in a human-readable narrative without changing any files.

Examples:

```sh
bottleneck explain
bottleneck explain --env=production
bottleneck explain --env=production --capability=Assurance
```

The command reuses the validation engine and adds:

- primary diagnosis when no capability filter is used
- owner mapping
- mapped bottlenecks
- why-this-matters explanation
- evidence/details
- evidence-quality score, missing evidence, and score impacts
- one recommended next action

### `bottleneck diagnose`

Shows the shortest diagnosis path: primary bottleneck, top contributing findings, recommended next action, and confidence level.

Examples:

```sh
bottleneck diagnose
bottleneck diagnose --env=production
bottleneck diagnose --format=json
bottleneck diagnose --format markdown
bottleneck diagnose --format github
bottleneck diagnose --env=production --strict --gate release
```

`diagnose` exits non-zero only when validation fails. Warning-only diagnosis remains a zero exit unless `--strict` promotes warnings to failures.

Use `bottleneck diagnose --format markdown` for compact PR comments, GitHub Actions Step Summary output, or release notes. The Markdown output includes the primary bottleneck, category scores, top findings, and the recommended next action. Use `bottleneck diagnose --format github` to emit GitHub Actions annotations. Use `bottleneck diagnose --gate release` to evaluate configured release gate thresholds.

### `bottleneck scorecard`

Summarizes release readiness in an evidence-backed scorecard with deterministic diagnosis, category gauges, category scores, resolved thresholds, release recommendation, evidence counts, missing evidence, score impacts, reasons, and recommended actions.

Examples:

```sh
bottleneck scorecard
bottleneck scorecard --env=production
bottleneck scorecard --format=json
bottleneck scorecard --format=markdown
bottleneck scorecard --view=executive
bottleneck scorecard --view=engineering
bottleneck scorecard --view=governance
```

Supported formats:

- `text`
- `json`
- `markdown`

Supported views:

- `executive`: short release decision summary
- `engineering`: detailed evidence and remediation view
- `governance`: policy-oriented security, assurance, execution, and missing-governance-evidence view

Release recommendations:

- `Proceed`: all assessed required evidence categories pass
- `Conditional`: no failures, but one or more evidence categories warn
- `Block`: one or more evidence categories fail
- `Unknown`: scorecard evidence is unavailable or not assessed

Like `validate`, `scorecard` returns a non-zero exit code when the system is failing.

Diagnosis scoring is derived from validation output. Passing BIASED categories start high, warnings score in the middle, failures score low, and missing, placeholder-heavy, thin, vague, unmeasurable, or disconnected evidence reduces the score further. Missing expected evidence IDs and broken traceability refs are included in score impacts. The primary bottleneck is the weakest BIASED category using this tie priority: Assurance, Security, Behavior, Intent, Execution, Design. When all assessed BIASED categories are strong, the primary bottleneck is `None`.

When `--github-annotations` is used, `validate` and `scorecard` emit GitHub Actions workflow commands for warning and failing validation results. Failing results are emitted as `::error`; warning results are emitted as `::warning`. File paths are included when bottleneck can tie a finding to an artifact.

### `bottleneck trace`

Traces a stable evidence ID across intent, behavior, design, assurance, security, and execution artifacts.

Examples:

```sh
bottleneck trace --id INTENT-001
bottleneck trace --id BEHAVIOR-001 --env=production
bottleneck trace --id ASSURANCE-001 --format=json
```

Supported formats:

- `text`
- `json`

Unknown IDs return a non-zero exit code with a useful error.
Positional IDs such as `bottleneck trace BEHAVIOR-001` remain supported.

## GitHub Actions Integration

Copyable workflow examples live in `examples/github-actions/`:

- `bottleneck-validate.yml`: runs `bottleneck validate`
- `bottleneck-scorecard.yml`: writes a Markdown scorecard to GitHub Actions Step Summary
- `bottleneck-pr-gate.yml`: runs validation, writes Step Summary output, emits annotations, and updates a stable PR comment

Copy the workflow you want into `.github/workflows/` in the repository that owns the evidence artifacts.

The PR gate uses the existing CLI exit behavior:

```sh
go build -o bottleneck .
set +e
./bottleneck validate --env=production --strict --github-annotations
VALIDATE_EXIT=$?
./bottleneck diagnose --env=production --strict --gate release --format markdown > bottleneck-diagnosis.md
DIAGNOSE_EXIT=$?
./bottleneck diagnose --env=production --strict --format github
./bottleneck scorecard --env=production --view=governance --format=markdown > bottleneck-scorecard.md
SCORECARD_EXIT=$?
cat bottleneck-diagnosis.md >> "$GITHUB_STEP_SUMMARY"
echo >> "$GITHUB_STEP_SUMMARY"
cat bottleneck-scorecard.md >> "$GITHUB_STEP_SUMMARY"
if [ "$VALIDATE_EXIT" -ne 0 ]; then exit "$VALIDATE_EXIT"; fi
if [ "$DIAGNOSE_EXIT" -ne 0 ]; then exit "$DIAGNOSE_EXIT"; fi
if [ "$SCORECARD_EXIT" -ne 0 ]; then exit "$SCORECARD_EXIT"; fi
```

`--env=dev`, `--env=stage`, and `--env=production` select thresholds from `bottleneck/config.yaml`. `--strict` promotes placeholder and insufficient-content findings from warnings to failures, which is useful for protected PR gates. Warnings do not fail jobs unless strict mode or production traceability rules promote them to failures.

The PR comment workflow uses `actions/github-script` and the hidden marker `<!-- bottleneck-diagnosis -->` so repeated runs update one comment instead of creating duplicates. The comment can include the compact diagnosis first and the detailed scorecard below it. It needs:

- `contents: read`
- `pull-requests: write`

When running in GitHub Actions, scorecard output includes pull request metadata from the event payload. If `GITHUB_TOKEN` is available, bottleneck optionally enriches the scorecard with changed files, review approval count, pending reviewers, and failed check runs. Local scorecard usage does not require GitHub credentials or network access.

PR risk signals are deterministic and informational. bottleneck reports warnings for large PRs by file count, large diffs by additions plus deletions, draft PRs, missing requested reviewers, AI-generated or AI-assisted labels, missing approvals when review data is available, pending reviewers, failed check runs, and release-relevant source changes without matching `bottleneck/` artifact changes. bottleneck integrates with GitHub Actions, but it does not replace GitHub branch protection, CodeQL, review rules, CI, security scanners, or deployment policies.

## Validation Rules

### Behavior

Artifact: `bottleneck/behavior/behavior-spec.md`

Must exist, must not be empty, and must contain:

- `## Expected Behavior`
- `## Unacceptable Behavior`

### Intent

Artifact: `bottleneck/intent/intent.md`

Must exist and contain:

- `## Outcomes`
- `## Constraints`
- `## Success Criteria`

### Design

Artifact: `bottleneck/design/architecture.md`

Must exist, must not be empty, and must contain at least one Markdown section header.

### Assurance

Artifact: `bottleneck/assurance/results.json`

bottleneck does not run tests. BDD tests are executed externally by tools such as Cucumber, SpecFlow, Cucumber.js, or Behave. bottleneck validates their output against assurance rules defined by the BIASED model.

Required schema:

```json
{
  "scenarios_total": 1,
  "scenarios_passed": 1,
  "scenarios_failed": 0,
  "failures": []
}
```

bottleneck computes `accuracy` as `scenarios_passed / scenarios_total`. Developers maintain only this file. The CLI interprets it against the selected environment. Assurance fails when JSON is invalid, required fields are missing, `scenarios_failed` exceeds the configured `max_failures`, or calculated accuracy is below the configured `min_accuracy`.

### Configuration

Artifact: `bottleneck/config.yaml`

Defines inherited environment thresholds:

```yaml
environments:
  default:
    assurance:
      min_accuracy: 0.90
      max_failures: 0
    execution:
      max_error_rate: 0.05
      min_adoption: 0.5
      telemetry:
        max_age_hours: 168
        min_deployments_per_week: 1
        max_change_failure_rate: 0.15
        max_error_rate: 0.05
        max_user_override_rate: 0.10
        min_adoption_rate: 0.50
        max_budget_variance: 0.20
    security:
      sarif:
        max_critical: 0
        max_high: 0
        max_medium: 5
        max_low: 20
        fail_on_unknown_severity: false
    gate:
      release:
        min_primary_score: 60
        required_categories:
          - Intent
          - Behavior
          - Assurance
          - Security
          - Execution
        require_traceability: true
        require_governance: false
  production:
    assurance:
      min_accuracy: 0.95
    gate:
      release:
        min_primary_score: 75
        require_traceability: true
        require_governance: true
```

`bottleneck validate --env=<environment>` starts from `default` and overlays only the explicitly configured values for the selected environment. Gate, SARIF, and telemetry settings are optional; older configs use safe defaults.

### Security

Artifact: `bottleneck/security/guardrails.json`

Required schema:

```json
{
  "violations": 0,
  "findings": {
    "critical": 0,
    "high": 0,
    "medium": 0,
    "low": 0,
    "note": 0,
    "unknown": 0
  }
}
```

For simple guardrail evidence without a `findings` map, Security fails when `violations > 0`. SARIF-ingested evidence uses configured severity thresholds for critical, high, medium, low, and unknown findings.

### Execution

Artifact: `bottleneck/execution/telemetry.json`

Required schema:

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
  "adoption_rate": 0.72,
  "error_rate": 0.02,
  "user_override_rate": 0.03,
  "cost": {
    "total": 120.5,
    "currency": "USD",
    "budget": 150,
    "trend": "stable"
  }
}
```

Older telemetry files with only `adoption_rate` and `error_rate` still validate. Extended telemetry also checks freshness, deployment frequency, change failure rate, user override rate, and budget variance. Failing reliability thresholds return `FAIL`; partial, stale, low-adoption, high-override, or over-budget telemetry returns `WARNING`.

### Traceability

Artifacts can declare stable evidence IDs using Markdown headings or optional JSON `evidence` arrays.

Markdown convention:

```markdown
### BEHAVIOR-001: Block production release when assurance fails
Critical: true
Refs:
- INTENT-001
- ASSURANCE-001
```

JSON convention:

```json
{
  "evidence": [
    {
      "id": "ASSURANCE-001",
      "refs": ["BEHAVIOR-001"],
      "source": "cucumber",
      "status": "pass"
    }
  ]
}
```

Supported ID format:

```text
^(INTENT|BEHAVIOR|DESIGN|ASSURANCE|SECURITY|EXECUTION)-[0-9]{3,}$
```

Traceability fails on duplicate IDs, invalid IDs, invalid references, or references to missing IDs. It warns by default when behavior is not linked to intent, critical behavior is not linked to assurance, or evidence is orphaned. Those behavior and critical-assurance mapping warnings become failures in `--strict` mode or `--env=production`.

## Continuous Validation Loops

The BIASED model treats release as the beginning of continuous validation, not the end of delivery.

Core loops:

- Assurance Loop: validates behavior against expectations over time.
- Execution Loop: uses real-world telemetry to refine behavior and intent.
- Cost Loop: treats cost as a continuous operating constraint.

> Shipping is not completion. Continuous validation is completion.

## Development

Build the CLI:

```sh
go build ./...
```

Run locally:

```sh
go run . init
go run . validate
go run . validate --env=production
go run . diagnose
go run . explain
go run . scorecard
go run . status
```

## Documentation

Additional strategy and validation references:

- `bottleneck-strategy-v1.md`
- `bottleneck/docs/validation.md`
- `bottleneck_GettingStarted.md`

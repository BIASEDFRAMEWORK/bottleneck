# bottleneck

Bottleneck helps a SaaS team diagnose delivery risk before a release by reading local evidence artifacts, producing a release readiness scorecard, and naming the primary bottleneck diagnosis with the next evidence action. It does not replace tests, CI, security scanners, or observability tools; it connects their outputs so a team can see whether behavior, intent, design, assurance, security, and execution evidence are strong enough to ship.

## SaaS Team Quickstart

Create a Subscription Billing Release starter:

```sh
bottleneck init --template saas
bottleneck scorecard
bottleneck diagnose
bottleneck trace BEHAVIOR-003
```

`scorecard` is the primary Day-One command: it puts the release recommendation, primary bottleneck, category results, why it matters, and next action at the top of the terminal output. Use `validate` as a supporting evidence-quality check and `scorecard --details` when you need the raw evidence, thresholds, missing evidence, and score impacts.

The SaaS template creates:

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

The starter models payment method updates and failed invoice retries. It intentionally leaves `BEHAVIOR-003` without mapped test evidence so the first scorecard shows a warning-only release posture instead of a perfect sample.

Bad starter scorecard example:

```text
Bottleneck Scorecard
Environment: default
Release Recommendation: Conditional
Primary Bottleneck: Assurance

Category Results:
- Intent: Pass
- Behavior: Pass
- Design: Pass
- Assurance: Warn
- Security: Pass
- Execution: Pass

Why:
BEHAVIOR-003 payment retry behavior has no mapped test evidence.

Next Action:
Add assurance evidence for payment retry behavior. Map it to BEHAVIOR-003.
```

JSON scorecard excerpt:

```sh
bottleneck scorecard --format=json
```

```json
{
  "schema_version": "scorecard.v2",
  "environment": "default",
  "system_status": "WARN",
  "release_recommendation": "Conditional",
  "primary_bottleneck": "Assurance",
  "diagnosis": {
    "recommended_action": "Add assurance evidence for payment retry behavior."
  }
}
```

Good scorecard example after adding mapped assurance evidence for `BEHAVIOR-003`:

```text
Bottleneck Scorecard
Environment: default
Release Recommendation: Proceed
Primary Bottleneck: None
```

Diagnosis and trace examples:

```text
Primary Bottleneck: Assurance
Reason: BEHAVIOR-003 is not linked to any passing test evidence.
Impact: Release confidence is reduced because payment retry behavior is unproven.
Next Action: Add or ingest test evidence mapped to BEHAVIOR-003.
Inspect: bottleneck trace BEHAVIOR-003
```

```text
Trace: BEHAVIOR-003
Missing links:
- BEHAVIOR-003 has no mapped test evidence
Recommendation:
Add assurance evidence for payment retry behavior.
```

For the full walkthrough, including how to break and fix the evidence gap, see [docs/quickstart-saas.md](docs/quickstart-saas.md).

For a copyable demo project with the intentional `BEHAVIOR-003` assurance gap, sample reports, and GitHub Actions workflow, see [examples/saas-billing](examples/saas-billing).

Sample SaaS report files live in `examples/saas/reports/` so you can try ingestion immediately:

```sh
bottleneck ingest cucumber --file reports/cucumber.json
bottleneck ingest sarif --file reports/codeql.sarif
bottleneck ingest test-summary --file reports/test-summary.json
bottleneck ingest telemetry --file reports/telemetry.json
```

The Cucumber and test-summary samples write `bottleneck/assurance/results.json` and can cover `BEHAVIOR-003` with mapped `ASSURANCE-*` evidence. The SARIF sample writes `bottleneck/security/guardrails.json` and links a low-severity `SECURITY-001` finding to billing retry behavior. The telemetry sample writes `bottleneck/execution/telemetry.json` and links `EXECUTION-001` to the billing behaviors and assurance IDs.

Detailed scorecard mode:

```sh
bottleneck scorecard --details
```

Minimal CI usage with currently supported commands:

```yaml
- name: Validate Bottleneck evidence
  run: bottleneck validate

- name: Generate Bottleneck scorecard
  run: bottleneck scorecard --format=markdown >> "$GITHUB_STEP_SUMMARY"

- name: Check release readiness
  run: bottleneck diagnose --gate=release
```

Use `bottleneck diagnose --format=github` or `bottleneck validate --github-annotations` when you want GitHub Actions annotations.

BIASED is the evidence model.
Bottleneck is the CLI that diagnoses delivery risk using that model.

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

The goal is to surface hidden delivery risk while the team can still add, fix, or connect the evidence.

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

Creates the local evidence artifact structure. Use the SaaS starter for the day-one Subscription Billing Release quickstart:

```sh
bottleneck init --template saas
```

The default starter remains available:

```sh
bottleneck init
```

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

First default-starter run:

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
bottleneck ingest sarif --file reports/codeql.sarif
bottleneck ingest codeql --file reports/codeql.sarif
bottleneck ingest test-summary --file reports/test-summary.json
bottleneck ingest telemetry --file reports/telemetry.json
```

Use `examples/saas/reports/*` to try these commands without real CI artifacts. Use `--dry-run` to parse and inspect normalized evidence without writing artifact files, and `--out` to override the default output path. Cucumber and test-summary ingestion write Assurance evidence to `bottleneck/assurance/results.json`; SARIF/CodeQL ingestion writes Security evidence to `bottleneck/security/guardrails.json`; telemetry ingestion writes Execution evidence to `bottleneck/execution/telemetry.json`. Evidence IDs are generated when the source format has no ID, and `refs` preserve links such as `BEHAVIOR-003` so `bottleneck scorecard` and `bottleneck trace BEHAVIOR-003` can show category impact after ingestion.

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
bottleneck trace BEHAVIOR-003
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

- `bottleneck-saas-scorecard.yml`: Day-One SaaS workflow that validates evidence, writes a Markdown scorecard to the GitHub Actions step summary, emits annotations, and uses the release gate as the blocking check
- `bottleneck-validate.yml`: runs `bottleneck validate`
- `bottleneck-scorecard.yml`: writes a Markdown scorecard to GitHub Actions Step Summary
- `bottleneck-pr-gate.yml`: runs validation, writes Step Summary output, emits annotations, and updates a stable PR comment

Copy the workflow you want into `.github/workflows/` in the repository that owns the evidence artifacts.

For the SaaS Day-One workflow:

```sh
mkdir -p .github/workflows
cp examples/github-actions/bottleneck-saas-scorecard.yml .github/workflows/bottleneck.yml
```

In a pull request, Bottleneck writes the scorecard to the GitHub Actions step summary. Developers see the release recommendation, primary bottleneck, category results, and next action in the check run. The workflow emits GitHub annotations with `bottleneck validate --github-annotations` and `bottleneck diagnose --format=github`; warnings appear as workflow warnings and blocking failures appear as errors.

The Day-One workflow keeps validation and scorecard generation visible even when evidence is weak. The blocking CI result comes from:

```sh
./bin/bottleneck diagnose --env="$BOTTLENECK_ENV" --gate=release
```

Warnings can appear in the scorecard without failing the job when the selected environment and release gate thresholds treat them as non-blocking. Release blockers include critical evidence gaps such as missing required assurance, broken traceability, critical security findings, missing required categories, or production gate failures where configured.

Tune the environment in the workflow by changing the command or workflow input:

```sh
./bin/bottleneck scorecard --env=stage --format=markdown >> "$GITHUB_STEP_SUMMARY"
./bin/bottleneck diagnose --env=production --gate=release
```

The older PR gate example uses the existing CLI exit behavior and posts a stable PR comment:

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
        max_age_hours: 0
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
        min_primary_score: 70
        required_categories:
          - Intent
          - Behavior
          - Assurance
          - Security
          - Execution
        require_traceability: false
        require_governance: false
  local:
    assurance:
      min_accuracy: 0.75
      max_failures: 5
    security:
      sarif:
        max_critical: 1
        max_high: 1
    gate:
      release:
        min_primary_score: 60
        require_traceability: false
  dev:
    assurance:
      min_accuracy: 0.85
      max_failures: 2
    gate:
      release:
        min_primary_score: 65
        require_traceability: false
  test:
    assurance:
      min_accuracy: 0.92
    security:
      sarif:
        max_medium: 2
    gate:
      release:
        min_primary_score: 75
        require_traceability: true
  stage:
    assurance:
      min_accuracy: 0.95
    execution:
      telemetry:
        max_age_hours: 168
    security:
      sarif:
        max_medium: 1
        fail_on_unknown_severity: true
    gate:
      release:
        min_primary_score: 80
        require_traceability: true
  production:
    assurance:
      min_accuracy: 0.97
    execution:
      telemetry:
        max_age_hours: 48
        max_error_rate: 0.02
        min_adoption_rate: 0.70
    security:
      sarif:
        max_critical: 0
        max_high: 0
        max_medium: 0
        max_low: 0
        fail_on_unknown_severity: true
    gate:
      release:
        min_primary_score: 85
        require_traceability: true
        require_governance: true
```

`bottleneck validate --env=<environment>` starts from `default` and overlays only the explicitly configured values for the selected environment. Gate, SARIF, and telemetry settings are optional; older configs use safe defaults. Generated configs include `local`, `dev`, `test`, `stage`, and `production`; unknown environment names fail with a message that lists supported environments. Local and dev are tuned for fast feedback, test and stage raise assurance and security expectations, and production uses the release gate to block critical readiness gaps.

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

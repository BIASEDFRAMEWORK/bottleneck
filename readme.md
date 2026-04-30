# bottleneck

bottleneck is an SDLC scorecard application for organizations adopting AI-era software delivery. It examines a code base and the surrounding delivery system, then turns evidence from development, security, operations, and product tools into release-readiness scorecards.

BIASED is the evidence framework behind bottleneck.

BIASED stands for:

- Behavior
- Intent
- Design
- Assurance
- Security
- Execution

The framework defines what evidence matters. bottleneck delivers that evidence by validating local artifacts today and, over time, integrating with the tools organizations already use to plan, build, secure, operate, and measure software.

## Product And Framework

BIASED is an evidence framework for AI-era software delivery. It helps teams prove that software is aligned to business intent, validated against expected behavior, designed coherently, governed by policy, secured before release, and measured in production.

bottleneck is the application implementation of that framework. It is not a replacement for GitHub, Jira, security scanners, observability platforms, product analytics, or CI/CD. It connects their signals back to framework evidence categories and presents the result as an SDLC scorecard.

The current implementation is a Go CLI that validates Git-based framework artifacts. The product direction is a broader scorecard surface that can ingest:

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

The CLI does not run tests, execute BDD frameworks, call external services, or store data in a database yet. It validates the artifacts and outcomes produced by the team and by external tools.

## Core Model

```text
Behavior -> Intent -> Design -> Assurance -> Security -> Execution -> Behavior
```

Execution reveals truth. Truth updates Behavior and Intent.

## CLI Commands

### `bottleneck init`

Creates the local framework artifact structure:

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

### `bottleneck validate`

Validates all six framework capabilities and prints system status.

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
Assurance: PASS
  accuracy: 1.00 (threshold: 0.90)
  scenarios_failed: 0 (allowed: 0)
Security: PASS
Execution: PASS

System Status: PASS
Primary Bottleneck: None
Environment: default
```

Exit behavior:

- Exit `0` when all capabilities are `PASS` or `WARNING`
- Exit `1` when any capability is `FAIL`

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
bottleneck ingest codeql --file results/codeql.sarif
bottleneck ingest test-summary --file results/test-summary.json
bottleneck ingest telemetry --file results/telemetry.json
```

Use `--dry-run` to parse and inspect normalized evidence without writing artifact files, and `--out` to override the default output path.

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
- one recommended next action

### `bottleneck scorecard`

Summarizes release readiness in an evidence-backed scorecard with deterministic diagnosis, category scores, resolved thresholds, release recommendation, evidence counts, missing evidence, reasons, and recommended actions.

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

- `Proceed`: all assessed required capabilities pass
- `Conditional`: no failures, but one or more capabilities warn
- `Block`: one or more capabilities fail
- `Unknown`: scorecard evidence is unavailable or not assessed

Like `validate`, `scorecard` returns a non-zero exit code when the system is failing.

Diagnosis scoring is derived from validation output. Passing BIASED categories start high, warnings score in the middle, failures score low, and missing, placeholder, weak, stale, or disconnected evidence reduces the score further. The primary bottleneck is the weakest BIASED category using this tie priority: Assurance, Security, Behavior, Intent, Execution, Design. When all assessed BIASED categories are strong, the primary bottleneck is `None`.

When `--github-annotations` is used, `validate` and `scorecard` emit GitHub Actions workflow commands for warning and failing validation results. Failing results are emitted as `::error`; warning results are emitted as `::warning`. File paths are included when bottleneck can tie a finding to an artifact.

### `bottleneck trace`

Traces a stable evidence ID across intent, behavior, assurance, security, and execution artifacts.

Examples:

```sh
bottleneck trace INTENT-001
bottleneck trace BEHAVIOR-001 --env=production
bottleneck trace ASSURANCE-001 --format=json
```

Supported formats:

- `text`
- `json`

Unknown IDs return a non-zero exit code with a useful error.

## GitHub Actions Integration

Copyable workflow examples live in `examples/github-actions/`:

- `bottleneck-validate.yml`: runs `bottleneck validate`
- `bottleneck-scorecard.yml`: writes a Markdown scorecard to GitHub Actions Step Summary
- `bottleneck-pr-gate.yml`: runs validation, writes Step Summary output, emits annotations, and updates a stable PR comment

Copy the workflow you want into `.github/workflows/` in the repository that owns the evidence artifacts.

The PR gate uses the existing CLI exit behavior:

```sh
go build -o bottleneck .
./bottleneck validate --env=production --strict --github-annotations
./bottleneck scorecard --env=production --view=governance --format=markdown > bottleneck-scorecard.md
cat bottleneck-scorecard.md >> "$GITHUB_STEP_SUMMARY"
```

`--env=dev`, `--env=stage`, and `--env=production` select thresholds from `bottleneck/config.yaml`. `--strict` promotes placeholder and insufficient-content findings from warnings to failures, which is useful for protected PR gates. Warnings do not fail jobs unless strict mode or production traceability rules promote them to failures.

The PR comment workflow uses `actions/github-script` and the hidden marker `<!-- bottleneck-scorecard -->` so repeated runs update one comment instead of creating duplicates. It needs:

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

bottleneck does not run tests. BDD tests are executed externally by tools such as Cucumber, SpecFlow, Cucumber.js, or Behave. bottleneck validates their output against assurance rules defined by the BIASED framework.

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
  production:
    assurance:
      min_accuracy: 0.95
```

`bottleneck validate --env=<environment>` starts from `default` and overlays only the explicitly configured values for the selected environment.

### Security

Artifact: `bottleneck/security/guardrails.json`

Required schema:

```json
{
  "violations": 0
}
```

Fails when `violations > 0`.

### Execution

Artifact: `bottleneck/execution/telemetry.json`

Required schema:

```json
{
  "adoption_rate": 0.9,
  "error_rate": 0.01
}
```

Fails when `error_rate` exceeds the configured `max_error_rate`. Returns `WARNING` when `adoption_rate` is below the configured `min_adoption`.

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

The BIASED framework treats release as the beginning of continuous validation, not the end of delivery.

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
go run . explain
go run . scorecard
go run . status
```

## Documentation

Additional framework and validation references:

- `bottleneck-strategy-v1.md`
- `bottleneck/docs/validation.md`
- `bottleneck_GettingStarted.md`

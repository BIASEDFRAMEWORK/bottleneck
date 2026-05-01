# Framework Validation

## Starter Sample

`bottleneck init` creates an AI PDF Risk Summarizer sample. The starter is intentionally weak in Assurance: one evaluation fails because ambiguous financial risk language was summarized as fact. Run this first:

~~~sh
bottleneck diagnose
bottleneck validate
bottleneck scorecard
~~~

Then replace the sample evidence with your system context, starting with `bottleneck/intent/intent.md`.

## 1. Capability Schemas

### Behavior

Artifact: /bottleneck/behavior/behavior-spec.md

Required structure:

- Must not be empty
- Must contain ## Expected Behavior
- Must contain ## Unacceptable Behavior

### Intent

Artifact: /bottleneck/intent/intent.md

Required structure:

- Must contain ## Outcomes
- Must contain ## Constraints
- Must contain ## Success Criteria

### Design

Artifact: /bottleneck/design/architecture.md

Required structure:

- Must not be empty
- Must contain at least one Markdown section header

### Assurance

Artifact: /bottleneck/assurance/results.json

Required JSON structure. Developers produce only this file; bottleneck computes metrics from it:

~~~json
{
  "scenarios_total": 2,
  "scenarios_passed": 1,
  "scenarios_failed": 1,
  "failures": [
    "Ambiguous risk clause was summarized as confirmed exposure"
  ],
  "evidence": [
    {
      "id": "ASSURANCE-001",
      "refs": ["BEHAVIOR-001"],
      "source": "sample evaluation",
      "status": "fail"
    }
  ]
}
~~~

### Configuration

Artifact: /bottleneck/config.yaml

Required YAML structure:

~~~yaml
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
~~~

### Security

Artifact: /bottleneck/security/guardrails.json

Required JSON structure:

~~~json
{
  "violations": 0,
  "findings": {
    "critical": 0,
    "high": 0,
    "medium": 0,
    "low": 0,
    "note": 0,
    "unknown": 0
  },
  "evidence": [
    {
      "id": "SECURITY-001",
      "refs": ["INTENT-001", "BEHAVIOR-001"],
      "source": "sample guardrail review",
      "status": "pass"
    }
  ]
}
~~~

### Execution

Artifact: /bottleneck/execution/telemetry.json

Required JSON structure:

~~~json
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
  "source_environment": "sample",
  "cost": {
    "total": 120.5,
    "currency": "USD",
    "budget": 150,
    "trend": "stable"
  },
  "evidence": [
    {
      "id": "EXECUTION-001",
      "refs": ["BEHAVIOR-001", "ASSURANCE-001"],
      "source": "sample telemetry",
      "status": "pass"
    }
  ]
}
~~~

## 2. Validation Rules

Behavior passes only when behavior-spec.md exists, is not empty, and includes both required behavior sections.

Intent passes only when intent.md exists and includes Outcomes, Constraints, and Success Criteria sections.

Design passes only when architecture.md exists, is not empty, and includes at least one Markdown section header.

Assurance passes only when results.json exists, parses as JSON, includes all required fields, has failed scenarios at or below the configured max_failures threshold, and has calculated accuracy greater than or equal to the configured min_accuracy threshold.

config.yaml must exist and parse as valid YAML before capability validation begins. When an environment is selected, unspecified values inherit from default. Gate, SARIF, and telemetry settings are optional; missing settings use safe defaults.

Security passes when guardrails.json exists, parses as JSON, includes violations, and either violations equals 0 or SARIF findings stay within configured severity thresholds.

Execution passes when telemetry.json exists, parses as JSON, includes adoption_rate and error_rate, and error_rate is less than or equal to the configured max_error_rate threshold. Execution returns WARNING when adoption_rate is below the configured min_adoption threshold. Extended telemetry also checks generated_at freshness, deployment frequency, change failure rate, user override rate, and budget variance.

## 3. CLI Mapping

bottleneck validate loads config.yaml first, resolves inherited thresholds, and then maps each capability to a dedicated validator. Use --env to select environment thresholds:

~~~sh
bottleneck validate --env=production
bottleneck validate --env=production --strict --github-annotations
bottleneck ingest cucumber --file reports/cucumber.json
bottleneck ingest sarif --file results/codeql.sarif
bottleneck ingest telemetry --file results/telemetry.json
~~~

- Behavior -> validateBehavior()
- Intent -> validateIntent()
- Design -> validateDesign()
- Assurance -> validateAssurance()
- Security -> validateSecurity()
- Execution -> validateExecution()
- Traceability -> validateTraceability()

The CLI enforces presence checks for required artifacts, schema checks for Markdown and JSON/YAML structure, evidence-quality checks for placeholder-heavy or thin content, expected evidence IDs, measurable intent language, environment-specific threshold checks for assurance accuracy and failures, SARIF security severity thresholds, execution telemetry freshness and health, execution adoption, and explicit traceability links between evidence IDs.

Related read-only commands built on the same validation results:

- `bottleneck explain`
  Produces a human-readable explanation with owner mapping, bottleneck mapping, evidence, missing evidence, score impacts, and recommended next actions.
- `bottleneck diagnose`
  Produces a focused bottleneck diagnosis with top contributing findings, recommended next action, and confidence level.
- `bottleneck scorecard`
  Produces an evidence-backed scorecard summarizing release recommendation, effective thresholds, capability gauges, capability status, owner, bottleneck, evidence counts, missing evidence, score impacts, reasons, and recommended actions.
- `bottleneck trace`
  Shows outbound references, inbound references, evidence chains, broken references, and orphan warnings for a stable evidence ID.

GitHub Actions integration uses the same validation engine and exit codes. `--github-annotations` emits workflow annotations for validation warnings and failures. `--strict` promotes placeholder or insufficient content from warnings to failures, making it appropriate for production pull request gates.

## 4. Example Output

~~~text
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
~~~

## 5. Interpretation Commands

### Explain

~~~sh
bottleneck explain --env=production --capability=Assurance
~~~

Use `explain` when an operator needs remediation context for one or more capabilities. When no capability filter is provided, the output starts with the primary diagnosis: weakest category, why it matters, and one recommended next action.

### Scorecard

~~~sh
bottleneck scorecard --env=production
bottleneck scorecard --env=production --format=json
bottleneck scorecard --env=production --format=markdown
bottleneck scorecard --env=production --view=executive
bottleneck scorecard --env=production --view=engineering
bottleneck scorecard --env=production --view=governance
bottleneck scorecard --env=production --view=governance --format=markdown
~~~

Use `scorecard` when an operator needs release-readiness context for terminal review, GitHub summaries, release notes, governance review, or downstream automation. The scorecard includes a release recommendation of `Proceed`, `Conditional`, `Block`, or `Unknown`, deterministic diagnosis, category scores, and the effective assurance and execution thresholds resolved for the selected environment.

Diagnosis scoring is derived from validation output. `PASS` categories start high, `WARNING` categories score in the middle, and `FAIL` categories score low. Missing, placeholder-heavy, thin, vague, unmeasurable, stale, or disconnected evidence reduces the score further. Missing expected evidence IDs and broken traceability refs are included in score impacts. The primary bottleneck is the weakest BIASED category, with ties resolved in this order: Assurance, Security, Behavior, Intent, Execution, Design. If all assessed BIASED categories are strong, the primary bottleneck is `None`.

### Diagnose

~~~sh
bottleneck diagnose
bottleneck diagnose --env=production
bottleneck diagnose --format=json
bottleneck diagnose --format=markdown
bottleneck diagnose --format=github
bottleneck diagnose --env=production --strict --gate=release
~~~

Use `diagnose` when an operator needs the shortest path to the primary bottleneck. The command includes the top contributing findings, the recommended next action, and a deterministic confidence level of `High`, `Medium`, or `Low`. Confidence is based on how many evidence categories contain meaningful content and whether traceability is clean. `diagnose --format=markdown` is safe for PR comments and GitHub Step Summary output. `diagnose --format=github` emits GitHub Actions workflow annotations. `diagnose --gate=release` evaluates configured release-gate thresholds and exits non-zero only when the gate fails.

### Trace

~~~sh
bottleneck trace --id INTENT-001
bottleneck trace --id BEHAVIOR-001 --env=production
bottleneck trace --id ASSURANCE-001 --format=json
~~~

Use `trace` when an operator or reviewer needs to audit how a single evidence ID connects to intent, behavior, design, tests, security, and telemetry. Positional IDs remain supported for older scripts.

Traceability supports Markdown evidence headings and optional JSON evidence arrays:

~~~markdown
### BEHAVIOR-001: Block production release when assurance fails
Critical: true
Refs:
- INTENT-001
- ASSURANCE-001
~~~

~~~json
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
~~~

Evidence IDs must match `^(INTENT|BEHAVIOR|DESIGN|ASSURANCE|SECURITY|EXECUTION)-[0-9]{3,}$`. Duplicate IDs, invalid references, and references to missing IDs fail validation. Orphaned or unmapped evidence creates warnings by default, and behavior-to-intent or critical-behavior-to-assurance gaps fail in `--strict` mode or production.

## 6. GitHub Actions And Pull Requests

Copy workflow examples from `examples/github-actions/` into `.github/workflows/`:

- `bottleneck-validate.yml`
- `bottleneck-scorecard.yml`
- `bottleneck-pr-gate.yml`

The PR gate pattern is:

~~~sh
go build -o bottleneck .
set +e
./bottleneck validate --env=production --strict --github-annotations
VALIDATE_EXIT=$?
./bottleneck diagnose --env=production --strict --gate=release --format=markdown > bottleneck-diagnosis.md
DIAGNOSE_EXIT=$?
./bottleneck diagnose --env=production --strict --format=github
./bottleneck scorecard --env=production --view=governance --format=markdown > bottleneck-scorecard.md
SCORECARD_EXIT=$?
cat bottleneck-diagnosis.md >> "$GITHUB_STEP_SUMMARY"
echo >> "$GITHUB_STEP_SUMMARY"
cat bottleneck-scorecard.md >> "$GITHUB_STEP_SUMMARY"
if [ "$VALIDATE_EXIT" -ne 0 ]; then exit "$VALIDATE_EXIT"; fi
if [ "$DIAGNOSE_EXIT" -ne 0 ]; then exit "$DIAGNOSE_EXIT"; fi
if [ "$SCORECARD_EXIT" -ne 0 ]; then exit "$SCORECARD_EXIT"; fi
~~~

The PR comment workflow uses `actions/github-script` to update a comment containing `<!-- bottleneck-diagnosis -->`. Required permissions are `contents: read` and `pull-requests: write`.

When running in GitHub Actions, scorecard output detects `GITHUB_ACTIONS`, the event name, repository, run ID, refs, SHA, and pull request fields from the event payload. When `GITHUB_TOKEN` is present, bottleneck optionally enriches the scorecard with changed files, review approvals, pending reviewers, and failed check runs. Missing token or unavailable metadata does not fail local or CI scorecard generation.

Pull request risk signals warn on large changed file count, large additions plus deletions, draft PRs, missing requested reviewers, AI-generated or AI-assisted labels, missing approvals when review data is available, pending reviewers, failed check runs, and release-relevant source changes without matching `bottleneck/` artifact changes.

Use `--env=dev`, `--env=stage`, or `--env=production` to choose thresholds. Use `--strict` when warnings for placeholder or insufficient evidence should block the pull request. bottleneck integrates with GitHub Actions and branch protection, but it does not replace CodeQL, CI, review rules, security scanners, or deployment controls.

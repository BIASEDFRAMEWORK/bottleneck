# Framework Validation

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
  "scenarios_total": 1,
  "scenarios_passed": 1,
  "scenarios_failed": 0,
  "failures": []
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
  production:
    assurance:
      min_accuracy: 0.95
~~~

### Security

Artifact: /bottleneck/security/guardrails.json

Required JSON structure:

~~~json
{
  "violations": 0
}
~~~

### Execution

Artifact: /bottleneck/execution/telemetry.json

Required JSON structure:

~~~json
{
  "adoption_rate": 0.9,
  "error_rate": 0.01
}
~~~

## 2. Validation Rules

Behavior passes only when behavior-spec.md exists, is not empty, and includes both required behavior sections.

Intent passes only when intent.md exists and includes Outcomes, Constraints, and Success Criteria sections.

Design passes only when architecture.md exists, is not empty, and includes at least one Markdown section header.

Assurance passes only when results.json exists, parses as JSON, includes all required fields, has failed scenarios at or below the configured max_failures threshold, and has calculated accuracy greater than or equal to the configured min_accuracy threshold.

config.yaml must exist and parse as valid YAML before capability validation begins. When an environment is selected, unspecified values inherit from default.

Security passes only when guardrails.json exists, parses as JSON, includes violations, and violations equals 0.

Execution passes when telemetry.json exists, parses as JSON, includes adoption_rate and error_rate, and error_rate is less than or equal to the configured max_error_rate threshold. Execution returns WARNING when adoption_rate is below the configured min_adoption threshold.

## 3. CLI Mapping

bottleneck validate loads config.yaml first, resolves inherited thresholds, and then maps each capability to a dedicated validator. Use --env to select environment thresholds:

~~~sh
bottleneck validate --env=production
bottleneck validate --env=production --strict --github-annotations
~~~

- Behavior -> validateBehavior()
- Intent -> validateIntent()
- Design -> validateDesign()
- Assurance -> validateAssurance()
- Security -> validateSecurity()
- Execution -> validateExecution()
- Traceability -> validateTraceability()

The CLI enforces presence checks for required artifacts, schema checks for Markdown and JSON/YAML structure, environment-specific threshold checks for assurance accuracy and failures, security violations, execution error rate, execution adoption, and explicit traceability links between evidence IDs.

Related read-only commands built on the same validation results:

- `bottleneck explain`
  Produces a human-readable explanation with owner mapping, bottleneck mapping, evidence, and recommended next actions.
- `bottleneck scorecard`
  Produces an evidence-backed scorecard summarizing release recommendation, effective thresholds, capability status, owner, bottleneck, evidence counts, missing evidence, reasons, and recommended actions.
- `bottleneck trace`
  Shows outbound references, inbound references, evidence chains, broken references, and orphan warnings for a stable evidence ID.

GitHub Actions integration uses the same validation engine and exit codes. `--github-annotations` emits workflow annotations for validation warnings and failures. `--strict` promotes placeholder or insufficient content from warnings to failures, making it appropriate for production pull request gates.

## 4. Example Output

~~~text
Behavior: PASS
Intent: PASS
Design: PASS
Assurance: FAIL (accuracy below threshold)
  accuracy: 0.90 (threshold: 0.95)
  scenarios_failed: 0 (allowed: 0)
Security: PASS
Execution: PASS

System Status: FAIL
Primary Bottleneck: Assurance
Environment: production
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

Diagnosis scoring is derived from validation output. `PASS` categories start high, `WARNING` categories score in the middle, and `FAIL` categories score low. Missing, placeholder, weak, stale, or disconnected evidence reduces the score further. The primary bottleneck is the weakest BIASED category, with ties resolved in this order: Assurance, Security, Behavior, Intent, Execution, Design. If all assessed BIASED categories are strong, the primary bottleneck is `None`.

### Trace

~~~sh
bottleneck trace INTENT-001
bottleneck trace BEHAVIOR-001 --env=production
bottleneck trace ASSURANCE-001 --format=json
~~~

Use `trace` when an operator or reviewer needs to audit how a single evidence ID connects to intent, behavior, tests, security, and telemetry.

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
./bottleneck validate --env=production --strict --github-annotations
./bottleneck scorecard --env=production --view=governance --format=markdown > bottleneck-scorecard.md
cat bottleneck-scorecard.md >> "$GITHUB_STEP_SUMMARY"
~~~

The PR comment workflow uses `actions/github-script` to update a comment containing `<!-- bottleneck-scorecard -->`. Required permissions are `contents: read` and `pull-requests: write`.

When running in GitHub Actions, scorecard output detects `GITHUB_ACTIONS`, the event name, repository, run ID, refs, SHA, and pull request fields from the event payload. When `GITHUB_TOKEN` is present, bottleneck optionally enriches the scorecard with changed files, review approvals, pending reviewers, and failed check runs. Missing token or unavailable metadata does not fail local or CI scorecard generation.

Pull request risk signals warn on large changed file count, large additions plus deletions, draft PRs, missing requested reviewers, AI-generated or AI-assisted labels, missing approvals when review data is available, pending reviewers, failed check runs, and release-relevant source changes without matching `bottleneck/` artifact changes.

Use `--env=dev`, `--env=stage`, or `--env=production` to choose thresholds. Use `--strict` when warnings for placeholder or insufficient evidence should block the pull request. bottleneck integrates with GitHub Actions and branch protection, but it does not replace CodeQL, CI, review rules, security scanners, or deployment controls.

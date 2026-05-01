package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initDirectories = []string{
	"bottleneck",
	"bottleneck/behavior",
	"bottleneck/intent",
	"bottleneck/design",
	"bottleneck/assurance",
	"bottleneck/assurance/features",
	"bottleneck/security",
	"bottleneck/execution",
	"bottleneck/docs",
}

var initFiles = map[string]string{
	"bottleneck/config.yaml":                       defaultConfigYAML,
	"bottleneck/behavior/behavior-spec.md":         defaultBehaviorSpec,
	"bottleneck/intent/intent.md":                  defaultIntent,
	"bottleneck/design/architecture.md":            defaultArchitecture,
	"bottleneck/assurance/features/sample.feature": defaultAssuranceFeature,
	"bottleneck/assurance/results.json":            defaultAssuranceResults,
	"bottleneck/security/guardrails.json":          defaultSecurityGuardrails,
	"bottleneck/execution/telemetry.json":          defaultExecutionTelemetry,
	"bottleneck/docs/validation.md":                validationDocumentation,
}

const defaultConfigYAML = `environments:
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

  dev:
    assurance:
      min_accuracy: 0.85

  test:
    assurance:
      min_accuracy: 0.90

  stage:
    assurance:
      min_accuracy: 0.93

  production:
    assurance:
      min_accuracy: 0.95
    gate:
      release:
        min_primary_score: 75
        require_traceability: true
        require_governance: true
`

const defaultIntent = `# Intent

<!-- Sample app: AI PDF Risk Summarizer. -->

## Outcomes

<!-- Replace this sample outcome with the release outcome your system must prove. -->

### INTENT-001: Summarize financial PDF risk without hiding uncertainty
Refs:
- BEHAVIOR-001

The system must summarize material risk clauses from financial PDFs while preserving uncertainty and caveats that affect release or investment decisions.

## Constraints

<!-- Replace these sample constraints with the boundaries your system must honor. -->

- The system must not invent risk facts that are not present in the source PDF.
- The system must flag ambiguous or qualified risk language instead of rewriting it as certainty.

## Success Criteria

<!-- Replace these sample criteria with measurable release checks for your system. -->

- At least 95% of evaluated summaries preserve material risk caveats.
- 100% of ambiguous risk clauses in the evaluation set are flagged as uncertain.
`

const defaultBehaviorSpec = `# Behavior Specification

## Expected Behavior

<!-- Replace this sample behavior with the behavior your release must prove. -->

### BEHAVIOR-001: Flag ambiguous financial risk language
Critical: true
Refs:
- INTENT-001
- ASSURANCE-001

When a PDF contains qualified risk language such as "may", "could", "subject to", or "material uncertainty", the summary must preserve that uncertainty and flag it for review.

## Unacceptable Behavior

<!-- Replace these examples with failures your system must prevent. -->

- The system must not summarize ambiguous risk language as a confirmed fact.
- The system must not omit material caveats from the risk summary.
`

const defaultArchitecture = `# Architecture

<!-- Replace this sample design with the architecture evidence reviewers need. -->

### DESIGN-001: Local PDF risk summarization flow
Refs:
- INTENT-001
- BEHAVIOR-001

The workflow extracts PDF text, identifies candidate risk clauses, asks the summarizer to produce a short risk summary, and runs a post-summary uncertainty check before showing the output to a reviewer.

Key components:
- PDF text extraction
- Risk clause detector
- Summary generator
- Uncertainty flagging check
- Reviewer-facing output
`

const defaultAssuranceFeature = `Feature: AI PDF risk summarization

  @BEHAVIOR-001
  Scenario: Ambiguous risk clause is flagged
    Given a financial PDF says exposure "may be material subject to market conditions"
    When the system summarizes the risk clause
    Then the summary should flag the exposure as uncertain

  @BEHAVIOR-001
  Scenario: Ambiguous risk clause is not stated as fact
    Given a financial PDF uses qualified risk language
    When the system produces a risk summary
    Then the summary should not state the risk as confirmed exposure
`

const defaultAssuranceResults = `{
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
      "status": "fail",
      "summary": "One evaluation failed because ambiguous risk language was summarized as fact."
    }
  ]
}
`

const defaultSecurityGuardrails = `{
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
      "status": "pass",
      "summary": "Guardrails require source-grounded summaries and prohibit unsupported risk claims."
    }
  ]
}
`

const defaultExecutionTelemetry = `{
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
      "status": "pass",
      "summary": "Sample pilot telemetry shows reviewers used the summary flow with low processing errors."
    }
  ]
}
`

const initSuccessMessage = `Bottleneck initialized.

This starter uses the AI PDF Risk Summarizer sample and intentionally leaves Assurance weak so diagnose can show a real bottleneck.

Next:
1. Run: bottleneck diagnose
2. Review the primary bottleneck
3. Run: bottleneck validate
4. Run: bottleneck scorecard
5. Replace sample intent and behavior with evidence from your own system

Start with: bottleneck/intent/intent.md
`

const validationDocumentation = `# Framework Validation

## Starter Sample

bottleneck init creates an AI PDF Risk Summarizer sample. The starter is intentionally weak in Assurance: one evaluation fails because ambiguous financial risk language was summarized as fact. Run this first:

~~~sh
bottleneck diagnose
bottleneck validate
bottleneck scorecard
~~~

Then replace the sample evidence with your system context, starting with bottleneck/intent/intent.md.

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

- bottleneck explain
  Produces a human-readable explanation with owner mapping, bottleneck mapping, evidence, missing evidence, score impacts, and recommended next actions.
- bottleneck diagnose
  Produces a focused bottleneck diagnosis with top contributing findings, recommended next action, and confidence level.
- bottleneck scorecard
  Produces an evidence-backed scorecard summarizing release recommendation, effective thresholds, capability gauges, capability status, owner, bottleneck, evidence counts, missing evidence, score impacts, reasons, and recommended actions.
- bottleneck trace
  Shows outbound references, inbound references, evidence chains, broken references, and orphan warnings for a stable evidence ID.

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

Use explain when an operator needs remediation context for one or more capabilities.

### Scorecard

~~~sh
bottleneck scorecard --env=production
bottleneck scorecard --env=production --format=json
bottleneck scorecard --env=production --format=markdown
bottleneck scorecard --env=production --view=executive
bottleneck scorecard --env=production --view=engineering
bottleneck scorecard --env=production --view=governance
~~~

Use scorecard when an operator needs release-readiness context for terminal review, GitHub summaries, release notes, governance review, or downstream automation. The scorecard includes a release recommendation of Proceed, Conditional, Block, or Unknown, and displays the effective assurance and execution thresholds resolved for the selected environment.

### Diagnose

~~~sh
bottleneck diagnose
bottleneck diagnose --env=production
bottleneck diagnose --format=json
bottleneck diagnose --format=markdown
bottleneck diagnose --format=github
bottleneck diagnose --env=production --strict --gate=release
~~~

Use diagnose when an operator needs the shortest path to the primary bottleneck. The command includes the top contributing findings, the recommended next action, and a deterministic confidence level of High, Medium, or Low. Confidence is based on how many evidence categories contain meaningful content and whether traceability is clean. diagnose --format=markdown is safe for PR comments and GitHub Step Summary output. diagnose --format=github emits GitHub Actions workflow annotations. diagnose --gate=release evaluates configured release-gate thresholds and exits non-zero only when the gate fails.

### Trace

~~~sh
bottleneck trace --id INTENT-001
bottleneck trace --id BEHAVIOR-001 --env=production
bottleneck trace --id ASSURANCE-001 --format=json
~~~

Use trace when an operator or reviewer needs to audit how a single evidence ID connects to intent, behavior, design, tests, security, and telemetry. Positional IDs remain supported for older scripts.

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

Evidence IDs must match ^(INTENT|BEHAVIOR|DESIGN|ASSURANCE|SECURITY|EXECUTION)-[0-9]{3,}$. Duplicate IDs, invalid references, and references to missing IDs fail validation. Orphaned or unmapped evidence creates warnings by default, and behavior-to-intent or critical-behavior-to-assurance gaps fail in --strict mode or production.
`

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize local evidence artifacts",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initializeProject("."); err != nil {
			return err
		}

		fmt.Print(initSuccessMessage)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func initializeProject(basePath string) error {
	for _, dir := range initDirectories {
		if err := os.MkdirAll(filepath.Join(basePath, dir), 0o755); err != nil {
			return err
		}
	}

	for relativePath, content := range initFiles {
		if err := writeFileIfMissing(filepath.Join(basePath, relativePath), content); err != nil {
			return err
		}
	}

	return nil
}

func writeFileIfMissing(path string, content string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}

	return os.WriteFile(path, []byte(content), 0o644)
}

package validator

import (
	"strings"
	"testing"

	"bottleneck/internal/models"
)

func TestSecuritySARIFThresholdsPassForAllowedFindings(t *testing.T) {
	basePath := t.TempDir()
	writeValidationProject(t, basePath, map[string]string{
		"bottleneck/security/guardrails.json": `{
  "violations": 2,
  "findings": {
    "critical": 0,
    "high": 0,
    "medium": 2,
    "low": 0,
    "note": 0,
    "unknown": 0
  },
  "evidence": [
    {"id": "SECURITY-001", "refs": ["BEHAVIOR-001"], "source": "codeql.sarif", "status": "fail", "severity": "medium"}
  ]
}`,
	})

	security := resultForCapability(t, NewEngine(basePath, "default").Validate(), "Security")
	if security.Status != models.StatusPass {
		t.Fatalf("expected allowed SARIF findings to pass, got %q with details %#v", security.Status, security.Details)
	}
}

func TestSecuritySARIFThresholdsFailForHighFindings(t *testing.T) {
	basePath := t.TempDir()
	writeValidationProject(t, basePath, map[string]string{
		"bottleneck/security/guardrails.json": `{
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
    {"id": "SECURITY-001", "refs": ["BEHAVIOR-001"], "source": "codeql.sarif", "status": "fail", "severity": "high"}
  ]
}`,
	})

	security := resultForCapability(t, NewEngine(basePath, "default").Validate(), "Security")
	if security.Status != models.StatusFail {
		t.Fatalf("expected high SARIF finding to fail, got %q", security.Status)
	}
	if !containsDetailSubstring(security.Details, "high SARIF findings above threshold") {
		t.Fatalf("expected high finding threshold detail, got %#v", security.Details)
	}
}

func TestSecuritySARIFUnknownSeverityWarnsUnlessConfiguredToFail(t *testing.T) {
	basePath := t.TempDir()
	writeValidationProject(t, basePath, map[string]string{
		"bottleneck/config.yaml": `environments:
  default:
    assurance:
      min_accuracy: 0.90
      max_failures: 0
    execution:
      max_error_rate: 0.05
      min_adoption: 0.5
    security:
      sarif:
        fail_on_unknown_severity: true
`,
		"bottleneck/security/guardrails.json": `{
  "violations": 1,
  "findings": {
    "critical": 0,
    "high": 0,
    "medium": 0,
    "low": 0,
    "note": 0,
    "unknown": 1
  },
  "evidence": [
    {"id": "SECURITY-001", "refs": ["BEHAVIOR-001"], "source": "codeql.sarif", "status": "fail", "severity": "unknown"}
  ]
}`,
	})

	security := resultForCapability(t, NewEngine(basePath, "default").Validate(), "Security")
	if security.Status != models.StatusFail {
		t.Fatalf("expected unknown severity to fail when configured, got %q", security.Status)
	}
	if !containsDetailSubstring(security.Details, "unknown SARIF severity above threshold") {
		t.Fatalf("expected unknown severity detail, got %#v", security.Details)
	}
}

func TestExecutionExtendedTelemetryHealthAndStaleness(t *testing.T) {
	basePath := t.TempDir()
	writeValidationProject(t, basePath, map[string]string{
		"bottleneck/execution/telemetry.json": healthyExtendedTelemetry("2999-04-30T12:00:00Z"),
	})

	execution := resultForCapability(t, NewEngine(basePath, "default").Validate(), "Execution")
	if execution.Status != models.StatusPass {
		t.Fatalf("expected healthy extended telemetry to pass, got %q with details %#v", execution.Status, execution.Details)
	}

	writeValidationProject(t, basePath, map[string]string{
		"bottleneck/execution/telemetry.json": healthyExtendedTelemetry("2000-01-01T00:00:00Z"),
	})
	execution = resultForCapability(t, NewEngine(basePath, "default").Validate(), "Execution")
	if execution.Status != models.StatusWarning {
		t.Fatalf("expected stale telemetry warning, got %q with details %#v", execution.Status, execution.Details)
	}
	if !containsDetailSubstring(execution.Details, "telemetry is stale") {
		t.Fatalf("expected stale telemetry detail, got %#v", execution.Details)
	}
}

func TestExecutionTelemetryFreshnessThresholdCanBeConfigured(t *testing.T) {
	basePath := t.TempDir()
	writeValidationProject(t, basePath, map[string]string{
		"bottleneck/config.yaml": `environments:
  default:
    assurance:
      min_accuracy: 0.90
      max_failures: 0
    execution:
      telemetry:
        max_age_hours: 300000
        min_deployments_per_week: 1
        max_change_failure_rate: 0.15
        max_error_rate: 0.05
        max_user_override_rate: 0.10
        min_adoption_rate: 0.50
        max_budget_variance: 0.20
`,
		"bottleneck/execution/telemetry.json": healthyExtendedTelemetry("2000-01-01T00:00:00Z"),
	})

	execution := resultForCapability(t, NewEngine(basePath, "default").Validate(), "Execution")
	if execution.Status != models.StatusPass {
		t.Fatalf("expected configured freshness threshold to pass old telemetry, got %q with details %#v", execution.Status, execution.Details)
	}
}

func TestExecutionPoorTelemetryFailsAndPartialTelemetryWarns(t *testing.T) {
	basePath := t.TempDir()
	writeValidationProject(t, basePath, map[string]string{
		"bottleneck/execution/telemetry.json": `{
  "generated_at": "2999-04-30T12:00:00Z",
  "adoption_rate": 0.72,
  "error_rate": 0.02,
  "evidence": [{"id": "EXECUTION-001", "refs": ["BEHAVIOR-001"], "source": "telemetry", "status": "pass"}]
}`,
	})

	execution := resultForCapability(t, NewEngine(basePath, "default").Validate(), "Execution")
	if execution.Status != models.StatusWarning {
		t.Fatalf("expected partial telemetry warning, got %q with details %#v", execution.Status, execution.Details)
	}
	if !containsDetailSubstring(execution.Details, "change_failure_rate telemetry missing") {
		t.Fatalf("expected missing change failure detail, got %#v", execution.Details)
	}

	writeValidationProject(t, basePath, map[string]string{
		"bottleneck/execution/telemetry.json": `{
  "generated_at": "2999-04-30T12:00:00Z",
  "deployment_frequency": {"deployments": 7, "period_days": 7},
  "change_failure_rate": 0.50,
  "adoption_rate": 0.72,
  "error_rate": 0.10,
  "user_override_rate": 0.03,
  "cost": {"total": 120, "budget": 150},
  "evidence": [{"id": "EXECUTION-001", "refs": ["BEHAVIOR-001"], "source": "telemetry", "status": "pass"}]
}`,
	})
	execution = resultForCapability(t, NewEngine(basePath, "default").Validate(), "Execution")
	if execution.Status != models.StatusFail {
		t.Fatalf("expected poor telemetry to fail, got %q with details %#v", execution.Status, execution.Details)
	}
	if !containsDetailSubstring(execution.Details, "error_rate above threshold") ||
		!containsDetailSubstring(execution.Details, "change_failure_rate above threshold") {
		t.Fatalf("expected poor telemetry details, got %#v", execution.Details)
	}
}

func TestExecutionTelemetryPercentagesAreScoredAsRatios(t *testing.T) {
	basePath := t.TempDir()
	writeValidationProject(t, basePath, map[string]string{
		"bottleneck/execution/telemetry.json": `{
  "generated_at": "2999-04-30T12:00:00Z",
  "deployment_frequency": {"deployments": 7, "period_days": 7},
  "change_failure_rate": 5,
  "adoption_rate": 72,
  "error_rate": 2,
  "user_override_rate": 3,
  "cost": {"budget_variance": 20},
  "evidence": [{"id": "EXECUTION-001", "refs": ["BEHAVIOR-001"], "source": "telemetry", "status": "pass"}]
}`,
	})

	execution := resultForCapability(t, NewEngine(basePath, "default").Validate(), "Execution")
	if execution.Status != models.StatusPass {
		t.Fatalf("expected percentage telemetry to pass as ratios, got %q with details %#v", execution.Status, execution.Details)
	}
}

func TestExecutionCostBudgetVarianceWarns(t *testing.T) {
	basePath := t.TempDir()
	writeValidationProject(t, basePath, map[string]string{
		"bottleneck/execution/telemetry.json": `{
  "generated_at": "2999-04-30T12:00:00Z",
  "deployment_frequency": {"deployments": 7, "period_days": 7},
  "change_failure_rate": 0.05,
  "adoption_rate": 0.72,
  "error_rate": 0.02,
  "user_override_rate": 0.03,
  "cost": {"total": 200, "budget": 100},
  "evidence": [{"id": "EXECUTION-001", "refs": ["BEHAVIOR-001"], "source": "telemetry", "status": "pass"}]
}`,
	})

	execution := resultForCapability(t, NewEngine(basePath, "default").Validate(), "Execution")
	if execution.Status != models.StatusWarning {
		t.Fatalf("expected cost variance warning, got %q with details %#v", execution.Status, execution.Details)
	}
	if !containsDetailSubstring(execution.Details, "cost budget variance above threshold") {
		t.Fatalf("expected cost variance detail, got %#v", execution.Details)
	}
}

func healthyExtendedTelemetry(generatedAt string) string {
	return `{
  "generated_at": "` + generatedAt + `",
  "window": {"start": "2026-04-23T00:00:00Z", "end": "2026-04-30T00:00:00Z"},
  "deployment_frequency": {"deployments": 7, "period_days": 7},
  "change_failure_rate": 0.05,
  "adoption_rate": 0.72,
  "error_rate": 0.02,
  "user_override_rate": 0.03,
  "cost": {"total": 120.5, "currency": "USD", "budget": 150, "trend": "stable"},
  "evidence": [{"id": "EXECUTION-001", "refs": ["BEHAVIOR-001"], "source": "telemetry", "status": "pass"}]
}`
}

func containsDetailSubstring(details []string, expected string) bool {
	for _, detail := range details {
		if strings.Contains(detail, expected) {
			return true
		}
	}
	return false
}

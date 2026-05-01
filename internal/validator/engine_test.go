package validator

import (
	"testing"
)

func TestEngineResultIncludesResolvedEffectiveThresholds(t *testing.T) {
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
      telemetry:
        max_age_hours: 96
        min_deployments_per_week: 2
        max_change_failure_rate: 0.12
        max_user_override_rate: 0.08
        max_budget_variance: 0.15
    security:
      sarif:
        max_medium: 10
        max_low: 30
  production:
    assurance:
      min_accuracy: 0.95
    execution:
      max_error_rate: 0.02
      telemetry:
        max_age_hours: 24
    security:
      sarif:
        max_high: 0
        fail_on_unknown_severity: true
`,
	})

	result := NewEngine(basePath, "production").Validate()
	thresholds := result.EffectiveThresholds

	if thresholds.Assurance.MinAccuracy != 0.95 {
		t.Fatalf("expected inherited production min accuracy 0.95, got %.2f", thresholds.Assurance.MinAccuracy)
	}
	if thresholds.Assurance.MaxFailures != 0 {
		t.Fatalf("expected inherited max failures 0, got %d", thresholds.Assurance.MaxFailures)
	}
	if thresholds.Execution.MaxErrorRate != 0.02 {
		t.Fatalf("expected production max error rate 0.02, got %.2f", thresholds.Execution.MaxErrorRate)
	}
	if thresholds.Execution.MinAdoption != 0.5 {
		t.Fatalf("expected inherited min adoption 0.50, got %.2f", thresholds.Execution.MinAdoption)
	}
	if thresholds.Execution.Telemetry.MaxAgeHours != 24 {
		t.Fatalf("expected production telemetry max age override 24, got %d", thresholds.Execution.Telemetry.MaxAgeHours)
	}
	if thresholds.Execution.Telemetry.MinDeploymentsPerWeek != 2 {
		t.Fatalf("expected inherited telemetry deployment threshold 2, got %.2f", thresholds.Execution.Telemetry.MinDeploymentsPerWeek)
	}
	if thresholds.Execution.Telemetry.MaxChangeFailureRate != 0.12 {
		t.Fatalf("expected inherited change failure threshold 0.12, got %.2f", thresholds.Execution.Telemetry.MaxChangeFailureRate)
	}
	if thresholds.Execution.Telemetry.MaxUserOverrideRate != 0.08 {
		t.Fatalf("expected inherited user override threshold 0.08, got %.2f", thresholds.Execution.Telemetry.MaxUserOverrideRate)
	}
	if thresholds.Execution.Telemetry.MaxBudgetVariance != 0.15 {
		t.Fatalf("expected inherited budget variance threshold 0.15, got %.2f", thresholds.Execution.Telemetry.MaxBudgetVariance)
	}
	if thresholds.Security.SARIF.MaxHigh != 0 {
		t.Fatalf("expected production max high threshold 0, got %d", thresholds.Security.SARIF.MaxHigh)
	}
	if thresholds.Security.SARIF.MaxMedium != 10 {
		t.Fatalf("expected inherited max medium threshold 10, got %d", thresholds.Security.SARIF.MaxMedium)
	}
	if thresholds.Security.SARIF.MaxLow != 30 {
		t.Fatalf("expected inherited max low threshold 30, got %d", thresholds.Security.SARIF.MaxLow)
	}
	if !thresholds.Security.SARIF.FailOnUnknownSeverity {
		t.Fatal("expected production unknown severity threshold policy")
	}
}

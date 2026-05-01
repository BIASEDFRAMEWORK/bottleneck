package validator

import (
	"strings"
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
    gate:
      release:
        min_primary_score: 85
        require_traceability: true
        require_governance: true
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
	if thresholds.Execution.Telemetry.StaleAllowed {
		t.Fatal("expected production telemetry staleness to be disallowed")
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
	if thresholds.Gate.Release.MinPrimaryScore != 85 {
		t.Fatalf("expected production release gate minimum score 85, got %d", thresholds.Gate.Release.MinPrimaryScore)
	}
	if !thresholds.Gate.Release.RequireTraceability {
		t.Fatal("expected production release gate to require traceability")
	}
	if !thresholds.Gate.Release.RequireGovernance {
		t.Fatal("expected production release gate to require governance")
	}
}

func TestEngineRejectsUnknownEnvironment(t *testing.T) {
	basePath := t.TempDir()
	writeValidationProject(t, basePath, map[string]string{
		"bottleneck/config.yaml": `environments:
  default:
    assurance:
      min_accuracy: 0.90
      max_failures: 0
  dev:
    assurance:
      min_accuracy: 0.85
`,
	})

	result := NewEngine(basePath, "not-real").Validate()
	if result.SystemStatus != "FAIL" || result.PrimaryBottleneck != "Config" {
		t.Fatalf("expected config failure for unknown environment, got %#v", result)
	}
	if len(result.Results) != 1 || result.Results[0].Capability != "Config" {
		t.Fatalf("expected config validation result, got %#v", result.Results)
	}
	if result.Results[0].Message != `unknown environment "not-real" (supported: default, dev)` {
		t.Fatalf("unexpected unknown environment message: %q", result.Results[0].Message)
	}
	for _, substring := range []string{
		"Next action: choose one of: dev.",
		"bottleneck scorecard --env=production",
	} {
		if !containsDetail(result.Results[0].Details, substring) {
			t.Fatalf("expected unknown environment guidance %q, got %#v", substring, result.Results[0].Details)
		}
	}
}

func TestEngineMissingConfigIncludesInitializationGuidance(t *testing.T) {
	result := NewEngine(t.TempDir(), "default").Validate()
	if result.SystemStatus != "FAIL" || result.PrimaryBottleneck != "Config" {
		t.Fatalf("expected config failure, got %#v", result)
	}
	check := resultForCapability(t, result, "Config")
	if check.Message != "No Bottleneck config found." {
		t.Fatalf("expected missing config reason, got %q", check.Message)
	}
	for _, substring := range []string{
		"Bottleneck has not been initialized",
		"Next action: initialize a SaaS starter project:",
		"bottleneck init --template saas",
	} {
		if !containsDetail(check.Details, substring) {
			t.Fatalf("expected missing config guidance %q, got %#v", substring, check.Details)
		}
	}
}

func containsDetail(details []string, expected string) bool {
	for _, detail := range details {
		if strings.Contains(detail, expected) {
			return true
		}
	}
	return false
}

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadExistingConfigWithoutGateSettings(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(`environments:
  default:
    assurance:
      min_accuracy: 0.90
      max_failures: 0
    execution:
      max_error_rate: 0.05
      min_adoption: 0.50
  production:
    assurance:
      min_accuracy: 0.95
`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	resolved := ResolveEnvironment(cfg, "production")
	if resolved.Gate.Release.MinPrimaryScore != 60 {
		t.Fatalf("expected safe default gate threshold, got %d", resolved.Gate.Release.MinPrimaryScore)
	}
	if !resolved.Gate.Release.RequireTraceability {
		t.Fatal("expected safe default traceability requirement")
	}
	if resolved.Security.SARIF.MaxHigh != 0 || resolved.Security.SARIF.MaxMedium != 5 {
		t.Fatalf("expected safe SARIF defaults, got %#v", resolved.Security.SARIF)
	}
	if resolved.Execution.Telemetry.MaxAgeHours != 168 || resolved.Execution.Telemetry.MinDeploymentsPerWeek != 1 {
		t.Fatalf("expected safe telemetry defaults, got %#v", resolved.Execution.Telemetry)
	}
}

func TestResolveEnvironmentUsesDefaultReleaseGateWhenMissing(t *testing.T) {
	cfg := Config{Environments: map[string]EnvironmentConfig{
		"default": {
			Assurance: AssuranceConfig{MinAccuracy: 0.90, MaxFailures: 0},
			Execution: ExecutionConfig{MaxErrorRate: 0.05, MinAdoption: 0.50},
		},
	}}

	resolved := ResolveEnvironment(cfg, "default")

	if resolved.Gate.Release.MinPrimaryScore != 60 {
		t.Fatalf("expected default min primary score 60, got %d", resolved.Gate.Release.MinPrimaryScore)
	}
	if !resolved.Gate.Release.RequireTraceability {
		t.Fatal("expected traceability to be required by default")
	}
	if resolved.Gate.Release.RequireGovernance {
		t.Fatal("expected governance not required by default")
	}
	expected := []string{"Intent", "Behavior", "Assurance", "Security", "Execution"}
	if !sameStrings(resolved.Gate.Release.RequiredCategories, expected) {
		t.Fatalf("expected required categories %#v, got %#v", expected, resolved.Gate.Release.RequiredCategories)
	}
}

func TestResolveEnvironmentInheritsReleaseGateSettings(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(`environments:
  default:
    gate:
      release:
        min_primary_score: 65
        required_categories:
          - Intent
          - Behavior
          - Assurance
        require_traceability: true
        require_governance: false
  production:
    gate:
      release:
        min_primary_score: 75
        require_governance: true
`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	resolved := ResolveEnvironment(cfg, "production")

	if resolved.Gate.Release.MinPrimaryScore != 75 {
		t.Fatalf("expected production min primary score override 75, got %d", resolved.Gate.Release.MinPrimaryScore)
	}
	if !sameStrings(resolved.Gate.Release.RequiredCategories, []string{"Intent", "Behavior", "Assurance"}) {
		t.Fatalf("expected inherited required categories, got %#v", resolved.Gate.Release.RequiredCategories)
	}
	if !resolved.Gate.Release.RequireTraceability {
		t.Fatal("expected inherited traceability requirement")
	}
	if !resolved.Gate.Release.RequireGovernance {
		t.Fatal("expected production governance requirement")
	}
}

func TestResolveEnvironmentInheritsSecurityAndTelemetrySettings(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(`environments:
  default:
    execution:
      max_error_rate: 0.05
      min_adoption: 0.50
      telemetry:
        max_age_hours: 96
        min_deployments_per_week: 2
        max_change_failure_rate: 0.12
    security:
      sarif:
        max_high: 1
        max_medium: 10
        fail_on_unknown_severity: false
  production:
    execution:
      telemetry:
        max_age_hours: 24
    security:
      sarif:
        max_high: 0
        fail_on_unknown_severity: true
`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	resolved := ResolveEnvironment(cfg, "production")
	if resolved.Execution.Telemetry.MaxAgeHours != 24 {
		t.Fatalf("expected production telemetry age override, got %d", resolved.Execution.Telemetry.MaxAgeHours)
	}
	if resolved.Execution.Telemetry.MinDeploymentsPerWeek != 2 {
		t.Fatalf("expected inherited deployment threshold, got %.2f", resolved.Execution.Telemetry.MinDeploymentsPerWeek)
	}
	if resolved.Execution.Telemetry.MaxChangeFailureRate != 0.12 {
		t.Fatalf("expected inherited change failure threshold, got %.2f", resolved.Execution.Telemetry.MaxChangeFailureRate)
	}
	if resolved.Security.SARIF.MaxHigh != 0 {
		t.Fatalf("expected production max high override, got %d", resolved.Security.SARIF.MaxHigh)
	}
	if resolved.Security.SARIF.MaxMedium != 10 {
		t.Fatalf("expected inherited medium threshold, got %d", resolved.Security.SARIF.MaxMedium)
	}
	if !resolved.Security.SARIF.FailOnUnknownSeverity {
		t.Fatal("expected production unknown severity policy")
	}
}

func TestResolveEnvironmentLegacyExecutionOverrideFeedsTelemetry(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(`environments:
  default:
    execution:
      max_error_rate: 0.05
      min_adoption: 0.50
      telemetry:
        max_error_rate: 0.04
        min_adoption_rate: 0.60
  production:
    execution:
      max_error_rate: 0.02
      min_adoption: 0.80
`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	resolved := ResolveEnvironment(cfg, "production")
	if resolved.Execution.Telemetry.MaxErrorRate != 0.02 {
		t.Fatalf("expected production legacy max_error_rate to override telemetry, got %.2f", resolved.Execution.Telemetry.MaxErrorRate)
	}
	if resolved.Execution.Telemetry.MinAdoptionRate != 0.80 {
		t.Fatalf("expected production legacy min_adoption to override telemetry, got %.2f", resolved.Execution.Telemetry.MinAdoptionRate)
	}
}

func sameStrings(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

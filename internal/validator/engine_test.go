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
  production:
    assurance:
      min_accuracy: 0.95
    execution:
      max_error_rate: 0.02
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
}

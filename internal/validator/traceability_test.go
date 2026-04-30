package validator

import (
	"testing"

	"bottleneck/internal/models"
)

func TestTraceabilityValidationResultWarnsAndFailsByMode(t *testing.T) {
	basePath := t.TempDir()
	writeValidationProject(t, basePath, map[string]string{
		"bottleneck/behavior/behavior-spec.md": "# Behavior Specification\n\n## Expected Behavior\n\n### BEHAVIOR-001: Block release\nRefs:\n- ASSURANCE-001\n\nThe system blocks unsafe releases when assurance fails.\n\n## Unacceptable Behavior\n\nThe system must not ignore failed assurance results.\n",
		"bottleneck/assurance/results.json":    "{\n  \"scenarios_total\": 1,\n  \"scenarios_passed\": 1,\n  \"scenarios_failed\": 0,\n  \"failures\": [],\n  \"evidence\": [{\"id\":\"ASSURANCE-001\",\"refs\":[\"BEHAVIOR-001\"],\"source\":\"cucumber\",\"status\":\"pass\"}]\n}\n",
	})

	defaultResult := NewEngine(basePath, "default").Validate()
	defaultTrace := resultForCapability(t, defaultResult, "Traceability")
	if defaultTrace.Status != models.StatusWarning {
		t.Fatalf("expected default traceability WARNING, got %q with details %#v", defaultTrace.Status, defaultTrace.Details)
	}

	strictResult := NewEngine(basePath, "default", WithStrictMode(true)).Validate()
	strictTrace := resultForCapability(t, strictResult, "Traceability")
	if strictTrace.Status != models.StatusFail {
		t.Fatalf("expected strict traceability FAIL, got %q with details %#v", strictTrace.Status, strictTrace.Details)
	}

	productionResult := NewEngine(basePath, "production").Validate()
	productionTrace := resultForCapability(t, productionResult, "Traceability")
	if productionTrace.Status != models.StatusFail {
		t.Fatalf("expected production traceability FAIL, got %q with details %#v", productionTrace.Status, productionTrace.Details)
	}
}

func TestSimpleJSONArtifactsWithoutEvidenceStillValidate(t *testing.T) {
	basePath := t.TempDir()
	writeValidationProject(t, basePath, nil)

	result := NewEngine(basePath, "default").Validate()

	for _, capability := range []string{"Assurance", "Security", "Execution"} {
		check := resultForCapability(t, result, capability)
		if check.Status != models.StatusPass {
			t.Fatalf("expected %s PASS without optional evidence arrays, got %q with details %#v", capability, check.Status, check.Details)
		}
	}

	traceability := resultForCapability(t, result, "Traceability")
	if traceability.Status != models.StatusPass {
		t.Fatalf("expected Traceability PASS for artifacts without IDs, got %q with details %#v", traceability.Status, traceability.Details)
	}
}

package validator

import (
	"testing"

	"bottleneck/internal/models"
)

func TestTraceabilityValidationResultWarnsAndFailsByMode(t *testing.T) {
	basePath := t.TempDir()
	writeValidationProject(t, basePath, map[string]string{
		"bottleneck/behavior/behavior-spec.md": "# Behavior Specification\n\n## Expected Behavior\n\n### BEHAVIOR-001: Block release\nRefs:\n- ASSURANCE-001\n\nThe system blocks unsafe releases when assurance fails.\n\n## Unacceptable Behavior\n\nThe system must not ignore failed assurance results.\n",
		"bottleneck/intent/intent.md":          "# Intent\n\n## Outcomes\n\nThe CLI identifies blocking capability failures before release approval.\n\n## Constraints\n\nValidation remains deterministic offline.\n\n## Success Criteria\n\n- At least 95% of warnings include an artifact path.\n",
		"bottleneck/design/architecture.md":    "# Architecture\n\nThe CLI reads local artifacts, validates them deterministically, and renders release posture.\n",
		"bottleneck/assurance/results.json":    "{\n  \"scenarios_total\": 1,\n  \"scenarios_passed\": 1,\n  \"scenarios_failed\": 0,\n  \"failures\": [],\n  \"evidence\": [{\"id\":\"ASSURANCE-001\",\"refs\":[\"BEHAVIOR-001\"],\"source\":\"cucumber\",\"status\":\"pass\"}]\n}\n",
	})

	defaultResult := NewEngine(basePath, "default").Validate()
	defaultTrace := resultForCapability(t, defaultResult, "Traceability")
	if defaultTrace.Status != models.StatusWarning {
		t.Fatalf("expected default traceability WARNING, got %q with details %#v", defaultTrace.Status, defaultTrace.Details)
	}
	if len(defaultTrace.EvidenceQuality.ScoreImpacts) == 0 || defaultTrace.EvidenceQuality.ScoreImpacts[0].Delta != -25 {
		t.Fatalf("expected strong behavior traceability score impact, got %#v", defaultTrace.EvidenceQuality.ScoreImpacts)
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

func TestSimpleJSONArtifactsWithoutEvidenceStillValidateButWarnTraceability(t *testing.T) {
	basePath := t.TempDir()
	writeValidationProject(t, basePath, withoutTraceEvidence(map[string]string{
		"bottleneck/behavior/behavior-spec.md": "# Behavior Specification\n\n## Expected Behavior\n\nThe system stores validated deployment evidence before rendering summaries.\n\n## Unacceptable Behavior\n\nThe system must not ignore failed assurance results.\n",
		"bottleneck/intent/intent.md":          "# Intent\n\n## Outcomes\n\nThe CLI identifies blocking capability failures before release approval.\n\n## Constraints\n\nValidation remains deterministic offline.\n\n## Success Criteria\n\n- At least 95% of warnings include an artifact path.\n",
		"bottleneck/design/architecture.md":    "# Architecture\n\nThe CLI reads local artifacts, validates them deterministically, and renders release posture.\n",
	}))

	result := NewEngine(basePath, "default").Validate()

	for _, capability := range []string{"Assurance", "Security", "Execution"} {
		check := resultForCapability(t, result, capability)
		if check.Status != models.StatusPass {
			t.Fatalf("expected %s PASS without optional evidence arrays, got %q with details %#v", capability, check.Status, check.Details)
		}
	}

	traceability := resultForCapability(t, result, "Traceability")
	if traceability.Status != models.StatusWarning {
		t.Fatalf("expected Traceability WARNING for artifacts without IDs, got %q with details %#v", traceability.Status, traceability.Details)
	}
	if !containsString(traceability.Details, "no traceability evidence IDs found") {
		t.Fatalf("expected missing traceability ID detail, got %#v", traceability.Details)
	}
}

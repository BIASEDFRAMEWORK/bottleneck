package explainer

import (
	"strings"
	"testing"

	"biased/internal/models"
)

func TestRenderMapsAssuranceFailure(t *testing.T) {
	result := models.EngineResult{
		Environment:       "production",
		SystemStatus:      models.StatusFail,
		PrimaryBottleneck: "Assurance",
		Results: []models.ValidationResult{{
			Capability: "Assurance",
			Status:     models.StatusFail,
			Message:    "accuracy below threshold",
			Details: []string{
				"accuracy: 0.90 (threshold: 0.95)",
				"scenarios_failed: 0 (allowed: 0)",
			},
		}},
	}

	output, err := Render(result, "")
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	if !strings.Contains(output, "Owner: Assurance Engineer") {
		t.Fatalf("expected assurance owner mapping in output:\n%s", output)
	}
	if !strings.Contains(output, "Mapped Bottleneck: Validation gaps") {
		t.Fatalf("expected assurance bottleneck mapping in output:\n%s", output)
	}
}

func TestRenderFiltersByCapability(t *testing.T) {
	result := models.EngineResult{
		Environment:       "default",
		SystemStatus:      models.StatusFail,
		PrimaryBottleneck: "Assurance",
		Results: []models.ValidationResult{
			{Capability: "Behavior", Status: models.StatusPass},
			{Capability: "Assurance", Status: models.StatusFail, Message: "accuracy below threshold"},
		},
	}

	output, err := Render(result, "Behavior")
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	if !strings.Contains(output, "Capability: Behavior") {
		t.Fatalf("expected filtered capability in output:\n%s", output)
	}
	if strings.Contains(output, "Capability: Assurance") {
		t.Fatalf("did not expect non-filtered capability in output:\n%s", output)
	}
}

func TestRenderIncludesSummaryEvidenceAndActions(t *testing.T) {
	result := models.EngineResult{
		Environment:       "production",
		SystemStatus:      models.StatusFail,
		PrimaryBottleneck: "Assurance",
		Results: []models.ValidationResult{{
			Capability: "Assurance",
			Status:     models.StatusFail,
			Message:    "accuracy below threshold",
			Details: []string{
				"accuracy: 0.90 (threshold: 0.95)",
				"scenarios_failed: 0 (allowed: 0)",
			},
		}},
	}

	output, err := Render(result, "")
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	expectedSubstrings := []string{
		"Environment: production",
		"System Status: FAIL",
		"Primary Bottleneck: Assurance",
		"Owner: Assurance Engineer",
		"Mapped Bottleneck: Validation gaps",
		"Evidence:",
		"accuracy below threshold",
		"accuracy: 0.90 (threshold: 0.95)",
		"Recommended Next Actions:",
		"Inspect biased/assurance/results.json and confirm the scenario counts are correct.",
	}

	for _, expected := range expectedSubstrings {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected %q in output:\n%s", expected, output)
		}
	}
}

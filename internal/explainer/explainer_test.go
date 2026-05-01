package explainer

import (
	"strings"
	"testing"

	"bottleneck/internal/models"
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

	if !strings.Contains(output, "Behavior Score:") {
		t.Fatalf("expected filtered capability in output:\n%s", output)
	}
	if strings.Contains(output, "Assurance Score:") {
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
		"Evidence found:",
		"accuracy below threshold",
		"accuracy: 0.90 (threshold: 0.95)",
		"Evidence missing:",
		"Score impact:",
		"Recommendation:",
		"Fix failing tests or add passing assurance evidence until accuracy meets the selected threshold.",
	}

	for _, expected := range expectedSubstrings {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected %q in output:\n%s", expected, output)
		}
	}
}

func TestRenderSurfacesContentQualityWarningDetails(t *testing.T) {
	result := models.EngineResult{
		Environment:       "default",
		SystemStatus:      models.StatusWarning,
		PrimaryBottleneck: "Behavior",
		Results: []models.ValidationResult{{
			Capability: "Behavior",
			Status:     models.StatusWarning,
			Message:    "content quality warnings detected",
			Details: []string{
				`bottleneck/behavior/behavior-spec.md section "Expected Behavior" still contains placeholder content`,
			},
		}},
	}

	output, err := Render(result, "")
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	expectedSubstrings := []string{
		"System Status: WARNING",
		"Primary Bottleneck: Behavior",
		"Status: WARNING",
		"content quality warnings detected",
		`bottleneck/behavior/behavior-spec.md section "Expected Behavior" still contains placeholder content`,
		"Replace placeholder behavior text with concrete expected and unacceptable behavior.",
	}

	for _, expected := range expectedSubstrings {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected %q in output:\n%s", expected, output)
		}
	}
}

func TestRenderIncludesEvidenceQualityMissingAndScoreImpacts(t *testing.T) {
	result := models.EngineResult{
		Environment:       "default",
		SystemStatus:      models.StatusWarning,
		PrimaryBottleneck: "Intent",
		Results: []models.ValidationResult{{
			Capability: "Intent",
			Status:     models.StatusWarning,
			Message:    "intent evidence quality is weak",
			EvidenceQuality: models.EvidenceQuality{
				Score:   80,
				Missing: []string{"Add an INTENT-* heading such as ### INTENT-001: ..."},
				ScoreImpacts: []models.ScoreImpact{{
					Reason: "bottleneck/intent/intent.md does not define an INTENT-* evidence ID",
					Delta:  -20,
				}},
			},
		}},
	}

	output, err := Render(result, "")
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	expectedSubstrings := []string{
		"evidence quality score: 80",
		"Add an INTENT-* heading such as ### INTENT-001: ...",
		"- -20 bottleneck/intent/intent.md does not define an INTENT-* evidence ID",
	}
	for _, expected := range expectedSubstrings {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected %q in output:\n%s", expected, output)
		}
	}
}

func TestRenderIncludesPrimaryDiagnosisWhenNoCapabilityFilter(t *testing.T) {
	result := models.EngineResult{
		Environment:       "production",
		SystemStatus:      models.StatusFail,
		PrimaryBottleneck: "Assurance",
		Results: []models.ValidationResult{
			{Capability: "Behavior", Status: models.StatusPass},
			{Capability: "Intent", Status: models.StatusPass},
			{Capability: "Design", Status: models.StatusPass},
			{Capability: "Assurance", Status: models.StatusFail, Message: "missing results.json"},
			{Capability: "Security", Status: models.StatusPass},
			{Capability: "Execution", Status: models.StatusPass},
		},
	}

	output, err := Render(result, "")
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	expected := []string{
		"Primary Diagnosis:",
		"Weakest Category: Assurance",
		"Top Evidence:",
		"- missing results.json",
		"Next Action: Add assurance evidence that maps test or evaluation results to BEHAVIOR-001.",
	}
	for _, substring := range expected {
		if !strings.Contains(output, substring) {
			t.Fatalf("expected %q in output:\n%s", substring, output)
		}
	}
}

func TestRenderCapabilityFilterPreservesCapabilitySpecificOutput(t *testing.T) {
	result := models.EngineResult{
		Environment:       "production",
		SystemStatus:      models.StatusFail,
		PrimaryBottleneck: "Assurance",
		Results: []models.ValidationResult{
			{Capability: "Behavior", Status: models.StatusPass},
			{Capability: "Assurance", Status: models.StatusFail, Message: "missing results.json"},
		},
	}

	output, err := Render(result, "Behavior")
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	if strings.Contains(output, "Primary Diagnosis:") {
		t.Fatalf("did not expect global diagnosis for filtered output:\n%s", output)
	}
	if !strings.Contains(output, "Behavior Score:") {
		t.Fatalf("expected behavior capability output:\n%s", output)
	}
}

func TestRenderUsesEvidenceDrivenSectionsAndAvoidsGenericDescriptions(t *testing.T) {
	result := models.EngineResult{
		Environment:       "default",
		SystemStatus:      models.StatusFail,
		PrimaryBottleneck: "Assurance",
		Results: []models.ValidationResult{{
			Capability: "Assurance",
			Status:     models.StatusFail,
			Message:    "scenarios_failed above threshold",
			Details: []string{
				"failure: Ambiguous risk clause was summarized as confirmed exposure",
				"accuracy: 0.50 (threshold: 0.90)",
				"scenarios_failed: 1 (allowed: 0)",
			},
			EvidenceQuality: models.EvidenceQuality{
				Score: 90,
				ScoreImpacts: []models.ScoreImpact{{
					Reason: "scenarios_failed above threshold",
					Delta:  -40,
				}},
			},
		}},
	}

	output, err := Render(result, "")
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	expected := []string{
		"Assurance Score:",
		"Evidence found:",
		"- bottleneck/assurance/results.json exists",
		"- failure: Ambiguous risk clause was summarized as confirmed exposure",
		"Evidence missing:",
		"Related IDs:",
		"Score impact:",
		"- -40 scenarios_failed above threshold",
		"Recommendation:",
	}
	for _, substring := range expected {
		if !strings.Contains(output, substring) {
			t.Fatalf("expected %q in output:\n%s", substring, output)
		}
	}
	if strings.Contains(output, "Assurance proves") || strings.Contains(output, "Intent defines") {
		t.Fatalf("explain output should avoid generic framework descriptions:\n%s", output)
	}
}

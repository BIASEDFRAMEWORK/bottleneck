package scorecard

import (
	"encoding/json"
	"strings"
	"testing"

	"biased/internal/models"
)

func TestBuildMapsAssuranceFailure(t *testing.T) {
	result := models.EngineResult{
		Environment:       "production",
		SystemStatus:      models.StatusFail,
		PrimaryBottleneck: "Assurance",
		Results: []models.ValidationResult{{
			Capability: "Assurance",
			Status:     models.StatusFail,
			Message:    "accuracy below threshold",
		}},
	}

	card := Build(result)
	if len(card.Capabilities) != 1 {
		t.Fatalf("expected 1 capability, got %d", len(card.Capabilities))
	}

	capability := card.Capabilities[0]
	if capability.Owner != "Assurance Engineer" {
		t.Fatalf("expected owner mapping, got %q", capability.Owner)
	}
	if capability.Bottleneck != "Validation gaps" {
		t.Fatalf("expected bottleneck mapping, got %q", capability.Bottleneck)
	}
}

func TestRenderTextIncludesSummaryAndRowData(t *testing.T) {
	result := models.EngineResult{
		Environment:       "production",
		SystemStatus:      models.StatusFail,
		PrimaryBottleneck: "Assurance",
		Results: []models.ValidationResult{
			{
				Capability: "Assurance",
				Status:     models.StatusFail,
				Message:    "accuracy below threshold",
				Details: []string{
					"accuracy: 0.90 (threshold: 0.95)",
				},
			},
		},
	}

	output, err := Render(result, "text")
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	expected := []string{
		"BIASED System Scorecard",
		"Environment: production",
		"System Status: FAIL",
		"Primary Bottleneck: Assurance",
		"Assurance",
		"Assurance Engineer",
		"Validation gaps",
		"accuracy below threshold",
		"Bottom line:",
		"The system is not valid for production. Primary ownership starts with Assurance Engineer.",
	}

	for _, substring := range expected {
		if !strings.Contains(output, substring) {
			t.Fatalf("expected %q in output:\n%s", substring, output)
		}
	}
}

func TestRenderJSONProducesValidJSON(t *testing.T) {
	result := models.EngineResult{
		Environment:       "default",
		SystemStatus:      models.StatusPass,
		PrimaryBottleneck: "None",
		Results: []models.ValidationResult{
			{
				Capability: "Intent",
				Status:     models.StatusPass,
			},
		},
	}

	output, err := Render(result, "json")
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	var decoded Scorecard
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("expected valid json, got error: %v\n%s", err, output)
	}

	if decoded.Environment != "default" {
		t.Fatalf("expected environment default, got %q", decoded.Environment)
	}
	if decoded.SystemStatus != models.StatusPass {
		t.Fatalf("expected PASS system status, got %q", decoded.SystemStatus)
	}
	if decoded.PrimaryBottleneck != "None" {
		t.Fatalf("expected primary bottleneck None, got %q", decoded.PrimaryBottleneck)
	}
	if len(decoded.Capabilities) != 1 {
		t.Fatalf("expected capability array, got %d", len(decoded.Capabilities))
	}
	if decoded.Capabilities[0].Owner != "Intent Engineer" {
		t.Fatalf("expected owner field in json, got %q", decoded.Capabilities[0].Owner)
	}
	if decoded.Capabilities[0].Bottleneck != "Ambiguous requirements" {
		t.Fatalf("expected bottleneck field in json, got %q", decoded.Capabilities[0].Bottleneck)
	}
}

func TestRenderRejectsUnsupportedFormat(t *testing.T) {
	result := models.EngineResult{}

	_, err := Render(result, "xml")
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Fatalf("unexpected error: %v", err)
	}
}

package cmd

import (
	"strings"
	"testing"

	"bottleneck/internal/models"
)

func TestRenderValidationOutputSurfacesPlaceholderDetails(t *testing.T) {
	result := models.EngineResult{
		Environment:       "default",
		SystemStatus:      models.StatusWarning,
		PrimaryBottleneck: "Intent",
		Results: []models.ValidationResult{{
			Capability: "Intent",
			Status:     models.StatusWarning,
			Message:    "content quality warnings detected",
			Details: []string{
				`bottleneck/intent/intent.md section "Outcomes" still contains placeholder content`,
			},
		}},
	}

	output := renderValidationOutput(result)

	expected := []string{
		"Intent: WARNING (content quality warnings detected)",
		`  bottleneck/intent/intent.md section "Outcomes" still contains placeholder content`,
		"System Status: WARNING",
		"Primary Bottleneck: Intent",
	}

	for _, substring := range expected {
		if !strings.Contains(output, substring) {
			t.Fatalf("expected %q in output:\n%s", substring, output)
		}
	}
}

package githubannotations

import (
	"strings"
	"testing"

	"bottleneck/internal/models"
)

func TestRenderEmitsWorkflowCommands(t *testing.T) {
	output := Render([]models.ValidationResult{
		{
			Capability: "Intent",
			Status:     models.StatusWarning,
			Message:    "content quality warnings detected",
			Details: []string{
				`bottleneck/intent/intent.md section "Outcomes" still contains placeholder content`,
			},
		},
		{
			Capability: "Assurance",
			Status:     models.StatusFail,
			Message:    "accuracy below threshold",
			Details: []string{
				"accuracy: 0.90 (threshold: 0.95)",
			},
		},
	})

	expected := []string{
		`::warning file=bottleneck/intent/intent.md::bottleneck/intent/intent.md section "Outcomes" still contains placeholder content`,
		`::error file=bottleneck/assurance/results.json::accuracy: 0.90 (threshold: 0.95)`,
	}
	for _, substring := range expected {
		if !strings.Contains(output, substring) {
			t.Fatalf("expected %q in annotations:\n%s", substring, output)
		}
	}
}

func TestRenderEscapesWorkflowCommandCharacters(t *testing.T) {
	output := Render([]models.ValidationResult{{
		Capability: "Intent",
		Status:     models.StatusFail,
		Findings: []models.ValidationFinding{{
			Level:   "error",
			Path:    "bottleneck/intent/intent.md:1,2",
			Message: "bad % value\nnext line",
		}},
	}})

	expected := `::error file=bottleneck/intent/intent.md%3A1%2C2::bad %25 value%0Anext line`
	if output != expected {
		t.Fatalf("expected escaped annotation %q, got %q", expected, output)
	}
}

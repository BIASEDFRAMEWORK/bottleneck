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
			Message: "bad % value\nnext: line, here",
		}},
	}})

	expected := `::error file=bottleneck/intent/intent.md%3A1%2C2::bad %25 value%0Anext: line, here`
	if output != expected {
		t.Fatalf("expected escaped annotation %q, got %q", expected, output)
	}
}

func TestRenderPreservesFindingPathAndLine(t *testing.T) {
	output := Render([]models.ValidationResult{{
		Capability: "Intent",
		Status:     models.StatusWarning,
		Findings: []models.ValidationFinding{{
			Level:   "warning",
			Path:    "bottleneck/intent/intent.md",
			Line:    7,
			Message: "Intent evidence contains placeholder content",
		}},
	}})

	expected := `::warning file=bottleneck/intent/intent.md,line=7::Intent evidence contains placeholder content`
	if output != expected {
		t.Fatalf("expected annotation with line %q, got %q", expected, output)
	}
}

func TestRenderInfersPathAndLineFromDetails(t *testing.T) {
	output := Render([]models.ValidationResult{{
		Capability: "Behavior",
		Status:     models.StatusWarning,
		Message:    "content quality warnings detected",
		Details:    []string{`bottleneck/behavior/behavior-spec.md:12 section "Expected Behavior" is too thin`},
	}})

	expected := `::warning file=bottleneck/behavior/behavior-spec.md,line=12::bottleneck/behavior/behavior-spec.md:12 section "Expected Behavior" is too thin`
	if output != expected {
		t.Fatalf("expected inferred path and line %q, got %q", expected, output)
	}
}

func TestRenderEscapesMessageNewlinesAndPropertySeparators(t *testing.T) {
	output := Render([]models.ValidationResult{{
		Capability: "Security",
		Status:     models.StatusFail,
		Findings: []models.ValidationFinding{{
			Level:   "error",
			Path:    "bottleneck/security/guardrails.json:1,2",
			Message: "bad % value\nnext line",
		}},
	}})

	expected := `::error file=bottleneck/security/guardrails.json%3A1%2C2::bad %25 value%0Anext line`
	if output != expected {
		t.Fatalf("expected escaped annotation %q, got %q", expected, output)
	}
}

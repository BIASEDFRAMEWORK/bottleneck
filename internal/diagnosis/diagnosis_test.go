package diagnosis

import (
	"testing"

	"bottleneck/internal/models"
)

func TestAnalyzeSelectsSingleWeakestCategory(t *testing.T) {
	result := resultWithCategories([]models.ValidationResult{
		{Capability: "Behavior", Status: models.StatusPass},
		{Capability: "Intent", Status: models.StatusPass},
		{Capability: "Design", Status: models.StatusPass},
		{Capability: "Assurance", Status: models.StatusFail, Message: "missing results.json"},
		{Capability: "Security", Status: models.StatusPass},
		{Capability: "Execution", Status: models.StatusPass},
	})

	diagnosis := Analyze(result)

	if diagnosis.PrimaryBottleneck != "Assurance" {
		t.Fatalf("expected Assurance bottleneck, got %q", diagnosis.PrimaryBottleneck)
	}
	if diagnosis.RecommendedAction != "Add assurance evidence that maps test or evaluation results to BEHAVIOR-001." {
		t.Fatalf("unexpected action %q", diagnosis.RecommendedAction)
	}
}

func TestAnalyzeHandlesTiedBottlenecksByPriority(t *testing.T) {
	result := resultWithCategories([]models.ValidationResult{
		{Capability: "Behavior", Status: models.StatusPass},
		{Capability: "Intent", Status: models.StatusPass},
		{Capability: "Design", Status: models.StatusPass},
		{Capability: "Assurance", Status: models.StatusFail, Message: "missing results.json"},
		{Capability: "Security", Status: models.StatusFail, Message: "missing guardrails.json"},
		{Capability: "Execution", Status: models.StatusPass},
	})

	diagnosis := Analyze(result)

	if diagnosis.PrimaryBottleneck != "Assurance" {
		t.Fatalf("expected Assurance priority winner, got %q", diagnosis.PrimaryBottleneck)
	}
	expected := []string{"Assurance", "Security"}
	if len(diagnosis.TiedBottlenecks) != len(expected) {
		t.Fatalf("expected tied bottlenecks %#v, got %#v", expected, diagnosis.TiedBottlenecks)
	}
	for index, value := range expected {
		if diagnosis.TiedBottlenecks[index] != value {
			t.Fatalf("expected tied bottlenecks %#v, got %#v", expected, diagnosis.TiedBottlenecks)
		}
	}
}

func TestAnalyzeReportsHealthyWhenAllCategoriesPass(t *testing.T) {
	result := resultWithCategories([]models.ValidationResult{
		{Capability: "Behavior", Status: models.StatusPass, Details: []string{"behavior evidence"}},
		{Capability: "Intent", Status: models.StatusPass, Details: []string{"intent evidence"}},
		{Capability: "Design", Status: models.StatusPass, Details: []string{"design evidence"}},
		{Capability: "Assurance", Status: models.StatusPass, Details: []string{"accuracy: 1.00"}},
		{Capability: "Security", Status: models.StatusPass, Details: []string{"violations: 0"}},
		{Capability: "Execution", Status: models.StatusPass, Details: []string{"error_rate: 0.01"}},
	})

	diagnosis := Analyze(result)

	if diagnosis.PrimaryBottleneck != HealthyPrimaryBottleneck {
		t.Fatalf("expected healthy bottleneck, got %q", diagnosis.PrimaryBottleneck)
	}
	if len(diagnosis.TiedBottlenecks) != 0 {
		t.Fatalf("expected no tied bottlenecks, got %#v", diagnosis.TiedBottlenecks)
	}
}

func TestAnalyzeReducesScoreForTraceabilityGaps(t *testing.T) {
	result := resultWithCategories([]models.ValidationResult{
		{Capability: "Behavior", Status: models.StatusPass, Details: []string{"behavior evidence"}},
		{Capability: "Intent", Status: models.StatusPass, Details: []string{"intent evidence"}},
		{Capability: "Design", Status: models.StatusPass, Details: []string{"design evidence"}},
		{Capability: "Assurance", Status: models.StatusPass, Details: []string{"accuracy: 1.00"}},
		{Capability: "Security", Status: models.StatusPass, Details: []string{"violations: 0"}},
		{Capability: "Execution", Status: models.StatusPass, Details: []string{"error_rate: 0.01"}},
		{
			Capability: "Traceability",
			Status:     models.StatusFail,
			Message:    "traceability failures detected",
			Details: []string{
				"bottleneck/behavior/behavior-spec.md BEHAVIOR-001 references missing ASSURANCE-001",
			},
		},
	})

	diagnosis := Analyze(result)

	if diagnosis.PrimaryBottleneck != "Behavior" {
		t.Fatalf("expected Behavior traceability bottleneck, got %q", diagnosis.PrimaryBottleneck)
	}
	if score := ScoreFor("Behavior", diagnosis.CategoryScores); score >= ScorePass {
		t.Fatalf("expected reduced behavior score, got %d", score)
	}
}

func TestRecommendedActionChangesForWeakStaleAndDisconnectedEvidence(t *testing.T) {
	tests := []struct {
		name       string
		validation models.ValidationResult
		expected   string
	}{
		{
			name: "weak",
			validation: models.ValidationResult{
				Capability: "Intent",
				Status:     models.StatusWarning,
				Message:    "content quality warnings detected",
				Details:    []string{`bottleneck/intent/intent.md section "Outcomes" still contains placeholder content`},
			},
			expected: "Replace placeholder intent statements with 1-3 measurable outcomes.",
		},
		{
			name: "stale",
			validation: models.ValidationResult{
				Capability: "Execution",
				Status:     models.StatusWarning,
				Message:    "stale telemetry evidence",
			},
			expected: "Refresh execution telemetry from the current environment before release.",
		},
		{
			name: "disconnected",
			validation: models.ValidationResult{
				Capability: "Behavior",
				Status:     models.StatusWarning,
				Message:    "traceability warnings detected",
				Details:    []string{"bottleneck/behavior/behavior-spec.md BEHAVIOR-001 is not linked to intent evidence"},
			},
			expected: "Link BEHAVIOR-001 to its supporting INTENT and ASSURANCE evidence.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if action := RecommendedAction(tt.validation); action != tt.expected {
				t.Fatalf("expected action %q, got %q", tt.expected, action)
			}
		})
	}
}

func resultWithCategories(results []models.ValidationResult) models.EngineResult {
	return models.EngineResult{
		Environment:       "production",
		SystemStatus:      models.StatusFail,
		PrimaryBottleneck: "Assurance",
		Results:           results,
	}
}

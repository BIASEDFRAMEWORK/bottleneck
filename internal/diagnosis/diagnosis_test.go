package diagnosis

import (
	"encoding/json"
	"strings"
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
	if diagnosis.Confidence != ConfidenceHigh {
		t.Fatalf("expected high confidence, got %q (%s)", diagnosis.Confidence, diagnosis.ConfidenceReason)
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
	if diagnosis.Confidence != ConfidenceMedium {
		t.Fatalf("expected broken traceability to reduce confidence to Medium, got %q", diagnosis.Confidence)
	}
}

func TestAnalyzeCollectsTopThreeContributingFindings(t *testing.T) {
	result := resultWithCategories([]models.ValidationResult{
		{
			Capability: "Behavior",
			Status:     models.StatusWarning,
			Message:    "behavior evidence quality is weak",
			Details: []string{
				`bottleneck/behavior/behavior-spec.md section "Expected Behavior" still contains placeholder content`,
			},
			EvidenceQuality: models.EvidenceQuality{
				Score: 25,
				ScoreImpacts: []models.ScoreImpact{
					{Reason: "missing BEHAVIOR-001", Delta: -20},
					{Reason: "placeholder-heavy behavior spec", Delta: -30},
					{Reason: "unacceptable behavior still placeholder", Delta: -20},
					{Reason: "extra lower priority reason", Delta: -10},
				},
			},
		},
		{Capability: "Intent", Status: models.StatusPass},
		{Capability: "Design", Status: models.StatusPass},
		{Capability: "Assurance", Status: models.StatusPass},
		{Capability: "Security", Status: models.StatusPass},
		{Capability: "Execution", Status: models.StatusPass},
	})

	diagnosis := Analyze(result)

	expected := []string{"missing BEHAVIOR-001", "placeholder-heavy behavior spec", "unacceptable behavior still placeholder"}
	if strings.Join(diagnosis.ContributingFindings, "|") != strings.Join(expected, "|") {
		t.Fatalf("expected top findings %#v, got %#v", expected, diagnosis.ContributingFindings)
	}
}

func TestConfidenceLevelsReflectEvidenceDepth(t *testing.T) {
	tests := []struct {
		name       string
		result     models.EngineResult
		confidence string
		reason     string
	}{
		{
			name: "low sparse",
			result: resultWithCategories([]models.ValidationResult{
				{Capability: "Behavior", Status: models.StatusPass},
				{Capability: "Intent", Status: models.StatusPass},
				{Capability: "Design", Status: models.StatusFail, Message: "missing architecture.md"},
				{Capability: "Assurance", Status: models.StatusFail, Message: "missing results.json"},
				{Capability: "Security", Status: models.StatusFail, Message: "missing guardrails.json"},
				{Capability: "Execution", Status: models.StatusFail, Message: "missing telemetry.json"},
			}),
			confidence: ConfidenceLow,
			reason:     "Only 2 of 6 evidence categories contain meaningful content.",
		},
		{
			name: "medium partial",
			result: resultWithCategories([]models.ValidationResult{
				{Capability: "Behavior", Status: models.StatusPass},
				{Capability: "Intent", Status: models.StatusPass},
				{Capability: "Design", Status: models.StatusPass},
				{Capability: "Assurance", Status: models.StatusPass},
				{Capability: "Security", Status: models.StatusFail, Message: "missing guardrails.json"},
				{Capability: "Execution", Status: models.StatusFail, Message: "missing telemetry.json"},
			}),
			confidence: ConfidenceMedium,
			reason:     "4 of 6 evidence categories contain meaningful content.",
		},
		{
			name: "high complete connected",
			result: resultWithCategories([]models.ValidationResult{
				{Capability: "Behavior", Status: models.StatusPass},
				{Capability: "Intent", Status: models.StatusPass},
				{Capability: "Design", Status: models.StatusPass},
				{Capability: "Assurance", Status: models.StatusPass},
				{Capability: "Security", Status: models.StatusPass},
				{Capability: "Execution", Status: models.StatusPass},
				{Capability: "Traceability", Status: models.StatusPass},
			}),
			confidence: ConfidenceHigh,
			reason:     "All 6 evidence categories contain meaningful, connected evidence.",
		},
		{
			name: "broken traceability lowers",
			result: resultWithCategories([]models.ValidationResult{
				{Capability: "Behavior", Status: models.StatusPass},
				{Capability: "Intent", Status: models.StatusPass},
				{Capability: "Design", Status: models.StatusPass},
				{Capability: "Assurance", Status: models.StatusPass},
				{Capability: "Security", Status: models.StatusPass},
				{Capability: "Execution", Status: models.StatusPass},
				{Capability: "Traceability", Status: models.StatusFail, Details: []string{"bottleneck/behavior/behavior-spec.md BEHAVIOR-001 references missing ASSURANCE-001"}},
			}),
			confidence: ConfidenceMedium,
			reason:     "All 6 evidence categories are present, but traceability has broken references.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confidence, reason := Confidence(tt.result)
			if confidence != tt.confidence || reason != tt.reason {
				t.Fatalf("expected %s/%q, got %s/%q", tt.confidence, tt.reason, confidence, reason)
			}
		})
	}
}

func TestRenderDiagnosisFormats(t *testing.T) {
	result := resultWithCategories([]models.ValidationResult{
		{Capability: "Behavior", Status: models.StatusWarning, Message: "behavior evidence quality is weak", Details: []string{"No unacceptable behaviors defined."}},
		{Capability: "Intent", Status: models.StatusPass},
		{Capability: "Design", Status: models.StatusPass},
		{Capability: "Assurance", Status: models.StatusPass},
		{Capability: "Security", Status: models.StatusPass},
		{Capability: "Execution", Status: models.StatusPass},
	})

	text, err := Render(result, FormatText)
	if err != nil {
		t.Fatalf("Render text returned error: %v", err)
	}
	for _, substring := range []string{
		"Primary Bottleneck: Behavior",
		"Contributing findings:",
		"1. behavior evidence quality is weak",
		"Recommended next action:",
		"Diagnosis Confidence:",
	} {
		if !strings.Contains(text, substring) {
			t.Fatalf("expected %q in text output:\n%s", substring, text)
		}
	}

	jsonOutput, err := Render(result, FormatJSON)
	if err != nil {
		t.Fatalf("Render json returned error: %v", err)
	}
	var report Report
	if err := json.Unmarshal([]byte(jsonOutput), &report); err != nil {
		t.Fatalf("expected stable JSON, got %v\n%s", err, jsonOutput)
	}
	if report.SchemaVersion != DiagnoseSchemaVersion || report.PrimaryBottleneck != "Behavior" {
		t.Fatalf("unexpected JSON report %#v", report)
	}
	if len(report.CategoryScores) == 0 {
		t.Fatalf("expected JSON report to include category scores: %#v", report)
	}

	markdown, err := Render(result, FormatMarkdown)
	if err != nil {
		t.Fatalf("Render markdown returned error: %v", err)
	}
	for _, substring := range []string{
		"## Bottleneck Diagnosis",
		"| Primary Bottleneck | Behavior |",
		"### Category Scores",
		"| Category | Score | Status |",
		"| Behavior | 60 | WARNING |",
		"### Top Findings",
		"### Recommended Next Action",
	} {
		if !strings.Contains(markdown, substring) {
			t.Fatalf("expected %q in markdown output:\n%s", substring, markdown)
		}
	}
	if strings.Contains(markdown, "\x1b[") {
		t.Fatalf("expected markdown output without ANSI formatting:\n%s", markdown)
	}

	github, err := Render(result, FormatGitHub)
	if err != nil {
		t.Fatalf("Render github returned error: %v", err)
	}
	if !strings.Contains(github, "::warning file=bottleneck/behavior/behavior-spec.md::No unacceptable behaviors defined.") {
		t.Fatalf("expected GitHub annotation output, got:\n%s", github)
	}
}

func TestRenderMarkdownDiagnosisIsPRFriendly(t *testing.T) {
	result := resultWithCategories([]models.ValidationResult{
		{Capability: "Behavior", Status: models.StatusPass},
		{Capability: "Intent", Status: models.StatusPass},
		{Capability: "Design", Status: models.StatusPass},
		{
			Capability: "Assurance",
			Status:     models.StatusFail,
			Message:    "accuracy below threshold",
			Details:    []string{"accuracy: 0.50 (threshold: 0.90)"},
		},
		{Capability: "Security", Status: models.StatusPass},
		{Capability: "Execution", Status: models.StatusPass},
	})

	markdown, err := Render(result, FormatMarkdown)
	if err != nil {
		t.Fatalf("Render markdown returned error: %v", err)
	}

	expected := []string{
		"## Bottleneck Diagnosis",
		"| Field | Value |",
		"| Primary Bottleneck | Assurance |",
		"### Category Scores",
		"| Category | Score | Status |",
		"| Assurance | 15 | FAIL |",
		"### Top Findings",
		"1. accuracy: 0.50 (threshold: 0.90)",
		"### Recommended Next Action",
		"Fix failing tests or add passing assurance evidence until accuracy meets the selected threshold.",
	}
	for _, substring := range expected {
		if !strings.Contains(markdown, substring) {
			t.Fatalf("expected %q in markdown output:\n%s", substring, markdown)
		}
	}
	if strings.Contains(markdown, "\x1b[") {
		t.Fatalf("expected markdown output without ANSI formatting:\n%s", markdown)
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

package report

import (
	"strings"
	"testing"
	"time"

	"bottleneck/internal/diagnosis"
	"bottleneck/internal/explainer"
	"bottleneck/internal/scorecard"
	"bottleneck/internal/trends"
)

func TestReportIncludesExecutiveSummary(t *testing.T) {
	output := RenderMarkdown(sampleReport(t, nil))
	if !strings.Contains(output, "## Executive Summary") ||
		!strings.Contains(output, "The current evidence suggests") {
		t.Fatalf("expected executive summary:\n%s", output)
	}
}

func TestReportIncludesCurrentStatus(t *testing.T) {
	output := RenderMarkdown(sampleReport(t, nil))
	for _, expected := range []string{"## Current Delivery-System Status", "Current status: `WARN`", "Release recommendation: `Conditional`"} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected %q in report:\n%s", expected, output)
		}
	}
}

func TestReportIncludesPrimaryBottleneck(t *testing.T) {
	output := RenderMarkdown(sampleReport(t, nil))
	if !strings.Contains(output, "## Primary Bottleneck") ||
		!strings.Contains(output, "The primary constraint appears to be `Assurance`.") {
		t.Fatalf("expected primary bottleneck section:\n%s", output)
	}
}

func TestReportIncludesCategoryScorecard(t *testing.T) {
	output := RenderMarkdown(sampleReport(t, nil))
	if !strings.Contains(output, "## Category Scorecard") ||
		!strings.Contains(output, "| Assurance | WARN | 60 | Assurance Engineer |") {
		t.Fatalf("expected category scorecard:\n%s", output)
	}
}

func TestReportIncludesTrendSummaryWhenSnapshotsExist(t *testing.T) {
	trend := sampleTrend()
	output := RenderMarkdown(sampleReport(t, &trend))
	if !strings.Contains(output, "## Trend Summary") ||
		!strings.Contains(output, "Overall direction is `improving`") ||
		!strings.Contains(output, "Assurance appeared as the primary bottleneck in 2 of 3 snapshots.") {
		t.Fatalf("expected trend summary:\n%s", output)
	}
}

func TestReportHandlesMissingTrendHistory(t *testing.T) {
	trend := trends.Analysis{
		Environment:       "default",
		SnapshotCount:     0,
		Window:            6,
		OverallDirection:  trends.DirectionInsufficientHistory,
		LeadershipSummary: "Not enough historical snapshots exist yet. Create snapshots over multiple delivery cycles to establish a trend.",
	}
	output := RenderMarkdown(sampleReport(t, &trend))
	if !strings.Contains(output, "Trend history is insufficient") ||
		!strings.Contains(output, "bottleneck snapshot") {
		t.Fatalf("expected missing trend history guidance:\n%s", output)
	}
}

func TestReportIncludesEvidenceFound(t *testing.T) {
	output := RenderMarkdown(sampleReport(t, nil))
	if !strings.Contains(output, "## Evidence Found") ||
		!strings.Contains(output, "Assurance: bottleneck/assurance/results.json exists") {
		t.Fatalf("expected evidence found:\n%s", output)
	}
}

func TestReportIncludesEvidenceMissing(t *testing.T) {
	output := RenderMarkdown(sampleReport(t, nil))
	if !strings.Contains(output, "## Evidence Missing") ||
		!strings.Contains(output, "Assurance: BEHAVIOR-003 has no mapped test evidence") {
		t.Fatalf("expected evidence missing:\n%s", output)
	}
}

func TestReportIncludesRisks(t *testing.T) {
	output := RenderMarkdown(sampleReport(t, nil))
	if !strings.Contains(output, "## Risk to Delivery") ||
		!strings.Contains(output, "cannot be proven against intended behavior") {
		t.Fatalf("expected risk section:\n%s", output)
	}
}

func TestReportIncludesRecommendations(t *testing.T) {
	output := RenderMarkdown(sampleReport(t, nil))
	if !strings.Contains(output, "## Recommended Actions") ||
		!strings.Contains(output, "Map BEHAVIOR-003 to passing test evidence.") {
		t.Fatalf("expected recommended actions:\n%s", output)
	}
}

func TestReportIncludesSuggestedOwners(t *testing.T) {
	output := RenderMarkdown(sampleReport(t, nil))
	if !strings.Contains(output, "## Suggested Owners") ||
		!strings.Contains(output, "QA/Assurance Engineer") {
		t.Fatalf("expected owners:\n%s", output)
	}
}

func TestReportIncludesSuggestedAutomation(t *testing.T) {
	output := RenderMarkdown(sampleReport(t, nil))
	if !strings.Contains(output, "## Suggested Automation") ||
		!strings.Contains(output, "Run Cucumber in GitHub Actions.") {
		t.Fatalf("expected automation:\n%s", output)
	}
}

func TestReportIncludesLeadershipDecision(t *testing.T) {
	output := RenderMarkdown(sampleReport(t, nil))
	if !strings.Contains(output, "## Leadership Decision Needed") ||
		!strings.Contains(output, "Approve time to map critical behaviors") {
		t.Fatalf("expected leadership decision:\n%s", output)
	}
}

func TestReportSupportsJSON(t *testing.T) {
	output, err := RenderJSON(sampleReport(t, nil))
	if err != nil {
		t.Fatalf("RenderJSON returned error: %v", err)
	}
	if !strings.Contains(output, `"schema_version": "sdlc.evidence.report.v1"`) ||
		!strings.Contains(output, `"leadership_decision"`) ||
		!strings.Contains(output, `"evidence_missing"`) {
		t.Fatalf("expected structured JSON report:\n%s", output)
	}
}

func sampleReport(t *testing.T, trend *trends.Analysis) Report {
	t.Helper()
	card := scorecard.Scorecard{
		SchemaVersion:         scorecard.SchemaVersion,
		Environment:           "default",
		SystemStatus:          scorecard.StatusWarn,
		ReleaseRecommendation: scorecard.RecommendationConditional,
		PrimaryBottleneck:     "Assurance",
		Capabilities: []scorecard.CapabilityScorecard{{
			Capability:        "Assurance",
			Status:            scorecard.StatusWarn,
			Score:             60,
			Owner:             "Assurance Engineer",
			RecommendedAction: "Map BEHAVIOR-003 to passing test evidence.",
		}},
	}
	explanation := explainer.Report{
		SchemaVersion:     explainer.SchemaVersion,
		Environment:       "default",
		SystemStatus:      scorecard.StatusWarn,
		PrimaryBottleneck: "Assurance",
		Diagnosis: diagnosis.Diagnosis{
			PrimaryBottleneck: "Assurance",
			WhyItMatters:      "Assurance evidence supports release confidence.",
		},
		Explanations: []explainer.CategoryExplanation{{
			Category:           "Assurance",
			Status:             scorecard.StatusWarn,
			Score:              60,
			EvidenceFound:      []explainer.EvidenceFact{{Text: "bottleneck/assurance/results.json exists"}},
			EvidenceMissing:    []explainer.EvidenceGap{{Text: "BEHAVIOR-003 has no mapped test evidence"}},
			RiskToDelivery:     "The team may ship functionality that cannot be proven against intended behavior.",
			RecommendedActions: []string{"Map BEHAVIOR-003 to passing test evidence."},
			SuggestedOwnerRoles: []string{
				"QA/Assurance Engineer",
				"Developer",
			},
			SuggestedAutomations: []string{"Run Cucumber in GitHub Actions."},
		}},
	}
	return Build(card, trend, explanation, Options{
		GeneratedAt: time.Date(2026, 5, 1, 15, 30, 0, 0, time.UTC),
	})
}

func sampleTrend() trends.Analysis {
	return trends.Analysis{
		Environment:      "default",
		SnapshotCount:    3,
		Window:           6,
		OverallDirection: trends.DirectionImproving,
		PersistentBottleneck: trends.PersistentBottleneck{
			Category: "Assurance",
			Count:    2,
			Total:    3,
			Summary:  "Assurance appeared as the primary bottleneck in 2 of 3 snapshots.",
		},
		LeadershipSummary: "The team is improving overall, but Assurance remains the most persistent delivery constraint.",
	}
}

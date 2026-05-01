package gate

import (
	"strings"
	"testing"

	"bottleneck/internal/config"
	"bottleneck/internal/models"
)

func TestEvaluateReleasePassesWhenThresholdsAreMet(t *testing.T) {
	report := EvaluateRelease(validEngineResult(), config.ReleaseGateConfig{
		MinPrimaryScore:     75,
		RequiredCategories:  []string{"Intent", "Behavior", "Assurance", "Security", "Execution"},
		RequireTraceability: true,
		RequireGovernance:   false,
	})

	if report.Status != StatusPass {
		t.Fatalf("expected release gate pass, got %#v", report)
	}
	if report.PrimaryScore < report.RequiredScore {
		t.Fatalf("expected primary score to satisfy threshold, got %#v", report)
	}
	if len(report.NotAssessed) == 0 || !strings.Contains(report.NotAssessed[0], "Governance") {
		t.Fatalf("expected governance not assessed note, got %#v", report.NotAssessed)
	}
}

func TestEvaluateReleaseFailsWhenPrimaryScoreBelowThreshold(t *testing.T) {
	result := validEngineResult()
	result.Results[3] = models.ValidationResult{
		Capability: "Assurance",
		Status:     models.StatusFail,
		Message:    "accuracy below threshold",
		Details:    []string{"accuracy: 0.50 (threshold: 0.90)"},
	}

	report := EvaluateRelease(result, config.ReleaseGateConfig{
		MinPrimaryScore:     60,
		RequiredCategories:  []string{"Intent", "Behavior", "Assurance", "Security", "Execution"},
		RequireTraceability: true,
	})

	expectGateReason(t, report, "Primary bottleneck score is below release threshold.")
}

func TestEvaluateReleaseFailsWhenRequiredCategoryIsMissing(t *testing.T) {
	result := validEngineResult()
	result.Results = []models.ValidationResult{
		{Capability: "Behavior", Status: models.StatusPass},
		{Capability: "Assurance", Status: models.StatusPass},
		{Capability: "Security", Status: models.StatusPass},
		{Capability: "Execution", Status: models.StatusPass},
		{Capability: "Traceability", Status: models.StatusPass},
	}

	report := EvaluateRelease(result, config.ReleaseGateConfig{
		MinPrimaryScore:     60,
		RequiredCategories:  []string{"Intent", "Behavior"},
		RequireTraceability: true,
	})

	expectGateReason(t, report, "Required category Intent is missing.")
}

func TestEvaluateReleaseFailsWhenTraceabilityIsBroken(t *testing.T) {
	result := validEngineResult()
	result.Results[6] = models.ValidationResult{
		Capability: "Traceability",
		Status:     models.StatusFail,
		Message:    "traceability failures detected",
		Details:    []string{"bottleneck/behavior/behavior-spec.md BEHAVIOR-001 references missing ASSURANCE-001"},
	}

	report := EvaluateRelease(result, config.ReleaseGateConfig{
		MinPrimaryScore:     60,
		RequiredCategories:  []string{"Intent", "Behavior", "Assurance", "Security", "Execution"},
		RequireTraceability: true,
	})

	expectGateReason(t, report, "Traceability is broken.")
}

func TestEvaluateReleaseFailsWhenSecurityFails(t *testing.T) {
	result := validEngineResult()
	result.Results[4] = models.ValidationResult{
		Capability: "Security",
		Status:     models.StatusFail,
		Message:    "violations above threshold",
	}

	report := EvaluateRelease(result, config.ReleaseGateConfig{
		MinPrimaryScore:     60,
		RequiredCategories:  []string{"Intent", "Behavior", "Assurance", "Security", "Execution"},
		RequireTraceability: true,
	})

	expectGateReason(t, report, "Security evidence fails release policy.")
}

func TestEvaluateReleaseFailsWhenGovernanceRequiredAndMissing(t *testing.T) {
	report := EvaluateRelease(validEngineResult(), config.ReleaseGateConfig{
		MinPrimaryScore:     60,
		RequiredCategories:  []string{"Intent", "Behavior", "Assurance", "Security", "Execution"},
		RequireTraceability: true,
		RequireGovernance:   true,
	})

	expectGateReason(t, report, "Governance evidence is required but not implemented or not assessed.")
}

func TestEvaluateReleaseFailsWhenGovernanceFails(t *testing.T) {
	result := validEngineResult()
	result.Results = append(result.Results, models.ValidationResult{
		Capability: "Governance",
		Status:     models.StatusFail,
		Message:    "approval evidence failed",
	})

	report := EvaluateRelease(result, config.ReleaseGateConfig{
		MinPrimaryScore:     60,
		RequiredCategories:  []string{"Intent", "Behavior", "Assurance", "Security", "Execution"},
		RequireTraceability: true,
		RequireGovernance:   true,
	})

	expectGateReason(t, report, "Governance evidence fails release policy.")
}

func TestRenderReleaseGateOutput(t *testing.T) {
	report := Result{
		Name:              ReleaseGateName,
		Status:            StatusFail,
		PrimaryBottleneck: "Assurance",
		PrimaryScore:      15,
		RequiredScore:     75,
		Reasons:           []string{"Primary bottleneck score is below release threshold."},
		RecommendedAction: "Fix failing tests.",
	}

	text := RenderText(report)
	for _, substring := range []string{
		"Release Gate: FAIL",
		"Primary Bottleneck: Assurance",
		"Primary Score: 15",
		"Required Score: 75",
		"1. Primary bottleneck score is below release threshold.",
		"Recommended next action:",
	} {
		if !strings.Contains(text, substring) {
			t.Fatalf("expected %q in text output:\n%s", substring, text)
		}
	}

	markdown := RenderMarkdown(report)
	for _, substring := range []string{
		"## Bottleneck Release Gate",
		"| Gate | release |",
		"| Result | FAIL |",
		"### Gate Reasons",
		"### Recommended Next Action",
	} {
		if !strings.Contains(markdown, substring) {
			t.Fatalf("expected %q in markdown output:\n%s", substring, markdown)
		}
	}

	github := RenderGitHub(report)
	if !strings.Contains(github, "::error::Bottleneck release gate failed") ||
		!strings.Contains(github, "::error::Primary bottleneck score is below release threshold.") {
		t.Fatalf("expected GitHub release gate annotations, got:\n%s", github)
	}
}

func expectGateReason(t *testing.T, report Result, reason string) {
	t.Helper()
	if report.Status != StatusFail {
		t.Fatalf("expected release gate fail, got %#v", report)
	}
	for _, actual := range report.Reasons {
		if actual == reason {
			return
		}
	}
	t.Fatalf("expected reason %q, got %#v", reason, report.Reasons)
}

func validEngineResult() models.EngineResult {
	return models.EngineResult{
		Environment:       "production",
		SystemStatus:      models.StatusPass,
		PrimaryBottleneck: "None",
		Results: []models.ValidationResult{
			{Capability: "Behavior", Status: models.StatusPass, Details: []string{"behavior evidence"}},
			{Capability: "Intent", Status: models.StatusPass, Details: []string{"intent evidence"}},
			{Capability: "Design", Status: models.StatusPass, Details: []string{"design evidence"}},
			{Capability: "Assurance", Status: models.StatusPass, Details: []string{"accuracy: 1.00"}},
			{Capability: "Security", Status: models.StatusPass, Details: []string{"violations: 0"}},
			{Capability: "Execution", Status: models.StatusPass, Details: []string{"error_rate: 0.01"}},
			{Capability: "Traceability", Status: models.StatusPass, Details: []string{"traceability_nodes: 6"}},
		},
	}
}

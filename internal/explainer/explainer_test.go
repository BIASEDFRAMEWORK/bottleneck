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

func TestExplainIncludesEvidenceFoundMissingRiskRecommendationsOwnersAndAutomation(t *testing.T) {
	result := models.EngineResult{
		Environment:       "default",
		SystemStatus:      models.StatusWarning,
		PrimaryBottleneck: "Assurance",
		Results: []models.ValidationResult{{
			Capability: "Assurance",
			Status:     models.StatusWarning,
			Message:    "BEHAVIOR-003 has no mapped test evidence",
			Details: []string{
				"accuracy: 1.00 (threshold: 0.90)",
				"bottleneck/behavior/behavior-spec.md BEHAVIOR-003 has no mapped test evidence",
			},
		}},
	}

	output, err := Render(result, "")
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	expected := []string{
		"Bottleneck Explanation",
		"Why this matters:",
		"Evidence found:",
		"accuracy: 1.00 (threshold: 0.90)",
		"Evidence missing:",
		"Tests exist but are not linked to behavior IDs.",
		"Risk to delivery:",
		"The team may ship functionality that appears complete but cannot be proven against intended behavior.",
		"Recommended actions:",
		"Add traceability references from test evidence to behavior expectations.",
		"Suggested owner roles:",
		"QA/Assurance Engineer",
		"Suggested automation:",
		"Run Cucumber or equivalent behavior tests in GitHub Actions.",
	}
	for _, substring := range expected {
		if !strings.Contains(output, substring) {
			t.Fatalf("expected %q in output:\n%s", substring, output)
		}
	}
}

func TestExplainMarkdownOutput(t *testing.T) {
	report, err := BuildReport(singleCategoryResult("Security", models.StatusFail, "high severity finding exists"), nil, "security")
	if err != nil {
		t.Fatalf("BuildReport returned error: %v", err)
	}
	output, err := RenderReport(report, FormatMarkdown)
	if err != nil {
		t.Fatalf("RenderReport returned error: %v", err)
	}
	expected := []string{
		"# Bottleneck Explanation",
		"## Security",
		"### Why This Matters",
		"### Evidence Found",
		"### Evidence Missing",
		"### Risk To Delivery",
		"### Recommended Actions",
		"### Suggested Owner Roles",
		"### Suggested Automation",
	}
	for _, substring := range expected {
		if !strings.Contains(output, substring) {
			t.Fatalf("expected %q in markdown output:\n%s", substring, output)
		}
	}
}

func TestExplainJSONOutput(t *testing.T) {
	report, err := BuildReport(singleCategoryResult("Execution", models.StatusWarning, "adoption below threshold"), nil, "execution")
	if err != nil {
		t.Fatalf("BuildReport returned error: %v", err)
	}
	output, err := RenderReport(report, FormatJSON)
	if err != nil {
		t.Fatalf("RenderReport returned error: %v", err)
	}
	if !strings.Contains(output, `"schema_version": "explain.v2"`) ||
		!strings.Contains(output, `"risk_to_delivery"`) ||
		!strings.Contains(output, `"suggested_automations"`) {
		t.Fatalf("expected structured explain JSON, got:\n%s", output)
	}
}

func TestExplainIntentRules(t *testing.T) {
	output := renderRuleOutput(t, "Intent", "intent evidence has placeholder content and lacks measurable outcomes")
	expected := []string{
		"Intent exists but does not clearly define measurable outcomes.",
		"Intent contains placeholder or thin content.",
		"Add observable outcomes, business constraints, and unacceptable outcomes.",
		"Replace template text with product-specific intent.",
		"Product Lead",
		"PR template requiring an intent reference",
	}
	assertExplainContains(t, output, expected)
}

func TestExplainBehaviorRules(t *testing.T) {
	output := renderRuleOutput(t, "Behavior", "behavior spec has no BEHAVIOR-* IDs and no mapped test evidence")
	expected := []string{
		"Behavior expectations are not traceable.",
		"Behavior is not validated.",
		"Add stable behavior IDs such as BEHAVIOR-001.",
		"Map each critical behavior to test evidence.",
		"QA/Assurance Engineer",
		"behavior specification linting",
	}
	assertExplainContains(t, output, expected)
}

func TestExplainDesignRules(t *testing.T) {
	output := renderRuleOutput(t, "Design", "architecture is missing tradeoffs and failure modes")
	expected := []string{
		"Architecture exists but does not explain tradeoffs.",
		"Architecture does not describe failure modes.",
		"Add decision records for key constraints and design choices.",
		"Add fallback, monitoring, and operational assumptions.",
		"Architect",
		"architecture decision record check",
	}
	assertExplainContains(t, output, expected)
}

func TestExplainAssuranceRules(t *testing.T) {
	output := renderRuleOutput(t, "Assurance", "missing assurance evidence, no mapped test evidence, and critical behavior validation coverage is low")
	expected := []string{
		"Missing automated validation evidence.",
		"Tests exist but are not linked to behavior IDs.",
		"Critical behaviors lack validation.",
		"Add test output or BDD evidence under bottleneck/assurance/.",
		"Prioritize tests for high-risk behaviors before expanding feature scope.",
		"Ingest test results",
	}
	assertExplainContains(t, output, expected)
}

func TestExplainSecurityRules(t *testing.T) {
	output := renderRuleOutput(t, "Security", "missing security guardrails and high severity security finding")
	expected := []string{
		"Missing security evidence.",
		"Security guardrails are not documented.",
		"High severity security findings exist.",
		"Add CodeQL, dependency review, secret scanning, or SARIF evidence.",
		"Block release until findings are triaged or resolved.",
		"Security Engineer",
		"CodeQL",
	}
	assertExplainContains(t, output, expected)
}

func TestExplainExecutionRules(t *testing.T) {
	output := renderRuleOutput(t, "Execution", "missing telemetry, weak adoption, user override, error rate, and latency are above threshold")
	expected := []string{
		"Missing execution evidence.",
		"Execution evidence suggests weak adoption or user trust.",
		"Execution evidence suggests operational instability.",
		"Add telemetry or production-readiness evidence.",
		"Review user workflow, training, and feedback loops.",
		"Address error rate, latency, or incident signals before accelerating release.",
		"SRE/Operations",
		"Ingest telemetry JSON",
	}
	assertExplainContains(t, output, expected)
}

func singleCategoryResult(category string, status string, message string) models.EngineResult {
	return models.EngineResult{
		Environment:       "default",
		SystemStatus:      status,
		PrimaryBottleneck: category,
		Results: []models.ValidationResult{{
			Capability: category,
			Status:     status,
			Message:    message,
		}},
	}
}

func renderRuleOutput(t *testing.T, category string, message string) string {
	t.Helper()
	output, err := Render(singleCategoryResult(category, models.StatusWarning, message), category)
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}
	return output
}

func assertExplainContains(t *testing.T, output string, expected []string) {
	t.Helper()
	for _, substring := range expected {
		if !strings.Contains(output, substring) {
			t.Fatalf("expected %q in output:\n%s", substring, output)
		}
	}
}

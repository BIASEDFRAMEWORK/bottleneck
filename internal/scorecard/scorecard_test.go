package scorecard

import (
	"encoding/json"
	"strings"
	"testing"

	"bottleneck/internal/githubactions"
	"bottleneck/internal/models"
	"bottleneck/internal/prrisk"
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
	if card.Diagnosis.PrimaryBottleneck != "Assurance" {
		t.Fatalf("expected diagnosis primary bottleneck Assurance, got %q", card.Diagnosis.PrimaryBottleneck)
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
		"Bottleneck Scorecard",
		"Environment: production",
		"Release Recommendation: Block",
		"Primary Bottleneck: Assurance",
		"Category Results:",
		"- Assurance: Fail",
		"Why:",
		"Next Action:",
		"Fix failing tests or add passing assurance evidence until accuracy meets the selected threshold.",
	}

	for _, substring := range expected {
		if !strings.Contains(output, substring) {
			t.Fatalf("expected %q in output:\n%s", substring, output)
		}
	}
	for _, verboseSection := range []string{"Effective Thresholds:", "Capability Details:", "Score Impacts:"} {
		if verboseSection == "Effective Thresholds:" {
			continue
		}
		if strings.Contains(output, verboseSection) {
			t.Fatalf("default scorecard output should omit %q:\n%s", verboseSection, output)
		}
	}
}

func TestGaugeHelperRendersAndClampsScores(t *testing.T) {
	tests := []struct {
		score    int
		expected string
	}{
		{score: -10, expected: "[----------]"},
		{score: 0, expected: "[----------]"},
		{score: 20, expected: "[##--------]"},
		{score: 60, expected: "[######----]"},
		{score: 80, expected: "[########--]"},
		{score: 100, expected: "[##########]"},
		{score: 120, expected: "[##########]"},
	}

	for _, tt := range tests {
		if actual := gauge(tt.score, 10); actual != tt.expected {
			t.Fatalf("expected gauge(%d) %q, got %q", tt.score, tt.expected, actual)
		}
	}
}

func TestRenderTextIncludesCategoryGaugeAndWeakestMarker(t *testing.T) {
	output, err := Render(sampleWarningResult(), "text")
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	expected := []string{
		"Category Results:",
		"- Behavior: Warn",
		"Primary Bottleneck: Behavior",
		"Why:",
		"Next Action:",
	}
	for _, substring := range expected {
		if !strings.Contains(output, substring) {
			t.Fatalf("expected %q in output:\n%s", substring, output)
		}
	}
	for _, verboseMarker := range []string{"Category Gauges:", "<-- primary bottleneck", "Capability Details:"} {
		if strings.Contains(output, verboseMarker) {
			t.Fatalf("default scorecard output should not include detailed marker %q:\n%s", verboseMarker, output)
		}
	}
}

func TestRenderDefaultSaaSScorecardIsConciseMainSurface(t *testing.T) {
	output, err := Render(sampleSaaSDayOneResult(), "text")
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	expected := []string{
		"Bottleneck Scorecard",
		"Environment: dev",
		"Release Recommendation: Conditional",
		"Primary Bottleneck: Assurance",
		"Effective Thresholds:",
		"- Minimum score: 65",
		"- Required traceability: false",
		"- Critical security findings allowed: 1",
		"- Stale telemetry allowed: true",
		"Category Results:",
		"- Intent: Pass",
		"- Behavior: Pass",
		"- Design: Pass",
		"- Assurance: Warn",
		"- Security: Pass",
		"- Execution: Pass",
		"Why:",
		"BEHAVIOR-003 payment retry behavior has no mapped test evidence.",
		"Next Action:",
		"Add assurance evidence for payment retry behavior. Map it to BEHAVIOR-003.",
	}
	for _, substring := range expected {
		if !strings.Contains(output, substring) {
			t.Fatalf("expected %q in concise SaaS scorecard:\n%s", substring, output)
		}
	}

	assertInOrder(t, output, []string{
		"- Intent: Pass",
		"- Behavior: Pass",
		"- Design: Pass",
		"- Assurance: Warn",
		"- Security: Pass",
		"- Execution: Pass",
	})

	lineCount := len(strings.Split(strings.TrimSpace(output), "\n"))
	if lineCount > 30 {
		t.Fatalf("expected concise scorecard to stay scan-friendly, got %d lines:\n%s", lineCount, output)
	}
	for _, verboseSection := range []string{"Effective Thresholds:", "Capability Details:", "Evidence:", "Score Impacts:"} {
		if verboseSection == "Effective Thresholds:" {
			continue
		}
		if strings.Contains(output, verboseSection) {
			t.Fatalf("concise scorecard should omit %q:\n%s", verboseSection, output)
		}
	}
}

func TestRenderJSONProducesValidJSON(t *testing.T) {
	result := models.EngineResult{
		Environment:         "default",
		SystemStatus:        models.StatusPass,
		PrimaryBottleneck:   "None",
		EffectiveThresholds: sampleThresholds(),
		Results: []models.ValidationResult{
			{
				Capability: "Intent",
				Status:     models.StatusPass,
				Details:    []string{"bottleneck/intent/intent.md contains required intent evidence"},
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
	if decoded.SchemaVersion != SchemaVersion {
		t.Fatalf("expected schema version %q, got %q", SchemaVersion, decoded.SchemaVersion)
	}
	if decoded.ReleaseRecommendation != RecommendationProceed {
		t.Fatalf("expected release recommendation Proceed, got %q", decoded.ReleaseRecommendation)
	}
	if decoded.Diagnosis.PrimaryBottleneck != "None" {
		t.Fatalf("expected healthy diagnosis in json, got %q", decoded.Diagnosis.PrimaryBottleneck)
	}
	if decoded.EffectiveThresholds.Assurance.MinAccuracy != 0.95 {
		t.Fatalf("expected min accuracy threshold in json, got %.2f", decoded.EffectiveThresholds.Assurance.MinAccuracy)
	}
	if decoded.EffectiveThresholds.Security.SARIF.MaxMedium != 5 {
		t.Fatalf("expected SARIF medium threshold in json, got %d", decoded.EffectiveThresholds.Security.SARIF.MaxMedium)
	}
	if decoded.EffectiveThresholds.Execution.Telemetry.MaxAgeHours != 168 {
		t.Fatalf("expected telemetry freshness threshold in json, got %d", decoded.EffectiveThresholds.Execution.Telemetry.MaxAgeHours)
	}
	if decoded.EffectiveThresholds.Gate.Release.MinPrimaryScore != 85 {
		t.Fatalf("expected release gate minimum score in json, got %d", decoded.EffectiveThresholds.Gate.Release.MinPrimaryScore)
	}
	if !decoded.EffectiveThresholds.Gate.Release.RequireTraceability {
		t.Fatal("expected release gate traceability threshold in json")
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
	if decoded.Capabilities[0].EvidenceCount != 1 {
		t.Fatalf("expected evidence count 1, got %d", decoded.Capabilities[0].EvidenceCount)
	}
	if decoded.Capabilities[0].Score < 80 {
		t.Fatalf("expected healthy capability score, got %d", decoded.Capabilities[0].Score)
	}
}

func TestScorecardJSONIncludesSchemaVersion(t *testing.T) {
	decoded := renderStableScorecardJSON(t, sampleWarningResult())
	if decoded.SchemaVersion != SchemaVersion {
		t.Fatalf("expected schema version %q, got %q", SchemaVersion, decoded.SchemaVersion)
	}
}

func TestScorecardJSONIncludesEnvironment(t *testing.T) {
	decoded := renderStableScorecardJSON(t, sampleWarningResult())
	if decoded.Environment != "production" {
		t.Fatalf("expected environment production, got %q", decoded.Environment)
	}
}

func TestScorecardJSONIncludesPrimaryBottleneck(t *testing.T) {
	decoded := renderStableScorecardJSON(t, sampleWarningResult())
	if decoded.PrimaryBottleneck != "Behavior" {
		t.Fatalf("expected primary bottleneck Behavior, got %q", decoded.PrimaryBottleneck)
	}
	if decoded.SystemStatus != StatusWarn || decoded.ReleaseRecommendation != RecommendationConditional {
		t.Fatalf("expected status and recommendation in stable contract, got %#v", decoded)
	}
}

func TestScorecardJSONIncludesCategories(t *testing.T) {
	decoded := renderStableScorecardJSON(t, sampleSaaSDayOneResult())
	if len(decoded.Categories) != 7 {
		t.Fatalf("expected stable category objects, got %#v", decoded.Categories)
	}
	expectedOrder := []string{"Intent", "Behavior", "Design", "Assurance", "Security", "Execution", "Traceability"}
	for index, expected := range expectedOrder {
		if decoded.Categories[index].Name != expected {
			t.Fatalf("expected category %d to be %s, got %#v", index, expected, decoded.Categories)
		}
	}
}

func TestScorecardJSONCategoryFields(t *testing.T) {
	missing := "Add behavior examples for failed payment retries."
	recommendation := "Replace placeholder behavior text with concrete expected and unacceptable behavior."
	result := models.EngineResult{
		Environment:       "default",
		SystemStatus:      models.StatusWarning,
		PrimaryBottleneck: "Behavior",
		Results: []models.ValidationResult{{
			Capability: "Behavior",
			Status:     models.StatusWarning,
			Details:    []string{"behavior evidence exists"},
			EvidenceQuality: models.EvidenceQuality{
				Missing: []string{missing},
			},
		}},
	}

	decoded := renderStableScorecardJSON(t, result)
	if len(decoded.Categories) != 1 {
		t.Fatalf("expected one category, got %#v", decoded.Categories)
	}
	category := decoded.Categories[0]
	if category.Name != "Behavior" ||
		category.Status != StatusWarn ||
		category.Score == 0 ||
		strings.TrimSpace(category.Summary) == "" ||
		!containsString(category.EvidenceFound, "behavior evidence exists") ||
		!containsString(category.EvidenceMissing, missing) ||
		!containsString(category.Recommendations, recommendation) {
		t.Fatalf("stable category contract missing required fields: %#v", category)
	}
}

func TestRenderSurfacesContentQualityDetailsInTextAndJSON(t *testing.T) {
	detail := `bottleneck/intent/intent.md section "Outcomes" still contains placeholder content`
	result := models.EngineResult{
		Environment:       "default",
		SystemStatus:      models.StatusWarning,
		PrimaryBottleneck: "Intent",
		Results: []models.ValidationResult{{
			Capability: "Intent",
			Status:     models.StatusWarning,
			Message:    "content quality warnings detected",
			Details:    []string{detail},
		}},
	}

	textOutput, err := RenderWithOptions(result, "text", Options{Details: true})
	if err != nil {
		t.Fatalf("Render text returned error: %v", err)
	}
	if !strings.Contains(textOutput, "WARN") {
		t.Fatalf("expected warning status in text output:\n%s", textOutput)
	}
	if !strings.Contains(textOutput, detail) {
		t.Fatalf("expected content quality detail in text output:\n%s", textOutput)
	}

	jsonOutput, err := Render(result, "json")
	if err != nil {
		t.Fatalf("Render json returned error: %v", err)
	}

	var decoded Scorecard
	if err := json.Unmarshal([]byte(jsonOutput), &decoded); err != nil {
		t.Fatalf("expected valid json, got error: %v\n%s", err, jsonOutput)
	}
	if len(decoded.Capabilities) != 1 {
		t.Fatalf("expected one capability in json, got %d", len(decoded.Capabilities))
	}
	if !containsString(decoded.Capabilities[0].Evidence, detail) {
		t.Fatalf("expected content quality detail in json evidence, got %#v", decoded.Capabilities[0].Evidence)
	}
	if !containsString(decoded.Capabilities[0].MissingEvidence, `bottleneck/intent/intent.md section "Outcomes" needs real evidence`) {
		t.Fatalf("expected content quality detail in json missing evidence, got %#v", decoded.Capabilities[0].MissingEvidence)
	}
}

func TestScorecardIncludesEvidenceQualityMissingAndScoreImpacts(t *testing.T) {
	missing := "Add an INTENT-* heading such as ### INTENT-001: ..."
	impact := models.ScoreImpact{
		Reason: "bottleneck/intent/intent.md does not define an INTENT-* evidence ID",
		Delta:  -20,
	}
	result := models.EngineResult{
		Environment:       "default",
		SystemStatus:      models.StatusWarning,
		PrimaryBottleneck: "Intent",
		Results: []models.ValidationResult{{
			Capability: "Intent",
			Status:     models.StatusWarning,
			Message:    "intent evidence quality is weak",
			EvidenceQuality: models.EvidenceQuality{
				Score:        80,
				Missing:      []string{missing},
				ScoreImpacts: []models.ScoreImpact{impact},
			},
		}},
	}

	card := Build(result)
	capability := card.Capabilities[0]
	if !containsString(capability.MissingEvidence, missing) {
		t.Fatalf("expected evidence-quality missing evidence, got %#v", capability.MissingEvidence)
	}
	if len(capability.ScoreImpacts) != 1 || capability.ScoreImpacts[0] != impact {
		t.Fatalf("expected score impact in scorecard, got %#v", capability.ScoreImpacts)
	}
}

func TestScorecardDiagnosisSelectsMissingAssuranceEvidence(t *testing.T) {
	result := models.EngineResult{
		Environment:       "production",
		SystemStatus:      models.StatusFail,
		PrimaryBottleneck: "Assurance",
		Results: []models.ValidationResult{
			{Capability: "Behavior", Status: models.StatusPass, Details: []string{"behavior evidence"}},
			{Capability: "Intent", Status: models.StatusPass, Details: []string{"intent evidence"}},
			{Capability: "Design", Status: models.StatusPass, Details: []string{"design evidence"}},
			{Capability: "Assurance", Status: models.StatusFail, Message: "missing results.json"},
			{Capability: "Security", Status: models.StatusPass, Details: []string{"violations: 0"}},
			{Capability: "Execution", Status: models.StatusPass, Details: []string{"error_rate: 0.01"}},
		},
	}

	card := Build(result)

	if card.Diagnosis.PrimaryBottleneck != "Assurance" {
		t.Fatalf("expected Assurance diagnosis, got %q", card.Diagnosis.PrimaryBottleneck)
	}
	if !strings.Contains(card.Diagnosis.WhyItMatters, "proof that the expected behavior was tested") {
		t.Fatalf("expected assurance why text, got %q", card.Diagnosis.WhyItMatters)
	}
	if card.Diagnosis.RecommendedAction != "Add assurance evidence that maps test or evaluation results to BEHAVIOR-001." {
		t.Fatalf("unexpected recommended action %q", card.Diagnosis.RecommendedAction)
	}
	if score := diagnosisScoreFor(card, "Assurance"); score >= diagnosisScoreFor(card, "Security") {
		t.Fatalf("expected Assurance score below Security, got Assurance=%d Security=%d", score, diagnosisScoreFor(card, "Security"))
	}
}

func TestScorecardDiagnosisHandlesTiesDeterministically(t *testing.T) {
	result := models.EngineResult{
		Environment:       "production",
		SystemStatus:      models.StatusFail,
		PrimaryBottleneck: "Security",
		Results: []models.ValidationResult{
			{Capability: "Behavior", Status: models.StatusPass},
			{Capability: "Intent", Status: models.StatusPass},
			{Capability: "Design", Status: models.StatusPass},
			{Capability: "Assurance", Status: models.StatusFail, Message: "missing results.json"},
			{Capability: "Security", Status: models.StatusFail, Message: "missing guardrails.json"},
			{Capability: "Execution", Status: models.StatusPass},
		},
	}

	card := Build(result)

	if card.Diagnosis.PrimaryBottleneck != "Assurance" {
		t.Fatalf("expected Assurance to win tie, got %q", card.Diagnosis.PrimaryBottleneck)
	}
	if got := strings.Join(card.Diagnosis.TiedBottlenecks, ","); got != "Assurance,Security" {
		t.Fatalf("expected deterministic tied bottlenecks, got %q", got)
	}
}

func TestReleaseRecommendationForPassWarningAndFailure(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		results  []models.ValidationResult
		expected string
	}{
		{
			name:   "proceed",
			status: models.StatusPass,
			results: []models.ValidationResult{{
				Capability: "Assurance",
				Status:     models.StatusPass,
				Details:    []string{"accuracy: 1.00 (threshold: 0.95)"},
			}},
			expected: RecommendationProceed,
		},
		{
			name:   "conditional",
			status: models.StatusWarning,
			results: []models.ValidationResult{{
				Capability: "Execution",
				Status:     models.StatusWarning,
				Message:    "low adoption",
				Details:    []string{"adoption_rate: 0.40 (min: 0.50)"},
			}},
			expected: RecommendationConditional,
		},
		{
			name:   "block",
			status: models.StatusFail,
			results: []models.ValidationResult{{
				Capability: "Assurance",
				Status:     models.StatusFail,
				Message:    "accuracy below threshold",
			}},
			expected: RecommendationBlock,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := Build(models.EngineResult{
				Environment:       "production",
				SystemStatus:      tt.status,
				PrimaryBottleneck: "Assurance",
				Results:           tt.results,
			})

			if card.ReleaseRecommendation != tt.expected {
				t.Fatalf("expected recommendation %q, got %q", tt.expected, card.ReleaseRecommendation)
			}
		})
	}
}

func TestCapabilityEntriesIncludeEvidenceDepthFields(t *testing.T) {
	result := models.EngineResult{
		Environment:       "default",
		SystemStatus:      models.StatusFail,
		PrimaryBottleneck: "Behavior",
		Results: []models.ValidationResult{{
			Capability: "Behavior",
			Status:     models.StatusFail,
			Message:    "missing behavior-spec.md",
		}},
	}

	card := Build(result)
	capability := card.Capabilities[0]

	if capability.EvidenceCount != 1 {
		t.Fatalf("expected message-backed evidence count 1, got %d", capability.EvidenceCount)
	}
	if capability.Reason != "missing behavior-spec.md" {
		t.Fatalf("expected reason from validation message, got %q", capability.Reason)
	}
	if capability.RecommendedAction == "" {
		t.Fatal("expected recommended action")
	}
	if !containsString(capability.MissingEvidence, "Create bottleneck/behavior/behavior-spec.md with Expected Behavior and Unacceptable Behavior evidence.") {
		t.Fatalf("expected actionable missing evidence, got %#v", capability.MissingEvidence)
	}
}

func TestRenderTextIncludesThresholdsAndReleaseRecommendation(t *testing.T) {
	output, err := RenderWithOptions(sampleWarningResult(), "text", Options{Details: true})
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	expected := []string{
		"Release Recommendation: Conditional",
		"Effective Thresholds:",
		"Minimum score: 85",
		"Required traceability: true",
		"Critical security findings allowed: 0",
		"Stale telemetry allowed: false",
		"gate.release.min_primary_score: 85",
		"assurance.min_accuracy: 0.95",
		"assurance.max_failures: 0",
		"execution.max_error_rate: 0.05",
		"execution.min_adoption: 0.50",
		"execution.telemetry.max_age_hours: 168",
		"execution.telemetry.min_deployments_per_week: 1.00",
		"security.sarif.max_high: 0",
		"security.sarif.max_medium: 5",
	}

	for _, substring := range expected {
		if !strings.Contains(output, substring) {
			t.Fatalf("expected %q in output:\n%s", substring, output)
		}
	}
}

func TestRenderMarkdownProducesGitHubReadableTable(t *testing.T) {
	output, err := Render(sampleWarningResult(), "markdown")
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	expected := []string{
		"# bottleneck Scorecard",
		"| Field | Value |",
		"## Effective Thresholds",
		"| Minimum score | 85 |",
		"| Required traceability | true |",
		"## Diagnosis",
		"| Capability | Status | Score | Evidence | Missing Evidence | Recommendation |",
		"| Behavior | WARN | 50 | 1 |",
	}

	for _, substring := range expected {
		if !strings.Contains(output, substring) {
			t.Fatalf("expected %q in markdown output:\n%s", substring, output)
		}
	}
}

func TestRenderJSONIncludesGitHubMetadataWhenDetected(t *testing.T) {
	changedFiles := 30
	additions := 800
	deletions := 250
	draft := false
	result := sampleWarningResult()
	output, err := RenderWithOptions(result, "json", Options{
		View: ViewGovernance,
		GitHub: &githubactions.Metadata{
			Detected:   true,
			EventName:  "pull_request",
			Repository: "acme/widgets",
			RunID:      "123456",
			PullRequest: &githubactions.PullRequestMetadata{
				Number:             42,
				Title:              "Add release gate",
				URL:                "https://github.com/acme/widgets/pull/42",
				BaseRef:            "main",
				HeadRef:            "feature/release-gate",
				Labels:             []string{"ai-assisted"},
				RequestedReviewers: []string{},
				ChangedFiles:       &changedFiles,
				Additions:          &additions,
				Deletions:          &deletions,
				Draft:              &draft,
				ChangedFileNames:   []string{"cmd/scorecard.go"},
			},
		},
	})
	if err != nil {
		t.Fatalf("RenderWithOptions returned error: %v", err)
	}

	var decoded Scorecard
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("expected valid json, got error: %v\n%s", err, output)
	}
	if decoded.GitHub == nil || decoded.GitHub.PullRequest == nil {
		t.Fatalf("expected GitHub metadata in scorecard JSON: %#v", decoded.GitHub)
	}
	if decoded.GitHub.PullRequest.Number != 42 {
		t.Fatalf("expected PR number 42, got %d", decoded.GitHub.PullRequest.Number)
	}
	if !containsRiskSignal(decoded.PullRequestRisk, "large_changed_file_count") {
		t.Fatalf("expected large PR risk signal, got %#v", decoded.PullRequestRisk)
	}
	if !containsRiskSignal(decoded.PullRequestRisk, "source_without_evidence_artifacts") {
		t.Fatalf("expected source without artifact risk signal, got %#v", decoded.PullRequestRisk)
	}
	if decoded.Diagnosis.PrimaryBottleneck != "Behavior" {
		t.Fatalf("expected diagnosis primary bottleneck Behavior, got %q", decoded.Diagnosis.PrimaryBottleneck)
	}
}

func TestRenderMarkdownIncludesGitHubMetadataWhenDetected(t *testing.T) {
	changedFiles := 2
	result := sampleWarningResult()
	output, err := RenderWithOptions(result, "markdown", Options{
		View: ViewGovernance,
		GitHub: &githubactions.Metadata{
			Detected:   true,
			EventName:  "pull_request",
			Repository: "acme/widgets",
			RunID:      "123456",
			PullRequest: &githubactions.PullRequestMetadata{
				Number:           42,
				Title:            "Add release gate",
				BaseRef:          "main",
				HeadRef:          "feature/release-gate",
				ChangedFiles:     &changedFiles,
				ChangedFileNames: []string{"bottleneck/intent/intent.md"},
			},
		},
	})
	if err != nil {
		t.Fatalf("RenderWithOptions returned error: %v", err)
	}

	expected := []string{
		"## GitHub Pull Request Context",
		"| Pull Request | #42 Add release gate |",
		"## Pull Request Risk Signals",
	}
	for _, substring := range expected {
		if !strings.Contains(output, substring) {
			t.Fatalf("expected %q in markdown output:\n%s", substring, output)
		}
	}
}

func TestExecutiveViewOmitsDetailedEvidence(t *testing.T) {
	detail := `bottleneck/behavior/behavior-spec.md section "Expected Behavior" still contains placeholder content`
	result := sampleWarningResult()
	result.Results[0].Details = []string{detail}

	output, err := Render(result, "text", ViewExecutive)
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	expected := []string{
		"Executive View",
		"Release Recommendation: Conditional",
		"Capability Status Summary:",
	}
	for _, substring := range expected {
		if !strings.Contains(output, substring) {
			t.Fatalf("expected %q in output:\n%s", substring, output)
		}
	}
	if strings.Contains(output, detail) {
		t.Fatalf("executive view should omit detailed evidence:\n%s", output)
	}
}

func TestEngineeringViewIncludesDetailedEvidenceAndMissingEvidence(t *testing.T) {
	detail := `bottleneck/behavior/behavior-spec.md section "Expected Behavior" still contains placeholder content`
	result := sampleWarningResult()
	result.Results[0].Details = []string{detail}

	output, err := RenderWithOptions(result, "text", Options{View: ViewEngineering, Details: true})
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	expected := []string{
		"Capability Details:",
		detail,
		`bottleneck/behavior/behavior-spec.md section "Expected Behavior" needs real evidence`,
	}
	for _, substring := range expected {
		if !strings.Contains(output, substring) {
			t.Fatalf("expected %q in output:\n%s", substring, output)
		}
	}
}

func TestGovernanceViewCallsOutGovernanceSignalsAndMissingGovernanceEvidence(t *testing.T) {
	result := models.EngineResult{
		Environment:         "production",
		SystemStatus:        models.StatusPass,
		PrimaryBottleneck:   "None",
		EffectiveThresholds: sampleThresholds(),
		Results: []models.ValidationResult{
			{Capability: "Security", Status: models.StatusPass, Details: []string{"violations: 0 (allowed: 0)"}},
			{Capability: "Assurance", Status: models.StatusPass, Details: []string{"accuracy: 1.00 (threshold: 0.95)"}},
			{Capability: "Execution", Status: models.StatusPass, Details: []string{"error_rate: 0.01 (max: 0.05)"}},
		},
	}

	output, err := Render(result, "text", ViewGovernance)
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	expected := []string{
		"Governance View",
		"Security: PASS",
		"Assurance: PASS",
		"Execution: PASS",
		"Governance Evidence: not assessed (no governance artifact exists yet)",
		"Governance evidence not assessed: no governance artifact exists yet.",
	}
	for _, substring := range expected {
		if !strings.Contains(output, substring) {
			t.Fatalf("expected %q in output:\n%s", substring, output)
		}
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

func TestRenderRejectsUnsupportedView(t *testing.T) {
	_, err := Render(models.EngineResult{}, "text", "audit")
	if err == nil {
		t.Fatal("expected error for unsupported view")
	}
	if !strings.Contains(err.Error(), "unsupported view") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func sampleWarningResult() models.EngineResult {
	return models.EngineResult{
		Environment:         "production",
		SystemStatus:        models.StatusWarning,
		PrimaryBottleneck:   "Behavior",
		EffectiveThresholds: sampleThresholds(),
		Results: []models.ValidationResult{{
			Capability: "Behavior",
			Status:     models.StatusWarning,
			Message:    "content quality warnings detected",
			Details: []string{
				`bottleneck/behavior/behavior-spec.md section "Expected Behavior" still contains placeholder content`,
			},
		}},
	}
}

func sampleSaaSDayOneResult() models.EngineResult {
	return models.EngineResult{
		Environment:         "dev",
		SystemStatus:        models.StatusWarning,
		PrimaryBottleneck:   "Traceability",
		EffectiveThresholds: sampleDevThresholds(),
		Results: []models.ValidationResult{
			{Capability: "Behavior", Status: models.StatusPass, Details: []string{"behavior evidence"}},
			{Capability: "Intent", Status: models.StatusPass, Details: []string{"intent evidence"}},
			{Capability: "Design", Status: models.StatusPass, Details: []string{"design evidence"}},
			{Capability: "Assurance", Status: models.StatusPass},
			{Capability: "Security", Status: models.StatusPass, Details: []string{"violations: 0"}},
			{Capability: "Execution", Status: models.StatusPass, Details: []string{"error_rate: 0.01", "adoption_rate: 0.68"}},
			{
				Capability: "Traceability",
				Status:     models.StatusWarning,
				Message:    "traceability warnings detected",
				Details:    []string{"bottleneck/behavior/behavior-spec.md BEHAVIOR-003 has no mapped test evidence"},
				EvidenceQuality: models.EvidenceQuality{
					Score: 75,
					ScoreImpacts: []models.ScoreImpact{{
						Reason: "bottleneck/behavior/behavior-spec.md BEHAVIOR-003 has no mapped test evidence",
						Delta:  -25,
					}},
				},
			},
		},
	}
}

func sampleDevThresholds() models.EffectiveThresholds {
	thresholds := sampleThresholds()
	thresholds.Execution.Telemetry.MaxAgeHours = 0
	thresholds.Execution.Telemetry.StaleAllowed = true
	thresholds.Security.SARIF.MaxCritical = 1
	thresholds.Gate.Release.MinPrimaryScore = 65
	thresholds.Gate.Release.RequireTraceability = false
	thresholds.Gate.Release.RequireGovernance = false
	return thresholds
}

func sampleThresholds() models.EffectiveThresholds {
	return models.EffectiveThresholds{
		Assurance: models.AssuranceThresholds{
			MinAccuracy: 0.95,
			MaxFailures: 0,
		},
		Execution: models.ExecutionThresholds{
			MaxErrorRate: 0.05,
			MinAdoption:  0.50,
			Telemetry: models.TelemetryThresholds{
				MaxAgeHours:           168,
				StaleAllowed:          false,
				MinDeploymentsPerWeek: 1,
				MaxChangeFailureRate:  0.15,
				MaxErrorRate:          0.05,
				MaxUserOverrideRate:   0.10,
				MinAdoptionRate:       0.50,
				MaxBudgetVariance:     0.20,
			},
		},
		Security: models.SecurityThresholds{
			SARIF: models.SARIFThresholds{
				MaxCritical:           0,
				MaxHigh:               0,
				MaxMedium:             5,
				MaxLow:                20,
				FailOnUnknownSeverity: false,
			},
		},
		Gate: models.GateThresholds{
			Release: models.ReleaseGateThresholds{
				MinPrimaryScore:     85,
				RequiredCategories:  []string{"Intent", "Behavior", "Assurance", "Security", "Execution"},
				RequireTraceability: true,
				RequireGovernance:   true,
			},
		},
	}
}

func renderStableScorecardJSON(t *testing.T, result models.EngineResult) Scorecard {
	t.Helper()
	output, err := Render(result, "json")
	if err != nil {
		t.Fatalf("Render json returned error: %v", err)
	}
	var decoded Scorecard
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("scorecard JSON did not parse: %v\n%s", err, output)
	}
	stable := StableJSONScorecard(decoded)
	if stable.SchemaVersion == "" ||
		stable.Environment == "" ||
		stable.SystemStatus == "" ||
		stable.ReleaseRecommendation == "" ||
		stable.PrimaryBottleneck == "" {
		t.Fatalf("stable scorecard contract missing required top-level fields: %#v", stable)
	}
	return decoded
}

func containsString(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func containsRiskSignal(values []prrisk.Signal, expected string) bool {
	for _, value := range values {
		if value.ID == expected {
			return true
		}
	}
	return false
}

func diagnosisScoreFor(card Scorecard, category string) int {
	for _, score := range card.Diagnosis.CategoryScores {
		if score.Category == category {
			return score.Score
		}
	}
	return 0
}

func assertInOrder(t *testing.T, output string, expected []string) {
	t.Helper()
	lastIndex := -1
	for _, substring := range expected {
		index := strings.Index(output, substring)
		if index == -1 {
			t.Fatalf("expected %q in output:\n%s", substring, output)
		}
		if index <= lastIndex {
			t.Fatalf("expected %q after prior category in output:\n%s", substring, output)
		}
		lastIndex = index
	}
}

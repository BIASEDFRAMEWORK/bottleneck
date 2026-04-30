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
		"bottleneck SDLC Scorecard",
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
	if decoded.EffectiveThresholds.Assurance.MinAccuracy != 0.95 {
		t.Fatalf("expected min accuracy threshold in json, got %.2f", decoded.EffectiveThresholds.Assurance.MinAccuracy)
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

	textOutput, err := Render(result, "text")
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
	output, err := Render(sampleWarningResult(), "text")
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	expected := []string{
		"Release Recommendation: Conditional",
		"Effective Thresholds:",
		"assurance.min_accuracy: 0.95",
		"assurance.max_failures: 0",
		"execution.max_error_rate: 0.05",
		"execution.min_adoption: 0.50",
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
		"| Capability | Status | Evidence | Missing Evidence | Recommendation |",
		"| Behavior | WARN | 1 |",
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

	output, err := Render(result, "text", ViewEngineering)
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

func sampleThresholds() models.EffectiveThresholds {
	return models.EffectiveThresholds{
		Assurance: models.AssuranceThresholds{
			MinAccuracy: 0.95,
			MaxFailures: 0,
		},
		Execution: models.ExecutionThresholds{
			MaxErrorRate: 0.05,
			MinAdoption:  0.50,
		},
	}
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

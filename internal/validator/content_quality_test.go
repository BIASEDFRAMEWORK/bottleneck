package validator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"bottleneck/internal/models"
)

const (
	templateBehavior = "# Behavior Specification\n\n## Expected Behavior\n\nDescribe intended system behavior.\n\n## Unacceptable Behavior\n\nDescribe behavior the system must prevent.\n"
	templateIntent   = "# Intent\n\n## Outcomes\n\nDescribe required outcomes.\n\n## Constraints\n\nDescribe system constraints.\n\n## Success Criteria\n\nDescribe measurable success criteria.\n"
	templateDesign   = "# Architecture\n\nDescribe system architecture.\n"

	validBehavior = "# Behavior Specification\n\n## Expected Behavior\n\n### BEHAVIOR-001: Block unsafe release\nRefs:\n- INTENT-001\n- ASSURANCE-001\n\nThe service returns a deterministic answer for every accepted workflow request and blocks release when assurance fails.\n\n## Unacceptable Behavior\n\n- Reject unsigned release requests before any deployment step starts.\n"
	validIntent   = "# Intent\n\n## Outcomes\n\n### INTENT-001: Reduce review time\nRefs:\n- BEHAVIOR-001\n\nThe platform reduces release review time by surfacing every failing capability.\n\n## Constraints\n\n- All validation rules remain deterministic and runnable without network access.\n\n## Success Criteria\n\n- At least 95% of release checks identify the primary bottleneck within 1 CLI run.\n"
	validDesign   = "# Architecture\n\n### DESIGN-001: Local deterministic validation\nRefs:\n- INTENT-001\n- BEHAVIOR-001\n\nThe CLI reads local evidence artifacts, applies deterministic validators, and renders the same result through validate, explain, and scorecard commands.\n"
)

func TestFreshTemplateArtifactsProduceWarningsByDefault(t *testing.T) {
	basePath := t.TempDir()
	writeValidationProject(t, basePath, withoutTraceEvidence(map[string]string{
		"bottleneck/behavior/behavior-spec.md": templateBehavior,
		"bottleneck/intent/intent.md":          templateIntent,
		"bottleneck/design/architecture.md":    templateDesign,
	}))

	result := NewEngine(basePath, "default").Validate()

	if result.SystemStatus != models.StatusWarning {
		t.Fatalf("expected system WARNING, got %q", result.SystemStatus)
	}
	if result.PrimaryBottleneck != "Behavior" {
		t.Fatalf("expected first warning capability as primary bottleneck, got %q", result.PrimaryBottleneck)
	}

	for _, capability := range []string{"Behavior", "Intent", "Design"} {
		check := resultForCapability(t, result, capability)
		if check.Status != models.StatusWarning {
			t.Fatalf("expected %s WARNING, got %q", capability, check.Status)
		}
	}
}

func TestFreshTemplateArtifactsProduceFailuresInStrictMode(t *testing.T) {
	basePath := t.TempDir()
	writeValidationProject(t, basePath, withoutTraceEvidence(map[string]string{
		"bottleneck/behavior/behavior-spec.md": templateBehavior,
		"bottleneck/intent/intent.md":          templateIntent,
		"bottleneck/design/architecture.md":    templateDesign,
	}))

	result := NewEngine(basePath, "default", WithStrictMode(true)).Validate()

	if result.SystemStatus != models.StatusFail {
		t.Fatalf("expected system FAIL, got %q", result.SystemStatus)
	}
	if result.PrimaryBottleneck != "Behavior" {
		t.Fatalf("expected first failing capability as primary bottleneck, got %q", result.PrimaryBottleneck)
	}

	for _, capability := range []string{"Behavior", "Intent", "Design"} {
		check := resultForCapability(t, result, capability)
		if check.Status != models.StatusFail {
			t.Fatalf("expected %s FAIL, got %q", capability, check.Status)
		}
		if check.Message != contentQualityStrictFailMessage {
			t.Fatalf("expected strict content-quality message for %s, got %q", capability, check.Message)
		}
	}
}

func TestPlaceholderWarningsIdentifyExactFileAndSection(t *testing.T) {
	basePath := t.TempDir()
	writeValidationProject(t, basePath, withoutTraceEvidence(map[string]string{
		"bottleneck/behavior/behavior-spec.md": templateBehavior,
		"bottleneck/intent/intent.md":          templateIntent,
		"bottleneck/design/architecture.md":    templateDesign,
	}))

	result := NewEngine(basePath, "default").Validate()

	expectedDetails := []string{
		`bottleneck/behavior/behavior-spec.md section "Expected Behavior" still contains placeholder content`,
		`bottleneck/behavior/behavior-spec.md section "Unacceptable Behavior" still contains placeholder content`,
		`bottleneck/intent/intent.md section "Outcomes" still contains placeholder content`,
		`bottleneck/intent/intent.md section "Constraints" still contains placeholder content`,
		`bottleneck/intent/intent.md section "Success Criteria" still contains placeholder content`,
		`bottleneck/design/architecture.md section "Architecture" still contains placeholder content`,
	}

	for _, detail := range expectedDetails {
		if !engineResultHasDetail(result, detail) {
			t.Fatalf("expected detail %q in result: %#v", detail, result.Results)
		}
	}
}

func TestPartiallyCompletedArtifactsWarnOnlyForUnchangedOrInsufficientSections(t *testing.T) {
	basePath := t.TempDir()
	writeValidationProject(t, basePath, map[string]string{
		"bottleneck/behavior/behavior-spec.md": "# Behavior Specification\n\n## Expected Behavior\n\n### BEHAVIOR-001: Store release evidence\nRefs:\n- INTENT-001\n- ASSURANCE-001\n\nThe system stores validated deployment evidence before rendering summaries.\n\n## Unacceptable Behavior\n\nDescribe behavior the system must prevent.\n",
		"bottleneck/intent/intent.md":          "# Intent\n\n## Outcomes\n\n### INTENT-001: Identify blocking evidence\nRefs:\n- BEHAVIOR-001\n\nThe CLI identifies blocking capability failures before release approval.\n\n## Constraints\n\nTBD\n\n## Success Criteria\n\n- At least 95% of warnings name the artifact section that needs real evidence.\n",
		"bottleneck/design/architecture.md":    validDesign,
	})

	result := NewEngine(basePath, "default").Validate()

	behavior := resultForCapability(t, result, "Behavior")
	if behavior.Status != models.StatusWarning {
		t.Fatalf("expected Behavior WARNING, got %q", behavior.Status)
	}
	assertOnlyDetails(t, behavior.Details, []string{
		`bottleneck/behavior/behavior-spec.md section "Unacceptable Behavior" still contains placeholder content`,
	})

	intent := resultForCapability(t, result, "Intent")
	if intent.Status != models.StatusWarning {
		t.Fatalf("expected Intent WARNING, got %q", intent.Status)
	}
	assertOnlyDetails(t, intent.Details, []string{
		`bottleneck/intent/intent.md section "Constraints" still contains placeholder content`,
	})

	design := resultForCapability(t, result, "Design")
	if design.Status != models.StatusPass {
		t.Fatalf("expected Design PASS, got %q with details %#v", design.Status, design.Details)
	}
}

func TestValidArtifactsPassContentQualityChecks(t *testing.T) {
	basePath := t.TempDir()
	writeValidationProject(t, basePath, nil)

	result := NewEngine(basePath, "default").Validate()

	if result.SystemStatus != models.StatusPass {
		t.Fatalf("expected system PASS, got %q with results %#v", result.SystemStatus, result.Results)
	}
	if result.PrimaryBottleneck != "None" {
		t.Fatalf("expected no primary bottleneck, got %q", result.PrimaryBottleneck)
	}

	for _, capability := range []string{"Behavior", "Intent", "Design"} {
		check := resultForCapability(t, result, capability)
		if check.Status != models.StatusPass {
			t.Fatalf("expected %s PASS, got %q with details %#v", capability, check.Status, check.Details)
		}
	}
}

func TestPlaceholderPhraseDetectionIncludesEpicTwoPhrases(t *testing.T) {
	phrases := []string{
		"Describe required outcomes",
		"Describe system constraints",
		"TODO",
		"TBD",
		"Add measurable success criteria",
	}

	for _, phrase := range phrases {
		if !containsPlaceholder("Before release: " + phrase) {
			t.Fatalf("expected placeholder phrase %q to be detected", phrase)
		}
	}
}

func TestMarkdownEvidenceQualityScoresDepth(t *testing.T) {
	requirements := []sectionContentRequirement{
		{section: "Outcomes", placeholder: placeholderRequiredOutcomes},
		{section: "Constraints", placeholder: placeholderSystemConstraints},
		{section: "Success Criteria", placeholder: placeholderSuccessCriteria},
	}

	empty := evaluateMarkdownEvidenceQuality("bottleneck", "intent/intent.md", "Intent", "", requirements)
	headerOnly := evaluateMarkdownEvidenceQuality("bottleneck", "intent/intent.md", "Intent", "# Intent\n\n## Outcomes\n\n## Constraints\n\n## Success Criteria\n", requirements)
	placeholder := evaluateMarkdownEvidenceQuality("bottleneck", "intent/intent.md", "Intent", templateIntent, requirements)
	missingID := evaluateMarkdownEvidenceQuality("bottleneck", "intent/intent.md", "Intent", "# Intent\n\n## Outcomes\n\nThe release review identifies every blocking capability.\n\n## Constraints\n\nValidation remains deterministic offline.\n\n## Success Criteria\n\n- At least 95% of warnings include an artifact path.\n", requirements)
	meaningful := evaluateMarkdownEvidenceQuality("bottleneck", "intent/intent.md", "Intent", validIntent, requirements)

	if empty.Score != 0 {
		t.Fatalf("expected empty file to score 0, got %d", empty.Score)
	}
	if headerOnly.Score <= empty.Score || headerOnly.Score >= placeholder.Score {
		t.Fatalf("expected header-only score between empty and placeholder, got empty=%d header=%d placeholder=%d", empty.Score, headerOnly.Score, placeholder.Score)
	}
	if placeholder.Score <= headerOnly.Score || placeholder.Score >= missingID.Score {
		t.Fatalf("expected placeholder score between header-only and missing-ID content, got header=%d placeholder=%d missingID=%d", headerOnly.Score, placeholder.Score, missingID.Score)
	}
	if missingID.Score >= meaningful.Score {
		t.Fatalf("expected missing ID to lower score below meaningful content, got missingID=%d meaningful=%d", missingID.Score, meaningful.Score)
	}
	if !containsString(placeholder.Details, "bottleneck/intent/intent.md is placeholder-heavy") {
		t.Fatalf("expected placeholder-heavy detail, got %#v", placeholder.Details)
	}
	if !containsString(missingID.Details, "bottleneck/intent/intent.md does not define an INTENT-* evidence ID") {
		t.Fatalf("expected missing ID detail, got %#v", missingID.Details)
	}
	if meaningful.Score != 100 {
		t.Fatalf("expected meaningful content to score 100, got %d", meaningful.Score)
	}
}

func TestVagueIntentLanguageLowersScore(t *testing.T) {
	requirements := []sectionContentRequirement{
		{section: "Outcomes", placeholder: placeholderRequiredOutcomes},
		{section: "Constraints", placeholder: placeholderSystemConstraints},
		{section: "Success Criteria", placeholder: placeholderSuccessCriteria},
	}

	vague := "# Intent\n\n## Outcomes\n\n### INTENT-001: Improve release quality\nRefs:\n- BEHAVIOR-001\n\nThe platform should improve release quality with better and easy workflows.\n\n## Constraints\n\nValidation remains deterministic offline.\n\n## Success Criteria\n\nOperators get better release feedback.\n"
	measurable := validIntent

	vagueQuality := evaluateMarkdownEvidenceQuality("bottleneck", "intent/intent.md", "Intent", vague, requirements)
	measurableQuality := evaluateMarkdownEvidenceQuality("bottleneck", "intent/intent.md", "Intent", measurable, requirements)

	if vagueQuality.Score >= measurableQuality.Score {
		t.Fatalf("expected vague intent score below measurable intent, got vague=%d measurable=%d", vagueQuality.Score, measurableQuality.Score)
	}
	if !containsString(vagueQuality.Details, `bottleneck/intent/intent.md section "Success Criteria" does not include measurable criteria`) {
		t.Fatalf("expected measurable criteria detail, got %#v", vagueQuality.Details)
	}
}

func TestEngineSystemStatusUsesWarningsWhenNoFailuresExist(t *testing.T) {
	basePath := t.TempDir()
	writeValidationProject(t, basePath, map[string]string{
		"bottleneck/intent/intent.md": "# Intent\n\n## Outcomes\n\n### INTENT-001: Placeholder outcome\nRefs:\n- BEHAVIOR-001\n\nDescribe required outcomes.\n\n## Constraints\n\nDescribe system constraints.\n\n## Success Criteria\n\nAdd measurable success criteria.\n",
	})

	result := NewEngine(basePath, "default").Validate()

	if result.SystemStatus != models.StatusWarning {
		t.Fatalf("expected system WARNING, got %q", result.SystemStatus)
	}
	if result.PrimaryBottleneck != "Intent" {
		t.Fatalf("expected Intent as primary warning bottleneck, got %q", result.PrimaryBottleneck)
	}
}

func writeValidationProject(t *testing.T, basePath string, overrides map[string]string) {
	t.Helper()

	files := map[string]string{
		"bottleneck/config.yaml": `environments:
  default:
    assurance:
      min_accuracy: 0.90
      max_failures: 0
    execution:
      max_error_rate: 0.05
      min_adoption: 0.5
`,
		"bottleneck/behavior/behavior-spec.md": validBehavior,
		"bottleneck/intent/intent.md":          validIntent,
		"bottleneck/design/architecture.md":    validDesign,
		"bottleneck/assurance/results.json":    "{\n  \"scenarios_total\": 1,\n  \"scenarios_passed\": 1,\n  \"scenarios_failed\": 0,\n  \"failures\": [],\n  \"evidence\": [{\"id\":\"ASSURANCE-001\",\"refs\":[\"BEHAVIOR-001\"],\"source\":\"cucumber\",\"status\":\"pass\"}]\n}\n",
		"bottleneck/security/guardrails.json":  "{\n  \"violations\": 0,\n  \"evidence\": [{\"id\":\"SECURITY-001\",\"refs\":[\"BEHAVIOR-001\"],\"source\":\"scanner\",\"status\":\"pass\"}]\n}\n",
		"bottleneck/execution/telemetry.json":  "{\n  \"adoption_rate\": 0.9,\n  \"error_rate\": 0.01,\n  \"evidence\": [{\"id\":\"EXECUTION-001\",\"refs\":[\"BEHAVIOR-001\",\"ASSURANCE-001\"],\"source\":\"telemetry\",\"status\":\"pass\"}]\n}\n",
	}

	for path, content := range overrides {
		files[path] = content
	}

	for relativePath, content := range files {
		fullPath := filepath.Join(basePath, relativePath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatalf("failed to create directory for %s: %v", relativePath, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatalf("failed to write %s: %v", relativePath, err)
		}
	}
}

func withoutTraceEvidence(overrides map[string]string) map[string]string {
	copy := map[string]string{
		"bottleneck/assurance/results.json":   "{\n  \"scenarios_total\": 1,\n  \"scenarios_passed\": 1,\n  \"scenarios_failed\": 0,\n  \"failures\": []\n}\n",
		"bottleneck/security/guardrails.json": "{\n  \"violations\": 0\n}\n",
		"bottleneck/execution/telemetry.json": "{\n  \"adoption_rate\": 0.9,\n  \"error_rate\": 0.01\n}\n",
	}
	for path, content := range overrides {
		copy[path] = content
	}
	return copy
}

func resultForCapability(t *testing.T, result models.EngineResult, capability string) models.ValidationResult {
	t.Helper()

	for _, check := range result.Results {
		if check.Capability == capability {
			return check
		}
	}

	t.Fatalf("missing capability %q in result: %#v", capability, result.Results)
	return models.ValidationResult{}
}

func engineResultHasDetail(result models.EngineResult, detail string) bool {
	for _, check := range result.Results {
		for _, candidate := range check.Details {
			if candidate == detail {
				return true
			}
		}
	}

	return false
}

func assertOnlyDetails(t *testing.T, actual []string, expected []string) {
	t.Helper()

	if len(actual) != len(expected) {
		t.Fatalf("expected details %#v, got %#v", expected, actual)
	}

	for index, detail := range expected {
		if actual[index] != detail {
			t.Fatalf("expected detail %d to be %q, got %q", index, detail, actual[index])
		}
	}

	for _, detail := range actual {
		if strings.Contains(detail, "Expected Behavior") || strings.Contains(detail, "Outcomes") || strings.Contains(detail, "Success Criteria") {
			t.Fatalf("unexpected warning for completed section in details %#v", actual)
		}
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

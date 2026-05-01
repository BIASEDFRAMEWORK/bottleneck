package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"bottleneck/internal/config"
	"bottleneck/internal/diagnosis"
	"bottleneck/internal/models"
	"bottleneck/internal/validator"
)

func TestInitializeProjectCreatesAIPDFRiskSummarizerStarter(t *testing.T) {
	basePath := t.TempDir()

	if err := initializeProject(basePath); err != nil {
		t.Fatalf("initializeProject returned error: %v", err)
	}

	expectedContent := map[string][]string{
		"bottleneck/intent/intent.md": {
			"AI PDF",
			"INTENT-001",
			"Summarize financial PDF risk without hiding uncertainty",
			"At least 95%",
			"100% of ambiguous risk clauses",
			"BEHAVIOR-001",
		},
		"bottleneck/behavior/behavior-spec.md": {
			"BEHAVIOR-001",
			"Flag ambiguous financial risk language",
			"Critical: true",
			"INTENT-001",
			"ASSURANCE-001",
			"must not summarize ambiguous risk language as a confirmed fact",
		},
		"bottleneck/design/architecture.md": {
			"DESIGN-001",
			"Local PDF risk summarization flow",
			"Risk clause detector",
			"Uncertainty flagging check",
		},
		"bottleneck/assurance/features/sample.feature": {
			"Feature: AI PDF risk summarization",
			"@BEHAVIOR-001",
			"Ambiguous risk clause is flagged",
		},
	}

	for relativePath, substrings := range expectedContent {
		content := readInitFile(t, basePath, relativePath)
		for _, substring := range substrings {
			if !strings.Contains(content, substring) {
				t.Fatalf("expected %s to contain %q:\n%s", relativePath, substring, content)
			}
		}
	}
}

func TestInitializeProjectCreatesParseableJSONEvidenceWithIntentionalAssuranceFailure(t *testing.T) {
	basePath := t.TempDir()

	if err := initializeProject(basePath); err != nil {
		t.Fatalf("initializeProject returned error: %v", err)
	}

	var assurance struct {
		ScenariosTotal  int `json:"scenarios_total"`
		ScenariosPassed int `json:"scenarios_passed"`
		ScenariosFailed int `json:"scenarios_failed"`
		Evidence        []struct {
			ID     string   `json:"id"`
			Refs   []string `json:"refs"`
			Status string   `json:"status"`
		} `json:"evidence"`
	}
	readInitJSON(t, basePath, "bottleneck/assurance/results.json", &assurance)
	if assurance.ScenariosTotal != 2 || assurance.ScenariosPassed != 1 || assurance.ScenariosFailed != 1 {
		t.Fatalf("expected one intentional assurance failure, got %#v", assurance)
	}
	if len(assurance.Evidence) != 1 || assurance.Evidence[0].ID != "ASSURANCE-001" || assurance.Evidence[0].Status != "fail" {
		t.Fatalf("expected ASSURANCE-001 failing evidence, got %#v", assurance.Evidence)
	}

	var security struct {
		Violations int `json:"violations"`
		Evidence   []struct {
			ID   string   `json:"id"`
			Refs []string `json:"refs"`
		} `json:"evidence"`
	}
	readInitJSON(t, basePath, "bottleneck/security/guardrails.json", &security)
	if security.Violations != 0 || len(security.Evidence) != 1 || security.Evidence[0].ID != "SECURITY-001" {
		t.Fatalf("expected SECURITY-001 passing guardrail evidence, got %#v", security)
	}
	if !stringSliceContains(security.Evidence[0].Refs, "INTENT-001") || !stringSliceContains(security.Evidence[0].Refs, "BEHAVIOR-001") {
		t.Fatalf("expected security refs to intent and behavior, got %#v", security.Evidence[0].Refs)
	}

	var execution struct {
		AdoptionRate float64 `json:"adoption_rate"`
		ErrorRate    float64 `json:"error_rate"`
		Evidence     []struct {
			ID   string   `json:"id"`
			Refs []string `json:"refs"`
		} `json:"evidence"`
	}
	readInitJSON(t, basePath, "bottleneck/execution/telemetry.json", &execution)
	if execution.AdoptionRate != 0.72 || execution.ErrorRate != 0.02 || len(execution.Evidence) != 1 || execution.Evidence[0].ID != "EXECUTION-001" {
		t.Fatalf("expected EXECUTION-001 telemetry evidence, got %#v", execution)
	}
}

func TestInitializeProjectDoesNotOverwriteExistingFiles(t *testing.T) {
	basePath := t.TempDir()
	intentPath := filepath.Join(basePath, "bottleneck", "intent", "intent.md")
	if err := os.MkdirAll(filepath.Dir(intentPath), 0o755); err != nil {
		t.Fatalf("create intent dir: %v", err)
	}
	if err := os.WriteFile(intentPath, []byte("custom intent"), 0o644); err != nil {
		t.Fatalf("write custom intent: %v", err)
	}

	if err := initializeProject(basePath); err != nil {
		t.Fatalf("initializeProject returned error: %v", err)
	}

	if content := readInitFile(t, basePath, "bottleneck/intent/intent.md"); content != "custom intent" {
		t.Fatalf("expected existing file not to be overwritten, got %q", content)
	}
}

func TestInitMessageGuidesFirstRun(t *testing.T) {
	expected := []string{
		"Bottleneck initialized.",
		"AI PDF Risk Summarizer",
		"intentionally leaves Assurance weak",
		"bottleneck diagnose",
		"bottleneck validate",
		"bottleneck scorecard",
		"Start with: bottleneck/intent/intent.md",
	}
	for _, substring := range expected {
		if !strings.Contains(initSuccessMessage, substring) {
			t.Fatalf("expected init message to contain %q:\n%s", substring, initSuccessMessage)
		}
	}
}

func TestGeneratedStarterValidatesWithAssuranceAsPrimaryBottleneck(t *testing.T) {
	basePath := t.TempDir()
	if err := initializeProject(basePath); err != nil {
		t.Fatalf("initializeProject returned error: %v", err)
	}

	result := validator.NewEngine(basePath, "default").Validate()
	if result.SystemStatus != models.StatusFail {
		t.Fatalf("expected generated starter to fail because assurance is intentionally weak, got %q with %#v", result.SystemStatus, result.Results)
	}
	if result.PrimaryBottleneck != "Assurance" {
		t.Fatalf("expected Assurance primary bottleneck, got %q", result.PrimaryBottleneck)
	}
	for _, capability := range []string{"Behavior", "Intent", "Design", "Security", "Execution", "Traceability"} {
		check := initResultForCapability(t, result, capability)
		if check.Status != models.StatusPass {
			t.Fatalf("expected %s PASS in generated starter, got %q with details %#v", capability, check.Status, check.Details)
		}
	}

	report := diagnosis.BuildReport(result)
	if report.PrimaryBottleneck != "Assurance" {
		t.Fatalf("expected diagnosis primary bottleneck Assurance, got %#v", report)
	}
	if len(report.ContributingFindings) == 0 || !strings.Contains(strings.Join(report.ContributingFindings, "\n"), "Ambiguous risk clause was summarized as confirmed exposure") {
		t.Fatalf("expected assurance failure in contributing findings, got %#v", report.ContributingFindings)
	}
	if report.RecommendedAction != "Add or fix evaluation evidence for BEHAVIOR-001 so ambiguous financial risk language is flagged as uncertain." {
		t.Fatalf("expected sample-specific recommended action, got %q", report.RecommendedAction)
	}
}

func TestDefaultConfigSupportsRoadmapEnvironmentsAndInheritance(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(configPath, []byte(defaultConfigYAML), 0o644); err != nil {
		t.Fatalf("write default config: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("load default config: %v", err)
	}

	expectedAccuracy := map[string]float64{
		"dev":        0.85,
		"test":       0.90,
		"stage":      0.93,
		"production": 0.95,
	}
	for env, expected := range expectedAccuracy {
		t.Run(env, func(t *testing.T) {
			resolved := config.ResolveEnvironment(cfg, env)
			if resolved.Assurance.MinAccuracy != expected {
				t.Fatalf("expected %s min accuracy %.2f, got %.2f", env, expected, resolved.Assurance.MinAccuracy)
			}
			if resolved.Assurance.MaxFailures != 0 {
				t.Fatalf("expected %s to inherit max failures 0, got %d", env, resolved.Assurance.MaxFailures)
			}
			if resolved.Execution.Telemetry.MaxAgeHours != 168 {
				t.Fatalf("expected %s to inherit telemetry freshness threshold, got %d", env, resolved.Execution.Telemetry.MaxAgeHours)
			}
			if resolved.Security.SARIF.MaxMedium != 5 {
				t.Fatalf("expected %s to inherit SARIF medium threshold, got %d", env, resolved.Security.SARIF.MaxMedium)
			}
		})
	}

	production := config.ResolveEnvironment(cfg, "production")
	if production.Gate.Release.MinPrimaryScore != 75 {
		t.Fatalf("expected production release gate primary threshold 75, got %d", production.Gate.Release.MinPrimaryScore)
	}
	if !production.Gate.Release.RequireGovernance {
		t.Fatal("expected production release gate to require governance")
	}
}

func readInitFile(t *testing.T, basePath string, relativePath string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(basePath, relativePath))
	if err != nil {
		t.Fatalf("read %s: %v", relativePath, err)
	}
	return string(content)
}

func readInitJSON(t *testing.T, basePath string, relativePath string, target any) {
	t.Helper()
	content := readInitFile(t, basePath, relativePath)
	if err := json.Unmarshal([]byte(content), target); err != nil {
		t.Fatalf("parse %s: %v\n%s", relativePath, err, content)
	}
}

func initResultForCapability(t *testing.T, result models.EngineResult, capability string) models.ValidationResult {
	t.Helper()
	for _, check := range result.Results {
		if check.Capability == capability {
			return check
		}
	}
	t.Fatalf("missing capability %q in %#v", capability, result.Results)
	return models.ValidationResult{}
}

func stringSliceContains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

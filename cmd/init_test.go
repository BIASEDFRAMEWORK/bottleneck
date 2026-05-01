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
	"bottleneck/internal/scorecard"
	"bottleneck/internal/traceability"
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

func TestInitializeProjectWithSaaSTemplateCreatesSubscriptionBillingStarter(t *testing.T) {
	basePath := t.TempDir()

	if err := initializeProjectWithTemplate(basePath, initTemplateSaaS); err != nil {
		t.Fatalf("initializeProjectWithTemplate returned error: %v", err)
	}

	expectedContent := map[string][]string{
		"bottleneck/intent/intent.md": {
			"Subscription Billing Release",
			"INTENT-001",
			"Customers must be able to update payment methods without duplicate charges",
			"BEHAVIOR-001",
			"BEHAVIOR-002",
			"BEHAVIOR-003",
		},
		"bottleneck/behavior/behavior-spec.md": {
			"BEHAVIOR-001",
			"Customer updates payment method",
			"BEHAVIOR-002",
			"Failed invoice is retried",
			"BEHAVIOR-003",
			"Duplicate charges are prevented",
			"idempotency key",
		},
		"bottleneck/design/architecture.md": {
			"DESIGN-001",
			"Tokenized billing retry flow",
			"Payment provider tokenization API",
			"Invoice retry worker",
		},
		"bottleneck/assurance/features/sample.feature": {
			"Feature: Subscription billing release",
			"@BEHAVIOR-003",
			"Duplicate retry charge is prevented",
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

	var assurance struct {
		ScenariosTotal  int `json:"scenarios_total"`
		ScenariosPassed int `json:"scenarios_passed"`
		ScenariosFailed int `json:"scenarios_failed"`
		Evidence        []struct {
			ID   string   `json:"id"`
			Refs []string `json:"refs"`
		} `json:"evidence"`
	}
	readInitJSON(t, basePath, "bottleneck/assurance/results.json", &assurance)
	if assurance.ScenariosTotal != 3 || assurance.ScenariosPassed != 3 || assurance.ScenariosFailed != 0 {
		t.Fatalf("expected SaaS starter to have passing test metrics with one intentional mapping gap, got %#v", assurance)
	}
	if len(assurance.Evidence) != 2 || assurance.Evidence[0].ID != "ASSURANCE-001" || assurance.Evidence[1].ID != "ASSURANCE-002" {
		t.Fatalf("expected two mapped SaaS assurance evidence entries, got %#v", assurance.Evidence)
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
		t.Fatalf("expected passing SaaS security evidence, got %#v", security)
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
	if execution.AdoptionRate != 0.68 || execution.ErrorRate != 0.015 || len(execution.Evidence) != 1 || execution.Evidence[0].ID != "EXECUTION-001" {
		t.Fatalf("expected SaaS execution telemetry evidence, got %#v", execution)
	}
}

func TestGeneratedSaaSStarterShowsAssuranceBottleneckAndTraceableGap(t *testing.T) {
	basePath := t.TempDir()
	if err := initializeProjectWithTemplate(basePath, initTemplateSaaS); err != nil {
		t.Fatalf("initializeProjectWithTemplate returned error: %v", err)
	}

	result := validator.NewEngine(basePath, "default").Validate()
	if result.SystemStatus != models.StatusWarning {
		t.Fatalf("expected SaaS starter to warn because BEHAVIOR-003 lacks mapped test evidence, got %q with %#v", result.SystemStatus, result.Results)
	}
	if result.PrimaryBottleneck != "Traceability" {
		t.Fatalf("expected engine primary bottleneck Traceability before diagnosis weighting, got %q", result.PrimaryBottleneck)
	}

	report := diagnosis.BuildReport(result)
	if report.PrimaryBottleneck != "Assurance" {
		t.Fatalf("expected diagnosis primary bottleneck Assurance, got %#v", report)
	}
	if report.RecommendedAction != "Add assurance evidence for payment retry behavior." {
		t.Fatalf("expected SaaS-specific recommended action, got %q", report.RecommendedAction)
	}

	output, err := scorecard.Render(result, "text")
	if err != nil {
		t.Fatalf("scorecard render returned error: %v", err)
	}
	for _, substring := range []string{
		"Release Recommendation: Conditional",
		"Primary Bottleneck: Assurance",
		"no mapped test evidence",
		"Next Action:",
		"Add assurance evidence for payment retry behavior. Map it to BEHAVIOR-003.",
	} {
		if !strings.Contains(output, substring) {
			t.Fatalf("expected %q in SaaS scorecard:\n%s", substring, output)
		}
	}

	graph, err := traceability.Build(filepath.Join(basePath, "bottleneck"), traceability.Options{Environment: "default"})
	if err != nil {
		t.Fatalf("traceability build returned error: %v", err)
	}
	trace, err := graph.Trace("BEHAVIOR-003")
	if err != nil {
		t.Fatalf("trace BEHAVIOR-003 returned error: %v", err)
	}
	traceOutput := traceability.RenderText(trace)
	for _, substring := range []string{
		"Trace: BEHAVIOR-003",
		"Duplicate charges are prevented during retry",
		"BEHAVIOR-003 has no mapped test evidence",
		"Recommendation:",
		"Add assurance evidence for payment retry behavior.",
	} {
		if !strings.Contains(traceOutput, substring) {
			t.Fatalf("expected %q in SaaS trace:\n%s", substring, traceOutput)
		}
	}
}

func TestInitializeProjectWithSaaSTemplateDoesNotOverwriteExistingFiles(t *testing.T) {
	basePath := t.TempDir()
	intentPath := filepath.Join(basePath, "bottleneck", "intent", "intent.md")
	if err := os.MkdirAll(filepath.Dir(intentPath), 0o755); err != nil {
		t.Fatalf("create intent dir: %v", err)
	}
	if err := os.WriteFile(intentPath, []byte("custom SaaS intent"), 0o644); err != nil {
		t.Fatalf("write custom intent: %v", err)
	}

	if err := initializeProjectWithTemplate(basePath, initTemplateSaaS); err != nil {
		t.Fatalf("initializeProjectWithTemplate returned error: %v", err)
	}

	if content := readInitFile(t, basePath, "bottleneck/intent/intent.md"); content != "custom SaaS intent" {
		t.Fatalf("expected existing file not to be overwritten, got %q", content)
	}
}

func TestInitializeProjectWithUnsupportedTemplateReturnsUsefulError(t *testing.T) {
	err := initializeProjectWithTemplate(t.TempDir(), "enterprise")
	if err == nil {
		t.Fatal("expected unsupported template error")
	}
	if !strings.Contains(err.Error(), `unsupported init template "enterprise"`) ||
		!strings.Contains(err.Error(), "supported: default, saas") {
		t.Fatalf("unexpected unsupported template error: %v", err)
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
		"local":      0.75,
		"dev":        0.85,
		"test":       0.92,
		"stage":      0.95,
		"production": 0.97,
	}
	for env, expected := range expectedAccuracy {
		t.Run(env, func(t *testing.T) {
			resolved, err := config.ResolveEnvironmentStrict(cfg, env)
			if err != nil {
				t.Fatalf("resolve %s: %v", env, err)
			}
			if resolved.Assurance.MinAccuracy != expected {
				t.Fatalf("expected %s min accuracy %.2f, got %.2f", env, expected, resolved.Assurance.MinAccuracy)
			}
			if env == "local" || env == "dev" {
				if resolved.Gate.Release.RequireTraceability {
					t.Fatalf("expected %s release gate to allow optional traceability", env)
				}
				if resolved.Execution.Telemetry.MaxAgeHours != 0 {
					t.Fatalf("expected %s to allow stale telemetry for local feedback, got max age %d", env, resolved.Execution.Telemetry.MaxAgeHours)
				}
				return
			}
			if !resolved.Gate.Release.RequireTraceability {
				t.Fatalf("expected %s release gate to require traceability", env)
			}
			if env == "test" && resolved.Security.SARIF.MaxMedium != 2 {
				t.Fatalf("expected test SARIF medium threshold 2, got %d", resolved.Security.SARIF.MaxMedium)
			}
			if env == "stage" && resolved.Security.SARIF.MaxMedium != 1 {
				t.Fatalf("expected stage SARIF medium threshold 1, got %d", resolved.Security.SARIF.MaxMedium)
			}
		})
	}

	for _, env := range []string{"default", "local", "dev", "test", "stage", "production"} {
		if _, ok := cfg.Environments[env]; !ok {
			t.Fatalf("expected generated default config to include %s environment", env)
		}
	}

	production, err := config.ResolveEnvironmentStrict(cfg, "production")
	if err != nil {
		t.Fatalf("resolve production: %v", err)
	}
	if production.Gate.Release.MinPrimaryScore != 85 {
		t.Fatalf("expected production release gate primary threshold 85, got %d", production.Gate.Release.MinPrimaryScore)
	}
	if !production.Gate.Release.RequireGovernance {
		t.Fatal("expected production release gate to require governance")
	}
	if production.Security.SARIF.MaxCritical != 0 || production.Security.SARIF.MaxMedium != 0 {
		t.Fatalf("expected production security thresholds to block critical and medium findings, got %#v", production.Security.SARIF)
	}
}

func TestSaaSConfigSupportsPracticalEnvironmentDefaults(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(configPath, []byte(saasConfigYAML), 0o644); err != nil {
		t.Fatalf("write SaaS config: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("load SaaS config: %v", err)
	}
	for _, env := range []string{"default", "local", "dev", "test", "stage", "production"} {
		if _, ok := cfg.Environments[env]; !ok {
			t.Fatalf("expected SaaS config to include %s environment", env)
		}
	}

	dev, err := config.ResolveEnvironmentStrict(cfg, "dev")
	if err != nil {
		t.Fatalf("resolve dev: %v", err)
	}
	production, err := config.ResolveEnvironmentStrict(cfg, "production")
	if err != nil {
		t.Fatalf("resolve production: %v", err)
	}
	if dev.Gate.Release.RequireTraceability {
		t.Fatal("expected dev release gate to allow traceability warnings")
	}
	if !production.Gate.Release.RequireTraceability {
		t.Fatal("expected production release gate to require traceability")
	}
	if dev.Gate.Release.MinPrimaryScore >= production.Gate.Release.MinPrimaryScore {
		t.Fatalf("expected production min score stricter than dev, got dev=%d production=%d", dev.Gate.Release.MinPrimaryScore, production.Gate.Release.MinPrimaryScore)
	}
	if dev.Security.SARIF.MaxCritical <= production.Security.SARIF.MaxCritical {
		t.Fatalf("expected production critical security threshold stricter than dev, got dev=%d production=%d", dev.Security.SARIF.MaxCritical, production.Security.SARIF.MaxCritical)
	}
	if dev.Execution.Telemetry.MaxAgeHours != 0 || production.Execution.Telemetry.MaxAgeHours == 0 {
		t.Fatalf("expected dev to allow stale telemetry and production to enforce freshness, got dev=%d production=%d", dev.Execution.Telemetry.MaxAgeHours, production.Execution.Telemetry.MaxAgeHours)
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

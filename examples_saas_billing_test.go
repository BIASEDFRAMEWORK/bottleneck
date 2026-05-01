package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"bottleneck/internal/diagnosis"
	"bottleneck/internal/ingest"
	"bottleneck/internal/models"
	"bottleneck/internal/scorecard"
	"bottleneck/internal/traceability"
	"bottleneck/internal/validator"

	"gopkg.in/yaml.v3"
)

func TestSaaSBillingExampleStructureAndReadme(t *testing.T) {
	exampleRoot := filepath.Join("examples", "saas-billing")

	requiredFiles := []string{
		"README.md",
		"bottleneck/config.yaml",
		"bottleneck/intent/intent.md",
		"bottleneck/behavior/behavior-spec.md",
		"bottleneck/design/architecture.md",
		"bottleneck/assurance/results.json",
		"bottleneck/security/guardrails.json",
		"bottleneck/execution/telemetry.json",
		"reports/cucumber.json",
		"reports/codeql.sarif",
		"reports/test-summary.json",
		"reports/telemetry.json",
		".github/workflows/bottleneck.yml",
	}
	for _, relativePath := range requiredFiles {
		assertFileExists(t, filepath.Join(exampleRoot, relativePath))
	}

	readme := readText(t, filepath.Join(exampleRoot, "README.md"))
	assertContainsAll(t, readme, []string{
		"Subscription Billing Release",
		"What Is Intentionally Broken",
		"BEHAVIOR-003",
		"duplicate charges",
		"cd examples/saas-billing",
		"bottleneck validate",
		"bottleneck scorecard",
		"bottleneck diagnose",
		"bottleneck trace BEHAVIOR-003",
		"bottleneck ingest cucumber --file reports/cucumber.json",
		"bottleneck ingest sarif --file reports/codeql.sarif",
		"bottleneck scorecard --env=production",
		".github/workflows/bottleneck.yml",
		"File Map",
	})

	intent := readText(t, filepath.Join(exampleRoot, "bottleneck", "intent", "intent.md"))
	assertContainsAll(t, intent, []string{
		"INTENT-001: Subscription Billing Release",
		"BEHAVIOR-001",
		"BEHAVIOR-002",
		"BEHAVIOR-003",
	})

	behavior := readText(t, filepath.Join(exampleRoot, "bottleneck", "behavior", "behavior-spec.md"))
	assertContainsAll(t, behavior, []string{
		"BEHAVIOR-001: Customer updates payment method",
		"BEHAVIOR-002: Failed invoice is retried after payment method update",
		"BEHAVIOR-003: Payment retry does not create duplicate charges",
		"The system must not create duplicate charges",
	})

	initialAssurance := readText(t, filepath.Join(exampleRoot, "bottleneck", "assurance", "results.json"))
	if strings.Contains(initialAssurance, "ASSURANCE-003") || strings.Contains(initialAssurance, "BEHAVIOR-003") {
		t.Fatalf("initial demo assurance should intentionally omit BEHAVIOR-003 coverage:\n%s", initialAssurance)
	}

	workflowPath := filepath.Join(exampleRoot, ".github", "workflows", "bottleneck.yml")
	workflow := readText(t, workflowPath)
	var parsedWorkflow map[string]interface{}
	if err := yaml.Unmarshal([]byte(workflow), &parsedWorkflow); err != nil {
		t.Fatalf("workflow YAML should parse: %v\n%s", err, workflow)
	}
	assertContainsAll(t, workflow, []string{
		"actions/checkout@v4",
		"actions/setup-go@v5",
		"go build -o ./bin/bottleneck .",
		"../../bin/bottleneck validate",
		"../../bin/bottleneck scorecard",
		"../../bin/bottleneck diagnose",
		"$GITHUB_STEP_SUMMARY",
	})
}

func TestSaaSBillingExampleDiagnosesGapAndIngestionImprovesCoverage(t *testing.T) {
	exampleRoot := filepath.Join("examples", "saas-billing")

	initial := validator.NewEngine(exampleRoot, "default").Validate()
	if initial.SystemStatus != models.StatusWarning {
		t.Fatalf("expected initial example to warn, got %q with results %#v", initial.SystemStatus, initial.Results)
	}

	initialDiagnosis := diagnosis.Analyze(initial)
	if initialDiagnosis.PrimaryBottleneck != "Assurance" {
		t.Fatalf("expected initial primary bottleneck to be Assurance, got %#v", initialDiagnosis)
	}

	initialCard := scorecard.Build(initial)
	if initialCard.ReleaseRecommendation != scorecard.RecommendationConditional {
		t.Fatalf("expected initial recommendation to be Conditional, got %#v", initialCard)
	}

	initialGraph, err := traceability.Build(filepath.Join(exampleRoot, "bottleneck"), traceability.Options{Environment: "default"})
	if err != nil {
		t.Fatalf("build initial traceability graph: %v", err)
	}
	initialTrace, err := initialGraph.Trace("BEHAVIOR-003")
	if err != nil {
		t.Fatalf("trace initial BEHAVIOR-003: %v", err)
	}
	if !hasSubstring(initialTrace.MissingLinks, "BEHAVIOR-003 has no mapped test evidence") {
		t.Fatalf("expected initial trace to show missing BEHAVIOR-003 assurance, got %#v", initialTrace.MissingLinks)
	}

	tempRoot := filepath.Join(t.TempDir(), "saas-billing")
	copyDir(t, exampleRoot, tempRoot)

	cucumberSummary, err := ingest.IngestCucumber(tempRoot, filepath.Join(tempRoot, "reports", "cucumber.json"), "", false, false)
	if err != nil {
		t.Fatalf("ingest cucumber sample: %v", err)
	}
	cucumberArtifact, ok := cucumberSummary.Artifact.(ingest.AssuranceArtifact)
	if !ok {
		t.Fatalf("expected cucumber ingest to return assurance artifact, got %T", cucumberSummary.Artifact)
	}
	if cucumberArtifact.ScenariosPassed != 3 || cucumberArtifact.ScenariosFailed != 0 {
		t.Fatalf("expected cucumber report to pass all scenarios, got %#v", cucumberArtifact)
	}

	updatedAssurance := readText(t, filepath.Join(tempRoot, "bottleneck", "assurance", "results.json"))
	assertContainsAll(t, updatedAssurance, []string{"ASSURANCE-003", "BEHAVIOR-003"})

	updatedGraph, err := traceability.Build(filepath.Join(tempRoot, "bottleneck"), traceability.Options{Environment: "default"})
	if err != nil {
		t.Fatalf("build updated traceability graph: %v", err)
	}
	updatedTrace, err := updatedGraph.Trace("BEHAVIOR-003")
	if err != nil {
		t.Fatalf("trace updated BEHAVIOR-003: %v", err)
	}
	if hasSubstring(updatedTrace.MissingLinks, "BEHAVIOR-003 has no mapped test evidence") {
		t.Fatalf("expected cucumber ingestion to remove missing assurance gap, got %#v", updatedTrace.MissingLinks)
	}
	if !evidenceLinksContain(updatedTrace.RelatedAssurance, "ASSURANCE-003") {
		t.Fatalf("expected updated trace to include ASSURANCE-003, got %#v", updatedTrace.RelatedAssurance)
	}

	sarifSummary, err := ingest.IngestSARIF(tempRoot, filepath.Join(tempRoot, "reports", "codeql.sarif"), "", false, false)
	if err != nil {
		t.Fatalf("ingest SARIF sample: %v", err)
	}
	if _, ok := sarifSummary.Artifact.(ingest.SecurityArtifact); !ok {
		t.Fatalf("expected SARIF ingest to return security artifact, got %T", sarifSummary.Artifact)
	}
	updatedSecurity := readText(t, filepath.Join(tempRoot, "bottleneck", "security", "guardrails.json"))
	assertContainsAll(t, updatedSecurity, []string{"SECURITY-001", "BEHAVIOR-003"})

	production := validator.NewEngine(tempRoot, "production").Validate()
	if production.SystemStatus != models.StatusPass {
		t.Fatalf("expected production scorecard to pass after ingestion, got %q with results %#v", production.SystemStatus, production.Results)
	}

	productionCard := scorecard.Build(production)
	if productionCard.ReleaseRecommendation != scorecard.RecommendationProceed {
		t.Fatalf("expected production recommendation to proceed after ingestion, got %#v", productionCard)
	}
	if productionCard.PrimaryBottleneck != diagnosis.HealthyPrimaryBottleneck {
		t.Fatalf("expected no production bottleneck after ingestion, got %#v", productionCard)
	}
}

func TestSaaSBillingDocsPointToDemoReadyExample(t *testing.T) {
	readme := readText(t, "readme.md")
	if !strings.Contains(readme, "examples/saas-billing") {
		t.Fatal("README should link the demo-ready SaaS billing example")
	}

	quickstart := readText(t, filepath.Join("docs", "quickstart-saas.md"))
	if !strings.Contains(quickstart, "examples/saas-billing") {
		t.Fatal("SaaS quickstart should point to the demo-ready example")
	}
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("expected %s to exist: %v", path, err)
	}
	if info.IsDir() {
		t.Fatalf("expected %s to be a file", path)
	}
}

func readText(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(content)
}

func assertContainsAll(t *testing.T, text string, expected []string) {
	t.Helper()
	for _, substring := range expected {
		if !strings.Contains(text, substring) {
			t.Fatalf("expected text to contain %q:\n%s", substring, text)
		}
	}
}

func hasSubstring(values []string, substring string) bool {
	for _, value := range values {
		if strings.Contains(value, substring) {
			return true
		}
	}
	return false
}

func evidenceLinksContain(links []traceability.EvidenceLink, id string) bool {
	for _, link := range links {
		if link.ID == id {
			return true
		}
	}
	return false
}

func copyDir(t *testing.T, src string, dst string) {
	t.Helper()

	info, err := os.Stat(src)
	if err != nil {
		t.Fatalf("stat source directory %s: %v", src, err)
	}
	if !info.IsDir() {
		t.Fatalf("source %s is not a directory", src)
	}
	if err := os.MkdirAll(dst, info.Mode()); err != nil {
		t.Fatalf("create destination directory %s: %v", dst, err)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatalf("read source directory %s: %v", src, err)
	}
	for _, entry := range entries {
		sourcePath := filepath.Join(src, entry.Name())
		destinationPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			copyDir(t, sourcePath, destinationPath)
			continue
		}

		entryInfo, err := entry.Info()
		if err != nil {
			t.Fatalf("stat source file %s: %v", sourcePath, err)
		}
		content, err := os.ReadFile(sourcePath)
		if err != nil {
			t.Fatalf("read source file %s: %v", sourcePath, err)
		}
		if err := os.WriteFile(destinationPath, content, entryInfo.Mode()); err != nil {
			t.Fatalf("write destination file %s: %v", destinationPath, err)
		}
	}
}

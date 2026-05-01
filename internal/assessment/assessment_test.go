package assessment

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"bottleneck/internal/discover"
	"bottleneck/internal/models"
)

func TestBuildReportsMaturityAndScoreConfidence(t *testing.T) {
	root := t.TempDir()
	writeAssessmentFile(t, root, "bottleneck/config.yaml", "version: 1\n")
	writeAssessmentFile(t, root, "bottleneck/assurance/results.json", `{"evidence":[{"generated_by":"junit","provenance":"test"}]}`)
	writeAssessmentFile(t, root, "bottleneck/security/guardrails.json", `{"evidence":[{"generated_by":"sarif","provenance":"test"}]}`)
	writeAssessmentFile(t, root, "bottleneck/execution/telemetry.json", `{"evidence":[{"generated_by":"telemetry","provenance":"test"}]}`)
	writeAssessmentFile(t, root, ".github/workflows/ci.yml", "name: ci\n")

	discovery, err := discover.Scan(root)
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	result := models.EngineResult{
		Environment:  "default",
		SystemStatus: models.StatusWarning,
		Results: []models.ValidationResult{
			{Capability: "Assurance", Status: models.StatusPass, Message: "ok"},
			{Capability: "Security", Status: models.StatusPass, Message: "ok"},
			{Capability: "Execution", Status: models.StatusPass, Message: "ok"},
		},
	}

	report := Build(result, discovery, Options{RootPath: root, Now: time.Now().UTC()})
	if report.Maturity.Level < 3 || report.ScoreConfidence != "high" || report.AIReadiness == "Blocked" {
		t.Fatalf("unexpected assessment report: %#v", report)
	}
	if len(report.Categories) != 3 || report.Categories[0].Provenance != "tool-generated" {
		t.Fatalf("expected tool-generated category provenance, got %#v", report.Categories)
	}
}

func TestRenderMarkdownContainsSummaryTable(t *testing.T) {
	report := Report{
		SchemaVersion:         SchemaVersion,
		Environment:           "default",
		Maturity:              Maturity{Level: 1, Label: "Documented"},
		AIReadiness:           "Limited",
		ReleaseRecommendation: "Conditional",
		PrimaryBottleneck:     "Assurance",
		ScoreConfidence:       "medium",
		NextAction:            "Add mapped assurance evidence.",
		UsefulCommands:        []string{"bottleneck discover"},
		Categories: []CategoryEvidence{{
			Name:       "Assurance",
			Status:     "WARN",
			Score:      70,
			Confidence: "medium",
			Provenance: "local artifact",
			Freshness:  "fresh",
		}},
	}

	output, err := Render(report, FormatMarkdown)
	if err != nil {
		t.Fatalf("render markdown: %v", err)
	}
	for _, expected := range []string{"# Bottleneck Assessment", "| Category | Status | Score |", "## Next Action", "`bottleneck discover`"} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected markdown to contain %q:\n%s", expected, output)
		}
	}
}

func writeAssessmentFile(t *testing.T, root string, rel string, content string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", rel, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
}

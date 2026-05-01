package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExplainCategoryFilter(t *testing.T) {
	binary := buildBottleneckBinary(t)
	projectDir := t.TempDir()
	initResult := runBottleneck(t, binary, projectDir, "init", "--template", "saas")
	assertExitCode(t, initResult, 0)

	result := runBottleneck(t, binary, projectDir, "explain", "--category=assurance")
	assertExitCode(t, result, 0)
	assertOutputContains(t, result, "Assurance Score:")
	if strings.Contains(result.stdout, "Intent Score:") {
		t.Fatalf("category filter should only render Assurance:\n%s", result.stdout)
	}
}

func TestExplainMarkdownOutputAndOutFile(t *testing.T) {
	binary := buildBottleneckBinary(t)
	projectDir := t.TempDir()
	initResult := runBottleneck(t, binary, projectDir, "init", "--template", "saas")
	assertExitCode(t, initResult, 0)

	result := runBottleneck(t, binary, projectDir, "explain", "--category=assurance", "--format=markdown", "--out=bottleneck/reports/bottleneck-explanation.md")
	assertExitCode(t, result, 0)
	assertOutputContains(t, result, "# Bottleneck Explanation")
	assertOutputContains(t, result, "## Assurance")
	assertOutputContains(t, result, "### Suggested Automation")

	reportPath := filepath.Join(projectDir, "bottleneck", "reports", "bottleneck-explanation.md")
	content, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("read explanation report: %v", err)
	}
	if !strings.Contains(string(content), "## Assurance") {
		t.Fatalf("expected markdown explanation report, got:\n%s", string(content))
	}
}

func TestExplainJSONOutput(t *testing.T) {
	binary := buildBottleneckBinary(t)
	projectDir := t.TempDir()
	initResult := runBottleneck(t, binary, projectDir, "init", "--template", "saas")
	assertExitCode(t, initResult, 0)

	result := runBottleneck(t, binary, projectDir, "explain", "--category=assurance", "--format=json")
	assertExitCode(t, result, 0)
	var decoded struct {
		SchemaVersion string `json:"schema_version"`
		Explanations  []struct {
			Category             string   `json:"category"`
			RiskToDelivery       string   `json:"risk_to_delivery"`
			SuggestedOwnerRoles  []string `json:"suggested_owner_roles"`
			SuggestedAutomations []string `json:"suggested_automations"`
		} `json:"explanations"`
	}
	if err := json.Unmarshal([]byte(result.stdout), &decoded); err != nil {
		t.Fatalf("explain JSON did not parse: %v\nstdout:\n%s\nstderr:\n%s", err, result.stdout, result.stderr)
	}
	if decoded.SchemaVersion != "explain.v2" || len(decoded.Explanations) != 1 || decoded.Explanations[0].Category != "Assurance" {
		t.Fatalf("unexpected explain JSON: %#v", decoded)
	}
	if decoded.Explanations[0].RiskToDelivery == "" || len(decoded.Explanations[0].SuggestedOwnerRoles) == 0 || len(decoded.Explanations[0].SuggestedAutomations) == 0 {
		t.Fatalf("expected risk, owner roles, and automations in JSON: %#v", decoded.Explanations[0])
	}
}

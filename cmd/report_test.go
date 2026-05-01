package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReportCreatesDefaultMarkdownFile(t *testing.T) {
	binary := buildBottleneckBinary(t)
	projectDir := t.TempDir()
	initResult := runBottleneck(t, binary, projectDir, "init", "--template", "saas")
	assertExitCode(t, initResult, 0)

	result := runBottleneck(t, binary, projectDir, "report")
	assertExitCode(t, result, 0)
	assertOutputContains(t, result, "SDLC evidence report created")
	assertOutputContains(t, result, "Report: bottleneck/reports/sdlc-evidence-report.md")

	reportPath := filepath.Join(projectDir, "bottleneck", "reports", "sdlc-evidence-report.md")
	content, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("read default report: %v", err)
	}
	for _, expected := range []string{
		"# SDLC Evidence Report",
		"## Executive Summary",
		"## Current Delivery-System Status",
		"## Primary Bottleneck",
		"## Category Scorecard",
		"## Trend Summary",
		"## Evidence Found",
		"## Evidence Missing",
		"## Risk to Delivery",
		"## Recommended Actions",
		"## Suggested Owners",
		"## Suggested Automation",
		"## Leadership Decision Needed",
		"## Appendix: Snapshot Metadata",
	} {
		if !strings.Contains(string(content), expected) {
			t.Fatalf("expected %q in default report:\n%s", expected, string(content))
		}
	}
}

func TestReportSupportsJSON(t *testing.T) {
	binary := buildBottleneckBinary(t)
	projectDir := t.TempDir()
	initResult := runBottleneck(t, binary, projectDir, "init", "--template", "saas")
	assertExitCode(t, initResult, 0)

	result := runBottleneck(t, binary, projectDir, "report", "--format=json")
	assertExitCode(t, result, 0)
	var decoded struct {
		SchemaVersion      string   `json:"schema_version"`
		PrimaryBottleneck  string   `json:"primary_bottleneck"`
		EvidenceMissing    []string `json:"evidence_missing"`
		LeadershipDecision string   `json:"leadership_decision"`
	}
	if err := json.Unmarshal([]byte(result.stdout), &decoded); err != nil {
		t.Fatalf("report JSON did not parse: %v\nstdout:\n%s\nstderr:\n%s", err, result.stdout, result.stderr)
	}
	if decoded.SchemaVersion != "sdlc.evidence.report.v1" || decoded.PrimaryBottleneck != "Assurance" || decoded.LeadershipDecision == "" {
		t.Fatalf("unexpected report JSON: %#v", decoded)
	}
}

func TestReportSupportsCustomOutputPath(t *testing.T) {
	binary := buildBottleneckBinary(t)
	projectDir := t.TempDir()
	initResult := runBottleneck(t, binary, projectDir, "init", "--template", "saas")
	assertExitCode(t, initResult, 0)

	result := runBottleneck(t, binary, projectDir, "report", "--out=bottleneck/reports/custom-sdlc-report.md")
	assertExitCode(t, result, 0)
	assertOutputContains(t, result, "Report: bottleneck/reports/custom-sdlc-report.md")

	if _, err := os.Stat(filepath.Join(projectDir, "bottleneck", "reports", "custom-sdlc-report.md")); err != nil {
		t.Fatalf("expected custom report path: %v", err)
	}
}

func TestReportIncludesTrendSummaryWhenSnapshotsExist(t *testing.T) {
	binary := buildBottleneckBinary(t)
	projectDir := t.TempDir()
	initResult := runBottleneck(t, binary, projectDir, "init", "--template", "saas")
	assertExitCode(t, initResult, 0)

	first := runBottleneck(t, binary, projectDir, "snapshot")
	assertExitCode(t, first, 0)
	second := runBottleneck(t, binary, projectDir, "snapshot", "--label=second")
	assertExitCode(t, second, 0)

	result := runBottleneck(t, binary, projectDir, "report")
	assertExitCode(t, result, 0)
	reportPath := filepath.Join(projectDir, "bottleneck", "reports", "sdlc-evidence-report.md")
	content, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}
	if !strings.Contains(string(content), "Overall direction is") ||
		!strings.Contains(string(content), "Assurance appeared as the primary bottleneck") {
		t.Fatalf("expected trend summary when snapshots exist:\n%s", string(content))
	}
}

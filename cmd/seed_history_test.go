package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSeedHistoryCreatesSnapshots(t *testing.T) {
	binary := buildBottleneckBinary(t)
	projectDir := t.TempDir()

	result := runBottleneck(t, binary, projectDir, "seed-history")
	assertExitCode(t, result, 0)
	for _, expected := range []string{
		"Bottleneck seed history created",
		"Scenario: saas-day-one",
		"Environment: default",
		"Snapshots: 6",
		"Output: bottleneck/history/scorecards",
		"Run `bottleneck trends`",
	} {
		assertOutputContains(t, result, expected)
	}

	files := seedHistoryCLIFiles(t, projectDir)
	if len(files) != 6 {
		t.Fatalf("expected 6 seeded snapshots, got %d: %#v", len(files), files)
	}
	decoded := readSeedHistoryCLISnapshot(t, files[0])
	if decoded.SchemaVersion != "scorecard.snapshot.v1" || decoded.Scorecard.SchemaVersion != "scorecard.v2" {
		t.Fatalf("unexpected seeded snapshot schema: %#v", decoded)
	}
}

func TestSeedHistoryDoesNotOverwriteByDefault(t *testing.T) {
	binary := buildBottleneckBinary(t)
	projectDir := t.TempDir()

	first := runBottleneck(t, binary, projectDir, "seed-history")
	assertExitCode(t, first, 0)

	second := runBottleneck(t, binary, projectDir, "seed-history")
	assertExitCode(t, second, 1)
	assertOutputContains(t, second, "Seed history already exists in bottleneck/history/scorecards.")
	assertOutputContains(t, second, "Next action: use --overwrite or choose a different --out path.")
}

func TestSeedHistoryOverwriteFlag(t *testing.T) {
	binary := buildBottleneckBinary(t)
	projectDir := t.TempDir()

	first := runBottleneck(t, binary, projectDir, "seed-history")
	assertExitCode(t, first, 0)
	second := runBottleneck(t, binary, projectDir, "seed-history", "--overwrite")
	assertExitCode(t, second, 0)

	files := seedHistoryCLIFiles(t, projectDir)
	if len(files) != 6 {
		t.Fatalf("expected 6 seeded snapshots after overwrite, got %d: %#v", len(files), files)
	}
}

func TestSeedHistoryWorksWithTrendsAndReport(t *testing.T) {
	binary := buildBottleneckBinary(t)
	projectDir := t.TempDir()

	initResult := runBottleneck(t, binary, projectDir, "init", "--template", "saas")
	assertExitCode(t, initResult, 0)

	seedResult := runBottleneck(t, binary, projectDir, "seed-history")
	assertExitCode(t, seedResult, 0)

	trendsResult := runBottleneck(t, binary, projectDir, "trends")
	assertExitCode(t, trendsResult, 0)
	assertOutputContains(t, trendsResult, "Snapshots analyzed: 6")
	assertOutputContains(t, trendsResult, "Overall direction: Improving")

	reportResult := runBottleneck(t, binary, projectDir, "report")
	assertExitCode(t, reportResult, 0)
	assertOutputContains(t, reportResult, "SDLC evidence report created")
	reportPath := filepath.Join(projectDir, "bottleneck", "reports", "sdlc-evidence-report.md")
	content, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}
	if !strings.Contains(string(content), "Overall direction is `improving`") ||
		!strings.Contains(string(content), "The team is improving overall") {
		t.Fatalf("expected seeded trend summary in report:\n%s", string(content))
	}
}

func TestSeedHistoryValidatesScenarioSnapshotCountAndEnvironment(t *testing.T) {
	binary := buildBottleneckBinary(t)
	projectDir := t.TempDir()

	unsupported := runBottleneck(t, binary, projectDir, "seed-history", "--scenario=security-regression")
	assertExitCode(t, unsupported, 1)
	assertOutputContains(t, unsupported, "unsupported seed-history scenario")

	invalidCount := runBottleneck(t, binary, projectDir, "seed-history", "--snapshots=0")
	assertExitCode(t, invalidCount, 1)
	assertOutputContains(t, invalidCount, "snapshots must be greater than 0")

	production := runBottleneck(t, binary, projectDir, "seed-history", "--env=production", "--out=bottleneck/history/production-scorecards", "--snapshots=1")
	assertExitCode(t, production, 0)
	assertOutputContains(t, production, "Environment: production")

	files, err := filepath.Glob(filepath.Join(projectDir, "bottleneck", "history", "production-scorecards", "*.json"))
	if err != nil {
		t.Fatalf("glob production seed history: %v", err)
	}
	if len(files) != 1 || !strings.Contains(filepath.Base(files[0]), "-production-") {
		t.Fatalf("expected one production snapshot, got %#v", files)
	}
}

func seedHistoryCLIFiles(t *testing.T, root string) []string {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(root, "bottleneck", "history", "scorecards", "*.json"))
	if err != nil {
		t.Fatalf("glob seed history: %v", err)
	}
	return files
}

func readSeedHistoryCLISnapshot(t *testing.T, path string) struct {
	SchemaVersion string `json:"schema_version"`
	Scorecard     struct {
		SchemaVersion string `json:"schema_version"`
	} `json:"scorecard"`
} {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read seeded snapshot: %v", err)
	}
	var decoded struct {
		SchemaVersion string `json:"schema_version"`
		Scorecard     struct {
			SchemaVersion string `json:"schema_version"`
		} `json:"scorecard"`
	}
	if err := json.Unmarshal(content, &decoded); err != nil {
		t.Fatalf("decode seeded snapshot: %v\n%s", err, string(content))
	}
	return decoded
}

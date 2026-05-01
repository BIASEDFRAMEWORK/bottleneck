package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTrendsCommandRendersTextFromSnapshots(t *testing.T) {
	binary := buildBottleneckBinary(t)
	projectDir := t.TempDir()
	writeCLISnapshot(t, projectDir, "2026-05-01T10:00:00Z", "default", "WARN", "Assurance", 40, "WARN")
	writeCLISnapshot(t, projectDir, "2026-05-02T10:00:00Z", "default", "WARN", "Assurance", 55, "WARN")

	result := runBottleneck(t, binary, projectDir, "trends")
	assertExitCode(t, result, 0)
	for _, expected := range []string{
		"Bottleneck Trends",
		"Environment: default",
		"Snapshots analyzed: 2",
		"Overall direction: Improving",
		"Persistent bottleneck:",
		"Assurance appeared as the primary bottleneck in 2 of 2 snapshots.",
	} {
		assertOutputContains(t, result, expected)
	}
}

func TestTrendsCommandRendersJSON(t *testing.T) {
	binary := buildBottleneckBinary(t)
	projectDir := t.TempDir()
	writeCLISnapshot(t, projectDir, "2026-05-01T10:00:00Z", "default", "WARN", "Assurance", 40, "WARN")
	writeCLISnapshot(t, projectDir, "2026-05-02T10:00:00Z", "default", "WARN", "Assurance", 55, "WARN")

	result := runBottleneck(t, binary, projectDir, "trends", "--format=json")
	assertExitCode(t, result, 0)
	var decoded struct {
		Environment      string `json:"environment"`
		SnapshotCount    int    `json:"snapshot_count"`
		OverallDirection string `json:"overall_direction"`
	}
	if err := json.Unmarshal([]byte(result.stdout), &decoded); err != nil {
		t.Fatalf("trend JSON did not parse: %v\nstdout:\n%s\nstderr:\n%s", err, result.stdout, result.stderr)
	}
	if decoded.Environment != "default" || decoded.SnapshotCount != 2 || decoded.OverallDirection != "improving" {
		t.Fatalf("unexpected trend JSON: %#v", decoded)
	}
}

func TestTrendsWritesMarkdownReport(t *testing.T) {
	binary := buildBottleneckBinary(t)
	projectDir := t.TempDir()
	writeCLISnapshot(t, projectDir, "2026-05-01T10:00:00Z", "default", "WARN", "Assurance", 40, "WARN")
	writeCLISnapshot(t, projectDir, "2026-05-02T10:00:00Z", "default", "WARN", "Assurance", 55, "WARN")

	result := runBottleneck(t, binary, projectDir, "trends", "--format=markdown", "--out=bottleneck/reports/trend-summary.md")
	assertExitCode(t, result, 0)
	assertOutputContains(t, result, "# Bottleneck Trend Summary")

	reportPath := filepath.Join(projectDir, "bottleneck", "reports", "trend-summary.md")
	content, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("read markdown report: %v", err)
	}
	if !strings.Contains(string(content), "## Leadership Interpretation") {
		t.Fatalf("expected markdown report content, got:\n%s", string(content))
	}
}

func TestTrendsCommandHandlesMissingHistoryDirectory(t *testing.T) {
	binary := buildBottleneckBinary(t)
	projectDir := t.TempDir()

	result := runBottleneck(t, binary, projectDir, "trends")
	assertExitCode(t, result, 0)
	assertOutputContains(t, result, "Overall direction: Insufficient history")
	assertOutputContains(t, result, "Run `bottleneck snapshot`")
}

func TestTrendsCommandValidatesWindow(t *testing.T) {
	binary := buildBottleneckBinary(t)
	projectDir := t.TempDir()

	result := runBottleneck(t, binary, projectDir, "trends", "--window=0")
	assertExitCode(t, result, 1)
	assertOutputContains(t, result, "window must be greater than 0")
}

func writeCLISnapshot(t *testing.T, root string, createdAt string, environment string, systemStatus string, primary string, assuranceScore int, assuranceStatus string) {
	t.Helper()
	history := filepath.Join(root, "bottleneck", "history", "scorecards")
	if err := os.MkdirAll(history, 0o755); err != nil {
		t.Fatalf("create history: %v", err)
	}
	content := fmt.Sprintf(`{
  "schema_version": "scorecard.snapshot.v1",
  "snapshot": {
    "created_at": %q,
    "environment": %q
  },
  "scorecard": {
    "schema_version": "scorecard.v2",
    "environment": %q,
    "system_status": %q,
    "primary_bottleneck": %q,
    "diagnosis": {
      "category_scores": [
        {"category": "Assurance", "score": %d, "status": %q}
      ]
    }
  }
}`, createdAt, environment, environment, systemStatus, primary, assuranceScore, assuranceStatus)
	filename := strings.NewReplacer(":", "", "-", "").Replace(createdAt)
	filename = strings.TrimSuffix(filename, "Z") + "-" + environment + "-scorecard.json"
	if err := os.WriteFile(filepath.Join(history, filename), []byte(content), 0o644); err != nil {
		t.Fatalf("write cli snapshot: %v", err)
	}
}

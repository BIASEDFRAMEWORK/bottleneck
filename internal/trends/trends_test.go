package trends

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTrendsInsufficientHistory(t *testing.T) {
	analysis := AnalyzeRecords([]SnapshotRecord{trendRecord("2026-05-01T10:00:00Z", "default", "WARN", "Assurance", map[string]CategoryPoint{
		"Assurance": point(40, statusWarn),
	})}, Options{Environment: "default", Window: 6})

	if analysis.OverallDirection != DirectionInsufficientHistory {
		t.Fatalf("expected insufficient history, got %#v", analysis)
	}
	if analysis.LeadershipSummary != "Not enough historical snapshots exist yet. Create snapshots over multiple delivery cycles to establish a trend." {
		t.Fatalf("unexpected leadership summary: %q", analysis.LeadershipSummary)
	}
}

func TestTrendsReadsSnapshots(t *testing.T) {
	root := t.TempDir()
	writeTrendSnapshot(t, root, "2026-05-01T10:00:00Z", "default", "WARN", "Assurance", map[string]trendScore{"Intent": {Score: 70, Status: "WARN"}})
	writeTrendSnapshot(t, root, "2026-05-02T10:00:00Z", "default", "WARN", "Assurance", map[string]trendScore{"Intent": {Score: 80, Status: "WARN"}})

	analysis, err := Analyze(Options{RootPath: root, Environment: "default", Window: 6})
	if err != nil {
		t.Fatalf("analyze trends: %v", err)
	}
	if analysis.SnapshotCount != 2 {
		t.Fatalf("expected 2 snapshots, got %#v", analysis)
	}
}

func TestTrendsReadsStableScorecardCategories(t *testing.T) {
	root := t.TempDir()
	writeStableCategorySnapshot(t, root, "2026-05-01T10:00:00Z", "default", "WARN", "Behavior", "Behavior", 45, "FAIL")
	writeStableCategorySnapshot(t, root, "2026-05-02T10:00:00Z", "default", "WARN", "Behavior", "Behavior", 70, "WARN")

	analysis, err := Analyze(Options{RootPath: root, Environment: "default", Window: 6})
	if err != nil {
		t.Fatalf("analyze stable category trends: %v", err)
	}
	behavior := findTrend(t, analysis, "Behavior")
	assertValues(t, behavior.Values, []float64{45, 70})
	if behavior.Statuses[0] != statusFail || behavior.Statuses[1] != statusWarn {
		t.Fatalf("expected statuses from stable categories, got %#v", behavior.Statuses)
	}
}

func TestTrendsSortsSnapshotsByCreatedAt(t *testing.T) {
	root := t.TempDir()
	writeTrendSnapshot(t, root, "2026-05-03T10:00:00Z", "default", "WARN", "Assurance", map[string]trendScore{"Intent": {Score: 90, Status: "PASS"}})
	writeTrendSnapshot(t, root, "2026-05-01T10:00:00Z", "default", "WARN", "Assurance", map[string]trendScore{"Intent": {Score: 70, Status: "WARN"}})
	writeTrendSnapshot(t, root, "2026-05-02T10:00:00Z", "default", "WARN", "Assurance", map[string]trendScore{"Intent": {Score: 80, Status: "WARN"}})

	analysis, err := Analyze(Options{RootPath: root, Environment: "default", Window: 6})
	if err != nil {
		t.Fatalf("analyze trends: %v", err)
	}
	intent := findTrend(t, analysis, "Intent")
	assertValues(t, intent.Values, []float64{70, 80, 90})
}

func TestTrendsFiltersByEnvironment(t *testing.T) {
	root := t.TempDir()
	writeTrendSnapshot(t, root, "2026-05-01T10:00:00Z", "default", "WARN", "Assurance", map[string]trendScore{"Intent": {Score: 70, Status: "WARN"}})
	writeTrendSnapshot(t, root, "2026-05-02T10:00:00Z", "production", "FAIL", "Security", map[string]trendScore{"Intent": {Score: 30, Status: "FAIL"}})

	analysis, err := Analyze(Options{RootPath: root, Environment: "production", Window: 6})
	if err != nil {
		t.Fatalf("analyze trends: %v", err)
	}
	if analysis.SnapshotCount != 1 || analysis.CurrentStatus != "FAIL" || analysis.CurrentPrimaryBottleneck != "Security" {
		t.Fatalf("expected only production snapshot, got %#v", analysis)
	}
}

func TestTrendsUsesLatestWindow(t *testing.T) {
	root := t.TempDir()
	writeTrendSnapshot(t, root, "2026-05-01T10:00:00Z", "default", "WARN", "Assurance", map[string]trendScore{"Intent": {Score: 10, Status: "FAIL"}})
	writeTrendSnapshot(t, root, "2026-05-02T10:00:00Z", "default", "WARN", "Assurance", map[string]trendScore{"Intent": {Score: 70, Status: "WARN"}})
	writeTrendSnapshot(t, root, "2026-05-03T10:00:00Z", "default", "WARN", "Assurance", map[string]trendScore{"Intent": {Score: 90, Status: "PASS"}})

	analysis, err := Analyze(Options{RootPath: root, Environment: "default", Window: 2})
	if err != nil {
		t.Fatalf("analyze trends: %v", err)
	}
	if analysis.SnapshotCount != 2 {
		t.Fatalf("expected latest 2 snapshots, got %#v", analysis)
	}
	intent := findTrend(t, analysis, "Intent")
	assertValues(t, intent.Values, []float64{70, 90})
}

func TestTrendsDetectsImprovingCategory(t *testing.T) {
	trend := directionTestTrend([]float64{40, 55}, []string{statusWarn, statusWarn})
	if trend.Direction != DirectionImproving {
		t.Fatalf("expected improving, got %#v", trend)
	}
}

func TestTrendsDetectsDecliningCategory(t *testing.T) {
	trend := directionTestTrend([]float64{80, 65}, []string{statusWarn, statusWarn})
	if trend.Direction != DirectionDeclining {
		t.Fatalf("expected declining, got %#v", trend)
	}
}

func TestTrendsDetectsStableCategory(t *testing.T) {
	trend := directionTestTrend([]float64{80, 85}, []string{statusWarn, statusWarn})
	if trend.Direction != DirectionStable {
		t.Fatalf("expected stable, got %#v", trend)
	}
}

func TestTrendsDetectsRecoveredCategory(t *testing.T) {
	trend := directionTestTrend([]float64{40, 100}, []string{statusWarn, statusPass})
	if trend.Direction != DirectionRecovered {
		t.Fatalf("expected recovered, got %#v", trend)
	}
}

func TestTrendsDetectsRegressedCategory(t *testing.T) {
	trend := directionTestTrend([]float64{100, 60}, []string{statusPass, statusWarn})
	if trend.Direction != DirectionRegressed {
		t.Fatalf("expected regressed, got %#v", trend)
	}
}

func TestTrendsIdentifiesPersistentBottleneck(t *testing.T) {
	records := []SnapshotRecord{
		trendRecord("2026-05-01T10:00:00Z", "default", "WARN", "Assurance", map[string]CategoryPoint{"Intent": point(70, statusWarn)}),
		trendRecord("2026-05-02T10:00:00Z", "default", "WARN", "Assurance", map[string]CategoryPoint{"Intent": point(80, statusWarn)}),
		trendRecord("2026-05-03T10:00:00Z", "default", "WARN", "Security", map[string]CategoryPoint{"Intent": point(90, statusPass)}),
	}
	analysis := AnalyzeRecords(records, Options{Environment: "default", Window: 6})
	if analysis.PersistentBottleneck.Category != "Assurance" || analysis.PersistentBottleneck.Count != 2 || analysis.PersistentBottleneck.Total != 3 {
		t.Fatalf("expected persistent Assurance bottleneck, got %#v", analysis.PersistentBottleneck)
	}
	if analysis.PersistentBottleneck.Summary != "Assurance appeared as the primary bottleneck in 2 of 3 snapshots." {
		t.Fatalf("unexpected persistent summary: %q", analysis.PersistentBottleneck.Summary)
	}
}

func TestTrendsHandlesMissingHistoryDirectory(t *testing.T) {
	analysis, err := Analyze(Options{RootPath: t.TempDir(), Environment: "default", Window: 6})
	if err != nil {
		t.Fatalf("missing history should not error: %v", err)
	}
	if analysis.SnapshotCount != 0 || analysis.OverallDirection != DirectionInsufficientHistory {
		t.Fatalf("expected insufficient history for missing history directory, got %#v", analysis)
	}
}

func TestTrendsIgnoresMalformedSnapshotsWithWarning(t *testing.T) {
	root := t.TempDir()
	history := filepath.Join(root, "bottleneck", "history", "scorecards")
	if err := os.MkdirAll(history, 0o755); err != nil {
		t.Fatalf("create history: %v", err)
	}
	if err := os.WriteFile(filepath.Join(history, "bad.json"), []byte("{not-json"), 0o644); err != nil {
		t.Fatalf("write malformed snapshot: %v", err)
	}
	writeTrendSnapshot(t, root, "2026-05-01T10:00:00Z", "default", "WARN", "Assurance", map[string]trendScore{"Intent": {Score: 70, Status: "WARN"}})

	analysis, err := Analyze(Options{RootPath: root, Environment: "default", Window: 6})
	if err != nil {
		t.Fatalf("malformed snapshot should be ignored with warning: %v", err)
	}
	if analysis.SnapshotCount != 1 || len(analysis.Warnings) != 1 {
		t.Fatalf("expected one valid snapshot and one warning, got %#v", analysis)
	}
}

func directionTestTrend(values []float64, statuses []string) CategoryTrend {
	records := make([]SnapshotRecord, 0, len(values))
	for i, value := range values {
		status := statuses[i]
		records = append(records, trendRecord(
			fmt.Sprintf("2026-05-%02dT10:00:00Z", i+1),
			"default",
			status,
			"Assurance",
			map[string]CategoryPoint{"Intent": point(value, status)},
		))
	}
	analysis := AnalyzeRecords(records, Options{Environment: "default", Window: 6})
	for _, trend := range analysis.CategoryTrends {
		if trend.Category == "Intent" {
			return trend
		}
	}
	return CategoryTrend{}
}

type trendScore struct {
	Score  float64
	Status string
}

func writeTrendSnapshot(t *testing.T, root string, createdAt string, environment string, systemStatus string, primary string, scores map[string]trendScore) {
	t.Helper()
	history := filepath.Join(root, "bottleneck", "history", "scorecards")
	if err := os.MkdirAll(history, 0o755); err != nil {
		t.Fatalf("create history: %v", err)
	}
	categoryScores := ""
	index := 0
	for category, score := range scores {
		if index > 0 {
			categoryScores += ","
		}
		categoryScores += fmt.Sprintf(`{"category":%q,"score":%v,"status":%q}`, category, score.Score, score.Status)
		index++
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
      "category_scores": [%s]
    }
  }
}`, createdAt, environment, environment, systemStatus, primary, categoryScores)
	filename := fmt.Sprintf("%s-%s-scorecard.json", stringsForFilename(createdAt), environment)
	if err := os.WriteFile(filepath.Join(history, filename), []byte(content), 0o644); err != nil {
		t.Fatalf("write snapshot: %v", err)
	}
}

func writeStableCategorySnapshot(t *testing.T, root string, createdAt string, environment string, systemStatus string, primary string, category string, score float64, status string) {
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
    "categories": [
      {"name": %q, "score": %v, "status": %q, "summary": "stable contract category"}
    ]
  }
}`, createdAt, environment, environment, systemStatus, primary, category, score, status)
	filename := fmt.Sprintf("%s-%s-stable-scorecard.json", stringsForFilename(createdAt), environment)
	if err := os.WriteFile(filepath.Join(history, filename), []byte(content), 0o644); err != nil {
		t.Fatalf("write stable category snapshot: %v", err)
	}
}

func trendRecord(createdAt string, environment string, systemStatus string, primary string, categories map[string]CategoryPoint) SnapshotRecord {
	parsed, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		panic(err)
	}
	return SnapshotRecord{
		CreatedAt:         parsed,
		Environment:       environment,
		SystemStatus:      systemStatus,
		PrimaryBottleneck: primary,
		Categories:        categories,
	}
}

func point(score float64, status string) CategoryPoint {
	return CategoryPoint{Score: &score, Status: status}
}

func findTrend(t *testing.T, analysis Analysis, category string) CategoryTrend {
	t.Helper()
	for _, trend := range analysis.CategoryTrends {
		if trend.Category == category {
			return trend
		}
	}
	t.Fatalf("expected trend for %s in %#v", category, analysis.CategoryTrends)
	return CategoryTrend{}
}

func assertValues(t *testing.T, actual []float64, expected []float64) {
	t.Helper()
	if len(actual) != len(expected) {
		t.Fatalf("expected values %#v, got %#v", expected, actual)
	}
	for i := range expected {
		if actual[i] != expected[i] {
			t.Fatalf("expected values %#v, got %#v", expected, actual)
		}
	}
}

func stringsForFilename(createdAt string) string {
	return strings.NewReplacer(":", "", "-", "", "T", "T").Replace(strings.TrimSuffix(createdAt, ":00Z")) + "Z"
}

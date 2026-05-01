package seedhistory

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"bottleneck/internal/diagnosis"
	"bottleneck/internal/scorecard"
	"bottleneck/internal/snapshot"
	"bottleneck/internal/trends"
)

func TestSeedHistoryCreatesSnapshots(t *testing.T) {
	root := t.TempDir()
	result := writeSeedHistory(t, root, Options{})

	if len(result.Snapshots) != 6 {
		t.Fatalf("expected 6 snapshots, got %d", len(result.Snapshots))
	}
	files := seedFiles(t, root)
	if len(files) != 6 {
		t.Fatalf("expected 6 snapshot files, got %d: %#v", len(files), files)
	}
	if !strings.Contains(filepath.Base(files[0]), "01-fast-demo-weak-evidence") ||
		!strings.Contains(filepath.Base(files[5]), "06-stable-release-candidate") {
		t.Fatalf("expected scenario labels in sorted filenames: %#v", files)
	}

	latest := readSeedSnapshot(t, files[5])
	if latest.Scorecard.SystemStatus != scorecard.StatusPass ||
		latest.Scorecard.PrimaryBottleneck != diagnosis.HealthyPrimaryBottleneck ||
		latest.Scorecard.BottomLine != "The delivery system can now prove intent, behavior, assurance, security, and execution evidence together." {
		t.Fatalf("unexpected final seed snapshot: %#v", latest.Scorecard)
	}
}

func TestSeedHistoryUsesSnapshotSchema(t *testing.T) {
	root := t.TempDir()
	result := writeSeedHistory(t, root, Options{SnapshotCount: 1})
	decoded := readSeedSnapshot(t, result.Snapshots[0].SnapshotPath)

	if decoded.SchemaVersion != snapshot.SchemaVersion {
		t.Fatalf("expected snapshot schema %q, got %q", snapshot.SchemaVersion, decoded.SchemaVersion)
	}
	if decoded.Scorecard.SchemaVersion != scorecard.SchemaVersion {
		t.Fatalf("expected scorecard schema %q, got %q", scorecard.SchemaVersion, decoded.Scorecard.SchemaVersion)
	}
	if decoded.Snapshot.Source != snapshot.Source {
		t.Fatalf("expected snapshot source %q, got %q", snapshot.Source, decoded.Snapshot.Source)
	}
	if len(decoded.Scorecard.Diagnosis.CategoryScores) != 6 || len(decoded.Scorecard.Capabilities) != 6 {
		t.Fatalf("expected six category scores and capabilities, got %#v", decoded.Scorecard)
	}
}

func TestSeedHistoryDoesNotOverwriteByDefault(t *testing.T) {
	root := t.TempDir()
	writeSeedHistory(t, root, Options{})

	_, err := Write(seedOptions(root, Options{}))
	if err == nil || !strings.Contains(err.Error(), "Seed history already exists") {
		t.Fatalf("expected existing history error, got %v", err)
	}
}

func TestSeedHistoryOverwriteFlag(t *testing.T) {
	root := t.TempDir()
	writeSeedHistory(t, root, Options{})

	result, err := Write(seedOptions(root, Options{Overwrite: true}))
	if err != nil {
		t.Fatalf("overwrite seed history: %v", err)
	}
	if len(result.Snapshots) != 6 {
		t.Fatalf("expected overwritten snapshot count, got %d", len(result.Snapshots))
	}
}

func TestSeedHistoryOverwriteReplacesPriorSeedFiles(t *testing.T) {
	root := t.TempDir()
	writeSeedHistory(t, root, Options{Now: time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)})
	writeSeedHistory(t, root, Options{Now: time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC), Overwrite: true})

	files := seedFiles(t, root)
	if len(files) != 6 {
		t.Fatalf("expected overwrite to leave 6 seeded snapshots, got %d: %#v", len(files), files)
	}
	for _, file := range files {
		if strings.Contains(filepath.Base(file), "2026-05-01") {
			t.Fatalf("expected prior seed file to be replaced, got %#v", files)
		}
	}
}

func TestSeedHistoryOverwriteWithSmallerSnapshotCountRemovesPriorSeedFiles(t *testing.T) {
	root := t.TempDir()
	writeSeedHistory(t, root, Options{})
	writeSeedHistory(t, root, Options{SnapshotCount: 3, Overwrite: true})

	files := seedFiles(t, root)
	if len(files) != 3 {
		t.Fatalf("expected overwrite with smaller count to leave 3 seeded snapshots, got %d: %#v", len(files), files)
	}
	if strings.Contains(strings.Join(files, "\n"), "04-tests-added-security-regresses") ||
		strings.Contains(strings.Join(files, "\n"), "06-stable-release-candidate") {
		t.Fatalf("expected later seed files to be removed, got %#v", files)
	}
}

func TestSeedHistoryWorksWithTrends(t *testing.T) {
	root := t.TempDir()
	writeSeedHistory(t, root, Options{})

	analysis, err := trends.Analyze(trends.Options{RootPath: root, Environment: DefaultEnvironment, Window: 6})
	if err != nil {
		t.Fatalf("analyze seeded trends: %v", err)
	}
	if analysis.SnapshotCount != 6 ||
		analysis.CurrentStatus != scorecard.StatusPass ||
		analysis.CurrentPrimaryBottleneck != diagnosis.HealthyPrimaryBottleneck ||
		analysis.OverallDirection != trends.DirectionImproving {
		t.Fatalf("unexpected seeded trend analysis: %#v", analysis)
	}
}

func TestSeedHistoryScenarioSaaSDayOne(t *testing.T) {
	root := t.TempDir()
	writeSeedHistory(t, root, Options{})

	files := seedFiles(t, root)
	expected := []struct {
		status  string
		primary string
		intent  string
	}{
		{scorecard.StatusFail, "Intent", scorecard.StatusFail},
		{scorecard.StatusFail, "Behavior", scorecard.StatusPass},
		{scorecard.StatusWarn, "Assurance", scorecard.StatusPass},
		{scorecard.StatusWarn, "Security", scorecard.StatusPass},
		{scorecard.StatusWarn, "Execution", scorecard.StatusPass},
		{scorecard.StatusPass, diagnosis.HealthyPrimaryBottleneck, scorecard.StatusPass},
	}
	for index, expectation := range expected {
		decoded := readSeedSnapshot(t, files[index])
		if decoded.Scorecard.SystemStatus != expectation.status || decoded.Scorecard.PrimaryBottleneck != expectation.primary {
			t.Fatalf("snapshot %d expected status %s primary %s, got %#v", index+1, expectation.status, expectation.primary, decoded.Scorecard)
		}
		if status := statusFor(decoded, "Intent"); status != expectation.intent {
			t.Fatalf("snapshot %d expected Intent %s, got %s", index+1, expectation.intent, status)
		}
	}
}

func TestSeedHistoryUnsupportedScenarioReturnsUsefulError(t *testing.T) {
	_, err := Write(seedOptions(t.TempDir(), Options{Scenario: "security-regression"}))
	if err == nil || !strings.Contains(err.Error(), "unsupported seed-history scenario") {
		t.Fatalf("expected unsupported scenario error, got %v", err)
	}
}

func TestSeedHistoryRejectsSnapshotCountLessThanOne(t *testing.T) {
	_, err := Write(seedOptions(t.TempDir(), Options{SnapshotCount: -1}))
	if err == nil || !strings.Contains(err.Error(), "snapshots must be greater than 0") {
		t.Fatalf("expected snapshots validation error, got %v", err)
	}
}

func TestSeedHistoryProductionEnvironmentMetadata(t *testing.T) {
	root := t.TempDir()
	result := writeSeedHistory(t, root, Options{Environment: "production", SnapshotCount: 1})
	decoded := readSeedSnapshot(t, result.Snapshots[0].SnapshotPath)

	if decoded.Snapshot.Environment != "production" || decoded.Scorecard.Environment != "production" {
		t.Fatalf("expected production metadata, got %#v", decoded)
	}
	if !strings.Contains(filepath.Base(result.Snapshots[0].SnapshotPath), "-production-") {
		t.Fatalf("expected production filename, got %s", result.Snapshots[0].SnapshotPath)
	}
}

func TestSeedHistoryOutputFilesAreSortedByTimestamp(t *testing.T) {
	root := t.TempDir()
	writeSeedHistory(t, root, Options{})
	files := seedFiles(t, root)

	createdAt := make([]string, 0, len(files))
	for _, file := range files {
		decoded := readSeedSnapshot(t, file)
		createdAt = append(createdAt, decoded.Snapshot.CreatedAt)
	}
	if !sort.StringsAreSorted(createdAt) {
		t.Fatalf("expected seed snapshots sorted by timestamp, got %#v", createdAt)
	}
}

func writeSeedHistory(t *testing.T, root string, options Options) Result {
	t.Helper()
	result, err := Write(seedOptions(root, options))
	if err != nil {
		t.Fatalf("write seed history: %v", err)
	}
	return result
}

func seedOptions(root string, options Options) Options {
	options.RootPath = root
	if options.Now.IsZero() {
		options.Now = time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	}
	return options
}

func seedFiles(t *testing.T, root string) []string {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(root, snapshot.DefaultScorecardDir, "*.json"))
	if err != nil {
		t.Fatalf("glob seed history: %v", err)
	}
	sort.Strings(files)
	return files
}

func readSeedSnapshot(t *testing.T, path string) snapshot.Snapshot {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read seed snapshot %s: %v", path, err)
	}
	var decoded snapshot.Snapshot
	if err := json.Unmarshal(content, &decoded); err != nil {
		t.Fatalf("decode seed snapshot %s: %v\n%s", path, err, string(content))
	}
	return decoded
}

func statusFor(decoded snapshot.Snapshot, category string) string {
	for _, score := range decoded.Scorecard.Diagnosis.CategoryScores {
		if score.Category == category {
			return score.Status
		}
	}
	return ""
}

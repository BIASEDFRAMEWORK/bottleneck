package snapshot

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"bottleneck/internal/gitinfo"
	"bottleneck/internal/scorecard"
)

func TestSnapshotCreatesHistoryDirectory(t *testing.T) {
	root := t.TempDir()
	result := writeTestSnapshot(t, root, Options{})

	info, err := os.Stat(filepath.Join(root, DefaultScorecardDir))
	if err != nil {
		t.Fatalf("expected history directory: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("expected %s to be a directory", filepath.Join(root, DefaultScorecardDir))
	}
	if result.SnapshotPath == "" {
		t.Fatal("expected snapshot path")
	}
}

func TestSnapshotWritesTimestampedFile(t *testing.T) {
	root := t.TempDir()
	result := writeTestSnapshot(t, root, Options{})

	expectedSuffix := filepath.Join("bottleneck", "history", "scorecards", "2026-05-01T141500Z-default-scorecard.json")
	if !strings.HasSuffix(result.SnapshotPath, expectedSuffix) {
		t.Fatalf("expected timestamped snapshot path ending %s, got %s", expectedSuffix, result.SnapshotPath)
	}
	assertFileExists(t, result.SnapshotPath)
}

func TestSnapshotWritesLatestFile(t *testing.T) {
	root := t.TempDir()
	result := writeTestSnapshot(t, root, Options{})

	expectedLatest := filepath.Join(root, "bottleneck", "history", "latest", "default.json")
	if result.LatestPath != expectedLatest {
		t.Fatalf("expected latest path %s, got %s", expectedLatest, result.LatestPath)
	}
	assertFileExists(t, expectedLatest)
}

func TestSnapshotNoLatestFlagSkipsLatestFile(t *testing.T) {
	root := t.TempDir()
	result := writeTestSnapshot(t, root, Options{NoLatest: true})

	if result.LatestPath != "" {
		t.Fatalf("expected no latest path, got %s", result.LatestPath)
	}
	if _, err := os.Stat(filepath.Join(root, "bottleneck", "history", "latest", "default.json")); !os.IsNotExist(err) {
		t.Fatalf("expected latest file to be skipped, stat err: %v", err)
	}
}

func TestSnapshotIncludesEnvironment(t *testing.T) {
	root := t.TempDir()
	result := writeTestSnapshot(t, root, Options{Environment: "production"})
	decoded := readSnapshot(t, result.SnapshotPath)

	if decoded.Snapshot.Environment != "production" {
		t.Fatalf("expected production snapshot metadata, got %#v", decoded.Snapshot)
	}
	if decoded.Scorecard.Environment != "production" {
		t.Fatalf("expected production scorecard, got %#v", decoded.Scorecard)
	}
	if !strings.HasSuffix(result.SnapshotPath, "2026-05-01T141500Z-production-scorecard.json") {
		t.Fatalf("expected production filename, got %s", result.SnapshotPath)
	}
}

func TestSnapshotIncludesLabel(t *testing.T) {
	root := t.TempDir()
	result := writeTestSnapshot(t, root, Options{Label: "release-candidate"})
	decoded := readSnapshot(t, result.SnapshotPath)

	if decoded.Snapshot.Label != "release-candidate" {
		t.Fatalf("expected label in metadata, got %#v", decoded.Snapshot)
	}
	if !strings.HasSuffix(result.SnapshotPath, "2026-05-01T141500Z-default-release-candidate-scorecard.json") {
		t.Fatalf("expected label in filename, got %s", result.SnapshotPath)
	}
}

func TestSnapshotSanitizesLabel(t *testing.T) {
	tests := map[string]string{
		"release candidate": "release-candidate",
		"Release_Candidate": "release-candidate",
		"prod/rc1":          "prod-rc1",
	}
	for input, expected := range tests {
		if actual := SanitizeLabel(input); actual != expected {
			t.Fatalf("expected sanitized label %q for %q, got %q", expected, input, actual)
		}
	}
}

func TestSnapshotJSONIncludesSchemaMetadataGitAndScorecard(t *testing.T) {
	root := t.TempDir()
	dirty := true
	result := writeTestSnapshot(t, root, Options{
		Git: gitinfo.Info{
			Commit: "abc1234",
			Branch: "main",
			Dirty:  &dirty,
		},
	})
	decoded := readSnapshot(t, result.SnapshotPath)

	if decoded.SchemaVersion != SchemaVersion {
		t.Fatalf("expected schema version %q, got %q", SchemaVersion, decoded.SchemaVersion)
	}
	if decoded.Snapshot.ID != "SNAPSHOT-20260501-141500" {
		t.Fatalf("expected deterministic snapshot id, got %q", decoded.Snapshot.ID)
	}
	if decoded.Snapshot.CreatedAt != "2026-05-01T14:15:00Z" {
		t.Fatalf("expected UTC created_at, got %q", decoded.Snapshot.CreatedAt)
	}
	if decoded.Snapshot.Source != Source {
		t.Fatalf("expected source %q, got %q", Source, decoded.Snapshot.Source)
	}
	if decoded.Snapshot.Git.Commit != "abc1234" || decoded.Snapshot.Git.Branch != "main" || decoded.Snapshot.Git.Dirty == nil || !*decoded.Snapshot.Git.Dirty {
		t.Fatalf("expected git metadata, got %#v", decoded.Snapshot.Git)
	}
	if decoded.Scorecard.SchemaVersion != scorecard.SchemaVersion || decoded.Scorecard.SystemStatus != scorecard.StatusWarn {
		t.Fatalf("expected scorecard data in snapshot, got %#v", decoded.Scorecard)
	}
}

func TestScorecardJSONBackwardsCompatibleForSnapshot(t *testing.T) {
	root := t.TempDir()
	result := writeTestSnapshot(t, root, Options{})
	decoded := readSnapshot(t, result.SnapshotPath)

	if decoded.Scorecard.SchemaVersion != scorecard.SchemaVersion ||
		decoded.Scorecard.Environment != "default" ||
		decoded.Scorecard.SystemStatus != scorecard.StatusWarn ||
		decoded.Scorecard.PrimaryBottleneck != "Assurance" {
		t.Fatalf("expected legacy scorecard fields in snapshot, got %#v", decoded.Scorecard)
	}
	if len(decoded.Scorecard.Categories) == 0 {
		t.Fatalf("expected stable categories in snapshot scorecard, got %#v", decoded.Scorecard)
	}
	category := decoded.Scorecard.Categories[0]
	if category.Name != "Assurance" ||
		category.Status != scorecard.StatusWarn ||
		category.Score == 0 ||
		category.EvidenceFound == nil ||
		category.EvidenceMissing == nil ||
		category.Recommendations == nil {
		t.Fatalf("expected stable category contract in snapshot, got %#v", category)
	}
}

func TestSnapshotDoesNotFailOutsideGitRepo(t *testing.T) {
	root := t.TempDir()
	result := writeTestSnapshot(t, root, Options{Git: gitinfo.Detect(root)})
	decoded := readSnapshot(t, result.SnapshotPath)

	if decoded.Snapshot.Git.Commit != "" || decoded.Snapshot.Git.Branch != "" || decoded.Snapshot.Git.Dirty != nil {
		t.Fatalf("expected empty git metadata outside git repo, got %#v", decoded.Snapshot.Git)
	}
}

func writeTestSnapshot(t *testing.T, root string, options Options) WriteResult {
	t.Helper()
	options.RootPath = root
	if options.CreatedAt.IsZero() {
		options.CreatedAt = time.Date(2026, 5, 1, 14, 15, 0, 0, time.UTC)
	}
	card := testScorecard(options.Environment)
	result, err := Write(card, options)
	if err != nil {
		t.Fatalf("write snapshot: %v", err)
	}
	return result
}

func testScorecard(environment string) scorecard.Scorecard {
	if environment == "" {
		environment = "default"
	}
	return scorecard.Scorecard{
		SchemaVersion:         scorecard.SchemaVersion,
		Environment:           environment,
		SystemStatus:          scorecard.StatusWarn,
		ReleaseRecommendation: scorecard.RecommendationConditional,
		PrimaryBottleneck:     "Assurance",
		Capabilities: []scorecard.CapabilityScorecard{{
			Capability:        "Assurance",
			Status:            scorecard.StatusWarn,
			Score:             65,
			Reason:            "Assurance evidence is incomplete.",
			Evidence:          []string{"partial assurance evidence"},
			MissingEvidence:   []string{"Map tests to behavior evidence."},
			RecommendedAction: "Map validation evidence to behavior evidence.",
		}},
	}
}

func readSnapshot(t *testing.T, path string) Snapshot {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read snapshot %s: %v", path, err)
	}
	var decoded Snapshot
	if err := json.Unmarshal(content, &decoded); err != nil {
		t.Fatalf("decode snapshot %s: %v\n%s", path, err, string(content))
	}
	return decoded
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("expected file %s: %v", path, err)
	}
	if info.IsDir() {
		t.Fatalf("expected file %s, got directory", path)
	}
}

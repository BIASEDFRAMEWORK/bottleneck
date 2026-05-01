package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSnapshotCommandCreatesHistoryFiles(t *testing.T) {
	binary := buildBottleneckBinary(t)
	projectDir := t.TempDir()

	initResult := runBottleneck(t, binary, projectDir, "init", "--template", "saas")
	assertExitCode(t, initResult, 0)

	result := runBottleneck(t, binary, projectDir, "snapshot", "--env=dev", "--label=release candidate")
	assertExitCode(t, result, 0)
	assertOutputContains(t, result, "Bottleneck snapshot created")
	assertOutputContains(t, result, "Environment: dev")
	assertOutputContains(t, result, "Status: WARN")
	assertOutputContains(t, result, "Primary bottleneck: Assurance")
	assertOutputContains(t, result, "Snapshot: bottleneck/history/scorecards/")
	assertOutputContains(t, result, "Latest: bottleneck/history/latest/dev.json")

	files, err := filepath.Glob(filepath.Join(projectDir, "bottleneck", "history", "scorecards", "*-dev-release-candidate-scorecard.json"))
	if err != nil {
		t.Fatalf("glob snapshots: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected one dev release-candidate snapshot, got %d: %#v", len(files), files)
	}

	decoded := readSnapshotJSON(t, files[0])
	if decoded.SchemaVersion != "scorecard.snapshot.v1" {
		t.Fatalf("expected snapshot schema, got %#v", decoded)
	}
	if decoded.Snapshot.Environment != "dev" || decoded.Snapshot.Label != "release-candidate" {
		t.Fatalf("expected dev release-candidate metadata, got %#v", decoded.Snapshot)
	}
	if decoded.Scorecard.Environment != "dev" || decoded.Scorecard.SystemStatus != "WARN" {
		t.Fatalf("expected dev warning scorecard, got %#v", decoded.Scorecard)
	}

	latest := filepath.Join(projectDir, "bottleneck", "history", "latest", "dev.json")
	latestSnapshot := readSnapshotJSON(t, latest)
	if latestSnapshot.Snapshot.ID != decoded.Snapshot.ID {
		t.Fatalf("expected latest snapshot to match timestamped snapshot, got %#v and %#v", latestSnapshot.Snapshot, decoded.Snapshot)
	}
}

func TestSnapshotCommandWritesFailingScorecardWithoutFailingExit(t *testing.T) {
	binary := buildBottleneckBinary(t)
	projectDir := t.TempDir()

	initResult := runBottleneck(t, binary, projectDir, "init")
	assertExitCode(t, initResult, 0)

	scorecardResult := runBottleneck(t, binary, projectDir, "scorecard")
	assertExitCode(t, scorecardResult, 1)

	result := runBottleneck(t, binary, projectDir, "snapshot")
	assertExitCode(t, result, 0)
	assertOutputContains(t, result, "Status: FAIL")

	files, err := filepath.Glob(filepath.Join(projectDir, "bottleneck", "history", "scorecards", "*-default-scorecard.json"))
	if err != nil {
		t.Fatalf("glob snapshots: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected one default snapshot, got %d: %#v", len(files), files)
	}
	decoded := readSnapshotJSON(t, files[0])
	if decoded.Scorecard.SystemStatus != "FAIL" || decoded.Scorecard.PrimaryBottleneck != "Assurance" {
		t.Fatalf("expected failing scorecard snapshot, got %#v", decoded.Scorecard)
	}
}

func TestSnapshotCommandProductionAndNoLatest(t *testing.T) {
	binary := buildBottleneckBinary(t)
	projectDir := t.TempDir()

	initResult := runBottleneck(t, binary, projectDir, "init", "--template", "saas")
	assertExitCode(t, initResult, 0)

	result := runBottleneck(t, binary, projectDir, "snapshot", "--env=production", "--no-latest")
	assertExitCode(t, result, 0)
	assertOutputContains(t, result, "Environment: production")
	assertOutputContains(t, result, "Latest: skipped")

	files, err := filepath.Glob(filepath.Join(projectDir, "bottleneck", "history", "scorecards", "*-production-scorecard.json"))
	if err != nil {
		t.Fatalf("glob production snapshots: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected one production snapshot, got %d: %#v", len(files), files)
	}
	if _, err := os.Stat(filepath.Join(projectDir, "bottleneck", "history", "latest", "production.json")); !os.IsNotExist(err) {
		t.Fatalf("expected production latest to be skipped, stat err: %v", err)
	}
}

func TestSnapshotDoesNotBreakExistingScorecardCommand(t *testing.T) {
	binary := buildBottleneckBinary(t)
	projectDir := t.TempDir()

	initResult := runBottleneck(t, binary, projectDir, "init", "--template", "saas")
	assertExitCode(t, initResult, 0)

	snapshotResult := runBottleneck(t, binary, projectDir, "snapshot", "--env=dev")
	assertExitCode(t, snapshotResult, 0)

	scorecardResult := runBottleneck(t, binary, projectDir, "scorecard", "--env=dev")
	assertExitCode(t, scorecardResult, 0)
	assertOutputContains(t, scorecardResult, "Bottleneck Scorecard")
	assertOutputContains(t, scorecardResult, "Environment: dev")
	assertOutputContains(t, scorecardResult, "Release Recommendation: Conditional")
}

func TestSnapshotCommandFailsWhenConfigCannotBeRead(t *testing.T) {
	binary := buildBottleneckBinary(t)
	projectDir := t.TempDir()

	result := runBottleneck(t, binary, projectDir, "snapshot")
	assertExitCode(t, result, 1)
	assertOutputContains(t, result, "No Bottleneck config found.")
	if _, err := os.Stat(filepath.Join(projectDir, "bottleneck", "history")); !os.IsNotExist(err) {
		t.Fatalf("snapshot should not write history for missing config, stat err: %v", err)
	}
}

func readSnapshotJSON(t *testing.T, path string) struct {
	SchemaVersion string `json:"schema_version"`
	Snapshot      struct {
		ID          string `json:"id"`
		Environment string `json:"environment"`
		Label       string `json:"label"`
		Git         struct {
			Commit string `json:"commit"`
			Branch string `json:"branch"`
			Dirty  *bool  `json:"dirty"`
		} `json:"git"`
	} `json:"snapshot"`
	Scorecard struct {
		SchemaVersion     string `json:"schema_version"`
		Environment       string `json:"environment"`
		SystemStatus      string `json:"system_status"`
		PrimaryBottleneck string `json:"primary_bottleneck"`
	} `json:"scorecard"`
} {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read snapshot %s: %v", path, err)
	}
	var decoded struct {
		SchemaVersion string `json:"schema_version"`
		Snapshot      struct {
			ID          string `json:"id"`
			Environment string `json:"environment"`
			Label       string `json:"label"`
			Git         struct {
				Commit string `json:"commit"`
				Branch string `json:"branch"`
				Dirty  *bool  `json:"dirty"`
			} `json:"git"`
		} `json:"snapshot"`
		Scorecard struct {
			SchemaVersion     string `json:"schema_version"`
			Environment       string `json:"environment"`
			SystemStatus      string `json:"system_status"`
			PrimaryBottleneck string `json:"primary_bottleneck"`
		} `json:"scorecard"`
	}
	if err := json.Unmarshal(content, &decoded); err != nil {
		t.Fatalf("snapshot JSON did not parse: %v\n%s", err, string(content))
	}
	if strings.TrimSpace(decoded.Snapshot.ID) == "" {
		t.Fatalf("expected snapshot id in %s: %#v", path, decoded)
	}
	return decoded
}

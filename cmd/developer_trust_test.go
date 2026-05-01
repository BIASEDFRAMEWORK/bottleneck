package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiscoverAndAutoIngestCLI(t *testing.T) {
	binary := buildBottleneckBinary(t)
	projectDir := t.TempDir()
	writeCLIFile(t, projectDir, "reports/junit.xml", `<testsuite tests="1" failures="0"><testcase classname="Checkout BEHAVIOR-003" name="happy path"></testcase></testsuite>`)
	writeCLIFile(t, projectDir, "coverage/lcov.info", "TN:\nSF:checkout.go\nDA:1,1\nDA:2,1\nend_of_record\n")
	writeCLIFile(t, projectDir, ".github/workflows/ci.yml", "name: ci\n")
	writeCLIFile(t, projectDir, "README.md", "# Service\n")

	discoverResult := runBottleneck(t, binary, projectDir, "discover", "--format=json")
	assertExitCode(t, discoverResult, 0)
	var discovery struct {
		Findings []struct {
			Kind             string `json:"kind"`
			Path             string `json:"path"`
			SuggestedCommand string `json:"suggested_command"`
		} `json:"findings"`
		Summary struct {
			TotalFindings int `json:"total_findings"`
		} `json:"summary"`
	}
	if err := json.Unmarshal([]byte(discoverResult.stdout), &discovery); err != nil {
		t.Fatalf("discover JSON did not parse: %v\n%s", err, discoverResult.stdout)
	}
	if discovery.Summary.TotalFindings < 4 {
		t.Fatalf("expected discovered evidence, got %#v", discovery)
	}

	dryRun := runBottleneck(t, binary, projectDir, "ingest", "--auto", "--dry-run", "--format=json")
	assertExitCode(t, dryRun, 0)
	if _, err := os.Stat(filepath.Join(projectDir, "bottleneck/assurance/results.json")); !os.IsNotExist(err) {
		t.Fatalf("dry-run should not write assurance output, stat err=%v", err)
	}

	syncResult := runBottleneck(t, binary, projectDir, "evidence", "sync", "--format=json")
	assertExitCode(t, syncResult, 0)
	var sync struct {
		Written []struct {
			Kind       string `json:"kind"`
			OutputPath string `json:"output_path"`
		} `json:"written"`
	}
	if err := json.Unmarshal([]byte(syncResult.stdout), &sync); err != nil {
		t.Fatalf("evidence sync JSON did not parse: %v\n%s", err, syncResult.stdout)
	}
	if len(sync.Written) == 0 || sync.Written[0].OutputPath != "bottleneck/assurance/results.json" {
		t.Fatalf("expected evidence sync to write assurance evidence, got %#v", sync)
	}
	assertFileContains(t, filepath.Join(projectDir, "bottleneck/assurance/results.json"), "generated_by")
}

func TestAssessAndExplainScoreJSON(t *testing.T) {
	binary := buildBottleneckBinary(t)
	projectDir := t.TempDir()

	initResult := runBottleneck(t, binary, projectDir, "init", "--template", "saas")
	assertExitCode(t, initResult, 0)

	checkResult := runBottleneck(t, binary, projectDir, "check")
	assertExitCode(t, checkResult, 0)
	assertOutputContains(t, checkResult, "System Status:")

	assessResult := runBottleneck(t, binary, projectDir, "assess", "--no-ingest", "--format=json")
	assertExitCode(t, assessResult, 0)
	var assessment struct {
		SchemaVersion         string `json:"schema_version"`
		AIReadiness           string `json:"ai_readiness"`
		ReleaseFriction       string `json:"release_friction"`
		PrimaryBottleneck     string `json:"primary_bottleneck"`
		ScoreConfidence       string `json:"score_confidence"`
		ReleaseRecommendation string `json:"release_recommendation"`
		Maturity              struct {
			Level int    `json:"level"`
			Label string `json:"label"`
		} `json:"maturity"`
		Categories []struct {
			Name       string `json:"name"`
			Provenance string `json:"provenance"`
		} `json:"categories"`
	}
	if err := json.Unmarshal([]byte(assessResult.stdout), &assessment); err != nil {
		t.Fatalf("assessment JSON did not parse: %v\n%s", err, assessResult.stdout)
	}
	if assessment.SchemaVersion != "assessment.v1" || assessment.Maturity.Label == "" || assessment.PrimaryBottleneck == "" || len(assessment.Categories) == 0 {
		t.Fatalf("unexpected assessment JSON: %#v", assessment)
	}

	explainResult := runBottleneck(t, binary, projectDir, "explain-score", "--format=json")
	assertExitCode(t, explainResult, 0)
	assertOutputContains(t, explainResult, `"score_rationale"`)
	assertOutputContains(t, explainResult, `"provenance"`)

	maturityResult := runBottleneck(t, binary, projectDir, "maturity", "--no-ingest")
	assertExitCode(t, maturityResult, 0)
	assertOutputContains(t, maturityResult, "Bottleneck Assessment")
}

func writeCLIFile(t *testing.T, root string, rel string, content string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", rel, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
}

func assertFileContains(t *testing.T, path string, expected string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !strings.Contains(string(content), expected) {
		t.Fatalf("expected %s to contain %q\n%s", path, expected, string(content))
	}
}

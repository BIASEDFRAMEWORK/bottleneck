package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIFlowInitValidateDiagnoseScorecardAndTrace(t *testing.T) {
	binary := buildBottleneckBinary(t)
	projectDir := t.TempDir()

	initResult := runBottleneck(t, binary, projectDir, "init")
	assertExitCode(t, initResult, 0)
	assertOutputContains(t, initResult, "Bottleneck initialized.")
	assertOutputContains(t, initResult, "intentionally leaves Assurance weak")

	validateResult := runBottleneck(t, binary, projectDir, "validate")
	assertExitCode(t, validateResult, 1)
	assertOutputContains(t, validateResult, "Assurance: FAIL")
	assertOutputContains(t, validateResult, "Primary Bottleneck: Assurance")

	diagnoseResult := runBottleneck(t, binary, projectDir, "diagnose", "--format=json")
	assertExitCode(t, diagnoseResult, 1)
	var diagnosis struct {
		SchemaVersion     string `json:"schema_version"`
		SystemStatus      string `json:"system_status"`
		PrimaryBottleneck string `json:"primary_bottleneck"`
		Confidence        string `json:"confidence"`
	}
	if err := json.Unmarshal([]byte(diagnoseResult.stdout), &diagnosis); err != nil {
		t.Fatalf("diagnose JSON did not parse: %v\n%s", err, diagnoseResult.stdout)
	}
	if diagnosis.SchemaVersion != "diagnose.v1" || diagnosis.SystemStatus != "FAIL" || diagnosis.PrimaryBottleneck != "Assurance" || diagnosis.Confidence != "High" {
		t.Fatalf("unexpected diagnosis JSON: %#v", diagnosis)
	}

	scorecardResult := runBottleneck(t, binary, projectDir, "scorecard", "--format=json")
	assertExitCode(t, scorecardResult, 1)
	var scorecard struct {
		SchemaVersion         string `json:"schema_version"`
		SystemStatus          string `json:"system_status"`
		ReleaseRecommendation string `json:"release_recommendation"`
		PrimaryBottleneck     string `json:"primary_bottleneck"`
		EffectiveThresholds   struct {
			Security struct {
				SARIF struct {
					MaxMedium int `json:"max_medium"`
				} `json:"sarif"`
			} `json:"security"`
			Execution struct {
				Telemetry struct {
					MaxAgeHours int `json:"max_age_hours"`
				} `json:"telemetry"`
			} `json:"execution"`
		} `json:"effective_thresholds"`
	}
	if err := json.Unmarshal([]byte(scorecardResult.stdout), &scorecard); err != nil {
		t.Fatalf("scorecard JSON did not parse: %v\n%s", err, scorecardResult.stdout)
	}
	if scorecard.SchemaVersion != "scorecard.v2" || scorecard.ReleaseRecommendation != "Block" || scorecard.PrimaryBottleneck != "Assurance" {
		t.Fatalf("unexpected scorecard JSON: %#v", scorecard)
	}
	if scorecard.EffectiveThresholds.Security.SARIF.MaxMedium != 5 || scorecard.EffectiveThresholds.Execution.Telemetry.MaxAgeHours != 168 {
		t.Fatalf("expected effective SARIF and telemetry thresholds in scorecard JSON, got %#v", scorecard.EffectiveThresholds)
	}

	traceResult := runBottleneck(t, binary, projectDir, "trace", "--id", "BEHAVIOR-001", "--format=json")
	assertExitCode(t, traceResult, 0)
	assertOutputContains(t, traceResult, `"query": "BEHAVIOR-001"`)
	assertOutputContains(t, traceResult, `"found": true`)
	assertOutputContains(t, traceResult, "ASSURANCE-001")

	executiveResult := runBottleneck(t, binary, projectDir, "scorecard", "--view=executive")
	assertExitCode(t, executiveResult, 1)
	assertOutputContains(t, executiveResult, "Executive View")
	assertOutputContains(t, executiveResult, "Release Recommendation: Block")
	if strings.Contains(executiveResult.stdout, "failure: Ambiguous risk clause was summarized as confirmed exposure") {
		t.Fatalf("executive view should stay concise and omit detailed evidence:\n%s", executiveResult.stdout)
	}

	governanceResult := runBottleneck(t, binary, projectDir, "scorecard", "--view=governance")
	assertExitCode(t, governanceResult, 1)
	assertOutputContains(t, governanceResult, "Governance View")
	assertOutputContains(t, governanceResult, "Governance Evidence: not assessed")
}

func TestCLIIngestCucumberDryRunDoesNotWriteArtifacts(t *testing.T) {
	binary := buildBottleneckBinary(t)
	projectDir := t.TempDir()
	inputPath := filepath.Join(projectDir, "cucumber.json")
	content := `[
  {
    "name": "Risk feature",
    "elements": [
      {
        "name": "Ambiguous risk clause is flagged",
        "type": "scenario",
        "tags": [{"name": "@BEHAVIOR-001"}],
        "steps": [{"result": {"status": "passed"}}]
      }
    ]
  }
]`
	if err := os.WriteFile(inputPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write cucumber input: %v", err)
	}

	result := runBottleneck(t, binary, projectDir, "ingest", "cucumber", "--file", inputPath, "--dry-run", "--format=json")
	assertExitCode(t, result, 0)
	assertOutputContains(t, result, `"scenarios_total": 1`)
	assertOutputContains(t, result, `"BEHAVIOR-001"`)

	if _, err := os.Stat(filepath.Join(projectDir, "bottleneck", "assurance", "results.json")); !os.IsNotExist(err) {
		t.Fatalf("dry-run should not write assurance artifact, stat err=%v", err)
	}
}

func TestCLITraceRequiresID(t *testing.T) {
	binary := buildBottleneckBinary(t)
	projectDir := t.TempDir()
	initResult := runBottleneck(t, binary, projectDir, "init")
	assertExitCode(t, initResult, 0)

	result := runBottleneck(t, binary, projectDir, "trace")
	assertExitCode(t, result, 1)
	assertOutputContains(t, result, "trace id required")
}

type cliResult struct {
	stdout   string
	stderr   string
	exitCode int
}

func buildBottleneckBinary(t *testing.T) string {
	t.Helper()

	binary := filepath.Join(t.TempDir(), "bottleneck")
	command := exec.Command("go", "build", "-o", binary, ".")
	command.Dir = ".."
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("build bottleneck binary: %v\n%s", err, string(output))
	}
	return binary
}

func runBottleneck(t *testing.T, binary string, dir string, args ...string) cliResult {
	t.Helper()

	command := exec.Command(binary, args...)
	command.Dir = dir
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	exitCode := 0
	if err := command.Run(); err != nil {
		exitError, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("run bottleneck %s: %v\nstdout:\n%s\nstderr:\n%s", strings.Join(args, " "), err, stdout.String(), stderr.String())
		}
		exitCode = exitError.ExitCode()
	}

	return cliResult{stdout: stdout.String(), stderr: stderr.String(), exitCode: exitCode}
}

func assertExitCode(t *testing.T, result cliResult, expected int) {
	t.Helper()
	if result.exitCode != expected {
		t.Fatalf("expected exit code %d, got %d\nstdout:\n%s\nstderr:\n%s", expected, result.exitCode, result.stdout, result.stderr)
	}
}

func assertOutputContains(t *testing.T, result cliResult, expected string) {
	t.Helper()
	combined := result.stdout + result.stderr
	if !strings.Contains(combined, expected) {
		t.Fatalf("expected output to contain %q\nstdout:\n%s\nstderr:\n%s", expected, result.stdout, result.stderr)
	}
}

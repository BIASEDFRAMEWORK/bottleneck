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
	if scorecard.EffectiveThresholds.Security.SARIF.MaxMedium != 5 || scorecard.EffectiveThresholds.Execution.Telemetry.MaxAgeHours != 0 {
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

func TestCLISaaSScorecardDefaultAndDetails(t *testing.T) {
	binary := buildBottleneckBinary(t)
	projectDir := t.TempDir()

	initResult := runBottleneck(t, binary, projectDir, "init", "--template", "saas")
	assertExitCode(t, initResult, 0)

	scorecardResult := runBottleneck(t, binary, projectDir, "scorecard", "--env=dev")
	assertExitCode(t, scorecardResult, 0)
	for _, substring := range []string{
		"Bottleneck Scorecard",
		"Environment: dev",
		"Release Recommendation: Conditional",
		"Primary Bottleneck: Assurance",
		"Effective Thresholds:",
		"- Minimum score: 65",
		"- Required traceability: false",
		"- Critical security findings allowed: 1",
		"- Stale telemetry allowed: true",
		"Category Results:",
		"- Intent: Pass",
		"- Behavior: Pass",
		"- Design: Pass",
		"- Assurance: Warn",
		"- Security: Pass",
		"- Execution: Pass",
		"Why:",
		"BEHAVIOR-003 payment retry behavior has no mapped test evidence.",
		"Next Action:",
		"Add assurance evidence for payment retry behavior. Map it to BEHAVIOR-003.",
	} {
		assertOutputContains(t, scorecardResult, substring)
	}
	for _, verboseSection := range []string{"Capability Details:", "Score Impacts:"} {
		if strings.Contains(scorecardResult.stdout, verboseSection) {
			t.Fatalf("default scorecard should be concise and omit %q:\n%s", verboseSection, scorecardResult.stdout)
		}
	}

	detailsResult := runBottleneck(t, binary, projectDir, "scorecard", "--env=dev", "--details")
	assertExitCode(t, detailsResult, 0)
	for _, substring := range []string{
		"Effective Thresholds:",
		"gate.release.min_primary_score: 65",
		"gate.release.require_traceability: false",
		"Capability Details:",
		"Missing Evidence:",
		"Score Impacts:",
		"bottleneck/behavior/behavior-spec.md BEHAVIOR-003 has no mapped test evidence",
	} {
		assertOutputContains(t, detailsResult, substring)
	}

	unsupportedFlagResult := runBottleneck(t, binary, projectDir, "scorecard", "--summary")
	assertExitCode(t, unsupportedFlagResult, 1)
	assertOutputContains(t, unsupportedFlagResult, "unknown flag: --summary")
}

func TestCLISaaSEnvironmentDefaultsAndReleaseGate(t *testing.T) {
	binary := buildBottleneckBinary(t)
	projectDir := t.TempDir()

	initResult := runBottleneck(t, binary, projectDir, "init", "--template", "saas")
	assertExitCode(t, initResult, 0)

	devGate := runBottleneck(t, binary, projectDir, "diagnose", "--env=dev", "--gate=release")
	assertExitCode(t, devGate, 0)
	assertOutputContains(t, devGate, "Release Gate: PASS")

	productionScorecard := runBottleneck(t, binary, projectDir, "scorecard", "--env=production")
	assertExitCode(t, productionScorecard, 1)
	for _, substring := range []string{
		"Environment: production",
		"Effective Thresholds:",
		"- Minimum score: 85",
		"- Required traceability: true",
		"- Critical security findings allowed: 0",
		"- Stale telemetry allowed: false",
		"Release Recommendation: Block",
	} {
		assertOutputContains(t, productionScorecard, substring)
	}

	productionGate := runBottleneck(t, binary, projectDir, "diagnose", "--env=production", "--gate=release")
	assertExitCode(t, productionGate, 1)
	assertOutputContains(t, productionGate, "Release Gate: FAIL")
	assertOutputContains(t, productionGate, "Traceability is broken.")

	unknownEnv := runBottleneck(t, binary, projectDir, "scorecard", "--env=not-real")
	assertExitCode(t, unknownEnv, 1)
	assertOutputContains(t, unknownEnv, `unknown environment "not-real"`)
	assertOutputContains(t, unknownEnv, "supported: default, dev, local, production, stage, test")
	assertOutputContains(t, unknownEnv, "bottleneck scorecard --env=production")
}

func TestCLIMissingConfigGuidesInitialization(t *testing.T) {
	binary := buildBottleneckBinary(t)
	projectDir := t.TempDir()

	result := runBottleneck(t, binary, projectDir, "scorecard")
	assertExitCode(t, result, 1)
	assertOutputContains(t, result, "No Bottleneck config found.")
	assertOutputContains(t, result, "Bottleneck has not been initialized")
	assertOutputContains(t, result, "bottleneck init --template saas")
}

func TestCLISaaSDiagnoseIsActionableAcrossFormats(t *testing.T) {
	binary := buildBottleneckBinary(t)
	projectDir := t.TempDir()

	initResult := runBottleneck(t, binary, projectDir, "init", "--template", "saas")
	assertExitCode(t, initResult, 0)

	textResult := runBottleneck(t, binary, projectDir, "diagnose")
	assertExitCode(t, textResult, 0)
	for _, substring := range []string{
		"Primary Bottleneck: Assurance",
		"Reason: BEHAVIOR-003 is not linked to any passing test evidence.",
		"Impact: Release confidence is reduced because payment retry behavior is unproven.",
		"Next Action: Add or ingest test evidence mapped to BEHAVIOR-003.",
		"Inspect: bottleneck trace BEHAVIOR-003",
		"Relevant Evidence: BEHAVIOR-003",
		"Supporting Issues:",
	} {
		assertOutputContains(t, textResult, substring)
	}

	jsonResult := runBottleneck(t, binary, projectDir, "diagnose", "--format=json")
	assertExitCode(t, jsonResult, 0)
	var report struct {
		SchemaVersion       string   `json:"schema_version"`
		PrimaryBottleneck   string   `json:"primary_bottleneck"`
		Reason              string   `json:"reason"`
		Impact              string   `json:"impact"`
		NextAction          string   `json:"next_action"`
		InspectCommand      string   `json:"inspect_command"`
		RelevantEvidenceIDs []string `json:"relevant_evidence_ids"`
	}
	if err := json.Unmarshal([]byte(jsonResult.stdout), &report); err != nil {
		t.Fatalf("diagnose JSON did not parse: %v\n%s", err, jsonResult.stdout)
	}
	if report.SchemaVersion != "diagnose.v1" ||
		report.PrimaryBottleneck != "Assurance" ||
		report.Reason != "BEHAVIOR-003 is not linked to any passing test evidence." ||
		report.Impact != "Release confidence is reduced because payment retry behavior is unproven." ||
		report.NextAction != "Add or ingest test evidence mapped to BEHAVIOR-003." ||
		report.InspectCommand != "bottleneck trace BEHAVIOR-003" ||
		len(report.RelevantEvidenceIDs) != 1 ||
		report.RelevantEvidenceIDs[0] != "BEHAVIOR-003" {
		t.Fatalf("unexpected actionable diagnose JSON: %#v", report)
	}

	markdownResult := runBottleneck(t, binary, projectDir, "diagnose", "--format=markdown")
	assertExitCode(t, markdownResult, 0)
	for _, substring := range []string{
		"### Reason",
		"BEHAVIOR-003 is not linked to any passing test evidence.",
		"### Impact",
		"Release confidence is reduced because payment retry behavior is unproven.",
		"### Recommended Next Action",
		"Add or ingest test evidence mapped to BEHAVIOR-003.",
		"### Inspect",
		"`bottleneck trace BEHAVIOR-003`",
	} {
		assertOutputContains(t, markdownResult, substring)
	}
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

func TestCLIIngestInvalidFileIncludesSampleFormatGuidance(t *testing.T) {
	binary := buildBottleneckBinary(t)
	projectDir := t.TempDir()
	inputPath := filepath.Join(projectDir, "bad-cucumber.json")
	if err := os.WriteFile(inputPath, []byte("not valid json"), 0o644); err != nil {
		t.Fatalf("write invalid input: %v", err)
	}

	result := runBottleneck(t, binary, projectDir, "ingest", "cucumber", "--file", inputPath, "--dry-run")
	assertExitCode(t, result, 1)
	assertOutputContains(t, result, "parse cucumber json")
	assertOutputContains(t, result, inputPath)
	assertOutputContains(t, result, "Next action: check the expected sample format in examples/saas/reports/cucumber.json")
}

func TestCLISaaSSampleCucumberIngestionImprovesScorecard(t *testing.T) {
	binary := buildBottleneckBinary(t)
	projectDir := t.TempDir()

	initResult := runBottleneck(t, binary, projectDir, "init", "--template", "saas")
	assertExitCode(t, initResult, 0)

	baseline := runBottleneck(t, binary, projectDir, "scorecard")
	assertExitCode(t, baseline, 0)
	assertOutputContains(t, baseline, "Release Recommendation: Conditional")
	assertOutputContains(t, baseline, "Primary Bottleneck: Assurance")
	assertOutputContains(t, baseline, "- Assurance: Warn")
	assertOutputContains(t, baseline, "BEHAVIOR-003 payment retry behavior has no mapped test evidence")

	ingestResult := runBottleneck(t, binary, projectDir, "ingest", "cucumber", "--file", cliSampleReportPath(t, "cucumber.json"))
	assertExitCode(t, ingestResult, 0)
	assertOutputContains(t, ingestResult, "assurance artifact: 3 scenarios (3 passed, 0 failed)")
	assertOutputContains(t, ingestResult, "Wrote bottleneck/assurance/results.json")

	assuranceContent, err := os.ReadFile(filepath.Join(projectDir, "bottleneck", "assurance", "results.json"))
	if err != nil {
		t.Fatalf("read ingested assurance artifact: %v", err)
	}
	if !strings.Contains(string(assuranceContent), "ASSURANCE-003") || !strings.Contains(string(assuranceContent), "BEHAVIOR-003") {
		t.Fatalf("expected ingested assurance artifact to link ASSURANCE-003 to BEHAVIOR-003:\n%s", string(assuranceContent))
	}

	updated := runBottleneck(t, binary, projectDir, "scorecard")
	assertExitCode(t, updated, 0)
	assertOutputContains(t, updated, "Release Recommendation: Proceed")
	assertOutputContains(t, updated, "Primary Bottleneck: None")
	assertOutputContains(t, updated, "- Assurance: Pass")
	if strings.Contains(updated.stdout, "no mapped test evidence") {
		t.Fatalf("updated scorecard should not report missing mapped test evidence:\n%s", updated.stdout)
	}

	traceResult := runBottleneck(t, binary, projectDir, "trace", "BEHAVIOR-003")
	assertExitCode(t, traceResult, 0)
	assertOutputContains(t, traceResult, "ASSURANCE-003")
	if strings.Contains(traceResult.stdout, "BEHAVIOR-003 has no mapped test evidence") {
		t.Fatalf("trace should show BEHAVIOR-003 covered after ingestion:\n%s", traceResult.stdout)
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

func cliSampleReportPath(t *testing.T, name string) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join("..", "examples", "saas", "reports", name))
	if err != nil {
		t.Fatalf("resolve sample report path %s: %v", name, err)
	}
	return path
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

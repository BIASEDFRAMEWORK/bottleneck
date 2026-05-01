package ingest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalizeCucumberCreatesAssuranceArtifact(t *testing.T) {
	tempDir := t.TempDir()
	filePath := fixturePath("cucumber-failing.json")

	result, err := IngestCucumber(tempDir, filePath, "", false, false)
	if err != nil {
		t.Fatalf("ingest cucumber: %v", err)
	}
	artifact, ok := result.Artifact.(AssuranceArtifact)
	if !ok {
		t.Fatalf("expected AssuranceArtifact, got %T", result.Artifact)
	}
	if artifact.ScenariosTotal != 1 {
		t.Fatalf("expected 1 scenario, got %d", artifact.ScenariosTotal)
	}
	if artifact.ScenariosPassed != 0 || artifact.ScenariosFailed != 1 {
		t.Fatalf("unexpected pass/fail counts: %d/%d", artifact.ScenariosPassed, artifact.ScenariosFailed)
	}
	if len(artifact.Evidence) != 1 {
		t.Fatalf("expected 1 evidence entry, got %d", len(artifact.Evidence))
	}
	if len(artifact.Evidence[0].Refs) != 1 || artifact.Evidence[0].Refs[0] != "BEHAVIOR-001" {
		t.Fatalf("expected BEHAVIOR-001 ref, got %#v", artifact.Evidence[0].Refs)
	}
	if len(result.Warnings) != 0 {
		t.Fatalf("expected no warnings, got %v", result.Warnings)
	}
}

func TestNormalizeCucumberReportsUnmappedAndUnmatchedBehavior(t *testing.T) {
	tempDir := t.TempDir()
	writeBehaviorSpec(t, tempDir, "BEHAVIOR-999")

	result, err := IngestCucumber(tempDir, fixturePath("cucumber-unmapped.json"), "", false, true)
	if err != nil {
		t.Fatalf("ingest cucumber: %v", err)
	}

	if !containsString(result.Warnings, `unmapped scenario "Scenario without behavior tag" has no BEHAVIOR-* tags`) {
		t.Fatalf("expected unmapped scenario warning, got %#v", result.Warnings)
	}
	if !containsString(result.Warnings, "behavior BEHAVIOR-999 has no matching Cucumber scenario evidence") {
		t.Fatalf("expected unmatched behavior warning, got %#v", result.Warnings)
	}
}

func TestNormalizeCucumberTreatsSkippedMissingAndMalformedAsNonPassing(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "cucumber.json")
	content := `[
  {
    "name": "Checkout feature",
    "elements": [
      {
        "name": "Skipped scenario",
        "type": "scenario",
        "tags": [{"name": "@BEHAVIOR-001"}],
        "steps": [{"result": {"status": "skipped"}}]
      },
      {
        "name": "Missing status scenario",
        "type": "scenario",
        "tags": [{"name": "@BEHAVIOR-002"}],
        "steps": [{"result": {}}]
      },
      {
        "name": "No steps scenario",
        "type": "scenario",
        "tags": [{"name": "@BEHAVIOR-003"}],
        "steps": []
      }
    ]
  }
]`
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatalf("write cucumber file: %v", err)
	}

	result, err := IngestCucumber(tempDir, filePath, "", false, true)
	if err != nil {
		t.Fatalf("ingest cucumber: %v", err)
	}
	artifact := result.Artifact.(AssuranceArtifact)
	if artifact.ScenariosTotal != 3 || artifact.ScenariosPassed != 0 || artifact.ScenariosFailed != 3 {
		t.Fatalf("expected all scenarios non-passing, got %#v", artifact)
	}

	emptyPath := filepath.Join(tempDir, "empty-cucumber.json")
	if err := os.WriteFile(emptyPath, []byte(`[]`), 0o644); err != nil {
		t.Fatalf("write empty cucumber file: %v", err)
	}
	emptyResult, err := IngestCucumber(tempDir, emptyPath, "", false, true)
	if err != nil {
		t.Fatalf("ingest empty cucumber: %v", err)
	}
	if !containsString(emptyResult.Warnings, "no Cucumber scenarios found") {
		t.Fatalf("expected empty Cucumber warning, got %#v", emptyResult.Warnings)
	}

	invalidPath := filepath.Join(tempDir, "invalid-cucumber.json")
	if err := os.WriteFile(invalidPath, []byte(`not valid json`), 0o644); err != nil {
		t.Fatalf("write invalid cucumber file: %v", err)
	}
	if _, err := IngestCucumber(tempDir, invalidPath, "", false, true); err == nil {
		t.Fatal("expected malformed Cucumber input to return an error")
	}
	if _, err := IngestCucumber(tempDir, filepath.Join(tempDir, "missing.json"), "", false, true); err == nil {
		t.Fatal("expected missing Cucumber input to return an error")
	}
}

func TestNormalizeCodeQLCreatesSecurityArtifact(t *testing.T) {
	tempDir := t.TempDir()
	filePath := fixturePath("sarif-mixed.sarif")

	artifact, err := IngestCodeQL(tempDir, filePath, "", false, false)
	if err != nil {
		t.Fatalf("ingest codeql: %v", err)
	}
	security, ok := artifact.Artifact.(SecurityArtifact)
	if !ok {
		t.Fatalf("expected SecurityArtifact, got %T", artifact.Artifact)
	}
	if security.Violations != 5 {
		t.Fatalf("expected 5 violations, got %d", security.Violations)
	}
	if security.Findings["high"] != 1 {
		t.Fatalf("expected high finding count 1, got %d", security.Findings["high"])
	}
	if security.Findings["critical"] != 1 || security.Findings["medium"] != 1 || security.Findings["low"] != 1 || security.Findings["unknown"] != 1 {
		t.Fatalf("unexpected severity counts: %#v", security.Findings)
	}
	if security.Evidence[0].RuleName != "SQL injection" || security.Evidence[0].Line == nil || *security.Evidence[0].Line != 42 {
		t.Fatalf("expected SARIF metadata preserved, got %#v", security.Evidence[0])
	}
}

func TestNormalizeSARIFNoFindingsAndCodeQLSeverity(t *testing.T) {
	tempDir := t.TempDir()
	empty, err := IngestSARIF(tempDir, fixturePath("sarif-none.sarif"), "", false, true)
	if err != nil {
		t.Fatalf("ingest empty sarif: %v", err)
	}
	emptyArtifact := empty.Artifact.(SecurityArtifact)
	if emptyArtifact.Violations != 0 || emptyArtifact.Evidence[0].Status != "pass" {
		t.Fatalf("expected passing no-findings SARIF artifact, got %#v", emptyArtifact)
	}

	codeql, err := IngestSARIF(tempDir, fixturePath("sarif-codeql.sarif"), "", false, true)
	if err != nil {
		t.Fatalf("ingest codeql sarif: %v", err)
	}
	codeqlArtifact := codeql.Artifact.(SecurityArtifact)
	if codeqlArtifact.Findings["high"] != 1 {
		t.Fatalf("expected CodeQL security-severity high, got %#v", codeqlArtifact.Findings)
	}
}

func TestNormalizeSARIFPreservesBottleneckRefsFromProperties(t *testing.T) {
	tempDir := t.TempDir()
	result, err := IngestSARIF(tempDir, sampleReportPath("codeql.sarif"), "", false, true)
	if err != nil {
		t.Fatalf("ingest sample sarif: %v", err)
	}
	artifact := result.Artifact.(SecurityArtifact)
	if len(artifact.Evidence) != 1 {
		t.Fatalf("expected one sample SARIF evidence item, got %d", len(artifact.Evidence))
	}
	refs := artifact.Evidence[0].Refs
	if !containsString(refs, "BEHAVIOR-003") || !containsString(refs, "INTENT-001") {
		t.Fatalf("expected SARIF refs to preserve BEHAVIOR-003 and INTENT-001, got %#v", refs)
	}
}

func TestSaaSReportSamplesParseAndWriteNormalizedEvidence(t *testing.T) {
	t.Run("cucumber", func(t *testing.T) {
		rootPath := t.TempDir()
		result, err := IngestCucumber(rootPath, sampleReportPath("cucumber.json"), "", false, false)
		if err != nil {
			t.Fatalf("ingest sample cucumber: %v", err)
		}
		artifact := result.Artifact.(AssuranceArtifact)
		if artifact.ScenariosTotal != 3 || artifact.ScenariosPassed != 3 || artifact.ScenariosFailed != 0 {
			t.Fatalf("unexpected sample cucumber counts: %#v", artifact)
		}
		if len(artifact.Evidence) != 3 || artifact.Evidence[2].ID != "ASSURANCE-003" || !containsString(artifact.Evidence[2].Refs, "BEHAVIOR-003") {
			t.Fatalf("expected generated ASSURANCE-003 linked to BEHAVIOR-003, got %#v", artifact.Evidence)
		}
		assertWrittenArtifactContains(t, rootPath, DefaultAssuranceOutput, "ASSURANCE-003", "BEHAVIOR-003")
	})

	t.Run("sarif", func(t *testing.T) {
		rootPath := t.TempDir()
		result, err := IngestSARIF(rootPath, sampleReportPath("codeql.sarif"), "", false, false)
		if err != nil {
			t.Fatalf("ingest sample sarif: %v", err)
		}
		artifact := result.Artifact.(SecurityArtifact)
		if artifact.Findings["low"] != 1 || artifact.Violations != 1 {
			t.Fatalf("expected one low SARIF finding, got %#v", artifact)
		}
		if len(artifact.Evidence) != 1 || artifact.Evidence[0].ID != "SECURITY-001" || !containsString(artifact.Evidence[0].Refs, "BEHAVIOR-003") {
			t.Fatalf("expected generated SECURITY-001 linked to BEHAVIOR-003, got %#v", artifact.Evidence)
		}
		assertWrittenArtifactContains(t, rootPath, DefaultSecurityOutput, "SECURITY-001", "BEHAVIOR-003")
	})

	t.Run("test summary", func(t *testing.T) {
		rootPath := t.TempDir()
		result, err := IngestTestSummary(rootPath, sampleReportPath("test-summary.json"), "", false, false)
		if err != nil {
			t.Fatalf("ingest sample test summary: %v", err)
		}
		artifact := result.Artifact.(AssuranceArtifact)
		if artifact.ScenariosTotal != 12 || artifact.ScenariosPassed != 12 || artifact.ScenariosFailed != 0 {
			t.Fatalf("unexpected sample test summary counts: %#v", artifact)
		}
		if len(artifact.Evidence) != 3 || artifact.Evidence[2].ID != "ASSURANCE-003" || !containsString(artifact.Evidence[2].Refs, "BEHAVIOR-003") {
			t.Fatalf("expected preserved ASSURANCE-003 linked to BEHAVIOR-003, got %#v", artifact.Evidence)
		}
		assertWrittenArtifactContains(t, rootPath, DefaultAssuranceOutput, "ASSURANCE-003", "BEHAVIOR-003")
	})

	t.Run("telemetry", func(t *testing.T) {
		rootPath := t.TempDir()
		result, err := IngestTelemetry(rootPath, sampleReportPath("telemetry.json"), "", false, false)
		if err != nil {
			t.Fatalf("ingest sample telemetry: %v", err)
		}
		artifact := result.Artifact.(ExecutionArtifact)
		if artifact.DeploymentFrequency == nil || artifact.DeploymentFrequency.Deployments != 6 {
			t.Fatalf("expected sample deployment frequency, got %#v", artifact.DeploymentFrequency)
		}
		if len(artifact.Evidence) != 1 || artifact.Evidence[0].ID != "EXECUTION-001" || !containsString(artifact.Evidence[0].Refs, "BEHAVIOR-003") {
			t.Fatalf("expected preserved EXECUTION-001 linked to BEHAVIOR-003, got %#v", artifact.Evidence)
		}
		assertWrittenArtifactContains(t, rootPath, DefaultExecutionOutput, "EXECUTION-001", "BEHAVIOR-003")
	})
}

func TestSaaSReportSampleDryRunsDoNotWriteNormalizedEvidence(t *testing.T) {
	cases := []struct {
		name string
		file string
		out  string
		run  func(rootPath, filePath string) (IngestSummary, error)
	}{
		{
			name: "cucumber",
			file: "cucumber.json",
			out:  DefaultAssuranceOutput,
			run: func(rootPath, filePath string) (IngestSummary, error) {
				return IngestCucumber(rootPath, filePath, "", false, true)
			},
		},
		{
			name: "sarif",
			file: "codeql.sarif",
			out:  DefaultSecurityOutput,
			run: func(rootPath, filePath string) (IngestSummary, error) {
				return IngestSARIF(rootPath, filePath, "", false, true)
			},
		},
		{
			name: "test-summary",
			file: "test-summary.json",
			out:  DefaultAssuranceOutput,
			run: func(rootPath, filePath string) (IngestSummary, error) {
				return IngestTestSummary(rootPath, filePath, "", false, true)
			},
		},
		{
			name: "telemetry",
			file: "telemetry.json",
			out:  DefaultExecutionOutput,
			run: func(rootPath, filePath string) (IngestSummary, error) {
				return IngestTelemetry(rootPath, filePath, "", false, true)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rootPath := t.TempDir()
			if _, err := tc.run(rootPath, sampleReportPath(tc.file)); err != nil {
				t.Fatalf("dry-run sample %s: %v", tc.name, err)
			}
			if _, err := os.Stat(filepath.Join(rootPath, tc.out)); !os.IsNotExist(err) {
				t.Fatalf("dry-run %s should not write %s, stat err=%v", tc.name, tc.out, err)
			}
		})
	}
}

func TestSaaSReportInvalidInputsFailCleanly(t *testing.T) {
	cases := []struct {
		name string
		run  func(rootPath, filePath string) (IngestSummary, error)
	}{
		{
			name: "cucumber",
			run: func(rootPath, filePath string) (IngestSummary, error) {
				return IngestCucumber(rootPath, filePath, "", false, true)
			},
		},
		{
			name: "sarif",
			run: func(rootPath, filePath string) (IngestSummary, error) {
				return IngestSARIF(rootPath, filePath, "", false, true)
			},
		},
		{
			name: "test-summary",
			run: func(rootPath, filePath string) (IngestSummary, error) {
				return IngestTestSummary(rootPath, filePath, "", false, true)
			},
		},
		{
			name: "telemetry",
			run: func(rootPath, filePath string) (IngestSummary, error) {
				return IngestTelemetry(rootPath, filePath, "", false, true)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rootPath := t.TempDir()
			invalidPath := filepath.Join(rootPath, tc.name+".json")
			if err := os.WriteFile(invalidPath, []byte("not valid json"), 0o644); err != nil {
				t.Fatalf("write invalid %s input: %v", tc.name, err)
			}
			_, err := tc.run(rootPath, invalidPath)
			if err == nil {
				t.Fatalf("expected invalid %s input to return an error", tc.name)
			}
			if !strings.Contains(err.Error(), "parse") {
				t.Fatalf("expected invalid %s error to mention parse, got %v", tc.name, err)
			}
			for _, substring := range []string{
				invalidPath,
				"Next action: check the expected sample format",
				"examples/saas/reports/",
			} {
				if !strings.Contains(err.Error(), substring) {
					t.Fatalf("expected invalid %s error to contain %q, got %v", tc.name, substring, err)
				}
			}
		})
	}
}

func TestNormalizeTestSummaryCreatesAssuranceArtifact(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "summary.json")
	content := `{
  "tests_total": 4,
  "tests_passed": 3,
  "tests_failed": 1,
  "tests_skipped": 0,
  "coverage": 0.85,
  "source": "go test ./...",
  "evidence": [
    {"id": "ASSURANCE-001", "status": "pass", "type": "test-summary", "source": "go test"}
  ]
}`
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatalf("write summary file: %v", err)
	}

	artifact, err := IngestTestSummary(tempDir, filePath, "", false, false)
	if err != nil {
		t.Fatalf("ingest test-summary: %v", err)
	}
	assurance, ok := artifact.Artifact.(AssuranceArtifact)
	if !ok {
		t.Fatalf("expected AssuranceArtifact, got %T", artifact.Artifact)
	}
	if assurance.ScenariosTotal != 4 || assurance.ScenariosFailed != 1 {
		t.Fatalf("unexpected assurance totals: %d/%d", assurance.ScenariosTotal, assurance.ScenariosFailed)
	}
	if len(assurance.Evidence) != 1 {
		t.Fatalf("expected evidence count 1, got %d", len(assurance.Evidence))
	}
}

func TestNormalizeTelemetryCreatesExecutionArtifact(t *testing.T) {
	tempDir := t.TempDir()
	filePath := fixturePath("telemetry-healthy.json")

	artifact, err := IngestTelemetry(tempDir, filePath, "", false, false)
	if err != nil {
		t.Fatalf("ingest telemetry: %v", err)
	}
	execution, ok := artifact.Artifact.(ExecutionArtifact)
	if !ok {
		t.Fatalf("expected ExecutionArtifact, got %T", artifact.Artifact)
	}
	if execution.AdoptionRate != 0.72 || execution.ErrorRate != 0.01 {
		t.Fatalf("unexpected telemetry values: %.2f/%.2f", execution.AdoptionRate, execution.ErrorRate)
	}
	if execution.DeploymentFrequency == nil || execution.DeploymentFrequency.Deployments != 7 {
		t.Fatalf("expected deployment frequency, got %#v", execution.DeploymentFrequency)
	}
	if len(execution.Evidence) != 1 || execution.Evidence[0].ID != "EXECUTION-001" {
		t.Fatalf("expected default execution evidence, got %#v", execution.Evidence)
	}
}

func TestIngestTelemetryNormalizesPercentageRates(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "telemetry-percentages.json")
	content := `{
  "deployment_frequency": {"deployments": 7, "period_days": 7},
  "change_failure_rate": 5,
  "error_rate": 2,
  "user_override_rate": 3,
  "adoption_rate": 72,
  "cost": {"budget_variance": 20}
}`
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatalf("write telemetry file: %v", err)
	}

	result, err := IngestTelemetry(tempDir, filePath, "", false, true)
	if err != nil {
		t.Fatalf("ingest telemetry: %v", err)
	}
	artifact := result.Artifact.(ExecutionArtifact)
	if artifact.ChangeFailureRate != 0.05 || artifact.ErrorRate != 0.02 || artifact.AdoptionRate != 0.72 {
		t.Fatalf("expected normalized core rates, got %#v", artifact)
	}
	if artifact.UserOverrideRate == nil || *artifact.UserOverrideRate != 0.03 {
		t.Fatalf("expected normalized user override rate, got %#v", artifact.UserOverrideRate)
	}
	if artifact.Cost == nil || artifact.Cost.BudgetVariance != 0.20 {
		t.Fatalf("expected normalized budget variance, got %#v", artifact.Cost)
	}
}

func TestIngestTelemetryInvalidInputReturnsError(t *testing.T) {
	_, err := IngestTelemetry(t.TempDir(), fixturePath("telemetry-invalid.json"), "", false, true)
	if err == nil {
		t.Fatal("expected error for invalid telemetry input")
	}
}

func TestIngestCucumberDryRunDoesNotWriteArtifact(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "cucumber.json")
	content := `[
  {
    "name": "Checkout feature",
    "elements": [
      {
        "name": "Valid checkout",
        "type": "scenario",
        "tags": [{"name": "@BEHAVIOR-001"}],
        "steps": [
          {"result": {"status": "passed"}}
        ]
      }
    ]
  }
]
`
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatalf("write cucumber file: %v", err)
	}

	result, err := IngestCucumber(tempDir, filePath, "bottleneck/assurance/results.json", false, true)
	if err != nil {
		t.Fatalf("dry-run ingest cucumber: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tempDir, "bottleneck/assurance/results.json")); err == nil {
		t.Fatal("expected output file to not be written in dry-run mode")
	}
	artifact, ok := result.Artifact.(AssuranceArtifact)
	if !ok {
		t.Fatalf("expected AssuranceArtifact, got %T", result.Artifact)
	}
	if artifact.ScenariosTotal != 1 {
		t.Fatalf("expected 1 scenario, got %d", artifact.ScenariosTotal)
	}
}

func TestIngestCucumberMergeAppendsExistingEvidence(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "cucumber.json")
	content := `[
  {
    "name": "Checkout feature",
    "elements": [
      {
        "name": "Valid checkout",
        "type": "scenario",
        "tags": [{"name": "@BEHAVIOR-001"}],
        "steps": [
          {"result": {"status": "passed"}}
        ]
      }
    ]
  }
]
`
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatalf("write cucumber file: %v", err)
	}

	existing := `{
  "scenarios_total": 1,
  "scenarios_passed": 1,
  "scenarios_failed": 0,
  "failures": [],
  "evidence": [
    {"id": "ASSURANCE-000", "type": "manual", "source": "manual", "status": "pass"}
  ]
}`
	existingPath := filepath.Join(tempDir, "bottleneck/assurance/results.json")
	if err := os.MkdirAll(filepath.Dir(existingPath), 0o755); err != nil {
		t.Fatalf("mkdir existing path: %v", err)
	}
	if err := os.WriteFile(existingPath, []byte(existing), 0o644); err != nil {
		t.Fatalf("write existing artifacts: %v", err)
	}

	result, err := IngestCucumber(tempDir, filePath, "bottleneck/assurance/results.json", true, false)
	if err != nil {
		t.Fatalf("merge ingest cucumber: %v", err)
	}
	artifact, ok := result.Artifact.(AssuranceArtifact)
	if !ok {
		t.Fatalf("expected AssuranceArtifact, got %T", result.Artifact)
	}
	if len(artifact.Evidence) != 2 {
		t.Fatalf("expected merged evidence count 2, got %d", len(artifact.Evidence))
	}
}

func TestIngestCodeQLInvalidInputReturnsError(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "codeql.sarif")
	if err := os.WriteFile(filePath, []byte("not valid json"), 0o644); err != nil {
		t.Fatalf("write invalid codeql file: %v", err)
	}

	_, err := IngestCodeQL(tempDir, filePath, "", false, false)
	if err == nil {
		t.Fatal("expected error for invalid SARIF input")
	}
}

func fixturePath(name string) string {
	return filepath.Join("testdata", name)
}

func sampleReportPath(name string) string {
	return filepath.Join("..", "..", "examples", "saas", "reports", name)
}

func assertWrittenArtifactContains(t *testing.T, rootPath string, relativePath string, expected ...string) {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(rootPath, relativePath))
	if err != nil {
		t.Fatalf("read written artifact %s: %v", relativePath, err)
	}
	text := string(content)
	for _, substring := range expected {
		if !strings.Contains(text, substring) {
			t.Fatalf("expected written artifact %s to contain %q\n%s", relativePath, substring, text)
		}
	}
}

func writeBehaviorSpec(t *testing.T, rootPath string, behaviorID string) {
	t.Helper()
	path := filepath.Join(rootPath, "bottleneck", "behavior", "behavior-spec.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create behavior dir: %v", err)
	}
	content := "# Behavior Specification\n\n## Expected Behavior\n\n### " + behaviorID + ": Test behavior\nCritical: true\n\nThe system does the expected thing.\n\n## Unacceptable Behavior\n\nThe system must not do the wrong thing.\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write behavior spec: %v", err)
	}
}

func containsString(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

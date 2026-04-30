package ingest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeCucumberCreatesAssuranceArtifact(t *testing.T) {
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
      },
      {
        "name": "Invalid payment",
        "type": "scenario",
        "tags": [{"name": "@BEHAVIOR-002"}],
        "steps": [
          {"result": {"status": "failed"}}
        ]
      }
    ]
  }
]
`
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatalf("write cucumber file: %v", err)
	}

	result, err := IngestCucumber(tempDir, filePath, "", false, false)
	if err != nil {
		t.Fatalf("ingest cucumber: %v", err)
	}
	artifact, ok := result.Artifact.(AssuranceArtifact)
	if !ok {
		t.Fatalf("expected AssuranceArtifact, got %T", result.Artifact)
	}
	if artifact.ScenariosTotal != 2 {
		t.Fatalf("expected 2 scenarios, got %d", artifact.ScenariosTotal)
	}
	if artifact.ScenariosPassed != 1 || artifact.ScenariosFailed != 1 {
		t.Fatalf("unexpected pass/fail counts: %d/%d", artifact.ScenariosPassed, artifact.ScenariosFailed)
	}
	if len(artifact.Evidence) != 2 {
		t.Fatalf("expected 2 evidence entries, got %d", len(artifact.Evidence))
	}
	if len(result.Warnings) != 0 {
		t.Fatalf("expected no warnings, got %v", result.Warnings)
	}
}

func TestNormalizeCodeQLCreatesSecurityArtifact(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "codeql.sarif")
	content := `{
  "runs": [
    {
      "results": [
        {
          "ruleId": "go/sql-injection",
          "level": "error",
          "message": {"text": "Database query built from user-controlled input"},
          "locations": [
            {"physicalLocation": {"artifactLocation": {"uri": "internal/db/query.go"}, "region": {"startLine": 42}}}
          ]
        }
      ]
    }
  ]
}`
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatalf("write codeql file: %v", err)
	}

	artifact, err := IngestCodeQL(tempDir, filePath, "", false, false)
	if err != nil {
		t.Fatalf("ingest codeql: %v", err)
	}
	security, ok := artifact.Artifact.(SecurityArtifact)
	if !ok {
		t.Fatalf("expected SecurityArtifact, got %T", artifact.Artifact)
	}
	if security.Violations != 1 {
		t.Fatalf("expected 1 violation, got %d", security.Violations)
	}
	if security.Findings["high"] != 1 {
		t.Fatalf("expected high finding count 1, got %d", security.Findings["high"])
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
	filePath := filepath.Join(tempDir, "telemetry.json")
	content := `{
  "adoption_rate": 0.72,
  "error_rate": 0.03,
  "evidence": [
    {"id": "EXECUTION-001", "type": "telemetry", "source": "snapshot", "status": "pass"}
  ]
}`
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatalf("write telemetry file: %v", err)
	}

	artifact, err := IngestTelemetry(tempDir, filePath, "", false, false)
	if err != nil {
		t.Fatalf("ingest telemetry: %v", err)
	}
	execution, ok := artifact.Artifact.(ExecutionArtifact)
	if !ok {
		t.Fatalf("expected ExecutionArtifact, got %T", artifact.Artifact)
	}
	if execution.AdoptionRate != 0.72 || execution.ErrorRate != 0.03 {
		t.Fatalf("unexpected telemetry values: %.2f/%.2f", execution.AdoptionRate, execution.ErrorRate)
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

package cmd

import (
	"strings"
	"testing"

	"bottleneck/internal/models"
)

func TestDiagnoseExitCodeMatchesSystemStatus(t *testing.T) {
	if code := diagnoseExitCode(models.EngineResult{SystemStatus: models.StatusFail}); code != 1 {
		t.Fatalf("expected FAIL diagnosis exit code 1, got %d", code)
	}
	if code := diagnoseExitCode(models.EngineResult{SystemStatus: models.StatusWarning}); code != 0 {
		t.Fatalf("expected WARNING diagnosis exit code 0, got %d", code)
	}
	if code := diagnoseExitCode(models.EngineResult{SystemStatus: models.StatusPass}); code != 0 {
		t.Fatalf("expected PASS diagnosis exit code 0, got %d", code)
	}
}

func TestDiagnoseCommandIsRegistered(t *testing.T) {
	if diagnoseCmd.Use != "diagnose" {
		t.Fatalf("expected diagnose command, got %q", diagnoseCmd.Use)
	}
	if diagnoseCmd.Flags().Lookup("gate") == nil {
		t.Fatal("expected diagnose --gate flag")
	}
	if diagnoseCmd.Flags().Lookup("format") == nil {
		t.Fatal("expected diagnose --format flag")
	}
}

func TestRenderDiagnoseGatePassesWhenReleaseGatePasses(t *testing.T) {
	output, exitCode, err := renderDiagnoseGate(models.EngineResult{
		Environment:       "default",
		SystemStatus:      models.StatusPass,
		PrimaryBottleneck: "None",
		Results: []models.ValidationResult{
			{Capability: "Behavior", Status: models.StatusPass, Details: []string{"behavior evidence"}},
			{Capability: "Intent", Status: models.StatusPass, Details: []string{"intent evidence"}},
			{Capability: "Design", Status: models.StatusPass, Details: []string{"design evidence"}},
			{Capability: "Assurance", Status: models.StatusPass, Details: []string{"accuracy: 1.00"}},
			{Capability: "Security", Status: models.StatusPass, Details: []string{"violations: 0"}},
			{Capability: "Execution", Status: models.StatusPass, Details: []string{"error_rate: 0.01"}},
			{Capability: "Traceability", Status: models.StatusPass, Details: []string{"traceability_nodes: 6"}},
		},
	}, "release", "text", "default")
	if err != nil {
		t.Fatalf("renderDiagnoseGate returned error: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("expected passing release gate exit 0, got %d\n%s", exitCode, output)
	}
	if !strings.Contains(output, "Release Gate: PASS") {
		t.Fatalf("expected release gate output, got:\n%s", output)
	}
}

func TestRenderDiagnoseGateFailsWhenReleaseGateFails(t *testing.T) {
	output, exitCode, err := renderDiagnoseGate(models.EngineResult{
		Environment:       "default",
		SystemStatus:      models.StatusFail,
		PrimaryBottleneck: "Assurance",
		Results: []models.ValidationResult{
			{Capability: "Behavior", Status: models.StatusPass, Details: []string{"behavior evidence"}},
			{Capability: "Intent", Status: models.StatusPass, Details: []string{"intent evidence"}},
			{Capability: "Design", Status: models.StatusPass, Details: []string{"design evidence"}},
			{Capability: "Assurance", Status: models.StatusFail, Message: "accuracy below threshold", Details: []string{"accuracy: 0.50 (threshold: 0.90)"}},
			{Capability: "Security", Status: models.StatusPass, Details: []string{"violations: 0"}},
			{Capability: "Execution", Status: models.StatusPass, Details: []string{"error_rate: 0.01"}},
			{Capability: "Traceability", Status: models.StatusPass, Details: []string{"traceability_nodes: 6"}},
		},
	}, "release", "markdown", "default")
	if err != nil {
		t.Fatalf("renderDiagnoseGate returned error: %v", err)
	}
	if exitCode != 1 {
		t.Fatalf("expected failing release gate exit 1, got %d\n%s", exitCode, output)
	}
	if !strings.Contains(output, "## Bottleneck Release Gate") ||
		!strings.Contains(output, "| Result | FAIL |") ||
		!strings.Contains(output, "Primary bottleneck score is below release threshold.") {
		t.Fatalf("expected failing release gate markdown, got:\n%s", output)
	}
}

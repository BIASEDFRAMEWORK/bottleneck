package traceability

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidEvidenceIDsParseIntoGraph(t *testing.T) {
	rootPath := writeTraceProject(t, nil)

	graph, err := Build(rootPath, Options{Environment: "default"})
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	expectedIDs := []string{"INTENT-001", "BEHAVIOR-001", "ASSURANCE-001", "SECURITY-001", "EXECUTION-001"}
	for _, id := range expectedIDs {
		if _, ok := graph.Nodes[id]; !ok {
			t.Fatalf("expected graph node %s in %#v", id, graph.Nodes)
		}
	}
	if len(graph.ValidateFindings()) != 0 {
		t.Fatalf("expected no graph findings, got %#v", graph.ValidateFindings())
	}
}

func TestDuplicateIDsFailValidation(t *testing.T) {
	rootPath := writeTraceProject(t, map[string]string{
		"design/architecture.md": "# Architecture\n\n### INTENT-001: Duplicate intent ID\n\nArchitecture notes.\n",
	})

	graph, err := Build(rootPath, Options{Environment: "default"})
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	assertFinding(t, graph.ValidateFindings(), SeverityFail, "duplicates evidence ID INTENT-001")
}

func TestInvalidIDSyntaxFailsValidation(t *testing.T) {
	rootPath := writeTraceProject(t, map[string]string{
		"intent/intent.md": "# Intent\n\n## Outcomes\n\n### INTENT-01: Invalid ID\nRefs:\n- BEHAVIOR-001\n\nRelease decisions have evidence.\n",
	})

	graph, err := Build(rootPath, Options{Environment: "default"})
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	assertFinding(t, graph.ValidateFindings(), SeverityFail, "invalid evidence ID INTENT-01")
}

func TestMissingReferencesFailValidationWithSourceAndReference(t *testing.T) {
	rootPath := writeTraceProject(t, map[string]string{
		"behavior/behavior-spec.md": "# Behavior Specification\n\n## Expected Behavior\n\n### BEHAVIOR-001: Block release\nCritical: true\nRefs:\n- INTENT-001\n- ASSURANCE-009\n\nThe system blocks unsafe releases.\n\n## Unacceptable Behavior\n\nThe system must not ignore failed assurance.\n",
	})

	graph, err := Build(rootPath, Options{Environment: "default"})
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	assertFinding(t, graph.ValidateFindings(), SeverityFail, "bottleneck/behavior/behavior-spec.md BEHAVIOR-001 references missing ASSURANCE-009")
}

func TestBehaviorWithoutIntentWarnsByDefaultAndFailsWhenStrictOrProduction(t *testing.T) {
	files := map[string]string{
		"intent/intent.md":          "# Intent\n\n## Outcomes\n\nA release decision has explicit evidence.\n",
		"behavior/behavior-spec.md": "# Behavior Specification\n\n## Expected Behavior\n\n### BEHAVIOR-001: Block release\nRefs:\n- ASSURANCE-001\n\nThe system blocks unsafe releases.\n\n## Unacceptable Behavior\n\nThe system must not ignore failed assurance.\n",
	}

	defaultGraph, err := Build(writeTraceProject(t, files), Options{Environment: "default"})
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	assertFinding(t, defaultGraph.ValidateFindings(), SeverityWarning, "BEHAVIOR-001 is not linked to intent evidence")

	strictGraph, err := Build(writeTraceProject(t, files), Options{Environment: "default", Strict: true})
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	assertFinding(t, strictGraph.ValidateFindings(), SeverityFail, "BEHAVIOR-001 is not linked to intent evidence")

	productionGraph, err := Build(writeTraceProject(t, files), Options{Environment: "production"})
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	assertFinding(t, productionGraph.ValidateFindings(), SeverityFail, "BEHAVIOR-001 is not linked to intent evidence")
}

func TestCriticalBehaviorWithoutAssuranceWarnsByDefaultAndFailsWhenStrictOrProduction(t *testing.T) {
	files := map[string]string{
		"behavior/behavior-spec.md": "# Behavior Specification\n\n## Expected Behavior\n\n### BEHAVIOR-001: Block release\nCritical: true\nRefs:\n- INTENT-001\n\nThe system blocks unsafe releases.\n\n## Unacceptable Behavior\n\nThe system must not ignore failed assurance.\n",
		"assurance/results.json":    "{\n  \"scenarios_total\": 1,\n  \"scenarios_passed\": 1,\n  \"scenarios_failed\": 0,\n  \"failures\": []\n}\n",
	}

	defaultGraph, err := Build(writeTraceProject(t, files), Options{Environment: "default"})
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	assertFinding(t, defaultGraph.ValidateFindings(), SeverityWarning, "BEHAVIOR-001 is critical but is not linked to assurance evidence")

	strictGraph, err := Build(writeTraceProject(t, files), Options{Environment: "default", Strict: true})
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	assertFinding(t, strictGraph.ValidateFindings(), SeverityFail, "BEHAVIOR-001 is critical but is not linked to assurance evidence")

	productionGraph, err := Build(writeTraceProject(t, files), Options{Environment: "production"})
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	assertFinding(t, productionGraph.ValidateFindings(), SeverityFail, "BEHAVIOR-001 is critical but is not linked to assurance evidence")
}

func TestBehaviorWithoutAssuranceReportsTraceabilityGap(t *testing.T) {
	files := map[string]string{
		"intent/intent.md":          "# Intent\n\n## Outcomes\n\n### INTENT-001: Evidence-backed release\nRefs:\n- BEHAVIOR-001\n\nRelease decisions have explicit evidence.\n",
		"behavior/behavior-spec.md": "# Behavior Specification\n\n## Expected Behavior\n\n### BEHAVIOR-001: Block release\nRefs:\n- INTENT-001\n\nThe system blocks unsafe releases.\n\n## Unacceptable Behavior\n\nThe system must not ignore failed assurance.\n",
		"assurance/results.json":    "{\n  \"scenarios_total\": 1,\n  \"scenarios_passed\": 1,\n  \"scenarios_failed\": 0,\n  \"failures\": []\n}\n",
		"security/guardrails.json":  "{\n  \"violations\": 0\n}\n",
		"execution/telemetry.json":  "{\n  \"adoption_rate\": 0.9,\n  \"error_rate\": 0.01\n}\n",
	}

	graph, err := Build(writeTraceProject(t, files), Options{Environment: "default"})
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	assertFinding(t, graph.ValidateFindings(), SeverityWarning, "Traceability Gap: BEHAVIOR-001 exists, but no assurance result references it")
}

func TestBehaviorWithOnlyOutboundAssuranceRefStillRequiresScenarioEvidence(t *testing.T) {
	files := map[string]string{
		"behavior/behavior-spec.md": "# Behavior Specification\n\n## Expected Behavior\n\n### BEHAVIOR-001: Block release\nRefs:\n- INTENT-001\n- ASSURANCE-001\n\nThe system blocks unsafe releases.\n\n## Unacceptable Behavior\n\nThe system must not ignore failed assurance.\n",
		"assurance/results.json":    "{\n  \"scenarios_total\": 1,\n  \"scenarios_passed\": 1,\n  \"scenarios_failed\": 0,\n  \"failures\": [],\n  \"evidence\": [{\"id\":\"ASSURANCE-001\",\"refs\":[],\"source\":\"cucumber\",\"status\":\"pass\"}]\n}\n",
	}

	graph, err := Build(writeTraceProject(t, files), Options{Environment: "default"})
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	assertFinding(t, graph.ValidateFindings(), SeverityWarning, "Traceability Gap: BEHAVIOR-001 exists, but no assurance result references it")
}

func TestOrphanedEvidenceIsReported(t *testing.T) {
	rootPath := writeTraceProject(t, map[string]string{
		"intent/intent.md":          "# Intent\n\n## Outcomes\n\n### INTENT-002: Orphan intent\n\nRelease decisions have evidence.\n",
		"behavior/behavior-spec.md": "# Behavior Specification\n\n## Expected Behavior\n\nUseful expected behavior content.\n\n## Unacceptable Behavior\n\nUseful unacceptable behavior content.\n",
		"design/architecture.md":    "# Architecture\n\nUseful design content.\n",
		"assurance/results.json":    "{\n  \"scenarios_total\": 1,\n  \"scenarios_passed\": 1,\n  \"scenarios_failed\": 0,\n  \"failures\": [],\n  \"evidence\": [{\"id\":\"ASSURANCE-003\",\"refs\":[],\"source\":\"cucumber\",\"status\":\"pass\"}]\n}\n",
		"security/guardrails.json":  "{\n  \"violations\": 0,\n  \"evidence\": [{\"id\":\"SECURITY-001\",\"refs\":[],\"source\":\"scanner\",\"status\":\"pass\"}]\n}\n",
		"execution/telemetry.json":  "{\n  \"adoption_rate\": 0.9,\n  \"error_rate\": 0.01,\n  \"evidence\": [{\"id\":\"EXECUTION-001\",\"refs\":[],\"source\":\"telemetry\",\"status\":\"pass\"}]\n}\n",
	})

	graph, err := Build(rootPath, Options{Environment: "default"})
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	findings := graph.ValidateFindings()
	for _, finding := range findings {
		if finding.Severity == SeverityFail {
			t.Fatalf("expected only orphan warnings, got failure %#v in %#v", finding, findings)
		}
	}
	assertFinding(t, findings, SeverityWarning, "INTENT-002 is not linked to any behavior")
	assertFinding(t, findings, SeverityWarning, "ASSURANCE-003 is not linked to any behavior")
	assertFinding(t, findings, SeverityWarning, "SECURITY-001 is not mapped to behavior or intent evidence")
	assertFinding(t, findings, SeverityWarning, "EXECUTION-001 is not tied to behavior or assurance release readiness evidence")
}

func TestTraceTextOutputShowsRefsInboundAndEvidenceChain(t *testing.T) {
	graph, err := Build(writeTraceProject(t, nil), Options{Environment: "production"})
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	trace, err := graph.Trace("BEHAVIOR-001")
	if err != nil {
		t.Fatalf("Trace returned error: %v", err)
	}

	output := RenderText(trace)
	expected := []string{
		"Trace: BEHAVIOR-001",
		"Behavior:",
		"- Found in bottleneck/behavior/behavior-spec.md",
		"Related intent:",
		"- INTENT-001 found in bottleneck/intent/intent.md",
		"Design evidence:",
		"- DESIGN-001 found in bottleneck/design/architecture.md",
		"Assurance evidence:",
		"- ASSURANCE-001 found in bottleneck/assurance/results.json",
		"Security evidence:",
		"- SECURITY-001 found in bottleneck/security/guardrails.json",
		"Execution evidence:",
		"- EXECUTION-001 found in bottleneck/execution/telemetry.json",
		"INTENT-001 -> BEHAVIOR-001 -> ASSURANCE-001 -> SECURITY-001 -> EXECUTION-001",
		"Missing links:",
		"- None.",
		"Recommendation:",
	}
	for _, substring := range expected {
		if !strings.Contains(output, substring) {
			t.Fatalf("expected %q in output:\n%s", substring, output)
		}
	}
}

func TestTraceJSONOutputIsStable(t *testing.T) {
	graph, err := Build(writeTraceProject(t, nil), Options{Environment: "production"})
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	trace, err := graph.Trace("ASSURANCE-001")
	if err != nil {
		t.Fatalf("Trace returned error: %v", err)
	}

	output, err := RenderJSON(trace)
	if err != nil {
		t.Fatalf("RenderJSON returned error: %v", err)
	}

	var decoded TraceResult
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("expected stable JSON, got error: %v\n%s", err, output)
	}
	if decoded.SchemaVersion != SchemaVersion {
		t.Fatalf("expected schema version %q, got %q", SchemaVersion, decoded.SchemaVersion)
	}
	if decoded.Query != "ASSURANCE-001" || decoded.ID != "ASSURANCE-001" || decoded.Kind != TypeAssurance || !decoded.Found {
		t.Fatalf("expected query ASSURANCE-001, got %q", decoded.Query)
	}
	if decoded.Node.ArtifactPath != "bottleneck/assurance/results.json" {
		t.Fatalf("expected artifact path, got %q", decoded.Node.ArtifactPath)
	}
}

func TestTraceIntentShowsRelatedBehaviorAndDownstreamEvidence(t *testing.T) {
	graph, err := Build(writeTraceProject(t, map[string]string{
		"behavior/behavior-spec.md": "# Behavior Specification\n\n## Expected Behavior\n\n### BEHAVIOR-001: Block release\nRefs:\n- INTENT-001\n- ASSURANCE-001\n\nThe system blocks unsafe releases.\n\n### BEHAVIOR-002: Summarize risk\nRefs:\n- INTENT-001\n\nThe system summarizes risk caveats.\n\n## Unacceptable Behavior\n\nThe system must not ignore failed assurance.\n",
		"assurance/results.json":    "{\n  \"scenarios_total\": 1,\n  \"scenarios_passed\": 1,\n  \"scenarios_failed\": 0,\n  \"failures\": [],\n  \"evidence\": [{\"id\":\"ASSURANCE-001\",\"refs\":[\"BEHAVIOR-001\"],\"source\":\"cucumber\",\"status\":\"pass\"}]\n}\n",
	}), Options{Environment: "default"})
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	trace, err := graph.Trace("INTENT-001")
	if err != nil {
		t.Fatalf("Trace returned error: %v", err)
	}

	output := RenderText(trace)
	expected := []string{
		"Trace: INTENT-001",
		"Related behavior:",
		"- BEHAVIOR-001 found in bottleneck/behavior/behavior-spec.md",
		"- BEHAVIOR-002 found in bottleneck/behavior/behavior-spec.md",
		"Assurance evidence:",
		"- ASSURANCE-001 found in bottleneck/assurance/results.json",
		"Missing links:",
		"BEHAVIOR-002 has no assurance result",
	}
	for _, substring := range expected {
		if !strings.Contains(output, substring) {
			t.Fatalf("expected %q in output:\n%s", substring, output)
		}
	}
}

func TestTraceBehaviorHighlightsMissingLinks(t *testing.T) {
	graph, err := Build(writeTraceProject(t, map[string]string{
		"design/architecture.md":   "# Architecture\n\n### DESIGN-002: Other design\nRefs:\n- INTENT-001\n\nDesign for a different path.\n",
		"assurance/results.json":   "{\n  \"scenarios_total\": 1,\n  \"scenarios_passed\": 1,\n  \"scenarios_failed\": 0,\n  \"failures\": [],\n  \"evidence\": [{\"id\":\"ASSURANCE-002\",\"refs\":[],\"source\":\"cucumber\",\"status\":\"pass\"}]\n}\n",
		"security/guardrails.json": "{\n  \"violations\": 0\n}\n",
		"execution/telemetry.json": "{\n  \"adoption_rate\": 0.9,\n  \"error_rate\": 0.01\n}\n",
	}), Options{Environment: "default"})
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	trace, err := graph.Trace("BEHAVIOR-001")
	if err != nil {
		t.Fatalf("Trace returned error: %v", err)
	}

	output := RenderText(trace)
	expected := []string{
		"Design evidence:",
		"- Missing: No design evidence references BEHAVIOR-001.",
		"Assurance evidence:",
		"- Missing: No assurance result references BEHAVIOR-001.",
		"Security evidence:",
		"- Missing: No security evidence references BEHAVIOR-001.",
		"Execution evidence:",
		"- Missing: No telemetry or execution signal references BEHAVIOR-001.",
		"BEHAVIOR-001 has no design reference",
		"BEHAVIOR-001 has no assurance result",
		"Recommendation:",
		"Add design, assurance, security, and execution evidence that references BEHAVIOR-001.",
	}
	for _, substring := range expected {
		if !strings.Contains(output, substring) {
			t.Fatalf("expected %q in output:\n%s", substring, output)
		}
	}
}

func TestTraceFixtureProjectsCoverCompleteAndMissingChains(t *testing.T) {
	complete, err := Build(filepath.Join("testdata", "complete-chain", "bottleneck"), Options{Environment: "default"})
	if err != nil {
		t.Fatalf("Build complete fixture returned error: %v", err)
	}
	completeTrace, err := complete.Trace("BEHAVIOR-001")
	if err != nil {
		t.Fatalf("Trace complete fixture returned error: %v", err)
	}
	if len(completeTrace.RelatedDesign) == 0 || len(completeTrace.RelatedAssurance) == 0 || len(completeTrace.RelatedSecurity) == 0 || len(completeTrace.RelatedExecution) == 0 {
		t.Fatalf("expected complete fixture to include downstream evidence, got %#v", completeTrace)
	}

	missing, err := Build(filepath.Join("testdata", "missing-links", "bottleneck"), Options{Environment: "default"})
	if err != nil {
		t.Fatalf("Build missing fixture returned error: %v", err)
	}
	missingTrace, err := missing.Trace("BEHAVIOR-001")
	if err != nil {
		t.Fatalf("Trace missing fixture returned error: %v", err)
	}
	output := RenderText(missingTrace)
	for _, substring := range []string{"BEHAVIOR-001 has no design reference", "BEHAVIOR-001 has no assurance result"} {
		if !strings.Contains(output, substring) {
			t.Fatalf("expected %q in missing fixture trace:\n%s", substring, output)
		}
	}
}

func TestUnknownTraceIDReturnsUsefulError(t *testing.T) {
	graph, err := Build(writeTraceProject(t, nil), Options{Environment: "default"})
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	_, err = graph.Trace("BEHAVIOR-999")
	if err == nil {
		t.Fatal("expected unknown trace ID error")
	}
	if !strings.Contains(err.Error(), `unknown trace id "BEHAVIOR-999"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func writeTraceProject(t *testing.T, overrides map[string]string) string {
	t.Helper()

	rootPath := filepath.Join(t.TempDir(), "bottleneck")
	files := map[string]string{
		"intent/intent.md":          "# Intent\n\n## Outcomes\n\n### INTENT-001: Reduce release risk\nRefs:\n- BEHAVIOR-001\n\nRelease decisions must be backed by evidence.\n",
		"behavior/behavior-spec.md": "# Behavior Specification\n\n## Expected Behavior\n\n### BEHAVIOR-001: Block unsafe release\nCritical: true\nRefs:\n- INTENT-001\n- ASSURANCE-001\n\nWhen assurance fails, release is blocked.\n\n## Unacceptable Behavior\n\nThe system must not ignore failed assurance.\n",
		"design/architecture.md":    "# Architecture\n\n### DESIGN-001: Scorecard flow\nRefs:\n- INTENT-001\n- BEHAVIOR-001\n\nThe CLI links evidence before rendering release posture.\n",
		"assurance/results.json":    "{\n  \"scenarios_total\": 1,\n  \"scenarios_passed\": 1,\n  \"scenarios_failed\": 0,\n  \"failures\": [],\n  \"evidence\": [{\"id\":\"ASSURANCE-001\",\"refs\":[\"BEHAVIOR-001\"],\"source\":\"cucumber\",\"status\":\"pass\"}]\n}\n",
		"security/guardrails.json":  "{\n  \"violations\": 0,\n  \"evidence\": [{\"id\":\"SECURITY-001\",\"refs\":[\"BEHAVIOR-001\",\"ASSURANCE-001\"],\"source\":\"scanner\",\"status\":\"pass\"}]\n}\n",
		"execution/telemetry.json":  "{\n  \"adoption_rate\": 0.9,\n  \"error_rate\": 0.01,\n  \"evidence\": [{\"id\":\"EXECUTION-001\",\"refs\":[\"BEHAVIOR-001\",\"ASSURANCE-001\"],\"source\":\"telemetry\",\"status\":\"pass\"}]\n}\n",
	}
	for path, content := range overrides {
		files[path] = content
	}
	for relativePath, content := range files {
		fullPath := filepath.Join(rootPath, relativePath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatalf("failed to create directory for %s: %v", relativePath, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatalf("failed to write %s: %v", relativePath, err)
		}
	}
	return rootPath
}

func assertFinding(t *testing.T, findings []Finding, severity string, substring string) {
	t.Helper()

	for _, finding := range findings {
		if finding.Severity == severity && strings.Contains(finding.Message, substring) {
			return
		}
	}
	t.Fatalf("expected %s finding containing %q, got %#v", severity, substring, findings)
}

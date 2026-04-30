package ingest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var currentTime = func() time.Time {
	return time.Now().UTC()
}

const (
	DefaultAssuranceOutput = "bottleneck/assurance/results.json"
	DefaultSecurityOutput  = "bottleneck/security/guardrails.json"
	DefaultExecutionOutput = "bottleneck/execution/telemetry.json"
)

var behaviorIDPattern = regexp.MustCompile(`@?(BEHAVIOR-[0-9]{3,})`)

type EvidenceItem struct {
	ID         string   `json:"id,omitempty"`
	Type       string   `json:"type"`
	Source     string   `json:"source"`
	Refs       []string `json:"refs,omitempty"`
	Severity   string   `json:"severity,omitempty"`
	RuleID     string   `json:"rule_id,omitempty"`
	Path       string   `json:"path,omitempty"`
	Line       *int     `json:"line,omitempty"`
	Status     string   `json:"status"`
	Summary    string   `json:"summary,omitempty"`
	IngestedAt string   `json:"ingested_at,omitempty"`
}

type AssuranceArtifact struct {
	ScenariosTotal  int            `json:"scenarios_total"`
	ScenariosPassed int            `json:"scenarios_passed"`
	ScenariosFailed int            `json:"scenarios_failed"`
	Failures        []string       `json:"failures"`
	Coverage        *float64       `json:"coverage,omitempty"`
	TestsSkipped    *int           `json:"tests_skipped,omitempty"`
	Evidence        []EvidenceItem `json:"evidence,omitempty"`
}

type SecurityArtifact struct {
	Violations int            `json:"violations"`
	Findings   map[string]int `json:"findings"`
	Evidence   []EvidenceItem `json:"evidence,omitempty"`
}

type ExecutionArtifact struct {
	SourceEnvironment string         `json:"source_environment,omitempty"`
	WindowStart       string         `json:"window_start,omitempty"`
	WindowEnd         string         `json:"window_end,omitempty"`
	AdoptionRate      float64        `json:"adoption_rate"`
	ErrorRate         float64        `json:"error_rate"`
	LatencyP95Ms      *float64       `json:"latency_p95_ms,omitempty"`
	RollbackRate      *float64       `json:"rollback_rate,omitempty"`
	UserOverrideRate  *float64       `json:"user_override_rate,omitempty"`
	Cost              *ExecutionCost `json:"cost,omitempty"`
	Evidence          []EvidenceItem `json:"evidence,omitempty"`
}

type ExecutionCost struct {
	TotalUSD    float64 `json:"total_usd,omitempty"`
	UnitCostUSD float64 `json:"unit_cost_usd,omitempty"`
}

type TestSummaryInput struct {
	TestsTotal   *int           `json:"tests_total"`
	TestsPassed  *int           `json:"tests_passed"`
	TestsFailed  *int           `json:"tests_failed"`
	TestsSkipped *int           `json:"tests_skipped"`
	Coverage     *float64       `json:"coverage"`
	Source       string         `json:"source"`
	Evidence     []EvidenceItem `json:"evidence,omitempty"`
}

// IngestSummary captures result metadata after an ingest operation.
type IngestSummary struct {
	Artifact interface{}
	Warnings []string
}

// DryRunPayload contains the artifact and warnings for dry-run JSON output.
type DryRunPayload struct {
	Artifact interface{} `json:"artifact"`
	Warnings []string    `json:"warnings,omitempty"`
}

// OutPayload contains the written path and warnings for JSON summaries.
type OutPayload struct {
	WrittenPath string   `json:"written_path"`
	Warnings    []string `json:"warnings,omitempty"`
}

func MarshalJSON(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

func (s IngestSummary) Text() string {
	lines := []string{}
	switch artifact := s.Artifact.(type) {
	case AssuranceArtifact:
		lines = append(lines, fmt.Sprintf("assurance artifact: %d scenarios (%d passed, %d failed)", artifact.ScenariosTotal, artifact.ScenariosPassed, artifact.ScenariosFailed))
	case SecurityArtifact:
		lines = append(lines, fmt.Sprintf("security artifact: %d violations", artifact.Violations))
	case ExecutionArtifact:
		lines = append(lines, fmt.Sprintf("execution artifact: adoption_rate=%.2f error_rate=%.2f", artifact.AdoptionRate, artifact.ErrorRate))
	default:
		lines = append(lines, "normalized artifact ready")
	}
	if len(s.Warnings) > 0 {
		lines = append(lines, "warnings:")
		for _, warning := range s.Warnings {
			lines = append(lines, fmt.Sprintf("  - %s", warning))
		}
	}
	return strings.Join(lines, "\n")
}

func IngestCucumber(rootPath, filePath, outPath string, merge, dryRun bool) (IngestSummary, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return IngestSummary{}, fmt.Errorf("read cucumber file: %w", err)
	}

	var report cucumberReport
	if err := json.Unmarshal(content, &report); err != nil {
		return IngestSummary{}, fmt.Errorf("parse cucumber json: %w", err)
	}

	artifact, warnings := normalizeCucumber(report, filePath)
	if outPath == "" {
		outPath = DefaultAssuranceOutput
	}

	if merge {
		artifact, err = mergeAssuranceArtifact(rootPath, outPath, artifact)
		if err != nil {
			return IngestSummary{}, err
		}
	}

	if !dryRun {
		if err := writeJSONArtifact(rootPath, outPath, artifact); err != nil {
			return IngestSummary{}, err
		}
	}

	return IngestSummary{Artifact: artifact, Warnings: warnings}, nil
}

func IngestCodeQL(rootPath, filePath, outPath string, merge, dryRun bool) (IngestSummary, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return IngestSummary{}, fmt.Errorf("read codeql file: %w", err)
	}

	var sarif sarifReport
	if err := json.Unmarshal(content, &sarif); err != nil {
		return IngestSummary{}, fmt.Errorf("parse sarif json: %w", err)
	}

	artifact := normalizeCodeQL(sarif, filePath)
	if outPath == "" {
		outPath = DefaultSecurityOutput
	}

	if merge {
		artifact, err = mergeSecurityArtifact(rootPath, outPath, artifact)
		if err != nil {
			return IngestSummary{}, err
		}
	}

	if !dryRun {
		if err := writeJSONArtifact(rootPath, outPath, artifact); err != nil {
			return IngestSummary{}, err
		}
	}

	return IngestSummary{Artifact: artifact}, nil
}

func IngestTestSummary(rootPath, filePath, outPath string, merge, dryRun bool) (IngestSummary, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return IngestSummary{}, fmt.Errorf("read test summary file: %w", err)
	}

	var input TestSummaryInput
	if err := json.Unmarshal(content, &input); err != nil {
		return IngestSummary{}, fmt.Errorf("parse test summary json: %w", err)
	}

	artifact, err := normalizeTestSummary(input, filePath)
	if err != nil {
		return IngestSummary{}, err
	}
	if outPath == "" {
		outPath = DefaultAssuranceOutput
	}

	if merge {
		artifact, err = mergeAssuranceArtifact(rootPath, outPath, artifact)
		if err != nil {
			return IngestSummary{}, err
		}
	}

	if !dryRun {
		if err := writeJSONArtifact(rootPath, outPath, artifact); err != nil {
			return IngestSummary{}, err
		}
	}

	return IngestSummary{Artifact: artifact}, nil
}

func IngestTelemetry(rootPath, filePath, outPath string, merge, dryRun bool) (IngestSummary, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return IngestSummary{}, fmt.Errorf("read telemetry file: %w", err)
	}

	var input ExecutionArtifact
	if err := json.Unmarshal(content, &input); err != nil {
		return IngestSummary{}, fmt.Errorf("parse telemetry json: %w", err)
	}

	for i := range input.Evidence {
		if input.Evidence[i].Type == "" {
			input.Evidence[i].Type = "telemetry"
		}
		if input.Evidence[i].Source == "" {
			input.Evidence[i].Source = filePath
		}
		input.Evidence[i].IngestedAt = currentTime().Format(time.RFC3339)
	}

	if outPath == "" {
		outPath = DefaultExecutionOutput
	}

	artifact := input
	if merge {
		artifact, err = mergeExecutionArtifact(rootPath, outPath, artifact)
		if err != nil {
			return IngestSummary{}, err
		}
	}

	if !dryRun {
		if err := writeJSONArtifact(rootPath, outPath, artifact); err != nil {
			return IngestSummary{}, err
		}
	}

	return IngestSummary{Artifact: artifact}, nil
}

func writeJSONArtifact(rootPath, outPath string, artifact interface{}) error {
	fullPath := filepath.Join(rootPath, outPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(fullPath, append(data, '\n'), 0o644)
}

func mergeAssuranceArtifact(rootPath, outPath string, artifact AssuranceArtifact) (AssuranceArtifact, error) {
	existing, err := readExistingAssuranceArtifact(rootPath, outPath)
	if err != nil {
		if os.IsNotExist(err) {
			return artifact, nil
		}
		return AssuranceArtifact{}, err
	}

	artifact.Evidence = mergeEvidence(existing.Evidence, artifact.Evidence)
	return artifact, nil
}

func mergeSecurityArtifact(rootPath, outPath string, artifact SecurityArtifact) (SecurityArtifact, error) {
	existing, err := readExistingSecurityArtifact(rootPath, outPath)
	if err != nil {
		if os.IsNotExist(err) {
			return artifact, nil
		}
		return SecurityArtifact{}, err
	}

	artifact.Evidence = mergeEvidence(existing.Evidence, artifact.Evidence)
	return artifact, nil
}

func mergeExecutionArtifact(rootPath, outPath string, artifact ExecutionArtifact) (ExecutionArtifact, error) {
	existing, err := readExistingExecutionArtifact(rootPath, outPath)
	if err != nil {
		if os.IsNotExist(err) {
			return artifact, nil
		}
		return ExecutionArtifact{}, err
	}

	artifact.Evidence = mergeEvidence(existing.Evidence, artifact.Evidence)
	return artifact, nil
}

func mergeEvidence(existing, incoming []EvidenceItem) []EvidenceItem {
	index := map[string]struct{}{}
	merged := make([]EvidenceItem, 0, len(existing)+len(incoming))
	for _, item := range existing {
		if item.ID != "" {
			index[item.ID] = struct{}{}
		}
		merged = append(merged, item)
	}
	for _, item := range incoming {
		if item.ID != "" {
			if _, ok := index[item.ID]; ok {
				continue
			}
			index[item.ID] = struct{}{}
		}
		merged = append(merged, item)
	}
	return merged
}

func readExistingAssuranceArtifact(rootPath, outPath string) (AssuranceArtifact, error) {
	fullPath := filepath.Join(rootPath, outPath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return AssuranceArtifact{}, err
	}
	var artifact AssuranceArtifact
	if err := json.Unmarshal(content, &artifact); err != nil {
		return AssuranceArtifact{}, err
	}
	return artifact, nil
}

func readExistingSecurityArtifact(rootPath, outPath string) (SecurityArtifact, error) {
	fullPath := filepath.Join(rootPath, outPath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return SecurityArtifact{}, err
	}
	var artifact SecurityArtifact
	if err := json.Unmarshal(content, &artifact); err != nil {
		return SecurityArtifact{}, err
	}
	return artifact, nil
}

func readExistingExecutionArtifact(rootPath, outPath string) (ExecutionArtifact, error) {
	fullPath := filepath.Join(rootPath, outPath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return ExecutionArtifact{}, err
	}
	var artifact ExecutionArtifact
	if err := json.Unmarshal(content, &artifact); err != nil {
		return ExecutionArtifact{}, err
	}
	return artifact, nil
}

func normalizeCucumber(report cucumberReport, source string) (AssuranceArtifact, []string) {
	total := 0
	passed := 0
	failed := 0
	failures := []string{}
	evidence := []EvidenceItem{}
	warnings := []string{}

	for _, feature := range report {
		for _, element := range feature.Elements {
			if element.Type != "scenario" && element.Type != "scenario_outline" && element.Type != "background" {
				continue
			}
			total++
			status := "pass"
			for _, step := range element.Steps {
				if strings.ToLower(step.Result.Status) != "passed" {
					status = "fail"
					break
				}
			}

			refs := extractBehaviorRefs(element.Tags)
			refs = append(refs, extractBehaviorRefs(feature.Tags)...)
			refs = uniqueStrings(refs)

			summary := element.Name
			if status == "fail" {
				failed++
				failures = append(failures, summary)
			} else {
				passed++
			}

			if len(refs) == 0 {
				warnings = append(warnings, fmt.Sprintf("unmapped scenario %q has no BEHAVIOR-* tags", element.Name))
			}

			evidence = append(evidence, EvidenceItem{
				ID:         fmt.Sprintf("ASSURANCE-%03d", len(evidence)+1),
				Type:       "cucumber",
				Source:     source,
				Refs:       refs,
				Status:     status,
				Summary:    summary,
				IngestedAt: currentTime().Format(time.RFC3339),
			})
		}
	}

	if total == 0 {
		warnings = append(warnings, "no Cucumber scenarios found")
	}

	return AssuranceArtifact{
		ScenariosTotal:  total,
		ScenariosPassed: passed,
		ScenariosFailed: failed,
		Failures:        failures,
		Evidence:        evidence,
	}, warnings
}

func normalizeCodeQL(report sarifReport, source string) SecurityArtifact {
	findings := map[string]int{
		"critical": 0,
		"high":     0,
		"medium":   0,
		"low":      0,
		"note":     0,
	}
	violations := 0
	evidence := []EvidenceItem{}

	for _, run := range report.Runs {
		for _, result := range run.Results {
			severity := determineSeverity(result)
			findings[severity]++
			violations++
			path, line := extractSarifLocation(result)
			message := result.Message.Text
			if message == "" {
				message = result.RuleID
			}

			evidence = append(evidence, EvidenceItem{
				ID:         fmt.Sprintf("SECURITY-%03d", len(evidence)+1),
				Type:       "codeql",
				Source:     source,
				Severity:   severity,
				RuleID:     result.RuleID,
				Path:       path,
				Line:       line,
				Status:     "fail",
				Summary:    message,
				IngestedAt: currentTime().Format(time.RFC3339),
			})
		}
	}

	return SecurityArtifact{
		Violations: violations,
		Findings:   findings,
		Evidence:   evidence,
	}
}

func normalizeTestSummary(input TestSummaryInput, source string) (AssuranceArtifact, error) {
	if input.TestsTotal == nil || input.TestsPassed == nil || input.TestsFailed == nil {
		return AssuranceArtifact{}, fmt.Errorf("required test summary fields missing")
	}

	failures := []string{}
	if *input.TestsFailed > 0 {
		failures = append(failures, fmt.Sprintf("%d tests failed", *input.TestsFailed))
	}

	artifact := AssuranceArtifact{
		ScenariosTotal:  *input.TestsTotal,
		ScenariosPassed: *input.TestsPassed,
		ScenariosFailed: *input.TestsFailed,
		Failures:        failures,
		Coverage:        input.Coverage,
		TestsSkipped:    input.TestsSkipped,
		Evidence:        []EvidenceItem{},
	}

	for _, item := range input.Evidence {
		item.Type = coalesce(item.Type, "test-summary")
		if item.Source == "" {
			item.Source = source
		}
		item.IngestedAt = currentTime().Format(time.RFC3339)
		artifact.Evidence = append(artifact.Evidence, item)
	}

	return artifact, nil
}

func determineSeverity(result sarifResult) string {
	if sev := extractStringProperty(result.Properties, "security-severity"); sev != "" {
		return normalizeSeverity(sev)
	}
	if sev := extractStringProperty(result.Properties, "severity"); sev != "" {
		return normalizeSeverity(sev)
	}
	if sev := extractStringProperty(result.Properties, "level"); sev != "" {
		return normalizeSeverity(sev)
	}
	if result.Level != "" {
		return normalizeSeverity(result.Level)
	}
	return "note"
}

func extractStringProperty(properties map[string]interface{}, key string) string {
	if properties == nil {
		return ""
	}
	value, ok := properties[key]
	if !ok {
		return ""
	}
	if s, ok := value.(string); ok {
		return s
	}
	return ""
}

func normalizeSeverity(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "critical", "high", "medium", "low", "note":
		return value
	case "error":
		return "high"
	case "warning":
		return "medium"
	case "none":
		return "note"
	}
	return "note"
}

func extractSarifLocation(result sarifResult) (string, *int) {
	for _, location := range result.Locations {
		if location.PhysicalLocation.ArtifactLocation.URI != "" {
			line := location.PhysicalLocation.Region.StartLine
			return location.PhysicalLocation.ArtifactLocation.URI, &line
		}
	}
	return "", nil
}

func coalesce(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func extractBehaviorRefs(tags []tag) []string {
	refs := []string{}
	for _, tag := range tags {
		matches := behaviorIDPattern.FindAllStringSubmatch(tag.Name, -1)
		for _, match := range matches {
			if len(match) > 1 {
				refs = append(refs, strings.TrimPrefix(match[1], "@"))
			}
		}
	}
	return refs
}

func uniqueStrings(items []string) []string {
	seen := map[string]struct{}{}
	result := []string{}
	for _, item := range items {
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}

type cucumberReport []cucumberFeature

type cucumberFeature struct {
	Name     string            `json:"name"`
	Elements []cucumberElement `json:"elements"`
	Tags     []tag             `json:"tags"`
}

type cucumberElement struct {
	Name  string         `json:"name"`
	Type  string         `json:"type"`
	Steps []cucumberStep `json:"steps"`
	Tags  []tag          `json:"tags"`
}

type cucumberStep struct {
	Result cucumberResult `json:"result"`
}

type cucumberResult struct {
	Status       string `json:"status"`
	ErrorMessage string `json:"error_message"`
}

type tag struct {
	Name string `json:"name"`
}

type sarifReport struct {
	Runs []sarifRun `json:"runs"`
}

type sarifRun struct {
	Results []sarifResult `json:"results"`
}

type sarifResult struct {
	RuleID     string                 `json:"ruleId"`
	Message    sarifMessage           `json:"message"`
	Level      string                 `json:"level"`
	Properties map[string]interface{} `json:"properties"`
	Locations  []sarifLocation        `json:"locations"`
}

type sarifMessage struct {
	Text string `json:"text"`
}

type sarifLocation struct {
	PhysicalLocation sarifPhysicalLocation `json:"physicalLocation"`
}

type sarifPhysicalLocation struct {
	ArtifactLocation sarifArtifactLocation `json:"artifactLocation"`
	Region           sarifRegion           `json:"region"`
}

type sarifArtifactLocation struct {
	URI string `json:"uri"`
}

type sarifRegion struct {
	StartLine int `json:"startLine"`
}

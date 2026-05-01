package ingest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
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

var (
	behaviorIDPattern = regexp.MustCompile(`@?(BEHAVIOR-[0-9]{3,})`)
	evidenceIDPattern = regexp.MustCompile(`\b(?:INTENT|BEHAVIOR|DESIGN|ASSURANCE|SECURITY|EXECUTION)-[0-9]{3,}\b`)
)

type EvidenceItem struct {
	ID         string   `json:"id,omitempty"`
	Type       string   `json:"type"`
	Source     string   `json:"source"`
	Refs       []string `json:"refs,omitempty"`
	Severity   string   `json:"severity,omitempty"`
	RuleID     string   `json:"rule_id,omitempty"`
	RuleName   string   `json:"rule_name,omitempty"`
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
	GeneratedAt         string               `json:"generated_at,omitempty"`
	SourceEnvironment   string               `json:"source_environment,omitempty"`
	Window              *TelemetryWindow     `json:"window,omitempty"`
	WindowStart         string               `json:"window_start,omitempty"`
	WindowEnd           string               `json:"window_end,omitempty"`
	DeploymentFrequency *DeploymentFrequency `json:"deployment_frequency,omitempty"`
	ChangeFailureRate   float64              `json:"change_failure_rate,omitempty"`
	AdoptionRate        float64              `json:"adoption_rate"`
	ErrorRate           float64              `json:"error_rate"`
	LatencyP95Ms        *float64             `json:"latency_p95_ms,omitempty"`
	RollbackRate        *float64             `json:"rollback_rate,omitempty"`
	UserOverrideRate    *float64             `json:"user_override_rate,omitempty"`
	Cost                *ExecutionCost       `json:"cost,omitempty"`
	Evidence            []EvidenceItem       `json:"evidence,omitempty"`
}

type TelemetryWindow struct {
	Start string `json:"start,omitempty"`
	End   string `json:"end,omitempty"`
}

type DeploymentFrequency struct {
	Deployments int     `json:"deployments"`
	PeriodDays  float64 `json:"period_days"`
}

type ExecutionCost struct {
	Total          float64 `json:"total,omitempty"`
	TotalUSD       float64 `json:"total_usd,omitempty"`
	Currency       string  `json:"currency,omitempty"`
	Budget         float64 `json:"budget,omitempty"`
	BudgetVariance float64 `json:"budget_variance,omitempty"`
	CostPerRequest float64 `json:"cost_per_request,omitempty"`
	UnitCostUSD    float64 `json:"unit_cost_usd,omitempty"`
	Trend          string  `json:"trend,omitempty"`
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
		return IngestSummary{}, ingestInputError("cucumber", filePath, "read cucumber file", err)
	}

	var report cucumberReport
	if err := json.Unmarshal(content, &report); err != nil {
		return IngestSummary{}, ingestInputError("cucumber", filePath, "parse cucumber json", err)
	}

	artifact, warnings := normalizeCucumber(report, filePath)
	warnings = append(warnings, unmatchedBehaviorWarnings(rootPath, artifact)...)
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
	return IngestSARIF(rootPath, filePath, outPath, merge, dryRun)
}

func ingestInputError(kind string, filePath string, action string, err error) error {
	return fmt.Errorf("%s %s: %w\nNext action: check the expected sample format in %s.", action, filePath, err, sampleFormatPath(kind))
}

func sampleFormatPath(kind string) string {
	switch kind {
	case "cucumber":
		return "examples/saas/reports/cucumber.json"
	case "sarif":
		return "examples/saas/reports/codeql.sarif"
	case "test-summary":
		return "examples/saas/reports/test-summary.json"
	case "telemetry":
		return "examples/saas/reports/telemetry.json"
	default:
		return "examples/saas/reports/"
	}
}

func IngestSARIF(rootPath, filePath, outPath string, merge, dryRun bool) (IngestSummary, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return IngestSummary{}, ingestInputError("sarif", filePath, "read sarif file", err)
	}

	var sarif sarifReport
	if err := json.Unmarshal(content, &sarif); err != nil {
		return IngestSummary{}, ingestInputError("sarif", filePath, "parse sarif json", err)
	}

	artifact := normalizeSARIF(sarif, filePath)
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
		return IngestSummary{}, ingestInputError("test-summary", filePath, "read test summary file", err)
	}

	var input TestSummaryInput
	if err := json.Unmarshal(content, &input); err != nil {
		return IngestSummary{}, ingestInputError("test-summary", filePath, "parse test summary json", err)
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
		return IngestSummary{}, ingestInputError("telemetry", filePath, "read telemetry file", err)
	}

	var input ExecutionArtifact
	if err := json.Unmarshal(content, &input); err != nil {
		return IngestSummary{}, ingestInputError("telemetry", filePath, "parse telemetry json", err)
	}

	normalizeTelemetryRates(&input)
	for i := range input.Evidence {
		if input.Evidence[i].Type == "" {
			input.Evidence[i].Type = "telemetry"
		}
		if input.Evidence[i].Source == "" {
			input.Evidence[i].Source = filePath
		}
		input.Evidence[i].IngestedAt = currentTime().Format(time.RFC3339)
	}
	if len(input.Evidence) == 0 {
		input.Evidence = []EvidenceItem{{
			ID:         "EXECUTION-001",
			Type:       "telemetry",
			Source:     filePath,
			Status:     "pass",
			Summary:    "Telemetry snapshot ingested.",
			IngestedAt: currentTime().Format(time.RFC3339),
		}}
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
		for _, element := range feature.scenarios() {
			if !element.isScenario() {
				continue
			}
			total++
			status := cucumberScenarioStatus(element)

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

func cucumberScenarioStatus(element cucumberElement) string {
	if len(element.Steps) == 0 {
		return "fail"
	}
	for _, step := range element.Steps {
		if strings.ToLower(strings.TrimSpace(step.Result.Status)) != "passed" {
			return "fail"
		}
	}
	return "pass"
}

func unmatchedBehaviorWarnings(rootPath string, artifact AssuranceArtifact) []string {
	behaviorIDs := behaviorIDsFromSpec(rootPath)
	if len(behaviorIDs) == 0 {
		return nil
	}
	covered := map[string]bool{}
	for _, evidence := range artifact.Evidence {
		for _, ref := range evidence.Refs {
			covered[ref] = true
		}
	}
	var warnings []string
	for _, id := range behaviorIDs {
		if !covered[id] {
			warnings = append(warnings, fmt.Sprintf("behavior %s has no matching Cucumber scenario evidence", id))
		}
	}
	return warnings
}

func behaviorIDsFromSpec(rootPath string) []string {
	content, err := os.ReadFile(filepath.Join(rootPath, "bottleneck", "behavior", "behavior-spec.md"))
	if err != nil {
		return nil
	}
	matches := regexp.MustCompile(`\bBEHAVIOR-[0-9]{3,}\b`).FindAllString(string(content), -1)
	return uniqueStrings(matches)
}

func normalizeSARIF(report sarifReport, source string) SecurityArtifact {
	findings := map[string]int{
		"critical": 0,
		"high":     0,
		"medium":   0,
		"low":      0,
		"note":     0,
		"unknown":  0,
	}
	violations := 0
	evidence := []EvidenceItem{}

	for _, run := range report.Runs {
		rules := sarifRulesByID(run)
		for _, result := range run.Results {
			rule := rules[result.RuleID]
			severity := determineSeverity(result, rule)
			findings[severity]++
			violations++
			path, line := extractSarifLocation(result)
			message := result.Message.Text
			if message == "" {
				message = result.RuleID
			}
			refs := extractSARIFRefs(result.Properties, rule.Properties)

			evidence = append(evidence, EvidenceItem{
				ID:         fmt.Sprintf("SECURITY-%03d", len(evidence)+1),
				Type:       "sarif",
				Source:     source,
				Refs:       refs,
				Severity:   severity,
				RuleID:     result.RuleID,
				RuleName:   rule.Name,
				Path:       path,
				Line:       line,
				Status:     "fail",
				Summary:    message,
				IngestedAt: currentTime().Format(time.RFC3339),
			})
		}
	}
	if violations == 0 {
		evidence = append(evidence, EvidenceItem{
			ID:         "SECURITY-001",
			Type:       "sarif",
			Source:     source,
			Status:     "pass",
			Summary:    "SARIF scan completed with no findings.",
			IngestedAt: currentTime().Format(time.RFC3339),
		})
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

func sarifRulesByID(run sarifRun) map[string]sarifRule {
	rules := map[string]sarifRule{}
	for _, rule := range run.Tool.Driver.Rules {
		if rule.ID != "" {
			rules[rule.ID] = rule
		}
	}
	return rules
}

func determineSeverity(result sarifResult, rule sarifRule) string {
	if sev := extractStringProperty(result.Properties, "security-severity"); sev != "" {
		return normalizeSeverity(sev)
	}
	if sev := extractStringProperty(result.Properties, "problem.severity"); sev != "" {
		return normalizeSeverity(sev)
	}
	if sev := extractStringProperty(result.Properties, "severity"); sev != "" {
		return normalizeSeverity(sev)
	}
	if sev := extractStringProperty(result.Properties, "level"); sev != "" {
		return normalizeSeverity(sev)
	}
	if sev := extractStringProperty(rule.Properties, "security-severity"); sev != "" {
		return normalizeSeverity(sev)
	}
	if sev := extractStringProperty(rule.Properties, "problem.severity"); sev != "" {
		return normalizeSeverity(sev)
	}
	if sev := extractStringProperty(rule.Properties, "severity"); sev != "" {
		return normalizeSeverity(sev)
	}
	if result.Level != "" {
		return normalizeSeverity(result.Level)
	}
	if rule.DefaultConfiguration.Level != "" {
		return normalizeSeverity(rule.DefaultConfiguration.Level)
	}
	return "unknown"
}

func normalizeTelemetryRates(input *ExecutionArtifact) {
	input.ChangeFailureRate = normalizeRatio(input.ChangeFailureRate)
	input.AdoptionRate = normalizeRatio(input.AdoptionRate)
	input.ErrorRate = normalizeRatio(input.ErrorRate)
	if input.UserOverrideRate != nil {
		normalized := normalizeRatio(*input.UserOverrideRate)
		input.UserOverrideRate = &normalized
	}
	if input.Cost != nil {
		input.Cost.BudgetVariance = normalizeRatio(input.Cost.BudgetVariance)
	}
}

func normalizeRatio(value float64) float64 {
	absolute := value
	if absolute < 0 {
		absolute = -absolute
	}
	if absolute > 1 && absolute <= 100 {
		return value / 100
	}
	return value
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
	if f, ok := value.(float64); ok {
		return strconv.FormatFloat(f, 'f', -1, 64)
	}
	return ""
}

func extractSARIFRefs(propertySets ...map[string]interface{}) []string {
	var refs []string
	keys := []string{
		"refs",
		"ref",
		"references",
		"bottleneck.refs",
		"bottleneckRefs",
		"evidence_refs",
		"evidenceRefs",
		"behavior_refs",
		"behaviorRefs",
		"tags",
	}
	for _, properties := range propertySets {
		if properties == nil {
			continue
		}
		for _, key := range keys {
			refs = append(refs, extractEvidenceRefsFromValue(properties[key])...)
		}
	}
	return uniqueStrings(refs)
}

func extractEvidenceRefsFromValue(value interface{}) []string {
	switch v := value.(type) {
	case string:
		return evidenceIDPattern.FindAllString(v, -1)
	case []interface{}:
		var refs []string
		for _, item := range v {
			refs = append(refs, extractEvidenceRefsFromValue(item)...)
		}
		return refs
	case []string:
		var refs []string
		for _, item := range v {
			refs = append(refs, evidenceIDPattern.FindAllString(item, -1)...)
		}
		return refs
	default:
		return nil
	}
}

func normalizeSeverity(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if numeric, err := strconv.ParseFloat(value, 64); err == nil {
		switch {
		case numeric >= 9.0:
			return "critical"
		case numeric >= 7.0:
			return "high"
		case numeric >= 4.0:
			return "medium"
		case numeric > 0:
			return "low"
		default:
			return "note"
		}
	}
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
	return "unknown"
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
	Name      string            `json:"name"`
	Elements  []cucumberElement `json:"elements"`
	Scenarios []cucumberElement `json:"scenarios"`
	Tags      []tag             `json:"tags"`
}

type cucumberElement struct {
	Name  string         `json:"name"`
	Type  string         `json:"type"`
	Steps []cucumberStep `json:"steps"`
	Tags  []tag          `json:"tags"`
}

func (f cucumberFeature) scenarios() []cucumberElement {
	if len(f.Elements) == 0 {
		return f.Scenarios
	}
	return append(append([]cucumberElement{}, f.Elements...), f.Scenarios...)
}

func (e cucumberElement) isScenario() bool {
	switch strings.ToLower(strings.TrimSpace(e.Type)) {
	case "", "scenario", "scenario_outline":
		return true
	default:
		return false
	}
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
	Tool    sarifTool     `json:"tool"`
}

type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

type sarifDriver struct {
	Rules []sarifRule `json:"rules"`
}

type sarifRule struct {
	ID                   string                 `json:"id"`
	Name                 string                 `json:"name"`
	Properties           map[string]interface{} `json:"properties"`
	DefaultConfiguration sarifDefaultConfig     `json:"defaultConfiguration"`
}

type sarifDefaultConfig struct {
	Level string `json:"level"`
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

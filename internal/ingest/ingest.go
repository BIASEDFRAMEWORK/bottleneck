package ingest

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
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
	ID          string   `json:"id,omitempty"`
	Type        string   `json:"type"`
	Source      string   `json:"source"`
	Refs        []string `json:"refs,omitempty"`
	Severity    string   `json:"severity,omitempty"`
	RuleID      string   `json:"rule_id,omitempty"`
	RuleName    string   `json:"rule_name,omitempty"`
	Path        string   `json:"path,omitempty"`
	Line        *int     `json:"line,omitempty"`
	Status      string   `json:"status"`
	Summary     string   `json:"summary,omitempty"`
	GeneratedBy string   `json:"generated_by,omitempty"`
	GeneratedAt string   `json:"generated_at,omitempty"`
	IngestedAt  string   `json:"ingested_at,omitempty"`
	Confidence  string   `json:"confidence,omitempty"`
	Provenance  string   `json:"provenance,omitempty"`
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
	case "junit":
		return "reports/junit.xml"
	case "coverage":
		return "coverage/lcov.info"
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

func IngestJUnit(rootPath, filePath, outPath string, merge, dryRun bool) (IngestSummary, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return IngestSummary{}, ingestInputError("junit", filePath, "read junit file", err)
	}

	artifact, warnings, err := normalizeJUnit(content, filePath)
	if err != nil {
		return IngestSummary{}, ingestInputError("junit", filePath, "parse junit xml", err)
	}
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

func IngestCoverage(rootPath, filePath, outPath string, merge, dryRun bool) (IngestSummary, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return IngestSummary{}, ingestInputError("coverage", filePath, "read coverage file", err)
	}
	defer file.Close()

	artifact, warnings, err := normalizeLCOV(file, filePath)
	if err != nil {
		return IngestSummary{}, ingestInputError("coverage", filePath, "parse lcov file", err)
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

	return IngestSummary{Artifact: artifact, Warnings: warnings}, nil
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
		enrichEvidence(&input.Evidence[i], "telemetry", confidenceForRefs(input.Evidence[i].Refs))
	}
	if len(input.Evidence) == 0 {
		item := EvidenceItem{
			ID:      "EXECUTION-001",
			Type:    "telemetry",
			Source:  filePath,
			Status:  "pass",
			Summary: "Telemetry snapshot ingested.",
		}
		enrichEvidence(&item, "telemetry", "medium")
		input.Evidence = []EvidenceItem{item}
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

	if allIncomingEvidenceAlreadyPresent(existing.Evidence, artifact.Evidence) {
		return existing, nil
	}

	merged := artifact
	if artifact.ScenariosTotal == 0 && existing.ScenariosTotal > 0 {
		merged.ScenariosTotal = existing.ScenariosTotal
		merged.ScenariosPassed = existing.ScenariosPassed
		merged.ScenariosFailed = existing.ScenariosFailed
		merged.Failures = existing.Failures
		merged.TestsSkipped = existing.TestsSkipped
	} else if artifact.ScenariosTotal > 0 && existing.ScenariosTotal > 0 {
		merged.ScenariosTotal = existing.ScenariosTotal + artifact.ScenariosTotal
		merged.ScenariosPassed = existing.ScenariosPassed + artifact.ScenariosPassed
		merged.ScenariosFailed = existing.ScenariosFailed + artifact.ScenariosFailed
		merged.Failures = append(existing.Failures, artifact.Failures...)
		if existing.TestsSkipped != nil || artifact.TestsSkipped != nil {
			totalSkipped := intValue(existing.TestsSkipped) + intValue(artifact.TestsSkipped)
			merged.TestsSkipped = &totalSkipped
		}
	}
	if artifact.Coverage == nil && existing.Coverage != nil {
		merged.Coverage = existing.Coverage
	}
	merged.Evidence = mergeEvidence(existing.Evidence, artifact.Evidence)
	return merged, nil
}

func allIncomingEvidenceAlreadyPresent(existing, incoming []EvidenceItem) bool {
	if len(incoming) == 0 {
		return false
	}
	index := map[string]struct{}{}
	for _, item := range existing {
		if key := evidenceKey(item); key != "" {
			index[key] = struct{}{}
		}
	}
	for _, item := range incoming {
		key := evidenceKey(item)
		if key == "" {
			return false
		}
		if _, ok := index[key]; !ok {
			return false
		}
	}
	return true
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
		if key := evidenceKey(item); key != "" {
			index[key] = struct{}{}
		}
		merged = append(merged, item)
	}
	for _, item := range incoming {
		if key := evidenceKey(item); key != "" {
			if _, ok := index[key]; ok {
				continue
			}
			index[key] = struct{}{}
		}
		merged = append(merged, item)
	}
	return merged
}

func evidenceKey(item EvidenceItem) string {
	if item.ID == "" {
		return ""
	}
	if item.Source != "" {
		return item.Source + "\x00" + item.ID
	}
	return item.ID
}

func intValue(value *int) int {
	if value == nil {
		return 0
	}
	return *value
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
			refs = append(refs, extractEvidenceRefsFromText(feature.Name+" "+element.Name)...)
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

			item := EvidenceItem{
				ID:      fmt.Sprintf("ASSURANCE-%03d", len(evidence)+1),
				Type:    "cucumber",
				Source:  source,
				Refs:    refs,
				Status:  status,
				Summary: summary,
			}
			enrichEvidence(&item, "cucumber", confidenceForRefs(refs))
			evidence = append(evidence, item)
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

			item := EvidenceItem{
				ID:       fmt.Sprintf("SECURITY-%03d", len(evidence)+1),
				Type:     "sarif",
				Source:   source,
				Refs:     refs,
				Severity: severity,
				RuleID:   result.RuleID,
				RuleName: rule.Name,
				Path:     path,
				Line:     line,
				Status:   "fail",
				Summary:  message,
			}
			enrichEvidence(&item, "sarif", confidenceForRefs(refs))
			evidence = append(evidence, item)
		}
	}
	if violations == 0 {
		item := EvidenceItem{
			ID:      "SECURITY-001",
			Type:    "sarif",
			Source:  source,
			Status:  "pass",
			Summary: "SARIF scan completed with no findings.",
		}
		enrichEvidence(&item, "sarif", "medium")
		evidence = append(evidence, item)
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
		enrichEvidence(&item, "test-summary", confidenceForRefs(item.Refs))
		artifact.Evidence = append(artifact.Evidence, item)
	}

	return artifact, nil
}

func normalizeJUnit(content []byte, source string) (AssuranceArtifact, []string, error) {
	var suites junitTestSuites
	if err := xml.Unmarshal(content, &suites); err != nil {
		var suite junitTestSuite
		if suiteErr := xml.Unmarshal(content, &suite); suiteErr != nil {
			return AssuranceArtifact{}, nil, err
		}
		suites = junitTestSuites{
			Tests:      suite.Tests,
			Failures:   suite.Failures,
			Errors:     suite.Errors,
			Skipped:    suite.Skipped,
			TestSuites: []junitTestSuite{suite},
		}
	}

	testCases := suites.allTestCases()
	total := 0
	passed := 0
	failed := 0
	skipped := 0
	failures := []string{}
	evidence := []EvidenceItem{}
	warnings := []string{}

	for _, testCase := range testCases {
		total++
		status := "pass"
		if testCase.isFailed() {
			status = "fail"
			failed++
			failures = append(failures, testCase.displayName())
		} else if testCase.isSkipped() {
			status = "warn"
			skipped++
		} else {
			passed++
		}
		refs := extractEvidenceRefsFromText(testCase.ClassName + " " + testCase.Name)
		if len(refs) == 0 {
			warnings = append(warnings, fmt.Sprintf("unmapped JUnit test %q has no evidence refs", testCase.displayName()))
		}

		item := EvidenceItem{
			ID:      fmt.Sprintf("ASSURANCE-%03d", len(evidence)+1),
			Type:    "junit",
			Source:  source,
			Refs:    refs,
			Status:  status,
			Summary: testCase.displayName(),
		}
		enrichEvidence(&item, "junit", confidenceForRefs(refs))
		evidence = append(evidence, item)
	}

	if len(testCases) == 0 {
		total = suites.Tests
		failed = suites.Failures + suites.Errors
		skipped = suites.Skipped
		passed = total - failed - skipped
		if passed < 0 {
			passed = 0
		}
		status := "pass"
		if failed > 0 {
			status = "fail"
			failures = append(failures, fmt.Sprintf("%d JUnit tests failed", failed))
		} else if skipped > 0 {
			status = "warn"
		}
		item := EvidenceItem{
			ID:      "ASSURANCE-001",
			Type:    "junit",
			Source:  source,
			Status:  status,
			Summary: "JUnit test suite summary ingested.",
		}
		enrichEvidence(&item, "junit", "medium")
		evidence = append(evidence, item)
	}

	if total == 0 {
		warnings = append(warnings, "no JUnit tests found")
	}

	skippedPtr := skipped
	return AssuranceArtifact{
		ScenariosTotal:  total,
		ScenariosPassed: passed,
		ScenariosFailed: failed,
		Failures:        failures,
		TestsSkipped:    &skippedPtr,
		Evidence:        evidence,
	}, warnings, nil
}

func normalizeLCOV(file *os.File, source string) (AssuranceArtifact, []string, error) {
	scanner := bufio.NewScanner(file)
	executable := 0
	covered := 0
	currentFile := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		switch {
		case strings.HasPrefix(line, "SF:"):
			currentFile = strings.TrimSpace(strings.TrimPrefix(line, "SF:"))
		case strings.HasPrefix(line, "DA:"):
			parts := strings.Split(strings.TrimPrefix(line, "DA:"), ",")
			if len(parts) < 2 {
				continue
			}
			hits, err := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				continue
			}
			executable++
			if hits > 0 {
				covered++
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return AssuranceArtifact{}, nil, err
	}

	warnings := []string{}
	ratio := 0.0
	if executable > 0 {
		ratio = float64(covered) / float64(executable)
	} else {
		warnings = append(warnings, "no executable LCOV lines found")
	}

	status := "warn"
	if ratio >= 0.80 {
		status = "pass"
	}
	summary := fmt.Sprintf("LCOV line coverage %.1f%% (%d/%d lines).", ratio*100, covered, executable)
	refs := extractEvidenceRefsFromText(source + " " + currentFile)
	item := EvidenceItem{
		ID:      "ASSURANCE-COVERAGE-001",
		Type:    "coverage",
		Source:  source,
		Refs:    refs,
		Status:  status,
		Summary: summary,
	}
	enrichEvidence(&item, "lcov", "medium")

	return AssuranceArtifact{
		Coverage: &ratio,
		Evidence: []EvidenceItem{
			item,
		},
	}, warnings, nil
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

func extractEvidenceRefsFromText(text string) []string {
	return uniqueStrings(evidenceIDPattern.FindAllString(text, -1))
}

func enrichEvidence(item *EvidenceItem, generatedBy string, confidence string) {
	if item.IngestedAt == "" {
		item.IngestedAt = currentTime().Format(time.RFC3339)
	}
	if item.GeneratedBy == "" {
		item.GeneratedBy = generatedBy
	}
	if item.GeneratedAt == "" {
		item.GeneratedAt = generatedAtForSource(item.Source)
	}
	if item.Confidence == "" {
		item.Confidence = confidence
	}
	if item.Provenance == "" && item.Source != "" {
		item.Provenance = fmt.Sprintf("Normalized from %s by bottleneck ingest.", item.Source)
	}
}

func generatedAtForSource(source string) string {
	if source == "" {
		return ""
	}
	info, err := os.Stat(source)
	if err != nil {
		return ""
	}
	return info.ModTime().UTC().Format(time.RFC3339)
}

func confidenceForRefs(refs []string) string {
	if len(refs) > 0 {
		return "high"
	}
	return "medium"
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

type junitTestSuites struct {
	XMLName    xml.Name         `xml:"testsuites"`
	Tests      int              `xml:"tests,attr"`
	Failures   int              `xml:"failures,attr"`
	Errors     int              `xml:"errors,attr"`
	Skipped    int              `xml:"skipped,attr"`
	TestSuites []junitTestSuite `xml:"testsuite"`
	TestCases  []junitTestCase  `xml:"testcase"`
}

func (s junitTestSuites) allTestCases() []junitTestCase {
	cases := append([]junitTestCase{}, s.TestCases...)
	for _, suite := range s.TestSuites {
		cases = append(cases, suite.allTestCases()...)
	}
	return cases
}

type junitTestSuite struct {
	Name       string           `xml:"name,attr"`
	Tests      int              `xml:"tests,attr"`
	Failures   int              `xml:"failures,attr"`
	Errors     int              `xml:"errors,attr"`
	Skipped    int              `xml:"skipped,attr"`
	TestCases  []junitTestCase  `xml:"testcase"`
	TestSuites []junitTestSuite `xml:"testsuite"`
}

func (s junitTestSuite) allTestCases() []junitTestCase {
	cases := append([]junitTestCase{}, s.TestCases...)
	for _, suite := range s.TestSuites {
		cases = append(cases, suite.allTestCases()...)
	}
	return cases
}

type junitTestCase struct {
	Name      string             `xml:"name,attr"`
	ClassName string             `xml:"classname,attr"`
	Failures  []junitTestFailure `xml:"failure"`
	Errors    []junitTestFailure `xml:"error"`
	Skipped   []struct{}         `xml:"skipped"`
}

type junitTestFailure struct {
	Message string `xml:"message,attr"`
	Text    string `xml:",chardata"`
}

func (c junitTestCase) isFailed() bool {
	return len(c.Failures) > 0 || len(c.Errors) > 0
}

func (c junitTestCase) isSkipped() bool {
	return len(c.Skipped) > 0
}

func (c junitTestCase) displayName() string {
	if c.ClassName == "" {
		return c.Name
	}
	if c.Name == "" {
		return c.ClassName
	}
	return c.ClassName + " " + c.Name
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

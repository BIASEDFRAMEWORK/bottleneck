package scorecard

import (
	"encoding/json"
	"fmt"
	"strings"

	"bottleneck/internal/diagnosis"
	"bottleneck/internal/githubactions"
	"bottleneck/internal/models"
	"bottleneck/internal/prrisk"
)

const (
	SchemaVersion = "scorecard.v2"

	FormatText     = "text"
	FormatJSON     = "json"
	FormatMarkdown = "markdown"

	ViewExecutive   = "executive"
	ViewEngineering = "engineering"
	ViewGovernance  = "governance"

	StatusPass    = "PASS"
	StatusWarn    = "WARN"
	StatusFail    = "FAIL"
	StatusUnknown = "UNKNOWN"

	RecommendationProceed     = "Proceed"
	RecommendationConditional = "Conditional"
	RecommendationBlock       = "Block"
	RecommendationUnknown     = "Unknown"
)

type Scorecard struct {
	SchemaVersion         string                     `json:"schema_version"`
	Environment           string                     `json:"environment"`
	SystemStatus          string                     `json:"system_status"`
	ReleaseRecommendation string                     `json:"release_recommendation"`
	PrimaryBottleneck     string                     `json:"primary_bottleneck"`
	EffectiveThresholds   models.EffectiveThresholds `json:"effective_thresholds"`
	Diagnosis             diagnosis.Diagnosis        `json:"diagnosis"`
	GitHub                *githubactions.Metadata    `json:"github,omitempty"`
	PullRequestRisk       []prrisk.Signal            `json:"pull_request_risk,omitempty"`
	Capabilities          []CapabilityScorecard      `json:"capabilities"`
	BottomLine            string                     `json:"bottom_line"`
}

type CapabilityScorecard struct {
	Capability        string               `json:"capability"`
	Status            string               `json:"status"`
	Score             int                  `json:"score"`
	Owner             string               `json:"owner"`
	Bottleneck        string               `json:"bottleneck"`
	EvidenceCount     int                  `json:"evidence_count"`
	MissingEvidence   []string             `json:"missing_evidence"`
	ScoreImpacts      []models.ScoreImpact `json:"score_impacts,omitempty"`
	Reason            string               `json:"reason"`
	RecommendedAction string               `json:"recommended_action"`
	Evidence          []string             `json:"evidence"`
}

type capabilityMetadata struct {
	owner             string
	bottleneck        string
	passingArtifact   string
	missingEvidence   string
	recommendedAction string
	passAction        string
}

type Options struct {
	View   string
	GitHub *githubactions.Metadata
}

var metadataByCapability = map[string]capabilityMetadata{
	"Intent": {
		owner:             "Intent Engineer",
		bottleneck:        "Ambiguous requirements",
		passingArtifact:   "bottleneck/intent/intent.md",
		missingEvidence:   "Add validation details for bottleneck/intent/intent.md showing concrete outcomes, constraints, and success criteria.",
		recommendedAction: "Update intent.md with concrete outcomes, constraints, and measurable success criteria.",
		passAction:        "Keep intent.md aligned with current release goals.",
	},
	"Behavior": {
		owner:             "Behavior Engineer",
		bottleneck:        "Non-deterministic outputs",
		passingArtifact:   "bottleneck/behavior/behavior-spec.md",
		missingEvidence:   "Add validation details for bottleneck/behavior/behavior-spec.md showing expected and unacceptable behavior.",
		recommendedAction: "Update behavior-spec.md with concrete expected and unacceptable behavior evidence.",
		passAction:        "Keep behavior-spec.md current as behavior changes.",
	},
	"Design": {
		owner:             "Design Engineer",
		bottleneck:        "Poor adoption / UX gaps",
		passingArtifact:   "bottleneck/design/architecture.md",
		missingEvidence:   "Add validation details for bottleneck/design/architecture.md showing the architecture is reviewable.",
		recommendedAction: "Expand architecture.md with concrete architecture and operational design evidence.",
		passAction:        "Keep architecture.md synchronized with implementation changes.",
	},
	"Assurance": {
		owner:             "Assurance Engineer",
		bottleneck:        "Validation gaps",
		passingArtifact:   "bottleneck/assurance/results.json",
		missingEvidence:   "Add assurance metrics from bottleneck/assurance/results.json.",
		recommendedAction: "Fix failing scenarios or regenerate external BDD results.",
		passAction:        "Keep assurance results current with the latest test run.",
	},
	"Security": {
		owner:             "Security Engineer",
		bottleneck:        "Risk & compliance",
		passingArtifact:   "bottleneck/security/guardrails.json",
		missingEvidence:   "Add guardrail evidence from bottleneck/security/guardrails.json.",
		recommendedAction: "Remove policy violations or regenerate guardrail evidence.",
		passAction:        "Keep guardrail evidence current with security policy changes.",
	},
	"Execution": {
		owner:             "Execution Engineer",
		bottleneck:        "Delivery friction",
		passingArtifact:   "bottleneck/execution/telemetry.json",
		missingEvidence:   "Add telemetry evidence from bottleneck/execution/telemetry.json.",
		recommendedAction: "Review telemetry and improve reliability or adoption before release.",
		passAction:        "Keep telemetry evidence current for the selected environment.",
	},
	"Traceability": {
		owner:             "Release Engineer",
		bottleneck:        "Traceability gaps",
		passingArtifact:   "bottleneck/*",
		missingEvidence:   "Add evidence IDs and Refs links across intent, behavior, assurance, security, and execution artifacts.",
		recommendedAction: "Run bottleneck trace --id <ID> for the affected ID and repair missing, duplicate, or orphaned evidence links.",
		passAction:        "Keep evidence IDs and Refs links current as release evidence changes.",
	},
	"Config": {
		owner:             "Execution Engineer",
		bottleneck:        "Delivery friction",
		passingArtifact:   "bottleneck/config.yaml",
		missingEvidence:   "Repair bottleneck/config.yaml so effective thresholds can be resolved.",
		recommendedAction: "Repair bottleneck/config.yaml and rerun the scorecard.",
		passAction:        "Keep config.yaml thresholds aligned with release policy.",
	},
}

func Build(result models.EngineResult) Scorecard {
	return BuildWithOptions(result, Options{})
}

func BuildWithOptions(result models.EngineResult, options Options) Scorecard {
	diagnosisResult := diagnosis.Analyze(result)
	capabilities := make([]CapabilityScorecard, 0, len(result.Results))
	for _, validation := range result.Results {
		score := diagnosis.ScoreFor(validation.Capability, diagnosisResult.CategoryScores)
		if !diagnosis.HasCategoryScore(validation.Capability, diagnosisResult.CategoryScores) {
			score = diagnosis.ScoreValidation(validation)
		}
		capabilities = append(capabilities, buildCapability(validation, score))
	}

	card := Scorecard{
		SchemaVersion:         SchemaVersion,
		Environment:           result.Environment,
		SystemStatus:          displayStatus(result.SystemStatus),
		PrimaryBottleneck:     diagnosisResult.PrimaryBottleneck,
		EffectiveThresholds:   result.EffectiveThresholds,
		Diagnosis:             diagnosisResult,
		Capabilities:          capabilities,
		ReleaseRecommendation: releaseRecommendationFor(capabilities, displayStatus(result.SystemStatus)),
	}
	if options.GitHub != nil && options.GitHub.Detected {
		metadata := *options.GitHub
		card.GitHub = &metadata
		card.PullRequestRisk = prrisk.Assess(metadata)
	}
	card.BottomLine = bottomLine(card)

	return card
}

func Render(result models.EngineResult, format string, viewValues ...string) (string, error) {
	view := ""
	if len(viewValues) > 0 {
		view = viewValues[0]
	}

	return RenderWithOptions(result, format, Options{View: view})
}

func RenderWithOptions(result models.EngineResult, format string, options Options) (string, error) {
	view, err := normalizeView(options.View)
	if err != nil {
		return "", err
	}

	card := BuildWithOptions(result, options)

	switch strings.ToLower(format) {
	case FormatText:
		return renderText(card, view), nil
	case FormatJSON:
		return renderJSON(card)
	case FormatMarkdown:
		return renderMarkdown(card, view), nil
	default:
		return "", fmt.Errorf("unsupported format %q (supported: text, json, markdown)", format)
	}
}

func buildCapability(validation models.ValidationResult, score int) CapabilityScorecard {
	meta := metadataFor(validation.Capability)
	evidence := evidenceItems(validation)
	missingEvidence := missingEvidenceFor(validation, evidence, meta)

	return CapabilityScorecard{
		Capability:        validation.Capability,
		Status:            displayStatus(validation.Status),
		Score:             score,
		Owner:             meta.owner,
		Bottleneck:        meta.bottleneck,
		EvidenceCount:     len(evidence),
		MissingEvidence:   missingEvidence,
		ScoreImpacts:      validation.EvidenceQuality.ScoreImpacts,
		Reason:            reasonFor(validation),
		RecommendedAction: recommendedActionFor(validation, meta),
		Evidence:          evidence,
	}
}

func renderText(card Scorecard, view string) string {
	switch view {
	case ViewExecutive:
		return renderExecutiveText(card)
	case ViewGovernance:
		return renderGovernanceText(card)
	default:
		return renderEngineeringText(card)
	}
}

func renderEngineeringText(card Scorecard) string {
	var lines []string
	lines = append(lines, gaugeScorecardHeader(card)...)
	lines = appendGitHubText(lines, card)
	lines = append(lines, "", "Category Gauges:")
	lines = appendGaugeLines(lines, card)
	lines = append(lines, "", "Effective Thresholds:")
	lines = append(lines, thresholdLines(card.EffectiveThresholds, "  ")...)

	lines = append(lines, "", "Capability Details:")
	for _, capability := range card.Capabilities {
		lines = append(lines, "")
		lines = append(lines,
			fmt.Sprintf("%s:", capability.Capability),
			fmt.Sprintf("  Status: %s", capability.Status),
			fmt.Sprintf("  Score: %d", capability.Score),
			fmt.Sprintf("  Owner: %s", capability.Owner),
			fmt.Sprintf("  Bottleneck: %s", capability.Bottleneck),
			fmt.Sprintf("  Reason: %s", capability.Reason),
			fmt.Sprintf("  Recommended Action: %s", capability.RecommendedAction),
			"  Evidence:",
		)
		lines = appendBulletLines(lines, capability.Evidence, "    ", "None reported.")
		lines = append(lines, "  Missing Evidence:")
		lines = appendBulletLines(lines, capability.MissingEvidence, "    ", "None.")
		lines = append(lines, "  Score Impacts:")
		lines = appendScoreImpactLines(lines, capability.ScoreImpacts, "    ")
	}

	lines = append(lines, "", "Bottom line:", card.BottomLine)
	return strings.Join(lines, "\n")
}

func gaugeScorecardHeader(card Scorecard) []string {
	lines := []string{
		"Bottleneck Scorecard",
		"",
		fmt.Sprintf("Environment: %s", card.Environment),
		fmt.Sprintf("System Status: %s", card.SystemStatus),
		fmt.Sprintf("Release Recommendation: %s", card.ReleaseRecommendation),
		"",
		fmt.Sprintf("Primary Bottleneck: %s", card.PrimaryBottleneck),
	}
	if len(card.Diagnosis.TiedBottlenecks) > 0 {
		lines = append(lines, fmt.Sprintf("Tied Bottlenecks: %s", strings.Join(card.Diagnosis.TiedBottlenecks, ", ")))
	}
	if card.Diagnosis.Confidence != "" {
		lines = append(lines, fmt.Sprintf("Diagnosis Confidence: %s", card.Diagnosis.Confidence))
	}
	lines = append(lines,
		"",
		"Overall Diagnosis:",
		fmt.Sprintf("Why: %s", card.Diagnosis.WhyItMatters),
		"",
		"Next action:",
		card.Diagnosis.RecommendedAction,
	)
	return lines
}

func renderExecutiveText(card Scorecard) string {
	var lines []string
	lines = append(lines,
		"bottleneck SDLC Scorecard - Executive View",
		"",
		fmt.Sprintf("Environment: %s", card.Environment),
		fmt.Sprintf("System Status: %s", card.SystemStatus),
		fmt.Sprintf("Release Recommendation: %s", card.ReleaseRecommendation),
		fmt.Sprintf("Primary Bottleneck: %s", card.PrimaryBottleneck),
	)
	lines = appendDiagnosisText(lines, card)
	lines = appendGitHubText(lines, card)
	lines = append(lines, "", "Capability Status Summary:")

	for _, line := range statusSummaryLines(card) {
		lines = append(lines, "  "+line)
	}

	lines = append(lines, "", "Capabilities:")
	for _, capability := range card.Capabilities {
		lines = append(lines, fmt.Sprintf("  %s: %s", capability.Capability, capability.Status))
	}

	lines = append(lines, "", "Bottom line:", card.BottomLine)
	return strings.Join(lines, "\n")
}

func renderGovernanceText(card Scorecard) string {
	var lines []string
	lines = append(lines,
		"bottleneck SDLC Scorecard - Governance View",
		"",
		fmt.Sprintf("Environment: %s", card.Environment),
		fmt.Sprintf("System Status: %s", card.SystemStatus),
		fmt.Sprintf("Release Recommendation: %s", card.ReleaseRecommendation),
		fmt.Sprintf("Primary Bottleneck: %s", card.PrimaryBottleneck),
	)
	lines = appendDiagnosisText(lines, card)
	lines = appendGitHubText(lines, card)
	lines = append(lines, "", "Effective Thresholds:")
	lines = append(lines, thresholdLines(card.EffectiveThresholds, "  ")...)
	lines = append(lines, "", "Governance Signals:")

	for _, capabilityName := range []string{"Security", "Assurance", "Execution"} {
		capability, ok := capabilityByName(card, capabilityName)
		if !ok {
			lines = append(lines, fmt.Sprintf("  %s: UNKNOWN - validation evidence unavailable", capabilityName))
			continue
		}
		lines = append(lines, fmt.Sprintf("  %s: %s - %s", capability.Capability, capability.Status, capability.Reason))
	}

	lines = append(lines, "  Governance Evidence: not assessed (no governance artifact exists yet)")
	lines = append(lines, "", "Missing Evidence Blocking Or Conditioning Release:")
	missing := governanceMissingEvidence(card)
	lines = appendBulletLines(lines, missing, "  ", "None.")
	lines = append(lines, "", "Release Decision Summary:", card.BottomLine)
	return strings.Join(lines, "\n")
}

func renderJSON(card Scorecard) (string, error) {
	content, err := json.MarshalIndent(card, "", "  ")
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func renderMarkdown(card Scorecard, view string) string {
	switch view {
	case ViewExecutive:
		return renderExecutiveMarkdown(card)
	case ViewGovernance:
		return renderGovernanceMarkdown(card)
	default:
		return renderEngineeringMarkdown(card)
	}
}

func renderEngineeringMarkdown(card Scorecard) string {
	var lines []string
	lines = append(lines, markdownSummary(card)...)
	lines = appendDiagnosisMarkdown(lines, card)
	lines = appendGitHubMarkdown(lines, card)
	lines = append(lines, "", "## Effective Thresholds", "", "| Threshold | Value |", "| --- | ---: |")
	for _, threshold := range thresholdRows(card.EffectiveThresholds) {
		lines = append(lines, fmt.Sprintf("| %s | %s |", threshold.name, threshold.value))
	}
	lines = append(lines, "", "## Capabilities", "", "| Capability | Status | Score | Evidence | Missing Evidence | Recommendation |", "| --- | --- | ---: | ---: | --- | --- |")
	for _, capability := range card.Capabilities {
		lines = append(lines, fmt.Sprintf(
			"| %s | %s | %d | %d | %s | %s |",
			markdownCell(capability.Capability),
			markdownCell(capability.Status),
			capability.Score,
			capability.EvidenceCount,
			markdownCell(joinMarkdownList(capability.MissingEvidence, "None")),
			markdownCell(capability.RecommendedAction),
		))
	}
	lines = append(lines, "", "## Evidence")
	for _, capability := range card.Capabilities {
		lines = append(lines, "", fmt.Sprintf("### %s", capability.Capability))
		lines = append(lines, fmt.Sprintf("- Reason: %s", markdownText(capability.Reason)))
		lines = append(lines, "- Evidence:")
		lines = appendMarkdownBullets(lines, capability.Evidence, "  ", "None reported.")
		lines = append(lines, "- Missing Evidence:")
		lines = appendMarkdownBullets(lines, capability.MissingEvidence, "  ", "None.")
		lines = append(lines, "- Score Impacts:")
		lines = appendMarkdownScoreImpacts(lines, capability.ScoreImpacts, "  ")
	}
	lines = append(lines, "", "## Bottom Line", "", card.BottomLine)
	return strings.Join(lines, "\n")
}

func renderExecutiveMarkdown(card Scorecard) string {
	var lines []string
	lines = append(lines, "# bottleneck Scorecard", "")
	lines = append(lines, "| Field | Value |", "| --- | --- |")
	lines = append(lines,
		fmt.Sprintf("| Environment | %s |", markdownCell(card.Environment)),
		fmt.Sprintf("| System Status | %s |", markdownCell(card.SystemStatus)),
		fmt.Sprintf("| Release Recommendation | %s |", markdownCell(card.ReleaseRecommendation)),
		fmt.Sprintf("| Primary Bottleneck | %s |", markdownCell(card.PrimaryBottleneck)),
	)
	lines = appendDiagnosisMarkdown(lines, card)
	lines = appendGitHubMarkdown(lines, card)
	lines = append(lines, "", "## Capability Status Summary", "", "| Status | Count |", "| --- | ---: |")
	for _, line := range statusSummaryRows(card) {
		lines = append(lines, fmt.Sprintf("| %s | %d |", line.status, line.count))
	}
	lines = append(lines, "", "## Bottom Line", "", card.BottomLine)
	return strings.Join(lines, "\n")
}

func renderGovernanceMarkdown(card Scorecard) string {
	var lines []string
	lines = append(lines, markdownSummary(card)...)
	lines = appendDiagnosisMarkdown(lines, card)
	lines = appendGitHubMarkdown(lines, card)
	lines = append(lines, "", "## Effective Thresholds", "", "| Threshold | Value |", "| --- | ---: |")
	for _, threshold := range thresholdRows(card.EffectiveThresholds) {
		lines = append(lines, fmt.Sprintf("| %s | %s |", threshold.name, threshold.value))
	}
	lines = append(lines, "", "## Governance Signals", "", "| Signal | Status | Reason |", "| --- | --- | --- |")
	for _, capabilityName := range []string{"Security", "Assurance", "Execution"} {
		capability, ok := capabilityByName(card, capabilityName)
		if !ok {
			lines = append(lines, fmt.Sprintf("| %s | UNKNOWN | validation evidence unavailable |", capabilityName))
			continue
		}
		lines = append(lines, fmt.Sprintf("| %s | %s | %s |", markdownCell(capability.Capability), markdownCell(capability.Status), markdownCell(capability.Reason)))
	}
	lines = append(lines, "| Governance Evidence | UNKNOWN | not assessed; no governance artifact exists yet |")
	lines = append(lines, "", "## Missing Evidence Blocking Or Conditioning Release")
	lines = appendMarkdownBullets(lines, governanceMissingEvidence(card), "", "None.")
	lines = append(lines, "", "## Release Decision Summary", "", card.BottomLine)
	return strings.Join(lines, "\n")
}

func markdownSummary(card Scorecard) []string {
	return []string{
		"# bottleneck Scorecard",
		"",
		"| Field | Value |",
		"| --- | --- |",
		fmt.Sprintf("| Environment | %s |", markdownCell(card.Environment)),
		fmt.Sprintf("| System Status | %s |", markdownCell(card.SystemStatus)),
		fmt.Sprintf("| Release Recommendation | %s |", markdownCell(card.ReleaseRecommendation)),
		fmt.Sprintf("| Primary Bottleneck | %s |", markdownCell(card.PrimaryBottleneck)),
	}
}

func scorecardHeader(card Scorecard) []string {
	return []string{
		"bottleneck SDLC Scorecard",
		"",
		fmt.Sprintf("Environment: %s", card.Environment),
		fmt.Sprintf("System Status: %s", card.SystemStatus),
		fmt.Sprintf("Release Recommendation: %s", card.ReleaseRecommendation),
		fmt.Sprintf("Primary Bottleneck: %s", card.PrimaryBottleneck),
	}
}

func appendDiagnosisText(lines []string, card Scorecard) []string {
	lines = append(lines,
		"",
		"Diagnosis:",
		fmt.Sprintf("  Primary Bottleneck: %s", card.Diagnosis.PrimaryBottleneck),
		fmt.Sprintf("  Why: %s", card.Diagnosis.WhyItMatters),
		fmt.Sprintf("  Next Action: %s", card.Diagnosis.RecommendedAction),
	)
	if len(card.Diagnosis.TiedBottlenecks) > 0 {
		lines = append(lines, fmt.Sprintf("  Tied Bottlenecks: %s", strings.Join(card.Diagnosis.TiedBottlenecks, ", ")))
	}
	lines = append(lines, "  Category Scores:")
	for _, score := range card.Diagnosis.CategoryScores {
		lines = append(lines, fmt.Sprintf("    - %s: %d (%s)", score.Category, score.Score, displayStatus(score.Status)))
	}
	if len(card.Diagnosis.CategoryScores) == 0 {
		lines = append(lines, "    - None.")
	}
	return lines
}

func appendDiagnosisMarkdown(lines []string, card Scorecard) []string {
	lines = append(lines,
		"",
		"## Diagnosis",
		"",
		fmt.Sprintf("- **Primary Bottleneck:** %s", markdownText(card.Diagnosis.PrimaryBottleneck)),
		fmt.Sprintf("- **Why:** %s", markdownText(card.Diagnosis.WhyItMatters)),
		fmt.Sprintf("- **Next action:** %s", markdownText(card.Diagnosis.RecommendedAction)),
	)
	if len(card.Diagnosis.TiedBottlenecks) > 0 {
		lines = append(lines, fmt.Sprintf("- **Tied bottlenecks:** %s", markdownText(strings.Join(card.Diagnosis.TiedBottlenecks, ", "))))
	}
	lines = append(lines, "", "| Category | Score | Status |", "| --- | ---: | --- |")
	for _, score := range card.Diagnosis.CategoryScores {
		lines = append(lines, fmt.Sprintf(
			"| %s | %d | %s |",
			markdownCell(score.Category),
			score.Score,
			markdownCell(displayStatus(score.Status)),
		))
	}
	return lines
}

func appendGitHubText(lines []string, card Scorecard) []string {
	if card.GitHub == nil {
		return lines
	}

	lines = append(lines, "", "GitHub Actions:")
	lines = append(lines, fmt.Sprintf("  Event: %s", emptyText(card.GitHub.EventName, "unknown")))
	lines = append(lines, fmt.Sprintf("  Repository: %s", emptyText(card.GitHub.Repository, "unknown")))
	if card.GitHub.RunID != "" {
		lines = append(lines, fmt.Sprintf("  Run ID: %s", card.GitHub.RunID))
	}
	if card.GitHub.PullRequest != nil {
		pr := card.GitHub.PullRequest
		lines = append(lines,
			fmt.Sprintf("  Pull Request: #%d %s", pr.Number, pr.Title),
			fmt.Sprintf("  Base: %s", emptyText(pr.BaseRef, "unknown")),
			fmt.Sprintf("  Head: %s", emptyText(pr.HeadRef, "unknown")),
		)
		if pr.Author != "" {
			lines = append(lines, fmt.Sprintf("  Author: %s", pr.Author))
		}
	}
	lines = append(lines, "  PR Risk Signals:")
	for _, signal := range card.PullRequestRisk {
		evidence := ""
		if signal.Evidence != "" {
			evidence = " (" + signal.Evidence + ")"
		}
		lines = append(lines, fmt.Sprintf("    - %s: %s%s", signal.Level, signal.Message, evidence))
	}
	if len(card.PullRequestRisk) == 0 {
		lines = append(lines, "    - None.")
	}
	for _, warning := range card.GitHub.Warnings {
		lines = append(lines, fmt.Sprintf("    - UNKNOWN: %s", warning))
	}

	return lines
}

func appendGitHubMarkdown(lines []string, card Scorecard) []string {
	if card.GitHub == nil {
		return lines
	}

	lines = append(lines, "", "## GitHub Pull Request Context", "", "| Field | Value |", "| --- | --- |")
	lines = append(lines,
		fmt.Sprintf("| Event | %s |", markdownCell(emptyText(card.GitHub.EventName, "unknown"))),
		fmt.Sprintf("| Repository | %s |", markdownCell(emptyText(card.GitHub.Repository, "unknown"))),
	)
	if card.GitHub.RunID != "" {
		lines = append(lines, fmt.Sprintf("| Run ID | %s |", markdownCell(card.GitHub.RunID)))
	}
	if card.GitHub.PullRequest != nil {
		pr := card.GitHub.PullRequest
		lines = append(lines,
			fmt.Sprintf("| Pull Request | #%d %s |", pr.Number, markdownCell(pr.Title)),
			fmt.Sprintf("| Base | %s |", markdownCell(emptyText(pr.BaseRef, "unknown"))),
			fmt.Sprintf("| Head | %s |", markdownCell(emptyText(pr.HeadRef, "unknown"))),
			fmt.Sprintf("| Author | %s |", markdownCell(emptyText(pr.Author, "unknown"))),
		)
		if pr.ChangedFiles != nil {
			lines = append(lines, fmt.Sprintf("| Changed Files | %d |", *pr.ChangedFiles))
		}
		if pr.Additions != nil && pr.Deletions != nil {
			lines = append(lines, fmt.Sprintf("| Diff Size | +%d / -%d |", *pr.Additions, *pr.Deletions))
		}
	}

	lines = append(lines, "", "## Pull Request Risk Signals")
	if len(card.PullRequestRisk) == 0 && len(card.GitHub.Warnings) == 0 {
		lines = append(lines, "- None.")
		return lines
	}
	for _, signal := range card.PullRequestRisk {
		evidence := ""
		if signal.Evidence != "" {
			evidence = " (" + markdownText(signal.Evidence) + ")"
		}
		lines = append(lines, fmt.Sprintf("- **%s**: %s%s", markdownText(signal.Level), markdownText(signal.Message), evidence))
	}
	for _, warning := range card.GitHub.Warnings {
		lines = append(lines, fmt.Sprintf("- **UNKNOWN**: %s", markdownText(warning)))
	}

	return lines
}

func emptyText(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func thresholdLines(thresholds models.EffectiveThresholds, prefix string) []string {
	rows := thresholdRows(thresholds)
	lines := make([]string, 0, len(rows))
	for _, row := range rows {
		lines = append(lines, fmt.Sprintf("%s%s: %s", prefix, row.name, row.value))
	}
	return lines
}

type thresholdRow struct {
	name  string
	value string
}

func thresholdRows(thresholds models.EffectiveThresholds) []thresholdRow {
	return []thresholdRow{
		{name: "assurance.min_accuracy", value: fmt.Sprintf("%.2f", thresholds.Assurance.MinAccuracy)},
		{name: "assurance.max_failures", value: fmt.Sprintf("%d", thresholds.Assurance.MaxFailures)},
		{name: "execution.max_error_rate", value: fmt.Sprintf("%.2f", thresholds.Execution.MaxErrorRate)},
		{name: "execution.min_adoption", value: fmt.Sprintf("%.2f", thresholds.Execution.MinAdoption)},
		{name: "execution.telemetry.max_age_hours", value: fmt.Sprintf("%d", thresholds.Execution.Telemetry.MaxAgeHours)},
		{name: "execution.telemetry.min_deployments_per_week", value: fmt.Sprintf("%.2f", thresholds.Execution.Telemetry.MinDeploymentsPerWeek)},
		{name: "execution.telemetry.max_change_failure_rate", value: fmt.Sprintf("%.2f", thresholds.Execution.Telemetry.MaxChangeFailureRate)},
		{name: "execution.telemetry.max_error_rate", value: fmt.Sprintf("%.2f", thresholds.Execution.Telemetry.MaxErrorRate)},
		{name: "execution.telemetry.max_user_override_rate", value: fmt.Sprintf("%.2f", thresholds.Execution.Telemetry.MaxUserOverrideRate)},
		{name: "execution.telemetry.min_adoption_rate", value: fmt.Sprintf("%.2f", thresholds.Execution.Telemetry.MinAdoptionRate)},
		{name: "execution.telemetry.max_budget_variance", value: fmt.Sprintf("%.2f", thresholds.Execution.Telemetry.MaxBudgetVariance)},
		{name: "security.sarif.max_critical", value: fmt.Sprintf("%d", thresholds.Security.SARIF.MaxCritical)},
		{name: "security.sarif.max_high", value: fmt.Sprintf("%d", thresholds.Security.SARIF.MaxHigh)},
		{name: "security.sarif.max_medium", value: fmt.Sprintf("%d", thresholds.Security.SARIF.MaxMedium)},
		{name: "security.sarif.max_low", value: fmt.Sprintf("%d", thresholds.Security.SARIF.MaxLow)},
		{name: "security.sarif.fail_on_unknown_severity", value: fmt.Sprintf("%t", thresholds.Security.SARIF.FailOnUnknownSeverity)},
	}
}

func appendGaugeLines(lines []string, card Scorecard) []string {
	scores := map[string]int{}
	for _, score := range card.Diagnosis.CategoryScores {
		scores[score.Category] = score.Score
	}

	for _, category := range []string{"Behavior", "Intent", "Design", "Assurance", "Security", "Execution"} {
		score, ok := scores[category]
		if !ok {
			lines = append(lines, fmt.Sprintf("%-10s [??????????] unknown%s", category, bottleneckMarker(card, category)))
			continue
		}
		lines = append(lines, fmt.Sprintf("%-10s %s %3d%s", category, gauge(score, 10), clampGaugeScore(score), bottleneckMarker(card, category)))
	}

	return lines
}

func gauge(score int, width int) string {
	if width <= 0 {
		return "[]"
	}
	score = clampGaugeScore(score)
	filled := score * width / 100
	if filled > width {
		filled = width
	}
	return "[" + strings.Repeat("#", filled) + strings.Repeat("-", width-filled) + "]"
}

func clampGaugeScore(score int) int {
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

func bottleneckMarker(card Scorecard, category string) string {
	if card.PrimaryBottleneck == "None" {
		return ""
	}
	if len(card.Diagnosis.TiedBottlenecks) > 0 {
		for _, tied := range card.Diagnosis.TiedBottlenecks {
			if tied == category {
				return "  <-- tied bottleneck"
			}
		}
		return ""
	}
	if card.PrimaryBottleneck == category {
		return "  <-- primary bottleneck"
	}
	return ""
}

func evidenceItems(validation models.ValidationResult) []string {
	evidence := append([]string{}, validation.Details...)
	if len(evidence) == 0 && validation.Message != "" && displayStatus(validation.Status) != StatusPass {
		evidence = append(evidence, validation.Message)
	}
	return evidence
}

func missingEvidenceFor(validation models.ValidationResult, evidence []string, meta capabilityMetadata) []string {
	var missing []string
	status := displayStatus(validation.Status)
	message := strings.ToLower(validation.Message)

	missing = append(missing, validation.EvidenceQuality.Missing...)

	for _, detail := range validation.Details {
		if missingDetail := missingEvidenceFromDetail(detail); missingDetail != "" {
			missing = append(missing, missingDetail)
		}
	}

	switch {
	case status == StatusUnknown:
		missing = append(missing, fmt.Sprintf("%s was not assessed by validation.", validation.Capability))
	case strings.Contains(message, "missing or invalid config.yaml"):
		missing = append(missing, "Repair bottleneck/config.yaml so effective thresholds can be resolved.")
	case strings.Contains(message, "missing behavior-spec.md"):
		missing = append(missing, "Create bottleneck/behavior/behavior-spec.md with Expected Behavior and Unacceptable Behavior evidence.")
	case strings.Contains(message, "missing intent.md"):
		missing = append(missing, "Create bottleneck/intent/intent.md with Outcomes, Constraints, and Success Criteria evidence.")
	case strings.Contains(message, "missing architecture.md"):
		missing = append(missing, "Create bottleneck/design/architecture.md with architecture evidence.")
	case strings.Contains(message, "missing results.json"):
		missing = append(missing, "Provide bottleneck/assurance/results.json from the external BDD results.")
	case strings.Contains(message, "missing guardrails.json"):
		missing = append(missing, "Provide bottleneck/security/guardrails.json with guardrail violation evidence.")
	case strings.Contains(message, "missing telemetry.json"):
		missing = append(missing, "Provide bottleneck/execution/telemetry.json with adoption and error-rate evidence.")
	case strings.Contains(message, "required sections missing"):
		missing = append(missing, requiredSectionsMissingEvidence(validation.Capability))
	case strings.Contains(message, "empty"):
		missing = append(missing, fmt.Sprintf("Add substantive evidence to %s.", meta.passingArtifact))
	case strings.Contains(message, "invalid results.json"):
		missing = append(missing, "Regenerate bottleneck/assurance/results.json as valid assurance JSON.")
	case strings.Contains(message, "invalid guardrails.json"):
		missing = append(missing, "Regenerate bottleneck/security/guardrails.json as valid guardrail JSON.")
	case strings.Contains(message, "invalid telemetry.json"):
		missing = append(missing, "Regenerate bottleneck/execution/telemetry.json as valid telemetry JSON.")
	case strings.Contains(message, "low adoption"):
		missing = append(missing, "Provide adoption evidence at or above execution.min_adoption.")
	case strings.Contains(message, "threshold"):
		missing = append(missing, fmt.Sprintf("Provide %s evidence that satisfies the effective thresholds.", strings.ToLower(validation.Capability)))
	case strings.Contains(message, "violations detected"):
		missing = append(missing, "Provide security evidence with zero guardrail violations.")
	}

	if len(evidence) == 0 {
		missing = append(missing, meta.missingEvidence)
	}

	return uniqueStrings(missing)
}

func missingEvidenceFromDetail(detail string) string {
	switch {
	case strings.Contains(detail, "still contains placeholder content"):
		return strings.Replace(detail, "still contains placeholder content", "needs real evidence", 1)
	case strings.Contains(detail, "is too thin to validate"):
		return strings.Replace(detail, "is too thin to validate", "needs a meaningful sentence or list item", 1)
	default:
		return ""
	}
}

func reasonFor(validation models.ValidationResult) string {
	if validation.Message != "" {
		return validation.Message
	}

	switch displayStatus(validation.Status) {
	case StatusPass:
		return "validation passed"
	case StatusWarn:
		return "validation warning reported"
	case StatusFail:
		return "validation failure reported"
	default:
		return "validation status unavailable"
	}
}

func recommendedActionFor(validation models.ValidationResult, meta capabilityMetadata) string {
	switch displayStatus(validation.Status) {
	case StatusPass:
		return meta.passAction
	case StatusUnknown:
		return "Run validation before using this scorecard for release decisions."
	default:
		return diagnosis.RecommendedAction(validation)
	}
}

func releaseRecommendationFor(capabilities []CapabilityScorecard, systemStatus string) string {
	if systemStatus == StatusUnknown || len(capabilities) == 0 {
		return RecommendationUnknown
	}

	hasWarning := false
	for _, capability := range capabilities {
		switch capability.Status {
		case StatusFail:
			return RecommendationBlock
		case StatusUnknown:
			return RecommendationUnknown
		case StatusWarn:
			hasWarning = true
		}
	}

	if hasWarning || systemStatus == StatusWarn {
		return RecommendationConditional
	}

	return RecommendationProceed
}

func bottomLine(card Scorecard) string {
	switch card.ReleaseRecommendation {
	case RecommendationBlock:
		return fmt.Sprintf(
			"The system is not valid for %s. Primary ownership starts with %s.",
			card.Environment,
			metadataFor(card.PrimaryBottleneck).owner,
		)
	case RecommendationConditional:
		return fmt.Sprintf(
			"The system has warnings for %s. Primary ownership starts with %s.",
			card.Environment,
			metadataFor(card.PrimaryBottleneck).owner,
		)
	case RecommendationProceed:
		return fmt.Sprintf(
			"The system is valid for %s. Continue monitoring all capability signals.",
			card.Environment,
		)
	default:
		return fmt.Sprintf(
			"The release posture for %s is unknown because required scorecard evidence is unavailable.",
			card.Environment,
		)
	}
}

func displayStatus(status string) string {
	switch strings.ToUpper(status) {
	case models.StatusPass:
		return StatusPass
	case models.StatusWarning, StatusWarn:
		return StatusWarn
	case models.StatusFail:
		return StatusFail
	case StatusUnknown:
		return StatusUnknown
	default:
		return StatusUnknown
	}
}

func metadataFor(capability string) capabilityMetadata {
	if meta, ok := metadataByCapability[capability]; ok {
		return meta
	}

	return capabilityMetadata{
		owner:             "Execution Engineer",
		bottleneck:        "Delivery friction",
		passingArtifact:   strings.ToLower(capability),
		missingEvidence:   fmt.Sprintf("Add validation evidence for %s.", capability),
		recommendedAction: "Inspect the underlying artifact and validation output for this capability.",
		passAction:        "Keep the artifact and observed evidence current.",
	}
}

func normalizeView(viewValues ...string) (string, error) {
	view := ViewEngineering
	if len(viewValues) > 0 && strings.TrimSpace(viewValues[0]) != "" {
		view = strings.ToLower(viewValues[0])
	}

	switch view {
	case ViewExecutive, ViewEngineering, ViewGovernance:
		return view, nil
	default:
		return "", fmt.Errorf("unsupported view %q (supported: executive, engineering, governance)", view)
	}
}

func capabilityByName(card Scorecard, capabilityName string) (CapabilityScorecard, bool) {
	for _, capability := range card.Capabilities {
		if capability.Capability == capabilityName {
			return capability, true
		}
	}
	return CapabilityScorecard{}, false
}

func governanceMissingEvidence(card Scorecard) []string {
	var missing []string
	for _, capability := range card.Capabilities {
		if capability.Status == StatusFail || capability.Status == StatusWarn || len(capability.MissingEvidence) > 0 {
			missing = append(missing, capability.MissingEvidence...)
		}
	}
	missing = append(missing, "Governance evidence not assessed: no governance artifact exists yet.")
	return uniqueStrings(missing)
}

func requiredSectionsMissingEvidence(capability string) string {
	switch capability {
	case "Behavior":
		return "Add ## Expected Behavior and ## Unacceptable Behavior sections to bottleneck/behavior/behavior-spec.md."
	case "Intent":
		return "Add ## Outcomes, ## Constraints, and ## Success Criteria sections to bottleneck/intent/intent.md."
	default:
		return fmt.Sprintf("Add the required Markdown sections for %s.", capability)
	}
}

func statusSummaryLines(card Scorecard) []string {
	rows := statusSummaryRows(card)
	lines := make([]string, 0, len(rows))
	for _, row := range rows {
		lines = append(lines, fmt.Sprintf("%s: %d", row.status, row.count))
	}
	return lines
}

type statusSummaryRow struct {
	status string
	count  int
}

func statusSummaryRows(card Scorecard) []statusSummaryRow {
	counts := map[string]int{
		StatusPass:    0,
		StatusWarn:    0,
		StatusFail:    0,
		StatusUnknown: 0,
	}
	for _, capability := range card.Capabilities {
		counts[capability.Status]++
	}

	return []statusSummaryRow{
		{status: StatusPass, count: counts[StatusPass]},
		{status: StatusWarn, count: counts[StatusWarn]},
		{status: StatusFail, count: counts[StatusFail]},
		{status: StatusUnknown, count: counts[StatusUnknown]},
	}
}

func appendBulletLines(lines []string, items []string, prefix string, emptyText string) []string {
	if len(items) == 0 {
		return append(lines, prefix+"- "+emptyText)
	}

	for _, item := range items {
		lines = append(lines, prefix+"- "+item)
	}
	return lines
}

func appendScoreImpactLines(lines []string, impacts []models.ScoreImpact, prefix string) []string {
	if len(impacts) == 0 {
		return append(lines, prefix+"- None.")
	}
	for _, impact := range impacts {
		lines = append(lines, fmt.Sprintf("%s- %s (%+d)", prefix, impact.Reason, impact.Delta))
	}
	return lines
}

func appendMarkdownBullets(lines []string, items []string, prefix string, emptyText string) []string {
	if len(items) == 0 {
		return append(lines, prefix+"- "+markdownText(emptyText))
	}

	for _, item := range items {
		lines = append(lines, prefix+"- "+markdownText(item))
	}
	return lines
}

func appendMarkdownScoreImpacts(lines []string, impacts []models.ScoreImpact, prefix string) []string {
	if len(impacts) == 0 {
		return append(lines, prefix+"- None.")
	}
	for _, impact := range impacts {
		lines = append(lines, fmt.Sprintf("%s- %s (%+d)", prefix, markdownText(impact.Reason), impact.Delta))
	}
	return lines
}

func joinMarkdownList(items []string, emptyText string) string {
	if len(items) == 0 {
		return emptyText
	}
	return strings.Join(items, "<br>")
}

func markdownCell(value string) string {
	value = strings.ReplaceAll(value, "|", `\|`)
	value = strings.ReplaceAll(value, "\n", "<br>")
	return value
}

func markdownText(value string) string {
	return strings.ReplaceAll(value, "|", `\|`)
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]bool, len(values))
	unique := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		unique = append(unique, value)
	}
	return unique
}

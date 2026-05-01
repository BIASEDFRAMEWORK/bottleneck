package diagnosis

import (
	"encoding/json"
	"fmt"
	"strings"

	"bottleneck/internal/githubannotations"
	"bottleneck/internal/models"
)

const (
	DiagnoseSchemaVersion = "diagnose.v1"
	FormatText            = "text"
	FormatJSON            = "json"
	FormatMarkdown        = "markdown"
	FormatGitHub          = "github"
)

type Report struct {
	SchemaVersion        string          `json:"schema_version"`
	Environment          string          `json:"environment"`
	SystemStatus         string          `json:"system_status"`
	PrimaryBottleneck    string          `json:"primary_bottleneck"`
	TiedBottlenecks      []string        `json:"tied_bottlenecks"`
	Rule                 string          `json:"rule,omitempty"`
	Reason               string          `json:"reason"`
	Impact               string          `json:"impact"`
	NextAction           string          `json:"next_action"`
	InspectCommand       string          `json:"inspect_command"`
	RelevantEvidenceIDs  []string        `json:"relevant_evidence_ids"`
	SupportingIssues     []string        `json:"supporting_issues"`
	WhyItMatters         string          `json:"why_it_matters"`
	ContributingFindings []string        `json:"contributing_findings"`
	RecommendedAction    string          `json:"recommended_action"`
	Confidence           string          `json:"confidence"`
	ConfidenceReason     string          `json:"confidence_reason"`
	CategoryScores       []CategoryScore `json:"category_scores"`
}

func BuildReport(result models.EngineResult) Report {
	diagnosis := Analyze(result)
	tied := append([]string{}, diagnosis.TiedBottlenecks...)
	findings := append([]string{}, diagnosis.ContributingFindings...)
	if tied == nil {
		tied = []string{}
	}
	if findings == nil {
		findings = []string{}
	}
	relevantIDs := append([]string{}, diagnosis.RelevantEvidenceIDs...)
	if relevantIDs == nil {
		relevantIDs = []string{}
	}
	supportingIssues := append([]string{}, diagnosis.SupportingIssues...)
	if supportingIssues == nil {
		supportingIssues = []string{}
	}
	scores := append([]CategoryScore{}, diagnosis.CategoryScores...)
	if scores == nil {
		scores = []CategoryScore{}
	}
	return Report{
		SchemaVersion:        DiagnoseSchemaVersion,
		Environment:          result.Environment,
		SystemStatus:         result.SystemStatus,
		PrimaryBottleneck:    diagnosis.PrimaryBottleneck,
		TiedBottlenecks:      tied,
		Rule:                 diagnosis.Rule,
		Reason:               diagnosis.Reason,
		Impact:               diagnosis.Impact,
		NextAction:           diagnosis.NextAction,
		InspectCommand:       diagnosis.InspectCommand,
		RelevantEvidenceIDs:  relevantIDs,
		SupportingIssues:     supportingIssues,
		WhyItMatters:         diagnosis.WhyItMatters,
		ContributingFindings: findings,
		RecommendedAction:    diagnosis.RecommendedAction,
		Confidence:           diagnosis.Confidence,
		ConfidenceReason:     diagnosis.ConfidenceReason,
		CategoryScores:       scores,
	}
}

func Render(result models.EngineResult, format string) (string, error) {
	report := BuildReport(result)
	switch strings.ToLower(format) {
	case "", FormatText:
		return renderText(report), nil
	case FormatJSON:
		return renderJSON(report)
	case FormatMarkdown:
		return renderMarkdown(report), nil
	case FormatGitHub:
		return githubannotations.Render(result.Results), nil
	default:
		return "", fmt.Errorf("unsupported format %q (supported: text, json, markdown, github)", format)
	}
}

func renderText(report Report) string {
	var lines []string
	lines = append(lines, fmt.Sprintf("Primary Bottleneck: %s", report.PrimaryBottleneck))
	if len(report.TiedBottlenecks) > 0 {
		lines = append(lines, fmt.Sprintf("Tied Bottlenecks: %s", strings.Join(report.TiedBottlenecks, ", ")))
	}
	lines = append(lines,
		fmt.Sprintf("Reason: %s", report.Reason),
		fmt.Sprintf("Impact: %s", report.Impact),
		fmt.Sprintf("Next Action: %s", effectiveNextAction(report)),
		fmt.Sprintf("Inspect: %s", report.InspectCommand),
	)
	if len(report.RelevantEvidenceIDs) > 0 {
		lines = append(lines, fmt.Sprintf("Relevant Evidence: %s", strings.Join(report.RelevantEvidenceIDs, ", ")))
	}

	lines = append(lines, "", "Supporting Issues:")
	appendNumberedLines(&lines, report.SupportingIssues, "No additional supporting issues were reported.")
	lines = append(lines, "", "Contributing Findings:")
	appendNumberedLines(&lines, report.ContributingFindings, "No specific findings were reported.")
	lines = append(lines, "", fmt.Sprintf("Diagnosis Confidence: %s", report.Confidence), "Confidence Reason:", report.ConfidenceReason)
	return strings.Join(lines, "\n")
}

func renderJSON(report Report) (string, error) {
	content, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func renderMarkdown(report Report) string {
	var lines []string
	lines = append(lines,
		"## Bottleneck Diagnosis",
		"",
		"| Field | Value |",
		"| --- | --- |",
		fmt.Sprintf("| Environment | %s |", markdownCell(report.Environment)),
		fmt.Sprintf("| System Status | %s |", markdownCell(report.SystemStatus)),
		fmt.Sprintf("| Primary Bottleneck | %s |", markdownCell(report.PrimaryBottleneck)),
		fmt.Sprintf("| Confidence | %s |", markdownCell(report.Confidence)),
	)
	if len(report.TiedBottlenecks) > 0 {
		lines = append(lines, fmt.Sprintf("| Tied Bottlenecks | %s |", markdownCell(strings.Join(report.TiedBottlenecks, ", "))))
	}
	lines = append(lines,
		"",
		"### Reason",
		"",
		markdownText(report.Reason),
		"",
		"### Impact",
		"",
		markdownText(report.Impact),
		"",
		"### Recommended Next Action",
		"",
		markdownText(effectiveNextAction(report)),
		"",
		"### Inspect",
		"",
		fmt.Sprintf("`%s`", markdownText(report.InspectCommand)),
		"",
		"### Relevant Evidence",
		"",
	)
	appendMarkdownBullets(&lines, report.RelevantEvidenceIDs, "None.")
	lines = append(lines,
		"",
		"### Why It Matters",
		"",
		markdownText(report.WhyItMatters),
		"",
		"### Category Scores",
		"",
		"| Category | Score | Status |",
		"| --- | ---: | --- |",
	)
	if len(report.CategoryScores) == 0 {
		lines = append(lines, "| None | 0 | UNKNOWN |")
	} else {
		for _, score := range report.CategoryScores {
			lines = append(lines, fmt.Sprintf("| %s | %d | %s |",
				markdownCell(score.Category),
				score.Score,
				markdownCell(score.Status),
			))
		}
	}
	lines = append(lines,
		"",
		"### Top Findings",
		"",
	)
	if len(report.ContributingFindings) == 0 {
		lines = append(lines, "1. No specific findings were reported.")
	} else {
		for index, finding := range report.ContributingFindings {
			lines = append(lines, fmt.Sprintf("%d. %s", index+1, markdownText(finding)))
		}
	}
	lines = append(lines,
		"",
		"### Supporting Issues",
		"",
	)
	appendMarkdownBullets(&lines, report.SupportingIssues, "None.")
	lines = append(lines,
		"",
		"### Confidence Reason",
		"",
		markdownText(report.ConfidenceReason),
	)
	return strings.Join(lines, "\n")
}

func effectiveNextAction(report Report) string {
	if strings.TrimSpace(report.NextAction) != "" {
		return report.NextAction
	}
	return report.RecommendedAction
}

func appendNumberedLines(lines *[]string, values []string, empty string) {
	if len(values) == 0 {
		*lines = append(*lines, "1. "+empty)
		return
	}
	for index, value := range values {
		*lines = append(*lines, fmt.Sprintf("%d. %s", index+1, value))
	}
}

func appendMarkdownBullets(lines *[]string, values []string, empty string) {
	if len(values) == 0 {
		*lines = append(*lines, "- "+empty)
		return
	}
	for _, value := range values {
		*lines = append(*lines, "- "+markdownText(value))
	}
}

func markdownCell(value string) string {
	value = strings.ReplaceAll(value, "|", `\|`)
	value = strings.ReplaceAll(value, "\n", "<br>")
	return value
}

func markdownText(value string) string {
	return strings.ReplaceAll(value, "|", `\|`)
}

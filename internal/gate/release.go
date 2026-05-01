package gate

import (
	"encoding/json"
	"fmt"
	"strings"

	"bottleneck/internal/config"
	"bottleneck/internal/diagnosis"
	"bottleneck/internal/models"
)

const (
	ReleaseGateName = "release"
	StatusPass      = "PASS"
	StatusFail      = "FAIL"
)

type Result struct {
	Name              string   `json:"name"`
	Status            string   `json:"status"`
	PrimaryBottleneck string   `json:"primary_bottleneck"`
	PrimaryScore      int      `json:"primary_score"`
	RequiredScore     int      `json:"required_score"`
	Reasons           []string `json:"reasons"`
	NotAssessed       []string `json:"not_assessed,omitempty"`
	RecommendedAction string   `json:"recommended_action"`
}

func EvaluateRelease(result models.EngineResult, settings config.ReleaseGateConfig) Result {
	analysis := diagnosis.Analyze(result)
	resultsByCategory := resultsByCapability(result.Results)

	report := Result{
		Name:              ReleaseGateName,
		Status:            StatusPass,
		PrimaryBottleneck: analysis.PrimaryBottleneck,
		PrimaryScore:      primaryScore(analysis),
		RequiredScore:     settings.MinPrimaryScore,
		Reasons:           []string{},
		RecommendedAction: analysis.RecommendedAction,
	}

	if report.PrimaryScore < settings.MinPrimaryScore {
		report.Reasons = append(report.Reasons, "Primary bottleneck score is below release threshold.")
	}

	for _, category := range settings.RequiredCategories {
		if strings.TrimSpace(category) == "" {
			continue
		}
		if _, ok := resultsByCategory[category]; !ok {
			report.Reasons = append(report.Reasons, fmt.Sprintf("Required category %s is missing.", category))
		}
	}

	if settings.RequireTraceability {
		traceability, ok := resultsByCategory["Traceability"]
		switch {
		case !ok:
			report.Reasons = append(report.Reasons, "Traceability result is missing.")
		case traceability.Status == models.StatusFail:
			report.Reasons = append(report.Reasons, "Traceability is broken.")
		}
	}

	if security, ok := resultsByCategory["Security"]; ok && security.Status == models.StatusFail {
		report.Reasons = append(report.Reasons, "Security evidence fails release policy.")
	}

	governance, governanceAssessed := resultsByCategory["Governance"]
	if governanceAssessed && governance.Status == models.StatusFail {
		report.Reasons = append(report.Reasons, "Governance evidence fails release policy.")
	}
	if settings.RequireGovernance && !governanceAssessed {
		report.Reasons = append(report.Reasons, "Governance evidence is required but not implemented or not assessed.")
	}
	if !settings.RequireGovernance && !governanceAssessed {
		report.NotAssessed = append(report.NotAssessed, "Governance evidence is not assessed because require_governance is false.")
	}

	if len(report.Reasons) > 0 {
		report.Status = StatusFail
	}

	return report
}

func Render(result Result, format string) (string, error) {
	switch strings.ToLower(format) {
	case "", "text":
		return RenderText(result), nil
	case "json":
		content, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return "", err
		}
		return string(content), nil
	case "markdown":
		return RenderMarkdown(result), nil
	case "github":
		return RenderGitHub(result), nil
	default:
		return "", fmt.Errorf("unsupported gate format %q (supported: text, json, markdown, github)", format)
	}
}

func RenderText(result Result) string {
	var lines []string
	lines = append(lines,
		fmt.Sprintf("Release Gate: %s", result.Status),
		"",
		fmt.Sprintf("Primary Bottleneck: %s", result.PrimaryBottleneck),
		fmt.Sprintf("Primary Score: %d", result.PrimaryScore),
		fmt.Sprintf("Required Score: %d", result.RequiredScore),
		"",
		"Reasons:",
	)
	if len(result.Reasons) == 0 {
		lines = append(lines, "1. No release gate failures detected.")
	} else {
		for index, reason := range result.Reasons {
			lines = append(lines, fmt.Sprintf("%d. %s", index+1, reason))
		}
	}
	if len(result.NotAssessed) > 0 {
		lines = append(lines, "", "Not assessed:")
		for index, reason := range result.NotAssessed {
			lines = append(lines, fmt.Sprintf("%d. %s", index+1, reason))
		}
	}
	lines = append(lines,
		"",
		"Recommended next action:",
		result.RecommendedAction,
	)
	return strings.Join(lines, "\n")
}

func RenderMarkdown(result Result) string {
	var lines []string
	lines = append(lines,
		"## Bottleneck Release Gate",
		"",
		"| Field | Value |",
		"| --- | --- |",
		fmt.Sprintf("| Gate | %s |", markdownCell(result.Name)),
		fmt.Sprintf("| Result | %s |", markdownCell(result.Status)),
		fmt.Sprintf("| Primary Bottleneck | %s |", markdownCell(result.PrimaryBottleneck)),
		fmt.Sprintf("| Primary Score | %d |", result.PrimaryScore),
		fmt.Sprintf("| Required Score | %d |", result.RequiredScore),
		"",
		"### Gate Reasons",
		"",
	)
	if len(result.Reasons) == 0 {
		lines = append(lines, "1. No release gate failures detected.")
	} else {
		for index, reason := range result.Reasons {
			lines = append(lines, fmt.Sprintf("%d. %s", index+1, markdownText(reason)))
		}
	}
	if len(result.NotAssessed) > 0 {
		lines = append(lines, "", "### Not Assessed", "")
		for index, reason := range result.NotAssessed {
			lines = append(lines, fmt.Sprintf("%d. %s", index+1, markdownText(reason)))
		}
	}
	lines = append(lines,
		"",
		"### Recommended Next Action",
		"",
		markdownText(result.RecommendedAction),
	)
	return strings.Join(lines, "\n")
}

func RenderGitHub(result Result) string {
	if result.Status != StatusFail {
		return "::notice::Bottleneck release gate passed"
	}
	var lines []string
	lines = append(lines, fmt.Sprintf("::error::%s", escapeWorkflowMessage("Bottleneck release gate failed")))
	for _, reason := range result.Reasons {
		lines = append(lines, fmt.Sprintf("::error::%s", escapeWorkflowMessage(reason)))
	}
	return strings.Join(lines, "\n")
}

func resultsByCapability(results []models.ValidationResult) map[string]models.ValidationResult {
	byCategory := map[string]models.ValidationResult{}
	for _, result := range results {
		byCategory[result.Capability] = result
	}
	return byCategory
}

func primaryScore(analysis diagnosis.Diagnosis) int {
	if analysis.PrimaryBottleneck != diagnosis.HealthyPrimaryBottleneck {
		return diagnosis.ScoreFor(analysis.PrimaryBottleneck, analysis.CategoryScores)
	}
	if len(analysis.CategoryScores) == 0 {
		return 0
	}
	lowest := 101
	for _, score := range analysis.CategoryScores {
		if score.Score < lowest {
			lowest = score.Score
		}
	}
	return lowest
}

func markdownCell(value string) string {
	value = strings.ReplaceAll(value, "|", `\|`)
	value = strings.ReplaceAll(value, "\n", "<br>")
	return value
}

func markdownText(value string) string {
	return strings.ReplaceAll(value, "|", `\|`)
}

func escapeWorkflowMessage(value string) string {
	value = strings.ReplaceAll(value, "%", "%25")
	value = strings.ReplaceAll(value, "\r", "%0D")
	value = strings.ReplaceAll(value, "\n", "%0A")
	return value
}

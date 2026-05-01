package trends

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Render(analysis Analysis, format string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "", FormatText:
		return RenderText(analysis), nil
	case FormatMarkdown:
		return RenderMarkdown(analysis), nil
	case FormatJSON:
		return RenderJSON(analysis)
	default:
		return "", fmt.Errorf("unsupported format %q (supported: text, markdown, json)", format)
	}
}

func RenderText(analysis Analysis) string {
	lines := []string{
		"Bottleneck Trends",
		"",
		fmt.Sprintf("Environment: %s", analysis.Environment),
		fmt.Sprintf("Snapshots analyzed: %d", analysis.SnapshotCount),
		fmt.Sprintf("Window: latest %d snapshots", analysis.Window),
		"",
		fmt.Sprintf("Overall direction: %s", displayDirection(analysis.OverallDirection)),
		fmt.Sprintf("Current status: %s", emptyFallback(analysis.CurrentStatus, "Unknown")),
		fmt.Sprintf("Current primary bottleneck: %s", emptyFallback(analysis.CurrentPrimaryBottleneck, "None")),
		"",
		"Category trends:",
	}
	if len(analysis.CategoryTrends) == 0 {
		lines = append(lines, "- Not enough category history yet.")
	} else {
		for _, trend := range analysis.CategoryTrends {
			lines = append(lines, fmt.Sprintf("- %-10s %s  %s", trend.Category+":", trendValuesText(trend), trendDirectionText(trend)))
		}
	}

	lines = append(lines,
		"",
		"Persistent bottleneck:",
		analysis.PersistentBottleneck.Summary,
		"",
		"Leadership summary:",
		analysis.LeadershipSummary,
	)
	if analysis.OverallDirection == DirectionInsufficientHistory {
		lines = append(lines, "", "Next:", "Run `bottleneck snapshot` after each delivery cycle to build trend history.")
	}
	if len(analysis.Warnings) > 0 {
		lines = append(lines, "", "Warnings:")
		for _, warning := range analysis.Warnings {
			lines = append(lines, "- "+warning)
		}
	}
	return strings.Join(lines, "\n")
}

func RenderMarkdown(analysis Analysis) string {
	lines := []string{
		"# Bottleneck Trend Summary",
		"",
		"## Executive Summary",
		analysis.LeadershipSummary,
		"",
		"## Snapshot Window",
		fmt.Sprintf("- Environment: `%s`", analysis.Environment),
		fmt.Sprintf("- Snapshots analyzed: `%d`", analysis.SnapshotCount),
		fmt.Sprintf("- Window: latest `%d` snapshots", analysis.Window),
		fmt.Sprintf("- Overall direction: `%s`", analysis.OverallDirection),
		fmt.Sprintf("- Current status: `%s`", emptyFallback(analysis.CurrentStatus, "UNKNOWN")),
		fmt.Sprintf("- Current primary bottleneck: `%s`", emptyFallback(analysis.CurrentPrimaryBottleneck, "None")),
		"",
		"## Category Trends",
		"",
		"| Category | Values | Direction | Delta | Summary |",
		"| --- | --- | --- | --- | --- |",
	}
	if len(analysis.CategoryTrends) == 0 {
		lines = append(lines, "| None | Not enough history | insufficient_history | 0 | Create more snapshots. |")
	} else {
		for _, trend := range analysis.CategoryTrends {
			lines = append(lines, fmt.Sprintf("| %s | %s | %s | %s | %s |",
				trend.Category,
				trendValuesText(trend),
				trend.Direction,
				formatDelta(trend.Delta),
				escapeMarkdownTable(trend.Summary),
			))
		}
	}
	lines = append(lines,
		"",
		"## Persistent Bottleneck",
		analysis.PersistentBottleneck.Summary,
		"",
		"## Leadership Interpretation",
		analysis.LeadershipSummary,
		"",
		"## Recommended Follow-Up",
		recommendedFollowUp(analysis),
	)
	if len(analysis.Warnings) > 0 {
		lines = append(lines, "", "## Warnings")
		for _, warning := range analysis.Warnings {
			lines = append(lines, "- "+warning)
		}
	}
	return strings.Join(lines, "\n")
}

func RenderJSON(analysis Analysis) (string, error) {
	content, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func WriteReport(path string, content string) error {
	if strings.TrimSpace(path) == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func trendValuesText(trend CategoryTrend) string {
	if len(trend.Values) == 0 {
		return "no data"
	}
	values := make([]string, 0, len(trend.Values))
	for _, value := range trend.Values {
		values = append(values, formatScore(value))
	}
	return strings.Join(values, " -> ")
}

func trendDirectionText(trend CategoryTrend) string {
	text := displayDirection(trend.Direction)
	if trend.Direction == DirectionImproving && trend.CurrentValue < 60 {
		return text + ", but still weak"
	}
	return text
}

func displayDirection(direction Direction) string {
	switch direction {
	case DirectionImproving:
		return "Improving"
	case DirectionDeclining:
		return "Declining"
	case DirectionStable:
		return "Stable"
	case DirectionRecovered:
		return "Recovered"
	case DirectionRegressed:
		return "Regressed"
	case DirectionInsufficientHistory:
		return "Insufficient history"
	default:
		return "Unknown"
	}
}

func recommendedFollowUp(analysis Analysis) string {
	if analysis.OverallDirection == DirectionInsufficientHistory {
		return "Run `bottleneck snapshot` after each delivery cycle until at least two snapshots exist for this environment."
	}
	if analysis.PersistentBottleneck.Category != "" {
		return fmt.Sprintf("Review `%s` evidence and decide whether it needs ownership, staffing, or process investment.", analysis.PersistentBottleneck.Category)
	}
	return "Keep creating snapshots and review categories that move toward warning or failing status."
}

func emptyFallback(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func escapeMarkdownTable(value string) string {
	return strings.ReplaceAll(value, "|", "\\|")
}

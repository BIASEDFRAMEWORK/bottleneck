package report

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Render(report Report, format string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "", FormatMarkdown:
		return RenderMarkdown(report), nil
	case FormatJSON:
		return RenderJSON(report)
	default:
		return "", fmt.Errorf("unsupported format %q (supported: markdown, json)", format)
	}
}

func RenderMarkdown(report Report) string {
	lines := []string{
		"# SDLC Evidence Report",
		"",
		"## Executive Summary",
		executiveSummary(report),
		"",
		"## Current Delivery-System Status",
		fmt.Sprintf("- Environment: `%s`", report.Environment),
		fmt.Sprintf("- Current status: `%s`", report.CurrentStatus),
		fmt.Sprintf("- Release recommendation: `%s`", report.ReleaseRecommendation),
		fmt.Sprintf("- Generated at: `%s`", report.GeneratedAt),
		"",
		"## Primary Bottleneck",
		primaryBottleneckText(report),
		"",
		"## Category Scorecard",
		"",
		"| Category | Status | Score | Owner | Recommended Action |",
		"| --- | --- | ---: | --- | --- |",
	}
	for _, capability := range report.Scorecard.Capabilities {
		lines = append(lines, fmt.Sprintf("| %s | %s | %d | %s | %s |",
			capability.Capability,
			capability.Status,
			capability.Score,
			capability.Owner,
			escapeTable(capability.RecommendedAction),
		))
	}
	lines = append(lines,
		"",
		"## Trend Summary",
		trendSummaryText(report),
		"",
		"## Evidence Found",
	)
	lines = appendBulletLines(lines, report.EvidenceFound)
	lines = append(lines, "", "## Evidence Missing")
	lines = appendBulletLines(lines, report.EvidenceMissing)
	lines = append(lines, "", "## Risk to Delivery")
	lines = appendBulletLines(lines, report.Risks)
	lines = append(lines, "", "## Recommended Actions")
	lines = appendNumberedLines(lines, report.RecommendedActions)
	lines = append(lines, "", "## Suggested Owners")
	lines = appendBulletLines(lines, report.SuggestedOwners)
	lines = append(lines, "", "## Suggested Automation")
	lines = appendBulletLines(lines, report.SuggestedAutomation)
	lines = append(lines, "", "## Leadership Decision Needed", report.LeadershipDecision)
	lines = append(lines, "", "## Appendix: Snapshot Metadata")
	lines = append(lines,
		fmt.Sprintf("- Snapshot count: `%d`", report.SnapshotMetadata.SnapshotCount),
		fmt.Sprintf("- Window: `%d`", report.SnapshotMetadata.Window),
		fmt.Sprintf("- Environment: `%s`", report.SnapshotMetadata.Environment),
	)
	if len(report.SnapshotMetadata.Warnings) > 0 {
		lines = append(lines, "- Warnings:")
		for _, warning := range report.SnapshotMetadata.Warnings {
			lines = append(lines, "  - "+warning)
		}
	}
	return strings.Join(lines, "\n")
}

func RenderJSON(report Report) (string, error) {
	content, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func WriteFile(path string, content string) error {
	if strings.TrimSpace(path) == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func executiveSummary(report Report) string {
	return fmt.Sprintf(
		"The current evidence suggests `%s` release readiness with `%s` as the primary constraint. Leadership should use this report to decide where evidence, ownership, or release controls need attention.",
		report.CurrentStatus,
		report.PrimaryBottleneck,
	)
}

func primaryBottleneckText(report Report) string {
	if strings.EqualFold(report.PrimaryBottleneck, "None") || strings.TrimSpace(report.PrimaryBottleneck) == "" {
		return "No primary bottleneck is currently identified. Continue monitoring evidence and maintaining release controls."
	}
	return fmt.Sprintf("The primary constraint appears to be `%s`.", report.PrimaryBottleneck)
}

func trendSummaryText(report Report) string {
	if report.TrendSummary == nil || report.TrendSummary.SnapshotCount < 2 {
		return "Trend history is insufficient. Create snapshots over multiple delivery cycles with `bottleneck snapshot` to establish direction."
	}
	summary := fmt.Sprintf(
		"Overall direction is `%s` across the latest `%d` snapshots. %s",
		report.TrendSummary.OverallDirection,
		report.TrendSummary.SnapshotCount,
		report.TrendSummary.LeadershipSummary,
	)
	if report.TrendSummary.PersistentBottleneck.Summary != "" {
		summary += " " + report.TrendSummary.PersistentBottleneck.Summary
	}
	return summary
}

func appendBulletLines(lines []string, values []string) []string {
	if len(values) == 0 {
		return append(lines, "- None.")
	}
	for _, value := range values {
		lines = append(lines, "- "+value)
	}
	return lines
}

func appendNumberedLines(lines []string, values []string) []string {
	if len(values) == 0 {
		return append(lines, "1. Continue monitoring evidence and maintain current release controls.")
	}
	for index, value := range values {
		lines = append(lines, fmt.Sprintf("%d. %s", index+1, value))
	}
	return lines
}

func escapeTable(value string) string {
	return strings.ReplaceAll(value, "|", "\\|")
}

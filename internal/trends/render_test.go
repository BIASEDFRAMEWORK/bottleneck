package trends

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTrendsRendersText(t *testing.T) {
	output := RenderText(sampleTrendAnalysis())
	for _, expected := range []string{
		"Bottleneck Trends",
		"Environment: default",
		"Snapshots analyzed: 3",
		"Overall direction: Improving",
		"Category trends:",
		"- Intent:",
		"Persistent bottleneck:",
		"Leadership summary:",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected text output to contain %q:\n%s", expected, output)
		}
	}
}

func TestTrendsRendersMarkdown(t *testing.T) {
	output := RenderMarkdown(sampleTrendAnalysis())
	for _, expected := range []string{
		"# Bottleneck Trend Summary",
		"## Executive Summary",
		"## Snapshot Window",
		"## Category Trends",
		"## Persistent Bottleneck",
		"## Leadership Interpretation",
		"## Recommended Follow-Up",
		"| Category | Values | Direction | Delta | Summary |",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected markdown output to contain %q:\n%s", expected, output)
		}
	}
}

func TestTrendsRendersJSON(t *testing.T) {
	output, err := RenderJSON(sampleTrendAnalysis())
	if err != nil {
		t.Fatalf("render JSON: %v", err)
	}
	var decoded Analysis
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("trend JSON did not parse: %v\n%s", err, output)
	}
	if decoded.Environment != "default" || decoded.OverallDirection != DirectionImproving {
		t.Fatalf("unexpected decoded analysis: %#v", decoded)
	}
}

func sampleTrendAnalysis() Analysis {
	return Analysis{
		Environment:              "default",
		SnapshotCount:            3,
		Window:                   6,
		OverallDirection:         DirectionImproving,
		CurrentStatus:            "WARN",
		CurrentPrimaryBottleneck: "Assurance",
		CategoryTrends: []CategoryTrend{{
			Category:      "Intent",
			Values:        []float64{70, 80, 90},
			Statuses:      []string{"WARN", "WARN", "PASS"},
			Direction:     DirectionRecovered,
			Delta:         20,
			CurrentValue:  90,
			PreviousValue: 80,
			Summary:       "Intent changed by +20 points across the selected window.",
		}},
		PersistentBottleneck: PersistentBottleneck{
			Category: "Assurance",
			Count:    2,
			Total:    3,
			Summary:  "Assurance appeared as the primary bottleneck in 2 of 3 snapshots.",
		},
		LeadershipSummary: "The team is improving overall, but Assurance remains the most persistent delivery constraint.",
	}
}

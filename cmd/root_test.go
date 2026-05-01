package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCommandProductLanguageMatchesRoadmap(t *testing.T) {
	description := rootCmd.Short + "\n" + rootCmd.Long
	expected := []string{
		"CLI",
		"measuring SDLC maturity",
		"local engineering evidence",
		"blocking release confidence",
		"what evidence proves it",
		"Start here:",
		"bottleneck init --template saas",
		"bottleneck assess",
		"bottleneck trace BEHAVIOR-003",
		"Common commands:",
		"assess      Show maturity, AI readiness, release friction, and next action.",
		"discover    Find local test, security, telemetry, design, and workflow evidence.",
		"evidence sync Discover and ingest supported local evidence automatically.",
		"explain-score Explain the maturity score with evidence provenance.",
		"Advanced commands:",
		"check       Check evidence files for missing, thin, or placeholder content.",
		"validate    Check evidence files for missing, thin, or placeholder content.",
		"scorecard   Show release readiness, primary bottleneck, and next action.",
		"diagnose    Explain what is blocking delivery and what to inspect next.",
		"trace       Follow one intent, behavior, or evidence ID end-to-end.",
		"ingest      Convert test, security, and telemetry reports into Bottleneck evidence.",
		"snapshot    Write scorecard snapshots for local trend history.",
		"seed-history Create demo scorecard history for local trends.",
		"trends      Analyze scorecard trends from local snapshots.",
		"report      Generate a leadership-ready SDLC evidence report.",
		"explain     Show how evidence affected category scores.",
		"main terminal surface",
	}

	for _, substring := range expected {
		if !strings.Contains(description, substring) {
			t.Fatalf("expected root command description to contain %q:\n%s", substring, description)
		}
	}

	if strings.Contains(strings.ToLower(description), "framework validation only") {
		t.Fatalf("root command description should not frame Bottleneck as framework validation only:\n%s", description)
	}
}

func TestRootHelpShowsDayOnePathAndCommonCommands(t *testing.T) {
	var buffer bytes.Buffer
	rootCmd.SetOut(&buffer)
	defer rootCmd.SetOut(nil)

	if err := rootCmd.Help(); err != nil {
		t.Fatalf("render root help: %v", err)
	}
	help := buffer.String()

	expected := []string{
		"Start here:\n  bottleneck init --template saas\n  bottleneck assess\n  bottleneck trace BEHAVIOR-003",
		"Common commands:",
		"assess      Show maturity, AI readiness, release friction, and next action.",
		"discover    Find local test, security, telemetry, design, and workflow evidence.",
		"evidence sync Discover and ingest supported local evidence automatically.",
		"explain-score Explain the maturity score with evidence provenance.",
		"Advanced commands:",
		"check       Check evidence files for missing, thin, or placeholder content.",
		"validate    Check evidence files for missing, thin, or placeholder content.",
		"scorecard   Show release readiness, primary bottleneck, and next action.",
		"diagnose    Explain what is blocking delivery and what to inspect next.",
		"trace       Follow one intent, behavior, or evidence ID end-to-end.",
		"ingest      Convert test, security, and telemetry reports into Bottleneck evidence.",
		"snapshot    Write scorecard snapshots for local trend history.",
		"seed-history Create demo scorecard history for local trends.",
		"trends      Analyze scorecard trends from local snapshots.",
		"report      Generate a leadership-ready SDLC evidence report.",
		"explain     Show how evidence affected category scores.",
	}
	for _, substring := range expected {
		if !strings.Contains(help, substring) {
			t.Fatalf("expected help to contain %q:\n%s", substring, help)
		}
	}
}

func TestScorecardCommandHelpPositionsScorecardAsMainSurface(t *testing.T) {
	description := scorecardCmd.Short + "\n" + scorecardCmd.Long
	expected := []string{
		"release readiness",
		"primary bottleneck",
		"next action",
		"Start here",
		"bottleneck init --template saas",
		"bottleneck scorecard",
		"bottleneck diagnose",
		"--details",
		"evidence, thresholds, and score impacts",
	}

	for _, substring := range expected {
		if !strings.Contains(description, substring) {
			t.Fatalf("expected scorecard command help to contain %q:\n%s", substring, description)
		}
	}
}

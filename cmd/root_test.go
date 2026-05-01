package cmd

import (
	"strings"
	"testing"
)

func TestRootCommandProductLanguageMatchesRoadmap(t *testing.T) {
	description := rootCmd.Short + "\n" + rootCmd.Long
	expected := []string{
		"CLI",
		"diagnoses delivery risk",
		"BIASED evidence model",
		"evidence artifacts",
		"scorecards",
		"release readiness",
		"bottleneck diagnosis",
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

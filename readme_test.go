package main

import (
	"os"
	"strings"
	"testing"
)

func TestREADMEProductScopeLanguage(t *testing.T) {
	content, err := os.ReadFile("readme.md")
	if err != nil {
		t.Fatalf("read readme.md: %v", err)
	}
	readme := string(content)

	expected := []string{
		"BIASED is the evidence model.",
		"Bottleneck is the CLI that diagnoses delivery risk using that model.",
		"Bottleneck evaluates a repo or release using local evidence artifacts.",
		"Team and organization-level views can come later by aggregating repo scorecards.",
		"single application repo",
		"service repo",
		"AI feature repo",
		"platform repo",
		"Bottleneck works for any software system, but it is especially useful for AI-enabled systems where behavior, drift, evaluation, and governance cannot be inferred from code alone.",
		"An AI PDF Risk Summarizer needs evidence that ambiguous financial risk language is flagged instead of summarized as fact.",
		"A payments service needs evidence that checkout behavior, test results, security checks, and production telemetry are connected before release.",
		"diagnoses delivery risk from local evidence artifacts",
		"release readiness",
		"hidden delivery risk",
	}

	for _, substring := range expected {
		if !strings.Contains(readme, substring) {
			t.Fatalf("expected README to contain %q", substring)
		}
	}

	if strings.Contains(strings.ToLower(readme), "framework validation only") {
		t.Fatal("README should not frame Bottleneck as framework validation only")
	}
}

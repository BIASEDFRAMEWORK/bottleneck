package main

import (
	"path/filepath"
	"strings"
	"testing"

	"bottleneck/internal/diagnosis"
	"bottleneck/internal/models"
	"bottleneck/internal/scorecard"
	"bottleneck/internal/traceability"
	"bottleneck/internal/validator"
)

func TestCompleteSaaSFixturePassesScorecardDiagnosisAndTrace(t *testing.T) {
	fixtureRoot := filepath.Join("internal", "traceability", "testdata", "complete-saas")

	result := validator.NewEngine(fixtureRoot, "default").Validate()
	if result.SystemStatus != models.StatusPass {
		t.Fatalf("expected complete SaaS fixture to pass, got %q with results %#v", result.SystemStatus, result.Results)
	}

	card := scorecard.Build(result)
	if card.ReleaseRecommendation != scorecard.RecommendationProceed {
		t.Fatalf("expected positive release recommendation, got %q", card.ReleaseRecommendation)
	}
	if card.PrimaryBottleneck != diagnosis.HealthyPrimaryBottleneck {
		t.Fatalf("expected no blocking bottleneck, got %q", card.PrimaryBottleneck)
	}

	report := diagnosis.Analyze(result)
	if report.PrimaryBottleneck != diagnosis.HealthyPrimaryBottleneck {
		t.Fatalf("expected healthy diagnosis, got %#v", report)
	}
	if !strings.Contains(report.RecommendedAction, "Keep evidence current") {
		t.Fatalf("expected healthy recommended action, got %q", report.RecommendedAction)
	}

	graph, err := traceability.Build(filepath.Join(fixtureRoot, "bottleneck"), traceability.Options{Environment: "default"})
	if err != nil {
		t.Fatalf("traceability build returned error: %v", err)
	}
	if findings := graph.ValidateFindings(); len(findings) != 0 {
		t.Fatalf("expected complete SaaS traceability fixture to have no findings, got %#v", findings)
	}

	trace, err := graph.Trace("BEHAVIOR-001")
	if err != nil {
		t.Fatalf("trace BEHAVIOR-001 returned error: %v", err)
	}
	if len(trace.RelatedIntent) == 0 || len(trace.RelatedDesign) == 0 || len(trace.RelatedAssurance) == 0 || len(trace.RelatedSecurity) == 0 || len(trace.RelatedExecution) == 0 {
		t.Fatalf("expected complete evidence chain for BEHAVIOR-001, got %#v", trace)
	}
	if len(trace.MissingLinks) != 0 {
		t.Fatalf("expected no missing links in complete SaaS trace, got %#v", trace.MissingLinks)
	}
}

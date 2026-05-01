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
		"Bottleneck helps a SaaS team diagnose delivery risk before a release",
		"local evidence artifacts",
		"release readiness scorecard",
		"primary bottleneck diagnosis",
		"delivery risk",
		"release readiness",
		"BIASED is the evidence model.",
		"Bottleneck is the CLI that diagnoses delivery risk using that model.",
		"Team and organization-level views can come later by aggregating repo scorecards.",
		"single application repo",
		"service repo",
		"AI feature repo",
		"platform repo",
		"Bottleneck works for any software system, but it is especially useful for AI-enabled systems where behavior, drift, evaluation, and governance cannot be inferred from code alone.",
		"An AI PDF Risk Summarizer needs evidence that ambiguous financial risk language is flagged instead of summarized as fact.",
		"A payments service needs evidence that checkout behavior, test results, security checks, and production telemetry are connected before release.",
		"diagnoses delivery risk from local evidence artifacts",
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

func TestREADMESaaSQuickstartContent(t *testing.T) {
	content, err := os.ReadFile("readme.md")
	if err != nil {
		t.Fatalf("read readme.md: %v", err)
	}
	readme := string(content)

	expected := []string{
		"## SaaS Team Quickstart",
		"bottleneck init --template saas",
		"bottleneck validate",
		"bottleneck scorecard",
		"bottleneck scorecard --format=json",
		"bottleneck scorecard --env=production",
		"bottleneck diagnose",
		"bottleneck trace BEHAVIOR-003",
		"bottleneck/intent/intent.md",
		"bottleneck/behavior/behavior-spec.md",
		"bottleneck/design/architecture.md",
		"bottleneck/assurance/results.json",
		"bottleneck/security/guardrails.json",
		"bottleneck/execution/telemetry.json",
		"Bad starter scorecard example",
		"`scorecard` is the primary Day-One command",
		"Release Recommendation: Conditional",
		"Primary Bottleneck: Assurance",
		"Reason: BEHAVIOR-003 is not linked to any passing test evidence.",
		"Impact: Release confidence is reduced because payment retry behavior is unproven.",
		"BEHAVIOR-003",
		"no mapped test evidence",
		"Next Action:",
		"Add or ingest test evidence mapped to BEHAVIOR-003.",
		"Inspect: bottleneck trace BEHAVIOR-003",
		"Good scorecard example",
		"Release Recommendation: Proceed",
		"bottleneck scorecard --details",
		"docs/quickstart-saas.md",
		"examples/saas/reports/",
		"bottleneck ingest cucumber --file reports/cucumber.json",
		"bottleneck ingest sarif --file reports/codeql.sarif",
		"bottleneck ingest test-summary --file reports/test-summary.json",
		"bottleneck ingest telemetry --file reports/telemetry.json",
		"Minimal CI usage",
		"examples/github-actions/bottleneck-saas-scorecard.yml",
		"cp examples/github-actions/bottleneck-saas-scorecard.yml .github/workflows/bottleneck.yml",
		"GitHub Actions step summary",
		"bottleneck diagnose --format=github",
		"bottleneck validate --github-annotations",
		"Warnings can appear in the scorecard without failing the job",
		"bottleneck scorecard --env=stage --format=markdown",
		"bottleneck diagnose --gate=release",
		"local`, `dev`, `test`, `stage`, and `production",
		"unknown environment names fail",
	}

	for _, substring := range expected {
		if !strings.Contains(readme, substring) {
			t.Fatalf("expected README SaaS quickstart to contain %q", substring)
		}
	}
}

func TestSaaSQuickstartGuideContent(t *testing.T) {
	content, err := os.ReadFile("docs/quickstart-saas.md")
	if err != nil {
		t.Fatalf("read docs/quickstart-saas.md: %v", err)
	}
	guide := string(content)

	expected := []string{
		"# SaaS Day-One Quickstart",
		"bottleneck init --template saas",
		"bottleneck validate",
		"bottleneck scorecard",
		"bottleneck scorecard --details",
		"bottleneck scorecard --format=json",
		"bottleneck scorecard --env=dev",
		"bottleneck scorecard --env=production",
		"bottleneck diagnose",
		"bottleneck trace BEHAVIOR-003",
		"bottleneck trace --id BEHAVIOR-003",
		"Break And Restore The Link",
		"Fix The Evidence Gap",
		"bottleneck/intent/intent.md",
		"bottleneck/behavior/behavior-spec.md",
		"bottleneck/design/architecture.md",
		"bottleneck/assurance/results.json",
		"bottleneck/security/guardrails.json",
		"bottleneck/execution/telemetry.json",
		"Primary Bottleneck: Assurance",
		"Reason: BEHAVIOR-003 is not linked to any passing test evidence.",
		"Impact: Release confidence is reduced because payment retry behavior is unproven.",
		"BEHAVIOR-003",
		"no mapped test evidence",
		"Next Action",
		"Inspect: bottleneck trace BEHAVIOR-003",
		"examples/saas/reports/",
		"bottleneck ingest cucumber --file reports/cucumber.json",
		"bottleneck ingest sarif --file reports/codeql.sarif",
		"bottleneck ingest test-summary --file reports/test-summary.json",
		"bottleneck ingest telemetry --file reports/telemetry.json",
		"bottleneck/assurance/results.json",
		"bottleneck/security/guardrails.json",
		"bottleneck/execution/telemetry.json",
		"After ingesting the Cucumber or test-summary sample, Assurance changes from `Warn` to `Pass`",
		"Choose An Environment",
		"Effective Thresholds:",
		"- Minimum score: 85",
		"- Required traceability: true",
		"- Critical security findings allowed: 0",
		"- Stale telemetry allowed: false",
		"unknown name and the supported environment list",
		"payment retry behavior",
		"Release Recommendation: Conditional",
		"bottleneck diagnose --format=github",
		"bottleneck validate --github-annotations",
		"examples/github-actions/bottleneck-saas-scorecard.yml",
		"cp examples/github-actions/bottleneck-saas-scorecard.yml .github/workflows/bottleneck.yml",
		"GitHub Actions step summary",
		"The workflow does not post a PR comment",
		"The release gate step controls whether CI blocks.",
		"Warnings can appear in the scorecard without failing the workflow",
		"bottleneck scorecard --env=stage --format=markdown",
		"bottleneck diagnose --env=production --gate=release",
	}

	for _, substring := range expected {
		if !strings.Contains(guide, substring) {
			t.Fatalf("expected SaaS quickstart guide to contain %q", substring)
		}
	}
}

func TestEnterpriseSDLCEvidenceGuideContent(t *testing.T) {
	readmeContent, err := os.ReadFile("readme.md")
	if err != nil {
		t.Fatalf("read readme.md: %v", err)
	}
	if !strings.Contains(string(readmeContent), "docs/enterprise-sdlc-evidence.md") {
		t.Fatal("README should link to enterprise SDLC evidence docs")
	}

	content, err := os.ReadFile("docs/enterprise-sdlc-evidence.md")
	if err != nil {
		t.Fatalf("read docs/enterprise-sdlc-evidence.md: %v", err)
	}
	guide := string(content)

	expected := []string{
		"# Enterprise SDLC Evidence with Bottleneck",
		"## What Bottleneck Does",
		"## What Bottleneck Does Not Do",
		"## Why Git Is the System of Record",
		"## Recommended Team Workflow",
		"## Recommended CI Workflow",
		"## Snapshot History",
		"## Trend Analysis",
		"## Evidence Explanation",
		"## Leadership Report",
		"## How to Interpret Results",
		"## Common Enterprise Scenarios",
		"## Frequently Asked Questions",
		"Bottleneck is not a dashboard.",
		"Bottleneck is not Jira replacement.",
		"Bottleneck is not a project-management system.",
		"Bottleneck is not a database-backed metrics warehouse.",
		"Bottleneck is a local-first evidence system for understanding SDLC bottlenecks.",
		"local files as the source of evidence",
		"Git is the system of record",
		"no external storage is required",
		"No dashboard is required",
		"Reports are generated from evidence, not opinion.",
		"Developers get a structured way",
		"Tech leads get a way",
		"Product leaders get a way",
		"Security teams get evidence",
		"Leadership gets a decision-ready summary",
		"bottleneck init --template saas",
		"bottleneck ingest cucumber --file reports/cucumber.json",
		"bottleneck ingest sarif --file reports/codeql.sarif",
		"bottleneck ingest telemetry --file reports/telemetry.json",
		"bottleneck scorecard",
		"bottleneck snapshot",
		"bottleneck trends",
		"bottleneck explain",
		"bottleneck report",
		"bottleneck scorecard --env=production --format=json",
		"bottleneck snapshot --env=production --label=ci",
		"bottleneck trends --env=production --window=6",
		"bottleneck report --env=production --format=markdown",
		"examples/github-actions/bottleneck-evidence-report.yml",
		"bottleneck/history/",
		"bottleneck/reports/",
		"does not commit generated files automatically",
		"CI-generated artifacts",
		"Committed snapshots",
		"Proceed",
		"Conditional",
		"Block",
		"Insufficient Evidence",
		"improving trends",
		"declining trends",
		"stable trends",
		"recovered categories",
		"regressed categories",
		"persistent bottleneck",
		"Team is shipping quickly but validation is weak",
		"Security is repeatedly discovered late",
		"Requirements are implemented but not traceable",
		"Production behavior does not match expected behavior",
		"Leadership wants faster delivery but the SDLC evidence is declining",
		"Does Bottleneck need a database?",
		"Does Bottleneck replace Jira?",
		"Do we need to commit snapshots?",
		"Can CI generate reports without committing them?",
		"What if we only have one snapshot?",
		"What if evidence is missing?",
		"How do teams use this in leadership reviews?",
	}

	for _, substring := range expected {
		if !strings.Contains(guide, substring) {
			t.Fatalf("expected enterprise SDLC evidence guide to contain %q", substring)
		}
	}
}

package report

import (
	"strings"
	"time"

	"bottleneck/internal/explainer"
	"bottleneck/internal/scorecard"
	"bottleneck/internal/trends"
)

const (
	SchemaVersion       = "sdlc.evidence.report.v1"
	FormatMarkdown      = "markdown"
	FormatJSON          = "json"
	DefaultReportPath   = "bottleneck/reports/sdlc-evidence-report.md"
	DefaultJSONFilePath = "bottleneck/reports/sdlc-evidence-report.json"
)

type Report struct {
	SchemaVersion         string                  `json:"schema_version"`
	GeneratedAt           string                  `json:"generated_at"`
	Environment           string                  `json:"environment"`
	CurrentStatus         string                  `json:"current_status"`
	ReleaseRecommendation string                  `json:"release_recommendation"`
	PrimaryBottleneck     string                  `json:"primary_bottleneck"`
	Scorecard             scorecard.Scorecard     `json:"scorecard"`
	TrendSummary          *trends.Analysis        `json:"trend_summary,omitempty"`
	Explanation           explainer.Report        `json:"explanation"`
	EvidenceFound         []string                `json:"evidence_found"`
	EvidenceMissing       []string                `json:"evidence_missing"`
	Risks                 []string                `json:"risks"`
	RecommendedActions    []string                `json:"recommended_actions"`
	SuggestedOwners       []string                `json:"suggested_owners"`
	SuggestedAutomation   []string                `json:"suggested_automation"`
	LeadershipDecision    string                  `json:"leadership_decision"`
	SnapshotMetadata      SnapshotMetadataSummary `json:"snapshot_metadata"`
}

type SnapshotMetadataSummary struct {
	SnapshotCount int      `json:"snapshot_count"`
	Window        int      `json:"window"`
	Environment   string   `json:"environment"`
	Warnings      []string `json:"warnings,omitempty"`
}

type Options struct {
	GeneratedAt time.Time
}

func Build(card scorecard.Scorecard, trendAnalysis *trends.Analysis, explanation explainer.Report, options Options) Report {
	card = scorecard.EnsureStableContract(card)
	generatedAt := options.GeneratedAt
	if generatedAt.IsZero() {
		generatedAt = time.Now().UTC()
	} else {
		generatedAt = generatedAt.UTC()
	}

	selected := selectedExplanations(card.PrimaryBottleneck, explanation.Explanations)
	report := Report{
		SchemaVersion:         SchemaVersion,
		GeneratedAt:           generatedAt.Format(time.RFC3339),
		Environment:           card.Environment,
		CurrentStatus:         card.SystemStatus,
		ReleaseRecommendation: card.ReleaseRecommendation,
		PrimaryBottleneck:     card.PrimaryBottleneck,
		Scorecard:             card,
		TrendSummary:          trendAnalysis,
		Explanation:           explanation,
		EvidenceFound:         evidenceFound(selected),
		EvidenceMissing:       evidenceMissing(selected),
		Risks:                 risks(selected),
		RecommendedActions:    recommendedActions(selected),
		SuggestedOwners:       suggestedOwners(selected),
		SuggestedAutomation:   suggestedAutomation(selected),
		LeadershipDecision:    leadershipDecision(card.PrimaryBottleneck),
	}
	if trendAnalysis != nil {
		report.SnapshotMetadata = SnapshotMetadataSummary{
			SnapshotCount: trendAnalysis.SnapshotCount,
			Window:        trendAnalysis.Window,
			Environment:   trendAnalysis.Environment,
			Warnings:      append([]string{}, trendAnalysis.Warnings...),
		}
	}
	return report
}

func selectedExplanations(primary string, explanations []explainer.CategoryExplanation) []explainer.CategoryExplanation {
	selected := []explainer.CategoryExplanation{}
	seen := map[string]struct{}{}
	for _, explanation := range explanations {
		if strings.EqualFold(explanation.Category, primary) || explanation.Status != scorecard.StatusPass || explanation.Score < 85 {
			selected = append(selected, explanation)
			seen[explanation.Category] = struct{}{}
		}
	}
	if len(selected) == 0 {
		for _, explanation := range explanations {
			if _, ok := seen[explanation.Category]; ok {
				continue
			}
			selected = append(selected, explanation)
		}
	}
	return selected
}

func evidenceFound(explanations []explainer.CategoryExplanation) []string {
	values := []string{}
	for _, explanation := range explanations {
		for _, fact := range explanation.EvidenceFound {
			values = append(values, prefixCategory(explanation.Category, fact.Text))
		}
	}
	return uniqueStrings(values, "No concrete evidence found.")
}

func evidenceMissing(explanations []explainer.CategoryExplanation) []string {
	values := []string{}
	for _, explanation := range explanations {
		for _, gap := range explanation.EvidenceMissing {
			if strings.EqualFold(strings.TrimSpace(gap.Text), "None.") {
				continue
			}
			values = append(values, prefixCategory(explanation.Category, gap.Text))
		}
	}
	return uniqueStrings(values, "No missing evidence was identified for the selected categories.")
}

func risks(explanations []explainer.CategoryExplanation) []string {
	values := []string{}
	for _, explanation := range explanations {
		values = append(values, prefixCategory(explanation.Category, explanation.RiskToDelivery))
	}
	return uniqueStrings(values, "No delivery risk was identified from the selected categories.")
}

func recommendedActions(explanations []explainer.CategoryExplanation) []string {
	values := []string{}
	for _, explanation := range explanations {
		for _, action := range explanation.RecommendedActions {
			values = append(values, action)
		}
	}
	return uniqueStrings(values, "Maintain current evidence and continue monitoring delivery risk.")
}

func suggestedOwners(explanations []explainer.CategoryExplanation) []string {
	values := []string{}
	for _, explanation := range explanations {
		values = append(values, explanation.SuggestedOwnerRoles...)
	}
	return uniqueStrings(values, "Technical Lead")
}

func suggestedAutomation(explanations []explainer.CategoryExplanation) []string {
	values := []string{}
	for _, explanation := range explanations {
		values = append(values, explanation.SuggestedAutomations...)
	}
	return uniqueStrings(values, "Run bottleneck validate in CI.")
}

func leadershipDecision(primaryBottleneck string) string {
	switch strings.ToLower(strings.TrimSpace(primaryBottleneck)) {
	case "assurance":
		return "Approve time to map critical behaviors to validation evidence before accelerating release."
	case "security":
		return "Decide whether release should be blocked until high-severity findings are resolved or formally accepted."
	case "intent":
		return "Align leadership, product, domain, and engineering on measurable intent before expanding implementation."
	case "behavior":
		return "Confirm expected and unacceptable behavior before expanding implementation or release scope."
	case "design":
		return "Decide whether architecture tradeoffs, dependencies, and failure modes are clear enough for release approval."
	case "execution":
		return "Prioritize production telemetry, adoption feedback, and operational readiness before scaling usage."
	case "", "none", "no bottleneck":
		return "Continue monitoring evidence trends and maintain current release controls."
	default:
		return "Review the primary bottleneck and decide whether ownership, time, or release gating needs adjustment."
	}
}

func prefixCategory(category string, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return category + ": " + value
}

func uniqueStrings(values []string, fallback string) []string {
	seen := map[string]struct{}{}
	unique := []string{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		unique = append(unique, value)
	}
	if len(unique) == 0 && fallback != "" {
		return []string{fallback}
	}
	return unique
}

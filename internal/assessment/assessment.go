package assessment

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"bottleneck/internal/discover"
	"bottleneck/internal/models"
	"bottleneck/internal/scorecard"
)

const (
	SchemaVersion = "assessment.v1"

	FormatText     = "text"
	FormatJSON     = "json"
	FormatMarkdown = "markdown"
)

type Options struct {
	RootPath     string
	Environment  string
	AutoIngested bool
	Warnings     []string
	Now          time.Time
}

type Report struct {
	SchemaVersion         string             `json:"schema_version"`
	Environment           string             `json:"environment"`
	Maturity              Maturity           `json:"maturity"`
	AIReadiness           string             `json:"ai_readiness"`
	ReleaseFriction       string             `json:"release_friction"`
	PrimaryBottleneck     string             `json:"primary_bottleneck"`
	ScoreConfidence       string             `json:"score_confidence"`
	ReleaseRecommendation string             `json:"release_recommendation"`
	ScoreRationale        []string           `json:"score_rationale"`
	Found                 []string           `json:"found"`
	Warnings              []string           `json:"warnings,omitempty"`
	NextAction            string             `json:"next_action"`
	UsefulCommands        []string           `json:"useful_commands"`
	Categories            []CategoryEvidence `json:"categories"`
}

type Maturity struct {
	Level int    `json:"level"`
	Label string `json:"label"`
}

type CategoryEvidence struct {
	Name            string   `json:"name"`
	Status          string   `json:"status"`
	Score           int      `json:"score"`
	Confidence      string   `json:"confidence"`
	Freshness       string   `json:"freshness,omitempty"`
	Provenance      string   `json:"provenance"`
	Rationale       string   `json:"rationale"`
	EvidenceFound   []string `json:"evidence_found,omitempty"`
	EvidenceMissing []string `json:"evidence_missing,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

func Build(result models.EngineResult, discovery discover.DiscoveryResult, options Options) Report {
	if options.RootPath == "" {
		options.RootPath = "."
	}
	if options.Now.IsZero() {
		options.Now = time.Now().UTC()
	}

	card := scorecard.Build(result)
	categories := categoryEvidence(card.Capabilities, discovery, options)
	maturity := maturityFor(card, discovery, options.RootPath)
	scoreConfidence := confidenceFor(categories, discovery)
	recommendation := card.ReleaseRecommendation
	primary := card.PrimaryBottleneck
	if strings.TrimSpace(primary) == "" {
		primary = result.PrimaryBottleneck
	}
	if strings.TrimSpace(primary) == "" {
		primary = "Unknown"
	}

	report := Report{
		SchemaVersion:         SchemaVersion,
		Environment:           coalesce(result.Environment, options.Environment, "default"),
		Maturity:              maturity,
		AIReadiness:           aiReadinessFor(maturity, recommendation, scoreConfidence),
		ReleaseFriction:       releaseFrictionFor(recommendation, maturity),
		PrimaryBottleneck:     primary,
		ScoreConfidence:       scoreConfidence,
		ReleaseRecommendation: recommendation,
		ScoreRationale:        rationaleFor(maturity, scoreConfidence, discovery, options.AutoIngested),
		Found:                 foundEvidence(discovery),
		Warnings:              append([]string{}, options.Warnings...),
		NextAction:            nextAction(card),
		UsefulCommands:        usefulCommands(primary),
		Categories:            categories,
	}
	report.Warnings = append(report.Warnings, discovery.Warnings...)
	if len(report.Found) == 0 {
		report.Warnings = append(report.Warnings, "No recognizable local evidence was discovered.")
	}
	return report
}

func Render(report Report, format string) (string, error) {
	switch strings.ToLower(format) {
	case FormatText:
		return renderText(report), nil
	case FormatJSON:
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return "", err
		}
		return string(data), nil
	case FormatMarkdown:
		return renderMarkdown(report), nil
	default:
		return "", fmt.Errorf("unsupported format %q (supported: text, json, markdown)", format)
	}
}

func categoryEvidence(capabilities []scorecard.CapabilityScorecard, discovery discover.DiscoveryResult, options Options) []CategoryEvidence {
	categories := make([]CategoryEvidence, 0, len(capabilities))
	for _, capability := range capabilities {
		freshness := freshnessFor(options.RootPath, capability.Capability, options.Now)
		provenance := provenanceFor(options.RootPath, capability.Capability, discovery)
		confidence := categoryConfidence(capability, provenance, freshness)
		categories = append(categories, CategoryEvidence{
			Name:            capability.Capability,
			Status:          capability.Status,
			Score:           capability.Score,
			Confidence:      confidence,
			Freshness:       freshness,
			Provenance:      provenance,
			Rationale:       capability.Reason,
			EvidenceFound:   append([]string{}, capability.Evidence...),
			EvidenceMissing: append([]string{}, capability.MissingEvidence...),
			Recommendations: []string{capability.RecommendedAction},
		})
	}
	sort.SliceStable(categories, func(i, j int) bool {
		return categoryRank(categories[i].Name) < categoryRank(categories[j].Name)
	})
	return categories
}

func maturityFor(card scorecard.Scorecard, discovery discover.DiscoveryResult, rootPath string) Maturity {
	hasAny := len(discovery.Findings) > 0
	hasNative := discovery.Summary.CountsByCategory[discover.CategoryNative] > 0
	hasAssurance := hasKind(discovery, "cucumber", "junit", "coverage", "test-summary", "bottleneck-assurance")
	hasSecurity := hasKind(discovery, "sarif", "bottleneck-security")
	hasTelemetry := hasKind(discovery, "telemetry", "bottleneck-execution")
	hasWorkflow := hasKind(discovery, "github-actions")
	hasDesign := discovery.Summary.CountsByCategory[discover.CategoryDesign] > 0 || hasKind(discovery, "bottleneck-doc")

	switch {
	case !hasAny:
		return Maturity{Level: 0, Label: "Ad Hoc"}
	case historicalSnapshots(rootPath) >= 3 && card.SystemStatus != scorecard.StatusFail && hasAssurance && hasSecurity && hasTelemetry:
		return Maturity{Level: 4, Label: "Optimized"}
	case card.SystemStatus != scorecard.StatusFail && hasAssurance && hasSecurity && hasTelemetry && hasWorkflow:
		return Maturity{Level: 3, Label: "Measured"}
	case hasAssurance || hasSecurity || hasTelemetry || hasNative:
		return Maturity{Level: 2, Label: "Managed"}
	case hasDesign:
		return Maturity{Level: 1, Label: "Documented"}
	default:
		return Maturity{Level: 0, Label: "Ad Hoc"}
	}
}

func confidenceFor(categories []CategoryEvidence, discovery discover.DiscoveryResult) string {
	high := 0
	medium := 0
	for _, category := range categories {
		switch category.Confidence {
		case "high":
			high++
		case "medium":
			medium++
		}
	}
	if high >= 3 && discovery.Summary.CountsByCategory[discover.CategoryNative] > 0 {
		return "high"
	}
	if high+medium >= 2 || len(discovery.Findings) > 0 {
		return "medium"
	}
	return "low"
}

func categoryConfidence(category scorecard.CapabilityScorecard, provenance string, freshness string) string {
	if category.Status == scorecard.StatusFail || freshness == "stale" || freshness == "missing" {
		if provenance == "tool-generated" || provenance == "native Bottleneck evidence" {
			return "medium"
		}
		return "low"
	}
	if provenance == "tool-generated" || provenance == "native Bottleneck evidence" {
		return "high"
	}
	return "medium"
}

func freshnessFor(rootPath string, category string, now time.Time) string {
	path, maxAge, ok := freshnessPath(category)
	if !ok {
		return "not age-gated"
	}
	info, err := os.Stat(filepath.Join(rootPath, path))
	if err != nil {
		return "missing"
	}
	if now.Sub(info.ModTime()) > maxAge {
		return "stale"
	}
	return "fresh"
}

func freshnessPath(category string) (string, time.Duration, bool) {
	switch strings.ToLower(category) {
	case "assurance":
		return "bottleneck/assurance/results.json", 14 * 24 * time.Hour, true
	case "security":
		return "bottleneck/security/guardrails.json", 14 * 24 * time.Hour, true
	case "execution":
		return "bottleneck/execution/telemetry.json", 7 * 24 * time.Hour, true
	default:
		return "", 0, false
	}
}

func provenanceFor(rootPath string, category string, discovery discover.DiscoveryResult) string {
	switch strings.ToLower(category) {
	case "assurance":
		if artifactHasToolProvenance(filepath.Join(rootPath, "bottleneck/assurance/results.json")) {
			return "tool-generated"
		}
	case "security":
		if artifactHasToolProvenance(filepath.Join(rootPath, "bottleneck/security/guardrails.json")) {
			return "tool-generated"
		}
	case "execution":
		if artifactHasToolProvenance(filepath.Join(rootPath, "bottleneck/execution/telemetry.json")) {
			return "tool-generated"
		}
	}
	if hasNativeForCategory(category, discovery) {
		return "native Bottleneck evidence"
	}
	if hasDiscoveryForCategory(category, discovery) {
		return "local artifact"
	}
	return "not found"
}

func artifactHasToolProvenance(path string) bool {
	content, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(content), `"generated_by"`) || strings.Contains(string(content), `"provenance"`)
}

func hasNativeForCategory(category string, discovery discover.DiscoveryResult) bool {
	switch strings.ToLower(category) {
	case "intent", "behavior", "design":
		return hasKind(discovery, "bottleneck-doc")
	case "assurance":
		return hasKind(discovery, "bottleneck-assurance")
	case "security":
		return hasKind(discovery, "bottleneck-security")
	case "execution":
		return hasKind(discovery, "bottleneck-execution")
	case "config":
		return hasKind(discovery, "config")
	default:
		return false
	}
}

func hasDiscoveryForCategory(category string, discovery discover.DiscoveryResult) bool {
	target := strings.ToLower(category)
	for _, finding := range discovery.Findings {
		switch target {
		case "assurance":
			if finding.Category == discover.CategoryAssurance {
				return true
			}
		case "security":
			if finding.Category == discover.CategorySecurity {
				return true
			}
		case "execution":
			if finding.Category == discover.CategoryExecution {
				return true
			}
		case "design":
			if finding.Category == discover.CategoryDesign {
				return true
			}
		}
	}
	return false
}

func aiReadinessFor(maturity Maturity, recommendation string, scoreConfidence string) string {
	if recommendation == scorecard.RecommendationBlock || maturity.Level == 0 {
		return "Blocked"
	}
	if maturity.Level <= 2 || scoreConfidence == "low" {
		return "Limited"
	}
	if maturity.Level >= 4 && scoreConfidence == "high" {
		return "Strong"
	}
	return "Ready With Guardrails"
}

func releaseFrictionFor(recommendation string, maturity Maturity) string {
	switch recommendation {
	case scorecard.RecommendationProceed:
		return "low"
	case scorecard.RecommendationConditional:
		return "medium"
	case scorecard.RecommendationBlock:
		return "high"
	default:
		if maturity.Level <= 1 {
			return "high"
		}
		return "medium"
	}
}

func rationaleFor(maturity Maturity, confidence string, discovery discover.DiscoveryResult, autoIngested bool) []string {
	rationale := []string{
		fmt.Sprintf("Maturity is Level %d (%s) based on discovered and native evidence coverage.", maturity.Level, maturity.Label),
		fmt.Sprintf("Score confidence is %s because Bottleneck found %d local evidence artifact(s).", confidence, len(discovery.Findings)),
	}
	if autoIngested {
		rationale = append(rationale, "Automatic ingestion normalized supported tool output before scoring.")
	}
	if len(discovery.Summary.Missing) > 0 {
		rationale = append(rationale, "Missing evidence: "+strings.Join(discovery.Summary.Missing, ", ")+".")
	}
	return rationale
}

func foundEvidence(discovery discover.DiscoveryResult) []string {
	found := make([]string, 0, len(discovery.Findings))
	for _, finding := range discovery.Findings {
		found = append(found, fmt.Sprintf("%s: %s", finding.Kind, finding.Path))
	}
	return found
}

func nextAction(card scorecard.Scorecard) string {
	if strings.TrimSpace(card.Diagnosis.RecommendedAction) != "" {
		return card.Diagnosis.RecommendedAction
	}
	for _, capability := range card.Capabilities {
		if capability.Capability == card.PrimaryBottleneck && capability.RecommendedAction != "" {
			return capability.RecommendedAction
		}
	}
	return "Run `bottleneck discover`, then add or ingest evidence for the weakest category."
}

func usefulCommands(primary string) []string {
	commands := []string{
		"bottleneck discover",
		"bottleneck ingest --auto",
		"bottleneck explain-score",
	}
	if primary != "" && primary != "Unknown" && primary != "Healthy" {
		commands = append(commands, "bottleneck trace BEHAVIOR-003")
	}
	return commands
}

func hasKind(discovery discover.DiscoveryResult, kinds ...string) bool {
	wanted := map[string]struct{}{}
	for _, kind := range kinds {
		wanted[kind] = struct{}{}
	}
	for _, finding := range discovery.Findings {
		if _, ok := wanted[finding.Kind]; ok {
			return true
		}
	}
	return false
}

func historicalSnapshots(rootPath string) int {
	entries, err := os.ReadDir(filepath.Join(rootPath, "bottleneck/history/scorecards"))
	if err != nil {
		return 0
	}
	count := 0
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			count++
		}
	}
	return count
}

func renderText(report Report) string {
	lines := []string{
		"Bottleneck Assessment",
		fmt.Sprintf("Environment: %s", report.Environment),
		fmt.Sprintf("Maturity: Level %d - %s", report.Maturity.Level, report.Maturity.Label),
		fmt.Sprintf("AI Readiness: %s", report.AIReadiness),
		fmt.Sprintf("Release Recommendation: %s", report.ReleaseRecommendation),
		fmt.Sprintf("Release Friction: %s", report.ReleaseFriction),
		fmt.Sprintf("Primary Bottleneck: %s", report.PrimaryBottleneck),
		fmt.Sprintf("Score Confidence: %s", report.ScoreConfidence),
		"",
		"Why:",
	}
	for _, rationale := range report.ScoreRationale {
		lines = append(lines, "  - "+rationale)
	}
	lines = append(lines, "", "Categories:")
	for _, category := range report.Categories {
		lines = append(lines, fmt.Sprintf("  - %s: %s score=%d confidence=%s provenance=%s freshness=%s", category.Name, category.Status, category.Score, category.Confidence, category.Provenance, category.Freshness))
	}
	if len(report.Found) > 0 {
		lines = append(lines, "", "Found:")
		for _, item := range report.Found {
			lines = append(lines, "  - "+item)
		}
	}
	if len(report.Warnings) > 0 {
		lines = append(lines, "", "Warnings:")
		for _, warning := range report.Warnings {
			lines = append(lines, "  - "+warning)
		}
	}
	lines = append(lines, "", "Next action: "+report.NextAction, "", "Useful commands:")
	for _, command := range report.UsefulCommands {
		lines = append(lines, "  - "+command)
	}
	return strings.Join(lines, "\n")
}

func renderMarkdown(report Report) string {
	lines := []string{
		"# Bottleneck Assessment",
		"",
		fmt.Sprintf("- **Environment:** %s", report.Environment),
		fmt.Sprintf("- **Maturity:** Level %d - %s", report.Maturity.Level, report.Maturity.Label),
		fmt.Sprintf("- **AI readiness:** %s", report.AIReadiness),
		fmt.Sprintf("- **Release recommendation:** %s", report.ReleaseRecommendation),
		fmt.Sprintf("- **Primary bottleneck:** %s", report.PrimaryBottleneck),
		fmt.Sprintf("- **Score confidence:** %s", report.ScoreConfidence),
		"",
		"| Category | Status | Score | Confidence | Provenance | Freshness |",
		"| --- | --- | ---: | --- | --- | --- |",
	}
	for _, category := range report.Categories {
		lines = append(lines, fmt.Sprintf("| %s | %s | %d | %s | %s | %s |", category.Name, category.Status, category.Score, category.Confidence, category.Provenance, category.Freshness))
	}
	if len(report.Found) > 0 {
		lines = append(lines, "", "## Found Evidence")
		for _, item := range report.Found {
			lines = append(lines, "- "+item)
		}
	}
	if len(report.Warnings) > 0 {
		lines = append(lines, "", "## Warnings")
		for _, warning := range report.Warnings {
			lines = append(lines, "- "+warning)
		}
	}
	lines = append(lines, "", "## Next Action", report.NextAction, "", "## Useful Commands")
	for _, command := range report.UsefulCommands {
		lines = append(lines, "- `"+command+"`")
	}
	return strings.Join(lines, "\n")
}

func categoryRank(name string) int {
	switch strings.ToLower(name) {
	case "config":
		return 0
	case "intent":
		return 1
	case "behavior":
		return 2
	case "design":
		return 3
	case "assurance":
		return 4
	case "security":
		return 5
	case "execution":
		return 6
	case "traceability":
		return 7
	default:
		return 99
	}
}

func coalesce(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

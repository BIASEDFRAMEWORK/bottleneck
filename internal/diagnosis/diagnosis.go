package diagnosis

import (
	"sort"
	"strconv"
	"strings"

	"bottleneck/internal/models"
)

const (
	HealthyPrimaryBottleneck = "None"

	ConfidenceHigh   = "High"
	ConfidenceMedium = "Medium"
	ConfidenceLow    = "Low"

	ScorePass    = 85
	ScoreWarning = 60
	ScoreFail    = 20
	ScoreUnknown = 40

	strongScoreThreshold = 80
)

type Diagnosis struct {
	PrimaryBottleneck    string          `json:"primary_bottleneck"`
	TiedBottlenecks      []string        `json:"tied_bottlenecks,omitempty"`
	Rule                 string          `json:"rule,omitempty"`
	Reason               string          `json:"reason"`
	Impact               string          `json:"impact"`
	NextAction           string          `json:"next_action"`
	InspectCommand       string          `json:"inspect_command"`
	RelevantEvidenceIDs  []string        `json:"relevant_evidence_ids,omitempty"`
	SupportingIssues     []string        `json:"supporting_issues,omitempty"`
	WhyItMatters         string          `json:"why_it_matters"`
	RecommendedAction    string          `json:"recommended_action"`
	ContributingFindings []string        `json:"contributing_findings,omitempty"`
	Confidence           string          `json:"confidence"`
	ConfidenceReason     string          `json:"confidence_reason"`
	CategoryScores       []CategoryScore `json:"category_scores"`
}

type CategoryScore struct {
	Category string `json:"category"`
	Score    int    `json:"score"`
	Status   string `json:"status"`
	Reason   string `json:"reason,omitempty"`
}

type CategoryInfo struct {
	Owner        string
	Bottleneck   string
	WhyItMatters string
	HealthyText  string
}

var categoryOrder = []string{"Behavior", "Intent", "Design", "Assurance", "Security", "Execution"}
var tiePriorityOrder = []string{"Assurance", "Security", "Behavior", "Intent", "Execution", "Design"}

var categoryInfo = map[string]CategoryInfo{
	"Config": {
		Owner:        "Execution Engineer",
		Bottleneck:   "Invalid configuration",
		WhyItMatters: "Environment and threshold configuration controls how evidence is interpreted. If it cannot be resolved, the release decision cannot be trusted.",
		HealthyText:  "Configuration thresholds are available for the selected environment.",
	},
	"Intent": {
		Owner:        "Intent Engineer",
		Bottleneck:   "Ambiguous requirements",
		WhyItMatters: "The team has not clearly defined what good looks like. This creates downstream ambiguity in design, testing, security, and release decisions.",
		HealthyText:  "Intent evidence is strong enough to guide downstream delivery decisions.",
	},
	"Behavior": {
		Owner:        "Behavior Engineer",
		Bottleneck:   "Non-deterministic outputs",
		WhyItMatters: "Expected and unacceptable behavior are not clear enough to verify. This makes generated or changed software harder to test, review, and release safely.",
		HealthyText:  "Behavior evidence is strong enough to describe what the system should and should not do.",
	},
	"Design": {
		Owner:        "Design Engineer",
		Bottleneck:   "Poor adoption / UX gaps",
		WhyItMatters: "The system design is not clear enough for engineers and operators to understand how the change works or how to evolve it safely.",
		HealthyText:  "Design evidence is strong enough to make the implementation reviewable.",
	},
	"Assurance": {
		Owner:        "Assurance Engineer",
		Bottleneck:   "Validation gaps",
		WhyItMatters: "There is not enough proof that the expected behavior was tested. Without assurance evidence, a release can look complete while still failing the behavior it was meant to satisfy.",
		HealthyText:  "Assurance evidence is strong enough to support the current release decision.",
	},
	"Security": {
		Owner:        "Security Engineer",
		Bottleneck:   "Risk & compliance",
		WhyItMatters: "Security or policy evidence is weak. That can let unsafe changes move forward before guardrails, compliance expectations, or operational risks are resolved.",
		HealthyText:  "Security evidence is strong enough to show no known guardrail violations.",
	},
	"Execution": {
		Owner:        "Execution Engineer",
		Bottleneck:   "Delivery friction",
		WhyItMatters: "Production or delivery evidence is weak. The team may not know whether the change is reliable, adopted, or creating operational friction after release.",
		HealthyText:  "Execution evidence is strong enough to show the system is working in practice.",
	},
	"Traceability": {
		Owner:        "Release Engineer",
		Bottleneck:   "Traceability gaps",
		WhyItMatters: "Traceability connects intent, behavior, assurance, security, and telemetry evidence so release decisions can be audited end to end.",
		HealthyText:  "Traceability evidence is strong enough to support the current release decision.",
	},
}

func Analyze(result models.EngineResult) Diagnosis {
	resultsByCategory := map[string]models.ValidationResult{}
	traceabilityResults := make([]models.ValidationResult, 0, 1)
	configFailure, hasConfigFailure := firstConfigFailure(result)
	for _, validation := range result.Results {
		if isDiagnosableCategory(validation.Capability) {
			resultsByCategory[validation.Capability] = validation
			continue
		}
		if validation.Capability == "Traceability" {
			traceabilityResults = append(traceabilityResults, validation)
		}
	}

	scores := make([]CategoryScore, 0, len(resultsByCategory))
	scoreByCategory := map[string]int{}
	for _, category := range categoryOrder {
		validation, ok := resultsByCategory[category]
		if !ok {
			continue
		}
		score := scoreValidation(validation)
		scoreByCategory[category] = score
		scores = append(scores, CategoryScore{
			Category: category,
			Score:    score,
			Status:   validation.Status,
			Reason:   reasonFor(validation),
		})
	}

	applyTraceabilityAdjustments(scoreByCategory, traceabilityResults)
	for index := range scores {
		scores[index].Score = clampScore(scoreByCategory[scores[index].Category])
	}
	confidence, confidenceReason := Confidence(result)

	if hasConfigFailure {
		findings := ContributingFindings(result, "Config")
		actionable := actionableFor(result, "Config", configFailure, traceabilityResults, findings)
		return Diagnosis{
			PrimaryBottleneck:    "Config",
			Rule:                 actionable.Rule,
			Reason:               actionable.Reason,
			Impact:               actionable.Impact,
			NextAction:           actionable.NextAction,
			InspectCommand:       actionable.InspectCommand,
			RelevantEvidenceIDs:  actionable.RelevantEvidenceIDs,
			SupportingIssues:     actionable.SupportingIssues,
			WhyItMatters:         WhyItMatters("Config"),
			RecommendedAction:    actionable.NextAction,
			ContributingFindings: findings,
			Confidence:           confidence,
			ConfidenceReason:     confidenceReason,
			CategoryScores:       scores,
		}
	}

	if len(scores) == 0 || allStrong(scores) {
		if rule, ok := highestActionableRule(result); ok && rule.Category == "Traceability" {
			findings := ContributingFindings(result, rule.Category)
			actionable := actionableFor(result, rule.Category, models.ValidationResult{}, traceabilityResults, findings)
			return Diagnosis{
				PrimaryBottleneck:    rule.Category,
				Rule:                 actionable.Rule,
				Reason:               actionable.Reason,
				Impact:               actionable.Impact,
				NextAction:           actionable.NextAction,
				InspectCommand:       actionable.InspectCommand,
				RelevantEvidenceIDs:  actionable.RelevantEvidenceIDs,
				SupportingIssues:     actionable.SupportingIssues,
				WhyItMatters:         WhyItMatters(rule.Category),
				RecommendedAction:    actionable.NextAction,
				ContributingFindings: findings,
				Confidence:           confidence,
				ConfidenceReason:     confidenceReason,
				CategoryScores:       scores,
			}
		}
		findings := ContributingFindings(result, HealthyPrimaryBottleneck)
		actionable := actionableFor(result, HealthyPrimaryBottleneck, models.ValidationResult{}, nil, findings)
		return Diagnosis{
			PrimaryBottleneck:    HealthyPrimaryBottleneck,
			Rule:                 actionable.Rule,
			Reason:               actionable.Reason,
			Impact:               actionable.Impact,
			NextAction:           actionable.NextAction,
			InspectCommand:       actionable.InspectCommand,
			RelevantEvidenceIDs:  actionable.RelevantEvidenceIDs,
			SupportingIssues:     actionable.SupportingIssues,
			WhyItMatters:         "All assessed delivery categories have enough evidence to support the current release decision.",
			RecommendedAction:    "Keep evidence current as intent, behavior, implementation, and production signals change.",
			ContributingFindings: findings,
			Confidence:           confidence,
			ConfidenceReason:     confidenceReason,
			CategoryScores:       scores,
		}
	}

	lowestScore := 101
	for _, score := range scores {
		if score.Score < lowestScore {
			lowestScore = score.Score
		}
	}

	tied := tiedBottlenecks(scores, lowestScore)
	primary := tied[0]
	rulePrimary := primaryFromActionableRules(result, primary)
	if rulePrimary != primary {
		primary = rulePrimary
		tied = []string{primary}
	}
	validation := resultsByCategory[primary]
	findings := ContributingFindings(result, primary)
	actionable := actionableFor(result, primary, validation, traceabilityResults, findings)

	return Diagnosis{
		PrimaryBottleneck:    primary,
		TiedBottlenecks:      tiedIfMultiple(tied),
		Rule:                 actionable.Rule,
		Reason:               actionable.Reason,
		Impact:               actionable.Impact,
		NextAction:           actionable.NextAction,
		InspectCommand:       actionable.InspectCommand,
		RelevantEvidenceIDs:  actionable.RelevantEvidenceIDs,
		SupportingIssues:     actionable.SupportingIssues,
		WhyItMatters:         WhyItMatters(primary),
		RecommendedAction:    recommendedActionForPrimary(primary, validation, traceabilityResults),
		ContributingFindings: findings,
		Confidence:           confidence,
		ConfidenceReason:     confidenceReason,
		CategoryScores:       scores,
	}
}

func firstConfigFailure(result models.EngineResult) (models.ValidationResult, bool) {
	for _, validation := range result.Results {
		if validation.Capability == "Config" && validation.Status != models.StatusPass {
			return validation, true
		}
	}
	return models.ValidationResult{}, false
}

func Info(category string) CategoryInfo {
	if info, ok := categoryInfo[category]; ok {
		return info
	}
	return CategoryInfo{
		Owner:        "Execution Engineer",
		Bottleneck:   "Delivery friction",
		WhyItMatters: "This capability contributes to the overall validity of the system and needs a clear owner for remediation.",
		HealthyText:  "Evidence is available for this capability.",
	}
}

func WhyItMatters(category string) string {
	return Info(category).WhyItMatters
}

func RecommendedAction(validation models.ValidationResult) string {
	category := validation.Capability
	if validation.Status == models.StatusPass {
		return Info(category).HealthyText
	}

	message := strings.ToLower(validation.Message)
	details := strings.ToLower(strings.Join(validation.Details, "\n"))
	combined := strings.TrimSpace(message + "\n" + details)

	switch {
	case category == "Config" && strings.Contains(combined, "no bottleneck config found"):
		return "Bottleneck has not been initialized in this directory. Initialize a SaaS starter project: bottleneck init --template saas."
	case category == "Config" && strings.Contains(combined, "unknown environment"):
		return "Choose a supported environment, for example: bottleneck scorecard --env=production."
	case category == "Assurance" && strings.Contains(combined, "no assurance evidence found"):
		return "Add test evidence manually or run: bottleneck ingest cucumber --file reports/cucumber.json."
	case category == "Assurance" && (strings.Contains(combined, "behavior-003") || strings.Contains(combined, "payment retry")):
		return "Add assurance evidence for payment retry behavior."
	case category == "Assurance" && strings.Contains(combined, "ambiguous risk"):
		return "Add or fix evaluation evidence for BEHAVIOR-001 so ambiguous financial risk language is flagged as uncertain."
	case strings.Contains(combined, "no mapped test evidence") ||
		strings.Contains(combined, "not linked") ||
		strings.Contains(combined, "references missing") ||
		strings.Contains(combined, "cannot reference") ||
		strings.Contains(combined, "orphan"):
		return disconnectedEvidenceAction(category)
	case strings.Contains(combined, "stale") ||
		strings.Contains(combined, "outdated") ||
		strings.Contains(combined, "expired"):
		return staleEvidenceAction(category)
	case strings.Contains(combined, "missing"):
		return missingEvidenceAction(category)
	case strings.Contains(combined, "placeholder") ||
		strings.Contains(combined, "too thin") ||
		strings.Contains(combined, "content quality"):
		return weakEvidenceAction(category)
	case strings.Contains(combined, "threshold") ||
		strings.Contains(combined, "low adoption") ||
		strings.Contains(combined, "violations"):
		return thresholdAction(category)
	}

	switch validation.Status {
	case models.StatusPass:
		return Info(category).HealthyText
	case models.StatusWarning:
		return weakEvidenceAction(category)
	case models.StatusFail:
		return missingEvidenceAction(category)
	default:
		return "Run validation and inspect the affected artifact before making a release decision."
	}
}

func recommendedActionForPrimary(primary string, validation models.ValidationResult, traceabilityResults []models.ValidationResult) string {
	if validation.Status == models.StatusPass {
		if action := traceabilityRecommendedAction(primary, traceabilityResults); action != "" {
			return action
		}
	}
	return RecommendedAction(validation)
}

func traceabilityRecommendedAction(primary string, traceabilityResults []models.ValidationResult) string {
	for _, validation := range traceabilityResults {
		if validation.Status == models.StatusPass || !traceabilityResultRelatesTo(validation, primary) {
			continue
		}
		combined := strings.ToLower(traceabilityCombinedText(validation))
		if primary == "Assurance" && (strings.Contains(combined, "behavior-003") ||
			strings.Contains(combined, "payment retry") ||
			strings.Contains(combined, "duplicate-charge") ||
			strings.Contains(combined, "duplicate charges")) {
			return "Add assurance evidence for payment retry behavior."
		}
		if primary == "Assurance" && (strings.Contains(combined, "no mapped test evidence") ||
			strings.Contains(combined, "no assurance result") ||
			strings.Contains(combined, "not linked to assurance evidence")) {
			return disconnectedEvidenceAction("Assurance")
		}
		return disconnectedEvidenceAction(primary)
	}
	return ""
}

func traceabilityCombinedText(validation models.ValidationResult) string {
	parts := []string{validation.Message}
	parts = append(parts, validation.Details...)
	for _, finding := range validation.Findings {
		parts = append(parts, finding.Message)
	}
	for _, impact := range validation.EvidenceQuality.ScoreImpacts {
		parts = append(parts, impact.Reason)
	}
	return strings.Join(parts, "\n")
}

func ScoreFor(category string, scores []CategoryScore) int {
	for _, score := range scores {
		if score.Category == category {
			return score.Score
		}
	}
	return 0
}

func HasCategoryScore(category string, scores []CategoryScore) bool {
	for _, score := range scores {
		if score.Category == category {
			return true
		}
	}
	return false
}

func ContributingFindings(result models.EngineResult, primary string) []string {
	if primary == HealthyPrimaryBottleneck {
		return []string{"All assessed delivery categories have enough evidence to support the current release decision."}
	}

	var findings []string
	for _, validation := range result.Results {
		if validation.Capability == primary {
			findings = appendFindingSources(findings, validation)
		}
	}
	for _, validation := range result.Results {
		if validation.Capability == "Traceability" && traceabilityResultRelatesTo(validation, primary) {
			findings = appendFindingSources(findings, validation)
		}
	}
	if len(findings) == 0 {
		findings = append(findings, Info(primary).Bottleneck)
	}

	return firstUnique(findings, 3)
}

func appendFindingSources(findings []string, validation models.ValidationResult) []string {
	for _, impact := range validation.EvidenceQuality.ScoreImpacts {
		findings = append(findings, impact.Reason)
	}
	for _, finding := range validation.Findings {
		findings = append(findings, finding.Message)
	}
	if validation.Capability == "Assurance" {
		findings = append(findings, validation.Details...)
		if validation.Message != "" && validation.Status != models.StatusPass {
			findings = append(findings, validation.Message)
		}
		findings = append(findings, validation.EvidenceQuality.Missing...)
		return findings
	}
	if validation.Message != "" && validation.Status != models.StatusPass {
		findings = append(findings, validation.Message)
	}
	findings = append(findings, validation.Details...)
	findings = append(findings, validation.EvidenceQuality.Missing...)
	return findings
}

func traceabilityResultRelatesTo(validation models.ValidationResult, primary string) bool {
	for _, detail := range validation.Details {
		if categoryFromTraceabilityDetail(detail) == primary {
			return true
		}
	}
	for _, finding := range validation.Findings {
		if categoryFromTraceabilityDetail(finding.Message) == primary {
			return true
		}
	}
	for _, impact := range validation.EvidenceQuality.ScoreImpacts {
		if categoryFromTraceabilityDetail(impact.Reason) == primary {
			return true
		}
	}
	return false
}

func firstUnique(values []string, limit int) []string {
	seen := map[string]bool{}
	unique := make([]string, 0, limit)
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		unique = append(unique, value)
		if len(unique) == limit {
			break
		}
	}
	return unique
}

func Confidence(result models.EngineResult) (string, string) {
	meaningful := 0
	for _, category := range categoryOrder {
		if validation, ok := validationForCapability(result, category); ok && hasMeaningfulEvidence(validation) {
			meaningful++
		}
	}

	traceabilityStatus := traceabilityStatus(result)
	stale := hasStaleEvidence(result)

	confidence := ConfidenceLow
	if meaningful >= 4 {
		confidence = ConfidenceMedium
	}
	if meaningful == len(categoryOrder) {
		confidence = ConfidenceHigh
	}
	if traceabilityStatus == models.StatusWarning || traceabilityStatus == models.StatusFail {
		confidence = downgradeConfidence(confidence)
	}
	if stale {
		confidence = downgradeConfidence(confidence)
	}

	return confidence, confidenceReason(meaningful, traceabilityStatus, stale)
}

func validationForCapability(result models.EngineResult, capability string) (models.ValidationResult, bool) {
	for _, validation := range result.Results {
		if validation.Capability == capability {
			return validation, true
		}
	}
	return models.ValidationResult{}, false
}

func hasMeaningfulEvidence(validation models.ValidationResult) bool {
	combined := strings.ToLower(strings.TrimSpace(validation.Message + "\n" + strings.Join(validation.Details, "\n") + "\n" + strings.Join(validation.EvidenceQuality.Details, "\n")))
	if validation.EvidenceQuality.Score > 0 {
		return validation.EvidenceQuality.Score >= strongScoreThreshold
	}
	if validation.Status == models.StatusPass {
		return true
	}
	if strings.Contains(combined, "missing") ||
		strings.Contains(combined, "invalid") ||
		strings.Contains(combined, "empty") ||
		strings.Contains(combined, "required sections missing") ||
		strings.Contains(combined, "placeholder") ||
		strings.Contains(combined, "too thin") ||
		strings.Contains(combined, "header-only") {
		return false
	}
	return len(validation.Details) > 0 || len(validation.Findings) > 0 || validation.Message != ""
}

func traceabilityStatus(result models.EngineResult) string {
	for _, validation := range result.Results {
		if validation.Capability == "Traceability" {
			return validation.Status
		}
	}
	return models.StatusPass
}

func hasStaleEvidence(result models.EngineResult) bool {
	for _, validation := range result.Results {
		combined := strings.ToLower(validation.Message + "\n" + strings.Join(validation.Details, "\n"))
		if strings.Contains(combined, "stale") || strings.Contains(combined, "outdated") || strings.Contains(combined, "expired") {
			return true
		}
	}
	return false
}

func downgradeConfidence(confidence string) string {
	switch confidence {
	case ConfidenceHigh:
		return ConfidenceMedium
	case ConfidenceMedium:
		return ConfidenceLow
	default:
		return ConfidenceLow
	}
}

func confidenceReason(meaningful int, traceabilityStatus string, stale bool) string {
	total := len(categoryOrder)
	if meaningful < 4 {
		return formatEvidenceCountReason("Only", meaningful, total)
	}
	if meaningful < total {
		return formatEvidenceCountReason("", meaningful, total)
	}
	if traceabilityStatus == models.StatusFail {
		return "All 6 evidence categories are present, but traceability has broken references."
	}
	if traceabilityStatus == models.StatusWarning {
		return "All 6 evidence categories are present, but traceability has warnings."
	}
	if stale {
		return "All 6 evidence categories contain meaningful content, but some evidence appears stale."
	}
	return "All 6 evidence categories contain meaningful, connected evidence."
}

func formatEvidenceCountReason(prefix string, count int, total int) string {
	if prefix == "" {
		return strings.TrimSpace(strings.Join([]string{formatInt(count), "of", formatInt(total), "evidence categories contain meaningful content."}, " "))
	}
	return strings.TrimSpace(strings.Join([]string{prefix, formatInt(count), "of", formatInt(total), "evidence categories contain meaningful content."}, " "))
}

func formatInt(value int) string {
	return strconv.Itoa(value)
}

func ScoreValidation(validation models.ValidationResult) int {
	return scoreValidation(validation)
}

func scoreValidation(validation models.ValidationResult) int {
	score := baseScore(validation.Status)
	message := strings.ToLower(validation.Message)
	details := strings.ToLower(strings.Join(validation.Details, "\n"))
	combined := strings.TrimSpace(message + "\n" + details)
	if validation.EvidenceQuality.Score > 0 {
		score = minInt(score, validation.EvidenceQuality.Score)
	}

	if validation.Status == models.StatusPass {
		if validation.EvidenceQuality.Score == 0 {
			score += minInt(15, len(validation.Details)*3)
		}
		if strings.Contains(combined, "non-placeholder") {
			score += 5
		}
		return clampScore(score)
	}

	if strings.Contains(combined, "missing") {
		score -= 15
	}
	if strings.Contains(combined, "invalid") {
		score -= 10
	}
	if strings.Contains(combined, "placeholder") || strings.Contains(combined, "too thin") || strings.Contains(combined, "content quality") {
		score -= 10
	}
	if strings.Contains(combined, "not linked") ||
		strings.Contains(combined, "references missing") ||
		strings.Contains(combined, "cannot reference") ||
		strings.Contains(combined, "orphan") {
		score -= 15
	}
	if strings.Contains(combined, "threshold") ||
		strings.Contains(combined, "violations") ||
		strings.Contains(combined, "low adoption") {
		score -= 5
	}
	if len(validation.Details) == 0 && validation.Status != models.StatusPass {
		score -= 5
	}
	for _, impact := range validation.EvidenceQuality.ScoreImpacts {
		if impact.Delta < 0 && !strings.Contains(combined, strings.ToLower(impact.Reason)) {
			score += impact.Delta
		}
	}

	return clampScore(score)
}

func baseScore(status string) int {
	switch status {
	case models.StatusPass:
		return ScorePass
	case models.StatusWarning:
		return ScoreWarning
	case models.StatusFail:
		return ScoreFail
	default:
		return ScoreUnknown
	}
}

func applyTraceabilityAdjustments(scoreByCategory map[string]int, traceabilityResults []models.ValidationResult) {
	if len(traceabilityResults) == 0 {
		return
	}

	for _, validation := range traceabilityResults {
		if validation.Status == models.StatusPass {
			continue
		}
		penalty := 10
		if validation.Status == models.StatusFail {
			penalty = 20
		}
		for _, detail := range validation.Details {
			category := categoryFromTraceabilityDetail(detail)
			if category == "" {
				continue
			}
			if _, ok := scoreByCategory[category]; ok {
				scoreByCategory[category] = clampScore(scoreByCategory[category] - penalty)
			}
		}
	}

}

func categoryFromTraceabilityDetail(detail string) string {
	lower := strings.ToLower(detail)
	if strings.Contains(lower, "no assurance result references") ||
		strings.Contains(lower, "no mapped test evidence") ||
		strings.Contains(lower, "not linked to assurance evidence") ||
		strings.Contains(lower, "without assurance") {
		return "Assurance"
	}

	pathToCategory := map[string]string{
		"bottleneck/behavior/":  "Behavior",
		"bottleneck/intent/":    "Intent",
		"bottleneck/design/":    "Design",
		"bottleneck/assurance/": "Assurance",
		"bottleneck/security/":  "Security",
		"bottleneck/execution/": "Execution",
	}
	for path, category := range pathToCategory {
		if strings.Contains(lower, path) {
			return category
		}
	}

	idToCategory := map[string]string{
		"BEHAVIOR-":  "Behavior",
		"INTENT-":    "Intent",
		"DESIGN-":    "Design",
		"ASSURANCE-": "Assurance",
		"SECURITY-":  "Security",
		"EXECUTION-": "Execution",
	}
	upper := strings.ToUpper(detail)
	for prefix, category := range idToCategory {
		if strings.Contains(upper, prefix) {
			return category
		}
	}

	return ""
}

func tiedBottlenecks(scores []CategoryScore, lowestScore int) []string {
	candidates := map[string]bool{}
	for _, score := range scores {
		if score.Score == lowestScore {
			candidates[score.Category] = true
		}
	}

	var tied []string
	for _, category := range tiePriorityOrder {
		if candidates[category] {
			tied = append(tied, category)
		}
	}
	return tied
}

func tiedIfMultiple(tied []string) []string {
	if len(tied) < 2 {
		return nil
	}
	return tied
}

func allStrong(scores []CategoryScore) bool {
	for _, score := range scores {
		if score.Score < strongScoreThreshold {
			return false
		}
	}
	return true
}

func reasonFor(validation models.ValidationResult) string {
	if validation.Message != "" {
		return validation.Message
	}
	if validation.Status != "" {
		return strings.ToLower(validation.Status)
	}
	return "not assessed"
}

func missingEvidenceAction(category string) string {
	switch category {
	case "Config":
		return "Repair bottleneck/config.yaml so the selected environment and thresholds can be resolved."
	case "Intent":
		return "Create bottleneck/intent/intent.md with 1-3 measurable outcomes, constraints, and success criteria."
	case "Behavior":
		return "Create bottleneck/behavior/behavior-spec.md with expected behavior, unacceptable behavior, and evidence IDs."
	case "Design":
		return "Create bottleneck/design/architecture.md with the architecture decisions reviewers need to evaluate the change."
	case "Assurance":
		return "Add assurance evidence that maps test or evaluation results to BEHAVIOR-001."
	case "Security":
		return "Add guardrail evidence in bottleneck/security/guardrails.json and confirm violations are zero."
	case "Execution":
		return "Add execution telemetry in bottleneck/execution/telemetry.json for adoption and error rate."
	case "Traceability":
		return "Add evidence IDs and Refs links across intent, behavior, assurance, security, and execution artifacts."
	default:
		return "Add the missing evidence artifact and rerun validation."
	}
}

func weakEvidenceAction(category string) string {
	switch category {
	case "Intent":
		return "Replace placeholder intent statements with 1-3 measurable outcomes."
	case "Behavior":
		return "Replace placeholder behavior text with concrete expected and unacceptable behavior."
	case "Design":
		return "Replace placeholder architecture notes with the decisions, dependencies, and operating constraints reviewers need."
	case "Assurance":
		return "Refresh assurance evidence with current test or evaluation results mapped to behavior IDs."
	case "Security":
		return "Refresh guardrail evidence with current policy scan results and zero unresolved violations."
	case "Execution":
		return "Refresh execution telemetry with current adoption and reliability evidence."
	default:
		return "Replace weak evidence with concrete, reviewable detail."
	}
}

func staleEvidenceAction(category string) string {
	switch category {
	case "Assurance":
		return "Regenerate assurance evidence from the latest test or evaluation run."
	case "Security":
		return "Regenerate security evidence from the latest guardrail or policy scan."
	case "Execution":
		return "Refresh execution telemetry from the current environment before release."
	default:
		return "Refresh the stale evidence artifact and rerun validation."
	}
}

func disconnectedEvidenceAction(category string) string {
	switch category {
	case "Intent":
		return "Link intent evidence to the behavior IDs it is meant to justify."
	case "Behavior":
		return "Link BEHAVIOR-001 to its supporting INTENT and ASSURANCE evidence."
	case "Design":
		return "Link design evidence to the intent and behavior it implements."
	case "Assurance":
		return "Link assurance results to the behavior IDs they verify."
	case "Security":
		return "Link security evidence to the behavior or intent evidence it protects."
	case "Execution":
		return "Link execution telemetry to the behavior or assurance evidence it validates in production."
	default:
		return "Repair missing, orphaned, or disconnected evidence links."
	}
}

func thresholdAction(category string) string {
	switch category {
	case "Assurance":
		return "Fix failing tests or add passing assurance evidence until accuracy meets the selected threshold."
	case "Security":
		return "Resolve guardrail violations before treating the release as valid."
	case "Execution":
		return "Improve adoption or reliability until telemetry meets the selected execution thresholds."
	default:
		return "Resolve the failed threshold and rerun validation."
	}
}

func isDiagnosableCategory(category string) bool {
	for _, candidate := range categoryOrder {
		if category == candidate {
			return true
		}
	}
	return false
}

func clampScore(score int) int {
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func SortScores(scores []CategoryScore) []CategoryScore {
	sorted := append([]CategoryScore{}, scores...)
	sort.SliceStable(sorted, func(i int, j int) bool {
		if sorted[i].Score != sorted[j].Score {
			return sorted[i].Score < sorted[j].Score
		}
		return indexOfPriority(sorted[i].Category) < indexOfPriority(sorted[j].Category)
	})
	return sorted
}

func indexOfPriority(category string) int {
	for index, candidate := range tiePriorityOrder {
		if candidate == category {
			return index
		}
	}
	return len(tiePriorityOrder)
}

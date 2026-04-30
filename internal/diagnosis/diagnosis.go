package diagnosis

import (
	"sort"
	"strings"

	"bottleneck/internal/models"
)

const (
	HealthyPrimaryBottleneck = "None"

	ScorePass    = 85
	ScoreWarning = 60
	ScoreFail    = 20
	ScoreUnknown = 40

	strongScoreThreshold = 80
)

type Diagnosis struct {
	PrimaryBottleneck string          `json:"primary_bottleneck"`
	TiedBottlenecks   []string        `json:"tied_bottlenecks,omitempty"`
	WhyItMatters      string          `json:"why_it_matters"`
	RecommendedAction string          `json:"recommended_action"`
	CategoryScores    []CategoryScore `json:"category_scores"`
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
}

func Analyze(result models.EngineResult) Diagnosis {
	resultsByCategory := map[string]models.ValidationResult{}
	traceabilityResults := make([]models.ValidationResult, 0, 1)
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

	if len(scores) == 0 || allStrong(scores) {
		return Diagnosis{
			PrimaryBottleneck: HealthyPrimaryBottleneck,
			WhyItMatters:      "All assessed delivery categories have enough evidence to support the current release decision.",
			RecommendedAction: "Keep evidence current as intent, behavior, implementation, and production signals change.",
			CategoryScores:    scores,
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
	validation := resultsByCategory[primary]

	return Diagnosis{
		PrimaryBottleneck: primary,
		TiedBottlenecks:   tiedIfMultiple(tied),
		WhyItMatters:      WhyItMatters(primary),
		RecommendedAction: RecommendedAction(validation),
		CategoryScores:    scores,
	}
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
	case strings.Contains(combined, "not linked") ||
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

func ScoreValidation(validation models.ValidationResult) int {
	return scoreValidation(validation)
}

func scoreValidation(validation models.ValidationResult) int {
	score := baseScore(validation.Status)
	message := strings.ToLower(validation.Message)
	details := strings.ToLower(strings.Join(validation.Details, "\n"))
	combined := strings.TrimSpace(message + "\n" + details)

	if validation.Status == models.StatusPass {
		score += minInt(15, len(validation.Details)*3)
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
		return "Link security evidence to the behavior or assurance evidence it protects."
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

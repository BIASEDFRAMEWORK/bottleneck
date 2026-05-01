package diagnosis

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"bottleneck/internal/models"
)

const (
	ruleHealthy                      = "healthy"
	ruleCriticalSecurityBlocker      = "critical_security_blocker"
	ruleMissingIntent                = "missing_intent"
	ruleBehaviorNotMappedToIntent    = "behavior_not_mapped_to_intent"
	ruleCriticalBehaviorWithoutTests = "critical_behavior_without_tests"
	ruleBrokenTraceability           = "broken_traceability"
	ruleMissingProductionReadiness   = "missing_production_readiness"
	ruleStaleTelemetry               = "stale_telemetry"
	ruleThinPlaceholderEvidence      = "thin_placeholder_evidence"
	ruleLowTraceabilityConfidence    = "low_traceability_confidence"
	ruleGenericPrimary               = "primary_bottleneck"
)

var evidenceIDPattern = regexp.MustCompile(`\b(?:INTENT|BEHAVIOR|DESIGN|ASSURANCE|SECURITY|EXECUTION)-[0-9]{3,}\b`)

type actionableDiagnosis struct {
	Rule                string
	Category            string
	Priority            int
	Reason              string
	Impact              string
	NextAction          string
	InspectCommand      string
	RelevantEvidenceIDs []string
	SupportingIssues    []string
}

func primaryFromActionableRules(result models.EngineResult, fallback string) string {
	rule, ok := highestActionableRule(result)
	if !ok {
		return fallback
	}
	if rule.Priority < 70 {
		return fallback
	}
	return rule.Category
}

func actionableFor(result models.EngineResult, primary string, validation models.ValidationResult, traceabilityResults []models.ValidationResult, findings []string) actionableDiagnosis {
	if primary == HealthyPrimaryBottleneck {
		return actionableDiagnosis{
			Rule:                ruleHealthy,
			Category:            primary,
			Reason:              "All assessed delivery evidence is strong enough for the current release decision.",
			Impact:              "Release confidence is high enough to proceed while keeping evidence current.",
			NextAction:          "Keep evidence current as intent, behavior, implementation, and production signals change.",
			InspectCommand:      "bottleneck scorecard --details",
			RelevantEvidenceIDs: evidenceIDsFromStrings(findings),
			SupportingIssues:    supportingIssues(result, primary),
		}
	}

	if rule, ok := highestActionableRuleForPrimary(result, primary); ok {
		rule.SupportingIssues = supportingIssues(result, primary)
		return rule
	}

	combined := strings.Join(append([]string{validation.Message}, append(validation.Details, findings...)...), "\n")
	ids := evidenceIDsFromStrings([]string{combined})
	return actionableDiagnosis{
		Rule:                ruleGenericPrimary,
		Category:            primary,
		Reason:              fallbackReason(primary, validation, findings),
		Impact:              WhyItMatters(primary),
		NextAction:          RecommendedAction(validation),
		InspectCommand:      inspectCommandFor(primary, ids),
		RelevantEvidenceIDs: ids,
		SupportingIssues:    supportingIssues(result, primary),
	}
}

func highestActionableRuleForPrimary(result models.EngineResult, primary string) (actionableDiagnosis, bool) {
	rules := actionableRules(result)
	filtered := make([]actionableDiagnosis, 0, len(rules))
	for _, rule := range rules {
		if rule.Category == primary {
			filtered = append(filtered, rule)
		}
	}
	return highestRule(filtered)
}

func highestActionableRule(result models.EngineResult) (actionableDiagnosis, bool) {
	return highestRule(actionableRules(result))
}

func highestRule(rules []actionableDiagnosis) (actionableDiagnosis, bool) {
	if len(rules) == 0 {
		return actionableDiagnosis{}, false
	}
	sort.SliceStable(rules, func(i, j int) bool {
		return rules[i].Priority > rules[j].Priority
	})
	return rules[0], true
}

func actionableRules(result models.EngineResult) []actionableDiagnosis {
	var rules []actionableDiagnosis
	for _, validation := range result.Results {
		combined := validationCombinedText(validation)
		lower := strings.ToLower(combined)
		ids := evidenceIDsFromStrings([]string{combined})

		switch validation.Capability {
		case "Security":
			if validation.Status == models.StatusFail && criticalSecuritySignal(lower) {
				rules = append(rules, actionableDiagnosis{
					Rule:                ruleCriticalSecurityBlocker,
					Category:            "Security",
					Priority:            100,
					Reason:              "Critical security findings are present.",
					Impact:              "Production release should not proceed while critical payment or account-security risk remains open.",
					NextAction:          "Resolve critical findings or add accepted-risk governance evidence.",
					InspectCommand:      "bottleneck scorecard --details",
					RelevantEvidenceIDs: ids,
				})
			}
		case "Intent":
			if validation.Status == models.StatusFail && missingIntentSignal(lower) {
				rules = append(rules, actionableDiagnosis{
					Rule:                ruleMissingIntent,
					Category:            "Intent",
					Priority:            90,
					Reason:              "No intent evidence describes the customer outcome.",
					Impact:              "The team cannot tell what release risk the evidence is meant to reduce.",
					NextAction:          "Add intent evidence with measurable SaaS outcome and related behavior IDs.",
					InspectCommand:      "bottleneck validate",
					RelevantEvidenceIDs: ids,
				})
			}
		case "Execution":
			if validation.Status == models.StatusFail && strings.Contains(lower, "missing telemetry") {
				rules = append(rules, actionableDiagnosis{
					Rule:                ruleMissingProductionReadiness,
					Category:            "Execution",
					Priority:            60,
					Reason:              "Production readiness evidence is missing.",
					Impact:              "The team cannot confirm current reliability, adoption, or billing telemetry before release.",
					NextAction:          "Add execution telemetry or ingest the latest production readiness metrics.",
					InspectCommand:      "bottleneck scorecard --details",
					RelevantEvidenceIDs: ids,
				})
			}
			if strings.Contains(lower, "stale") || strings.Contains(lower, "outdated") || strings.Contains(lower, "expired") {
				rules = append(rules, actionableDiagnosis{
					Rule:                ruleStaleTelemetry,
					Category:            "Execution",
					Priority:            50,
					Reason:              "Execution telemetry is stale.",
					Impact:              "Release readiness is based on old production behavior and may miss current billing failures.",
					NextAction:          "Refresh telemetry evidence or ingest the latest execution metrics.",
					InspectCommand:      inspectCommandFor("Execution", ids),
					RelevantEvidenceIDs: ids,
				})
			}
		}

		if validation.Status != models.StatusPass && weakEvidenceSignal(lower) {
			rules = append(rules, actionableDiagnosis{
				Rule:                ruleThinPlaceholderEvidence,
				Category:            validation.Capability,
				Priority:            40,
				Reason:              fmt.Sprintf("%s evidence is too thin or placeholder-heavy.", validation.Capability),
				Impact:              "Reviewers cannot trust the release decision until the evidence describes real system behavior.",
				NextAction:          RecommendedAction(validation),
				InspectCommand:      inspectCommandFor(validation.Capability, ids),
				RelevantEvidenceIDs: ids,
			})
		}

		if validation.Capability == "Traceability" && validation.Status != models.StatusPass {
			rules = append(rules, traceabilityActionableRules(result, validation)...)
		}
	}
	return rules
}

func traceabilityActionableRules(result models.EngineResult, validation models.ValidationResult) []actionableDiagnosis {
	var rules []actionableDiagnosis
	for _, detail := range validation.Details {
		lower := strings.ToLower(detail)
		ids := evidenceIDsFromStrings([]string{detail})
		behaviorID := firstEvidenceIDWithPrefix(ids, "BEHAVIOR-")
		if behaviorID == "" {
			behaviorID = firstEvidenceIDWithPrefix(evidenceIDsFromStrings([]string{validationCombinedText(validation)}), "BEHAVIOR-")
		}

		if strings.Contains(lower, "no mapped test evidence") ||
			strings.Contains(lower, "not linked to assurance evidence") ||
			strings.Contains(lower, "no assurance result references") {
			rules = append(rules, assuranceCoverageRule(result.Environment, detail, behaviorID, ids))
			continue
		}

		if strings.Contains(lower, "not linked to intent evidence") {
			rules = append(rules, actionableDiagnosis{
				Rule:                ruleBehaviorNotMappedToIntent,
				Category:            "Behavior",
				Priority:            85,
				Reason:              behaviorIDReason(behaviorID, "is not linked to intent evidence."),
				Impact:              "The behavior is not traceable to a customer or release outcome.",
				NextAction:          behaviorIDAction(behaviorID, "Add an intent reference to the behavior or update the intent evidence."),
				InspectCommand:      inspectCommandFor("Behavior", ids),
				RelevantEvidenceIDs: ids,
			})
			continue
		}

		if validation.Status == models.StatusFail {
			category := categoryFromTraceabilityDetail(detail)
			if category == "" {
				category = "Behavior"
			}
			rules = append(rules, actionableDiagnosis{
				Rule:                ruleBrokenTraceability,
				Category:            category,
				Priority:            70,
				Reason:              detail,
				Impact:              "Required release evidence cannot be audited end to end until the broken reference is fixed.",
				NextAction:          "Repair the broken evidence reference and rerun diagnosis.",
				InspectCommand:      inspectCommandFor(category, ids),
				RelevantEvidenceIDs: ids,
			})
		}
	}

	if len(rules) == 0 && validation.Status == models.StatusWarning {
		rules = append(rules, actionableDiagnosis{
			Rule:                ruleLowTraceabilityConfidence,
			Category:            "Traceability",
			Priority:            35,
			Reason:              "Traceability confidence is low.",
			Impact:              "Release evidence is harder to audit because some artifacts are not connected.",
			NextAction:          "Repair missing, duplicate, orphaned, or weak evidence links.",
			InspectCommand:      "bottleneck scorecard --details",
			RelevantEvidenceIDs: evidenceIDsFromStrings([]string{validationCombinedText(validation)}),
		})
	}

	return rules
}

func assuranceCoverageRule(environment string, detail string, behaviorID string, ids []string) actionableDiagnosis {
	if behaviorID == "" {
		behaviorID = "BEHAVIOR evidence"
	}
	reason := behaviorID + " is not linked to any passing test evidence."
	impact := "Release confidence is reduced because the behavior is unproven."
	if strings.EqualFold(environment, "production") {
		impact = "Production release should not proceed because required behavior is unproven."
	}
	if isPaymentRetrySignal(detail, behaviorID) {
		impact = "Release confidence is reduced because payment retry behavior is unproven."
		if strings.EqualFold(environment, "production") {
			impact = "Production release should not proceed because payment retry behavior is unproven."
		}
	}
	return actionableDiagnosis{
		Rule:                ruleCriticalBehaviorWithoutTests,
		Category:            "Assurance",
		Priority:            80,
		Reason:              reason,
		Impact:              impact,
		NextAction:          "Add or ingest test evidence mapped to " + behaviorID + ".",
		InspectCommand:      "bottleneck trace " + behaviorID,
		RelevantEvidenceIDs: ids,
	}
}

func validationCombinedText(validation models.ValidationResult) string {
	parts := []string{validation.Message}
	parts = append(parts, validation.Details...)
	parts = append(parts, validation.EvidenceQuality.Details...)
	parts = append(parts, validation.EvidenceQuality.Missing...)
	for _, finding := range validation.Findings {
		parts = append(parts, finding.Message)
	}
	for _, impact := range validation.EvidenceQuality.ScoreImpacts {
		parts = append(parts, impact.Reason)
	}
	return strings.Join(parts, "\n")
}

func criticalSecuritySignal(lower string) bool {
	if strings.Contains(lower, "violations detected") || strings.Contains(lower, "critical security") {
		return true
	}
	for _, line := range strings.Split(lower, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "critical_findings:") {
			fields := strings.Fields(strings.TrimPrefix(line, "critical_findings:"))
			if len(fields) > 0 {
				value, err := strconv.Atoi(fields[0])
				return err == nil && value > 0
			}
		}
	}
	return false
}

func missingIntentSignal(lower string) bool {
	return strings.Contains(lower, "missing intent.md") ||
		strings.Contains(lower, "missing intent") ||
		strings.Contains(lower, "no intent evidence")
}

func weakEvidenceSignal(lower string) bool {
	return strings.Contains(lower, "placeholder") ||
		strings.Contains(lower, "too thin") ||
		strings.Contains(lower, "content quality") ||
		strings.Contains(lower, "required sections missing")
}

func evidenceIDsFromStrings(values []string) []string {
	seen := map[string]bool{}
	var ids []string
	for _, value := range values {
		for _, id := range evidenceIDPattern.FindAllString(value, -1) {
			if seen[id] {
				continue
			}
			seen[id] = true
			ids = append(ids, id)
		}
	}
	return ids
}

func firstEvidenceIDWithPrefix(ids []string, prefix string) string {
	for _, id := range ids {
		if strings.HasPrefix(id, prefix) {
			return id
		}
	}
	return ""
}

func fallbackReason(primary string, validation models.ValidationResult, findings []string) string {
	if len(findings) > 0 {
		return findings[0]
	}
	if validation.Message != "" {
		return validation.Message
	}
	return Info(primary).Bottleneck
}

func inspectCommandFor(category string, ids []string) string {
	for _, prefix := range []string{"BEHAVIOR-", "INTENT-", "ASSURANCE-", "SECURITY-", "EXECUTION-"} {
		if id := firstEvidenceIDWithPrefix(ids, prefix); id != "" {
			return "bottleneck trace " + id
		}
	}
	switch category {
	case "Intent":
		return "bottleneck validate"
	default:
		return "bottleneck scorecard --details"
	}
}

func behaviorIDReason(behaviorID string, suffix string) string {
	if behaviorID == "" {
		return "Behavior evidence " + suffix
	}
	return behaviorID + " " + suffix
}

func behaviorIDAction(behaviorID string, fallback string) string {
	if behaviorID == "" {
		return fallback
	}
	return "Add an intent reference to " + behaviorID + " or update the intent evidence."
}

func isPaymentRetrySignal(detail string, behaviorID string) bool {
	lower := strings.ToLower(detail + " " + behaviorID)
	return strings.Contains(lower, "behavior-003") ||
		strings.Contains(lower, "payment retry") ||
		strings.Contains(lower, "duplicate-charge") ||
		strings.Contains(lower, "duplicate charges")
}

func supportingIssues(result models.EngineResult, primary string) []string {
	type issue struct {
		priority int
		text     string
	}
	var issues []issue
	for _, validation := range result.Results {
		if validation.Status == models.StatusPass {
			continue
		}
		text := supportIssueText(validation)
		if text == "" {
			continue
		}
		category := validation.Capability
		if validation.Capability == "Traceability" {
			category = categoryFromTraceabilityDetail(text)
			if category == "" {
				category = "Traceability"
			}
		}
		if category == primary {
			continue
		}
		issues = append(issues, issue{priority: issuePriority(validation.Status), text: text})
	}
	sort.SliceStable(issues, func(i, j int) bool {
		return issues[i].priority > issues[j].priority
	})
	values := make([]string, 0, len(issues))
	for _, issue := range issues {
		values = append(values, issue.text)
	}
	return firstUnique(values, 5)
}

func supportIssueText(validation models.ValidationResult) string {
	if len(validation.Details) > 0 {
		return validation.Capability + ": " + validation.Details[0]
	}
	if validation.Message != "" {
		return validation.Capability + ": " + validation.Message
	}
	if len(validation.EvidenceQuality.Missing) > 0 {
		return validation.Capability + ": " + validation.EvidenceQuality.Missing[0]
	}
	return ""
}

func issuePriority(status string) int {
	switch status {
	case models.StatusFail:
		return 2
	case models.StatusWarning:
		return 1
	default:
		return 0
	}
}

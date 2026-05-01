package explainer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"bottleneck/internal/diagnosis"
	"bottleneck/internal/models"
	"bottleneck/internal/traceability"
)

type capabilityMetadata struct {
	owner                 string
	ownerRoles            []string
	bottleneck            string
	whyThisMatters        string
	riskToDelivery        string
	nextActionBase        []string
	automationSuggestions []string
}

const (
	SchemaVersion  = "explain.v2"
	FormatText     = "text"
	FormatMarkdown = "markdown"
	FormatJSON     = "json"
)

type Report struct {
	SchemaVersion     string                `json:"schema_version"`
	Environment       string                `json:"environment"`
	SystemStatus      string                `json:"system_status"`
	PrimaryBottleneck string                `json:"primary_bottleneck"`
	Diagnosis         diagnosis.Diagnosis   `json:"diagnosis"`
	Explanations      []CategoryExplanation `json:"explanations"`
}

type CategoryExplanation struct {
	Category             string               `json:"category"`
	Score                int                  `json:"score"`
	Status               string               `json:"status"`
	WhyThisMatters       string               `json:"why_this_matters"`
	EvidenceFound        []EvidenceFact       `json:"evidence_found"`
	EvidenceMissing      []EvidenceGap        `json:"evidence_missing"`
	RelatedIDs           []string             `json:"related_ids"`
	ScoreImpacts         []models.ScoreImpact `json:"score_impacts"`
	RiskToDelivery       string               `json:"risk_to_delivery"`
	Recommendation       string               `json:"recommendation"`
	RecommendedActions   []string             `json:"recommended_actions"`
	SuggestedOwnerRoles  []string             `json:"suggested_owner_roles"`
	SuggestedAutomations []string             `json:"suggested_automations"`
	Owner                string               `json:"owner"`
	MappedBottleneck     string               `json:"mapped_bottleneck"`
}

type EvidenceFact struct {
	Text string `json:"text"`
}

type EvidenceGap struct {
	Text string `json:"text"`
}

var evidenceIDPattern = regexp.MustCompile(`\b(INTENT|BEHAVIOR|DESIGN|ASSURANCE|SECURITY|EXECUTION)-[0-9]{3,}\b`)

var metadataByCapability = map[string]capabilityMetadata{
	"Intent": {
		owner:          "Intent Engineer",
		ownerRoles:     []string{"Product Lead", "Domain Expert", "Technical Lead"},
		bottleneck:     "Ambiguous requirements",
		whyThisMatters: "Intent evidence defines the business outcome, constraints, and success criteria that downstream behavior, design, assurance, security, and execution evidence must support.",
		riskToDelivery: "The team may build or approve work without a shared, measurable definition of the intended outcome.",
		nextActionBase: []string{
			"Update bottleneck/intent/intent.md with concrete outcomes, constraints, success criteria, and INTENT-* refs.",
		},
		automationSuggestions: []string{
			"Add a PR template requiring an intent reference.",
			"Run Markdown quality checks for measurable outcomes and constraints.",
			"Add a commit hook or CI check for required INTENT-* IDs.",
		},
	},
	"Behavior": {
		owner:          "Behavior Engineer",
		ownerRoles:     []string{"Product Lead", "Domain Expert", "QA/Assurance Engineer"},
		bottleneck:     "Non-deterministic outputs",
		whyThisMatters: "Behavior evidence describes what the system must and must not do so tests and release decisions can be tied to observable expectations.",
		riskToDelivery: "The team may ship behavior that appears complete but is not traceable to expected and unacceptable outcomes.",
		nextActionBase: []string{
			"Update bottleneck/behavior/behavior-spec.md with concrete BEHAVIOR-* expectations and refs.",
		},
		automationSuggestions: []string{
			"Run behavior specification linting in GitHub Actions.",
			"Check behavior IDs and traceability refs before merging.",
			"Fail release gates when critical behavior lacks mapped evidence.",
		},
	},
	"Design": {
		owner:          "Design Engineer",
		ownerRoles:     []string{"Architect", "Technical Lead", "Platform Engineer"},
		bottleneck:     "Poor adoption / UX gaps",
		whyThisMatters: "Design evidence explains architecture, boundaries, dependencies, tradeoffs, and operational assumptions before teams rely on the implementation.",
		riskToDelivery: "The team may approve a change without understanding system boundaries, tradeoffs, failure modes, or operational risk.",
		nextActionBase: []string{
			"Update bottleneck/design/architecture.md with DESIGN-* evidence that references the relevant intent and behavior IDs.",
		},
		automationSuggestions: []string{
			"Require architecture decision record checks for major changes.",
			"Run diagram or architecture document freshness checks.",
			"Gate production approval on current design evidence for risky changes.",
		},
	},
	"Assurance": {
		owner:          "Assurance Engineer",
		ownerRoles:     []string{"QA/Assurance Engineer", "Developer", "Product Lead"},
		bottleneck:     "Validation gaps",
		whyThisMatters: "Assurance evidence proves that intended behavior was tested before release and that critical behavior is mapped to passing validation evidence.",
		riskToDelivery: "The team may ship functionality that appears complete but cannot be proven against intended behavior.",
		nextActionBase: []string{
			"Add Cucumber, evaluation, or test evidence in bottleneck/assurance/results.json that references the affected BEHAVIOR-* IDs.",
		},
		automationSuggestions: []string{
			"Run Cucumber or equivalent behavior tests in GitHub Actions.",
			"Ingest test results into bottleneck/assurance/results.json.",
			"Fail the release gate when critical behaviors lack mapped tests.",
		},
	},
	"Security": {
		owner:          "Security Engineer",
		ownerRoles:     []string{"Security Engineer", "Platform Engineer", "Technical Lead"},
		bottleneck:     "Risk & compliance",
		whyThisMatters: "Security evidence shows known risks, guardrails, and policy findings before release acceleration hides unresolved vulnerabilities or compliance gaps.",
		riskToDelivery: "The team may ship known vulnerabilities, missing guardrails, or unreviewed compliance risk.",
		nextActionBase: []string{
			"Regenerate bottleneck/security/guardrails.json from SARIF or guardrail evidence and resolve findings above threshold.",
		},
		automationSuggestions: []string{
			"Run CodeQL in GitHub Actions.",
			"Enable dependency review and secret scanning.",
			"Ingest SARIF into bottleneck/security/guardrails.json.",
		},
	},
	"Execution": {
		owner:          "Execution Engineer",
		ownerRoles:     []string{"SRE/Operations", "Product Lead", "Customer Success/Adoption Lead"},
		bottleneck:     "Delivery friction",
		whyThisMatters: "Execution evidence confirms whether the system is reliable, adopted, and healthy in delivery or production-like use.",
		riskToDelivery: "The team may accelerate release without knowing whether reliability, adoption, or operational signals support the decision.",
		nextActionBase: []string{
			"Refresh bottleneck/execution/telemetry.json with current reliability, adoption, override, deployment, and cost signals.",
		},
		automationSuggestions: []string{
			"Ingest telemetry JSON from release health checks.",
			"Run production signal review before release approval.",
			"Fail or warn release gates when reliability or adoption metrics fall below threshold.",
		},
	},
	"Traceability": {
		owner:          "Release Engineer",
		ownerRoles:     []string{"Release Engineer", "Technical Lead"},
		bottleneck:     "Traceability gaps",
		whyThisMatters: "Traceability evidence connects intent, behavior, tests, security, and execution signals so release decisions can be audited.",
		riskToDelivery: "The team may make release decisions from disconnected artifacts that cannot prove which behavior is covered.",
		nextActionBase: []string{
			"Run bottleneck trace --id <ID> for the affected evidence ID and repair missing, duplicate, or orphaned links.",
		},
		automationSuggestions: []string{
			"Run bottleneck validate in GitHub Actions.",
			"Fail release gates on broken evidence references.",
		},
	},
	"Config": {
		owner:          "Execution Engineer",
		ownerRoles:     []string{"Technical Lead", "Platform Engineer"},
		bottleneck:     "Delivery friction",
		whyThisMatters: "Configuration controls how evidence thresholds are resolved for each environment.",
		riskToDelivery: "The team may trust a release decision based on invalid or unknown environment thresholds.",
		nextActionBase: []string{
			"Repair bottleneck/config.yaml so the selected environment and thresholds resolve cleanly.",
		},
		automationSuggestions: []string{
			"Validate bottleneck/config.yaml in CI.",
			"Use environment-specific release gate checks.",
		},
	},
}

func Render(result models.EngineResult, capabilityFilter string) (string, error) {
	return RenderWithGraph(result, nil, capabilityFilter)
}

func RenderWithGraph(result models.EngineResult, graph *traceability.Graph, capabilityFilter string) (string, error) {
	return RenderWithOptions(result, graph, Options{Filter: capabilityFilter, Format: FormatText})
}

type Options struct {
	Filter string
	Format string
}

func RenderWithOptions(result models.EngineResult, graph *traceability.Graph, options Options) (string, error) {
	report, err := BuildReport(result, graph, options.Filter)
	if err != nil {
		return "", err
	}
	return RenderReport(report, options.Format)
}

func BuildReport(result models.EngineResult, graph *traceability.Graph, capabilityFilter string) (Report, error) {
	filtered, err := filterResults(result.Results, capabilityFilter)
	if err != nil {
		return Report{}, err
	}
	diagnosisResult := diagnosis.Analyze(result)
	explanations := categoryExplanations(result, filtered, diagnosisResult.CategoryScores, graph)

	return Report{
		SchemaVersion:     SchemaVersion,
		Environment:       result.Environment,
		SystemStatus:      result.SystemStatus,
		PrimaryBottleneck: diagnosisResult.PrimaryBottleneck,
		Diagnosis:         diagnosisResult,
		Explanations:      explanations,
	}, nil
}

func RenderReport(report Report, format string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "", FormatText:
		return renderText(report), nil
	case FormatMarkdown:
		return renderMarkdown(report), nil
	case FormatJSON:
		return renderJSON(report)
	default:
		return "", fmt.Errorf("unsupported format %q (supported: text, markdown, json)", format)
	}
}

func WriteOutput(path string, output string) error {
	if strings.TrimSpace(path) == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(output), 0o644)
}

func renderText(report Report) string {
	var lines []string
	lines = append(lines,
		"Bottleneck Explanation",
		"",
		fmt.Sprintf("Environment: %s", report.Environment),
		fmt.Sprintf("System Status: %s", report.SystemStatus),
		fmt.Sprintf("Primary bottleneck: %s", report.PrimaryBottleneck),
		fmt.Sprintf("Primary Bottleneck: %s", report.PrimaryBottleneck),
	)
	if len(report.Explanations) > 1 {
		lines = append(lines,
			"",
			"Primary Diagnosis:",
			fmt.Sprintf("Weakest Category: %s", report.PrimaryBottleneck),
			"Top Evidence:",
		)
		lines = appendBulletLines(lines, report.Diagnosis.ContributingFindings, "None.")
		lines = append(lines, fmt.Sprintf("Next Action: %s", report.Diagnosis.RecommendedAction))
		if len(report.Diagnosis.TiedBottlenecks) > 0 {
			lines = append(lines, fmt.Sprintf("Tied Bottlenecks: %s", strings.Join(report.Diagnosis.TiedBottlenecks, ", ")))
		}
	}

	for _, explanation := range report.Explanations {
		lines = append(lines, "")
		lines = append(lines,
			fmt.Sprintf("%s Score: %d", explanation.Category, explanation.Score),
			fmt.Sprintf("Status: %s", explanation.Status),
			fmt.Sprintf("Owner: %s", explanation.Owner),
			fmt.Sprintf("Mapped Bottleneck: %s", explanation.MappedBottleneck),
			"",
			"Why this matters:",
			explanation.WhyThisMatters,
			"",
			"Evidence found:",
		)
		lines = appendFactLines(lines, explanation.EvidenceFound)
		lines = append(lines, "", "Evidence missing:")
		lines = appendGapLines(lines, explanation.EvidenceMissing)
		lines = append(lines, "", "Related IDs:")
		lines = appendBulletLines(lines, explanation.RelatedIDs, "None found.")
		lines = append(lines, "", "Score impact:")
		lines = appendImpactLines(lines, explanation.ScoreImpacts)
		lines = append(lines, "", "Recommendation:", explanation.Recommendation)
		lines = append(lines, "", "Risk to delivery:", explanation.RiskToDelivery)
		lines = append(lines, "", "Recommended actions:")
		lines = appendNumberedLines(lines, explanation.RecommendedActions)
		lines = append(lines, "", "Suggested owner roles:")
		lines = appendBulletLines(lines, explanation.SuggestedOwnerRoles, "None.")
		lines = append(lines, "", "Suggested automation:")
		lines = appendBulletLines(lines, explanation.SuggestedAutomations, "None.")
	}

	return strings.Join(lines, "\n")
}

func renderMarkdown(report Report) string {
	lines := []string{
		"# Bottleneck Explanation",
		"",
		fmt.Sprintf("- Environment: `%s`", report.Environment),
		fmt.Sprintf("- System status: `%s`", report.SystemStatus),
		fmt.Sprintf("- Primary bottleneck: `%s`", report.PrimaryBottleneck),
		"",
		"## Executive Diagnosis",
		report.Diagnosis.WhyItMatters,
		"",
	}
	if len(report.Diagnosis.ContributingFindings) > 0 {
		lines = append(lines, "### Top Evidence")
		lines = appendBulletLines(lines, report.Diagnosis.ContributingFindings, "None.")
		lines = append(lines, "")
	}
	for _, explanation := range report.Explanations {
		lines = append(lines,
			fmt.Sprintf("## %s", explanation.Category),
			"",
			fmt.Sprintf("- Status: `%s`", explanation.Status),
			fmt.Sprintf("- Score: `%d`", explanation.Score),
			fmt.Sprintf("- Mapped bottleneck: `%s`", explanation.MappedBottleneck),
			"",
			"### Why This Matters",
			explanation.WhyThisMatters,
			"",
			"### Evidence Found",
		)
		lines = appendFactLines(lines, explanation.EvidenceFound)
		lines = append(lines, "", "### Evidence Missing")
		lines = appendGapLines(lines, explanation.EvidenceMissing)
		lines = append(lines, "", "### Risk To Delivery", explanation.RiskToDelivery)
		lines = append(lines, "", "### Recommended Actions")
		lines = appendNumberedLines(lines, explanation.RecommendedActions)
		lines = append(lines, "", "### Suggested Owner Roles")
		lines = appendBulletLines(lines, explanation.SuggestedOwnerRoles, "None.")
		lines = append(lines, "", "### Suggested Automation")
		lines = appendBulletLines(lines, explanation.SuggestedAutomations, "None.")
		lines = append(lines, "")
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func renderJSON(report Report) (string, error) {
	content, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func categoryExplanations(result models.EngineResult, filtered []models.ValidationResult, scores []diagnosis.CategoryScore, graph *traceability.Graph) []CategoryExplanation {
	allResults := map[string]models.ValidationResult{}
	for _, validation := range result.Results {
		allResults[validation.Capability] = validation
	}

	explanations := make([]CategoryExplanation, 0, len(filtered))
	for _, validation := range filtered {
		explanations = append(explanations, explainCategory(validation, scores, graph, allResults))
	}
	return explanations
}

func explainCategory(validation models.ValidationResult, scores []diagnosis.CategoryScore, graph *traceability.Graph, allResults map[string]models.ValidationResult) CategoryExplanation {
	category := validation.Capability
	meta := metadataFor(category)
	found := evidenceFound(validation, graph)
	missing := evidenceMissing(validation, graph, allResults)
	relatedIDs := relatedIDsFor(category, validation, graph)
	impacts := scoreImpactsFor(validation)
	recommendation := diagnosis.RecommendedAction(validation)
	if recommendation == "" && len(meta.nextActionBase) > 0 {
		recommendation = meta.nextActionBase[0]
	}
	recommendedActions := recommendedActionsFor(validation, missing, recommendation, meta)

	return CategoryExplanation{
		Category:             category,
		Score:                diagnosis.ScoreFor(category, scores),
		Status:               validation.Status,
		WhyThisMatters:       meta.whyThisMatters,
		EvidenceFound:        factsFromStrings(found),
		EvidenceMissing:      gapsFromStrings(missing),
		RelatedIDs:           relatedIDs,
		ScoreImpacts:         impacts,
		RiskToDelivery:       riskToDeliveryFor(validation, missing, meta),
		Recommendation:       recommendation,
		RecommendedActions:   recommendedActions,
		SuggestedOwnerRoles:  append([]string{}, meta.ownerRoles...),
		SuggestedAutomations: append([]string{}, meta.automationSuggestions...),
		Owner:                meta.owner,
		MappedBottleneck:     meta.bottleneck,
	}
}

func filterResults(results []models.ValidationResult, capabilityFilter string) ([]models.ValidationResult, error) {
	capabilityFilter = strings.TrimSpace(capabilityFilter)
	if capabilityFilter == "" {
		filtered := make([]models.ValidationResult, 0, len(results))
		for _, result := range results {
			if isScoredCategory(result.Capability) {
				filtered = append(filtered, result)
			}
		}
		return filtered, nil
	}

	for _, result := range results {
		if strings.EqualFold(result.Capability, capabilityFilter) {
			return []models.ValidationResult{result}, nil
		}
	}

	return nil, fmt.Errorf("unknown capability %q", capabilityFilter)
}

func isScoredCategory(category string) bool {
	switch category {
	case "Intent", "Behavior", "Design", "Assurance", "Security", "Execution":
		return true
	default:
		return false
	}
}

func metadataFor(capability string) capabilityMetadata {
	if meta, ok := metadataByCapability[capability]; ok {
		return meta
	}

	return capabilityMetadata{
		owner:          "Execution Engineer",
		ownerRoles:     []string{"Technical Lead"},
		bottleneck:     "Delivery friction",
		whyThisMatters: "This evidence contributes to the overall validity of the release decision.",
		riskToDelivery: "The team may miss a delivery risk if this evidence is not understood.",
		nextActionBase: []string{
			"Inspect the affected artifact and validation output for this capability.",
		},
		automationSuggestions: []string{
			"Run bottleneck validate in CI.",
		},
	}
}

func evidenceFound(validation models.ValidationResult, graph *traceability.Graph) []string {
	var found []string
	if artifact := artifactForCapability(validation.Capability); artifact != "" && !isMissingArtifactMessage(validation.Message) {
		found = append(found, artifact+" exists")
	}
	if graph != nil {
		for _, node := range nodesForCapability(graph, validation.Capability) {
			found = append(found, nodeFact(node))
		}
	}
	for _, detail := range validation.Details {
		if isGapDetail(detail) {
			continue
		}
		found = append(found, detail)
	}
	if validation.Message != "" && !isGapDetail(validation.Message) && validation.Status != models.StatusPass {
		found = append(found, "validation reported: "+validation.Message)
	}
	if validation.EvidenceQuality.Score > 0 && !containsQualityScore(found) {
		found = append(found, fmt.Sprintf("evidence quality score: %d", validation.EvidenceQuality.Score))
	}
	if len(found) == 0 {
		found = append(found, "No concrete evidence found.")
	}
	return uniqueStrings(found)
}

func containsQualityScore(values []string) bool {
	for _, value := range values {
		if strings.Contains(strings.ToLower(value), "evidence quality score") {
			return true
		}
	}
	return false
}

func evidenceMissing(validation models.ValidationResult, graph *traceability.Graph, allResults map[string]models.ValidationResult) []string {
	var missing []string
	missing = append(missing, ruleEvidenceMissing(validation)...)
	missing = append(missing, missingFromMessage(validation)...)
	missing = append(missing, validation.EvidenceQuality.Missing...)
	for _, detail := range validation.Details {
		if isGapDetail(detail) {
			missing = append(missing, detail)
		}
	}
	if graph != nil {
		for _, finding := range graph.ValidateFindings() {
			if findingBelongsToCategory(finding, validation.Capability) {
				missing = append(missing, finding.Message)
			}
		}
	}
	if validation.Capability != "Traceability" {
		if traceResult, ok := allResults["Traceability"]; ok {
			for _, detail := range traceResult.Details {
				if detailReferencesCategory(detail, validation.Capability) {
					missing = append(missing, detail)
				}
			}
		}
	}
	if len(missing) == 0 {
		missing = append(missing, "None.")
	}
	return uniqueStrings(missing)
}

func relatedIDsFor(category string, validation models.ValidationResult, graph *traceability.Graph) []string {
	ids := []string{}
	if graph != nil {
		for _, node := range nodesForCapability(graph, category) {
			ids = append(ids, node.ID)
			ids = append(ids, node.Refs...)
			for _, inbound := range graph.InboundRefs(node.ID) {
				ids = append(ids, inbound)
			}
		}
	}
	ids = append(ids, extractIDs(validation.Message)...)
	ids = append(ids, extractIDs(strings.Join(validation.Details, "\n"))...)
	ids = append(ids, extractIDs(strings.Join(validation.EvidenceQuality.Details, "\n"))...)
	ids = append(ids, extractIDs(strings.Join(validation.EvidenceQuality.Missing, "\n"))...)
	sort.Slice(ids, func(i, j int) bool {
		return nodeSortKey(ids[i]) < nodeSortKey(ids[j])
	})
	return uniqueStrings(ids)
}

func scoreImpactsFor(validation models.ValidationResult) []models.ScoreImpact {
	if len(validation.EvidenceQuality.ScoreImpacts) > 0 {
		return validation.EvidenceQuality.ScoreImpacts
	}
	switch validation.Status {
	case models.StatusFail:
		return []models.ScoreImpact{{Reason: fallbackImpactReason(validation), Delta: -40}}
	case models.StatusWarning:
		return []models.ScoreImpact{{Reason: fallbackImpactReason(validation), Delta: -20}}
	default:
		return nil
	}
}

func fallbackImpactReason(validation models.ValidationResult) string {
	if validation.Message != "" {
		return validation.Message
	}
	return strings.ToLower(validation.Capability) + " validation did not pass"
}

func missingFromMessage(validation models.ValidationResult) []string {
	message := strings.ToLower(validation.Message)
	artifact := artifactForCapability(validation.Capability)
	switch {
	case strings.Contains(message, "missing") && artifact != "":
		return []string{"Create or regenerate " + artifact + "."}
	case strings.Contains(message, "invalid") && artifact != "":
		return []string{"Regenerate " + artifact + " as valid evidence."}
	case strings.Contains(message, "threshold"):
		return []string{"Provide " + strings.ToLower(validation.Capability) + " evidence that satisfies the configured threshold."}
	default:
		return nil
	}
}

func ruleEvidenceMissing(validation models.ValidationResult) []string {
	if validation.Status == models.StatusPass {
		return nil
	}
	combined := combinedEvidenceText(validation)
	switch validation.Capability {
	case "Intent":
		return ruleMatches(combined, []ruleMatch{
			{markers: []string{"missing", "intent"}, text: "Missing intent evidence."},
			{markers: []string{"measurable"}, text: "Intent exists but does not clearly define measurable outcomes."},
			{markers: []string{"placeholder"}, text: "Intent contains placeholder or thin content."},
			{markers: []string{"too thin"}, text: "Intent contains placeholder or thin content."},
			{markers: []string{"thin content"}, text: "Intent contains placeholder or thin content."},
		})
	case "Behavior":
		return ruleMatches(combined, []ruleMatch{
			{markers: []string{"missing", "behavior"}, text: "Missing behavior specification."},
			{markers: []string{"behavior-*"}, text: "Behavior expectations are not traceable."},
			{markers: []string{"behavior id"}, text: "Behavior expectations are not traceable."},
			{markers: []string{"no mapped test"}, text: "Behavior is not validated."},
			{markers: []string{"not linked", "test"}, text: "Behavior is not validated."},
		})
	case "Design":
		return ruleMatches(combined, []ruleMatch{
			{markers: []string{"missing", "architecture"}, text: "Missing architecture evidence."},
			{markers: []string{"tradeoff"}, text: "Architecture exists but does not explain tradeoffs."},
			{markers: []string{"failure mode"}, text: "Architecture does not describe failure modes."},
			{markers: []string{"operational"}, text: "Architecture does not describe failure modes."},
			{markers: []string{"fallback"}, text: "Architecture does not describe failure modes."},
			{markers: []string{"monitoring"}, text: "Architecture does not describe failure modes."},
		})
	case "Assurance":
		return ruleMatches(combined, []ruleMatch{
			{markers: []string{"missing", "assurance"}, text: "Missing automated validation evidence."},
			{markers: []string{"no assurance"}, text: "Missing automated validation evidence."},
			{markers: []string{"no mapped test"}, text: "Tests exist but are not linked to behavior IDs."},
			{markers: []string{"not linked", "test"}, text: "Tests exist but are not linked to behavior IDs."},
			{markers: []string{"coverage"}, text: "Critical behaviors lack validation."},
			{markers: []string{"critical", "validation"}, text: "Critical behaviors lack validation."},
		})
	case "Security":
		return ruleMatches(combined, []ruleMatch{
			{markers: []string{"missing", "security"}, text: "Missing security evidence."},
			{markers: []string{"missing", "guardrail"}, text: "Security guardrails are not documented."},
			{markers: []string{"guardrail"}, text: "Security guardrails are not documented."},
			{markers: []string{"high"}, text: "High severity security findings exist."},
			{markers: []string{"critical"}, text: "High severity security findings exist."},
		})
	case "Execution":
		return ruleMatches(combined, []ruleMatch{
			{markers: []string{"missing", "telemetry"}, text: "Missing execution evidence."},
			{markers: []string{"missing", "execution"}, text: "Missing execution evidence."},
			{markers: []string{"adoption"}, text: "Execution evidence suggests weak adoption or user trust."},
			{markers: []string{"user trust"}, text: "Execution evidence suggests weak adoption or user trust."},
			{markers: []string{"override"}, text: "Execution evidence suggests weak adoption or user trust."},
			{markers: []string{"error"}, text: "Execution evidence suggests operational instability."},
			{markers: []string{"latency"}, text: "Execution evidence suggests operational instability."},
			{markers: []string{"incident"}, text: "Execution evidence suggests operational instability."},
			{markers: []string{"change_failure"}, text: "Execution evidence suggests operational instability."},
		})
	default:
		return nil
	}
}

type ruleMatch struct {
	markers []string
	text    string
}

func ruleMatches(combined string, matches []ruleMatch) []string {
	var values []string
	for _, match := range matches {
		if allMarkersPresent(combined, match.markers) {
			values = append(values, match.text)
		}
	}
	return uniqueStrings(values)
}

func allMarkersPresent(text string, markers []string) bool {
	for _, marker := range markers {
		if !strings.Contains(text, marker) {
			return false
		}
	}
	return true
}

func recommendedActionsFor(validation models.ValidationResult, missing []string, recommendation string, meta capabilityMetadata) []string {
	actions := []string{}
	for _, gap := range missing {
		actions = append(actions, recommendationForGap(validation.Capability, gap))
	}
	if recommendation != "" {
		actions = append(actions, recommendation)
	}
	actions = append(actions, meta.nextActionBase...)
	return uniqueStrings(actions)
}

func recommendationForGap(category string, gap string) string {
	lower := strings.ToLower(gap)
	switch category {
	case "Intent":
		switch {
		case strings.Contains(lower, "missing intent"):
			return "Create bottleneck/intent/intent.md with measurable outcomes and constraints."
		case strings.Contains(lower, "measurable outcomes"):
			return "Add observable outcomes, business constraints, and unacceptable outcomes."
		case strings.Contains(lower, "placeholder") || strings.Contains(lower, "thin content"):
			return "Replace template text with product-specific intent."
		}
	case "Behavior":
		switch {
		case strings.Contains(lower, "missing behavior"):
			return "Create behavior-spec.md with expected and unacceptable behaviors."
		case strings.Contains(lower, "not traceable"):
			return "Add stable behavior IDs such as BEHAVIOR-001."
		case strings.Contains(lower, "not validated"):
			return "Map each critical behavior to test evidence."
		}
	case "Design":
		switch {
		case strings.Contains(lower, "missing architecture"):
			return "Document major components, boundaries, dependencies, and risk decisions."
		case strings.Contains(lower, "tradeoffs"):
			return "Add decision records for key constraints and design choices."
		case strings.Contains(lower, "failure modes"):
			return "Add fallback, monitoring, and operational assumptions."
		}
	case "Assurance":
		switch {
		case strings.Contains(lower, "missing automated validation"):
			return "Add test output or BDD evidence under bottleneck/assurance/."
		case strings.Contains(lower, "not linked"):
			return "Add traceability references from test evidence to behavior expectations."
		case strings.Contains(lower, "critical behaviors"):
			return "Prioritize tests for high-risk behaviors before expanding feature scope."
		}
	case "Security":
		switch {
		case strings.Contains(lower, "missing security"):
			return "Add CodeQL, dependency review, secret scanning, or SARIF evidence."
		case strings.Contains(lower, "high severity"):
			return "Block release until findings are triaged or resolved."
		case strings.Contains(lower, "guardrails"):
			return "Add security/guardrails.json or equivalent evidence."
		}
	case "Execution":
		switch {
		case strings.Contains(lower, "missing execution"):
			return "Add telemetry or production-readiness evidence."
		case strings.Contains(lower, "weak adoption") || strings.Contains(lower, "user trust"):
			return "Review user workflow, training, and feedback loops."
		case strings.Contains(lower, "operational instability"):
			return "Address error rate, latency, or incident signals before accelerating release."
		}
	}
	return ""
}

func riskToDeliveryFor(validation models.ValidationResult, missing []string, meta capabilityMetadata) string {
	if validation.Status == models.StatusPass && onlyNoneMissing(missing) {
		return "Current evidence lowers delivery risk for this category, but it should stay current as the release changes."
	}
	return meta.riskToDelivery
}

func onlyNoneMissing(missing []string) bool {
	if len(missing) == 0 {
		return true
	}
	if len(missing) == 1 && strings.EqualFold(strings.TrimSpace(missing[0]), "None.") {
		return true
	}
	return false
}

func combinedEvidenceText(validation models.ValidationResult) string {
	parts := []string{
		validation.Message,
		strings.Join(validation.Details, "\n"),
		strings.Join(validation.EvidenceQuality.Details, "\n"),
		strings.Join(validation.EvidenceQuality.Missing, "\n"),
	}
	for _, impact := range validation.EvidenceQuality.ScoreImpacts {
		parts = append(parts, impact.Reason)
	}
	return strings.ToLower(strings.Join(parts, "\n"))
}

func artifactForCapability(capability string) string {
	switch capability {
	case "Intent":
		return "bottleneck/intent/intent.md"
	case "Behavior":
		return "bottleneck/behavior/behavior-spec.md"
	case "Design":
		return "bottleneck/design/architecture.md"
	case "Assurance":
		return "bottleneck/assurance/results.json"
	case "Security":
		return "bottleneck/security/guardrails.json"
	case "Execution":
		return "bottleneck/execution/telemetry.json"
	case "Config":
		return "bottleneck/config.yaml"
	default:
		return ""
	}
}

func isMissingArtifactMessage(message string) bool {
	lower := strings.ToLower(message)
	return strings.Contains(lower, "missing ") || strings.Contains(lower, "missing or invalid")
}

func isGapDetail(detail string) bool {
	lower := strings.ToLower(detail)
	if strings.HasPrefix(lower, "failure:") {
		return false
	}
	gapMarkers := []string{
		"missing",
		"not linked",
		"no ",
		"placeholder",
		"too thin",
		"stale",
		"below threshold",
		"above threshold",
		"invalid",
		"orphan",
		"cannot reference",
		"references missing",
		"does not",
	}
	for _, marker := range gapMarkers {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func nodesForCapability(graph *traceability.Graph, capability string) []traceability.Node {
	if graph == nil {
		return nil
	}
	nodeType := capability
	var nodes []traceability.Node
	for _, id := range graph.OrderedIDs {
		node := graph.Nodes[id]
		if node.Type == nodeType {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

func nodeFact(node traceability.Node) string {
	text := fmt.Sprintf("%s found in %s", node.ID, node.ArtifactPath)
	if node.Title != "" {
		text += " (" + node.Title + ")"
	}
	if node.Status != "" {
		text += " status=" + node.Status
	}
	return text
}

func findingBelongsToCategory(finding traceability.Finding, category string) bool {
	return idBelongsToCategory(finding.SourceID, category) || idBelongsToCategory(finding.ReferenceID, category)
}

func detailReferencesCategory(detail string, category string) bool {
	lower := strings.ToLower(detail)
	if category == "Assurance" && (strings.Contains(lower, "no mapped test evidence") || strings.Contains(lower, "test evidence")) {
		return true
	}
	for _, id := range extractIDs(detail) {
		if idBelongsToCategory(id, category) {
			return true
		}
	}
	return false
}

func idBelongsToCategory(id string, category string) bool {
	prefix := strings.ToUpper(category)
	if category == "Traceability" || category == "Config" {
		return false
	}
	return strings.HasPrefix(id, prefix+"-")
}

func extractIDs(text string) []string {
	return evidenceIDPattern.FindAllString(text, -1)
}

func factsFromStrings(values []string) []EvidenceFact {
	values = uniqueStrings(values)
	facts := make([]EvidenceFact, 0, len(values))
	for _, value := range values {
		facts = append(facts, EvidenceFact{Text: value})
	}
	return facts
}

func gapsFromStrings(values []string) []EvidenceGap {
	values = uniqueStrings(values)
	gaps := make([]EvidenceGap, 0, len(values))
	for _, value := range values {
		gaps = append(gaps, EvidenceGap{Text: value})
	}
	return gaps
}

func appendFactLines(lines []string, facts []EvidenceFact) []string {
	if len(facts) == 0 {
		return append(lines, "- None.")
	}
	for _, fact := range facts {
		lines = append(lines, "- "+fact.Text)
	}
	return lines
}

func appendGapLines(lines []string, gaps []EvidenceGap) []string {
	if len(gaps) == 0 {
		return append(lines, "- None.")
	}
	for _, gap := range gaps {
		lines = append(lines, "- "+gap.Text)
	}
	return lines
}

func appendBulletLines(lines []string, values []string, empty string) []string {
	if len(values) == 0 {
		return append(lines, "- "+empty)
	}
	for _, value := range values {
		lines = append(lines, "- "+value)
	}
	return lines
}

func appendNumberedLines(lines []string, values []string) []string {
	values = uniqueStrings(values)
	if len(values) == 0 {
		return append(lines, "1. Inspect the category evidence and repair the highest-risk gap.")
	}
	for index, value := range values {
		lines = append(lines, fmt.Sprintf("%d. %s", index+1, value))
	}
	return lines
}

func appendImpactLines(lines []string, impacts []models.ScoreImpact) []string {
	if len(impacts) == 0 {
		return append(lines, "- None.")
	}
	for _, impact := range impacts {
		lines = append(lines, fmt.Sprintf("- %+d %s", impact.Delta, impact.Reason))
	}
	return lines
}

func uniqueStrings(values []string) []string {
	seen := map[string]struct{}{}
	unique := make([]string, 0, len(values))
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
	return unique
}

func nodeSortKey(id string) string {
	order := 99
	switch {
	case strings.HasPrefix(id, "INTENT-"):
		order = 1
	case strings.HasPrefix(id, "BEHAVIOR-"):
		order = 2
	case strings.HasPrefix(id, "DESIGN-"):
		order = 3
	case strings.HasPrefix(id, "ASSURANCE-"):
		order = 4
	case strings.HasPrefix(id, "SECURITY-"):
		order = 5
	case strings.HasPrefix(id, "EXECUTION-"):
		order = 6
	}
	return fmt.Sprintf("%02d:%s", order, id)
}

package explainer

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"bottleneck/internal/diagnosis"
	"bottleneck/internal/models"
	"bottleneck/internal/traceability"
)

type capabilityMetadata struct {
	owner          string
	bottleneck     string
	nextActionBase []string
}

type CategoryExplanation struct {
	Category        string               `json:"category"`
	Score           int                  `json:"score"`
	Status          string               `json:"status"`
	EvidenceFound   []EvidenceFact       `json:"evidence_found"`
	EvidenceMissing []EvidenceGap        `json:"evidence_missing"`
	RelatedIDs      []string             `json:"related_ids"`
	ScoreImpacts    []models.ScoreImpact `json:"score_impacts"`
	Recommendation  string               `json:"recommendation"`
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
		owner:      "Intent Engineer",
		bottleneck: "Ambiguous requirements",
		nextActionBase: []string{
			"Update bottleneck/intent/intent.md with concrete outcomes, constraints, success criteria, and INTENT-* refs.",
		},
	},
	"Behavior": {
		owner:      "Behavior Engineer",
		bottleneck: "Non-deterministic outputs",
		nextActionBase: []string{
			"Update bottleneck/behavior/behavior-spec.md with concrete BEHAVIOR-* expectations and refs.",
		},
	},
	"Design": {
		owner:      "Design Engineer",
		bottleneck: "Poor adoption / UX gaps",
		nextActionBase: []string{
			"Update bottleneck/design/architecture.md with DESIGN-* evidence that references the relevant intent and behavior IDs.",
		},
	},
	"Assurance": {
		owner:      "Assurance Engineer",
		bottleneck: "Validation gaps",
		nextActionBase: []string{
			"Add Cucumber, evaluation, or test evidence in bottleneck/assurance/results.json that references the affected BEHAVIOR-* IDs.",
		},
	},
	"Security": {
		owner:      "Security Engineer",
		bottleneck: "Risk & compliance",
		nextActionBase: []string{
			"Regenerate bottleneck/security/guardrails.json from SARIF or guardrail evidence and resolve findings above threshold.",
		},
	},
	"Execution": {
		owner:      "Execution Engineer",
		bottleneck: "Delivery friction",
		nextActionBase: []string{
			"Refresh bottleneck/execution/telemetry.json with current reliability, adoption, override, deployment, and cost signals.",
		},
	},
	"Traceability": {
		owner:      "Release Engineer",
		bottleneck: "Traceability gaps",
		nextActionBase: []string{
			"Run bottleneck trace --id <ID> for the affected evidence ID and repair missing, duplicate, or orphaned links.",
		},
	},
	"Config": {
		owner:      "Execution Engineer",
		bottleneck: "Delivery friction",
		nextActionBase: []string{
			"Repair bottleneck/config.yaml so the selected environment and thresholds resolve cleanly.",
		},
	},
}

func Render(result models.EngineResult, capabilityFilter string) (string, error) {
	return RenderWithGraph(result, nil, capabilityFilter)
}

func RenderWithGraph(result models.EngineResult, graph *traceability.Graph, capabilityFilter string) (string, error) {
	filtered, err := filterResults(result.Results, capabilityFilter)
	if err != nil {
		return "", err
	}
	diagnosisResult := diagnosis.Analyze(result)
	explanations := categoryExplanations(result, filtered, diagnosisResult.CategoryScores, graph)

	var lines []string
	lines = append(lines,
		fmt.Sprintf("Environment: %s", result.Environment),
		fmt.Sprintf("System Status: %s", result.SystemStatus),
		fmt.Sprintf("Primary Bottleneck: %s", diagnosisResult.PrimaryBottleneck),
	)
	if capabilityFilter == "" {
		lines = append(lines,
			"",
			"Primary Diagnosis:",
			fmt.Sprintf("Weakest Category: %s", diagnosisResult.PrimaryBottleneck),
			"Top Evidence:",
		)
		lines = appendBulletLines(lines, diagnosisResult.ContributingFindings, "None.")
		lines = append(lines, fmt.Sprintf("Next Action: %s", diagnosisResult.RecommendedAction))
		if len(diagnosisResult.TiedBottlenecks) > 0 {
			lines = append(lines, fmt.Sprintf("Tied Bottlenecks: %s", strings.Join(diagnosisResult.TiedBottlenecks, ", ")))
		}
	}

	for _, explanation := range explanations {
		meta := metadataFor(explanation.Category)
		lines = append(lines, "")
		lines = append(lines,
			fmt.Sprintf("%s Score: %d", explanation.Category, explanation.Score),
			fmt.Sprintf("Status: %s", explanation.Status),
			fmt.Sprintf("Owner: %s", meta.owner),
			fmt.Sprintf("Mapped Bottleneck: %s", meta.bottleneck),
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
	}

	return strings.Join(lines, "\n"), nil
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

	return CategoryExplanation{
		Category:        category,
		Score:           diagnosis.ScoreFor(category, scores),
		Status:          validation.Status,
		EvidenceFound:   factsFromStrings(found),
		EvidenceMissing: gapsFromStrings(missing),
		RelatedIDs:      relatedIDs,
		ScoreImpacts:    impacts,
		Recommendation:  recommendation,
	}
}

func filterResults(results []models.ValidationResult, capabilityFilter string) ([]models.ValidationResult, error) {
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
		owner:      "Execution Engineer",
		bottleneck: "Delivery friction",
		nextActionBase: []string{
			"Inspect the affected artifact and validation output for this capability.",
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

package validator

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"bottleneck/internal/models"
)

const (
	placeholderIntendedBehavior      = "Describe intended system behavior."
	placeholderUnacceptableBehavior  = "Describe behavior the system must prevent."
	placeholderRequiredOutcomes      = "Describe required outcomes."
	placeholderSystemConstraints     = "Describe system constraints."
	placeholderSuccessCriteria       = "Describe measurable success criteria."
	placeholderSystemArchitecture    = "Describe system architecture."
	contentQualityWarningMessage     = "content quality warnings detected"
	contentQualityStrictFailMessage  = "content quality failures detected"
	contentQualityPlaceholderMessage = "%s section %q still contains placeholder content"
	contentQualityThinMessage        = "%s section %q is too thin to validate"
)

var (
	contentQualityTokens       = regexp.MustCompile(`[A-Za-z0-9]+`)
	evidenceIDPattern          = regexp.MustCompile(`\b(?:INTENT|BEHAVIOR|DESIGN|ASSURANCE|SECURITY|EXECUTION)-[0-9]{3,}\b`)
	numberOrPercentPattern     = regexp.MustCompile(`\b[0-9]+(?:\.[0-9]+)?%?\b`)
	contentQualityPlaceholders = []string{
		placeholderIntendedBehavior,
		placeholderUnacceptableBehavior,
		placeholderRequiredOutcomes,
		placeholderSystemConstraints,
		placeholderSuccessCriteria,
		placeholderSystemArchitecture,
		"Describe required outcomes",
		"Describe system constraints",
		"TODO",
		"TBD",
		"Add measurable success criteria",
	}
	measurablePhrases = []string{
		"at least",
		"no more than",
		"below",
		"above",
		"within",
		"less than",
		"greater than",
		"threshold",
		"percent",
		"percentage",
		"by ",
		"before ",
		"after ",
	}
	weakLanguageWords = map[string]bool{
		"better":        true,
		"improve":       true,
		"fast":          true,
		"easy":          true,
		"robust":        true,
		"user-friendly": true,
	}
	genericContentWords = map[string]bool{
		"a":            true,
		"an":           true,
		"and":          true,
		"architecture": true,
		"behavior":     true,
		"constraints":  true,
		"content":      true,
		"criteria":     true,
		"describe":     true,
		"description":  true,
		"design":       true,
		"details":      true,
		"intended":     true,
		"na":           true,
		"none":         true,
		"outcomes":     true,
		"placeholder":  true,
		"required":     true,
		"stuff":        true,
		"success":      true,
		"system":       true,
		"tbd":          true,
		"things":       true,
		"todo":         true,
	}
)

type sectionContentRequirement struct {
	section     string
	placeholder string
}

func contentQualityResult(capability string, details []string, strict bool) models.ValidationResult {
	status := models.StatusWarning
	message := contentQualityWarningMessage
	if strict {
		status = models.StatusFail
		message = contentQualityStrictFailMessage
	}

	return models.ValidationResult{
		Capability: capability,
		Status:     status,
		Message:    message,
		Details:    details,
	}
}

func contentQualityResultWithQuality(capability string, quality models.EvidenceQuality, strict bool) models.ValidationResult {
	result := contentQualityResult(capability, quality.Details, strict)
	result.Findings = findingsForQuality(result.Status, quality)
	result.EvidenceQuality = quality
	return result
}

func markdownSectionContentDetails(rootPath string, relativePath string, content string, requirements []sectionContentRequirement) []string {
	var details []string
	for _, requirement := range requirements {
		body, ok := markdownSectionBody(content, requirement.section)
		if !ok {
			continue
		}

		details = append(details, contentQualityDetails(
			contentQualityArtifactPath(rootPath, relativePath),
			requirement.section,
			body,
			requirement.placeholder,
		)...)
	}

	return details
}

func markdownDocumentContentDetails(rootPath string, relativePath string, section string, content string, placeholder string) []string {
	return contentQualityDetails(
		contentQualityArtifactPath(rootPath, relativePath),
		section,
		markdownDocumentBody(content),
		placeholder,
	)
}

func contentQualityDetails(artifactPath string, section string, body string, placeholder string) []string {
	if containsPlaceholder(body) || (placeholder != "" && strings.Contains(body, placeholder)) {
		return []string{fmt.Sprintf(contentQualityPlaceholderMessage, artifactPath, section)}
	}

	if !hasSufficientContent(body) {
		return []string{fmt.Sprintf(contentQualityThinMessage, artifactPath, section)}
	}

	return nil
}

func evaluateMarkdownEvidenceQuality(rootPath string, relativePath string, artifactType string, content string, requirements []sectionContentRequirement) models.EvidenceQuality {
	artifactPath := contentQualityArtifactPath(rootPath, relativePath)
	quality := models.EvidenceQuality{Score: 100}

	text := strings.TrimSpace(content)
	if text == "" {
		return evidenceQualityWithImpact(0, fmt.Sprintf("%s is empty", artifactPath), artifactPath+" is empty", "Add substantive evidence to "+artifactPath+".")
	}

	headerOnly := isHeaderOnlyMarkdown(text)
	if headerOnly {
		addQualityIssue(&quality, fmt.Sprintf("%s is header-only", artifactPath), -45)
		quality.Missing = append(quality.Missing, "Add body evidence under the required Markdown sections.")
	}

	placeholderLineCount := countPlaceholderLines(text)
	meaningfulLineCount := countMeaningfulLines(text)
	placeholderHeavy := placeholderLineCount > 0 && placeholderLineCount >= meaningfulLineCount
	if placeholderHeavy {
		addQualityIssue(&quality, fmt.Sprintf("%s is placeholder-heavy", artifactPath), -30)
		quality.Missing = append(quality.Missing, "Replace starter placeholder content with project-specific evidence.")
	}

	for _, requirement := range requirements {
		body, ok := markdownSectionBody(text, requirement.section)
		if !ok {
			continue
		}
		details := contentQualityDetails(artifactPath, requirement.section, body, requirement.placeholder)
		for _, detail := range details {
			delta := -20
			if strings.Contains(detail, "too thin") {
				delta = -15
			}
			addQualityIssue(&quality, detail, delta)
			if strings.Contains(detail, "placeholder") {
				quality.Missing = append(quality.Missing, fmt.Sprintf("%s section %q needs real evidence", artifactPath, requirement.section))
			}
			if strings.Contains(detail, "too thin") {
				quality.Missing = append(quality.Missing, fmt.Sprintf("%s section %q needs a meaningful sentence or list item", artifactPath, requirement.section))
			}
		}
	}

	switch artifactType {
	case "Intent":
		if !hasEvidenceID(text, "INTENT") {
			addQualityIssue(&quality, fmt.Sprintf("%s does not define an INTENT-* evidence ID", artifactPath), -20)
			quality.Missing = append(quality.Missing, "Add an INTENT-* heading such as ### INTENT-001: ...")
		}
		if body, ok := markdownSectionBody(text, "Success Criteria"); ok && !hasMeasurableLanguage(body) {
			addQualityIssue(&quality, fmt.Sprintf("%s section %q does not include measurable criteria", artifactPath, "Success Criteria"), -15)
			quality.Missing = append(quality.Missing, "Add measurable success criteria with numbers, thresholds, dates, or time windows.")
		}
		if hasWeakLanguage(text) && !hasMeasurableLanguage(text) {
			addQualityIssue(&quality, fmt.Sprintf("%s relies on vague outcome language without measurable criteria", artifactPath), -10)
		}
	case "Behavior":
		if !hasEvidenceID(text, "BEHAVIOR") {
			addQualityIssue(&quality, fmt.Sprintf("%s does not define a BEHAVIOR-* evidence ID", artifactPath), -20)
			quality.Missing = append(quality.Missing, "Add a BEHAVIOR-* heading such as ### BEHAVIOR-001: ...")
		}
	case "Design":
		if !hasEvidenceID(text, "DESIGN") && meaningfulLineCount < 3 {
			addQualityIssue(&quality, fmt.Sprintf("%s does not define a DESIGN-* evidence ID or enough concrete architecture evidence", artifactPath), -15)
			quality.Missing = append(quality.Missing, "Add a DESIGN-* heading or enough concrete architecture decisions to make the design reviewable.")
		}
	}

	if headerOnly && quality.Score < 20 {
		quality.Score = 20
	}
	if placeholderHeavy && quality.Score < 25 {
		quality.Score = 25
	}
	quality.Score = clampQualityScore(quality.Score)
	quality.Details = uniqueStrings(quality.Details)
	quality.Missing = uniqueStrings(quality.Missing)
	quality.ScoreImpacts = uniqueImpacts(quality.ScoreImpacts)
	return quality
}

func evidenceQualityWithImpact(score int, detail string, impactReason string, missing string) models.EvidenceQuality {
	quality := models.EvidenceQuality{Score: score}
	addQualityIssue(&quality, detail, score-100)
	quality.ScoreImpacts[0].Reason = impactReason
	if missing != "" {
		quality.Missing = append(quality.Missing, missing)
	}
	return quality
}

func addQualityIssue(quality *models.EvidenceQuality, detail string, delta int) {
	if detail == "" {
		return
	}
	quality.Details = append(quality.Details, detail)
	quality.ScoreImpacts = append(quality.ScoreImpacts, models.ScoreImpact{
		Reason: detail,
		Delta:  delta,
	})
	quality.Score = clampQualityScore(quality.Score + delta)
}

func qualityStatus(capability string, quality models.EvidenceQuality, strict bool) (string, string) {
	if quality.Score >= 80 && len(quality.Details) == 0 {
		return models.StatusPass, ""
	}
	if strict {
		return models.StatusFail, contentQualityStrictFailMessage
	}
	return models.StatusWarning, fmt.Sprintf("%s evidence quality is weak", strings.ToLower(capability))
}

func strictValue(values []bool) bool {
	return len(values) > 0 && values[0]
}

func qualityPassDetails(rootPath string, relativePath string, sections []string) []string {
	artifactPath := contentQualityArtifactPath(rootPath, relativePath)
	details := make([]string, 0, len(sections)+1)
	for _, section := range sections {
		details = append(details, artifactPath+fmt.Sprintf(` section %q contains meaningful evidence`, section))
	}
	details = append(details, fmt.Sprintf("%s evidence quality score: 100", artifactPath))
	return details
}

func findingsForQuality(status string, quality models.EvidenceQuality) []models.ValidationFinding {
	if len(quality.Details) == 0 {
		return nil
	}

	level := "warning"
	if status == models.StatusFail {
		level = "error"
	}

	findings := make([]models.ValidationFinding, 0, len(quality.Details))
	for _, detail := range quality.Details {
		findings = append(findings, models.ValidationFinding{
			Level:   level,
			Message: detail,
			Path:    pathFromQualityDetail(detail),
		})
	}
	return findings
}

func pathFromQualityDetail(detail string) string {
	match := regexp.MustCompile(`(bottleneck/[^\s:]+)`).FindStringSubmatch(detail)
	if len(match) < 2 {
		return ""
	}
	return strings.TrimRight(match[1], ".,)")
}

func containsPlaceholder(content string) bool {
	lowerContent := strings.ToLower(content)
	for _, placeholder := range contentQualityPlaceholders {
		if strings.Contains(lowerContent, strings.ToLower(placeholder)) {
			return true
		}
	}
	return false
}

func countPlaceholderLines(content string) int {
	count := 0
	for _, line := range strings.Split(content, "\n") {
		if containsPlaceholder(line) {
			count++
		}
	}
	return count
}

func countMeaningfulLines(content string) int {
	count := 0
	for _, line := range strings.Split(content, "\n") {
		if isMeaningfulContentLine(line) {
			count++
		}
	}
	return count
}

func isHeaderOnlyMarkdown(content string) bool {
	hasHeading := false
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if _, _, ok := markdownHeading(trimmed); ok {
			hasHeading = true
			continue
		}
		return false
	}
	return hasHeading
}

func hasEvidenceID(content string, prefix string) bool {
	for _, id := range evidenceIDPattern.FindAllString(content, -1) {
		if strings.HasPrefix(id, prefix+"-") {
			return true
		}
	}
	return false
}

func hasMeasurableLanguage(content string) bool {
	lowerContent := strings.ToLower(content)
	if numberOrPercentPattern.MatchString(lowerContent) {
		return true
	}
	for _, phrase := range measurablePhrases {
		if strings.Contains(lowerContent, phrase) {
			return true
		}
	}
	return false
}

func hasWeakLanguage(content string) bool {
	lowerContent := strings.ToLower(content)
	for word := range weakLanguageWords {
		if strings.Contains(lowerContent, word) {
			return true
		}
	}
	return false
}

func hasSufficientContent(body string) bool {
	cleaned := body
	for _, placeholder := range contentQualityPlaceholders {
		cleaned = strings.ReplaceAll(cleaned, placeholder, " ")
	}

	for _, line := range strings.Split(cleaned, "\n") {
		if isMeaningfulContentLine(line) {
			return true
		}
	}

	return false
}

func isMeaningfulContentLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return false
	}
	if _, _, ok := markdownHeading(trimmed); ok {
		return false
	}

	isListItem := strings.HasPrefix(trimmed, "- ") ||
		strings.HasPrefix(trimmed, "* ") ||
		strings.HasPrefix(trimmed, "+ ") ||
		isNumberedListItem(trimmed)

	trimmed = strings.TrimPrefix(trimmed, ">")
	trimmed = strings.TrimSpace(trimmed)
	trimmed = strings.TrimLeft(trimmed, "-*+0123456789.)] \t")
	trimmed = strings.TrimSpace(trimmed)

	tokens := contentQualityTokens.FindAllString(trimmed, -1)
	if len(tokens) == 0 {
		return false
	}
	if len(tokens) == 1 {
		return false
	}
	if len(tokens) == 2 {
		if allGenericContentWords(tokens) {
			return false
		}

		return isListItem || strings.HasSuffix(trimmed, ".") || strings.HasSuffix(trimmed, "!") || strings.HasSuffix(trimmed, "?")
	}

	return true
}

func allGenericContentWords(tokens []string) bool {
	for _, token := range tokens {
		if !genericContentWords[strings.ToLower(token)] {
			return false
		}
	}

	return true
}

func isNumberedListItem(line string) bool {
	for index, char := range line {
		if char < '0' || char > '9' {
			return index > 0 && (char == '.' || char == ')')
		}
	}

	return false
}

func contentQualityArtifactPath(rootPath string, relativePath string) string {
	return filepath.ToSlash(filepath.Join(filepath.Base(rootPath), relativePath))
}

func clampQualityScore(score int) int {
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]bool, len(values))
	unique := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) == "" || seen[value] {
			continue
		}
		seen[value] = true
		unique = append(unique, value)
	}
	sort.Strings(unique)
	return unique
}

func uniqueImpacts(values []models.ScoreImpact) []models.ScoreImpact {
	seen := make(map[string]bool, len(values))
	unique := make([]models.ScoreImpact, 0, len(values))
	for _, value := range values {
		key := fmt.Sprintf("%s:%d", value.Reason, value.Delta)
		if strings.TrimSpace(value.Reason) == "" || seen[key] {
			continue
		}
		seen[key] = true
		unique = append(unique, value)
	}
	return unique
}

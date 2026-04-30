package validator

import (
	"fmt"
	"path/filepath"
	"regexp"
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
	contentQualityPlaceholders = []string{
		placeholderIntendedBehavior,
		placeholderUnacceptableBehavior,
		placeholderRequiredOutcomes,
		placeholderSystemConstraints,
		placeholderSuccessCriteria,
		placeholderSystemArchitecture,
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
	if strings.Contains(body, placeholder) {
		return []string{fmt.Sprintf(contentQualityPlaceholderMessage, artifactPath, section)}
	}

	if !hasSufficientContent(body) {
		return []string{fmt.Sprintf(contentQualityThinMessage, artifactPath, section)}
	}

	return nil
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

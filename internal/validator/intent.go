package validator

import (
	"os"
	"path/filepath"
	"strings"

	"bottleneck/internal/models"
)

type IntentValidator struct {
	rootPath string
	strict   bool
}

func NewIntentValidator(rootPath string, strict bool) *IntentValidator {
	return &IntentValidator{rootPath: rootPath, strict: strict}
}

func (v *IntentValidator) Validate() []models.ValidationResult {
	return []models.ValidationResult{validateIntent(v.rootPath, v.strict)}
}

func validateIntent(rootPath string, strict bool) models.ValidationResult {
	path := filepath.Join(rootPath, "intent", "intent.md")
	content, err := os.ReadFile(path)
	if err != nil {
		return models.ValidationResult{
			Capability: "Intent",
			Status:     models.StatusFail,
			Message:    "missing intent.md",
		}
	}

	text := strings.TrimSpace(string(content))
	if !containsMarkdownSection(text, "Outcomes") ||
		!containsMarkdownSection(text, "Constraints") ||
		!containsMarkdownSection(text, "Success Criteria") {
		return models.ValidationResult{
			Capability: "Intent",
			Status:     models.StatusFail,
			Message:    "required sections missing",
		}
	}

	requirements := []sectionContentRequirement{
		{section: "Outcomes", placeholder: placeholderRequiredOutcomes},
		{section: "Constraints", placeholder: placeholderSystemConstraints},
		{section: "Success Criteria", placeholder: placeholderSuccessCriteria},
	}
	quality := evaluateMarkdownEvidenceQuality(rootPath, "intent/intent.md", "Intent", text, requirements)
	if status, message := qualityStatus("Intent", quality, strict); status != models.StatusPass {
		result := contentQualityResultWithQuality("Intent", quality, strict)
		result.Status = status
		result.Message = message
		return result
	}

	return models.ValidationResult{
		Capability:      "Intent",
		Status:          models.StatusPass,
		Details:         qualityPassDetails(rootPath, "intent/intent.md", []string{"Outcomes", "Constraints", "Success Criteria"}),
		EvidenceQuality: quality,
	}
}

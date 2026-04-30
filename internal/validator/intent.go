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

	details := markdownSectionContentDetails(rootPath, "intent/intent.md", text, []sectionContentRequirement{
		{section: "Outcomes", placeholder: placeholderRequiredOutcomes},
		{section: "Constraints", placeholder: placeholderSystemConstraints},
		{section: "Success Criteria", placeholder: placeholderSuccessCriteria},
	})
	if len(details) > 0 {
		return contentQualityResult("Intent", details, strict)
	}

	return models.ValidationResult{
		Capability: "Intent",
		Status:     models.StatusPass,
		Details: []string{
			contentQualityArtifactPath(rootPath, "intent/intent.md") + ` section "Outcomes" contains non-placeholder content`,
			contentQualityArtifactPath(rootPath, "intent/intent.md") + ` section "Constraints" contains non-placeholder content`,
			contentQualityArtifactPath(rootPath, "intent/intent.md") + ` section "Success Criteria" contains non-placeholder content`,
		},
	}
}

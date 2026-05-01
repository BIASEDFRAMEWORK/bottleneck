package validator

import (
	"os"
	"path/filepath"
	"strings"

	"bottleneck/internal/models"
)

type BehaviorValidator struct {
	rootPath string
	strict   bool
}

func NewBehaviorValidator(rootPath string, strict bool) *BehaviorValidator {
	return &BehaviorValidator{rootPath: rootPath, strict: strict}
}

func (v *BehaviorValidator) Validate() []models.ValidationResult {
	return []models.ValidationResult{validateBehavior(v.rootPath, v.strict)}
}

func validateBehavior(rootPath string, strict bool) models.ValidationResult {
	path := filepath.Join(rootPath, "behavior", "behavior-spec.md")

	content, err := os.ReadFile(path)
	if err != nil {
		return models.ValidationResult{
			Capability: "Behavior",
			Status:     models.StatusFail,
			Message:    "missing behavior-spec.md",
		}
	}

	text := strings.TrimSpace(string(content))
	if text == "" {
		return models.ValidationResult{
			Capability: "Behavior",
			Status:     models.StatusFail,
			Message:    "behavior-spec.md is empty",
		}
	}

	if !containsMarkdownSection(text, "Expected Behavior") || !containsMarkdownSection(text, "Unacceptable Behavior") {
		return models.ValidationResult{
			Capability: "Behavior",
			Status:     models.StatusFail,
			Message:    "required sections missing",
		}
	}

	requirements := []sectionContentRequirement{
		{section: "Expected Behavior", placeholder: placeholderIntendedBehavior},
		{section: "Unacceptable Behavior", placeholder: placeholderUnacceptableBehavior},
	}
	quality := evaluateMarkdownEvidenceQuality(rootPath, "behavior/behavior-spec.md", "Behavior", text, requirements)
	if status, message := qualityStatus("Behavior", quality, strict); status != models.StatusPass {
		result := contentQualityResultWithQuality("Behavior", quality, strict)
		result.Status = status
		result.Message = message
		return result
	}

	return models.ValidationResult{
		Capability:      "Behavior",
		Status:          models.StatusPass,
		Details:         qualityPassDetails(rootPath, "behavior/behavior-spec.md", []string{"Expected Behavior", "Unacceptable Behavior"}),
		EvidenceQuality: quality,
	}
}

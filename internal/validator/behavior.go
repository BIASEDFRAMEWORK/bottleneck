package validator

import (
	"os"
	"path/filepath"
	"strings"

	"biased/internal/models"
)

type BehaviorValidator struct {
	rootPath string
}

func NewBehaviorValidator(rootPath string) *BehaviorValidator {
	return &BehaviorValidator{rootPath: rootPath}
}

func (v *BehaviorValidator) Validate() []models.ValidationResult {
	return []models.ValidationResult{validateBehavior(v.rootPath)}
}

func validateBehavior(rootPath string) models.ValidationResult {
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

	return models.ValidationResult{
		Capability: "Behavior",
		Status:     models.StatusPass,
	}
}

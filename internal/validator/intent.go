package validator

import (
	"os"
	"path/filepath"
	"strings"

	"biased/internal/models"
)

type IntentValidator struct {
	rootPath string
}

func NewIntentValidator(rootPath string) *IntentValidator {
	return &IntentValidator{rootPath: rootPath}
}

func (v *IntentValidator) Validate() []models.ValidationResult {
	return []models.ValidationResult{validateIntent(v.rootPath)}
}

func validateIntent(rootPath string) models.ValidationResult {
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

	return models.ValidationResult{
		Capability: "Intent",
		Status:     models.StatusPass,
	}
}

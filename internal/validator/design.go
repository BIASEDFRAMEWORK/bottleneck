package validator

import (
	"os"
	"path/filepath"
	"strings"

	"biased/internal/models"
)

type DesignValidator struct {
	rootPath string
}

func NewDesignValidator(rootPath string) *DesignValidator {
	return &DesignValidator{rootPath: rootPath}
}

func (v *DesignValidator) Validate() []models.ValidationResult {
	return []models.ValidationResult{validateDesign(v.rootPath)}
}

func validateDesign(rootPath string) models.ValidationResult {
	path := filepath.Join(rootPath, "design", "architecture.md")
	content, err := os.ReadFile(path)
	if err != nil {
		return models.ValidationResult{
			Capability: "Design",
			Status:     models.StatusFail,
			Message:    "missing architecture.md",
		}
	}

	text := strings.TrimSpace(string(content))
	if text == "" {
		return models.ValidationResult{
			Capability: "Design",
			Status:     models.StatusFail,
			Message:    "architecture.md is empty",
		}
	}

	if !containsAnyMarkdownSection(text) {
		return models.ValidationResult{
			Capability: "Design",
			Status:     models.StatusFail,
			Message:    "section header missing",
		}
	}

	return models.ValidationResult{
		Capability: "Design",
		Status:     models.StatusPass,
	}
}

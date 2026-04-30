package validator

import (
	"os"
	"path/filepath"
	"strings"

	"bottleneck/internal/models"
)

type DesignValidator struct {
	rootPath string
	strict   bool
}

func NewDesignValidator(rootPath string, strict bool) *DesignValidator {
	return &DesignValidator{rootPath: rootPath, strict: strict}
}

func (v *DesignValidator) Validate() []models.ValidationResult {
	return []models.ValidationResult{validateDesign(v.rootPath, v.strict)}
}

func validateDesign(rootPath string, strict bool) models.ValidationResult {
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

	details := markdownDocumentContentDetails(rootPath, "design/architecture.md", "Architecture", text, placeholderSystemArchitecture)
	if len(details) > 0 {
		return contentQualityResult("Design", details, strict)
	}

	return models.ValidationResult{
		Capability: "Design",
		Status:     models.StatusPass,
		Details: []string{
			contentQualityArtifactPath(rootPath, "design/architecture.md") + ` section "Architecture" contains non-placeholder content`,
		},
	}
}

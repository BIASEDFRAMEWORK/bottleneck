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
		return artifactReadErrorResult(rootPath, "Design", "design/architecture.md", err)
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

	requirements := []sectionContentRequirement{
		{section: "Architecture", placeholder: placeholderSystemArchitecture},
	}
	quality := evaluateMarkdownEvidenceQuality(rootPath, "design/architecture.md", "Design", text, requirements)
	if status, message := qualityStatus("Design", quality, strict); status != models.StatusPass {
		result := contentQualityResultWithQuality("Design", quality, strict)
		result.Status = status
		result.Message = message
		return result
	}

	return models.ValidationResult{
		Capability:      "Design",
		Status:          models.StatusPass,
		Details:         qualityPassDetails(rootPath, "design/architecture.md", []string{"Architecture"}),
		EvidenceQuality: quality,
	}
}

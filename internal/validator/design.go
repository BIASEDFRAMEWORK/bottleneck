package validator

import (
	"os"
	"path/filepath"

	"biased/internal/models"
)

type DesignValidator struct {
	rootPath string
}

func NewDesignValidator(rootPath string) *DesignValidator {
	return &DesignValidator{rootPath: rootPath}
}

func (v *DesignValidator) Validate() []models.ValidationResult {
	path := filepath.Join(v.rootPath, "design", "architecture.md")
	if _, err := os.Stat(path); err != nil {
		return []models.ValidationResult{{
			Capability: "Design",
			Status:     models.StatusFail,
			Message:    "missing architecture.md",
		}}
	}

	return []models.ValidationResult{{
		Capability: "Design",
		Status:     models.StatusPass,
	}}
}

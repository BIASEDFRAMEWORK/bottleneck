package validator

import (
	"os"
	"path/filepath"

	"biased/internal/models"
)

type IntentValidator struct {
	rootPath string
}

func NewIntentValidator(rootPath string) *IntentValidator {
	return &IntentValidator{rootPath: rootPath}
}

func (v *IntentValidator) Validate() []models.ValidationResult {
	path := filepath.Join(v.rootPath, "intent", "intent.md")
	if _, err := os.Stat(path); err != nil {
		return []models.ValidationResult{{
			Capability: "Intent",
			Status:     models.StatusFail,
			Message:    "missing intent.md",
		}}
	}

	return []models.ValidationResult{{
		Capability: "Intent",
		Status:     models.StatusPass,
	}}
}

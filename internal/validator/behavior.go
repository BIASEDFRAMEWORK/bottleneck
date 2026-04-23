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
	path := filepath.Join(v.rootPath, "behavior", "behavior-spec.md")

	content, err := os.ReadFile(path)
	if err != nil {
		return []models.ValidationResult{{
			Capability: "Behavior",
			Status:     models.StatusFail,
			Message:    "missing behavior-spec.md",
		}}
	}

	if strings.TrimSpace(string(content)) == "" {
		return []models.ValidationResult{{
			Capability: "Behavior",
			Status:     models.StatusFail,
			Message:    "behavior-spec.md is empty",
		}}
	}

	return []models.ValidationResult{{
		Capability: "Behavior",
		Status:     models.StatusPass,
	}}
}

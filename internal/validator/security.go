package validator

import (
	"encoding/json"
	"os"
	"path/filepath"

	"biased/internal/models"
)

type SecurityValidator struct {
	rootPath string
}

type securityFile struct {
	Violations *int `json:"violations"`
}

func NewSecurityValidator(rootPath string) *SecurityValidator {
	return &SecurityValidator{rootPath: rootPath}
}

func (v *SecurityValidator) Validate() []models.ValidationResult {
	return []models.ValidationResult{validateSecurity(v.rootPath)}
}

func validateSecurity(rootPath string) models.ValidationResult {
	path := filepath.Join(rootPath, "security", "guardrails.json")
	content, err := os.ReadFile(path)
	if err != nil {
		return models.ValidationResult{
			Capability: "Security",
			Status:     models.StatusFail,
			Message:    "missing guardrails.json",
		}
	}

	var data securityFile
	if err := json.Unmarshal(content, &data); err != nil || data.Violations == nil {
		return models.ValidationResult{
			Capability: "Security",
			Status:     models.StatusFail,
			Message:    "invalid guardrails.json",
		}
	}

	if *data.Violations > 0 {
		return models.ValidationResult{
			Capability: "Security",
			Status:     models.StatusFail,
			Message:    "violations detected",
		}
	}

	return models.ValidationResult{
		Capability: "Security",
		Status:     models.StatusPass,
	}
}

package validator

import (
	"encoding/json"
	"os"
	"path/filepath"

	"biased/internal/models"
)

type ExecutionValidator struct {
	rootPath string
}

type executionFile struct {
	AdoptionRate *float64 `json:"adoption_rate"`
	ErrorRate    *float64 `json:"error_rate"`
}

func NewExecutionValidator(rootPath string) *ExecutionValidator {
	return &ExecutionValidator{rootPath: rootPath}
}

func (v *ExecutionValidator) Validate() []models.ValidationResult {
	path := filepath.Join(v.rootPath, "execution", "telemetry.json")
	content, err := os.ReadFile(path)
	if err != nil {
		return []models.ValidationResult{{
			Capability: "Execution",
			Status:     models.StatusFail,
			Message:    "missing telemetry.json",
		}}
	}

	var data executionFile
	if err := json.Unmarshal(content, &data); err != nil || data.AdoptionRate == nil || data.ErrorRate == nil {
		return []models.ValidationResult{{
			Capability: "Execution",
			Status:     models.StatusFail,
			Message:    "invalid telemetry.json",
		}}
	}

	if *data.ErrorRate > 0.05 {
		return []models.ValidationResult{{
			Capability: "Execution",
			Status:     models.StatusFail,
			Message:    "error rate above threshold",
		}}
	}

	if *data.AdoptionRate < 0.5 {
		return []models.ValidationResult{{
			Capability: "Execution",
			Status:     models.StatusWarning,
			Message:    "low adoption",
		}}
	}

	return []models.ValidationResult{{
		Capability: "Execution",
		Status:     models.StatusPass,
	}}
}

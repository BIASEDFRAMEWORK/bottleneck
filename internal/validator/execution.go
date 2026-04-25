package validator

import (
	"encoding/json"
	"os"
	"path/filepath"

	"biased/internal/config"
	"biased/internal/models"
)

type ExecutionValidator struct {
	rootPath string
	config   config.ExecutionConfig
}

type executionFile struct {
	AdoptionRate *float64 `json:"adoption_rate"`
	ErrorRate    *float64 `json:"error_rate"`
}

func NewExecutionValidator(rootPath string, cfg config.ExecutionConfig) *ExecutionValidator {
	return &ExecutionValidator{rootPath: rootPath, config: cfg}
}

func (v *ExecutionValidator) Validate() []models.ValidationResult {
	return []models.ValidationResult{validateExecution(v.rootPath, v.config)}
}

func validateExecution(rootPath string, cfg config.ExecutionConfig) models.ValidationResult {
	path := filepath.Join(rootPath, "execution", "telemetry.json")
	content, err := os.ReadFile(path)
	if err != nil {
		return models.ValidationResult{
			Capability: "Execution",
			Status:     models.StatusFail,
			Message:    "missing telemetry.json",
		}
	}

	var data executionFile
	if err := json.Unmarshal(content, &data); err != nil || data.AdoptionRate == nil || data.ErrorRate == nil {
		return models.ValidationResult{
			Capability: "Execution",
			Status:     models.StatusFail,
			Message:    "invalid telemetry.json",
		}
	}

	if *data.ErrorRate > cfg.MaxErrorRate {
		return models.ValidationResult{
			Capability: "Execution",
			Status:     models.StatusFail,
			Message:    "error rate above threshold",
		}
	}

	if *data.AdoptionRate < cfg.MinAdoption {
		return models.ValidationResult{
			Capability: "Execution",
			Status:     models.StatusWarning,
			Message:    "low adoption",
		}
	}

	return models.ValidationResult{
		Capability: "Execution",
		Status:     models.StatusPass,
	}
}

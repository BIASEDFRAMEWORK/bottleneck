package validator

import (
	"encoding/json"
	"os"
	"path/filepath"

	"bottleneck/internal/config"
	"bottleneck/internal/metrics"
	"bottleneck/internal/models"
)

type AssuranceValidator struct {
	rootPath string
	config   config.AssuranceConfig
}

func NewAssuranceValidator(rootPath string, cfg config.AssuranceConfig) *AssuranceValidator {
	return &AssuranceValidator{rootPath: rootPath, config: cfg}
}

func (v *AssuranceValidator) Validate() []models.ValidationResult {
	return []models.ValidationResult{validateAssurance(v.rootPath, v.config)}
}

func validateAssurance(rootPath string, cfg config.AssuranceConfig) models.ValidationResult {
	path := filepath.Join(rootPath, "assurance", "results.json")
	content, err := os.ReadFile(path)
	if err != nil {
		return models.ValidationResult{
			Capability: "Assurance",
			Status:     models.StatusFail,
			Message:    "missing results.json",
		}
	}

	var results metrics.AssuranceResults
	if err := json.Unmarshal(content, &results); err != nil {
		return models.ValidationResult{
			Capability: "Assurance",
			Status:     models.StatusFail,
			Message:    "invalid results.json",
		}
	}

	calculated, err := metrics.CalculateAssuranceMetrics(results)
	if err != nil {
		return models.ValidationResult{
			Capability: "Assurance",
			Status:     models.StatusFail,
			Message:    err.Error(),
		}
	}

	details := []string{
		formatFloatDetail("accuracy", calculated.Accuracy, cfg.MinAccuracy, "threshold"),
		formatIntDetail("scenarios_failed", calculated.ScenariosFailed, cfg.MaxFailures, "allowed"),
	}

	if calculated.ScenariosFailed > cfg.MaxFailures {
		return models.ValidationResult{
			Capability: "Assurance",
			Status:     models.StatusFail,
			Message:    "scenarios_failed above threshold",
			Details:    details,
		}
	}

	if calculated.Accuracy < cfg.MinAccuracy {
		return models.ValidationResult{
			Capability: "Assurance",
			Status:     models.StatusFail,
			Message:    "accuracy below threshold",
			Details:    details,
		}
	}

	return models.ValidationResult{
		Capability: "Assurance",
		Status:     models.StatusPass,
		Details:    details,
	}
}

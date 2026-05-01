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
	strict   bool
}

func NewAssuranceValidator(rootPath string, cfg config.AssuranceConfig, strictValues ...bool) *AssuranceValidator {
	return &AssuranceValidator{rootPath: rootPath, config: cfg, strict: strictValue(strictValues)}
}

func (v *AssuranceValidator) Validate() []models.ValidationResult {
	return []models.ValidationResult{validateAssurance(v.rootPath, v.config, v.strict)}
}

func validateAssurance(rootPath string, cfg config.AssuranceConfig, strictValues ...bool) models.ValidationResult {
	strict := strictValue(strictValues)
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

	details := append([]string{}, assuranceFailureDetails(results)...)
	details = append(details,
		formatFloatDetail("accuracy", calculated.Accuracy, cfg.MinAccuracy, "threshold"),
		formatIntDetail("scenarios_failed", calculated.ScenariosFailed, cfg.MaxFailures, "allowed"),
	)
	quality := evaluateJSONEvidenceQuality(rootPath, "assurance/results.json", "Assurance", content)

	if calculated.ScenariosFailed > cfg.MaxFailures {
		return models.ValidationResult{
			Capability:      "Assurance",
			Status:          models.StatusFail,
			Message:         "scenarios_failed above threshold",
			Details:         append(details, quality.Details...),
			EvidenceQuality: quality,
		}
	}

	if calculated.Accuracy < cfg.MinAccuracy {
		return models.ValidationResult{
			Capability:      "Assurance",
			Status:          models.StatusFail,
			Message:         "accuracy below threshold",
			Details:         append(details, quality.Details...),
			EvidenceQuality: quality,
		}
	}
	if quality.Score < 80 {
		result := jsonQualityWarningResult("Assurance", quality, strict)
		result.Details = append(details, result.Details...)
		return result
	}

	return models.ValidationResult{
		Capability:      "Assurance",
		Status:          models.StatusPass,
		Details:         details,
		EvidenceQuality: quality,
	}
}

func assuranceFailureDetails(results metrics.AssuranceResults) []string {
	if results.Failures == nil {
		return nil
	}

	details := make([]string, 0, len(*results.Failures))
	for _, failure := range *results.Failures {
		if failure == "" {
			continue
		}
		details = append(details, "failure: "+failure)
	}
	return details
}

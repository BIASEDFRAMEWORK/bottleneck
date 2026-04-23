package validator

import (
	"path/filepath"

	"biased/internal/models"
)

type Validator interface {
	Validate() []models.ValidationResult
}

type Engine struct {
	validators []Validator
}

func NewEngine(basePath string) *Engine {
	biasedRoot := filepath.Join(basePath, "biased")

	return &Engine{
		validators: []Validator{
			NewBehaviorValidator(biasedRoot),
			NewIntentValidator(biasedRoot),
			NewDesignValidator(biasedRoot),
			NewAssuranceValidator(biasedRoot),
			NewSecurityValidator(biasedRoot),
			NewExecutionValidator(biasedRoot),
		},
	}
}

func (e *Engine) Validate() models.EngineResult {
	results := make([]models.ValidationResult, 0, len(e.validators))
	systemStatus := models.StatusPass
	primaryBottleneck := "None"

	for _, validator := range e.validators {
		for _, result := range validator.Validate() {
			results = append(results, result)
			if systemStatus == models.StatusPass && result.Status == models.StatusFail {
				systemStatus = models.StatusFail
				primaryBottleneck = result.Capability
			}
		}
	}

	return models.EngineResult{
		Results:           results,
		SystemStatus:      systemStatus,
		PrimaryBottleneck: primaryBottleneck,
	}
}

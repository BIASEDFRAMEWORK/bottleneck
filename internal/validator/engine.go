package validator

import (
	"path/filepath"

	"biased/internal/config"
	"biased/internal/models"
)

type Validator interface {
	Validate() []models.ValidationResult
}

type Engine struct {
	validators    []Validator
	environment   string
	configLoadErr error
}

func NewEngine(basePath string, env string) *Engine {
	biasedRoot := filepath.Join(basePath, "biased")
	cfg, err := config.Load(filepath.Join(biasedRoot, "config.yaml"))
	if err != nil {
		return &Engine{
			environment:   env,
			configLoadErr: err,
		}
	}

	envConfig := config.ResolveEnvironment(cfg, env)

	return &Engine{
		environment: env,
		validators: []Validator{
			NewBehaviorValidator(biasedRoot),
			NewIntentValidator(biasedRoot),
			NewDesignValidator(biasedRoot),
			NewAssuranceValidator(biasedRoot, envConfig.Assurance),
			NewSecurityValidator(biasedRoot),
			NewExecutionValidator(biasedRoot, envConfig.Execution),
		},
	}
}

func (e *Engine) Validate() models.EngineResult {
	if e.configLoadErr != nil {
		return models.EngineResult{
			Results: []models.ValidationResult{{
				Capability: "Config",
				Status:     models.StatusFail,
				Message:    "missing or invalid config.yaml",
			}},
			SystemStatus:      models.StatusFail,
			PrimaryBottleneck: "Config",
			Environment:       e.environment,
		}
	}

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
		Environment:       e.environment,
	}
}

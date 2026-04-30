package validator

import (
	"path/filepath"

	"bottleneck/internal/config"
	"bottleneck/internal/models"
)

type Validator interface {
	Validate() []models.ValidationResult
}

type Engine struct {
	validators          []Validator
	environment         string
	effectiveThresholds models.EffectiveThresholds
	configLoadErr       error
}

type EngineOptions struct {
	Strict bool
}

type EngineOption func(*EngineOptions)

func WithStrictMode(strict bool) EngineOption {
	return func(options *EngineOptions) {
		options.Strict = strict
	}
}

func NewEngine(basePath string, env string, optionFuncs ...EngineOption) *Engine {
	options := EngineOptions{}
	for _, optionFunc := range optionFuncs {
		optionFunc(&options)
	}

	bottleneckRoot := filepath.Join(basePath, "bottleneck")
	cfg, err := config.Load(filepath.Join(bottleneckRoot, "config.yaml"))
	if err != nil {
		return &Engine{
			environment:   env,
			configLoadErr: err,
		}
	}

	envConfig := config.ResolveEnvironment(cfg, env)

	return &Engine{
		environment:         env,
		effectiveThresholds: effectiveThresholdsFromConfig(envConfig),
		validators: []Validator{
			NewBehaviorValidator(bottleneckRoot, options.Strict),
			NewIntentValidator(bottleneckRoot, options.Strict),
			NewDesignValidator(bottleneckRoot, options.Strict),
			NewAssuranceValidator(bottleneckRoot, envConfig.Assurance),
			NewSecurityValidator(bottleneckRoot),
			NewExecutionValidator(bottleneckRoot, envConfig.Execution),
			NewTraceabilityValidator(bottleneckRoot, env, options.Strict),
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
			SystemStatus:        models.StatusFail,
			PrimaryBottleneck:   "Config",
			Environment:         e.environment,
			EffectiveThresholds: e.effectiveThresholds,
		}
	}

	results := make([]models.ValidationResult, 0, len(e.validators))
	systemStatus := models.StatusPass
	primaryBottleneck := "None"

	for _, validator := range e.validators {
		for _, result := range validator.Validate() {
			results = append(results, result)

			if result.Status == models.StatusFail && systemStatus != models.StatusFail {
				systemStatus = models.StatusFail
				primaryBottleneck = result.Capability
				continue
			}

			if result.Status == models.StatusWarning && systemStatus == models.StatusPass {
				systemStatus = models.StatusWarning
				primaryBottleneck = result.Capability
			}
		}
	}

	return models.EngineResult{
		Results:             results,
		SystemStatus:        systemStatus,
		PrimaryBottleneck:   primaryBottleneck,
		Environment:         e.environment,
		EffectiveThresholds: e.effectiveThresholds,
	}
}

func effectiveThresholdsFromConfig(envConfig config.EnvironmentConfig) models.EffectiveThresholds {
	return models.EffectiveThresholds{
		Assurance: models.AssuranceThresholds{
			MinAccuracy: envConfig.Assurance.MinAccuracy,
			MaxFailures: envConfig.Assurance.MaxFailures,
		},
		Execution: models.ExecutionThresholds{
			MaxErrorRate: envConfig.Execution.MaxErrorRate,
			MinAdoption:  envConfig.Execution.MinAdoption,
		},
	}
}

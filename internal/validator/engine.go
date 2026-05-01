package validator

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

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
	configLoadDetails   []string
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
		message, details := configLoadGuidance(err)
		return &Engine{
			environment:       env,
			configLoadErr:     errConfigLoad(message, err),
			configLoadDetails: details,
		}
	}

	envConfig, err := config.ResolveEnvironmentStrict(cfg, env)
	if err != nil {
		return &Engine{
			environment:       env,
			configLoadErr:     err,
			configLoadDetails: environmentGuidanceDetails(cfg),
		}
	}

	return &Engine{
		environment:         env,
		effectiveThresholds: effectiveThresholdsFromConfig(envConfig),
		validators: []Validator{
			NewBehaviorValidator(bottleneckRoot, options.Strict),
			NewIntentValidator(bottleneckRoot, options.Strict),
			NewDesignValidator(bottleneckRoot, options.Strict),
			NewAssuranceValidator(bottleneckRoot, envConfig.Assurance, options.Strict),
			NewSecurityValidator(bottleneckRoot, envConfig.Security, options.Strict),
			NewExecutionValidator(bottleneckRoot, envConfig.Execution, options.Strict),
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
				Message:    e.configLoadErr.Error(),
				Details:    e.configLoadDetails,
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
			Telemetry: models.TelemetryThresholds{
				MaxAgeHours:           envConfig.Execution.Telemetry.MaxAgeHours,
				StaleAllowed:          envConfig.Execution.Telemetry.MaxAgeHours <= 0,
				MinDeploymentsPerWeek: envConfig.Execution.Telemetry.MinDeploymentsPerWeek,
				MaxChangeFailureRate:  envConfig.Execution.Telemetry.MaxChangeFailureRate,
				MaxErrorRate:          envConfig.Execution.Telemetry.MaxErrorRate,
				MaxUserOverrideRate:   envConfig.Execution.Telemetry.MaxUserOverrideRate,
				MinAdoptionRate:       envConfig.Execution.Telemetry.MinAdoptionRate,
				MaxBudgetVariance:     envConfig.Execution.Telemetry.MaxBudgetVariance,
			},
		},
		Security: models.SecurityThresholds{
			SARIF: models.SARIFThresholds{
				MaxCritical:           envConfig.Security.SARIF.MaxCritical,
				MaxHigh:               envConfig.Security.SARIF.MaxHigh,
				MaxMedium:             envConfig.Security.SARIF.MaxMedium,
				MaxLow:                envConfig.Security.SARIF.MaxLow,
				FailOnUnknownSeverity: envConfig.Security.SARIF.FailOnUnknownSeverity,
			},
		},
		Gate: models.GateThresholds{
			Release: models.ReleaseGateThresholds{
				MinPrimaryScore:     envConfig.Gate.Release.MinPrimaryScore,
				RequiredCategories:  append([]string{}, envConfig.Gate.Release.RequiredCategories...),
				RequireTraceability: envConfig.Gate.Release.RequireTraceability,
				RequireGovernance:   envConfig.Gate.Release.RequireGovernance,
			},
		},
	}
}

func configLoadGuidance(err error) (string, []string) {
	if errors.Is(err, os.ErrNotExist) {
		return "No Bottleneck config found.", []string{
			"Bottleneck has not been initialized in this directory.",
			"Next action: initialize a SaaS starter project:",
			"  bottleneck init --template saas",
		}
	}
	return "Invalid Bottleneck config.", []string{
		"bottleneck/config.yaml could not be parsed: " + err.Error(),
		"Next action: repair bottleneck/config.yaml or run `bottleneck init --template saas` in a new project for an example.",
	}
}

func environmentGuidanceDetails(cfg config.Config) []string {
	supported := config.SupportedEnvironments(cfg)
	display := make([]string, 0, len(supported))
	for _, env := range supported {
		if env == "default" {
			continue
		}
		display = append(display, env)
	}
	if len(display) == 0 {
		display = supported
	}
	return []string{
		"Next action: choose one of: " + strings.Join(display, ", ") + ".",
		"Example:",
		"  bottleneck scorecard --env=production",
	}
}

func errConfigLoad(message string, err error) error {
	return configLoadError{message: message, err: err}
}

type configLoadError struct {
	message string
	err     error
}

func (e configLoadError) Error() string {
	return e.message
}

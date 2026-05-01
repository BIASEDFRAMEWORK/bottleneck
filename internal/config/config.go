package config

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Environments map[string]EnvironmentConfig `yaml:"environments"`
}

type EnvironmentConfig struct {
	Assurance AssuranceConfig `yaml:"assurance"`
	Execution ExecutionConfig `yaml:"execution"`
	Security  SecurityConfig  `yaml:"security"`
	Gate      GateConfig      `yaml:"gate"`
}

type AssuranceConfig struct {
	MinAccuracy float64 `yaml:"min_accuracy"`
	MaxFailures int     `yaml:"max_failures"`

	minAccuracySet bool
	maxFailuresSet bool
}

type ExecutionConfig struct {
	MaxErrorRate float64         `yaml:"max_error_rate"`
	MinAdoption  float64         `yaml:"min_adoption"`
	Telemetry    TelemetryConfig `yaml:"telemetry"`

	maxErrorRateSet bool
	minAdoptionSet  bool
}

type SecurityConfig struct {
	SARIF SARIFConfig `yaml:"sarif"`
}

type SARIFConfig struct {
	MaxCritical           int  `yaml:"max_critical"`
	MaxHigh               int  `yaml:"max_high"`
	MaxMedium             int  `yaml:"max_medium"`
	MaxLow                int  `yaml:"max_low"`
	FailOnUnknownSeverity bool `yaml:"fail_on_unknown_severity"`

	maxCriticalSet           bool
	maxHighSet               bool
	maxMediumSet             bool
	maxLowSet                bool
	failOnUnknownSeveritySet bool
}

type TelemetryConfig struct {
	MaxAgeHours           int     `yaml:"max_age_hours"`
	MinDeploymentsPerWeek float64 `yaml:"min_deployments_per_week"`
	MaxChangeFailureRate  float64 `yaml:"max_change_failure_rate"`
	MaxErrorRate          float64 `yaml:"max_error_rate"`
	MaxUserOverrideRate   float64 `yaml:"max_user_override_rate"`
	MinAdoptionRate       float64 `yaml:"min_adoption_rate"`
	MaxBudgetVariance     float64 `yaml:"max_budget_variance"`

	maxAgeHoursSet           bool
	minDeploymentsPerWeekSet bool
	maxChangeFailureRateSet  bool
	maxErrorRateSet          bool
	maxUserOverrideRateSet   bool
	minAdoptionRateSet       bool
	maxBudgetVarianceSet     bool
}

type GateConfig struct {
	Release ReleaseGateConfig `yaml:"release"`
}

type ReleaseGateConfig struct {
	MinPrimaryScore     int      `yaml:"min_primary_score"`
	RequiredCategories  []string `yaml:"required_categories"`
	RequireTraceability bool     `yaml:"require_traceability"`
	RequireGovernance   bool     `yaml:"require_governance"`

	minPrimaryScoreSet     bool
	requiredCategoriesSet  bool
	requireTraceabilitySet bool
	requireGovernanceSet   bool
}

func Load(path string) (Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func ResolveEnvironment(cfg Config, env string) EnvironmentConfig {
	resolved, err := ResolveEnvironmentStrict(cfg, env)
	if err != nil {
		return resolveEnvironmentUnchecked(cfg, "default")
	}
	return resolved
}

func ResolveEnvironmentStrict(cfg Config, env string) (EnvironmentConfig, error) {
	env = strings.TrimSpace(env)
	if env == "" {
		env = "default"
	}
	if env != "default" {
		if _, ok := cfg.Environments[env]; !ok {
			return EnvironmentConfig{}, fmt.Errorf("unknown environment %q (supported: %s)", env, strings.Join(SupportedEnvironments(cfg), ", "))
		}
	}
	return resolveEnvironmentUnchecked(cfg, env), nil
}

func SupportedEnvironments(cfg Config) []string {
	seen := map[string]struct{}{"default": {}}
	for env := range cfg.Environments {
		if strings.TrimSpace(env) == "" {
			continue
		}
		seen[env] = struct{}{}
	}
	values := make([]string, 0, len(seen))
	for env := range seen {
		values = append(values, env)
	}
	sort.Strings(values)
	return values
}

func resolveEnvironmentUnchecked(cfg Config, env string) EnvironmentConfig {
	resolved := cfg.Environments["default"]
	resolved.Security.SARIF = mergeSARIF(defaultSARIFConfig(), resolved.Security.SARIF)
	resolved.Execution.Telemetry = mergeTelemetry(defaultTelemetryConfig(), resolved.Execution.Telemetry)
	resolved.Execution.Telemetry = syncTelemetryWithLegacyExecution(resolved.Execution)
	resolved.Gate.Release = mergeReleaseGate(defaultReleaseGateConfig(), resolved.Gate.Release)
	override, ok := cfg.Environments[env]
	if !ok || env == "default" {
		return resolved
	}

	if override.Assurance.minAccuracySet {
		resolved.Assurance.MinAccuracy = override.Assurance.MinAccuracy
	}
	if override.Assurance.maxFailuresSet {
		resolved.Assurance.MaxFailures = override.Assurance.MaxFailures
	}
	if override.Execution.maxErrorRateSet {
		resolved.Execution.MaxErrorRate = override.Execution.MaxErrorRate
	}
	if override.Execution.minAdoptionSet {
		resolved.Execution.MinAdoption = override.Execution.MinAdoption
	}
	resolved.Security.SARIF = mergeSARIF(resolved.Security.SARIF, override.Security.SARIF)
	resolved.Execution.Telemetry = mergeTelemetry(resolved.Execution.Telemetry, override.Execution.Telemetry)
	resolved.Execution.Telemetry = syncTelemetryWithLegacyExecution(resolved.Execution)
	if override.Execution.maxErrorRateSet && !override.Execution.Telemetry.maxErrorRateSet {
		resolved.Execution.Telemetry.MaxErrorRate = override.Execution.MaxErrorRate
	}
	if override.Execution.minAdoptionSet && !override.Execution.Telemetry.minAdoptionRateSet {
		resolved.Execution.Telemetry.MinAdoptionRate = override.Execution.MinAdoption
	}
	if override.Gate.Release.minPrimaryScoreSet {
		resolved.Gate.Release.MinPrimaryScore = override.Gate.Release.MinPrimaryScore
	}
	if override.Gate.Release.requiredCategoriesSet {
		resolved.Gate.Release.RequiredCategories = append([]string{}, override.Gate.Release.RequiredCategories...)
	}
	if override.Gate.Release.requireTraceabilitySet {
		resolved.Gate.Release.RequireTraceability = override.Gate.Release.RequireTraceability
	}
	if override.Gate.Release.requireGovernanceSet {
		resolved.Gate.Release.RequireGovernance = override.Gate.Release.RequireGovernance
	}

	return resolved
}

func defaultSARIFConfig() SARIFConfig {
	return SARIFConfig{
		MaxCritical: 0,
		MaxHigh:     0,
		MaxMedium:   5,
		MaxLow:      20,
	}
}

func defaultTelemetryConfig() TelemetryConfig {
	return TelemetryConfig{
		MaxAgeHours:           168,
		MinDeploymentsPerWeek: 1,
		MaxChangeFailureRate:  0.15,
		MaxErrorRate:          0.05,
		MaxUserOverrideRate:   0.10,
		MinAdoptionRate:       0.50,
		MaxBudgetVariance:     0.20,
	}
}

func mergeSARIF(base SARIFConfig, override SARIFConfig) SARIFConfig {
	merged := base
	if override.maxCriticalSet {
		merged.MaxCritical = override.MaxCritical
	}
	if override.maxHighSet {
		merged.MaxHigh = override.MaxHigh
	}
	if override.maxMediumSet {
		merged.MaxMedium = override.MaxMedium
	}
	if override.maxLowSet {
		merged.MaxLow = override.MaxLow
	}
	if override.failOnUnknownSeveritySet {
		merged.FailOnUnknownSeverity = override.FailOnUnknownSeverity
	}
	return merged
}

func mergeTelemetry(base TelemetryConfig, override TelemetryConfig) TelemetryConfig {
	merged := base
	if override.maxAgeHoursSet {
		merged.MaxAgeHours = override.MaxAgeHours
		merged.maxAgeHoursSet = true
	}
	if override.minDeploymentsPerWeekSet {
		merged.MinDeploymentsPerWeek = override.MinDeploymentsPerWeek
		merged.minDeploymentsPerWeekSet = true
	}
	if override.maxChangeFailureRateSet {
		merged.MaxChangeFailureRate = override.MaxChangeFailureRate
		merged.maxChangeFailureRateSet = true
	}
	if override.maxErrorRateSet {
		merged.MaxErrorRate = override.MaxErrorRate
		merged.maxErrorRateSet = true
	}
	if override.maxUserOverrideRateSet {
		merged.MaxUserOverrideRate = override.MaxUserOverrideRate
		merged.maxUserOverrideRateSet = true
	}
	if override.minAdoptionRateSet {
		merged.MinAdoptionRate = override.MinAdoptionRate
		merged.minAdoptionRateSet = true
	}
	if override.maxBudgetVarianceSet {
		merged.MaxBudgetVariance = override.MaxBudgetVariance
		merged.maxBudgetVarianceSet = true
	}
	return merged
}

func syncTelemetryWithLegacyExecution(cfg ExecutionConfig) TelemetryConfig {
	telemetry := cfg.Telemetry
	if cfg.maxErrorRateSet && !telemetry.maxErrorRateSet {
		telemetry.MaxErrorRate = cfg.MaxErrorRate
	}
	if cfg.minAdoptionSet && !telemetry.minAdoptionRateSet {
		telemetry.MinAdoptionRate = cfg.MinAdoption
	}
	return telemetry
}

func DefaultReleaseGateConfig() ReleaseGateConfig {
	return defaultReleaseGateConfig()
}

func defaultReleaseGateConfig() ReleaseGateConfig {
	return ReleaseGateConfig{
		MinPrimaryScore:     60,
		RequiredCategories:  []string{"Intent", "Behavior", "Assurance", "Security", "Execution"},
		RequireTraceability: true,
		RequireGovernance:   false,
	}
}

func mergeReleaseGate(base ReleaseGateConfig, override ReleaseGateConfig) ReleaseGateConfig {
	merged := base
	if override.minPrimaryScoreSet {
		merged.MinPrimaryScore = override.MinPrimaryScore
	}
	if override.requiredCategoriesSet {
		merged.RequiredCategories = append([]string{}, override.RequiredCategories...)
	}
	if override.requireTraceabilitySet {
		merged.RequireTraceability = override.RequireTraceability
	}
	if override.requireGovernanceSet {
		merged.RequireGovernance = override.RequireGovernance
	}
	return merged
}

func (c *AssuranceConfig) UnmarshalYAML(value *yaml.Node) error {
	type plain AssuranceConfig
	var decoded plain
	if err := value.Decode(&decoded); err != nil {
		return err
	}

	*c = AssuranceConfig(decoded)
	for i := 0; i < len(value.Content)-1; i += 2 {
		switch value.Content[i].Value {
		case "min_accuracy":
			c.minAccuracySet = true
		case "max_failures":
			c.maxFailuresSet = true
		}
	}

	return nil
}

func (c *ExecutionConfig) UnmarshalYAML(value *yaml.Node) error {
	type plain ExecutionConfig
	var decoded plain
	if err := value.Decode(&decoded); err != nil {
		return err
	}

	*c = ExecutionConfig(decoded)
	for i := 0; i < len(value.Content)-1; i += 2 {
		switch value.Content[i].Value {
		case "max_error_rate":
			c.maxErrorRateSet = true
		case "min_adoption":
			c.minAdoptionSet = true
		}
	}

	return nil
}

func (c *SARIFConfig) UnmarshalYAML(value *yaml.Node) error {
	type plain SARIFConfig
	var decoded plain
	if err := value.Decode(&decoded); err != nil {
		return err
	}

	*c = SARIFConfig(decoded)
	for i := 0; i < len(value.Content)-1; i += 2 {
		switch value.Content[i].Value {
		case "max_critical":
			c.maxCriticalSet = true
		case "max_high":
			c.maxHighSet = true
		case "max_medium":
			c.maxMediumSet = true
		case "max_low":
			c.maxLowSet = true
		case "fail_on_unknown_severity":
			c.failOnUnknownSeveritySet = true
		}
	}

	return nil
}

func (c *TelemetryConfig) UnmarshalYAML(value *yaml.Node) error {
	type plain TelemetryConfig
	var decoded plain
	if err := value.Decode(&decoded); err != nil {
		return err
	}

	*c = TelemetryConfig(decoded)
	for i := 0; i < len(value.Content)-1; i += 2 {
		switch value.Content[i].Value {
		case "max_age_hours":
			c.maxAgeHoursSet = true
		case "min_deployments_per_week":
			c.minDeploymentsPerWeekSet = true
		case "max_change_failure_rate":
			c.maxChangeFailureRateSet = true
		case "max_error_rate":
			c.maxErrorRateSet = true
		case "max_user_override_rate":
			c.maxUserOverrideRateSet = true
		case "min_adoption_rate":
			c.minAdoptionRateSet = true
		case "max_budget_variance":
			c.maxBudgetVarianceSet = true
		}
	}

	return nil
}

func (c *ReleaseGateConfig) UnmarshalYAML(value *yaml.Node) error {
	type plain ReleaseGateConfig
	var decoded plain
	if err := value.Decode(&decoded); err != nil {
		return err
	}

	*c = ReleaseGateConfig(decoded)
	for i := 0; i < len(value.Content)-1; i += 2 {
		switch value.Content[i].Value {
		case "min_primary_score":
			c.minPrimaryScoreSet = true
		case "required_categories":
			c.requiredCategoriesSet = true
		case "require_traceability":
			c.requireTraceabilitySet = true
		case "require_governance":
			c.requireGovernanceSet = true
		}
	}

	return nil
}

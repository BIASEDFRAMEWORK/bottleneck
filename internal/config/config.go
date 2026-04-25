package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Environments map[string]EnvironmentConfig `yaml:"environments"`
}

type EnvironmentConfig struct {
	Assurance AssuranceConfig `yaml:"assurance"`
	Execution ExecutionConfig `yaml:"execution"`
}

type AssuranceConfig struct {
	MinAccuracy float64 `yaml:"min_accuracy"`
	MaxFailures int     `yaml:"max_failures"`

	minAccuracySet bool
	maxFailuresSet bool
}

type ExecutionConfig struct {
	MaxErrorRate float64 `yaml:"max_error_rate"`
	MinAdoption  float64 `yaml:"min_adoption"`

	maxErrorRateSet bool
	minAdoptionSet  bool
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
	resolved := cfg.Environments["default"]
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

	return resolved
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

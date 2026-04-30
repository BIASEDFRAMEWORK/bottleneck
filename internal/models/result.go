package models

const (
	StatusPass    = "PASS"
	StatusFail    = "FAIL"
	StatusWarning = "WARNING"
)

type ValidationResult struct {
	Capability string
	Status     string
	Message    string
	Details    []string
	Findings   []ValidationFinding
}

type ValidationFinding struct {
	Level   string `json:"level,omitempty"`
	Message string `json:"message,omitempty"`
	Path    string `json:"path,omitempty"`
	Line    int    `json:"line,omitempty"`
}

type EngineResult struct {
	Results             []ValidationResult
	SystemStatus        string
	PrimaryBottleneck   string
	Environment         string
	EffectiveThresholds EffectiveThresholds
}

type EffectiveThresholds struct {
	Assurance AssuranceThresholds `json:"assurance"`
	Execution ExecutionThresholds `json:"execution"`
}

type AssuranceThresholds struct {
	MinAccuracy float64 `json:"min_accuracy"`
	MaxFailures int     `json:"max_failures"`
}

type ExecutionThresholds struct {
	MaxErrorRate float64 `json:"max_error_rate"`
	MinAdoption  float64 `json:"min_adoption"`
}

package models

const (
	StatusPass    = "PASS"
	StatusFail    = "FAIL"
	StatusWarning = "WARNING"
)

type ValidationResult struct {
	Capability      string
	Status          string
	Message         string
	Details         []string
	Findings        []ValidationFinding
	EvidenceQuality EvidenceQuality
}

type ValidationFinding struct {
	Level   string `json:"level,omitempty"`
	Message string `json:"message,omitempty"`
	Path    string `json:"path,omitempty"`
	Line    int    `json:"line,omitempty"`
}

type EvidenceQuality struct {
	Score        int           `json:"score,omitempty"`
	Details      []string      `json:"details,omitempty"`
	Missing      []string      `json:"missing,omitempty"`
	ScoreImpacts []ScoreImpact `json:"score_impacts,omitempty"`
}

type ScoreImpact struct {
	Reason string `json:"reason"`
	Delta  int    `json:"delta"`
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
	Security  SecurityThresholds  `json:"security"`
}

type AssuranceThresholds struct {
	MinAccuracy float64 `json:"min_accuracy"`
	MaxFailures int     `json:"max_failures"`
}

type ExecutionThresholds struct {
	MaxErrorRate float64             `json:"max_error_rate"`
	MinAdoption  float64             `json:"min_adoption"`
	Telemetry    TelemetryThresholds `json:"telemetry"`
}

type TelemetryThresholds struct {
	MaxAgeHours           int     `json:"max_age_hours"`
	MinDeploymentsPerWeek float64 `json:"min_deployments_per_week"`
	MaxChangeFailureRate  float64 `json:"max_change_failure_rate"`
	MaxErrorRate          float64 `json:"max_error_rate"`
	MaxUserOverrideRate   float64 `json:"max_user_override_rate"`
	MinAdoptionRate       float64 `json:"min_adoption_rate"`
	MaxBudgetVariance     float64 `json:"max_budget_variance"`
}

type SecurityThresholds struct {
	SARIF SARIFThresholds `json:"sarif"`
}

type SARIFThresholds struct {
	MaxCritical           int  `json:"max_critical"`
	MaxHigh               int  `json:"max_high"`
	MaxMedium             int  `json:"max_medium"`
	MaxLow                int  `json:"max_low"`
	FailOnUnknownSeverity bool `json:"fail_on_unknown_severity"`
}

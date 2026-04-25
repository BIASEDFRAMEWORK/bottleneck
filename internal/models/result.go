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
}

type EngineResult struct {
	Results           []ValidationResult
	SystemStatus      string
	PrimaryBottleneck string
	Environment       string
}

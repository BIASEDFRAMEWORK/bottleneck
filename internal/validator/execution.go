package validator

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"bottleneck/internal/config"
	"bottleneck/internal/models"
)

type ExecutionValidator struct {
	rootPath string
	config   config.ExecutionConfig
	strict   bool
}

type executionFile struct {
	GeneratedAt         string               `json:"generated_at"`
	Window              *executionWindow     `json:"window"`
	DeploymentFrequency *deploymentFrequency `json:"deployment_frequency"`
	ChangeFailureRate   *float64             `json:"change_failure_rate"`
	ErrorRate           *float64             `json:"error_rate"`
	UserOverrideRate    *float64             `json:"user_override_rate"`
	AdoptionRate        *float64             `json:"adoption_rate"`
	Cost                *executionCost       `json:"cost"`
}

type executionWindow struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type deploymentFrequency struct {
	Deployments int     `json:"deployments"`
	PeriodDays  float64 `json:"period_days"`
}

type executionCost struct {
	Total          *float64 `json:"total"`
	TotalUSD       *float64 `json:"total_usd"`
	CostPerRequest *float64 `json:"cost_per_request"`
	UnitCostUSD    *float64 `json:"unit_cost_usd"`
	Budget         *float64 `json:"budget"`
	BudgetVariance *float64 `json:"budget_variance"`
	Trend          string   `json:"trend"`
	Currency       string   `json:"currency"`
}

func NewExecutionValidator(rootPath string, cfg config.ExecutionConfig, strictValues ...bool) *ExecutionValidator {
	return &ExecutionValidator{rootPath: rootPath, config: cfg, strict: strictValue(strictValues)}
}

func (v *ExecutionValidator) Validate() []models.ValidationResult {
	return []models.ValidationResult{validateExecution(v.rootPath, v.config, v.strict)}
}

func validateExecution(rootPath string, cfg config.ExecutionConfig, strictValues ...bool) models.ValidationResult {
	strict := strictValue(strictValues)
	path := filepath.Join(rootPath, "execution", "telemetry.json")
	content, err := os.ReadFile(path)
	if err != nil {
		return models.ValidationResult{
			Capability: "Execution",
			Status:     models.StatusFail,
			Message:    "missing telemetry.json",
		}
	}

	var data executionFile
	if err := json.Unmarshal(content, &data); err != nil || data.AdoptionRate == nil || data.ErrorRate == nil {
		return models.ValidationResult{
			Capability: "Execution",
			Status:     models.StatusFail,
			Message:    "invalid telemetry.json",
		}
	}

	details := []string{
		formatFloatDetail("error_rate", normalizeRate(*data.ErrorRate), cfg.Telemetry.MaxErrorRate, "max"),
		formatFloatDetail("adoption_rate", normalizeRate(*data.AdoptionRate), cfg.Telemetry.MinAdoptionRate, "min"),
	}
	quality := evaluateJSONEvidenceQuality(rootPath, "execution/telemetry.json", "Execution", content)

	failures, warnings := executionTelemetryFindings(data, cfg.Telemetry)
	details = append(details, telemetryDetails(data, cfg.Telemetry)...)

	if len(failures) > 0 {
		return models.ValidationResult{
			Capability:      "Execution",
			Status:          models.StatusFail,
			Message:         "telemetry outside release thresholds",
			Details:         append(append(details, failures...), quality.Details...),
			EvidenceQuality: quality,
		}
	}

	if len(warnings) > 0 {
		return models.ValidationResult{
			Capability:      "Execution",
			Status:          models.StatusWarning,
			Message:         "telemetry evidence is incomplete",
			Details:         append(append(details, warnings...), quality.Details...),
			EvidenceQuality: quality,
		}
	}
	if quality.Score < 80 {
		result := jsonQualityWarningResult("Execution", quality, strict)
		result.Details = append(details, result.Details...)
		return result
	}

	return models.ValidationResult{
		Capability:      "Execution",
		Status:          models.StatusPass,
		Details:         details,
		EvidenceQuality: quality,
	}
}

func executionTelemetryFindings(data executionFile, cfg config.TelemetryConfig) ([]string, []string) {
	var failures []string
	var warnings []string
	extended := usesExtendedTelemetry(data)

	if data.ErrorRate != nil && normalizeRate(*data.ErrorRate) > cfg.MaxErrorRate {
		failures = append(failures, "error_rate above threshold")
	}
	if data.AdoptionRate != nil && normalizeRate(*data.AdoptionRate) < cfg.MinAdoptionRate {
		warnings = append(warnings, "adoption_rate below threshold")
	}
	if extended && data.ChangeFailureRate == nil {
		warnings = append(warnings, "change_failure_rate telemetry missing")
	} else if data.ChangeFailureRate != nil && normalizeRate(*data.ChangeFailureRate) > cfg.MaxChangeFailureRate {
		failures = append(failures, "change_failure_rate above threshold")
	}
	if extended && data.UserOverrideRate == nil {
		warnings = append(warnings, "user_override_rate telemetry missing")
	} else if data.UserOverrideRate != nil && normalizeRate(*data.UserOverrideRate) > cfg.MaxUserOverrideRate {
		warnings = append(warnings, "user_override_rate above threshold")
	}
	if extended && data.DeploymentFrequency == nil {
		warnings = append(warnings, "deployment_frequency telemetry missing")
	} else if data.DeploymentFrequency != nil && deploymentsPerWeek(*data.DeploymentFrequency) < cfg.MinDeploymentsPerWeek {
		warnings = append(warnings, "deployment_frequency below threshold")
	}
	if extended && data.GeneratedAt == "" {
		warnings = append(warnings, "generated_at telemetry timestamp missing")
	} else if data.GeneratedAt != "" {
		generatedAt, err := time.Parse(time.RFC3339, data.GeneratedAt)
		if err != nil {
			failures = append(failures, "generated_at telemetry timestamp is invalid")
		} else if telemetryIsStale(generatedAt.UTC(), cfg.MaxAgeHours) {
			warnings = append(warnings, "telemetry is stale")
		}
	}
	if extended && data.Cost == nil {
		warnings = append(warnings, "cost telemetry missing")
	} else if budgetVariance(data.Cost) > cfg.MaxBudgetVariance {
		warnings = append(warnings, "cost budget variance above threshold")
	}

	return failures, warnings
}

func usesExtendedTelemetry(data executionFile) bool {
	return data.GeneratedAt != "" ||
		data.Window != nil ||
		data.DeploymentFrequency != nil ||
		data.ChangeFailureRate != nil ||
		data.UserOverrideRate != nil ||
		data.Cost != nil
}

func telemetryDetails(data executionFile, cfg config.TelemetryConfig) []string {
	var details []string
	if data.ChangeFailureRate != nil {
		details = append(details, formatFloatDetail("change_failure_rate", normalizeRate(*data.ChangeFailureRate), cfg.MaxChangeFailureRate, "max"))
	}
	if data.UserOverrideRate != nil {
		details = append(details, formatFloatDetail("user_override_rate", normalizeRate(*data.UserOverrideRate), cfg.MaxUserOverrideRate, "max"))
	}
	if data.DeploymentFrequency != nil {
		details = append(details, formatFloatDetail("deployments_per_week", deploymentsPerWeek(*data.DeploymentFrequency), cfg.MinDeploymentsPerWeek, "min"))
	}
	if data.Cost != nil {
		details = append(details, formatFloatDetail("budget_variance", budgetVariance(data.Cost), cfg.MaxBudgetVariance, "max"))
	}
	if data.GeneratedAt != "" {
		details = append(details, "generated_at: "+data.GeneratedAt)
	}
	return details
}

func deploymentsPerWeek(frequency deploymentFrequency) float64 {
	if frequency.PeriodDays <= 0 {
		return 0
	}
	return float64(frequency.Deployments) / frequency.PeriodDays * 7
}

func normalizeRate(value float64) float64 {
	absolute := value
	if absolute < 0 {
		absolute = -absolute
	}
	if absolute > 1 && absolute <= 100 {
		return value / 100
	}
	return value
}

func telemetryIsStale(generatedAt time.Time, maxAgeHours int) bool {
	if maxAgeHours <= 0 {
		return false
	}
	elapsed := time.Since(generatedAt)
	if elapsed <= 0 {
		return false
	}
	maxDurationHours := int64(1<<63-1) / int64(time.Hour)
	if int64(maxAgeHours) > maxDurationHours {
		return false
	}
	return elapsed > time.Duration(maxAgeHours)*time.Hour
}

func budgetVariance(cost *executionCost) float64 {
	if cost == nil {
		return 0
	}
	if cost.BudgetVariance != nil {
		variance := normalizeRate(*cost.BudgetVariance)
		if variance < 0 {
			return -variance
		}
		return variance
	}
	total := cost.Total
	if total == nil {
		total = cost.TotalUSD
	}
	if total == nil || cost.Budget == nil || *cost.Budget == 0 {
		return 0
	}
	variance := (*total - *cost.Budget) / *cost.Budget
	if variance < 0 {
		return -variance
	}
	return variance
}

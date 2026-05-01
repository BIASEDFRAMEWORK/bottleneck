package validator

import (
	"encoding/json"
	"os"
	"path/filepath"

	"bottleneck/internal/config"
	"bottleneck/internal/models"
)

type SecurityValidator struct {
	rootPath string
	config   config.SecurityConfig
	strict   bool
}

type securityFile struct {
	Violations *int           `json:"violations"`
	Findings   map[string]int `json:"findings"`
}

func NewSecurityValidator(rootPath string, cfg config.SecurityConfig, strictValues ...bool) *SecurityValidator {
	return &SecurityValidator{rootPath: rootPath, config: cfg, strict: strictValue(strictValues)}
}

func (v *SecurityValidator) Validate() []models.ValidationResult {
	return []models.ValidationResult{validateSecurity(v.rootPath, v.config, v.strict)}
}

func validateSecurity(rootPath string, cfg config.SecurityConfig, strictValues ...bool) models.ValidationResult {
	strict := strictValue(strictValues)
	path := filepath.Join(rootPath, "security", "guardrails.json")
	content, err := os.ReadFile(path)
	if err != nil {
		return artifactReadErrorResult(rootPath, "Security", "security/guardrails.json", err)
	}

	var data securityFile
	if err := json.Unmarshal(content, &data); err != nil || data.Violations == nil {
		return models.ValidationResult{
			Capability: "Security",
			Status:     models.StatusFail,
			Message:    "invalid guardrails.json",
		}
	}

	quality := evaluateJSONEvidenceQuality(rootPath, "security/guardrails.json", "Security", content)
	if len(data.Findings) > 0 {
		return validateSARIFSecurity(data, cfg.SARIF, quality, strict)
	}

	if *data.Violations > 0 {
		return models.ValidationResult{
			Capability:      "Security",
			Status:          models.StatusFail,
			Message:         "violations detected",
			EvidenceQuality: quality,
			Details: []string{
				formatIntDetail("violations", *data.Violations, 0, "allowed"),
			},
		}
	}

	if quality.Score < 80 {
		result := jsonQualityWarningResult("Security", quality, strict)
		result.Details = append([]string{formatIntDetail("violations", *data.Violations, 0, "allowed")}, result.Details...)
		return result
	}

	return models.ValidationResult{
		Capability:      "Security",
		Status:          models.StatusPass,
		EvidenceQuality: quality,
		Details: []string{
			formatIntDetail("violations", *data.Violations, 0, "allowed"),
		},
	}
}

func validateSARIFSecurity(data securityFile, cfg config.SARIFConfig, quality models.EvidenceQuality, strict bool) models.ValidationResult {
	details := []string{
		formatIntDetail("critical_findings", data.Findings["critical"], cfg.MaxCritical, "max"),
		formatIntDetail("high_findings", data.Findings["high"], cfg.MaxHigh, "max"),
		formatIntDetail("medium_findings", data.Findings["medium"], cfg.MaxMedium, "max"),
		formatIntDetail("low_findings", data.Findings["low"], cfg.MaxLow, "max"),
		formatIntDetail("unknown_findings", data.Findings["unknown"], 0, "allowed when fail_on_unknown_severity"),
	}

	failReasons := sarifThresholdFailures(data.Findings, cfg)
	if len(failReasons) > 0 {
		return models.ValidationResult{
			Capability:      "Security",
			Status:          models.StatusFail,
			Message:         "sarif findings exceed thresholds",
			Details:         append(append(details, failReasons...), quality.Details...),
			EvidenceQuality: quality,
		}
	}

	if data.Findings["unknown"] > 0 {
		status := models.StatusWarning
		message := "sarif findings include unknown severity"
		if strict {
			status = models.StatusFail
			message = "sarif findings exceed thresholds"
		}
		return models.ValidationResult{
			Capability:      "Security",
			Status:          status,
			Message:         message,
			Details:         append(details, quality.Details...),
			EvidenceQuality: quality,
		}
	}

	if quality.Score < 80 {
		result := jsonQualityWarningResult("Security", quality, strict)
		result.Details = append(details, result.Details...)
		return result
	}

	return models.ValidationResult{
		Capability:      "Security",
		Status:          models.StatusPass,
		Details:         details,
		EvidenceQuality: quality,
	}
}

func sarifThresholdFailures(findings map[string]int, cfg config.SARIFConfig) []string {
	var failures []string
	if findings["critical"] > cfg.MaxCritical {
		failures = append(failures, "critical SARIF findings above threshold")
	}
	if findings["high"] > cfg.MaxHigh {
		failures = append(failures, "high SARIF findings above threshold")
	}
	if findings["medium"] > cfg.MaxMedium {
		failures = append(failures, "medium SARIF findings above threshold")
	}
	if findings["low"] > cfg.MaxLow {
		failures = append(failures, "low SARIF findings above threshold")
	}
	if cfg.FailOnUnknownSeverity && findings["unknown"] > 0 {
		failures = append(failures, "unknown SARIF severity above threshold")
	}
	return failures
}

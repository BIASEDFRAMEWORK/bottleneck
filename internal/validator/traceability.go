package validator

import (
	"strings"

	"bottleneck/internal/models"
	"bottleneck/internal/traceability"
)

type TraceabilityValidator struct {
	rootPath    string
	environment string
	strict      bool
}

func NewTraceabilityValidator(rootPath string, environment string, strict bool) *TraceabilityValidator {
	return &TraceabilityValidator{rootPath: rootPath, environment: environment, strict: strict}
}

func (v *TraceabilityValidator) Validate() []models.ValidationResult {
	return []models.ValidationResult{validateTraceability(v.rootPath, v.environment, v.strict)}
}

func validateTraceability(rootPath string, environment string, strict bool) models.ValidationResult {
	graph, err := traceability.Build(rootPath, traceability.Options{
		Environment: environment,
		Strict:      strict,
	})
	if err != nil {
		return models.ValidationResult{
			Capability: "Traceability",
			Status:     models.StatusFail,
			Message:    "traceability graph could not be parsed",
			Details:    []string{err.Error()},
		}
	}

	findings := graph.ValidateFindings()
	if len(findings) == 0 {
		if len(graph.OrderedIDs) == 0 {
			status := models.StatusWarning
			message := "traceability evidence IDs missing"
			if strict {
				status = models.StatusFail
				message = "traceability failures detected"
			}
			quality := models.EvidenceQuality{
				Score:   50,
				Details: []string{"no traceability evidence IDs found"},
				Missing: []string{"Add evidence IDs and refs across intent, behavior, assurance, security, and execution artifacts."},
				ScoreImpacts: []models.ScoreImpact{{
					Reason: "no traceability evidence IDs found",
					Delta:  -50,
				}},
			}
			return models.ValidationResult{
				Capability:      "Traceability",
				Status:          status,
				Message:         message,
				Details:         quality.Details,
				EvidenceQuality: quality,
			}
		}
		return models.ValidationResult{
			Capability: "Traceability",
			Status:     models.StatusPass,
			Details:    traceabilityPassDetails(graph),
		}
	}

	status := models.StatusWarning
	message := "traceability warnings detected"
	for _, finding := range findings {
		if finding.Severity == traceability.SeverityFail {
			status = models.StatusFail
			message = "traceability failures detected"
			break
		}
	}

	quality := traceabilityEvidenceQuality(findings)
	return models.ValidationResult{
		Capability:      "Traceability",
		Status:          status,
		Message:         message,
		Details:         traceabilityFindingDetails(findings),
		Findings:        traceabilityValidationFindings(findings),
		EvidenceQuality: quality,
	}
}

func traceabilityPassDetails(graph traceability.Graph) []string {
	if len(graph.OrderedIDs) == 0 {
		return []string{"no traceability evidence IDs found"}
	}

	return []string{
		formatIntDetail("traceability_nodes", len(graph.OrderedIDs), 1, "minimum"),
	}
}

func traceabilityFindingDetails(findings []traceability.Finding) []string {
	details := make([]string, 0, len(findings))
	for _, finding := range findings {
		details = append(details, finding.Message)
	}
	return details
}

func traceabilityEvidenceQuality(findings []traceability.Finding) models.EvidenceQuality {
	quality := models.EvidenceQuality{Score: 100}
	for _, finding := range findings {
		delta := traceabilityPenalty(finding)
		quality.Details = append(quality.Details, finding.Message)
		quality.ScoreImpacts = append(quality.ScoreImpacts, models.ScoreImpact{
			Reason: finding.Message,
			Delta:  delta,
		})
		if finding.Severity == traceability.SeverityFail {
			quality.Missing = append(quality.Missing, "Repair broken or invalid evidence references.")
		}
		quality.Score += delta
	}
	if quality.Score < 0 {
		quality.Score = 0
	}
	return quality
}

func traceabilityPenalty(finding traceability.Finding) int {
	if finding.Severity == traceability.SeverityFail {
		return -30
	}
	if strings.HasPrefix(finding.SourceID, "BEHAVIOR-") &&
		(strings.Contains(finding.Message, "no assurance result references") ||
			strings.Contains(finding.Message, "not linked to assurance evidence") ||
			strings.Contains(finding.Message, "not linked to intent evidence")) {
		return -25
	}
	return -15
}

func traceabilityValidationFindings(findings []traceability.Finding) []models.ValidationFinding {
	validationFindings := make([]models.ValidationFinding, 0, len(findings))
	for _, finding := range findings {
		level := "warning"
		if finding.Severity == traceability.SeverityFail {
			level = "error"
		}
		validationFindings = append(validationFindings, models.ValidationFinding{
			Level:   level,
			Message: finding.Message,
			Path:    finding.ArtifactPath,
		})
	}
	return validationFindings
}

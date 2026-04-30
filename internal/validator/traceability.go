package validator

import (
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

	return models.ValidationResult{
		Capability: "Traceability",
		Status:     status,
		Message:    message,
		Details:    traceabilityFindingDetails(findings),
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

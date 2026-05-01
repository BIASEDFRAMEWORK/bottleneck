package validator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"bottleneck/internal/models"
)

func missingArtifactResult(rootPath string, capability string, relativePath string) models.ValidationResult {
	artifactPath := contentQualityArtifactPath(rootPath, relativePath)
	details := missingArtifactDetails(rootPath, relativePath, artifactPath)

	if capability == "Assurance" {
		details = append(details,
			"Next action: Add test evidence manually or run:",
			"  bottleneck ingest cucumber --file reports/cucumber.json",
		)
		return models.ValidationResult{
			Capability: capability,
			Status:     models.StatusFail,
			Message:    "No assurance evidence found.",
			Details:    details,
		}
	}

	details = append(details,
		fmt.Sprintf("Next action: run bottleneck init --template saas or add %s.", artifactPath),
	)
	return models.ValidationResult{
		Capability: capability,
		Status:     models.StatusFail,
		Message:    missingArtifactMessage(relativePath),
		Details:    details,
	}
}

func artifactReadErrorResult(rootPath string, capability string, relativePath string, err error) models.ValidationResult {
	if os.IsNotExist(err) {
		return missingArtifactResult(rootPath, capability, relativePath)
	}

	artifactPath := contentQualityArtifactPath(rootPath, relativePath)
	return models.ValidationResult{
		Capability: capability,
		Status:     models.StatusFail,
		Message:    "could not read " + filepath.Base(relativePath),
		Details: []string{
			artifactPath + " could not be read: " + err.Error(),
			"Next action: check the file path and permissions, then rerun bottleneck validate.",
		},
	}
}

func missingArtifactDetails(rootPath string, relativePath string, artifactPath string) []string {
	details := []string{}
	relativeDir := filepath.Dir(relativePath)
	if relativeDir != "." {
		fullDir := filepath.Join(rootPath, relativeDir)
		if _, err := os.Stat(fullDir); os.IsNotExist(err) {
			details = append(details, "Missing evidence directory: "+filepath.ToSlash(filepath.Join(filepath.Base(rootPath), relativeDir)))
		}
	}
	details = append(details, "Expected file: "+artifactPath)
	return details
}

func missingArtifactMessage(relativePath string) string {
	switch filepath.ToSlash(relativePath) {
	case "behavior/behavior-spec.md":
		return "missing behavior-spec.md"
	case "intent/intent.md":
		return "missing intent.md"
	case "design/architecture.md":
		return "missing architecture.md"
	case "security/guardrails.json":
		return "missing guardrails.json"
	case "execution/telemetry.json":
		return "missing telemetry.json"
	default:
		return "missing " + filepath.Base(relativePath)
	}
}

func placeholderGuidanceDetails(details []string) []string {
	if !hasPlaceholderHeavyDetail(details) {
		return details
	}
	return append(details,
		"Placeholder content does not support release readiness.",
		"Next action: replace it with real SaaS evidence or run:",
		"  bottleneck init --template saas",
	)
}

func hasPlaceholderHeavyDetail(details []string) bool {
	for _, detail := range details {
		if strings.Contains(detail, "placeholder-heavy") {
			return true
		}
	}
	return false
}

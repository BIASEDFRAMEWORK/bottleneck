package validator

import (
	"encoding/json"
	"fmt"
	"strings"

	"bottleneck/internal/models"
)

type evidenceEnvelope struct {
	Evidence []evidenceEntry `json:"evidence"`
}

type evidenceEntry struct {
	ID     string   `json:"id"`
	Refs   []string `json:"refs"`
	Source string   `json:"source"`
	Status string   `json:"status"`
}

func evaluateJSONEvidenceQuality(rootPath string, relativePath string, capability string, content []byte) models.EvidenceQuality {
	artifactPath := contentQualityArtifactPath(rootPath, relativePath)
	quality := models.EvidenceQuality{Score: 100}

	var envelope evidenceEnvelope
	if err := json.Unmarshal(content, &envelope); err != nil {
		addQualityIssue(&quality, fmt.Sprintf("%s evidence quality could not inspect JSON evidence array", artifactPath), -10)
		return quality
	}

	expectedPrefix := strings.ToUpper(capability)
	if len(envelope.Evidence) == 0 {
		addQualityIssue(&quality, fmt.Sprintf("%s does not include %s-* evidence IDs", artifactPath, expectedPrefix), -15)
		quality.Missing = append(quality.Missing, fmt.Sprintf("Add an evidence array with %s-* IDs and refs.", expectedPrefix))
		return finalizeQuality(quality)
	}

	hasExpectedID := false
	hasRefs := false
	hasSourceOrStatus := false
	for _, evidence := range envelope.Evidence {
		if strings.HasPrefix(evidence.ID, expectedPrefix+"-") {
			hasExpectedID = true
		}
		if len(evidence.Refs) > 0 {
			hasRefs = true
		}
		if strings.TrimSpace(evidence.Source) != "" || strings.TrimSpace(evidence.Status) != "" {
			hasSourceOrStatus = true
		}
	}

	if !hasExpectedID {
		addQualityIssue(&quality, fmt.Sprintf("%s does not include %s-* evidence IDs", artifactPath, expectedPrefix), -20)
		quality.Missing = append(quality.Missing, fmt.Sprintf("Add %s-* IDs to JSON evidence entries.", expectedPrefix))
	}
	if !hasRefs {
		addQualityIssue(&quality, fmt.Sprintf("%s evidence IDs do not reference related release evidence", artifactPath), -15)
		quality.Missing = append(quality.Missing, "Add refs from JSON evidence entries to behavior, intent, or assurance IDs.")
	}
	if !hasSourceOrStatus {
		addQualityIssue(&quality, fmt.Sprintf("%s evidence entries do not include source or status context", artifactPath), -10)
	}

	return finalizeQuality(quality)
}

func jsonQualityWarningResult(capability string, quality models.EvidenceQuality, strict bool) models.ValidationResult {
	status := models.StatusWarning
	message := fmt.Sprintf("%s evidence quality is weak", strings.ToLower(capability))
	if strict {
		status = models.StatusFail
		message = contentQualityStrictFailMessage
	}
	return models.ValidationResult{
		Capability:      capability,
		Status:          status,
		Message:         message,
		Details:         quality.Details,
		Findings:        findingsForQuality(status, quality),
		EvidenceQuality: quality,
	}
}

func finalizeQuality(quality models.EvidenceQuality) models.EvidenceQuality {
	quality.Score = clampQualityScore(quality.Score)
	quality.Details = uniqueStrings(quality.Details)
	quality.Missing = uniqueStrings(quality.Missing)
	quality.ScoreImpacts = uniqueImpacts(quality.ScoreImpacts)
	return quality
}

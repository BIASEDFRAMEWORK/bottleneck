package validator

import (
	"encoding/json"
	"os"
	"path/filepath"

	"biased/internal/models"
)

type AssuranceValidator struct {
	rootPath string
}

type assuranceFile struct {
	ScenariosTotal  *int      `json:"scenarios_total"`
	ScenariosPassed *int      `json:"scenarios_passed"`
	ScenariosFailed *int      `json:"scenarios_failed"`
	Accuracy        *float64  `json:"accuracy"`
	Failures        *[]string `json:"failures"`
}

func NewAssuranceValidator(rootPath string) *AssuranceValidator {
	return &AssuranceValidator{rootPath: rootPath}
}

func (v *AssuranceValidator) Validate() []models.ValidationResult {
	path := filepath.Join(v.rootPath, "assurance", "results.json")
	content, err := os.ReadFile(path)
	if err != nil {
		return []models.ValidationResult{{
			Capability: "Assurance",
			Status:     models.StatusFail,
			Message:    "missing results.json",
		}}
	}

	var data assuranceFile
	if err := json.Unmarshal(content, &data); err != nil {
		return []models.ValidationResult{{
			Capability: "Assurance",
			Status:     models.StatusFail,
			Message:    "invalid results.json",
		}}
	}

	if data.ScenariosTotal == nil ||
		data.ScenariosPassed == nil ||
		data.ScenariosFailed == nil ||
		data.Accuracy == nil ||
		data.Failures == nil {
		return []models.ValidationResult{{
			Capability: "Assurance",
			Status:     models.StatusFail,
			Message:    "required fields missing",
		}}
	}

	if *data.ScenariosFailed > 0 {
		return []models.ValidationResult{{
			Capability: "Assurance",
			Status:     models.StatusFail,
			Message:    "scenarios failed",
		}}
	}

	if *data.Accuracy < 0.90 {
		return []models.ValidationResult{{
			Capability: "Assurance",
			Status:     models.StatusFail,
			Message:    "accuracy below threshold",
		}}
	}

	return []models.ValidationResult{{
		Capability: "Assurance",
		Status:     models.StatusPass,
	}}
}

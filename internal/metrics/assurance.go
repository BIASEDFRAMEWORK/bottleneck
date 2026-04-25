package metrics

import "fmt"

type AssuranceResults struct {
	ScenariosTotal  *int      `json:"scenarios_total"`
	ScenariosPassed *int      `json:"scenarios_passed"`
	ScenariosFailed *int      `json:"scenarios_failed"`
	Failures        *[]string `json:"failures"`
}

type AssuranceMetrics struct {
	Accuracy        float64
	ScenariosFailed int
}

func CalculateAssuranceMetrics(results AssuranceResults) (AssuranceMetrics, error) {
	if results.ScenariosTotal == nil ||
		results.ScenariosPassed == nil ||
		results.ScenariosFailed == nil ||
		results.Failures == nil {
		return AssuranceMetrics{}, fmt.Errorf("required fields missing")
	}

	if *results.ScenariosTotal <= 0 {
		return AssuranceMetrics{}, fmt.Errorf("scenarios_total must be greater than 0")
	}

	return AssuranceMetrics{
		Accuracy:        float64(*results.ScenariosPassed) / float64(*results.ScenariosTotal),
		ScenariosFailed: *results.ScenariosFailed,
	}, nil
}

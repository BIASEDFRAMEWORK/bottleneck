package explainer

import (
	"fmt"
	"strings"

	"bottleneck/internal/models"
)

type capabilityMetadata struct {
	owner          string
	bottleneck     string
	whyItMatters   string
	nextActionBase []string
}

var metadataByCapability = map[string]capabilityMetadata{
	"Intent": {
		owner:        "Intent Engineer",
		bottleneck:   "Ambiguous requirements",
		whyItMatters: "Intent defines the outcomes, constraints, and success criteria the system is accountable to.",
		nextActionBase: []string{
			"Review intent.md for missing or unclear outcomes, constraints, and success criteria.",
			"Clarify the intended business outcome before changing downstream artifacts.",
		},
	},
	"Behavior": {
		owner:        "Behavior Engineer",
		bottleneck:   "Non-deterministic outputs",
		whyItMatters: "Behavior defines the expected and unacceptable system behavior the rest of the model validates against.",
		nextActionBase: []string{
			"Update behavior-spec.md so expected and unacceptable behavior are explicit.",
			"Remove ambiguity before changing design or assurance artifacts.",
		},
	},
	"Design": {
		owner:        "Design Engineer",
		bottleneck:   "Poor adoption / UX gaps",
		whyItMatters: "Design explains how the system works and gives operators and engineers the structure needed to evolve it safely.",
		nextActionBase: []string{
			"Expand architecture.md so the design is explicit and reviewable.",
			"Use the design artifact to close traceability or usability gaps before release.",
		},
	},
	"Assurance": {
		owner:        "Assurance Engineer",
		bottleneck:   "Validation gaps",
		whyItMatters: "Assurance proves the implemented system still behaves as intended using externally produced BDD results.",
		nextActionBase: []string{
			"Inspect bottleneck/assurance/results.json and confirm the scenario counts are correct.",
			"Fix failing scenarios or regenerate the external BDD results before promoting the system.",
		},
	},
	"Security": {
		owner:        "Security Engineer",
		bottleneck:   "Risk & compliance",
		whyItMatters: "Security establishes whether policy violations or governance failures make the system unsafe to operate.",
		nextActionBase: []string{
			"Inspect guardrails.json for policy violations and remove the underlying risk.",
			"Re-run validation after the violations count returns to zero.",
		},
	},
	"Execution": {
		owner:        "Execution Engineer",
		bottleneck:   "Delivery friction",
		whyItMatters: "Execution reflects real-world adoption and reliability, which determines whether the delivered system is actually working in practice.",
		nextActionBase: []string{
			"Review telemetry.json for reliability or adoption regressions in the target environment.",
			"Adjust delivery, UX, or rollout mechanics before the issue compounds in production.",
		},
	},
	"Traceability": {
		owner:        "Release Engineer",
		bottleneck:   "Traceability gaps",
		whyItMatters: "Traceability connects intent, behavior, assurance, security, and telemetry evidence so release decisions can be audited end to end.",
		nextActionBase: []string{
			"Inspect evidence IDs and Refs entries across the framework artifacts.",
			"Run bottleneck trace <id> for the affected evidence ID and repair missing or orphaned links.",
		},
	},
	"Config": {
		owner:        "Execution Engineer",
		bottleneck:   "Delivery friction",
		whyItMatters: "Configuration selects the thresholds that interpret the same artifacts across environments and must resolve cleanly before validation can run.",
		nextActionBase: []string{
			"Repair bottleneck/config.yaml so the default environment and overrides parse correctly.",
			"Confirm the selected environment inherits the intended thresholds before re-running validation.",
		},
	},
}

func Render(result models.EngineResult, capabilityFilter string) (string, error) {
	filtered, err := filterResults(result.Results, capabilityFilter)
	if err != nil {
		return "", err
	}

	var lines []string
	lines = append(lines,
		fmt.Sprintf("Environment: %s", result.Environment),
		fmt.Sprintf("System Status: %s", result.SystemStatus),
		fmt.Sprintf("Primary Bottleneck: %s", result.PrimaryBottleneck),
	)

	for _, validation := range filtered {
		meta := metadataFor(validation.Capability)

		lines = append(lines, "")
		lines = append(lines,
			fmt.Sprintf("Capability: %s", validation.Capability),
			fmt.Sprintf("Status: %s", validation.Status),
			fmt.Sprintf("Owner: %s", meta.owner),
			fmt.Sprintf("Mapped Bottleneck: %s", meta.bottleneck),
			fmt.Sprintf("Why It Matters: %s", meta.whyItMatters),
			"Evidence:",
		)

		evidence := collectEvidence(validation)
		for _, item := range evidence {
			lines = append(lines, fmt.Sprintf("- %s", item))
		}

		lines = append(lines, "Recommended Next Actions:")
		for _, action := range recommendedNextActions(validation, meta) {
			lines = append(lines, fmt.Sprintf("- %s", action))
		}
	}

	return strings.Join(lines, "\n"), nil
}

func filterResults(results []models.ValidationResult, capabilityFilter string) ([]models.ValidationResult, error) {
	if capabilityFilter == "" {
		return results, nil
	}

	for _, result := range results {
		if strings.EqualFold(result.Capability, capabilityFilter) {
			return []models.ValidationResult{result}, nil
		}
	}

	return nil, fmt.Errorf("unknown capability %q", capabilityFilter)
}

func metadataFor(capability string) capabilityMetadata {
	if meta, ok := metadataByCapability[capability]; ok {
		return meta
	}

	return capabilityMetadata{
		owner:        "Execution Engineer",
		bottleneck:   "Delivery friction",
		whyItMatters: "This capability contributes to the overall validity of the system and needs a clear owner for remediation.",
		nextActionBase: []string{
			"Inspect the underlying artifact and validation output for this capability.",
		},
	}
}

func collectEvidence(validation models.ValidationResult) []string {
	var evidence []string
	if validation.Message != "" {
		evidence = append(evidence, validation.Message)
	}
	evidence = append(evidence, validation.Details...)
	if len(evidence) == 0 {
		evidence = append(evidence, "No additional evidence reported.")
	}
	return evidence
}

func recommendedNextActions(validation models.ValidationResult, meta capabilityMetadata) []string {
	actions := append([]string{}, meta.nextActionBase...)

	switch validation.Status {
	case models.StatusPass:
		actions = append(actions, "Keep the artifact and observed evidence current as the system evolves.")
	case models.StatusWarning:
		actions = append(actions, "Treat this warning as an early signal and correct it before it becomes a failing bottleneck.")
	case models.StatusFail:
		actions = append(actions, "Resolve this bottleneck before relying on the system as valid for the selected environment.")
	}

	return actions
}

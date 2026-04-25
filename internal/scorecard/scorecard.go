package scorecard

import (
	"encoding/json"
	"fmt"
	"strings"

	"biased/internal/models"
)

type Scorecard struct {
	Environment       string                `json:"environment"`
	SystemStatus      string                `json:"system_status"`
	PrimaryBottleneck string                `json:"primary_bottleneck"`
	Capabilities      []CapabilityScorecard `json:"capabilities"`
	BottomLine        string                `json:"bottom_line"`
}

type CapabilityScorecard struct {
	Capability string `json:"capability"`
	Status     string `json:"status"`
	Owner      string `json:"owner"`
	Bottleneck string `json:"bottleneck"`
	Evidence   string `json:"evidence"`
}

type capabilityMetadata struct {
	owner      string
	bottleneck string
}

var metadataByCapability = map[string]capabilityMetadata{
	"Intent": {
		owner:      "Intent Engineer",
		bottleneck: "Ambiguous requirements",
	},
	"Behavior": {
		owner:      "Behavior Engineer",
		bottleneck: "Non-deterministic outputs",
	},
	"Design": {
		owner:      "Design Engineer",
		bottleneck: "Poor adoption / UX gaps",
	},
	"Assurance": {
		owner:      "Assurance Engineer",
		bottleneck: "Validation gaps",
	},
	"Security": {
		owner:      "Security Engineer",
		bottleneck: "Risk & compliance",
	},
	"Execution": {
		owner:      "Execution Engineer",
		bottleneck: "Delivery friction",
	},
	"Config": {
		owner:      "Execution Engineer",
		bottleneck: "Delivery friction",
	},
}

func Build(result models.EngineResult) Scorecard {
	capabilities := make([]CapabilityScorecard, 0, len(result.Results))
	for _, validation := range result.Results {
		meta := metadataFor(validation.Capability)
		capabilities = append(capabilities, CapabilityScorecard{
			Capability: validation.Capability,
			Status:     displayStatus(validation.Status),
			Owner:      meta.owner,
			Bottleneck: meta.bottleneck,
			Evidence:   evidenceFor(validation),
		})
	}

	return Scorecard{
		Environment:       result.Environment,
		SystemStatus:      result.SystemStatus,
		PrimaryBottleneck: result.PrimaryBottleneck,
		Capabilities:      capabilities,
		BottomLine:        bottomLine(result),
	}
}

func Render(result models.EngineResult, format string) (string, error) {
	card := Build(result)

	switch strings.ToLower(format) {
	case "text":
		return renderText(card), nil
	case "json":
		return renderJSON(card)
	default:
		return "", fmt.Errorf("unsupported format %q", format)
	}
}

func renderText(card Scorecard) string {
	var lines []string
	lines = append(lines,
		"BIASED System Scorecard",
		"",
		fmt.Sprintf("Environment: %s", card.Environment),
		fmt.Sprintf("System Status: %s", card.SystemStatus),
		fmt.Sprintf("Primary Bottleneck: %s", card.PrimaryBottleneck),
		"",
		fmt.Sprintf("%-12s %-8s %-24s %-28s %s", "Capability", "Status", "Owner", "Bottleneck", "Evidence"),
		fmt.Sprintf("%-12s %-8s %-24s %-28s %s", "----------", "------", "-----", "----------", "--------"),
	)

	for _, capability := range card.Capabilities {
		lines = append(lines, fmt.Sprintf(
			"%-12s %-8s %-24s %-28s %s",
			capability.Capability,
			capability.Status,
			capability.Owner,
			capability.Bottleneck,
			capability.Evidence,
		))
	}

	lines = append(lines, "", "Bottom line:", card.BottomLine)
	return strings.Join(lines, "\n")
}

func renderJSON(card Scorecard) (string, error) {
	content, err := json.MarshalIndent(card, "", "  ")
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func metadataFor(capability string) capabilityMetadata {
	if meta, ok := metadataByCapability[capability]; ok {
		return meta
	}

	return capabilityMetadata{
		owner:      "Execution Engineer",
		bottleneck: "Delivery friction",
	}
}

func evidenceFor(validation models.ValidationResult) string {
	if validation.Message != "" {
		return validation.Message
	}
	if len(validation.Details) > 0 {
		return strings.Join(validation.Details, "; ")
	}

	switch validation.Capability {
	case "Intent":
		return "intent.md valid"
	case "Behavior":
		return "behavior-spec.md valid"
	case "Design":
		return "architecture.md valid"
	case "Security":
		return "guardrails valid"
	case "Execution":
		return "telemetry valid"
	case "Config":
		return "config.yaml valid"
	default:
		return "artifact valid"
	}
}

func bottomLine(result models.EngineResult) string {
	if result.SystemStatus == models.StatusFail {
		return fmt.Sprintf(
			"The system is not valid for %s. Primary ownership starts with %s.",
			result.Environment,
			metadataFor(result.PrimaryBottleneck).owner,
		)
	}

	return fmt.Sprintf(
		"The system is valid for %s. Continue monitoring all capability signals.",
		result.Environment,
	)
}

func displayStatus(status string) string {
	if status == models.StatusWarning {
		return "WARN"
	}
	return status
}

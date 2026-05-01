package githubannotations

import (
	"fmt"
	"regexp"
	"strings"

	"bottleneck/internal/models"
)

var detailPathPattern = regexp.MustCompile(`(bottleneck/[^\s:]+)(?::([0-9]+))?`)

func Render(results []models.ValidationResult) string {
	var lines []string
	for _, result := range results {
		findings := findingsForResult(result)
		for _, finding := range findings {
			command := commandForFinding(finding)
			if command != "" {
				lines = append(lines, command)
			}
		}
	}
	return strings.Join(lines, "\n")
}

func findingsForResult(result models.ValidationResult) []models.ValidationFinding {
	if len(result.Findings) > 0 {
		return normalizeFindings(result)
	}
	if result.Status == models.StatusPass {
		return nil
	}

	level := levelForStatus(result.Status)
	if level == "" {
		return nil
	}

	var findings []models.ValidationFinding
	for _, detail := range result.Details {
		path, line := locationFromDetail(detail)
		if path == "" {
			path = defaultPathForCapability(result.Capability)
		}
		findings = append(findings, models.ValidationFinding{
			Level:   level,
			Message: detail,
			Path:    path,
			Line:    line,
		})
	}

	if len(findings) == 0 && result.Message != "" {
		findings = append(findings, models.ValidationFinding{
			Level:   level,
			Message: result.Capability + ": " + result.Message,
			Path:    defaultPathForCapability(result.Capability),
		})
	}

	return findings
}

func normalizeFindings(result models.ValidationResult) []models.ValidationFinding {
	level := levelForStatus(result.Status)
	normalized := make([]models.ValidationFinding, 0, len(result.Findings))
	for _, finding := range result.Findings {
		if finding.Level == "" {
			finding.Level = level
		}
		if finding.Message == "" {
			finding.Message = result.Message
		}
		if finding.Path == "" {
			path, line := locationFromDetail(finding.Message)
			if path == "" {
				for _, detail := range result.Details {
					path, line = locationFromDetail(detail)
					if path != "" {
						break
					}
				}
			}
			if path == "" {
				path = defaultPathForCapability(result.Capability)
			}
			finding.Path = path
			if finding.Line == 0 {
				finding.Line = line
			}
		}
		normalized = append(normalized, finding)
	}
	return normalized
}

func commandForFinding(finding models.ValidationFinding) string {
	level := strings.ToLower(finding.Level)
	if level != "warning" && level != "error" {
		return ""
	}
	if strings.TrimSpace(finding.Message) == "" {
		return ""
	}

	var props []string
	if finding.Path != "" {
		props = append(props, "file="+escapeProperty(finding.Path))
	}
	if finding.Line > 0 {
		props = append(props, fmt.Sprintf("line=%d", finding.Line))
	}

	if len(props) > 0 {
		return fmt.Sprintf("::%s %s::%s", level, strings.Join(props, ","), escapeMessage(finding.Message))
	}
	return fmt.Sprintf("::%s::%s", level, escapeMessage(finding.Message))
}

func levelForStatus(status string) string {
	switch status {
	case models.StatusFail:
		return "error"
	case models.StatusWarning:
		return "warning"
	default:
		return ""
	}
}

func pathFromDetail(detail string) string {
	path, _ := locationFromDetail(detail)
	return path
}

func locationFromDetail(detail string) (string, int) {
	match := detailPathPattern.FindStringSubmatch(detail)
	if len(match) < 2 {
		return "", 0
	}
	path := strings.TrimRight(match[1], ".,)")
	line := 0
	if len(match) > 2 && match[2] != "" {
		fmt.Sscanf(match[2], "%d", &line)
	}
	return path, line
}

func defaultPathForCapability(capability string) string {
	switch capability {
	case "Behavior":
		return "bottleneck/behavior/behavior-spec.md"
	case "Intent":
		return "bottleneck/intent/intent.md"
	case "Design":
		return "bottleneck/design/architecture.md"
	case "Assurance":
		return "bottleneck/assurance/results.json"
	case "Security":
		return "bottleneck/security/guardrails.json"
	case "Execution":
		return "bottleneck/execution/telemetry.json"
	case "Config":
		return "bottleneck/config.yaml"
	default:
		return ""
	}
}

func escapeMessage(value string) string {
	value = strings.ReplaceAll(value, "%", "%25")
	value = strings.ReplaceAll(value, "\r", "%0D")
	value = strings.ReplaceAll(value, "\n", "%0A")
	return value
}

func escapeProperty(value string) string {
	value = escapeMessage(value)
	value = strings.ReplaceAll(value, ":", "%3A")
	value = strings.ReplaceAll(value, ",", "%2C")
	return value
}

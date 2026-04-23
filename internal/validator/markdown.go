package validator

import "strings"

func containsMarkdownSection(content string, section string) bool {
	expected := "## " + section
	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) == expected {
			return true
		}
	}

	return false
}

func containsAnyMarkdownSection(content string) bool {
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") && len(trimmed) > 1 {
			return true
		}
	}

	return false
}

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

func markdownSectionBody(content string, section string) (string, bool) {
	lines := strings.Split(content, "\n")

	for index, line := range lines {
		level, heading, ok := markdownHeading(line)
		if !ok || heading != section {
			continue
		}

		var body []string
		for _, bodyLine := range lines[index+1:] {
			nextLevel, _, isHeading := markdownHeading(bodyLine)
			if isHeading && nextLevel <= level {
				break
			}

			body = append(body, bodyLine)
		}

		return strings.Join(body, "\n"), true
	}

	return "", false
}

func markdownDocumentBody(content string) string {
	var body []string
	for _, line := range strings.Split(content, "\n") {
		if _, _, ok := markdownHeading(line); ok {
			continue
		}

		body = append(body, line)
	}

	return strings.Join(body, "\n")
}

func markdownHeading(line string) (int, string, bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || trimmed[0] != '#' {
		return 0, "", false
	}

	level := 0
	for level < len(trimmed) && trimmed[level] == '#' {
		level++
	}
	if level == 0 || level > 6 || level == len(trimmed) {
		return 0, "", false
	}
	if trimmed[level] != ' ' && trimmed[level] != '\t' {
		return 0, "", false
	}

	heading := strings.TrimSpace(trimmed[level:])
	heading = strings.TrimSpace(strings.TrimRight(heading, "#"))
	if heading == "" {
		return 0, "", false
	}

	return level, heading, true
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

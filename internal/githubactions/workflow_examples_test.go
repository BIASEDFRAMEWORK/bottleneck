package githubactions

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowExamplesContainExpectedCommands(t *testing.T) {
	examples := []string{
		"bottleneck-validate.yml",
		"bottleneck-scorecard.yml",
		"bottleneck-pr-gate.yml",
	}

	for _, example := range examples {
		t.Run(example, func(t *testing.T) {
			path := filepath.Join("..", "..", "examples", "github-actions", example)
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read workflow example: %v", err)
			}
			text := string(content)

			for _, substring := range []string{"pull_request:", "workflow_dispatch:", "go build -o bottleneck ."} {
				if !strings.Contains(text, substring) {
					t.Fatalf("expected %q in %s:\n%s", substring, example, text)
				}
			}

			for _, substring := range []string{"$GITHUB_STEP_SUMMARY", "scorecard", "--format=markdown"} {
				if !strings.Contains(text, substring) {
					t.Fatalf("expected %q in %s:\n%s", substring, example, text)
				}
			}

			if strings.Contains(example, "validate") || strings.Contains(example, "gate") {
				if !strings.Contains(text, "validate") {
					t.Fatalf("expected validate command in %s:\n%s", example, text)
				}
			}

			if strings.Contains(example, "gate") {
				for _, substring := range []string{"diagnose", "--gate release", "--format github", "--format markdown", "bottleneck-diagnosis.md", "<!-- bottleneck-diagnosis -->"} {
					if !strings.Contains(text, substring) {
						t.Fatalf("expected %q in %s:\n%s", substring, example, text)
					}
				}
			}
		})
	}
}

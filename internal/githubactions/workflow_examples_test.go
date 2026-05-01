package githubactions

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestWorkflowExamplesContainExpectedCommands(t *testing.T) {
	examples := []string{
		"bottleneck-validate.yml",
		"bottleneck-scorecard.yml",
		"bottleneck-pr-gate.yml",
		"bottleneck-saas-scorecard.yml",
	}

	for _, example := range examples {
		t.Run(example, func(t *testing.T) {
			path := filepath.Join("..", "..", "examples", "github-actions", example)
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read workflow example: %v", err)
			}
			text := string(content)

			for _, substring := range []string{"pull_request:", "workflow_dispatch:"} {
				if !strings.Contains(text, substring) {
					t.Fatalf("expected %q in %s:\n%s", substring, example, text)
				}
			}
			if !strings.Contains(text, "go build -o bottleneck .") && !strings.Contains(text, "go build -o ./bin/bottleneck .") {
				t.Fatalf("expected workflow to build bottleneck in %s:\n%s", example, text)
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

func TestSaaSScorecardWorkflowIsCopyPasteSafe(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "github-actions", "bottleneck-saas-scorecard.yml")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read SaaS workflow example: %v", err)
	}
	text := string(content)

	var parsed map[string]interface{}
	if err := yaml.Unmarshal(content, &parsed); err != nil {
		t.Fatalf("SaaS workflow YAML should parse: %v\n%s", err, text)
	}

	expected := []string{
		"name: bottleneck SaaS scorecard",
		"pull_request:",
		"push:",
		"workflow_dispatch:",
		"actions/checkout@v4",
		"actions/setup-go@v5",
		"go-version-file: go.mod",
		"go build -o ./bin/bottleneck .",
		"./bin/bottleneck validate --env=\"$BOTTLENECK_ENV\" --github-annotations",
		"./bin/bottleneck scorecard --env=\"$BOTTLENECK_ENV\" --format=markdown >> \"$GITHUB_STEP_SUMMARY\"",
		"./bin/bottleneck diagnose --env=\"$BOTTLENECK_ENV\" --format=github",
		"./bin/bottleneck diagnose --env=\"$BOTTLENECK_ENV\" --gate=release",
		"continue-on-error: true",
		"permissions:",
		"contents: read",
	}
	for _, substring := range expected {
		if !strings.Contains(text, substring) {
			t.Fatalf("expected SaaS workflow to contain %q:\n%s", substring, text)
		}
	}

	for _, forbidden := range []string{
		"secrets.",
		"/Users/",
		"/tmp/",
		"/private/",
		"~/",
		"pull-requests: write",
		"actions/github-script",
		"bottleneck-pr-comment",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("SaaS workflow should not contain %q:\n%s", forbidden, text)
		}
	}
}

func TestSaaSScorecardWorkflowReferencesValidCommands(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "github-actions", "bottleneck-saas-scorecard.yml")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read SaaS workflow example: %v", err)
	}
	text := string(content)

	commands := map[string]string{
		"validate":  "validate.go",
		"scorecard": "scorecard.go",
		"diagnose":  "diagnose.go",
	}
	for command, file := range commands {
		if !strings.Contains(text, "bottleneck "+command) {
			t.Fatalf("expected SaaS workflow to reference bottleneck %s:\n%s", command, text)
		}
		commandPath := filepath.Join("..", "..", "cmd", file)
		source, err := os.ReadFile(commandPath)
		if err != nil {
			t.Fatalf("read command source %s: %v", file, err)
		}
		if !strings.Contains(string(source), `Use:   "`+command+`"`) {
			t.Fatalf("expected cmd/%s to register %q command", file, command)
		}
	}

	diagnoseSource, err := os.ReadFile(filepath.Join("..", "..", "cmd", "diagnose.go"))
	if err != nil {
		t.Fatalf("read diagnose command: %v", err)
	}
	if strings.Contains(text, "diagnose --env=\"$BOTTLENECK_ENV\" --format=github") &&
		!strings.Contains(string(diagnoseSource), "text, json, markdown, or github") {
		t.Fatal("workflow emits diagnose --format=github but diagnose help does not advertise github format support")
	}
}

func TestEvidenceReportWorkflowGeneratesAndUploadsArtifacts(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "github-actions", "bottleneck-evidence-report.yml")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read evidence report workflow example: %v", err)
	}
	text := string(content)

	var parsed map[string]interface{}
	if err := yaml.Unmarshal(content, &parsed); err != nil {
		t.Fatalf("evidence report workflow YAML should parse: %v\n%s", err, text)
	}

	expected := []string{
		"name: Bottleneck Evidence Report",
		"pull_request:",
		"workflow_dispatch:",
		"actions/checkout@v4",
		"actions/setup-go@v5",
		"go-version-file: go.mod",
		"go build -o bottleneck-cli .",
		"./bottleneck-cli validate",
		"./bottleneck-cli scorecard --format=markdown --details",
		"./bottleneck-cli snapshot --label=ci",
		"./bottleneck-cli trends --format=markdown --out=bottleneck/reports/trend-summary.md",
		"./bottleneck-cli report --format=markdown --out=bottleneck/reports/sdlc-evidence-report.md",
		"actions/upload-artifact@v4",
		"name: bottleneck-reports",
		"bottleneck/history/",
		"bottleneck/reports/",
	}
	for _, substring := range expected {
		if !strings.Contains(text, substring) {
			t.Fatalf("expected evidence report workflow to contain %q:\n%s", substring, text)
		}
	}

	for _, forbidden := range []string{
		"git commit",
		"secrets.",
		"/Users/",
		"/tmp/",
		"/private/",
		"~/",
		"actions/github-script",
		"pull-requests: write",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("evidence report workflow should not contain %q:\n%s", forbidden, text)
		}
	}
}

func TestEvidenceReportWorkflowReferencesValidCommands(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "github-actions", "bottleneck-evidence-report.yml")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read evidence report workflow example: %v", err)
	}
	text := string(content)

	commands := map[string]string{
		"validate":  "validate.go",
		"scorecard": "scorecard.go",
		"snapshot":  "snapshot.go",
		"trends":    "trends.go",
		"report":    "report.go",
	}
	for command, file := range commands {
		if !strings.Contains(text, "bottleneck-cli "+command) {
			t.Fatalf("expected evidence report workflow to reference bottleneck %s:\n%s", command, text)
		}
		commandPath := filepath.Join("..", "..", "cmd", file)
		source, err := os.ReadFile(commandPath)
		if err != nil {
			t.Fatalf("read command source %s: %v", file, err)
		}
		if !strings.Contains(string(source), `Use:   "`+command+`"`) {
			t.Fatalf("expected cmd/%s to register %q command", file, command)
		}
	}
}

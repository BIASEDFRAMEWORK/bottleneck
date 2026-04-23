package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initDirectories = []string{
	"biased",
	"biased/behavior",
	"biased/intent",
	"biased/design",
	"biased/assurance",
	"biased/assurance/features",
	"biased/security",
	"biased/execution",
}

var initFiles = map[string]string{
	"biased/behavior/behavior-spec.md":         "# Behavior Specification\n",
	"biased/intent/intent.md":                  "# Intent\n",
	"biased/design/architecture.md":            "# Architecture\n",
	"biased/assurance/features/sample.feature": "Feature: Sample system validation\n\n  Scenario: System behaves as intended\n    Given the system is initialized\n    When validation artifacts are present\n    Then the outcomes should satisfy BIASED\n",
	"biased/assurance/results.json":            "{\n  \"scenarios_total\": 1,\n  \"scenarios_passed\": 1,\n  \"scenarios_failed\": 0,\n  \"accuracy\": 1.0,\n  \"failures\": []\n}\n",
	"biased/security/guardrails.json":          "{\n  \"violations\": 0\n}\n",
	"biased/execution/telemetry.json":          "{\n  \"adoption_rate\": 0.9,\n  \"error_rate\": 0.01\n}\n",
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the BIASED directory structure",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initializeProject("."); err != nil {
			return err
		}

		fmt.Println("BIASED project initialized")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func initializeProject(basePath string) error {
	for _, dir := range initDirectories {
		if err := os.MkdirAll(filepath.Join(basePath, dir), 0o755); err != nil {
			return err
		}
	}

	for relativePath, content := range initFiles {
		if err := writeFileIfMissing(filepath.Join(basePath, relativePath), content); err != nil {
			return err
		}
	}

	return nil
}

func writeFileIfMissing(path string, content string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}

	return os.WriteFile(path, []byte(content), 0o644)
}

package cmd

import (
	"fmt"
	"os"
	"strings"

	"bottleneck/internal/githubannotations"
	"bottleneck/internal/models"
	"bottleneck/internal/validator"

	"github.com/spf13/cobra"
)

var validateEnv string
var validateStrict bool
var validateGitHubAnnotations bool

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate framework evidence artifacts",
	Run: func(cmd *cobra.Command, args []string) {
		engine := validator.NewEngine(".", validateEnv, validator.WithStrictMode(validateStrict))
		result := engine.Validate()

		fmt.Println(renderValidationOutput(result))
		if validateGitHubAnnotations {
			if annotations := githubannotations.Render(result.Results); annotations != "" {
				fmt.Fprintln(os.Stderr, annotations)
			}
		}

		if result.SystemStatus == models.StatusFail {
			os.Exit(1)
		}
	},
}

func init() {
	validateCmd.Flags().StringVar(&validateEnv, "env", "default", "environment config to use")
	validateCmd.Flags().BoolVar(&validateStrict, "strict", false, "treat placeholder and insufficient content as failures")
	validateCmd.Flags().BoolVar(&validateGitHubAnnotations, "github-annotations", false, "emit GitHub Actions warning and error annotations")
	rootCmd.AddCommand(validateCmd)
}

func renderValidationOutput(result models.EngineResult) string {
	var lines []string
	for _, check := range result.Results {
		line := fmt.Sprintf("%s: %s", check.Capability, check.Status)
		if check.Message != "" {
			line = fmt.Sprintf("%s (%s)", line, check.Message)
		}

		lines = append(lines, line)
		for _, detail := range check.Details {
			lines = append(lines, fmt.Sprintf("  %s", detail))
		}
	}

	lines = append(lines,
		"",
		fmt.Sprintf("System Status: %s", result.SystemStatus),
		fmt.Sprintf("Primary Bottleneck: %s", result.PrimaryBottleneck),
		fmt.Sprintf("Environment: %s", result.Environment),
	)

	return strings.Join(lines, "\n")
}

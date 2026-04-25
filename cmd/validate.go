package cmd

import (
	"fmt"
	"os"

	"biased/internal/models"
	"biased/internal/validator"

	"github.com/spf13/cobra"
)

var validateEnv string

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the BIASED directory structure",
	Run: func(cmd *cobra.Command, args []string) {
		engine := validator.NewEngine(".", validateEnv)
		result := engine.Validate()

		for _, check := range result.Results {
			line := fmt.Sprintf("%s: %s", check.Capability, check.Status)
			if check.Message != "" {
				line = fmt.Sprintf("%s (%s)", line, check.Message)
			}

			fmt.Println(line)
			for _, detail := range check.Details {
				fmt.Printf("  %s\n", detail)
			}
		}

		fmt.Println()
		fmt.Printf("System Status: %s\n", result.SystemStatus)
		fmt.Printf("Primary Bottleneck: %s\n", result.PrimaryBottleneck)
		fmt.Printf("Environment: %s\n", result.Environment)

		if result.SystemStatus == models.StatusFail {
			os.Exit(1)
		}
	},
}

func init() {
	validateCmd.Flags().StringVar(&validateEnv, "env", "default", "environment config to use")
	rootCmd.AddCommand(validateCmd)
}

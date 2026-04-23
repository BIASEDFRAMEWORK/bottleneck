package cmd

import (
	"fmt"
	"os"

	"biased/internal/models"
	"biased/internal/validator"

	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the BIASED directory structure",
	Run: func(cmd *cobra.Command, args []string) {
		engine := validator.NewEngine(".")
		result := engine.Validate()

		for _, check := range result.Results {
			line := fmt.Sprintf("%s: %s", check.Capability, check.Status)
			if check.Message != "" {
				line = fmt.Sprintf("%s (%s)", line, check.Message)
			}

			fmt.Println(line)
		}

		fmt.Println()
		fmt.Printf("System Status: %s\n", result.SystemStatus)
		fmt.Printf("Primary Bottleneck: %s\n", result.PrimaryBottleneck)

		if result.SystemStatus == models.StatusFail {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

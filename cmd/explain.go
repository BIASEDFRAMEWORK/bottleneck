package cmd

import (
	"fmt"

	"biased/internal/explainer"
	"biased/internal/validator"

	"github.com/spf13/cobra"
)

var explainEnv string
var explainCapability string

var explainCmd = &cobra.Command{
	Use:   "explain",
	Short: "Explain the current validation state in human-readable form",
	RunE: func(cmd *cobra.Command, args []string) error {
		engine := validator.NewEngine(".", explainEnv)
		result := engine.Validate()

		output, err := explainer.Render(result, explainCapability)
		if err != nil {
			return err
		}

		fmt.Println(output)
		return nil
	},
}

func init() {
	explainCmd.Flags().StringVar(&explainEnv, "env", "default", "environment config to use")
	explainCmd.Flags().StringVar(&explainCapability, "capability", "", "single capability to explain")
	rootCmd.AddCommand(explainCmd)
}

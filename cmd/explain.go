package cmd

import (
	"fmt"
	"os"
	"strings"

	"bottleneck/internal/explainer"
	"bottleneck/internal/models"
	"bottleneck/internal/traceability"
	"bottleneck/internal/validator"

	"github.com/spf13/cobra"
)

var explainEnv string
var explainCapability string
var explainCategory string
var explainFormat string
var explainOut string
var explainStrict bool

var explainCmd = &cobra.Command{
	Use:   "explain",
	Short: "Show how evidence affected category scores",
	RunE: func(cmd *cobra.Command, args []string) error {
		engine := validator.NewEngine(".", explainEnv, validator.WithStrictMode(explainStrict))
		result := engine.Validate()
		graph, graphErr := traceability.Build("bottleneck", traceability.Options{
			Environment: explainEnv,
			Strict:      explainStrict,
		})
		var graphPtr *traceability.Graph
		if graphErr == nil {
			graphPtr = &graph
		}

		filter, err := explainFilter()
		if err != nil {
			return err
		}

		output, err := explainer.RenderWithOptions(result, graphPtr, explainer.Options{
			Filter: filter,
			Format: explainFormat,
		})
		if err != nil {
			return err
		}
		if explainOut != "" {
			if err := explainer.WriteOutput(explainOut, output+"\n"); err != nil {
				return fmt.Errorf("write explanation output %s: %w", explainOut, err)
			}
		}

		fmt.Println(output)
		if result.SystemStatus == models.StatusFail {
			os.Exit(1)
		}

		return nil
	},
}

func init() {
	explainCmd.Flags().StringVar(&explainEnv, "env", "default", "environment config to use")
	explainCmd.Flags().StringVar(&explainCapability, "capability", "", "single capability to explain")
	explainCmd.Flags().StringVar(&explainCategory, "category", "", "single category to explain")
	explainCmd.Flags().StringVar(&explainFormat, "format", explainer.FormatText, "output format: text, markdown, or json")
	explainCmd.Flags().StringVar(&explainOut, "out", "", "optional file path to write rendered explanation")
	explainCmd.Flags().BoolVar(&explainStrict, "strict", false, "treat placeholder and insufficient content as failures")
	rootCmd.AddCommand(explainCmd)
}

func explainFilter() (string, error) {
	category := strings.TrimSpace(explainCategory)
	capability := strings.TrimSpace(explainCapability)
	if category != "" && capability != "" && !strings.EqualFold(category, capability) {
		return "", fmt.Errorf("use either --category or --capability, not both")
	}
	if category != "" {
		return category, nil
	}
	return capability, nil
}

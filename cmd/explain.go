package cmd

import (
	"fmt"
	"os"

	"bottleneck/internal/explainer"
	"bottleneck/internal/models"
	"bottleneck/internal/traceability"
	"bottleneck/internal/validator"

	"github.com/spf13/cobra"
)

var explainEnv string
var explainCapability string
var explainStrict bool

var explainCmd = &cobra.Command{
	Use:   "explain",
	Short: "Explain the current validation state in human-readable form",
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

		output, err := explainer.RenderWithGraph(result, graphPtr, explainCapability)
		if err != nil {
			return err
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
	explainCmd.Flags().BoolVar(&explainStrict, "strict", false, "treat placeholder and insufficient content as failures")
	rootCmd.AddCommand(explainCmd)
}

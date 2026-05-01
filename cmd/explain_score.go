package cmd

import (
	"fmt"
	"os"

	"bottleneck/internal/assessment"
	"bottleneck/internal/discover"
	"bottleneck/internal/validator"

	"github.com/spf13/cobra"
)

var explainScoreFormat string
var explainScoreStrict bool
var explainScoreEnvironment string

var explainScoreCmd = &cobra.Command{
	Use:   "explain-score",
	Short: "Explain the evidence and rationale behind the maturity score",
	Long: `Explain maturity, release recommendation, primary bottleneck, score
confidence, AI readiness, and per-category evidence provenance without writing
files.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		rootPath, err := os.Getwd()
		if err != nil {
			return err
		}
		discovery, err := discover.Scan(rootPath)
		if err != nil {
			return err
		}
		engine := validator.NewEngine(".", explainScoreEnvironment, validator.WithStrictMode(explainScoreStrict))
		result := engine.Validate()
		report := assessment.Build(result, discovery, assessment.Options{
			RootPath:    rootPath,
			Environment: explainScoreEnvironment,
		})
		output, err := assessment.Render(report, explainScoreFormat)
		if err != nil {
			return err
		}
		fmt.Println(output)
		return nil
	},
}

func init() {
	explainScoreCmd.Flags().StringVar(&explainScoreFormat, "format", "text", "output format: text, json, or markdown")
	explainScoreCmd.Flags().BoolVar(&explainScoreStrict, "strict", false, "treat placeholder and insufficient content as failures")
	explainScoreCmd.Flags().StringVar(&explainScoreEnvironment, "environment", "default", "environment config to use")
	explainScoreCmd.Flags().StringVar(&explainScoreEnvironment, "env", "default", "environment config to use")
	rootCmd.AddCommand(explainScoreCmd)
}

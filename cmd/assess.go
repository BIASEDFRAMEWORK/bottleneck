package cmd

import (
	"fmt"
	"os"

	"bottleneck/internal/assessment"
	"bottleneck/internal/discover"
	"bottleneck/internal/validator"

	"github.com/spf13/cobra"
)

var assessNoIngest bool
var assessFormat string
var assessStrict bool
var assessEnvironment string

var assessCmd = &cobra.Command{
	Use:     "assess",
	Aliases: []string{"maturity"},
	Short:   "Assess SDLC maturity from local engineering evidence",
	Long: `Assess local engineering evidence and summarize SDLC maturity, AI readiness,
release friction, the primary bottleneck, score confidence, and the next action.

By default assess discovers and auto-ingests supported local evidence before
running the scorecard. Use --no-ingest for a read-only assessment.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		rootPath, err := os.Getwd()
		if err != nil {
			return err
		}

		var warnings []string
		autoIngested := false
		if !assessNoIngest {
			autoResult, err := executeAutoIngest(rootPath, false)
			if err != nil {
				warnings = append(warnings, err.Error())
			} else {
				warnings = append(warnings, autoResult.Warnings...)
				autoIngested = len(autoResult.Written) > 0
			}
		}

		discovery, err := discover.Scan(rootPath)
		if err != nil {
			return err
		}

		engine := validator.NewEngine(".", assessEnvironment, validator.WithStrictMode(assessStrict))
		result := engine.Validate()
		report := assessment.Build(result, discovery, assessment.Options{
			RootPath:     rootPath,
			Environment:  assessEnvironment,
			AutoIngested: autoIngested,
			Warnings:     warnings,
		})

		output, err := assessment.Render(report, assessFormat)
		if err != nil {
			return err
		}
		fmt.Println(output)
		return nil
	},
}

func init() {
	assessCmd.Flags().BoolVar(&assessNoIngest, "no-ingest", false, "skip automatic evidence ingestion before scoring")
	assessCmd.Flags().StringVar(&assessFormat, "format", "text", "output format: text, json, or markdown")
	assessCmd.Flags().BoolVar(&assessStrict, "strict", false, "treat placeholder and insufficient content as failures")
	assessCmd.Flags().StringVar(&assessEnvironment, "environment", "default", "environment config to use")
	assessCmd.Flags().StringVar(&assessEnvironment, "env", "default", "environment config to use")
	rootCmd.AddCommand(assessCmd)
}

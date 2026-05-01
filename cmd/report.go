package cmd

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"bottleneck/internal/explainer"
	"bottleneck/internal/models"
	"bottleneck/internal/report"
	"bottleneck/internal/scorecard"
	"bottleneck/internal/traceability"
	"bottleneck/internal/trends"
	"bottleneck/internal/validator"

	"github.com/spf13/cobra"
)

var reportEnv string
var reportFormat string
var reportOut string
var reportWindow int
var reportStrict bool

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate a leadership-ready SDLC evidence report",
	Long: `Generate a leadership-ready SDLC evidence report from local Bottleneck evidence.

The report combines the current scorecard, local snapshot trends when present,
and evidence-backed explanation data. It writes local Markdown or JSON output
without a database, backend, dashboard, or external service.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		env := strings.TrimSpace(reportEnv)
		if env == "" {
			env = "default"
		}
		if reportWindow <= 0 {
			return errors.New("window must be greater than 0")
		}

		engine := validator.NewEngine(".", env, validator.WithStrictMode(reportStrict))
		result := engine.Validate()
		if configFailure := firstReportConfigFailure(result); configFailure != "" {
			return errors.New(configFailure)
		}

		card := scorecard.Build(result)
		trendAnalysis, err := trends.Analyze(trends.Options{
			RootPath:    ".",
			Environment: env,
			Window:      reportWindow,
		})
		if err != nil {
			return err
		}

		graph, graphErr := traceability.Build("bottleneck", traceability.Options{
			Environment: env,
			Strict:      reportStrict,
		})
		var graphPtr *traceability.Graph
		if graphErr == nil {
			graphPtr = &graph
		}

		explanation, err := explainer.BuildReport(result, graphPtr, "")
		if err != nil {
			return err
		}
		sdlcReport := report.Build(card, &trendAnalysis, explanation, report.Options{GeneratedAt: time.Now().UTC()})
		outputFormat := strings.TrimSpace(reportFormat)
		if outputFormat == "" {
			outputFormat = report.FormatMarkdown
		}
		output, err := report.Render(sdlcReport, outputFormat)
		if err != nil {
			return err
		}

		outPath := strings.TrimSpace(reportOut)
		if outPath == "" && outputFormat == report.FormatMarkdown {
			outPath = report.DefaultReportPath
		}
		if outPath != "" {
			if err := report.WriteFile(outPath, output+"\n"); err != nil {
				return fmt.Errorf("write SDLC evidence report %s: %w", outPath, err)
			}
			fmt.Println(renderReportSuccess(sdlcReport, outPath))
			return nil
		}

		fmt.Println(output)
		return nil
	},
}

func init() {
	reportCmd.Flags().StringVar(&reportEnv, "env", "default", "environment config to use")
	reportCmd.Flags().StringVar(&reportFormat, "format", report.FormatMarkdown, "output format: markdown or json")
	reportCmd.Flags().StringVar(&reportOut, "out", "", "optional file path to write rendered report")
	reportCmd.Flags().IntVar(&reportWindow, "window", trends.DefaultWindow, "number of latest snapshots to analyze")
	reportCmd.Flags().BoolVar(&reportStrict, "strict", false, "treat placeholder and insufficient content as failures")
	rootCmd.AddCommand(reportCmd)
}

func firstReportConfigFailure(result models.EngineResult) string {
	for _, validation := range result.Results {
		if validation.Capability == "Config" && validation.Status == models.StatusFail {
			if strings.TrimSpace(validation.Message) != "" {
				return validation.Message
			}
			return "invalid Bottleneck config"
		}
	}
	return ""
}

func renderReportSuccess(sdlcReport report.Report, outPath string) string {
	trend := "Insufficient history"
	if sdlcReport.TrendSummary != nil {
		trend = string(sdlcReport.TrendSummary.OverallDirection)
		if sdlcReport.TrendSummary.PersistentBottleneck.Category != "" {
			trend = fmt.Sprintf("%s, but %s remains persistent", trend, sdlcReport.TrendSummary.PersistentBottleneck.Category)
		}
	}
	return strings.Join([]string{
		"SDLC evidence report created",
		"",
		fmt.Sprintf("Status: %s", sdlcReport.CurrentStatus),
		fmt.Sprintf("Primary bottleneck: %s", sdlcReport.PrimaryBottleneck),
		fmt.Sprintf("Trend: %s", trend),
		fmt.Sprintf("Report: %s", pathForDisplay(outPath)),
	}, "\n")
}

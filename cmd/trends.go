package cmd

import (
	"errors"
	"fmt"
	"strings"

	"bottleneck/internal/trends"

	"github.com/spf13/cobra"
)

var trendsEnv string
var trendsWindow int
var trendsFormat string
var trendsOut string

var trendsCmd = &cobra.Command{
	Use:   "trends",
	Short: "Analyze scorecard trends from local snapshots",
	Long: `Analyze local Bottleneck scorecard snapshots over time.

Trends reads bottleneck/history/scorecards/, filters snapshots by environment,
compares category scores, identifies persistent bottlenecks, and renders a
deterministic local trend summary without a database or external service.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		env := strings.TrimSpace(trendsEnv)
		if env == "" {
			env = "default"
		}
		if trendsWindow <= 0 {
			return errors.New("window must be greater than 0")
		}

		analysis, err := trends.Analyze(trends.Options{
			RootPath:    ".",
			Environment: env,
			Window:      trendsWindow,
		})
		if err != nil {
			return err
		}

		output, err := trends.Render(analysis, trendsFormat)
		if err != nil {
			return err
		}
		if trendsOut != "" {
			if err := trends.WriteReport(trendsOut, output+"\n"); err != nil {
				return fmt.Errorf("write trends report %s: %w", trendsOut, err)
			}
		}
		fmt.Println(output)
		return nil
	},
}

func init() {
	trendsCmd.Flags().StringVar(&trendsEnv, "env", "default", "environment config to analyze")
	trendsCmd.Flags().IntVar(&trendsWindow, "window", trends.DefaultWindow, "number of latest snapshots to analyze")
	trendsCmd.Flags().StringVar(&trendsFormat, "format", trends.FormatText, "output format: text, markdown, or json")
	trendsCmd.Flags().StringVar(&trendsOut, "out", "", "optional file path to write rendered trend output")
	rootCmd.AddCommand(trendsCmd)
}

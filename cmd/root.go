package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "bottleneck",
	Short: "Diagnose delivery risk with the BIASED evidence model",
	Long: `Bottleneck is a CLI that diagnoses delivery risk using the BIASED evidence model. It validates local evidence artifacts, renders scorecards, checks release readiness, and explains the primary bottleneck diagnosis.

Start here:
  bottleneck init --template saas
  bottleneck scorecard
  bottleneck diagnose

Common commands:
  validate    Check evidence files for missing, thin, or placeholder content.
  scorecard   Show release readiness, primary bottleneck, and next action.
  diagnose    Explain what is blocking delivery and what to inspect next.
  trace       Follow one intent, behavior, or evidence ID end-to-end.
  ingest      Convert test, security, and telemetry reports into Bottleneck evidence.
  explain     Show how evidence affected category scores.

Use scorecard as the main terminal surface for release readiness, primary bottleneck, and next action.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

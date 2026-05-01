package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "bottleneck",
	Short: "Measure SDLC maturity from local engineering evidence",
	Long: `Bottleneck is a CLI for measuring SDLC maturity from local engineering evidence. It tells you what is blocking release confidence and what evidence proves it.

Start here:
  bottleneck init --template saas
  bottleneck assess
  bottleneck trace BEHAVIOR-003

Common commands:
  assess      Show maturity, AI readiness, release friction, and next action.
  discover    Find local test, security, telemetry, design, and workflow evidence.
  evidence sync Discover and ingest supported local evidence automatically.
  trace       Follow one intent, behavior, or evidence ID end-to-end.
  explain-score Explain the maturity score with evidence provenance.
  report      Generate a leadership-ready SDLC evidence report.

Advanced commands:
  check       Check evidence files for missing, thin, or placeholder content.
  validate    Check evidence files for missing, thin, or placeholder content.
  scorecard   Show release readiness, primary bottleneck, and next action.
  diagnose    Explain what is blocking delivery and what to inspect next.
  ingest      Convert test, security, and telemetry reports into Bottleneck evidence.
  snapshot    Write scorecard snapshots for local trend history.
  seed-history Create demo scorecard history for local trends.
  trends      Analyze scorecard trends from local snapshots.
  explain     Show how evidence affected category scores.

Use assess as the main terminal surface for SDLC maturity, release confidence, and the next action.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

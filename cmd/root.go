package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "bottleneck",
	Short:         "Diagnose delivery risk with the BIASED evidence model",
	Long:          "Bottleneck is a CLI that diagnoses delivery risk using the BIASED evidence model. It validates local evidence artifacts, renders scorecards, checks release readiness, and explains the primary bottleneck diagnosis.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

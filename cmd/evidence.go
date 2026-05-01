package cmd

import (
	"github.com/spf13/cobra"
)

var evidenceCmd = &cobra.Command{
	Use:   "evidence",
	Short: "Work with Bottleneck evidence artifacts",
}

var evidenceSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Discover and ingest supported evidence automatically",
	RunE: func(cmd *cobra.Command, args []string) error {
		ingestAuto = true
		return runAutoIngest()
	},
}

func init() {
	evidenceSyncCmd.Flags().BoolVar(&ingestDryRun, "dry-run", false, "parse and print normalized evidence without writing files")
	evidenceSyncCmd.Flags().StringVar(&ingestFormat, "format", "text", "output format: text or json")
	evidenceSyncCmd.Flags().BoolVar(&ingestMerge, "merge", false, "merge with existing artifact evidence instead of replacing it")
	evidenceCmd.AddCommand(evidenceSyncCmd)
	rootCmd.AddCommand(evidenceCmd)
}

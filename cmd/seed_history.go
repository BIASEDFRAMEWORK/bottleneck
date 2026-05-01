package cmd

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"bottleneck/internal/gitinfo"
	"bottleneck/internal/seedhistory"
	"bottleneck/internal/snapshot"

	"github.com/spf13/cobra"
)

var seedHistoryScenario string
var seedHistoryEnv string
var seedHistorySnapshots int
var seedHistoryOut string
var seedHistoryOverwrite bool

var seedHistoryCmd = &cobra.Command{
	Use:   "seed-history",
	Short: "Create demo scorecard history for local trends",
	Long: `Create deterministic local scorecard snapshots for demos and tests.

Seed history writes Bottleneck scorecard snapshot JSON files that work with
bottleneck trends and bottleneck report. It uses local files only and does not
require Git, a database, a backend, or an external metrics service.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		env := strings.TrimSpace(seedHistoryEnv)
		if env == "" {
			env = seedhistory.DefaultEnvironment
		}
		if seedHistorySnapshots < 1 {
			return errors.New("snapshots must be greater than 0")
		}

		result, err := seedhistory.Write(seedhistory.Options{
			RootPath:      ".",
			Scenario:      seedHistoryScenario,
			Environment:   env,
			SnapshotCount: seedHistorySnapshots,
			OutDir:        seedHistoryOut,
			Overwrite:     seedHistoryOverwrite,
			Now:           time.Now().UTC(),
			Git:           gitinfo.Detect("."),
		})
		if err != nil {
			return err
		}

		fmt.Println(renderSeedHistorySuccess(result))
		return nil
	},
}

func init() {
	seedHistoryCmd.Flags().StringVar(&seedHistoryScenario, "scenario", seedhistory.ScenarioSaaSDayOne, "seed history scenario to generate")
	seedHistoryCmd.Flags().StringVar(&seedHistoryEnv, "env", seedhistory.DefaultEnvironment, "environment metadata to write")
	seedHistoryCmd.Flags().IntVar(&seedHistorySnapshots, "snapshots", seedhistory.DefaultSnapshotCount, "number of scenario snapshots to generate")
	seedHistoryCmd.Flags().StringVar(&seedHistoryOut, "out", snapshot.DefaultScorecardDir, "directory for seeded scorecard snapshots")
	seedHistoryCmd.Flags().BoolVar(&seedHistoryOverwrite, "overwrite", false, "overwrite existing seed history in the output directory")
	rootCmd.AddCommand(seedHistoryCmd)
}

func renderSeedHistorySuccess(result seedhistory.Result) string {
	return strings.Join([]string{
		"Bottleneck seed history created",
		"",
		fmt.Sprintf("Scenario: %s", result.Scenario),
		fmt.Sprintf("Environment: %s", result.Environment),
		fmt.Sprintf("Snapshots: %d", len(result.Snapshots)),
		fmt.Sprintf("Output: %s", pathForDisplay(result.OutDir)),
		"",
		"Next:",
		"Run `bottleneck trends` to see SDLC evidence direction over time.",
	}, "\n")
}

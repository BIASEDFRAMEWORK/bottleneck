package cmd

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"bottleneck/internal/gitinfo"
	"bottleneck/internal/models"
	"bottleneck/internal/scorecard"
	"bottleneck/internal/snapshot"
	"bottleneck/internal/validator"

	"github.com/spf13/cobra"
)

var snapshotEnv string
var snapshotLabel string
var snapshotOut string
var snapshotStrict bool
var snapshotNoLatest bool

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Write a timestamped scorecard snapshot for trend history",
	Long: `Write the current Bottleneck scorecard as a local JSON snapshot.

Snapshots are written to bottleneck/history/scorecards/ and the latest
snapshot for the selected environment is copied to bottleneck/history/latest/.
Snapshot creation succeeds even when release evidence is failing, because a
failing scorecard is useful historical evidence.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		env := strings.TrimSpace(snapshotEnv)
		if env == "" {
			env = "default"
		}

		engine := validator.NewEngine(".", env, validator.WithStrictMode(snapshotStrict))
		result := engine.Validate()
		if configFailure := firstSnapshotConfigFailure(result); configFailure != "" {
			return errors.New(configFailure)
		}

		card := scorecard.Build(result)
		written, err := snapshot.Write(card, snapshot.Options{
			RootPath:    ".",
			OutDir:      snapshotOut,
			Environment: env,
			Label:       snapshotLabel,
			NoLatest:    snapshotNoLatest,
			CreatedAt:   time.Now().UTC(),
			Git:         gitinfo.Detect("."),
		})
		if err != nil {
			return err
		}

		fmt.Println(renderSnapshotSuccess(card, written))
		return nil
	},
}

func init() {
	snapshotCmd.Flags().StringVar(&snapshotEnv, "env", "default", "environment config to use")
	snapshotCmd.Flags().StringVar(&snapshotLabel, "label", "", "optional snapshot label")
	snapshotCmd.Flags().StringVar(&snapshotOut, "out", snapshot.DefaultScorecardDir, "directory for timestamped scorecard snapshots")
	snapshotCmd.Flags().BoolVar(&snapshotStrict, "strict", false, "treat placeholder and insufficient content as failures")
	snapshotCmd.Flags().BoolVar(&snapshotNoLatest, "no-latest", false, "skip updating bottleneck/history/latest/<env>.json")
	rootCmd.AddCommand(snapshotCmd)
}

func firstSnapshotConfigFailure(result models.EngineResult) string {
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

func renderSnapshotSuccess(card scorecard.Scorecard, written snapshot.WriteResult) string {
	latest := pathForDisplay(written.LatestPath)
	if latest == "" {
		latest = "skipped"
	}

	lines := []string{
		"Bottleneck snapshot created",
		"",
		fmt.Sprintf("Environment: %s", written.Snapshot.Snapshot.Environment),
		fmt.Sprintf("Status: %s", card.SystemStatus),
		fmt.Sprintf("Primary bottleneck: %s", card.PrimaryBottleneck),
		fmt.Sprintf("Snapshot: %s", pathForDisplay(written.SnapshotPath)),
		fmt.Sprintf("Latest: %s", latest),
		"",
		"Next:",
		"Commit this snapshot so Bottleneck can compare SDLC evidence over time.",
	}
	return strings.Join(lines, "\n")
}

func pathForDisplay(path string) string {
	if strings.TrimSpace(path) == "" {
		return ""
	}
	clean := filepath.Clean(path)
	if rel, err := filepath.Rel(".", clean); err == nil && !strings.HasPrefix(rel, "..") {
		clean = rel
	}
	return filepath.ToSlash(clean)
}

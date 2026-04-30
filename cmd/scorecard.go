package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"bottleneck/internal/githubactions"
	"bottleneck/internal/githubannotations"
	"bottleneck/internal/models"
	"bottleneck/internal/scorecard"
	"bottleneck/internal/validator"

	"github.com/spf13/cobra"
)

var scorecardEnv string
var scorecardFormat string
var scorecardView string
var scorecardStrict bool
var scorecardGitHubAnnotations bool

var scorecardCmd = &cobra.Command{
	Use:   "scorecard",
	Short: "Summarize the current SDLC scorecard state",
	RunE: func(cmd *cobra.Command, args []string) error {
		engine := validator.NewEngine(".", scorecardEnv, validator.WithStrictMode(scorecardStrict))
		result := engine.Validate()

		githubMetadata := githubactions.DetectFromEnv()
		if githubMetadata.Detected {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			githubactions.EnrichFromEnv(ctx, &githubMetadata, os.Getenv("GITHUB_TOKEN"))
		}

		output, err := scorecard.RenderWithOptions(result, scorecardFormat, scorecard.Options{
			View:   scorecardView,
			GitHub: &githubMetadata,
		})
		if err != nil {
			return err
		}

		fmt.Println(output)
		if scorecardGitHubAnnotations {
			if annotations := githubannotations.Render(result.Results); annotations != "" {
				fmt.Fprintln(os.Stderr, annotations)
			}
		}

		if result.SystemStatus == models.StatusFail {
			os.Exit(1)
		}

		return nil
	},
}

func init() {
	scorecardCmd.Flags().StringVar(&scorecardEnv, "env", "default", "environment config to use")
	scorecardCmd.Flags().StringVar(&scorecardFormat, "format", "text", "output format: text, json, or markdown")
	scorecardCmd.Flags().StringVar(&scorecardView, "view", "engineering", "scorecard view: executive, engineering, or governance")
	scorecardCmd.Flags().BoolVar(&scorecardStrict, "strict", false, "treat placeholder and insufficient content as failures")
	scorecardCmd.Flags().BoolVar(&scorecardGitHubAnnotations, "github-annotations", false, "emit GitHub Actions warning and error annotations")
	rootCmd.AddCommand(scorecardCmd)
}

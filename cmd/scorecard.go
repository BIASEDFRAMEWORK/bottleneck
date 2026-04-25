package cmd

import (
	"fmt"
	"os"

	"biased/internal/models"
	"biased/internal/scorecard"
	"biased/internal/validator"

	"github.com/spf13/cobra"
)

var scorecardEnv string
var scorecardFormat string

var scorecardCmd = &cobra.Command{
	Use:   "scorecard",
	Short: "Summarize the current BIASED validation state",
	RunE: func(cmd *cobra.Command, args []string) error {
		engine := validator.NewEngine(".", scorecardEnv)
		result := engine.Validate()

		output, err := scorecard.Render(result, scorecardFormat)
		if err != nil {
			return err
		}

		fmt.Println(output)

		if result.SystemStatus == models.StatusFail {
			os.Exit(1)
		}

		return nil
	},
}

func init() {
	scorecardCmd.Flags().StringVar(&scorecardEnv, "env", "default", "environment config to use")
	scorecardCmd.Flags().StringVar(&scorecardFormat, "format", "text", "output format: text or json")
	rootCmd.AddCommand(scorecardCmd)
}

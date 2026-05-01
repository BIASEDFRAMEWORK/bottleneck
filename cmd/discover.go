package cmd

import (
	"fmt"
	"os"

	"bottleneck/internal/discover"

	"github.com/spf13/cobra"
)

var discoverFormat string

var discoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Find local SDLC evidence that Bottleneck can use",
	Long: `Find local SDLC evidence artifacts without writing files.

Discovery detects common test, security, telemetry, design, GitHub Actions,
and native Bottleneck evidence files, then recommends ingest commands for
supported formats.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		rootPath, err := os.Getwd()
		if err != nil {
			return err
		}

		result, err := discover.Scan(rootPath)
		if err != nil {
			return err
		}

		switch discoverFormat {
		case "text":
			fmt.Println(discover.RenderText(result))
		case "json":
			encoded, err := discover.MarshalJSON(result)
			if err != nil {
				return err
			}
			fmt.Println(string(encoded))
		default:
			return fmt.Errorf("invalid format %q, expected text or json", discoverFormat)
		}
		return nil
	},
}

func init() {
	discoverCmd.Flags().StringVar(&discoverFormat, "format", "text", "output format: text or json")
	rootCmd.AddCommand(discoverCmd)
}

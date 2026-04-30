package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"bottleneck/internal/traceability"

	"github.com/spf13/cobra"
)

var traceEnv string
var traceFormat string
var traceStrict bool

var traceCmd = &cobra.Command{
	Use:   "trace <id>",
	Short: "Trace evidence links for a framework evidence ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		graph, err := traceability.Build(filepath.Join(".", "bottleneck"), traceability.Options{
			Environment: traceEnv,
			Strict:      traceStrict,
		})
		if err != nil {
			return err
		}

		result, err := graph.Trace(args[0])
		if err != nil {
			return err
		}

		switch strings.ToLower(traceFormat) {
		case "text":
			fmt.Println(traceability.RenderText(result))
		case "json":
			output, err := traceability.RenderJSON(result)
			if err != nil {
				return err
			}
			fmt.Println(output)
		default:
			return fmt.Errorf("unsupported format %q (supported: text, json)", traceFormat)
		}

		return nil
	},
}

func init() {
	traceCmd.Flags().StringVar(&traceEnv, "env", "default", "environment config to use")
	traceCmd.Flags().StringVar(&traceFormat, "format", "text", "output format: text or json")
	traceCmd.Flags().BoolVar(&traceStrict, "strict", false, "treat traceability warnings as failures where applicable")
	rootCmd.AddCommand(traceCmd)
}

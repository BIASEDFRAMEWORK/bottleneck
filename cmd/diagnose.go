package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"bottleneck/internal/config"
	"bottleneck/internal/diagnosis"
	"bottleneck/internal/gate"
	"bottleneck/internal/models"
	"bottleneck/internal/validator"

	"github.com/spf13/cobra"
)

var diagnoseEnv string
var diagnoseFormat string
var diagnoseStrict bool
var diagnoseGate string

var diagnoseCmd = &cobra.Command{
	Use:   "diagnose",
	Short: "Diagnose the primary bottleneck",
	RunE: func(cmd *cobra.Command, args []string) error {
		engine := validator.NewEngine(".", diagnoseEnv, validator.WithStrictMode(diagnoseStrict))
		result := engine.Validate()

		if diagnoseGate != "" {
			output, exitCode, err := renderDiagnoseGate(result, diagnoseGate, diagnoseFormat, diagnoseEnv)
			if err != nil {
				return err
			}
			if output != "" {
				fmt.Println(output)
			}
			if exitCode != 0 {
				os.Exit(exitCode)
			}
			return nil
		}

		output, err := diagnosis.Render(result, diagnoseFormat)
		if err != nil {
			return err
		}

		fmt.Println(output)
		if diagnoseExitCode(result) != 0 {
			os.Exit(diagnoseExitCode(result))
		}

		return nil
	},
}

func init() {
	diagnoseCmd.Flags().StringVar(&diagnoseEnv, "env", "default", "environment config to use")
	diagnoseCmd.Flags().StringVar(&diagnoseFormat, "format", "text", "output format: text, json, markdown, or github")
	diagnoseCmd.Flags().BoolVar(&diagnoseStrict, "strict", false, "treat placeholder and insufficient content as failures")
	diagnoseCmd.Flags().StringVar(&diagnoseGate, "gate", "", "optional gate to evaluate: release")
	rootCmd.AddCommand(diagnoseCmd)
}

func diagnoseExitCode(result models.EngineResult) int {
	if result.SystemStatus == models.StatusFail {
		return 1
	}
	return 0
}

func renderDiagnoseGate(result models.EngineResult, gateName string, format string, env string) (string, int, error) {
	if !strings.EqualFold(gateName, gate.ReleaseGateName) {
		return "", 0, fmt.Errorf("unsupported gate %q (supported: release)", gateName)
	}

	settings := config.DefaultReleaseGateConfig()
	if cfg, err := config.Load(filepath.Join("bottleneck", "config.yaml")); err == nil {
		settings = config.ResolveEnvironment(cfg, env).Gate.Release
	}

	report := gate.EvaluateRelease(result, settings)
	output, err := gate.Render(report, format)
	if err != nil {
		return "", 0, err
	}
	if report.Status == gate.StatusFail {
		return output, 1, nil
	}
	return output, 0, nil
}

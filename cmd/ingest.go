package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"bottleneck/internal/ingest"

	"github.com/spf13/cobra"
)

var ingestFilePath string
var ingestOutPath string
var ingestFormat string
var ingestDryRun bool
var ingestMerge bool

var ingestCmd = &cobra.Command{
	Use:   "ingest",
	Short: "Convert test, security, and telemetry reports into Bottleneck evidence",
}

func init() {
	rootCmd.AddCommand(ingestCmd)

	cucumberCmd := &cobra.Command{
		Use:   "cucumber",
		Short: "Ingest Cucumber JSON test results",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runIngestCommand("cucumber")
		},
	}
	codeqlCmd := &cobra.Command{
		Use:   "codeql",
		Short: "Ingest CodeQL SARIF results",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runIngestCommand("codeql")
		},
	}
	sarifCmd := &cobra.Command{
		Use:   "sarif",
		Short: "Ingest SARIF security scan results",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runIngestCommand("sarif")
		},
	}
	testSummaryCmd := &cobra.Command{
		Use:   "test-summary",
		Short: "Ingest generic test summary JSON",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runIngestCommand("test-summary")
		},
	}
	telemetryCmd := &cobra.Command{
		Use:   "telemetry",
		Short: "Ingest telemetry snapshots",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runIngestCommand("telemetry")
		},
	}

	for _, command := range []*cobra.Command{cucumberCmd, codeqlCmd, sarifCmd, testSummaryCmd, telemetryCmd} {
		command.Flags().StringVar(&ingestFilePath, "file", "", "input file to ingest")
		command.Flags().StringVar(&ingestOutPath, "out", "", "explicit output artifact path")
		command.Flags().BoolVar(&ingestDryRun, "dry-run", false, "parse and print normalized evidence without writing files")
		command.Flags().StringVar(&ingestFormat, "format", "text", "output format: text or json")
		command.Flags().BoolVar(&ingestMerge, "merge", false, "merge with existing artifact evidence instead of replacing it")
		command.MarkFlagRequired("file")
		ingestCmd.AddCommand(command)
	}
}

func runIngestCommand(kind string) error {
	if ingestFormat != "text" && ingestFormat != "json" {
		return fmt.Errorf("invalid format %q, expected text or json", ingestFormat)
	}

	rootPath, err := os.Getwd()
	if err != nil {
		return err
	}

	var outputPath string
	if ingestOutPath != "" {
		outputPath = ingestOutPath
	} else {
		switch kind {
		case "cucumber", "test-summary":
			outputPath = "bottleneck/assurance/results.json"
		case "codeql", "sarif":
			outputPath = "bottleneck/security/guardrails.json"
		case "telemetry":
			outputPath = "bottleneck/execution/telemetry.json"
		}
	}
	outputPath = filepath.Clean(outputPath)

	var summary ingest.IngestSummary
	var writeArtifact bool = !ingestDryRun

	switch kind {
	case "cucumber":
		summary, err = ingest.IngestCucumber(rootPath, ingestFilePath, outputPath, ingestMerge, ingestDryRun)
	case "codeql":
		summary, err = ingest.IngestCodeQL(rootPath, ingestFilePath, outputPath, ingestMerge, ingestDryRun)
	case "sarif":
		summary, err = ingest.IngestSARIF(rootPath, ingestFilePath, outputPath, ingestMerge, ingestDryRun)
	case "test-summary":
		summary, err = ingest.IngestTestSummary(rootPath, ingestFilePath, outputPath, ingestMerge, ingestDryRun)
	case "telemetry":
		summary, err = ingest.IngestTelemetry(rootPath, ingestFilePath, outputPath, ingestMerge, ingestDryRun)
	default:
		err = fmt.Errorf("unknown ingest kind %q", kind)
	}
	if err != nil {
		return err
	}

	if ingestDryRun {
		if ingestFormat == "json" {
			payload := ingest.DryRunPayload{
				Artifact: summary.Artifact,
				Warnings: summary.Warnings,
			}
			encoded, err := ingest.MarshalJSON(payload)
			if err != nil {
				return err
			}
			fmt.Println(string(encoded))
			return nil
		}
		fmt.Println(summary.Text())
		return nil
	}

	if ingestFormat == "json" {
		payload := ingest.OutPayload{
			WrittenPath: outputPath,
			Warnings:    summary.Warnings,
		}
		encoded, err := ingest.MarshalJSON(payload)
		if err != nil {
			return err
		}
		fmt.Println(string(encoded))
		return nil
	}

	fmt.Println(summary.Text())
	if summary.Warnings != nil {
		for _, warning := range summary.Warnings {
			fmt.Printf("WARNING: %s\n", warning)
		}
	}

	if writeArtifact {
		fmt.Printf("Wrote %s\n", outputPath)
	}

	return nil
}

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"bottleneck/internal/discover"
	"bottleneck/internal/ingest"

	"github.com/spf13/cobra"
)

var ingestFilePath string
var ingestOutPath string
var ingestFormat string
var ingestDryRun bool
var ingestMerge bool
var ingestAuto bool

var ingestCmd = &cobra.Command{
	Use:   "ingest",
	Short: "Convert test, security, and telemetry reports into Bottleneck evidence",
	RunE: func(cmd *cobra.Command, args []string) error {
		if ingestAuto {
			return runAutoIngest()
		}
		return cmd.Help()
	},
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
	junitCmd := &cobra.Command{
		Use:   "junit",
		Short: "Ingest JUnit XML test results",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runIngestCommand("junit")
		},
	}
	coverageCmd := &cobra.Command{
		Use:   "coverage",
		Short: "Ingest LCOV coverage results",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runIngestCommand("coverage")
		},
	}
	telemetryCmd := &cobra.Command{
		Use:   "telemetry",
		Short: "Ingest telemetry snapshots",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runIngestCommand("telemetry")
		},
	}

	ingestCmd.Flags().BoolVar(&ingestAuto, "auto", false, "discover and ingest supported local evidence automatically")
	ingestCmd.Flags().BoolVar(&ingestDryRun, "dry-run", false, "parse and print normalized evidence without writing files")
	ingestCmd.Flags().StringVar(&ingestFormat, "format", "text", "output format: text or json")
	ingestCmd.Flags().BoolVar(&ingestMerge, "merge", false, "merge with existing artifact evidence instead of replacing it")

	for _, command := range []*cobra.Command{cucumberCmd, codeqlCmd, sarifCmd, testSummaryCmd, junitCmd, coverageCmd, telemetryCmd} {
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
		case "cucumber", "test-summary", "junit", "coverage":
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
	case "junit":
		summary, err = ingest.IngestJUnit(rootPath, ingestFilePath, outputPath, ingestMerge, ingestDryRun)
	case "coverage":
		summary, err = ingest.IngestCoverage(rootPath, ingestFilePath, outputPath, ingestMerge, ingestDryRun)
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

type autoIngestResult struct {
	Mode       string              `json:"mode"`
	DryRun     bool                `json:"dry_run"`
	Written    []autoIngestWritten `json:"written,omitempty"`
	Skipped    []autoIngestSkipped `json:"skipped,omitempty"`
	Warnings   []string            `json:"warnings,omitempty"`
	NextAction string              `json:"next_action"`
}

type autoIngestWritten struct {
	Kind          string `json:"kind"`
	SourcePath    string `json:"source_path"`
	OutputPath    string `json:"output_path"`
	EvidenceCount int    `json:"evidence_count"`
}

type autoIngestSkipped struct {
	Kind   string `json:"kind"`
	Path   string `json:"path"`
	Reason string `json:"reason"`
}

func runAutoIngest() error {
	if ingestFormat != "text" && ingestFormat != "json" {
		return fmt.Errorf("invalid format %q, expected text or json", ingestFormat)
	}

	rootPath, err := os.Getwd()
	if err != nil {
		return err
	}

	result, err := executeAutoIngest(rootPath, ingestDryRun)
	if err != nil {
		return err
	}

	if ingestFormat == "json" {
		encoded, err := ingest.MarshalJSON(result)
		if err != nil {
			return err
		}
		fmt.Println(string(encoded))
		return nil
	}

	fmt.Println(renderAutoIngestText(result))
	return nil
}

func executeAutoIngest(rootPath string, dryRun bool) (autoIngestResult, error) {
	discovery, err := discover.Scan(rootPath)
	if err != nil {
		return autoIngestResult{}, err
	}

	result := autoIngestResult{
		Mode:       "auto",
		DryRun:     dryRun,
		Warnings:   append([]string{}, discovery.Warnings...),
		NextAction: "Run `bottleneck assess` to see maturity, release friction, and the primary bottleneck.",
	}

	seen := map[string]struct{}{}
	for _, finding := range discovery.Findings {
		if isNativeBottleneckPath(finding.Path) {
			result.Skipped = append(result.Skipped, autoIngestSkipped{
				Kind:   finding.Kind,
				Path:   finding.Path,
				Reason: "already a Bottleneck evidence artifact",
			})
			continue
		}

		outputPath, supported := autoOutputPath(finding.Kind)
		if !supported {
			if finding.SuggestedCommand == "" {
				result.Skipped = append(result.Skipped, autoIngestSkipped{
					Kind:   finding.Kind,
					Path:   finding.Path,
					Reason: "discovered for context but not supported by automatic ingestion yet",
				})
			}
			continue
		}

		key := finding.Kind + "\x00" + finding.Path
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		summary, err := runAutoIngestOne(rootPath, finding.Kind, finding.Path, outputPath, dryRun)
		if err != nil {
			result.Warnings = append(result.Warnings, err.Error())
			result.Skipped = append(result.Skipped, autoIngestSkipped{
				Kind:   finding.Kind,
				Path:   finding.Path,
				Reason: "ingestion failed; inspect the file format and run the suggested ingest command directly",
			})
			continue
		}
		result.Warnings = append(result.Warnings, summary.Warnings...)
		result.Written = append(result.Written, autoIngestWritten{
			Kind:          finding.Kind,
			SourcePath:    finding.Path,
			OutputPath:    outputPath,
			EvidenceCount: evidenceCount(summary.Artifact),
		})
	}

	return result, nil
}

func runAutoIngestOne(rootPath string, kind string, inputPath string, outputPath string, dryRun bool) (ingest.IngestSummary, error) {
	switch kind {
	case "cucumber":
		return ingest.IngestCucumber(rootPath, inputPath, outputPath, true, dryRun)
	case "junit":
		return ingest.IngestJUnit(rootPath, inputPath, outputPath, true, dryRun)
	case "coverage":
		return ingest.IngestCoverage(rootPath, inputPath, outputPath, true, dryRun)
	case "sarif":
		return ingest.IngestSARIF(rootPath, inputPath, outputPath, true, dryRun)
	case "test-summary":
		return ingest.IngestTestSummary(rootPath, inputPath, outputPath, true, dryRun)
	case "telemetry":
		return ingest.IngestTelemetry(rootPath, inputPath, outputPath, true, dryRun)
	default:
		return ingest.IngestSummary{}, fmt.Errorf("unsupported auto-ingest kind %q", kind)
	}
}

func autoOutputPath(kind string) (string, bool) {
	switch kind {
	case "cucumber", "junit", "coverage", "test-summary":
		return ingest.DefaultAssuranceOutput, true
	case "sarif":
		return ingest.DefaultSecurityOutput, true
	case "telemetry":
		return ingest.DefaultExecutionOutput, true
	default:
		return "", false
	}
}

func isNativeBottleneckPath(path string) bool {
	return path == ingest.DefaultAssuranceOutput ||
		path == ingest.DefaultSecurityOutput ||
		path == ingest.DefaultExecutionOutput
}

func evidenceCount(artifact interface{}) int {
	switch value := artifact.(type) {
	case ingest.AssuranceArtifact:
		return len(value.Evidence)
	case ingest.SecurityArtifact:
		return len(value.Evidence)
	case ingest.ExecutionArtifact:
		return len(value.Evidence)
	default:
		return 0
	}
}

func renderAutoIngestText(result autoIngestResult) string {
	lines := []string{"Bottleneck Auto Evidence Ingest"}
	if result.DryRun {
		lines = append(lines, "Mode: dry-run")
	} else {
		lines = append(lines, "Mode: write")
	}
	if len(result.Written) == 0 {
		lines = append(lines, "No supported evidence was ingested.")
	} else {
		lines = append(lines, "Ingested:")
		for _, item := range result.Written {
			verb := "Wrote"
			if result.DryRun {
				verb = "Would write"
			}
			lines = append(lines, fmt.Sprintf("  - %s %s from %s (%d evidence items)", verb, item.OutputPath, item.SourcePath, item.EvidenceCount))
		}
	}
	if len(result.Skipped) > 0 {
		lines = append(lines, "Skipped:")
		for _, item := range result.Skipped {
			lines = append(lines, fmt.Sprintf("  - %s (%s): %s", item.Path, item.Kind, item.Reason))
		}
	}
	if len(result.Warnings) > 0 {
		lines = append(lines, "Warnings:")
		for _, warning := range result.Warnings {
			lines = append(lines, "  - "+warning)
		}
	}
	lines = append(lines, "Next action: "+result.NextAction)
	return strings.Join(lines, "\n")
}

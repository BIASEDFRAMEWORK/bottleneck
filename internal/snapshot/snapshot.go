package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"bottleneck/internal/gitinfo"
	"bottleneck/internal/scorecard"
)

const (
	SchemaVersion        = "scorecard.snapshot.v1"
	Source               = "bottleneck snapshot"
	DefaultScorecardDir  = "bottleneck/history/scorecards"
	DefaultLatestDir     = "bottleneck/history/latest"
	timestampFileLayout  = "2006-01-02T150405Z"
	timestampIDLayout    = "20060102-150405"
	timestampJSONLayout  = time.RFC3339
	defaultFileMode      = 0o644
	defaultDirectoryMode = 0o755
)

var labelSeparatorPattern = regexp.MustCompile(`[^a-z0-9]+`)

type Metadata struct {
	ID          string       `json:"id"`
	CreatedAt   string       `json:"created_at"`
	Environment string       `json:"environment"`
	Label       string       `json:"label,omitempty"`
	Source      string       `json:"source"`
	Git         gitinfo.Info `json:"git"`
}

type Snapshot struct {
	SchemaVersion string              `json:"schema_version"`
	Snapshot      Metadata            `json:"snapshot"`
	Scorecard     scorecard.Scorecard `json:"scorecard"`
}

type Options struct {
	RootPath    string
	OutDir      string
	LatestDir   string
	Environment string
	Label       string
	NoLatest    bool
	CreatedAt   time.Time
	Git         gitinfo.Info
}

type WriteResult struct {
	Snapshot     Snapshot
	SnapshotPath string
	LatestPath   string
}

func Write(card scorecard.Scorecard, options Options) (WriteResult, error) {
	options = normalizeOptions(card, options)
	snapshot := Build(card, options)

	snapshotPath := TimestampedPath(options.RootPath, options.OutDir, options.CreatedAt, options.Environment, snapshot.Snapshot.Label)
	if err := writeJSON(snapshotPath, snapshot); err != nil {
		return WriteResult{}, err
	}

	result := WriteResult{
		Snapshot:     snapshot,
		SnapshotPath: snapshotPath,
	}
	if !options.NoLatest {
		latestPath := LatestPath(options.RootPath, options.LatestDir, options.Environment)
		if err := writeJSON(latestPath, snapshot); err != nil {
			return WriteResult{}, err
		}
		result.LatestPath = latestPath
	}

	return result, nil
}

func Build(card scorecard.Scorecard, options Options) Snapshot {
	options = normalizeOptions(card, options)
	card = scorecard.EnsureStableContract(card)
	label := SanitizeLabel(options.Label)
	return Snapshot{
		SchemaVersion: SchemaVersion,
		Snapshot: Metadata{
			ID:          "SNAPSHOT-" + options.CreatedAt.UTC().Format(timestampIDLayout),
			CreatedAt:   options.CreatedAt.UTC().Format(timestampJSONLayout),
			Environment: options.Environment,
			Label:       label,
			Source:      Source,
			Git:         options.Git,
		},
		Scorecard: card,
	}
}

func TimestampedPath(rootPath string, outDir string, createdAt time.Time, environment string, label string) string {
	if outDir == "" {
		outDir = DefaultScorecardDir
	}
	filename := TimestampedFilename(createdAt, environment, label)
	return resolvePath(rootPath, filepath.Join(outDir, filename))
}

func TimestampedFilename(createdAt time.Time, environment string, label string) string {
	parts := []string{
		createdAt.UTC().Format(timestampFileLayout),
		SanitizeFilenamePart(environment, "default"),
	}
	if sanitizedLabel := SanitizeLabel(label); sanitizedLabel != "" {
		parts = append(parts, sanitizedLabel)
	}
	parts = append(parts, "scorecard")
	return strings.Join(parts, "-") + ".json"
}

func LatestPath(rootPath string, latestDir string, environment string) string {
	if latestDir == "" {
		latestDir = DefaultLatestDir
	}
	return resolvePath(rootPath, filepath.Join(latestDir, SanitizeFilenamePart(environment, "default")+".json"))
}

func SanitizeLabel(label string) string {
	return SanitizeFilenamePart(label, "")
}

func SanitizeFilenamePart(value string, fallback string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = labelSeparatorPattern.ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	if value == "" {
		return fallback
	}
	return value
}

func normalizeOptions(card scorecard.Scorecard, options Options) Options {
	if options.RootPath == "" {
		options.RootPath = "."
	}
	if strings.TrimSpace(options.Environment) == "" {
		options.Environment = card.Environment
	}
	if strings.TrimSpace(options.Environment) == "" {
		options.Environment = "default"
	}
	if options.CreatedAt.IsZero() {
		options.CreatedAt = time.Now().UTC()
	} else {
		options.CreatedAt = options.CreatedAt.UTC()
	}
	return options
}

func writeJSON(path string, snapshot Snapshot) error {
	if err := os.MkdirAll(filepath.Dir(path), defaultDirectoryMode); err != nil {
		return fmt.Errorf("create snapshot directory %s: %w", filepath.Dir(path), err)
	}

	content, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal snapshot json: %w", err)
	}
	return os.WriteFile(path, append(content, '\n'), defaultFileMode)
}

func resolvePath(rootPath string, path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	if rootPath == "" {
		rootPath = "."
	}
	return filepath.Clean(filepath.Join(rootPath, path))
}

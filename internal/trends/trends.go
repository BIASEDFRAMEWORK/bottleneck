package trends

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"bottleneck/internal/snapshot"
)

const (
	DefaultWindow = 6

	FormatText     = "text"
	FormatMarkdown = "markdown"
	FormatJSON     = "json"
)

type Direction string

const (
	DirectionImproving           Direction = "improving"
	DirectionDeclining           Direction = "declining"
	DirectionStable              Direction = "stable"
	DirectionRecovered           Direction = "recovered"
	DirectionRegressed           Direction = "regressed"
	DirectionInsufficientHistory Direction = "insufficient_history"
)

type Options struct {
	RootPath    string
	HistoryDir  string
	Environment string
	Window      int
}

type Analysis struct {
	Environment              string               `json:"environment"`
	SnapshotCount            int                  `json:"snapshot_count"`
	Window                   int                  `json:"window"`
	OverallDirection         Direction            `json:"overall_direction"`
	CurrentStatus            string               `json:"current_status"`
	CurrentPrimaryBottleneck string               `json:"current_primary_bottleneck"`
	CategoryTrends           []CategoryTrend      `json:"category_trends"`
	PersistentBottleneck     PersistentBottleneck `json:"persistent_bottleneck"`
	LeadershipSummary        string               `json:"leadership_summary"`
	Warnings                 []string             `json:"warnings,omitempty"`
}

type CategoryTrend struct {
	Category      string    `json:"category"`
	Values        []float64 `json:"values"`
	Statuses      []string  `json:"statuses"`
	Direction     Direction `json:"direction"`
	Delta         float64   `json:"delta"`
	CurrentValue  float64   `json:"current_value"`
	PreviousValue float64   `json:"previous_value"`
	Summary       string    `json:"summary"`
}

type PersistentBottleneck struct {
	Category string `json:"category"`
	Count    int    `json:"count"`
	Total    int    `json:"total"`
	Summary  string `json:"summary"`
}

type SnapshotRecord struct {
	Path              string
	CreatedAt         time.Time
	Environment       string
	SystemStatus      string
	PrimaryBottleneck string
	Categories        map[string]CategoryPoint
}

type CategoryPoint struct {
	Score  *float64
	Status string
}

type snapshotFile struct {
	Snapshot struct {
		CreatedAt   string `json:"created_at"`
		Environment string `json:"environment"`
	} `json:"snapshot"`
	Scorecard struct {
		Environment       string `json:"environment"`
		SystemStatus      string `json:"system_status"`
		PrimaryBottleneck string `json:"primary_bottleneck"`
		Diagnosis         struct {
			CategoryScores []struct {
				Category string  `json:"category"`
				Score    float64 `json:"score"`
				Status   string  `json:"status"`
			} `json:"category_scores"`
		} `json:"diagnosis"`
		Capabilities []struct {
			Capability string  `json:"capability"`
			Status     string  `json:"status"`
			Score      float64 `json:"score"`
		} `json:"capabilities"`
		Categories []struct {
			Category   string  `json:"category"`
			Capability string  `json:"capability"`
			Name       string  `json:"name"`
			Status     string  `json:"status"`
			Score      float64 `json:"score"`
		} `json:"categories"`
	} `json:"scorecard"`
}

var categoryOrder = []string{"Intent", "Behavior", "Design", "Assurance", "Security", "Execution"}

func Analyze(options Options) (Analysis, error) {
	options = normalizeOptions(options)
	records, warnings, err := ReadSnapshots(options)
	if err != nil {
		return Analysis{}, err
	}
	analysis := AnalyzeRecords(records, options)
	analysis.Warnings = append(analysis.Warnings, warnings...)
	return analysis, nil
}

func ReadSnapshots(options Options) ([]SnapshotRecord, []string, error) {
	options = normalizeOptions(options)
	historyDir := resolvePath(options.RootPath, options.HistoryDir)
	entries, err := os.ReadDir(historyDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("read snapshot history %s: %w", historyDir, err)
	}

	records := []SnapshotRecord{}
	warnings := []string{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".json") {
			continue
		}
		path := filepath.Join(historyDir, entry.Name())
		record, err := readSnapshotFile(path)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("ignored malformed snapshot %s: %v", displayPath(options.RootPath, path), err))
			continue
		}
		if record.Environment != options.Environment {
			continue
		}
		records = append(records, record)
	}

	sort.SliceStable(records, func(i int, j int) bool {
		if !records[i].CreatedAt.Equal(records[j].CreatedAt) {
			return records[i].CreatedAt.Before(records[j].CreatedAt)
		}
		return records[i].Path < records[j].Path
	})
	return records, warnings, nil
}

func AnalyzeRecords(records []SnapshotRecord, options Options) Analysis {
	options = normalizeOptions(options)
	records = latestWindow(records, options.Window)

	analysis := Analysis{
		Environment:      options.Environment,
		SnapshotCount:    len(records),
		Window:           options.Window,
		CategoryTrends:   categoryTrends(records),
		OverallDirection: DirectionInsufficientHistory,
	}
	if len(records) > 0 {
		latest := records[len(records)-1]
		analysis.CurrentStatus = latest.SystemStatus
		analysis.CurrentPrimaryBottleneck = latest.PrimaryBottleneck
	}
	analysis.PersistentBottleneck = persistentBottleneck(records)
	analysis.OverallDirection = overallDirection(analysis.SnapshotCount, analysis.CategoryTrends)
	analysis.LeadershipSummary = leadershipSummary(analysis)
	return analysis
}

func readSnapshotFile(path string) (SnapshotRecord, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return SnapshotRecord{}, err
	}

	var file snapshotFile
	if err := json.Unmarshal(content, &file); err != nil {
		return SnapshotRecord{}, err
	}

	createdAt, err := time.Parse(time.RFC3339, strings.TrimSpace(file.Snapshot.CreatedAt))
	if err != nil {
		return SnapshotRecord{}, fmt.Errorf("invalid snapshot.created_at")
	}

	environment := strings.TrimSpace(file.Snapshot.Environment)
	if environment == "" {
		environment = strings.TrimSpace(file.Scorecard.Environment)
	}
	if environment == "" {
		environment = "default"
	}

	return SnapshotRecord{
		Path:              path,
		CreatedAt:         createdAt.UTC(),
		Environment:       environment,
		SystemStatus:      normalizeStatus(file.Scorecard.SystemStatus),
		PrimaryBottleneck: strings.TrimSpace(file.Scorecard.PrimaryBottleneck),
		Categories:        extractCategories(file),
	}, nil
}

func extractCategories(file snapshotFile) map[string]CategoryPoint {
	categories := map[string]CategoryPoint{}
	for _, score := range file.Scorecard.Diagnosis.CategoryScores {
		category := strings.TrimSpace(score.Category)
		if category == "" {
			continue
		}
		value := score.Score
		categories[category] = CategoryPoint{Score: &value, Status: normalizeStatus(score.Status)}
	}
	if len(categories) > 0 {
		return categories
	}

	for _, capability := range file.Scorecard.Capabilities {
		category := strings.TrimSpace(capability.Capability)
		if category == "" {
			continue
		}
		value := capability.Score
		point := CategoryPoint{Status: normalizeStatus(capability.Status)}
		if value != 0 || point.Status == statusFail {
			point.Score = &value
		}
		if point.Score == nil {
			point.Score = scoreFromStatus(point.Status)
		}
		categories[category] = point
	}
	if len(categories) > 0 {
		return categories
	}

	for _, categoryScore := range file.Scorecard.Categories {
		category := strings.TrimSpace(categoryScore.Category)
		if category == "" {
			category = strings.TrimSpace(categoryScore.Capability)
		}
		if category == "" {
			category = strings.TrimSpace(categoryScore.Name)
		}
		if category == "" {
			continue
		}
		value := categoryScore.Score
		point := CategoryPoint{Status: normalizeStatus(categoryScore.Status)}
		if value != 0 || point.Status == statusFail {
			point.Score = &value
		}
		if point.Score == nil {
			point.Score = scoreFromStatus(point.Status)
		}
		categories[category] = point
	}
	return categories
}

func categoryTrends(records []SnapshotRecord) []CategoryTrend {
	categories := orderedCategories(records)
	trends := make([]CategoryTrend, 0, len(categories))
	for _, category := range categories {
		values := []float64{}
		statuses := []string{}
		for _, record := range records {
			point, ok := record.Categories[category]
			if !ok {
				statuses = append(statuses, statusUnknown)
				continue
			}
			statuses = append(statuses, point.Status)
			if point.Score != nil {
				values = append(values, *point.Score)
			} else if score := scoreFromStatus(point.Status); score != nil {
				values = append(values, *score)
			}
		}
		trend := CategoryTrend{
			Category:  category,
			Values:    values,
			Statuses:  statuses,
			Direction: directionFor(len(records), values, statuses),
			Summary:   categorySummary(category, values, statuses),
		}
		if len(values) > 0 {
			trend.CurrentValue = values[len(values)-1]
		}
		if len(values) > 1 {
			trend.PreviousValue = values[len(values)-2]
			trend.Delta = values[len(values)-1] - values[0]
		}
		trends = append(trends, trend)
	}
	return trends
}

func orderedCategories(records []SnapshotRecord) []string {
	seen := map[string]struct{}{}
	for _, record := range records {
		for category := range record.Categories {
			seen[category] = struct{}{}
		}
	}
	ordered := []string{}
	for _, category := range categoryOrder {
		if _, ok := seen[category]; ok {
			ordered = append(ordered, category)
			delete(seen, category)
		}
	}
	extra := make([]string, 0, len(seen))
	for category := range seen {
		extra = append(extra, category)
	}
	sort.Strings(extra)
	return append(ordered, extra...)
}

func directionFor(snapshotCount int, values []float64, statuses []string) Direction {
	if snapshotCount < 2 || len(values) < 2 {
		return DirectionInsufficientHistory
	}

	latestStatus := latestKnownStatus(statuses)
	earlierStatuses := knownStatusesBeforeLatest(statuses)
	if statusIsPassEarlier(earlierStatuses) && (latestStatus == statusWarn || latestStatus == statusFail) {
		return DirectionRegressed
	}
	if statusIsWeakEarlier(earlierStatuses) && latestStatus == statusPass {
		return DirectionRecovered
	}

	delta := values[len(values)-1] - values[0]
	switch {
	case delta >= 10:
		return DirectionImproving
	case delta <= -10:
		return DirectionDeclining
	default:
		return DirectionStable
	}
}

func overallDirection(snapshotCount int, trends []CategoryTrend) Direction {
	if snapshotCount < 2 {
		return DirectionInsufficientHistory
	}

	positive := 0
	negative := 0
	stable := 0
	for _, trend := range trends {
		switch trend.Direction {
		case DirectionImproving, DirectionRecovered:
			positive++
		case DirectionDeclining, DirectionRegressed:
			negative++
		case DirectionStable:
			stable++
		}
	}

	switch {
	case positive > negative:
		return DirectionImproving
	case negative > positive:
		return DirectionDeclining
	case stable >= positive+negative:
		return DirectionStable
	default:
		return DirectionStable
	}
}

func persistentBottleneck(records []SnapshotRecord) PersistentBottleneck {
	counts := map[string]int{}
	for _, record := range records {
		category := strings.TrimSpace(record.PrimaryBottleneck)
		if ignoredBottleneck(category) {
			continue
		}
		counts[category]++
	}

	persistent := PersistentBottleneck{Total: len(records)}
	for _, category := range orderedBottleneckCategories(counts) {
		if counts[category] > persistent.Count {
			persistent.Category = category
			persistent.Count = counts[category]
		}
	}
	if persistent.Category == "" {
		persistent.Summary = fmt.Sprintf("No persistent primary bottleneck was detected across %d snapshots.", len(records))
		return persistent
	}
	persistent.Summary = fmt.Sprintf("%s appeared as the primary bottleneck in %d of %d snapshots.", persistent.Category, persistent.Count, len(records))
	return persistent
}

func orderedBottleneckCategories(counts map[string]int) []string {
	seen := map[string]struct{}{}
	for category := range counts {
		seen[category] = struct{}{}
	}
	ordered := []string{}
	for _, category := range categoryOrder {
		if _, ok := seen[category]; ok {
			ordered = append(ordered, category)
			delete(seen, category)
		}
	}
	extra := make([]string, 0, len(seen))
	for category := range seen {
		extra = append(extra, category)
	}
	sort.Strings(extra)
	return append(ordered, extra...)
}

func leadershipSummary(analysis Analysis) string {
	if analysis.OverallDirection == DirectionInsufficientHistory {
		return "Not enough historical snapshots exist yet. Create snapshots over multiple delivery cycles to establish a trend."
	}
	if analysis.OverallDirection == DirectionDeclining {
		return "The team's SDLC evidence is declining. Leadership should review the categories with recent regressions before approving additional release acceleration."
	}
	if analysis.OverallDirection == DirectionImproving && analysis.PersistentBottleneck.Category != "" {
		return fmt.Sprintf("The team is improving overall, but %s remains the most persistent delivery constraint.", analysis.PersistentBottleneck.Category)
	}
	if analysis.OverallDirection == DirectionImproving {
		return "The team's SDLC evidence is improving across the selected snapshot window."
	}
	if analysis.OverallDirection == DirectionStable && weakCurrentPosture(analysis) {
		return "The team is stable but not improving. The current bottleneck has persisted across multiple snapshots and likely requires deliberate investment."
	}
	return "The team is stable across the selected snapshot window."
}

func categorySummary(category string, values []float64, statuses []string) string {
	if len(values) < 2 {
		return fmt.Sprintf("%s does not have enough historical data for a trend.", category)
	}
	return fmt.Sprintf("%s changed by %s points across the selected window.", category, formatDelta(values[len(values)-1]-values[0]))
}

func latestWindow(records []SnapshotRecord, window int) []SnapshotRecord {
	if window <= 0 {
		window = DefaultWindow
	}
	if len(records) <= window {
		return append([]SnapshotRecord{}, records...)
	}
	return append([]SnapshotRecord{}, records[len(records)-window:]...)
}

func normalizeOptions(options Options) Options {
	if strings.TrimSpace(options.RootPath) == "" {
		options.RootPath = "."
	}
	if strings.TrimSpace(options.HistoryDir) == "" {
		options.HistoryDir = snapshot.DefaultScorecardDir
	}
	if strings.TrimSpace(options.Environment) == "" {
		options.Environment = "default"
	}
	if options.Window <= 0 {
		options.Window = DefaultWindow
	}
	return options
}

const (
	statusPass    = "PASS"
	statusWarn    = "WARN"
	statusFail    = "FAIL"
	statusUnknown = "UNKNOWN"
)

func normalizeStatus(status string) string {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "PASS", "PASSED", "OK":
		return statusPass
	case "WARN", "WARNING":
		return statusWarn
	case "FAIL", "FAILED", "ERROR":
		return statusFail
	default:
		return statusUnknown
	}
}

func scoreFromStatus(status string) *float64 {
	var score float64
	switch normalizeStatus(status) {
	case statusPass:
		score = 100
	case statusWarn:
		score = 60
	case statusFail:
		score = 0
	default:
		return nil
	}
	return &score
}

func latestKnownStatus(statuses []string) string {
	for i := len(statuses) - 1; i >= 0; i-- {
		status := normalizeStatus(statuses[i])
		if status != statusUnknown {
			return status
		}
	}
	return statusUnknown
}

func knownStatusesBeforeLatest(statuses []string) []string {
	if len(statuses) <= 1 {
		return nil
	}
	latestKnownIndex := -1
	for i := len(statuses) - 1; i >= 0; i-- {
		if normalizeStatus(statuses[i]) != statusUnknown {
			latestKnownIndex = i
			break
		}
	}
	if latestKnownIndex <= 0 {
		return nil
	}
	return statuses[:latestKnownIndex]
}

func statusIsPassEarlier(statuses []string) bool {
	for _, status := range statuses {
		if normalizeStatus(status) == statusPass {
			return true
		}
	}
	return false
}

func statusIsWeakEarlier(statuses []string) bool {
	for _, status := range statuses {
		normalized := normalizeStatus(status)
		if normalized == statusWarn || normalized == statusFail {
			return true
		}
	}
	return false
}

func ignoredBottleneck(category string) bool {
	normalized := strings.ToLower(strings.TrimSpace(category))
	return normalized == "" || normalized == "none" || normalized == "no bottleneck"
}

func weakCurrentPosture(analysis Analysis) bool {
	if !ignoredBottleneck(analysis.CurrentPrimaryBottleneck) {
		return true
	}
	status := normalizeStatus(analysis.CurrentStatus)
	return status == statusWarn || status == statusFail || status == statusUnknown
}

func resolvePath(rootPath string, path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Clean(filepath.Join(rootPath, path))
}

func displayPath(rootPath string, path string) string {
	if rel, err := filepath.Rel(rootPath, path); err == nil && !strings.HasPrefix(rel, "..") {
		return filepath.ToSlash(rel)
	}
	return filepath.ToSlash(filepath.Clean(path))
}

func formatDelta(delta float64) string {
	if delta > 0 {
		return "+" + formatScore(delta)
	}
	return formatScore(delta)
}

func formatScore(value float64) string {
	if math.Abs(value-math.Round(value)) < 0.000001 {
		return fmt.Sprintf("%.0f", value)
	}
	return fmt.Sprintf("%.1f", value)
}

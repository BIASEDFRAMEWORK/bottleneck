package seedhistory

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"bottleneck/internal/diagnosis"
	"bottleneck/internal/gitinfo"
	"bottleneck/internal/models"
	"bottleneck/internal/scorecard"
	"bottleneck/internal/snapshot"
)

const (
	ScenarioSaaSDayOne    = "saas-day-one"
	DefaultSnapshotCount  = 6
	DefaultEnvironment    = "default"
	deliveryCycleSpacing  = 7 * 24 * time.Hour
	defaultDirectoryMode  = 0o755
	defaultSeedConfidence = "High"
)

var categoryOrder = []string{"Intent", "Behavior", "Design", "Assurance", "Security", "Execution"}

type Options struct {
	RootPath      string
	Scenario      string
	Environment   string
	SnapshotCount int
	OutDir        string
	Overwrite     bool
	Now           time.Time
	Git           gitinfo.Info
}

type Result struct {
	Scenario    string
	Environment string
	OutDir      string
	Snapshots   []snapshot.WriteResult
}

type scenarioPoint struct {
	Label             string
	SystemStatus      string
	PrimaryBottleneck string
	CategoryStatuses  map[string]string
	Narrative         string
}

func Write(options Options) (Result, error) {
	options = normalizeOptions(options)
	allPoints, err := scenario(options.Scenario)
	if err != nil {
		return Result{}, err
	}
	if options.SnapshotCount < 1 {
		return Result{}, errors.New("snapshots must be greater than 0")
	}
	if options.SnapshotCount > len(allPoints) {
		return Result{}, fmt.Errorf("scenario %q supports at most %d snapshots", options.Scenario, len(allPoints))
	}

	points := append([]scenarioPoint{}, allPoints[:options.SnapshotCount]...)
	if err := ensureOutputWritable(options); err != nil {
		return Result{}, err
	}
	if options.Overwrite {
		if err := removeExistingScenarioFiles(options, allPoints); err != nil {
			return Result{}, err
		}
	}

	end := options.Now.UTC().Truncate(time.Second)
	start := end.Add(-time.Duration(len(points)-1) * deliveryCycleSpacing)
	written := make([]snapshot.WriteResult, 0, len(points))
	for index, point := range points {
		createdAt := start.Add(time.Duration(index) * deliveryCycleSpacing)
		card := scorecardFor(point, options.Environment)
		result, err := snapshot.Write(card, snapshot.Options{
			RootPath:    options.RootPath,
			OutDir:      options.OutDir,
			Environment: options.Environment,
			Label:       point.Label,
			NoLatest:    true,
			CreatedAt:   createdAt,
			Git:         options.Git,
		})
		if err != nil {
			return Result{}, err
		}
		written = append(written, result)
	}

	return Result{
		Scenario:    options.Scenario,
		Environment: options.Environment,
		OutDir:      options.OutDir,
		Snapshots:   written,
	}, nil
}

func normalizeOptions(options Options) Options {
	if strings.TrimSpace(options.RootPath) == "" {
		options.RootPath = "."
	}
	if strings.TrimSpace(options.Scenario) == "" {
		options.Scenario = ScenarioSaaSDayOne
	}
	if strings.TrimSpace(options.Environment) == "" {
		options.Environment = DefaultEnvironment
	}
	if strings.TrimSpace(options.OutDir) == "" {
		options.OutDir = snapshot.DefaultScorecardDir
	}
	if options.SnapshotCount == 0 {
		options.SnapshotCount = DefaultSnapshotCount
	}
	if options.Now.IsZero() {
		options.Now = time.Now().UTC()
	}
	return options
}

func scenario(name string) ([]scenarioPoint, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case ScenarioSaaSDayOne:
		return saasDayOneScenario(), nil
	default:
		return nil, fmt.Errorf("unsupported seed-history scenario %q (supported: %s)", name, ScenarioSaaSDayOne)
	}
}

func ensureOutputWritable(options Options) error {
	outDir := resolvePath(options.RootPath, options.OutDir)
	entries, err := os.ReadDir(outDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read seed history output %s: %w", outDir, err)
	}
	if len(entries) > 0 && !options.Overwrite {
		return fmt.Errorf("Seed history already exists in %s.\nNext action: use --overwrite or choose a different --out path.", displayPath(options.RootPath, outDir))
	}
	if options.Overwrite {
		return os.MkdirAll(outDir, defaultDirectoryMode)
	}
	return nil
}

func removeExistingScenarioFiles(options Options, points []scenarioPoint) error {
	outDir := resolvePath(options.RootPath, options.OutDir)
	entries, err := os.ReadDir(outDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read seed history output %s: %w", outDir, err)
	}

	labels := map[string]struct{}{}
	for _, point := range points {
		labels[snapshot.SanitizeLabel(point.Label)] = struct{}{}
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), "-scorecard.json") {
			continue
		}
		if !filenameHasScenarioLabel(entry.Name(), labels) {
			continue
		}
		path := filepath.Join(outDir, entry.Name())
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("remove existing seed snapshot %s: %w", displayPath(options.RootPath, path), err)
		}
	}
	return nil
}

func filenameHasScenarioLabel(filename string, labels map[string]struct{}) bool {
	for label := range labels {
		if strings.Contains(filename, "-"+label+"-scorecard.json") {
			return true
		}
	}
	return false
}

func scorecardFor(point scenarioPoint, environment string) scorecard.Scorecard {
	categoryScores := make([]diagnosis.CategoryScore, 0, len(categoryOrder))
	capabilities := make([]scorecard.CapabilityScorecard, 0, len(categoryOrder))
	for _, category := range categoryOrder {
		status := normalizeStatus(point.CategoryStatuses[category])
		score := scoreForStatus(status)
		categoryScores = append(categoryScores, diagnosis.CategoryScore{
			Category: category,
			Score:    score,
			Status:   status,
			Reason:   reasonFor(category, status),
		})
		capabilities = append(capabilities, capabilityFor(category, status, score))
	}

	return scorecard.EnsureStableContract(scorecard.Scorecard{
		SchemaVersion:         scorecard.SchemaVersion,
		Environment:           environment,
		SystemStatus:          point.SystemStatus,
		ReleaseRecommendation: recommendationFor(point.SystemStatus),
		PrimaryBottleneck:     point.PrimaryBottleneck,
		Diagnosis: diagnosis.Diagnosis{
			PrimaryBottleneck: point.PrimaryBottleneck,
			Rule:              "seed-history/saas-day-one",
			Reason:            point.Narrative,
			Impact:            impactFor(point.PrimaryBottleneck),
			NextAction:        nextActionFor(point.PrimaryBottleneck),
			RecommendedAction: recommendedActionFor(point.PrimaryBottleneck),
			WhyItMatters:      "Historical evidence lets leaders see whether SDLC capability is improving, regressing, or blocked by a persistent constraint.",
			Confidence:        defaultSeedConfidence,
			ConfidenceReason:  "Seed history uses deterministic scenario data.",
			CategoryScores:    categoryScores,
		},
		Capabilities: capabilities,
		BottomLine:   point.Narrative,
	})
}

func capabilityFor(category string, status string, score int) scorecard.CapabilityScorecard {
	capability := scorecard.CapabilityScorecard{
		Capability:        category,
		Status:            status,
		Score:             score,
		Owner:             ownerFor(category),
		Bottleneck:        bottleneckFor(category),
		Reason:            reasonFor(category, status),
		RecommendedAction: recommendedActionFor(category),
	}
	if status == scorecard.StatusPass {
		capability.EvidenceCount = 2
		capability.Evidence = []string{passingEvidenceFor(category)}
		capability.RecommendedAction = passActionFor(category)
		return capability
	}
	capability.EvidenceCount = 1
	capability.Evidence = []string{partialEvidenceFor(category)}
	capability.MissingEvidence = []string{missingEvidenceFor(category)}
	capability.ScoreImpacts = []models.ScoreImpact{{
		Reason: reasonFor(category, status),
		Delta:  score - 100,
	}}
	return capability
}

func saasDayOneScenario() []scenarioPoint {
	return []scenarioPoint{
		{
			Label:             "01-fast-demo-weak-evidence",
			SystemStatus:      scorecard.StatusFail,
			PrimaryBottleneck: "Intent",
			CategoryStatuses: map[string]string{
				"Intent": scorecard.StatusFail, "Behavior": scorecard.StatusFail, "Design": scorecard.StatusWarn,
				"Assurance": scorecard.StatusFail, "Security": scorecard.StatusWarn, "Execution": scorecard.StatusFail,
			},
			Narrative: "The team has code, but cannot clearly explain what outcome the system is designed to produce.",
		},
		{
			Label:             "02-intent-clarified",
			SystemStatus:      scorecard.StatusFail,
			PrimaryBottleneck: "Behavior",
			CategoryStatuses: map[string]string{
				"Intent": scorecard.StatusPass, "Behavior": scorecard.StatusFail, "Design": scorecard.StatusWarn,
				"Assurance": scorecard.StatusFail, "Security": scorecard.StatusWarn, "Execution": scorecard.StatusFail,
			},
			Narrative: "Intent improved, but expected and unacceptable behaviors are not measurable.",
		},
		{
			Label:             "03-behavior-documented-assurance-weak",
			SystemStatus:      scorecard.StatusWarn,
			PrimaryBottleneck: "Assurance",
			CategoryStatuses: map[string]string{
				"Intent": scorecard.StatusPass, "Behavior": scorecard.StatusPass, "Design": scorecard.StatusWarn,
				"Assurance": scorecard.StatusFail, "Security": scorecard.StatusWarn, "Execution": scorecard.StatusFail,
			},
			Narrative: "Behavior is documented, but tests are not mapped to behavior evidence.",
		},
		{
			Label:             "04-tests-added-security-regresses",
			SystemStatus:      scorecard.StatusWarn,
			PrimaryBottleneck: "Security",
			CategoryStatuses: map[string]string{
				"Intent": scorecard.StatusPass, "Behavior": scorecard.StatusPass, "Design": scorecard.StatusPass,
				"Assurance": scorecard.StatusWarn, "Security": scorecard.StatusFail, "Execution": scorecard.StatusWarn,
			},
			Narrative: "Validation improved, but a high-severity security issue was introduced.",
		},
		{
			Label:             "05-security-recovered-execution-weak",
			SystemStatus:      scorecard.StatusWarn,
			PrimaryBottleneck: "Execution",
			CategoryStatuses: map[string]string{
				"Intent": scorecard.StatusPass, "Behavior": scorecard.StatusPass, "Design": scorecard.StatusPass,
				"Assurance": scorecard.StatusPass, "Security": scorecard.StatusPass, "Execution": scorecard.StatusWarn,
			},
			Narrative: "Pre-release evidence improved, but production/adoption evidence is incomplete.",
		},
		{
			Label:             "06-stable-release-candidate",
			SystemStatus:      scorecard.StatusPass,
			PrimaryBottleneck: diagnosis.HealthyPrimaryBottleneck,
			CategoryStatuses: map[string]string{
				"Intent": scorecard.StatusPass, "Behavior": scorecard.StatusPass, "Design": scorecard.StatusPass,
				"Assurance": scorecard.StatusPass, "Security": scorecard.StatusPass, "Execution": scorecard.StatusPass,
			},
			Narrative: "The delivery system can now prove intent, behavior, assurance, security, and execution evidence together.",
		},
	}
}

func recommendationFor(status string) string {
	switch normalizeStatus(status) {
	case scorecard.StatusPass:
		return scorecard.RecommendationProceed
	case scorecard.StatusWarn:
		return scorecard.RecommendationConditional
	case scorecard.StatusFail:
		return scorecard.RecommendationBlock
	default:
		return scorecard.RecommendationUnknown
	}
}

func scoreForStatus(status string) int {
	switch normalizeStatus(status) {
	case scorecard.StatusPass:
		return 95
	case scorecard.StatusWarn:
		return 65
	case scorecard.StatusFail:
		return 25
	default:
		return 40
	}
}

func normalizeStatus(status string) string {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case scorecard.StatusPass:
		return scorecard.StatusPass
	case scorecard.StatusWarn, "WARNING":
		return scorecard.StatusWarn
	case scorecard.StatusFail:
		return scorecard.StatusFail
	default:
		return scorecard.StatusUnknown
	}
}

func ownerFor(category string) string {
	switch category {
	case "Intent":
		return "Product Lead"
	case "Behavior":
		return "Engineering Lead"
	case "Design":
		return "Architecture Lead"
	case "Assurance":
		return "QA Lead"
	case "Security":
		return "Security Lead"
	case "Execution":
		return "Operations Lead"
	default:
		return "Delivery Lead"
	}
}

func bottleneckFor(category string) string {
	switch category {
	case "Intent":
		return "Unclear outcomes"
	case "Behavior":
		return "Unmeasurable behavior"
	case "Design":
		return "Unreviewed design"
	case "Assurance":
		return "Validation gaps"
	case "Security":
		return "Security risk"
	case "Execution":
		return "Operational readiness"
	default:
		return "Delivery evidence"
	}
}

func reasonFor(category string, status string) string {
	if normalizeStatus(status) == scorecard.StatusPass {
		return fmt.Sprintf("%s evidence is strong enough for the seeded delivery cycle.", category)
	}
	return fmt.Sprintf("%s evidence is not yet strong enough for the seeded delivery cycle.", category)
}

func missingEvidenceFor(category string) string {
	switch category {
	case "Intent":
		return "Add measurable outcomes, constraints, and success criteria."
	case "Behavior":
		return "Add expected and unacceptable behaviors that can be validated."
	case "Design":
		return "Add architecture and operational design evidence."
	case "Assurance":
		return "Map tests to behavior evidence and current results."
	case "Security":
		return "Resolve or formally accept high-severity security findings."
	case "Execution":
		return "Add production telemetry, adoption, and operational readiness evidence."
	default:
		return "Add evidence for this category."
	}
}

func recommendedActionFor(category string) string {
	switch category {
	case "Intent":
		return "Clarify measurable release intent before expanding implementation."
	case "Behavior":
		return "Document measurable expected and unacceptable behaviors."
	case "Design":
		return "Make the system design reviewable before release expansion."
	case "Assurance":
		return "Map validation evidence to the behaviors it proves."
	case "Security":
		return "Resolve high-severity findings or record a formal risk decision."
	case "Execution":
		return "Add production telemetry and adoption evidence before scaling usage."
	case diagnosis.HealthyPrimaryBottleneck:
		return "Continue monitoring evidence trends and maintain current release controls."
	default:
		return "Review the weakest evidence category and add missing proof."
	}
}

func nextActionFor(category string) string {
	if category == diagnosis.HealthyPrimaryBottleneck {
		return "Run bottleneck trends after future delivery cycles to confirm the release posture remains stable."
	}
	return recommendedActionFor(category)
}

func passActionFor(category string) string {
	return fmt.Sprintf("Keep %s evidence current as the delivery system changes.", strings.ToLower(category))
}

func passingEvidenceFor(category string) string {
	return fmt.Sprintf("%s evidence present in seeded history.", category)
}

func partialEvidenceFor(category string) string {
	return fmt.Sprintf("Partial %s evidence present in seeded history.", strings.ToLower(category))
}

func impactFor(category string) string {
	if category == diagnosis.HealthyPrimaryBottleneck {
		return "The seeded release candidate has evidence across the primary SDLC categories."
	}
	return fmt.Sprintf("%s is the primary constraint in this seeded delivery cycle.", category)
}

func resolvePath(rootPath string, path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	if strings.TrimSpace(rootPath) == "" {
		rootPath = "."
	}
	return filepath.Clean(filepath.Join(rootPath, path))
}

func displayPath(rootPath string, path string) string {
	if strings.TrimSpace(path) == "" {
		return ""
	}
	clean := filepath.Clean(path)
	if rel, err := filepath.Rel(rootPath, clean); err == nil && !strings.HasPrefix(rel, "..") {
		clean = rel
	}
	return filepath.ToSlash(clean)
}

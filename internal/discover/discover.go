package discover

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	CategoryAssurance = "assurance"
	CategorySecurity  = "security"
	CategoryExecution = "execution"
	CategoryDesign    = "design"
	CategoryNative    = "native"
)

type DiscoveryResult struct {
	RootPath string             `json:"root_path"`
	Findings []DiscoveryFinding `json:"findings"`
	Summary  Summary            `json:"summary"`
	Warnings []string           `json:"warnings,omitempty"`
}

type DiscoveryFinding struct {
	Category         string   `json:"category"`
	Kind             string   `json:"kind"`
	Path             string   `json:"path"`
	Confidence       string   `json:"confidence"`
	SuggestedCommand string   `json:"suggested_command,omitempty"`
	Notes            []string `json:"notes,omitempty"`
}

type Summary struct {
	CountsByCategory map[string]int `json:"counts_by_category"`
	CountsByKind     map[string]int `json:"counts_by_kind"`
	Missing          []string       `json:"missing,omitempty"`
	TotalFindings    int            `json:"total_findings"`
}

type Options struct {
	RootPath string
}

func Scan(rootPath string) (DiscoveryResult, error) {
	if rootPath == "" {
		rootPath = "."
	}
	rootPath = filepath.Clean(rootPath)

	result := DiscoveryResult{
		RootPath: rootPath,
		Summary: Summary{
			CountsByCategory: map[string]int{},
			CountsByKind:     map[string]int{},
		},
	}

	seen := map[string]struct{}{}
	err := filepath.WalkDir(rootPath, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("could not read %s: %v", displayPath(rootPath, path), err))
			if entry != nil && entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		rel, relErr := filepath.Rel(rootPath, path)
		if relErr != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("could not resolve %s: %v", path, relErr))
			return nil
		}
		rel = filepath.ToSlash(rel)
		if rel == "." {
			return nil
		}

		if entry.IsDir() {
			if shouldSkipDir(rel, entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		for _, finding := range classify(rel) {
			key := finding.Kind + "\x00" + finding.Path
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			result.Findings = append(result.Findings, finding)
		}
		return nil
	})
	if err != nil {
		return DiscoveryResult{}, err
	}

	sort.Slice(result.Findings, func(i, j int) bool {
		if result.Findings[i].Category == result.Findings[j].Category {
			if result.Findings[i].Kind == result.Findings[j].Kind {
				return result.Findings[i].Path < result.Findings[j].Path
			}
			return result.Findings[i].Kind < result.Findings[j].Kind
		}
		return categoryRank(result.Findings[i].Category) < categoryRank(result.Findings[j].Category)
	})

	for _, finding := range result.Findings {
		result.Summary.CountsByCategory[finding.Category]++
		result.Summary.CountsByKind[finding.Kind]++
	}
	result.Summary.TotalFindings = len(result.Findings)
	result.Summary.Missing = missingCategories(result.Summary.CountsByCategory, result.Summary.CountsByKind)

	return result, nil
}

func MarshalJSON(result DiscoveryResult) ([]byte, error) {
	return json.MarshalIndent(result, "", "  ")
}

func RenderText(result DiscoveryResult) string {
	lines := []string{
		"Bottleneck Evidence Discovery",
		fmt.Sprintf("Root: %s", result.RootPath),
		"",
	}

	if len(result.Findings) == 0 {
		lines = append(lines,
			"No recognizable SDLC evidence artifacts found.",
			"",
			"Next action: run `bottleneck init --template saas` or add existing test, security, telemetry, or design evidence.",
		)
		return strings.Join(lines, "\n")
	}

	for _, category := range []string{CategoryAssurance, CategorySecurity, CategoryExecution, CategoryDesign, CategoryNative} {
		findings := findingsForCategory(result.Findings, category)
		if len(findings) == 0 {
			continue
		}
		lines = append(lines, titleCategory(category)+":")
		for _, finding := range findings {
			line := fmt.Sprintf("  - %s (%s, %s confidence)", finding.Path, finding.Kind, finding.Confidence)
			lines = append(lines, line)
			if finding.SuggestedCommand != "" {
				lines = append(lines, "    command: "+finding.SuggestedCommand)
			}
			for _, note := range finding.Notes {
				lines = append(lines, "    note: "+note)
			}
		}
		lines = append(lines, "")
	}

	if len(result.Warnings) > 0 {
		lines = append(lines, "Warnings:")
		for _, warning := range result.Warnings {
			lines = append(lines, "  - "+warning)
		}
		lines = append(lines, "")
	}

	if len(result.Summary.Missing) > 0 {
		lines = append(lines, "Missing:")
		for _, missing := range result.Summary.Missing {
			lines = append(lines, "  - "+missing)
		}
		lines = append(lines, "")
	}

	lines = append(lines, "Next action: run `bottleneck ingest --auto` to normalize supported evidence, then `bottleneck assess`.")
	return strings.TrimRight(strings.Join(lines, "\n"), "\n")
}

func shouldSkipDir(rel string, name string) bool {
	switch name {
	case ".git", "node_modules", "vendor", ".next", "dist":
		return true
	}
	return false
}

func classify(rel string) []DiscoveryFinding {
	lower := strings.ToLower(rel)
	var findings []DiscoveryFinding

	switch {
	case lower == "reports/cucumber.json" || lower == "target/cucumber.json" || lower == "cucumber.json":
		findings = append(findings, finding(CategoryAssurance, "cucumber", rel, "high"))
	case lower == "reports/junit.xml" || pathMatches(lower, "build/test-results/", ".xml") || pathMatches(lower, "test-results/", ".xml"):
		findings = append(findings, finding(CategoryAssurance, "junit", rel, "high"))
	case lower == "coverage/lcov.info":
		findings = append(findings, finding(CategoryAssurance, "coverage", rel, "medium"))
	case lower == "coverage/cobertura.xml":
		findings = append(findings, findingWithNotes(CategoryAssurance, "cobertura", rel, "medium", "Cobertura discovery is available; LCOV ingestion is supported in this release."))
	case lower == "coverage/coverage-final.json":
		findings = append(findings, findingWithNotes(CategoryAssurance, "coverage-final", rel, "medium", "Istanbul coverage JSON is discoverable; LCOV ingestion is supported in this release."))
	case lower == "playwright-report/results.json":
		findings = append(findings, findingWithNotes(CategoryAssurance, "playwright", rel, "medium", "Playwright result discovery is available; use JUnit or test-summary ingestion for normalization."))
	case lower == "reports/test-summary.json" || lower == "test-summary.json":
		findings = append(findings, finding(CategoryAssurance, "test-summary", rel, "high"))
	case strings.HasPrefix(lower, "cypress/results/") && strings.HasSuffix(lower, ".json"):
		findings = append(findings, findingWithNotes(CategoryAssurance, "cypress", rel, "medium", "Cypress result discovery is available; use JUnit or test-summary ingestion for normalization."))
	}

	switch {
	case strings.HasSuffix(lower, ".sarif") && (strings.HasPrefix(lower, "reports/") || strings.HasPrefix(lower, "results/") || lower == "codeql-results.sarif" || lower == "semgrep.sarif" || lower == "trivy.sarif"):
		findings = append(findings, finding(CategorySecurity, "sarif", rel, "high"))
	case lower == "npm-audit.json":
		findings = append(findings, findingWithNotes(CategorySecurity, "npm-audit", rel, "medium", "npm audit discovery is available; SARIF ingestion is supported in this release."))
	case lower == "osv-scanner.json":
		findings = append(findings, findingWithNotes(CategorySecurity, "osv-scanner", rel, "medium", "OSV scanner discovery is available; SARIF ingestion is supported in this release."))
	}

	switch {
	case strings.HasPrefix(lower, ".github/workflows/") && (strings.HasSuffix(lower, ".yml") || strings.HasSuffix(lower, ".yaml")):
		findings = append(findings, finding(CategoryExecution, "github-actions", rel, "high"))
	case lower == "reports/telemetry.json" || lower == "bottleneck/execution/telemetry.json":
		findings = append(findings, finding(CategoryExecution, "telemetry", rel, "high"))
	case lower == "deployment-summary.json":
		findings = append(findings, findingWithNotes(CategoryExecution, "deployment-summary", rel, "medium", "Deployment summary discovery is available; telemetry JSON ingestion is supported in this release."))
	case lower == "release-summary.json":
		findings = append(findings, findingWithNotes(CategoryExecution, "release-summary", rel, "medium", "Release summary discovery is available; telemetry JSON ingestion is supported in this release."))
	}

	switch {
	case lower == "readme.md":
		findings = append(findings, finding(CategoryDesign, "readme", rel, "medium"))
	case lower == "docs/architecture.md":
		findings = append(findings, finding(CategoryDesign, "architecture", rel, "high"))
	case strings.HasPrefix(lower, "docs/adr/") && strings.HasSuffix(lower, ".md"):
		findings = append(findings, finding(CategoryDesign, "adr", rel, "high"))
	case lower == "openapi.yaml" || lower == "openapi.yml" || lower == "swagger.yaml" || lower == "swagger.yml":
		findings = append(findings, finding(CategoryDesign, "openapi", rel, "medium"))
	case strings.HasPrefix(lower, "docs/") && strings.HasSuffix(lower, ".md"):
		findings = append(findings, finding(CategoryDesign, "docs", rel, "medium"))
	}

	switch {
	case lower == "bottleneck/config.yaml":
		findings = append(findings, finding(CategoryNative, "config", rel, "high"))
	case lower == "bottleneck/assurance/results.json":
		findings = append(findings, finding(CategoryNative, "bottleneck-assurance", rel, "high"))
	case lower == "bottleneck/security/guardrails.json":
		findings = append(findings, finding(CategoryNative, "bottleneck-security", rel, "high"))
	case lower == "bottleneck/execution/telemetry.json":
		findings = append(findings, finding(CategoryNative, "bottleneck-execution", rel, "high"))
	case strings.HasPrefix(lower, "bottleneck/intent/") || strings.HasPrefix(lower, "bottleneck/behavior/") || strings.HasPrefix(lower, "bottleneck/design/"):
		findings = append(findings, finding(CategoryNative, "bottleneck-doc", rel, "high"))
	}

	return findings
}

func finding(category, kind, path, confidence string) DiscoveryFinding {
	return findingWithNotes(category, kind, path, confidence)
}

func findingWithNotes(category, kind, path, confidence string, notes ...string) DiscoveryFinding {
	return DiscoveryFinding{
		Category:         category,
		Kind:             kind,
		Path:             path,
		Confidence:       confidence,
		SuggestedCommand: suggestedCommand(kind, path),
		Notes:            notes,
	}
}

func suggestedCommand(kind, path string) string {
	switch kind {
	case "cucumber":
		return fmt.Sprintf("bottleneck ingest cucumber --file %s --merge", path)
	case "junit":
		return fmt.Sprintf("bottleneck ingest junit --file %s --merge", path)
	case "coverage":
		return fmt.Sprintf("bottleneck ingest coverage --file %s --merge", path)
	case "sarif":
		return fmt.Sprintf("bottleneck ingest sarif --file %s --merge", path)
	case "test-summary":
		return fmt.Sprintf("bottleneck ingest test-summary --file %s --merge", path)
	case "telemetry":
		return fmt.Sprintf("bottleneck ingest telemetry --file %s --merge", path)
	}
	return ""
}

func pathMatches(path string, prefix string, suffix string) bool {
	return strings.HasPrefix(path, prefix) && strings.HasSuffix(path, suffix)
}

func missingCategories(counts map[string]int, kindCounts map[string]int) []string {
	var missing []string
	for _, category := range []string{CategoryAssurance, CategorySecurity, CategoryExecution, CategoryDesign, CategoryNative} {
		if counts[category] == 0 {
			missing = append(missing, category+" evidence")
		}
	}
	if kindCounts["telemetry"] == 0 && kindCounts["bottleneck-execution"] == 0 {
		missing = append(missing, "execution telemetry")
	}
	return missing
}

func findingsForCategory(findings []DiscoveryFinding, category string) []DiscoveryFinding {
	var filtered []DiscoveryFinding
	for _, finding := range findings {
		if finding.Category == category {
			filtered = append(filtered, finding)
		}
	}
	return filtered
}

func categoryRank(category string) int {
	switch category {
	case CategoryAssurance:
		return 0
	case CategorySecurity:
		return 1
	case CategoryExecution:
		return 2
	case CategoryDesign:
		return 3
	case CategoryNative:
		return 4
	default:
		return 99
	}
}

func titleCategory(category string) string {
	if category == "" {
		return ""
	}
	return strings.ToUpper(category[:1]) + category[1:]
}

func displayPath(rootPath string, path string) string {
	if rel, err := filepath.Rel(rootPath, path); err == nil && rel != "." && !strings.HasPrefix(rel, "..") {
		return filepath.ToSlash(rel)
	}
	if _, err := os.Stat(path); err == nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(path)
}

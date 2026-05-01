package discover

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScanFindsEvidenceAndSuggestedCommands(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "reports/cucumber.json", "[]")
	writeFile(t, root, "reports/codeql.sarif", "{}")
	writeFile(t, root, "reports/telemetry.json", "{}")
	writeFile(t, root, "README.md", "# App\n")
	writeFile(t, root, ".github/workflows/ci.yml", "name: ci\n")
	writeFile(t, root, "node_modules/pkg/reports/cucumber.json", "[]")

	result, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}

	assertFinding(t, result, "cucumber", "reports/cucumber.json", "bottleneck ingest cucumber --file reports/cucumber.json --merge")
	assertFinding(t, result, "sarif", "reports/codeql.sarif", "bottleneck ingest sarif --file reports/codeql.sarif --merge")
	assertFinding(t, result, "telemetry", "reports/telemetry.json", "bottleneck ingest telemetry --file reports/telemetry.json --merge")
	assertFinding(t, result, "readme", "README.md", "")
	assertFinding(t, result, "github-actions", ".github/workflows/ci.yml", "")

	if result.Summary.CountsByCategory[CategoryAssurance] != 1 ||
		result.Summary.CountsByCategory[CategorySecurity] != 1 ||
		result.Summary.CountsByCategory[CategoryExecution] != 2 ||
		result.Summary.CountsByCategory[CategoryDesign] != 1 {
		t.Fatalf("unexpected discovery counts: %#v", result.Summary.CountsByCategory)
	}
	for _, finding := range result.Findings {
		if strings.Contains(finding.Path, "node_modules") {
			t.Fatalf("expected ignored directory to be skipped, got finding %#v", finding)
		}
	}
}

func TestScanReportsMissingTelemetry(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "reports/cucumber.json", "[]")

	result, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}

	if !contains(result.Summary.Missing, "execution telemetry") {
		t.Fatalf("expected missing telemetry summary, got %#v", result.Summary.Missing)
	}
}

func TestMarshalJSONIncludesStableFields(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "coverage/lcov.info", "TN:\nSF:app.go\nDA:1,1\nend_of_record\n")

	result, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	encoded, err := MarshalJSON(result)
	if err != nil {
		t.Fatalf("MarshalJSON returned error: %v", err)
	}

	var decoded DiscoveryResult
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("discovery JSON did not parse: %v\n%s", err, string(encoded))
	}
	if decoded.RootPath == "" || decoded.Summary.TotalFindings != 1 || decoded.Findings[0].Kind != "coverage" {
		t.Fatalf("unexpected discovery JSON: %#v", decoded)
	}
}

func assertFinding(t *testing.T, result DiscoveryResult, kind string, path string, command string) {
	t.Helper()
	for _, finding := range result.Findings {
		if finding.Kind == kind && finding.Path == path {
			if command != "" && finding.SuggestedCommand != command {
				t.Fatalf("expected command %q for %s, got %q", command, path, finding.SuggestedCommand)
			}
			return
		}
	}
	t.Fatalf("expected finding kind=%s path=%s in %#v", kind, path, result.Findings)
}

func writeFile(t *testing.T, root string, rel string, content string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", rel, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
}

func contains(items []string, wanted string) bool {
	for _, item := range items {
		if item == wanted {
			return true
		}
	}
	return false
}

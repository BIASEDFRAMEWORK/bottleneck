package validator

import "testing"

func TestJSONEvidenceQualityScoresIDsRefsAndContext(t *testing.T) {
	weak := evaluateJSONEvidenceQuality("bottleneck", "assurance/results.json", "Assurance", []byte(`{
  "evidence": [{"id": "SECURITY-001"}]
}`))

	expectedDetails := []string{
		"bottleneck/assurance/results.json does not include ASSURANCE-* evidence IDs",
		"bottleneck/assurance/results.json evidence IDs do not reference related release evidence",
		"bottleneck/assurance/results.json evidence entries do not include source or status context",
	}
	for _, detail := range expectedDetails {
		if !containsString(weak.Details, detail) {
			t.Fatalf("expected detail %q in weak JSON quality %#v", detail, weak)
		}
	}
	if weak.Score >= 100 {
		t.Fatalf("expected weak JSON evidence score below 100, got %d", weak.Score)
	}

	strong := evaluateJSONEvidenceQuality("bottleneck", "assurance/results.json", "Assurance", []byte(`{
  "evidence": [{"id": "ASSURANCE-001", "refs": ["BEHAVIOR-001"], "source": "cucumber", "status": "pass"}]
}`))
	if strong.Score != 100 {
		t.Fatalf("expected strong JSON evidence to score 100, got %d with details %#v", strong.Score, strong.Details)
	}
}

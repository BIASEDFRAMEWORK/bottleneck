# AI Implementation Prompt: Generate Leadership-Ready SDLC Evidence Report

You are working in the Bottleneck Go CLI codebase.

Implement **Implementation Epic 4: Generate leadership-ready SDLC evidence report** from the `Enterprise SDLC Evidence Package` milestone.

## Milestone Context

Bottleneck should help a development team generate a leadership-ready SDLC evidence report from local evidence, local snapshots, and deterministic analysis.

This epic adds:

```sh
bottleneck report
```

The report should summarize current state, trends over time, primary bottleneck, evidence found, evidence missing, risks, recommended actions, owners, automation, and leadership decisions.

No database, SaaS backend, external storage, dashboard, or project-management dependency may be introduced.

## Scope

Add a new command:

```sh
bottleneck report
```

Supported flags:

```sh
bottleneck report --env=default
bottleneck report --format=markdown
bottleneck report --format=json
bottleneck report --out=bottleneck/reports/sdlc-evidence-report.md
bottleneck report --window=6
bottleneck report --strict
```

Default behavior:

1. Run current validation and scorecard logic.
2. Load recent trend analysis if snapshot history exists.
3. Generate explanation for the primary bottleneck and weak categories.
4. Render a Markdown report.
5. Write to `bottleneck/reports/sdlc-evidence-report.md`.
6. Print the output path and short summary.

## Current Code To Inspect

Read before changing code:

- `cmd/report.go`, if any
- `cmd/scorecard.go`
- `cmd/explain.go`
- `cmd/diagnose.go`
- `cmd/root.go`
- `internal/scorecard/*`
- `internal/trends/*`, if implemented
- `internal/diagnosis/*`
- `internal/report/*`, if present
- `internal/validator/*`
- existing renderer and output helpers
- docs and README references

If trends or enhanced explain are not implemented yet, handle those dependencies gracefully and keep the report useful with current scorecard and validation data.

## Report Sections

Markdown report must include:

```markdown
# SDLC Evidence Report

## Executive Summary
## Current Delivery-System Status
## Primary Bottleneck
## Category Scorecard
## Trend Summary
## Evidence Found
## Evidence Missing
## Risk to Delivery
## Recommended Actions
## Suggested Owners
## Suggested Automation
## Leadership Decision Needed
## Appendix: Snapshot Metadata
```

If no snapshots exist, `Trend Summary` should say there is insufficient history and recommend creating snapshots over multiple delivery cycles.

## Recommended Tone

Use enterprise-friendly, evidence-based language.

Avoid:

```text
Your team failed.
Your Scrum is broken.
This is bad.
```

Prefer:

```text
The current evidence suggests...
The primary constraint appears to be...
The team has evidence for...
The team is missing evidence for...
Leadership may need to decide whether to...
```

## Suggested Internal Package

Add if appropriate:

```text
internal/report/
  report.go
  render.go
  report_test.go
  render_test.go
```

Suggested model:

```go
type Report struct {
    GeneratedAt             time.Time
    Environment             string
    CurrentStatus           string
    ReleaseRecommendation   string
    PrimaryBottleneck       string
    Scorecard               any
    TrendSummary            *trends.Analysis
    Explanation             any
    EvidenceFound           []string
    EvidenceMissing         []string
    Risks                   []string
    RecommendedActions      []string
    SuggestedOwners         []string
    SuggestedAutomation     []string
    LeadershipDecision      string
}
```

Use concrete DTOs where they exist.

## Leadership Decision Generation

Generate deterministic leadership decision text.

Examples:

- Assurance:
  - `Approve time to map critical behaviors to validation evidence before accelerating release.`
- Security:
  - `Decide whether release should be blocked until high-severity findings are resolved or formally accepted.`
- Intent:
  - `Align leadership, product, domain, and engineering on measurable intent before expanding implementation.`
- Execution:
  - `Prioritize production telemetry, adoption feedback, and operational readiness before scaling usage.`
- No primary bottleneck:
  - `Continue monitoring evidence trends and maintain current release controls.`

## Example Terminal Output

```text
SDLC evidence report created

Status: WARN
Primary bottleneck: Assurance
Trend: Improving, but Assurance remains persistent
Report: bottleneck/reports/sdlc-evidence-report.md
```

## JSON Output

`--format=json` should render a structured report object with stable snake_case fields and all major sections represented.

If `--format=json` and `--out` are both provided, write JSON to the output path.

## Acceptance Criteria

- `bottleneck report` creates `bottleneck/reports/sdlc-evidence-report.md`.
- Report includes executive summary.
- Report includes current status.
- Report includes primary bottleneck.
- Report includes category scorecard.
- Report includes trend summary if snapshots exist.
- Report handles missing trend history gracefully.
- Report includes evidence found.
- Report includes evidence missing.
- Report includes risk to delivery.
- Report includes recommended actions.
- Report includes suggested owners.
- Report includes suggested automation.
- Report includes leadership decision needed.
- Report supports JSON output.
- Report supports custom output path.
- Report does not require a database or external service.
- Existing commands continue to work.

## Tests To Add

Add tests under `internal/report` and command tests where useful:

- `TestReportCreatesDefaultMarkdownFile`
- `TestReportIncludesExecutiveSummary`
- `TestReportIncludesCurrentStatus`
- `TestReportIncludesPrimaryBottleneck`
- `TestReportIncludesCategoryScorecard`
- `TestReportIncludesTrendSummaryWhenSnapshotsExist`
- `TestReportHandlesMissingTrendHistory`
- `TestReportIncludesEvidenceFound`
- `TestReportIncludesEvidenceMissing`
- `TestReportIncludesRisks`
- `TestReportIncludesRecommendations`
- `TestReportIncludesSuggestedOwners`
- `TestReportIncludesSuggestedAutomation`
- `TestReportIncludesLeadershipDecision`
- `TestReportSupportsJSON`
- `TestReportSupportsCustomOutputPath`

Use deterministic fixtures. Avoid full Markdown golden files unless that is the local testing style.

## Implementation Constraints

- Do not add a database or external storage.
- Do not call an LLM.
- Do not require trend history to exist.
- Do not fail solely because snapshots are missing.
- Do not break scorecard, explain, diagnose, validate, trace, or ingest.
- Keep report language deterministic and testable.

## Verification

Run:

```sh
go test ./...
```

If feasible, manually verify:

```sh
bottleneck seed-history
bottleneck report
bottleneck report --format=json
bottleneck report --out=bottleneck/reports/custom-sdlc-report.md
```

Use a temporary directory for manual report generation.

## Final Response Requirements

When finished, report:

1. Report command behavior.
2. Report sections implemented.
3. Trend and missing-history behavior.
4. JSON and Markdown support.
5. Tests added or changed.
6. Exact commands run and results.
7. Any acceptance criteria intentionally deferred and why.


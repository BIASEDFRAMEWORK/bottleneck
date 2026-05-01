Below is an implementation list for the milestone:

# Milestone: Enterprise SDLC Evidence Package

## Milestone intent

Enable a development team to use Bottleneck as an **enterprise SDLC evidence system** that helps them explain to leadership:

```text
Where are we today?
Where have we been?
Are we improving or getting worse?
What is the primary SDLC bottleneck?
What evidence supports that conclusion?
What should we address next?
What decision or support do we need from leadership?
```

The current codebase already has the right starting point: a Cobra-based CLI with commands such as `scorecard`, `diagnose`, `trace`, `ingest`, `validate`, and `explain`, plus evidence categories for intent, behavior, design, assurance, security, and execution. 

The milestone should add three connected capabilities:

```text
1. Snapshot evidence over time.
2. Trend evidence over time.
3. Generate leadership-ready SDLC evidence reports.
```

This should be implemented without adding:

```text
- A database
- SaaS backend
- External storage
- Dashboard
- Jira dependency
- Third-party metrics platform
```

Git and local files remain the system of record.

---

# Guiding architecture principle

## Bottleneck does not store metrics in a database.

Instead:

```text
Bottleneck materializes evidence snapshots.
Git versions those snapshots.
Bottleneck reads those snapshots to explain trends.
Reports translate the evidence into enterprise decision language.
```

This keeps Bottleneck aligned with its positioning:

```text
Metrics are derived from evidence, tools, artifacts, and Git history — not Jira.
```

---

# Target enterprise workflow

The intended user flow should become:

```sh
bottleneck init --template saas

# Team adds or ingests evidence:
bottleneck ingest --type cucumber --file reports/cucumber.json
bottleneck ingest --type sarif --file reports/codeql.sarif
bottleneck ingest --type telemetry --file reports/telemetry.json

# Team checks current state:
bottleneck scorecard

# Team creates an auditable snapshot:
bottleneck snapshot

# Team sees whether SDLC evidence is improving:
bottleneck trends

# Team generates leadership-facing report:
bottleneck report
```

Optional release flow:

```sh
bottleneck snapshot --env=production --label=release-candidate
bottleneck trends --env=production --window=6
bottleneck report --env=production --format=markdown
```

Optional CI flow:

```sh
bottleneck scorecard --env=production --format=json
bottleneck snapshot --env=production --label=ci
bottleneck report --env=production --format=markdown
```

---

# Proposed command set

Add or enhance these commands:

```text
bottleneck snapshot
bottleneck trends
bottleneck report
```

Enhance:

```text
bottleneck explain
bottleneck scorecard --format=json
```

Do not remove or break:

```text
bottleneck init
bottleneck validate
bottleneck scorecard
bottleneck diagnose
bottleneck trace
bottleneck ingest
bottleneck explain
```

---

# Proposed folder structure

Add this generated artifact structure:

```text
bottleneck/
  history/
    scorecards/
      2026-05-01T141500Z-default-scorecard.json
      2026-05-08T141500Z-default-scorecard.json
      2026-05-15T141500Z-production-scorecard.json

    latest/
      default.json
      production.json

  reports/
    trend-summary.md
    bottleneck-explanation.md
    sdlc-evidence-report.md
```

Optional future folder:

```text
bottleneck/
  history/
    decisions/
      2026-05-01T141500Z-release-decision.json
```

Do not require this future folder in the current milestone unless it is already natural to add.

---

# Implementation Epic 1: Snapshot current scorecard evidence

## Feature name

```text
bottleneck snapshot
```

## Purpose

Create an immutable, timestamped scorecard snapshot that can be committed to Git and used later for trend analysis.

## User story

```text
As a development team,
I want to create a timestamped snapshot of Bottleneck’s current SDLC evidence,
so that I can compare delivery-system maturity over time without using a database.
```

## CLI command

```sh
bottleneck snapshot
```

## CLI options

```sh
bottleneck snapshot --env=default
bottleneck snapshot --env=production
bottleneck snapshot --label=release-candidate
bottleneck snapshot --out=bottleneck/history/scorecards
bottleneck snapshot --format=json
bottleneck snapshot --commit
bottleneck snapshot --no-latest
bottleneck snapshot --strict
```

## Recommended first version options

Implement these in the first pass:

```text
--env
--label
--out
--strict
--no-latest
```

Defer this unless easy:

```text
--commit
```

Auto-committing can be useful, but it introduces Git safety questions. It should not block the milestone.

## Default behavior

When the user runs:

```sh
bottleneck snapshot
```

Bottleneck should:

```text
1. Run the existing validation and scorecard logic.
2. Generate a JSON scorecard snapshot.
3. Add snapshot metadata.
4. Write a timestamped file to bottleneck/history/scorecards/.
5. Write or update bottleneck/history/latest/default.json.
6. Print a concise success message.
7. Exit non-zero if the system status is FAIL, unless existing behavior suggests snapshots should still be allowed.
```

Important design decision:

I recommend snapshot creation should still happen even when status is `FAIL`.

Reason:

```text
A failed scorecard is still valuable historical evidence.
```

But the command should exit with the same failure semantics as `scorecard` only if explicitly requested later.

For first implementation, I would make `snapshot` always write the file and exit zero unless there is an actual runtime error.

Potential future flag:

```sh
bottleneck snapshot --fail-on-scorecard-fail
```

## Example terminal output

```text
Bottleneck snapshot created

Environment: default
Status: WARN
Primary bottleneck: Assurance
Snapshot: bottleneck/history/scorecards/2026-05-01T141500Z-default-scorecard.json
Latest: bottleneck/history/latest/default.json

Next:
Commit this snapshot so Bottleneck can compare SDLC evidence over time.
```

## Snapshot filename convention

Use UTC timestamp.

```text
YYYY-MM-DDTHHMMSSZ-{env}-scorecard.json
```

Example:

```text
2026-05-01T141500Z-default-scorecard.json
2026-05-01T141500Z-production-scorecard.json
```

If a label is provided:

```text
YYYY-MM-DDTHHMMSSZ-{env}-{label}-scorecard.json
```

Example:

```text
2026-05-01T141500Z-production-release-candidate-scorecard.json
```

Sanitize labels:

```text
release candidate -> release-candidate
Release_Candidate -> release-candidate
prod/rc1 -> prod-rc1
```

## Snapshot metadata schema

Add a top-level snapshot block to JSON output.

```json
{
  "schema_version": "scorecard.snapshot.v1",
  "snapshot": {
    "id": "SNAPSHOT-20260501-141500",
    "created_at": "2026-05-01T14:15:00Z",
    "environment": "default",
    "label": "release-candidate",
    "source": "bottleneck snapshot",
    "git": {
      "commit": "abc1234",
      "branch": "main",
      "dirty": false
    }
  },
  "scorecard": {
    "system_status": "WARN",
    "release_recommendation": "Conditional",
    "primary_bottleneck": "Assurance",
    "categories": []
  }
}
```

If Git metadata cannot be detected:

```json
"git": {
  "commit": "",
  "branch": "",
  "dirty": null
}
```

Do not fail snapshot generation just because Git metadata cannot be detected.

## Git metadata helper

Add an internal package if one does not already exist:

```text
internal/gitinfo/
  gitinfo.go
  gitinfo_test.go
```

Functions:

```go
type Info struct {
    Commit string `json:"commit"`
    Branch string `json:"branch"`
    Dirty  *bool  `json:"dirty,omitempty"`
}

func Detect(root string) Info
```

Implementation can shell out to:

```sh
git rev-parse --short HEAD
git rev-parse --abbrev-ref HEAD
git status --porcelain
```

Rules:

```text
- If not in a Git repo, return empty values.
- Do not fail the command.
- Keep timeouts short if command execution uses context.
- Tests should not require a real remote repository.
```

## Internal package

Add:

```text
internal/snapshot/
  snapshot.go
  snapshot_test.go
```

Suggested types:

```go
type Metadata struct {
    ID          string      `json:"id"`
    CreatedAt   time.Time   `json:"created_at"`
    Environment string      `json:"environment"`
    Label       string      `json:"label,omitempty"`
    Source      string      `json:"source"`
    Git         gitinfo.Info `json:"git"`
}

type Snapshot struct {
    SchemaVersion string      `json:"schema_version"`
    Snapshot      Metadata    `json:"snapshot"`
    Scorecard     interface{} `json:"scorecard"`
}
```

Better than `interface{}` if existing scorecard model has a concrete type.

## Acceptance criteria

```text
AC1. Running `bottleneck snapshot` creates `bottleneck/history/scorecards/`.
AC2. Running `bottleneck snapshot` writes a timestamped JSON snapshot file.
AC3. Running `bottleneck snapshot` updates `bottleneck/history/latest/default.json` unless `--no-latest` is used.
AC4. Running `bottleneck snapshot --env=production` writes a production-specific snapshot.
AC5. Running `bottleneck snapshot --label=release-candidate` includes the label in metadata and filename.
AC6. Snapshot JSON includes schema_version, snapshot metadata, git metadata, and scorecard data.
AC7. Snapshot creation does not require a database or external service.
AC8. Snapshot creation does not break existing scorecard behavior.
AC9. Snapshot command has tests for filename generation, metadata generation, latest-file behavior, and label sanitization.
AC10. Snapshot command is included in root help text.
```

## Test cases

```text
TestSnapshotCreatesHistoryDirectory
TestSnapshotWritesTimestampedFile
TestSnapshotWritesLatestFile
TestSnapshotNoLatestFlagSkipsLatestFile
TestSnapshotIncludesEnvironment
TestSnapshotIncludesLabel
TestSnapshotSanitizesLabel
TestSnapshotIncludesGitMetadataWhenAvailable
TestSnapshotDoesNotFailOutsideGitRepo
TestSnapshotDoesNotBreakExistingScorecardCommand
```

## Codex prompt packet: Snapshot feature

```text
You are working in the Bottleneck Go CLI codebase.

Goal:
Implement a new `bottleneck snapshot` command that creates timestamped, Git-versionable scorecard snapshots without using a database or external storage.

Context:
The existing CLI already has commands such as scorecard, validate, diagnose, trace, ingest, and explain. Reuse the existing validator and scorecard rendering/model logic. Do not change or break existing command behavior.

Requirements:
1. Add a new Cobra command: `snapshot`.
2. Default command: `bottleneck snapshot`.
3. Supported flags:
   - `--env`, default `default`
   - `--label`, optional
   - `--out`, default `bottleneck/history/scorecards`
   - `--strict`, default false
   - `--no-latest`, default false
4. The command should run the same validation/scorecard logic used by `bottleneck scorecard`.
5. It should create a JSON snapshot file under `bottleneck/history/scorecards/`.
6. Filename format:
   - `YYYY-MM-DDTHHMMSSZ-{env}-scorecard.json`
   - If label is provided: `YYYY-MM-DDTHHMMSSZ-{env}-{label}-scorecard.json`
7. The command should also update `bottleneck/history/latest/{env}.json` unless `--no-latest` is passed.
8. Snapshot JSON should include:
   - `schema_version`
   - `snapshot.id`
   - `snapshot.created_at`
   - `snapshot.environment`
   - `snapshot.label`
   - `snapshot.source`
   - `snapshot.git.commit`
   - `snapshot.git.branch`
   - `snapshot.git.dirty`
   - `scorecard`
9. Add internal package `internal/snapshot` if appropriate.
10. Add internal package `internal/gitinfo` if appropriate.
11. Git metadata detection must not fail the command if Git is unavailable.
12. Add tests for snapshot file creation, latest file creation, label sanitization, environment handling, and no Git repo behavior.
13. Update root command help text to include `snapshot`.
14. Do not add any database, SaaS, or external storage dependency.
15. Run `go test ./...` and fix failures.

Acceptance:
- Existing tests continue passing.
- New tests cover the snapshot behavior.
- `bottleneck snapshot` creates valid JSON history artifacts.
```

---

# Implementation Epic 2: Trend analysis from local snapshots

## Feature name

```text
bottleneck trends
```

## Purpose

Compare historical Bottleneck snapshots to show whether the team’s SDLC evidence is improving, declining, stable, or regressing.

## User story

```text
As a development team,
I want Bottleneck to compare prior evidence snapshots,
so that I can tell leadership whether our SDLC is getting better or worse.
```

## CLI command

```sh
bottleneck trends
```

## CLI options

```sh
bottleneck trends --env=default
bottleneck trends --window=6
bottleneck trends --since=30d
bottleneck trends --format=text
bottleneck trends --format=markdown
bottleneck trends --format=json
bottleneck trends --out=bottleneck/reports/trend-summary.md
```

## Recommended first version options

Implement:

```text
--env
--window
--format
--out
```

Defer if needed:

```text
--since
```

Time-window parsing can be added later if it slows implementation.

## Default behavior

When the user runs:

```sh
bottleneck trends
```

Bottleneck should:

```text
1. Read snapshots from bottleneck/history/scorecards/.
2. Filter to environment `default`.
3. Sort by snapshot.created_at ascending.
4. Use the latest 6 snapshots by default.
5. Compare category scores/statuses.
6. Identify improving, declining, stable, insufficient-history, and recurring bottlenecks.
7. Render a text trend summary to stdout.
```

## Example terminal output

```text
Bottleneck Trends

Environment: default
Snapshots analyzed: 6
Window: latest 6 snapshots

Overall direction: Improving
Current status: WARN
Current primary bottleneck: Assurance

Category trends:
- Intent:     70 → 80 → 85 → 90  Improving
- Behavior:  55 → 60 → 70 → 75  Improving
- Design:    80 → 80 → 80 → 80  Stable
- Assurance: 30 → 35 → 40 → 45  Improving, but still weak
- Security:  90 → 90 → 60 → 85  Recovered
- Execution: 75 → 70 → 65 → 60  Declining

Persistent bottleneck:
Assurance appeared as the primary bottleneck in 4 of 6 snapshots.

Leadership summary:
The team is improving overall, but assurance remains the most persistent delivery constraint.
```

## Trend model

Create an internal package:

```text
internal/trends/
  trends.go
  trends_test.go
  render.go
  render_test.go
```

## Suggested trend types

```go
type Direction string

const (
    DirectionImproving Direction = "improving"
    DirectionDeclining Direction = "declining"
    DirectionStable Direction = "stable"
    DirectionRecovered Direction = "recovered"
    DirectionRegressed Direction = "regressed"
    DirectionInsufficientHistory Direction = "insufficient_history"
)
```

## Suggested structs

```go
type Analysis struct {
    Environment              string
    SnapshotCount            int
    Window                   int
    OverallDirection          Direction
    CurrentStatus             string
    CurrentPrimaryBottleneck  string
    CategoryTrends            []CategoryTrend
    PersistentBottleneck      PersistentBottleneck
    LeadershipSummary         string
}

type CategoryTrend struct {
    Category       string
    Values         []float64
    Statuses       []string
    Direction      Direction
    Delta          float64
    CurrentValue   float64
    PreviousValue  float64
    Summary        string
}

type PersistentBottleneck struct {
    Category string
    Count    int
    Total    int
    Summary  string
}
```

## Category score extraction

Use whatever score model currently exists. If the scorecard does not have numeric category scores, use a normalized status model:

```text
PASS = 100
WARN = 60
FAIL = 0
UNKNOWN = null
```

If numeric category scores already exist, use those.

## Direction calculation rules

First implementation can be simple and deterministic.

Given a category’s first and last score in the selected window:

```text
delta = last - first
```

Rules:

```text
If fewer than 2 snapshots: insufficient_history
If delta >= +10: improving
If delta <= -10: declining
If absolute delta < 10: stable
```

Add special classifications:

```text
Recovered:
- Category had FAIL/WARN in earlier snapshot and PASS in latest snapshot.

Regressed:
- Category had PASS earlier and latest is WARN/FAIL.
```

Precedence:

```text
1. insufficient_history
2. regressed
3. recovered
4. improving
5. declining
6. stable
```

## Overall direction rules

Simple version:

```text
If more categories improving/recovered than declining/regressed: improving
If more categories declining/regressed than improving/recovered: declining
If all or most stable: stable
If insufficient snapshots: insufficient_history
```

## Persistent bottleneck rules

Use each snapshot’s `primary_bottleneck`.

```text
Count how often each category appears.
The category with the highest count is the persistent bottleneck.
```

Ignore empty values:

```text
None
No bottleneck
""
```

Example:

```text
Assurance appeared in 4 of 6 snapshots.
```

## Leadership summary generation

Deterministic text based on analysis.

Examples:

```text
Improving overall + persistent bottleneck:
"The team is improving overall, but Assurance remains the most persistent delivery constraint."

Declining:
"The team’s SDLC evidence is declining. Leadership should review the categories with recent regressions before approving additional release acceleration."

Stable but weak:
"The team is stable but not improving. The current bottleneck has persisted across multiple snapshots and likely requires deliberate investment."

Insufficient history:
"Not enough historical snapshots exist yet. Create snapshots over multiple delivery cycles to establish a trend."
```

## Markdown output

```sh
bottleneck trends --format=markdown --out=bottleneck/reports/trend-summary.md
```

Should write:

```markdown
# Bottleneck Trend Summary

## Executive Summary

## Snapshot Window

## Category Trends

## Persistent Bottleneck

## Leadership Interpretation

## Recommended Follow-Up
```

## JSON output

```sh
bottleneck trends --format=json
```

Should output machine-readable trend analysis.

## Acceptance criteria

```text
AC1. Running `bottleneck trends` reads local scorecard snapshots.
AC2. Trends do not require a database or external service.
AC3. Trends filter by environment.
AC4. Trends sort snapshots by created_at.
AC5. Trends compare the latest N snapshots using `--window`.
AC6. Trends identify category direction.
AC7. Trends identify persistent primary bottleneck.
AC8. Trends generate text output by default.
AC9. Trends support markdown output.
AC10. Trends support JSON output.
AC11. Trends can write markdown output to `bottleneck/reports/trend-summary.md`.
AC12. Trends handle insufficient history gracefully.
AC13. Trends ignore malformed snapshots with a warning or return a clear error.
AC14. Existing commands and tests remain unchanged.
```

## Test cases

```text
TestTrendsInsufficientHistory
TestTrendsReadsSnapshots
TestTrendsSortsSnapshotsByCreatedAt
TestTrendsFiltersByEnvironment
TestTrendsUsesLatestWindow
TestTrendsDetectsImprovingCategory
TestTrendsDetectsDecliningCategory
TestTrendsDetectsStableCategory
TestTrendsDetectsRecoveredCategory
TestTrendsDetectsRegressedCategory
TestTrendsIdentifiesPersistentBottleneck
TestTrendsRendersText
TestTrendsRendersMarkdown
TestTrendsRendersJSON
TestTrendsWritesMarkdownReport
TestTrendsHandlesMissingHistoryDirectory
```

## Codex prompt packet: Trends feature

```text
You are working in the Bottleneck Go CLI codebase.

Goal:
Implement `bottleneck trends`, which analyzes local Bottleneck scorecard snapshots over time and reports whether SDLC evidence is improving, declining, stable, recovered, or regressed.

Context:
A previous feature should have introduced `bottleneck snapshot`, which writes JSON snapshots to `bottleneck/history/scorecards/`. Do not use a database, external storage, SaaS backend, or dashboard.

Requirements:
1. Add a new Cobra command: `trends`.
2. Supported flags:
   - `--env`, default `default`
   - `--window`, default 6
   - `--format`, default `text`, allowed values `text`, `markdown`, `json`
   - `--out`, optional path for writing rendered output
3. Read snapshot JSON files from `bottleneck/history/scorecards/`.
4. Filter snapshots by environment.
5. Sort snapshots by `snapshot.created_at` ascending.
6. Analyze the latest N snapshots based on `--window`.
7. Compare category scores or normalized category statuses over time.
8. Direction rules:
   - fewer than 2 snapshots = insufficient_history
   - delta >= +10 = improving
   - delta <= -10 = declining
   - absolute delta < 10 = stable
   - earlier WARN/FAIL and latest PASS = recovered
   - earlier PASS and latest WARN/FAIL = regressed
9. Identify persistent bottleneck by counting `primary_bottleneck` across the selected window.
10. Generate an overall direction.
11. Generate deterministic leadership summary text.
12. Support text, markdown, and JSON output.
13. If `--out` is provided, write the rendered output to that file, creating parent directories as needed.
14. Add internal package `internal/trends`.
15. Add tests for reading, sorting, filtering, windowing, direction detection, persistent bottleneck detection, and rendering.
16. Update root help text to include `trends`.
17. Run `go test ./...` and fix failures.

Acceptance:
- `bottleneck trends` works with snapshot files only.
- No database or external storage is added.
- Existing commands continue to work.
```

---

# Implementation Epic 3: Enhance explain into evidence-backed diagnosis

## Feature name

```text
bottleneck explain
```

## Purpose

Make `explain` more useful for enterprise teams by clearly describing:

```text
- Why the category scored the way it did
- What evidence was found
- What evidence is missing
- Why leadership should care
- What the team should do next
```

The current command already exists, so this should be an enhancement, not a replacement.

## User story

```text
As a development team,
I want Bottleneck to explain the evidence behind each weak SDLC category,
so that I can defend recommendations to leadership using facts instead of opinions.
```

## CLI options to add or confirm

```sh
bottleneck explain
bottleneck explain --category=assurance
bottleneck explain --category=security
bottleneck explain --capability=billing
bottleneck explain --format=text
bottleneck explain --format=markdown
bottleneck explain --format=json
bottleneck explain --out=bottleneck/reports/bottleneck-explanation.md
```

Current implementation appears to support capability. Add category and format/out if missing.

## Output model

For each category, explain:

```text
Category
Status
Score
Why this matters
Evidence found
Evidence missing
Risk to delivery
Recommended next actions
Suggested owner roles
Suggested GitHub Actions or hooks
```

## Example output

```text
Bottleneck Explanation

Primary bottleneck: Assurance
Status: FAIL

Why this matters:
The team can describe intended behavior, but cannot prove enough behavior before release.

Evidence found:
- behavior-spec.md exists
- Cucumber results were ingested
- Some test evidence references behavior IDs

Evidence missing:
- 5 expected behaviors have no mapped tests
- 2 tests do not reference behavior IDs
- No release validation evidence was found

Risk to delivery:
The team may ship functionality that appears complete but cannot be proven against intended behavior.

Recommended actions:
1. Add behavior IDs to Cucumber scenarios.
2. Map each critical behavior to at least one automated test.
3. Add release validation evidence before production approval.

Suggested owner roles:
- Assurance Engineer
- Developer
- Product/Domain Expert

Suggested automation:
- Run Cucumber in GitHub Actions
- Upload test output to bottleneck/assurance/results.json
- Fail release gate when critical behaviors lack mapped tests
```

## Rule-based explanation engine

Do not make this AI-powered yet.

Build deterministic rules.

### Intent rules

If intent file missing:

```text
Missing intent evidence.
Recommendation: Create bottleneck/intent/intent.md with measurable outcomes and constraints.
```

If intent exists but lacks measurable outcomes:

```text
Intent exists but does not clearly define measurable outcomes.
Recommendation: Add observable outcomes, business constraints, and unacceptable outcomes.
```

If intent has placeholders:

```text
Intent contains placeholder or thin content.
Recommendation: Replace template text with product-specific intent.
```

### Behavior rules

If behavior spec missing:

```text
Missing behavior specification.
Recommendation: Create behavior-spec.md with expected and unacceptable behaviors.
```

If behaviors lack IDs:

```text
Behavior expectations are not traceable.
Recommendation: Add stable behavior IDs such as BEHAVIOR-001.
```

If behavior has no tests:

```text
Behavior is not validated.
Recommendation: Map each critical behavior to test evidence.
```

### Design rules

If architecture doc missing:

```text
Missing architecture evidence.
Recommendation: Document major components, boundaries, dependencies, and risk decisions.
```

If design lacks tradeoffs:

```text
Architecture exists but does not explain tradeoffs.
Recommendation: Add decision records for key constraints and design choices.
```

If design lacks operational or failure-mode content:

```text
Architecture does not describe failure modes.
Recommendation: Add fallback, monitoring, and operational assumptions.
```

### Assurance rules

If Cucumber or test evidence missing:

```text
Missing automated validation evidence.
Recommendation: Add test output or BDD evidence under bottleneck/assurance/.
```

If tests exist but are not mapped:

```text
Tests exist but are not linked to behavior IDs.
Recommendation: Add traceability references from test evidence to behavior expectations.
```

If coverage is low:

```text
Critical behaviors lack validation.
Recommendation: Prioritize tests for high-risk behaviors before expanding feature scope.
```

### Security rules

If SARIF/security evidence missing:

```text
Missing security evidence.
Recommendation: Add CodeQL, dependency review, secret scanning, or SARIF evidence.
```

If high severity finding exists:

```text
High severity security findings exist.
Recommendation: Block release until findings are triaged or resolved.
```

If guardrails missing:

```text
Security guardrails are not documented.
Recommendation: Add security/guardrails.json or equivalent evidence.
```

### Execution rules

If telemetry missing:

```text
Missing execution evidence.
Recommendation: Add telemetry or production-readiness evidence.
```

If adoption/usage signals weak:

```text
Execution evidence suggests weak adoption or user trust.
Recommendation: Review user workflow, training, and feedback loops.
```

If reliability metrics weak:

```text
Execution evidence suggests operational instability.
Recommendation: Address error rate, latency, or incident signals before accelerating release.
```

## Suggested owner roles

Add deterministic owner suggestions:

```text
Intent:
- Product Lead
- Domain Expert
- Technical Lead

Behavior:
- Product Lead
- Domain Expert
- QA/Assurance Engineer

Design:
- Architect
- Technical Lead
- Platform Engineer

Assurance:
- QA/Assurance Engineer
- Developer
- Product Lead

Security:
- Security Engineer
- Platform Engineer
- Technical Lead

Execution:
- SRE/Operations
- Product Lead
- Customer Success/Adoption Lead
```

## Suggested automation mapping

```text
Intent:
- PR template requiring intent reference
- Markdown quality checks
- Commit hook for required intent IDs

Behavior:
- Behavior spec linting
- Traceability checks

Design:
- Architecture decision record check
- Diagram/doc freshness check

Assurance:
- Cucumber in GitHub Actions
- Test result ingestion
- Behavior-to-test traceability gate

Security:
- CodeQL
- Dependency review
- Secret scanning
- SARIF ingestion

Execution:
- Telemetry JSON ingestion
- Release health check
- Production signal review
```

## Acceptance criteria

```text
AC1. `bottleneck explain` continues to work with existing behavior.
AC2. `bottleneck explain --category=assurance` explains only Assurance.
AC3. Explanation includes evidence found.
AC4. Explanation includes evidence missing.
AC5. Explanation includes risk to delivery.
AC6. Explanation includes recommended actions.
AC7. Explanation includes suggested owner roles.
AC8. Explanation includes suggested automations.
AC9. Explanation supports text output.
AC10. Explanation supports markdown output.
AC11. Explanation supports JSON output.
AC12. Explanation can write to `bottleneck/reports/bottleneck-explanation.md`.
AC13. Explanation is deterministic and does not require an LLM.
AC14. Existing explain tests continue passing or are updated intentionally.
```

## Test cases

```text
TestExplainCategoryFilter
TestExplainIncludesEvidenceFound
TestExplainIncludesEvidenceMissing
TestExplainIncludesRisk
TestExplainIncludesRecommendations
TestExplainIncludesOwnerRoles
TestExplainIncludesAutomationSuggestions
TestExplainMarkdownOutput
TestExplainJSONOutput
TestExplainWritesOutputFile
TestExplainIntentRules
TestExplainBehaviorRules
TestExplainDesignRules
TestExplainAssuranceRules
TestExplainSecurityRules
TestExplainExecutionRules
```

## Codex prompt packet: Explain enhancement

```text
You are working in the Bottleneck Go CLI codebase.

Goal:
Enhance the existing `bottleneck explain` command so it produces deterministic, evidence-backed diagnosis useful for enterprise SDLC conversations.

Context:
The existing explain command should not be removed. Preserve backward compatibility where possible. Bottleneck should not use an LLM for this feature. The explanation should be rule-based and derived from existing validation, scorecard, traceability, and evidence models.

Requirements:
1. Add or confirm support for:
   - `--category`, optional
   - `--capability`, existing behavior should continue
   - `--format`, allowed values `text`, `markdown`, `json`
   - `--out`, optional report output path
   - `--env`
   - `--strict`
2. For each category, generate:
   - category name
   - status
   - score if available
   - why this matters
   - evidence found
   - evidence missing
   - risk to delivery
   - recommended next actions
   - suggested owner roles
   - suggested automation or GitHub Actions hooks
3. Implement deterministic rules for Intent, Behavior, Design, Assurance, Security, and Execution.
4. Do not call an LLM or external service.
5. Support text, markdown, and JSON rendering.
6. If `--out` is provided, write output to the path, creating parent directories as needed.
7. Add tests for category filtering, evidence found, evidence missing, risk, recommendations, owner roles, automation suggestions, markdown, JSON, and file output.
8. Update help text.
9. Preserve existing explain functionality and tests unless intentionally updated.
10. Run `go test ./...` and fix failures.

Acceptance:
- `bottleneck explain --category=assurance --format=markdown` returns a leadership-useful explanation.
- Existing commands continue to work.
- The feature remains deterministic and local-only.
```

---

# Implementation Epic 4: Generate leadership-ready SDLC evidence report

## Feature name

```text
bottleneck report
```

## Purpose

Generate a single Markdown or JSON report that a team can share with leadership to explain current state, trend, bottleneck, evidence, missing evidence, and recommended actions.

## User story

```text
As a development team,
I want Bottleneck to generate a leadership-ready SDLC evidence report,
so that I can communicate what needs to be fixed in our delivery system using evidence.
```

## CLI command

```sh
bottleneck report
```

## CLI options

```sh
bottleneck report --env=default
bottleneck report --format=markdown
bottleneck report --format=json
bottleneck report --out=bottleneck/reports/sdlc-evidence-report.md
bottleneck report --window=6
bottleneck report --strict
```

## Default behavior

When user runs:

```sh
bottleneck report
```

Bottleneck should:

```text
1. Run current validation/scorecard.
2. Load recent trends if snapshot history exists.
3. Generate explanation for primary bottleneck and weak categories.
4. Render a Markdown report.
5. Write to bottleneck/reports/sdlc-evidence-report.md.
6. Print the output path and short summary.
```

## Example terminal output

```text
SDLC evidence report created

Status: WARN
Primary bottleneck: Assurance
Trend: Improving, but Assurance remains persistent
Report: bottleneck/reports/sdlc-evidence-report.md
```

## Report sections

The Markdown report should include:

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

## Recommended report language

Keep the tone enterprise-friendly.

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

## Example report content

```markdown
# SDLC Evidence Report

## Executive Summary

The system is currently in WARN status. The primary bottleneck is Assurance.

The team has evidence of intent, behavior, architecture, and security controls. However, several expected behaviors are not mapped to automated validation evidence. This limits the team's ability to prove that the system behaves as intended before release.

## Current Delivery-System Status

- Status: WARN
- Release recommendation: Conditional
- Primary bottleneck: Assurance
- Environment: default

## Category Scorecard

| Category | Status | Summary |
|---|---|---|
| Intent | PASS | Intent evidence exists and includes measurable outcomes. |
| Behavior | WARN | Behavior expectations exist but are not fully validated. |
| Design | PASS | Architecture evidence exists. |
| Assurance | FAIL | Critical behaviors lack mapped validation evidence. |
| Security | PASS | Security evidence exists. |
| Execution | WARN | Production telemetry is incomplete. |

## Trend Summary

The team is improving overall, but Assurance has appeared as the primary bottleneck in 4 of the last 6 snapshots.

## Evidence Found

- Intent documentation exists.
- Behavior specification exists.
- Test evidence exists.
- Security guardrail evidence exists.

## Evidence Missing

- Some behavior IDs are not mapped to tests.
- Release validation evidence is missing.
- Execution telemetry is incomplete.

## Risk to Delivery

The team may ship software that appears complete but cannot be proven against intended behavior.

## Recommended Actions

1. Map critical behaviors to automated tests.
2. Add behavior IDs to Cucumber scenarios.
3. Add release validation evidence.
4. Add execution telemetry for production behavior.

## Suggested Owners

- Assurance Engineer
- Technical Lead
- Product Lead
- Domain Expert

## Leadership Decision Needed

Approve time in the next delivery cycle to close Assurance evidence gaps before increasing delivery throughput.
```

## Internal package

Add:

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
    GeneratedAt          time.Time
    Environment          string
    CurrentStatus        string
    ReleaseRecommendation string
    PrimaryBottleneck    string
    Scorecard            interface{}
    TrendSummary         *trends.Analysis
    Explanation          interface{}
    EvidenceFound        []string
    EvidenceMissing      []string
    Risks                []string
    RecommendedActions   []string
    SuggestedOwners      []string
    SuggestedAutomation  []string
    LeadershipDecision   string
}
```

## Leadership decision generation

Deterministic examples:

### If primary bottleneck is Assurance

```text
Approve time to map critical behaviors to validation evidence before accelerating release.
```

### If primary bottleneck is Security

```text
Decide whether release should be blocked until high-severity findings are resolved or formally accepted.
```

### If primary bottleneck is Intent

```text
Align leadership, product, domain, and engineering on measurable intent before expanding implementation.
```

### If primary bottleneck is Execution

```text
Prioritize production telemetry, adoption feedback, and operational readiness before scaling usage.
```

### If no primary bottleneck

```text
Continue monitoring evidence trends and maintain current release controls.
```

## Acceptance criteria

```text
AC1. `bottleneck report` creates `bottleneck/reports/sdlc-evidence-report.md`.
AC2. Report includes executive summary.
AC3. Report includes current status.
AC4. Report includes primary bottleneck.
AC5. Report includes category scorecard.
AC6. Report includes trend summary if snapshots exist.
AC7. Report handles missing trend history gracefully.
AC8. Report includes evidence found.
AC9. Report includes evidence missing.
AC10. Report includes risk to delivery.
AC11. Report includes recommended actions.
AC12. Report includes suggested owners.
AC13. Report includes suggested automation.
AC14. Report includes leadership decision needed.
AC15. Report supports JSON output.
AC16. Report supports custom output path.
AC17. Report does not require a database or external service.
AC18. Existing commands continue to work.
```

## Test cases

```text
TestReportCreatesDefaultMarkdownFile
TestReportIncludesExecutiveSummary
TestReportIncludesCurrentStatus
TestReportIncludesPrimaryBottleneck
TestReportIncludesCategoryScorecard
TestReportIncludesTrendSummaryWhenSnapshotsExist
TestReportHandlesMissingTrendHistory
TestReportIncludesEvidenceFound
TestReportIncludesEvidenceMissing
TestReportIncludesRisks
TestReportIncludesRecommendations
TestReportIncludesSuggestedOwners
TestReportIncludesSuggestedAutomation
TestReportIncludesLeadershipDecision
TestReportSupportsJSON
TestReportSupportsCustomOutputPath
```

## Codex prompt packet: Report feature

```text
You are working in the Bottleneck Go CLI codebase.

Goal:
Implement `bottleneck report`, which generates a leadership-ready SDLC evidence report from the current scorecard, explanation output, and trend history.

Context:
Bottleneck is a local-first Go CLI. It should not use a database, SaaS backend, dashboard, or external storage. Reports should be generated from local evidence files and local snapshot history.

Requirements:
1. Add a new Cobra command: `report`.
2. Supported flags:
   - `--env`, default `default`
   - `--format`, default `markdown`, allowed values `markdown`, `json`
   - `--out`, default `bottleneck/reports/sdlc-evidence-report.md`
   - `--window`, default 6
   - `--strict`, default false
3. The command should run current validation/scorecard logic.
4. It should load trend analysis from local snapshots if available.
5. It should generate explanation information for the primary bottleneck and weak categories.
6. It should render a Markdown report with these sections:
   - SDLC Evidence Report
   - Executive Summary
   - Current Delivery-System Status
   - Primary Bottleneck
   - Category Scorecard
   - Trend Summary
   - Evidence Found
   - Evidence Missing
   - Risk to Delivery
   - Recommended Actions
   - Suggested Owners
   - Suggested Automation
   - Leadership Decision Needed
   - Appendix: Snapshot Metadata
7. If no snapshots exist, the trend section should say that there is insufficient history and recommend running `bottleneck snapshot` over multiple delivery cycles.
8. Support JSON output.
9. Write the report to the `--out` path, creating parent directories as needed.
10. Print a concise success message showing status, primary bottleneck, trend if available, and report path.
11. Add tests for report generation, report sections, missing trend history, custom output path, markdown rendering, and JSON rendering.
12. Update root help text to include `report`.
13. Run `go test ./...` and fix failures.

Acceptance:
- `bottleneck report` produces a Markdown file useful for leadership conversations.
- It works without a database or external service.
- It does not break existing commands.
```

---

# Implementation Epic 5: Seed enterprise history for demos and tests

## Feature name

```text
bottleneck seed-history
```

## Purpose

Generate realistic historical snapshot data that demonstrates Bottleneck’s value in enterprise conversations.

This is useful for:

```text
- Demos
- Tests
- Documentation
- Conference talks
- Example repos
- Onboarding new users
```

## User story

```text
As a Bottleneck evaluator,
I want to generate realistic SDLC history,
so that I can understand how Bottleneck identifies improvement, regression, and persistent bottlenecks over time.
```

## CLI command

```sh
bottleneck seed-history
```

## CLI options

```sh
bottleneck seed-history --scenario=saas-day-one
bottleneck seed-history --env=default
bottleneck seed-history --snapshots=6
bottleneck seed-history --out=bottleneck/history/scorecards
bottleneck seed-history --overwrite
```

## Recommended first implementation

Implement one scenario:

```text
saas-day-one
```

Later scenarios:

```text
ai-product
regulated-enterprise
security-regression
execution-drift
```

## Seed scenario: SaaS day-one success

Generate 6 snapshots.

### Snapshot 1: Fast demo, weak evidence

```text
Status: FAIL
Primary bottleneck: Intent
Intent: FAIL
Behavior: FAIL
Design: WARN
Assurance: FAIL
Security: WARN
Execution: FAIL
```

Narrative:

```text
The team has code, but cannot clearly explain what outcome the system is designed to produce.
```

### Snapshot 2: Intent clarified

```text
Status: FAIL
Primary bottleneck: Behavior
Intent: PASS
Behavior: FAIL
Design: WARN
Assurance: FAIL
Security: WARN
Execution: FAIL
```

Narrative:

```text
Intent improved, but expected and unacceptable behaviors are not measurable.
```

### Snapshot 3: Behavior documented, assurance weak

```text
Status: WARN
Primary bottleneck: Assurance
Intent: PASS
Behavior: PASS
Design: WARN
Assurance: FAIL
Security: WARN
Execution: FAIL
```

Narrative:

```text
Behavior is documented, but tests are not mapped to behavior evidence.
```

### Snapshot 4: Tests added, security regresses

```text
Status: WARN
Primary bottleneck: Security
Intent: PASS
Behavior: PASS
Design: PASS
Assurance: WARN
Security: FAIL
Execution: WARN
```

Narrative:

```text
Validation improved, but a high-severity security issue was introduced.
```

### Snapshot 5: Security recovered, execution weak

```text
Status: WARN
Primary bottleneck: Execution
Intent: PASS
Behavior: PASS
Design: PASS
Assurance: PASS
Security: PASS
Execution: WARN
```

Narrative:

```text
Pre-release evidence improved, but production/adoption evidence is incomplete.
```

### Snapshot 6: Stable release candidate

```text
Status: PASS
Primary bottleneck: None
Intent: PASS
Behavior: PASS
Design: PASS
Assurance: PASS
Security: PASS
Execution: PASS
```

Narrative:

```text
The delivery system can now prove intent, behavior, assurance, security, and execution evidence together.
```

## Acceptance criteria

```text
AC1. `bottleneck seed-history` creates realistic snapshot files.
AC2. Seed snapshots use the same schema as real snapshots.
AC3. Seed snapshots work with `bottleneck trends`.
AC4. Seed snapshots work with `bottleneck report`.
AC5. Default scenario is `saas-day-one`.
AC6. Existing snapshot files are not overwritten unless `--overwrite` is passed.
AC7. Command does not require a database or external service.
```

## Test cases

```text
TestSeedHistoryCreatesSnapshots
TestSeedHistoryUsesSnapshotSchema
TestSeedHistoryDoesNotOverwriteByDefault
TestSeedHistoryOverwriteFlag
TestSeedHistoryWorksWithTrends
TestSeedHistoryScenarioSaaSDayOne
```

## Codex prompt packet: Seed history feature

```text
You are working in the Bottleneck Go CLI codebase.

Goal:
Implement `bottleneck seed-history`, which generates realistic historical scorecard snapshots for demos, tests, and onboarding.

Context:
Bottleneck uses local snapshot files under `bottleneck/history/scorecards/` for trend analysis. This command should generate sample history using the same snapshot schema as `bottleneck snapshot`.

Requirements:
1. Add a new Cobra command: `seed-history`.
2. Supported flags:
   - `--scenario`, default `saas-day-one`
   - `--env`, default `default`
   - `--snapshots`, default 6
   - `--out`, default `bottleneck/history/scorecards`
   - `--overwrite`, default false
3. Implement scenario `saas-day-one`.
4. Generate 6 snapshots showing maturity progression:
   - Snapshot 1: primary bottleneck Intent, status FAIL
   - Snapshot 2: primary bottleneck Behavior, status FAIL
   - Snapshot 3: primary bottleneck Assurance, status WARN
   - Snapshot 4: primary bottleneck Security, status WARN
   - Snapshot 5: primary bottleneck Execution, status WARN
   - Snapshot 6: no primary bottleneck, status PASS
5. Use the same snapshot schema used by `bottleneck snapshot`.
6. Do not overwrite existing files unless `--overwrite` is provided.
7. Generated snapshots must work with `bottleneck trends`.
8. Generated snapshots must work with `bottleneck report`.
9. Add tests for file creation, schema compatibility, overwrite behavior, and trends compatibility.
10. Update root help text.
11. Run `go test ./...` and fix failures.

Acceptance:
- A user can run `bottleneck seed-history`, then `bottleneck trends`, then `bottleneck report`.
- No database or external service is added.
```

---

# Implementation Epic 6: Improve scorecard JSON stability

## Feature name

```text
Stable scorecard JSON contract
```

## Purpose

Ensure that `snapshot`, `trends`, and `report` can rely on a stable machine-readable scorecard structure.

## User story

```text
As a Bottleneck feature developer,
I want a stable scorecard JSON contract,
so that snapshot, trends, report, and future integrations do not break when text output changes.
```

## Problem

Text output can evolve. JSON output should be treated as a contract.

## Implementation requirements

Create or formalize a scorecard DTO.

Suggested structure:

```json
{
  "schema_version": "scorecard.v1",
  "environment": "default",
  "system_status": "WARN",
  "release_recommendation": "Conditional",
  "primary_bottleneck": "Assurance",
  "categories": [
    {
      "name": "Intent",
      "status": "PASS",
      "score": 90,
      "summary": "Intent evidence exists and includes measurable outcomes.",
      "evidence_found": [],
      "evidence_missing": [],
      "recommendations": []
    }
  ]
}
```

## Rules

```text
- JSON field names should use snake_case.
- Schema version should be explicit.
- Unknown fields may be added later.
- Existing JSON output should not be broken unless tests are updated intentionally.
- Trends should consume this stable structure.
- Reports should consume this stable structure.
```

## Acceptance criteria

```text
AC1. Scorecard JSON includes schema_version.
AC2. Scorecard JSON includes environment.
AC3. Scorecard JSON includes system_status.
AC4. Scorecard JSON includes release_recommendation.
AC5. Scorecard JSON includes primary_bottleneck.
AC6. Scorecard JSON includes category objects.
AC7. Each category includes name, status, score if available, summary, evidence_found, evidence_missing, and recommendations if available.
AC8. Snapshot uses this stable scorecard structure.
AC9. Trends uses this stable scorecard structure.
AC10. Tests protect this JSON contract.
```

## Test cases

```text
TestScorecardJSONIncludesSchemaVersion
TestScorecardJSONIncludesEnvironment
TestScorecardJSONIncludesPrimaryBottleneck
TestScorecardJSONIncludesCategories
TestScorecardJSONCategoryFields
TestScorecardJSONBackwardsCompatibleForSnapshot
```

## Codex prompt packet: Stable scorecard JSON

```text
You are working in the Bottleneck Go CLI codebase.

Goal:
Stabilize the JSON output contract for `bottleneck scorecard --format=json` so that snapshot, trends, report, and future integrations can rely on it.

Requirements:
1. Ensure scorecard JSON includes:
   - schema_version
   - environment
   - system_status
   - release_recommendation
   - primary_bottleneck
   - categories
2. Each category should include:
   - name
   - status
   - score, if available
   - summary, if available
   - evidence_found, if available
   - evidence_missing, if available
   - recommendations, if available
3. Use snake_case JSON field names.
4. Do not break text or markdown scorecard output.
5. Update or add tests that protect the JSON contract.
6. Ensure `bottleneck snapshot` uses this stable JSON structure.
7. Ensure `bottleneck trends` can parse this structure.
8. Run `go test ./...` and fix failures.

Acceptance:
- `bottleneck scorecard --format=json` produces predictable machine-readable output.
- Existing CLI behavior remains intact.
```

---

# Implementation Epic 7: GitHub Actions example for enterprise evidence package

## Feature name

```text
Enterprise evidence workflow example
```

## Purpose

Show teams how to generate Bottleneck evidence in CI without requiring SaaS infrastructure.

## User story

```text
As a development team,
I want a GitHub Actions example that generates Bottleneck snapshots and reports,
so that I can include SDLC evidence in pull requests, release reviews, and leadership updates.
```

## Add example file

```text
examples/github-actions/bottleneck-evidence-report.yml
```

## Example workflow

```yaml
name: Bottleneck Evidence Report

on:
  pull_request:
  workflow_dispatch:

jobs:
  bottleneck-evidence:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Build Bottleneck
        run: go build -o bottleneck-cli .

      - name: Validate evidence
        run: ./bottleneck-cli validate

      - name: Generate scorecard
        run: ./bottleneck-cli scorecard --format=markdown --details

      - name: Create snapshot
        run: ./bottleneck-cli snapshot --label=ci

      - name: Generate trends
        run: ./bottleneck-cli trends --format=markdown --out=bottleneck/reports/trend-summary.md

      - name: Generate SDLC evidence report
        run: ./bottleneck-cli report --format=markdown --out=bottleneck/reports/sdlc-evidence-report.md

      - name: Upload Bottleneck reports
        uses: actions/upload-artifact@v4
        with:
          name: bottleneck-reports
          path: |
            bottleneck/history/
            bottleneck/reports/
```

Do not auto-commit snapshots in the default workflow.

Provide a separate release workflow later if needed.

## Acceptance criteria

```text
AC1. Add GitHub Actions example for evidence reports.
AC2. Workflow runs validate, scorecard, snapshot, trends, and report.
AC3. Workflow uploads generated history and reports as artifacts.
AC4. Workflow does not require external services.
AC5. Documentation explains that teams may commit snapshots intentionally if they want Git-tracked trend history.
```

## Codex prompt packet: GitHub Actions evidence example

```text
You are working in the Bottleneck Go CLI codebase.

Goal:
Add a GitHub Actions example showing how an enterprise team can generate Bottleneck SDLC evidence reports in CI.

Requirements:
1. Add `examples/github-actions/bottleneck-evidence-report.yml`.
2. The workflow should:
   - checkout code
   - set up Go
   - build Bottleneck
   - run `bottleneck validate`
   - run `bottleneck scorecard --format=markdown --details`
   - run `bottleneck snapshot --label=ci`
   - run `bottleneck trends --format=markdown --out=bottleneck/reports/trend-summary.md`
   - run `bottleneck report --format=markdown --out=bottleneck/reports/sdlc-evidence-report.md`
   - upload `bottleneck/history/` and `bottleneck/reports/` as artifacts
3. Do not auto-commit generated files in this workflow.
4. Update documentation to explain how CI-generated reports differ from committed trend history.
5. Add or update tests if the repo already tests workflow examples.
6. Run `go test ./...` and fix failures.

Acceptance:
- The example workflow is present.
- Documentation explains intended usage.
- Existing tests pass.
```

---

# Implementation Epic 8: Documentation for enterprise usage

## Feature name

```text
Enterprise evidence documentation
```

## Purpose

Make the feature understandable to enterprise teams.

## Add document

```text
docs/enterprise-sdlc-evidence.md
```

## Document sections

```markdown
# Enterprise SDLC Evidence with Bottleneck

## What Bottleneck Does

## What Bottleneck Does Not Do

## Why Git Is the System of Record

## Recommended Team Workflow

## Recommended CI Workflow

## Snapshot History

## Trend Analysis

## Evidence Explanation

## Leadership Report

## How to Interpret Results

## Common Enterprise Scenarios

## Frequently Asked Questions
```

## Important positioning

Include:

```text
Bottleneck is not a dashboard.
Bottleneck is not Jira replacement.
Bottleneck is not a project-management system.
Bottleneck is not a database-backed metrics warehouse.
Bottleneck is a local-first evidence system for understanding SDLC bottlenecks.
```

## Explain team value

```text
Developers get a structured way to show missing evidence.
Tech leads get a way to prioritize systemic fixes.
Product leaders get a way to connect intent to delivery.
Security teams get evidence of guardrails and risk.
Leadership gets a decision-ready summary.
```

## Common enterprise scenarios

```text
Scenario 1: Team is shipping quickly but failing validation.
Scenario 2: Security is repeatedly discovered late.
Scenario 3: Requirements are implemented but not traceable.
Scenario 4: Production behavior does not match expected behavior.
Scenario 5: Leadership wants faster delivery but the SDLC evidence is declining.
```

## Acceptance criteria

```text
AC1. Documentation explains snapshot.
AC2. Documentation explains trends.
AC3. Documentation explains report.
AC4. Documentation explains no database requirement.
AC5. Documentation includes local workflow.
AC6. Documentation includes CI workflow.
AC7. Documentation includes enterprise interpretation examples.
AC8. Documentation is linked from README.
```

## Codex prompt packet: Documentation

```text
You are working in the Bottleneck Go CLI codebase.

Goal:
Add enterprise-facing documentation for the SDLC Evidence Package milestone.

Requirements:
1. Add `docs/enterprise-sdlc-evidence.md`.
2. Explain:
   - what Bottleneck does
   - what Bottleneck does not do
   - why Git is the system of record
   - how snapshots work
   - how trends work
   - how explain works
   - how reports work
   - how to use Bottleneck locally
   - how to use Bottleneck in CI
   - how leadership should interpret results
3. Include common enterprise scenarios:
   - shipping fast but validation is weak
   - late security discovery
   - missing traceability
   - production behavior drift
   - leadership asking for faster delivery despite declining evidence
4. Link the document from README.
5. Keep tone enterprise-friendly and practical.
6. Run `go test ./...` if documentation tests exist.

Acceptance:
- Documentation exists.
- README links to it.
- Documentation reinforces local-first, no-database architecture.
```

---

# Recommended implementation order

Use this order to reduce risk:

```text
1. Stabilize scorecard JSON contract.
2. Add snapshot.
3. Add seed-history.
4. Add trends.
5. Enhance explain.
6. Add report.
7. Add GitHub Actions example.
8. Add enterprise documentation.
```

Why this order:

```text
Stable JSON gives the rest of the milestone a contract.
Snapshot creates historical data.
Seed-history makes trends easy to test and demo.
Trends creates the historical insight.
Explain creates the diagnosis.
Report packages everything for leadership.
CI example shows enterprise usage.
Docs make it adoptable.
```

---

# Milestone-level acceptance criteria

The milestone is complete when:

```text
AC1. A team can run `bottleneck snapshot` and create a local, timestamped scorecard snapshot.
AC2. A team can commit Bottleneck snapshots to Git.
AC3. A team can run `bottleneck trends` and see improvement, decline, stability, regressions, and persistent bottlenecks.
AC4. A team can run `bottleneck explain` and understand why a category is weak.
AC5. A team can run `bottleneck report` and generate a leadership-ready SDLC evidence report.
AC6. A team can generate seed history for demos and tests.
AC7. A GitHub Actions example exists for generating evidence artifacts in CI.
AC8. Enterprise documentation explains how and why to use the workflow.
AC9. No database or third-party storage is introduced.
AC10. Existing commands continue to work.
AC11. `go test ./...` passes.
```

---

# Milestone non-goals

Do not implement these in this milestone:

```text
Dashboard
Database
SaaS backend
Multi-repo aggregation
Organization-level portfolio view
Jira integration
Linear integration
ADO integration
Role-based access control
Authentication
Long-running agent service
LLM-generated recommendations
Auto-remediation
Automatic Git commits by default
Pull request comment bot
```

Those are later maturity features.

The goal of this milestone is narrower and stronger:

```text
A single development team can generate an evidence-backed SDLC report from its repo.
```

---

# Final milestone prompt for local Codex

Use this as the umbrella prompt to ask Codex to generate the sub-feature prompts or implementation plan.

```text
You are working in the Bottleneck Go CLI codebase.

Milestone:
Enterprise SDLC Evidence Package

Objective:
Enable a development team to use Bottleneck to gather local evidence, compare SDLC maturity over time, explain current bottlenecks, and generate a leadership-ready report describing what the team needs to address in its SDLC.

Architecture principle:
Do not add a database, SaaS backend, external storage, dashboard, or project-management dependency. Git and local files are the system of record.

Existing product:
Bottleneck is a Go CLI with commands such as init, validate, scorecard, diagnose, trace, ingest, and explain. It uses local evidence artifacts under the `bottleneck/` folder and categories such as Intent, Behavior, Design, Assurance, Security, and Execution.

Milestone features:
1. Stabilize scorecard JSON output.
2. Add `bottleneck snapshot`.
3. Add `bottleneck seed-history`.
4. Add `bottleneck trends`.
5. Enhance `bottleneck explain`.
6. Add `bottleneck report`.
7. Add GitHub Actions example for evidence reports.
8. Add enterprise documentation.

Implementation constraints:
- Preserve existing command behavior.
- Preserve existing tests unless intentionally updated.
- Add tests for every new command and rendering path.
- Prefer deterministic rule-based logic over LLM calls.
- Use local files only.
- Use Git metadata only when available.
- Do not fail if Git metadata is unavailable.
- Run `go test ./...` after each feature.

Desired workflow:
A user should be able to run:

  bottleneck snapshot
  bottleneck trends
  bottleneck explain
  bottleneck report

And receive:
- timestamped scorecard snapshots
- trend analysis
- evidence-backed bottleneck explanation
- leadership-ready SDLC evidence report

Create implementation prompts for each sub-feature. Each prompt should be independently executable by Codex, include acceptance criteria, describe tests to add, and explicitly state that existing functionality must not be broken.
```

# AI Implementation Prompt: Snapshot Current Scorecard Evidence

You are working in the Bottleneck Go CLI codebase.

Implement **Implementation Epic 1: Snapshot current scorecard evidence** from the `Enterprise SDLC Evidence Package` milestone.

## Milestone Context

Bottleneck should become a local-first enterprise SDLC evidence system. The team should be able to answer:

- Where are we today?
- Where have we been?
- Are we improving or getting worse?
- What is the primary SDLC bottleneck?
- What evidence supports that conclusion?
- What should we address next?
- What decision or support do we need from leadership?

The architecture principle is strict:

- Do not add a database.
- Do not add a SaaS backend.
- Do not add external storage.
- Do not add a dashboard.
- Do not add Jira, Linear, ADO, or third-party metrics dependencies.
- Git and local files are the system of record.

This epic creates the first historical artifact: a timestamped scorecard snapshot that can be committed to Git and used later by trend and report features.

## Scope

Add:

```sh
bottleneck snapshot
```

Recommended first-version flags:

```sh
bottleneck snapshot --env=default
bottleneck snapshot --env=production
bottleneck snapshot --label=release-candidate
bottleneck snapshot --out=bottleneck/history/scorecards
bottleneck snapshot --strict
bottleneck snapshot --no-latest
```

Defer `--commit` unless it is already trivial and safe. Auto-committing introduces Git safety questions and should not block this milestone.

## Current Code To Inspect

Read the existing implementation before changing code:

- `cmd/scorecard.go`
- `cmd/validate.go`
- `cmd/root.go`
- `cmd/*_test.go`
- `internal/scorecard/*`
- `internal/validator/*`
- `internal/gate/*`
- any existing config/environment loading code
- any existing JSON rendering helpers

Reuse the current validation and scorecard model. Do not duplicate scorecard logic if it can be factored cleanly.

## Behavior

When the user runs:

```sh
bottleneck snapshot
```

Bottleneck should:

1. Run the existing validation and scorecard logic.
2. Generate a JSON scorecard snapshot.
3. Add snapshot metadata.
4. Write a timestamped file to `bottleneck/history/scorecards/`.
5. Write or update `bottleneck/history/latest/default.json`.
6. Print a concise success message.

Important behavior decision:

- Snapshot creation should still happen when the scorecard status is FAIL.
- A failed scorecard is valuable historical evidence.
- For this first implementation, `snapshot` should exit zero when the snapshot was written successfully, even if the scorecard status is FAIL.
- Exit non-zero for real runtime errors such as unreadable config, invalid flags, or inability to write output.

Do not change `bottleneck scorecard` exit semantics.

## Output Paths

Default timestamped snapshot directory:

```text
bottleneck/history/scorecards/
```

Default latest snapshot directory:

```text
bottleneck/history/latest/
```

Filename convention:

```text
YYYY-MM-DDTHHMMSSZ-{env}-scorecard.json
```

Examples:

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

Use UTC timestamps.

## Snapshot JSON Schema

Snapshot JSON should be stable and machine-readable:

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
    "schema_version": "scorecard.v1",
    "environment": "default",
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

If Epic 6 has already stabilized the scorecard JSON contract, use that DTO directly. If not, use the current scorecard JSON model and keep the snapshot wrapper stable enough for trends to parse later.

## Suggested Internal Packages

Add if appropriate:

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
    SchemaVersion string   `json:"schema_version"`
    Snapshot      Metadata `json:"snapshot"`
    Scorecard     any      `json:"scorecard"`
}
```

Use a concrete scorecard DTO instead of `any` if one already exists.

Add if appropriate:

```text
internal/gitinfo/
  gitinfo.go
  gitinfo_test.go
```

Suggested type:

```go
type Info struct {
    Commit string `json:"commit"`
    Branch string `json:"branch"`
    Dirty  *bool  `json:"dirty,omitempty"`
}
```

Suggested function:

```go
func Detect(root string) Info
```

Implementation may shell out to:

```sh
git rev-parse --short HEAD
git rev-parse --abbrev-ref HEAD
git status --porcelain
```

Rules:

- If not in a Git repo, return empty values.
- Do not fail the command.
- Keep command timeouts short if using context.
- Tests should not require a remote repository.

## CLI Output

Recommended success output:

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

When `--no-latest` is used, omit `Latest` or show `Latest: skipped`.

## Acceptance Criteria

- Running `bottleneck snapshot` creates `bottleneck/history/scorecards/`.
- Running `bottleneck snapshot` writes a timestamped JSON snapshot file.
- Running `bottleneck snapshot` updates `bottleneck/history/latest/default.json` unless `--no-latest` is used.
- Running `bottleneck snapshot --env=production` writes a production-specific snapshot.
- Running `bottleneck snapshot --label=release-candidate` includes the label in metadata and filename.
- Snapshot JSON includes `schema_version`, snapshot metadata, Git metadata, and scorecard data.
- Snapshot creation does not require a database or external service.
- Snapshot creation does not break existing scorecard behavior.
- Snapshot command has tests for filename generation, metadata generation, latest-file behavior, and label sanitization.
- Snapshot command is included in root help text.

## Tests To Add

Add tests in the most appropriate packages:

- `cmd/snapshot_test.go` for command wiring and temp-directory behavior.
- `internal/snapshot/snapshot_test.go` for filename, metadata, latest path, and JSON behavior.
- `internal/gitinfo/gitinfo_test.go` for no-Git behavior and Git repo behavior where practical.

Required tests:

- `TestSnapshotCreatesHistoryDirectory`
- `TestSnapshotWritesTimestampedFile`
- `TestSnapshotWritesLatestFile`
- `TestSnapshotNoLatestFlagSkipsLatestFile`
- `TestSnapshotIncludesEnvironment`
- `TestSnapshotIncludesLabel`
- `TestSnapshotSanitizesLabel`
- `TestSnapshotIncludesGitMetadataWhenAvailable`
- `TestSnapshotDoesNotFailOutsideGitRepo`
- `TestSnapshotDoesNotBreakExistingScorecardCommand`

Prefer injecting a clock into snapshot generation tests so filename tests are deterministic.

Use temporary directories. Do not create history files in the repository root during tests.

## Implementation Constraints

- Keep changes small and reviewable.
- Do not break existing commands.
- Do not add external storage.
- Do not add a database.
- Do not introduce a dashboard or service process.
- Do not require Git to be available.
- Do not auto-commit snapshots by default.
- Preserve existing scorecard behavior and output formats.

## Verification

Run:

```sh
go test ./...
```

If feasible, manually verify from a temporary directory:

```sh
bottleneck init --template saas
bottleneck snapshot
bottleneck snapshot --env=production --label=release-candidate
bottleneck snapshot --no-latest
```

For manual verification, confirm the timestamped files and latest files contain valid JSON.

## Final Response Requirements

When finished, report:

1. Snapshot command behavior.
2. Files and packages added or changed.
3. Snapshot schema details.
4. Git metadata behavior.
5. Tests added or changed.
6. Exact commands run and results.
7. Any acceptance criteria intentionally deferred and why.


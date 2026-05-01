# AI Implementation Prompt: GitHub Actions Enterprise Evidence Workflow

You are working in the Bottleneck Go CLI codebase.

Implement **Implementation Epic 7: GitHub Actions example for enterprise evidence package** from the `Enterprise SDLC Evidence Package` milestone.

## Milestone Context

Bottleneck should let teams generate SDLC evidence artifacts in CI without requiring SaaS infrastructure. CI should validate evidence, create snapshots, generate trends, generate reports, and upload those artifacts for review.

This epic adds a GitHub Actions example only. It should not add a hosted service, PR bot, database, or automatic Git commit behavior.

## Scope

Add:

```text
examples/github-actions/bottleneck-evidence-report.yml
```

The workflow should:

- checkout code
- set up Go
- build Bottleneck
- run `bottleneck validate`
- run `bottleneck scorecard --format=markdown --details`
- run `bottleneck snapshot --label=ci`
- run `bottleneck trends --format=markdown --out=bottleneck/reports/trend-summary.md`
- run `bottleneck report --format=markdown --out=bottleneck/reports/sdlc-evidence-report.md`
- upload `bottleneck/history/` and `bottleneck/reports/` as artifacts

Do not auto-commit snapshots in the default workflow.

## Current Code To Inspect

Read before changing files:

- existing `examples/github-actions/*`
- existing `.github/workflows/*`
- README CI documentation
- `docs/enterprise-sdlc-evidence.md`, if implemented
- command availability for `snapshot`, `trends`, `report`
- existing workflow validation tests

If a command is not implemented yet, do not create a workflow that references a non-existent command unless this example is intentionally staged for the completed milestone and the docs clearly say it requires the enterprise evidence package features.

## Workflow File

Create:

```text
examples/github-actions/bottleneck-evidence-report.yml
```

Recommended workflow:

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
          go-version-file: go.mod

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

Use `go-version-file: go.mod` if that is compatible with existing repo workflows. If the repo pins Go versions another way, follow the local convention.

## Documentation Requirements

Update README or enterprise docs to explain:

- This workflow generates evidence artifacts in CI.
- It uploads `bottleneck/history/` and `bottleneck/reports/` as artifacts.
- It does not auto-commit generated snapshots.
- Teams may intentionally commit snapshots to Git if they want Git-tracked trend history.
- CI-generated reports are useful for PR and release review, but committed snapshots are needed for long-lived trend history across runs unless artifacts are retained.

Recommended wording:

```text
The CI workflow generates a snapshot and report for the current run and uploads them as artifacts. It does not commit those files automatically. If your team wants Git to be the long-term trend history, review and commit snapshot files intentionally.
```

## Tests To Add

If the repo already tests workflow examples, update those tests. Otherwise add a small test that verifies:

- `examples/github-actions/bottleneck-evidence-report.yml` exists.
- Workflow contains `actions/checkout`.
- Workflow sets up Go.
- Workflow builds Bottleneck.
- Workflow runs `validate`.
- Workflow runs `scorecard --format=markdown`.
- Workflow runs `snapshot --label=ci`.
- Workflow runs `trends --format=markdown --out=bottleneck/reports/trend-summary.md`.
- Workflow runs `report --format=markdown --out=bottleneck/reports/sdlc-evidence-report.md`.
- Workflow uploads `bottleneck/history/` and `bottleneck/reports/`.
- Workflow does not contain `git commit`.
- Workflow does not contain `secrets.`.
- Workflow does not contain machine-specific absolute paths.

Use a YAML parser only if the repository already uses one. Otherwise, lightweight content checks are acceptable.

## Acceptance Criteria

- Add GitHub Actions example for evidence reports.
- Workflow runs validate, scorecard, snapshot, trends, and report.
- Workflow uploads generated history and reports as artifacts.
- Workflow does not require external services.
- Documentation explains that teams may commit snapshots intentionally if they want Git-tracked trend history.

## Implementation Constraints

- Do not add a PR comment bot.
- Do not add auto-commit behavior.
- Do not require secrets.
- Do not require external services.
- Do not reference commands that are not implemented unless clearly staged for the completed milestone.
- Keep workflow copy/paste friendly.

## Verification

Run:

```sh
go test ./...
```

If a workflow validator exists, run it. Do not add a network-dependent linter.

Manually inspect:

```sh
sed -n '1,220p' examples/github-actions/bottleneck-evidence-report.yml
```

## Final Response Requirements

When finished, report:

1. Workflow file added or changed.
2. Workflow steps included.
3. Documentation updates.
4. Workflow tests added or changed.
5. Exact commands run and results.
6. Any acceptance criteria intentionally deferred and why.


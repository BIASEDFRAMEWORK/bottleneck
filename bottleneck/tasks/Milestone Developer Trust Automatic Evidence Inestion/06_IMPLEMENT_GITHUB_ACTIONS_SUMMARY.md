# Codex Prompt 06: Implement GitHub Actions assessment summary support

## Objective

Make Bottleneck easy to use in GitHub Actions by generating Markdown suitable for `$GITHUB_STEP_SUMMARY` and optional machine-readable artifacts.

This reinforces developer trust by making maturity evidence visible in PR and CI workflows without requiring a dashboard.

## Required behavior

Add a Markdown-focused output option to `assess` and/or a dedicated command.

Preferred option:

```bash
bottleneck assess --format markdown
```

Optional dedicated command if cleaner:

```bash
bottleneck github-summary
```

The output should be suitable for:

```bash
bottleneck assess --format markdown >> "$GITHUB_STEP_SUMMARY"
```

## Markdown output requirements

The Markdown should include:

```markdown
# Bottleneck SDLC Maturity Assessment

| Field | Result |
|---|---|
| Overall Maturity | Level 2 - Managed |
| AI Readiness | Limited |
| Release Friction | Medium |
| Primary Bottleneck | Assurance |
| Score Confidence | Medium |
| Release Recommendation | Conditional |

## What Bottleneck Found

- ✅ GitHub Actions workflow detected
- ✅ Cucumber test report detected and ingested
- ✅ SARIF security report detected and ingested
- ⚠️ No production telemetry freshness signal detected
- ⚠️ 1 behavior has no mapped passing assurance evidence

## Next Action

Add or ingest assurance evidence mapped to `BEHAVIOR-003`.

## Useful Commands

```bash
bottleneck trace BEHAVIOR-003
bottleneck explain-score
bottleneck report
```
```

Make sure code fences render correctly in the actual implementation.

## GitHub Actions example

Create or update:

```text
docs/github-actions.md
examples/github-actions/bottleneck-assessment.yml
```

Example workflow should include:

```yaml
name: Bottleneck SDLC Assessment

on:
  pull_request:
  workflow_dispatch:

jobs:
  bottleneck:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Install Bottleneck
        run: go install ./...

      - name: Discover evidence
        run: bottleneck discover

      - name: Ingest evidence
        run: bottleneck ingest --auto

      - name: Publish assessment summary
        run: bottleneck assess --format markdown >> "$GITHUB_STEP_SUMMARY"

      - name: Save assessment JSON
        run: bottleneck assess --format json > bottleneck-assessment.json

      - name: Upload assessment artifact
        uses: actions/upload-artifact@v4
        with:
          name: bottleneck-assessment
          path: bottleneck-assessment.json
```

Adjust install steps to match the repo's actual module setup. If `go install ./...` is not appropriate, use the correct local CLI build command.

## Optional release gate mode

If the existing CLI already supports gates, document it. If not, add only a simple flag if low-risk:

```bash
bottleneck assess --gate release
```

Behavior:

- Exit code 0 for Ready or Conditional.
- Exit code 1 for Blocked.
- Avoid making this the default.

Important: Do not make teams fail CI by default. The milestone is about trust first, enforcement later.

## Files likely to change

- `cmd/assess.go`
- `cmd/assess_test.go`
- possibly new `internal/markdown/*`
- `docs/github-actions.md`
- `examples/github-actions/bottleneck-assessment.yml`
- root README if needed

## Implementation constraints

- Markdown output must not include ANSI color codes.
- JSON output must remain JSON only.
- Text output should remain friendly terminal output.
- Do not require GitHub environment variables for markdown generation.
- Do not make network calls.
- Do not require a GitHub token.

## Tests to add

Required tests:

1. `assess --format markdown` emits a Markdown heading and a summary table.
2. Markdown output includes maturity, AI readiness, primary bottleneck, score confidence, and release recommendation.
3. Markdown output includes next action and useful commands.
4. JSON output remains valid JSON and unaffected by Markdown support.
5. Text output remains unchanged or intentionally updated.
6. Example workflow file exists and references `bottleneck assess --format markdown`.
7. `go test ./...` passes.

## Acceptance criteria

- Teams can publish Bottleneck assessment results into GitHub Actions summaries.
- CI usage is visible but not punitive by default.
- Markdown, text, and JSON outputs are distinct and stable.

## Commit message suggestion

```text
Add GitHub Actions assessment summary output
```

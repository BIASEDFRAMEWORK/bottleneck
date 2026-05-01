# Codex Prompt 01: Implement `bottleneck discover`

## Objective

Add a new `bottleneck discover` command that scans the local repository for SDLC evidence sources without modifying files.

This command reduces onboarding overload by showing developers what Bottleneck can already use before they manually configure anything.

## Required behavior

Add:

```bash
bottleneck discover
```

Default output should be human-readable text.

Add optional JSON output:

```bash
bottleneck discover --format json
```

Do not write evidence files in this command. This is read-only discovery.

## Evidence sources to detect

Implement a discovery package, likely under:

```text
internal/discover/
```

Detect these categories.

### Assurance evidence

Detect common test and coverage outputs:

- `reports/cucumber.json`
- `target/cucumber.json`
- `cucumber.json`
- `reports/junit.xml`
- `build/test-results/**/*.xml`
- `test-results/**/*.xml`
- `coverage/lcov.info`
- `coverage/cobertura.xml`
- `coverage/coverage-final.json`
- `playwright-report/results.json`
- `cypress/results/*.json`

### Security evidence

Detect:

- `reports/*.sarif`
- `results/*.sarif`
- `codeql-results.sarif`
- `semgrep.sarif`
- `trivy.sarif`
- `npm-audit.json`
- `osv-scanner.json`

### CI/CD and execution evidence

Detect:

- `.github/workflows/*.yml`
- `.github/workflows/*.yaml`
- `reports/telemetry.json`
- `bottleneck/execution/telemetry.json`
- `deployment-summary.json`
- `release-summary.json`

### Design and architecture evidence

Detect:

- `README.md` or `readme.md`
- `docs/architecture.md`
- `docs/adr/*.md`
- `openapi.yaml`
- `openapi.yml`
- `swagger.yaml`
- `swagger.yml`
- `docs/**/*.md`

### Bottleneck-native evidence

Detect:

- `bottleneck/config.yaml`
- `bottleneck/assurance/results.json`
- `bottleneck/security/guardrails.json`
- `bottleneck/execution/telemetry.json`
- existing intent, behavior, design, or other native artifacts created by `init --template saas`

## Suggested data model

Create a model similar to:

```go
type DiscoveryResult struct {
    RootPath string `json:"root_path"`
    Findings []DiscoveryFinding `json:"findings"`
    Summary DiscoverySummary `json:"summary"`
}

type DiscoveryFinding struct {
    Category string `json:"category"` // assurance, security, execution, design, intent, behavior, native
    Kind string `json:"kind"` // cucumber, junit, sarif, github-actions, telemetry, readme, adr, etc.
    Path string `json:"path"`
    Confidence string `json:"confidence"` // high, medium, low
    SuggestedCommand string `json:"suggested_command,omitempty"`
    Notes []string `json:"notes,omitempty"`
}

type DiscoverySummary struct {
    AssuranceSources int `json:"assurance_sources"`
    SecuritySources int `json:"security_sources"`
    ExecutionSources int `json:"execution_sources"`
    DesignSources int `json:"design_sources"`
    NativeSources int `json:"native_sources"`
    Missing []string `json:"missing"`
}
```

Use existing models if appropriate, but keep discovery separate from scoring.

## Suggested text output

Example:

```text
Bottleneck evidence discovery

Found evidence sources:

Assurance
  ✓ reports/cucumber.json                 cucumber       high
    Suggested: bottleneck ingest cucumber --file reports/cucumber.json
  ✓ coverage/lcov.info                    coverage       medium

Security
  ✓ reports/codeql.sarif                  sarif          high
    Suggested: bottleneck ingest sarif --file reports/codeql.sarif

Execution
  ✓ .github/workflows/ci.yml              github-actions high
  ⚠ No telemetry evidence detected

Native Bottleneck Evidence
  ✓ bottleneck/config.yaml
  ✓ bottleneck/assurance/results.json

Next action:
  Run bottleneck ingest --auto to convert detected tool outputs into Bottleneck evidence.
```

## Suggested commands for findings

For known ingestable evidence, include suggested commands:

- Cucumber: `bottleneck ingest cucumber --file <path>`
- SARIF: `bottleneck ingest sarif --file <path>`
- generic test summary: `bottleneck ingest test-summary --file <path>`
- telemetry: `bottleneck ingest telemetry --file <path>`

For JUnit and coverage, include a suggested command even if the ingest implementation comes later:

- `bottleneck ingest junit --file <path>`
- `bottleneck ingest coverage --file <path>`

These may be implemented in Prompt 02.

## Files likely to change

- `cmd/root.go`
- new `cmd/discover.go`
- new `cmd/discover_test.go`
- new `internal/discover/*`
- possibly `readme.md` only if command list needs a small update

## Implementation constraints

- Do not change current behavior of `scorecard`, `diagnose`, `validate`, `trace`, or existing `ingest` subcommands.
- Do not require GitHub access or network access.
- Do not parse large files deeply. Existence/path detection is enough for v1.
- Use filepath-safe and OS-portable logic.
- Ignore `.git`, `node_modules`, `vendor`, `.next`, `dist`, `build` except known report paths under build/test-results.
- Avoid failing the command if one directory cannot be read; report a warning in JSON or text output.

## Tests to add

Add tests that create temporary repo structures and verify discovery results.

Required tests:

1. Discovers Cucumber, SARIF, telemetry, README, and GitHub workflow files.
2. Returns useful suggested ingest commands for Cucumber, SARIF, and telemetry.
3. JSON output includes category, kind, path, confidence, and suggested command.
4. Missing evidence summary includes missing telemetry when none exists.
5. Does not traverse ignored directories such as `.git` or `node_modules`.
6. Existing command tests still pass.

## Acceptance criteria

- `go test ./...` passes.
- `bottleneck discover` works in a repo with no Bottleneck files and does not fail.
- `bottleneck discover --format json` emits valid JSON.
- Output clearly tells the user what was found and what to do next.

## Commit message suggestion

```text
Add evidence discovery command
```

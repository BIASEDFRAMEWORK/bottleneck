# Codex Prompt 05: Update docs and add low-overload command aliases

## Objective

Reduce onboarding overload by documenting the simplest day-one path and adding friendly command aliases without removing existing commands.

This prompt should be run after discovery, auto-ingest, assess, and explain-score exist.

## Required documentation updates

Rewrite or restructure the README and docs around three user journeys.

### Journey 1: Developer day one

Title:

```text
Get your first SDLC maturity assessment in 5 minutes
```

Commands:

```bash
bottleneck init --template saas
bottleneck assess
bottleneck trace BEHAVIOR-003
```

Explain:

```text
You do not need to understand the full BIASED model first. Start by seeing what Bottleneck finds, what it can prove, and where your SDLC is weakest.
```

### Journey 2: Add automated evidence

Title:

```text
Connect Bottleneck to evidence your team already produces
```

Commands:

```bash
bottleneck discover
bottleneck ingest --auto
bottleneck assess
bottleneck explain-score
```

Explain evidence sources:

- Cucumber
- JUnit
- coverage/lcov
- SARIF / CodeQL
- telemetry JSON
- GitHub Actions workflows

### Journey 3: Use Bottleneck in CI/CD

Title:

```text
Make SDLC maturity visible in every pull request
```

Example:

```yaml
- name: Discover SDLC evidence
  run: bottleneck discover

- name: Ingest SDLC evidence
  run: bottleneck ingest --auto

- name: Assess SDLC maturity
  run: bottleneck assess --format json > bottleneck-assessment.json

- name: Publish Bottleneck summary
  run: bottleneck assess >> "$GITHUB_STEP_SUMMARY"
```

Include guidance that CI gate behavior should be introduced after teams trust the score.

## Positioning copy

Update the README introduction to position Bottleneck as an SDLC maturity tool.

Suggested copy:

```text
Bottleneck is a CLI for measuring SDLC maturity from local engineering evidence.

It helps teams understand what makes releases hard: missing behavior proof, weak traceability, stale security evidence, poor telemetry, or unclear intent.

The more mature your SDLC is, the easier your releases become. The easier your releases become, the safer it is to enable AI-assisted development.
```

Use this core sentence prominently:

```text
Bottleneck tells you what is blocking release confidence and what evidence proves it.
```

## Command aliases

Add friendly aliases without removing existing commands.

### Required aliases

Add:

```bash
bottleneck check
```

Alias for:

```bash
bottleneck validate
```

Add:

```bash
bottleneck evidence sync
```

Alias for:

```bash
bottleneck ingest --auto
```

### Optional alias

If simple to implement, add:

```bash
bottleneck maturity
```

Alias for:

```bash
bottleneck assess
```

## Help text updates

Update root command help so the start-here path is:

```text
Start here:
  bottleneck init --template saas
  bottleneck assess
  bottleneck trace BEHAVIOR-003
```

Common commands should prioritize:

```text
assess          Show SDLC maturity, AI readiness, release friction, and primary bottleneck.
discover        Find evidence sources in the repo.
evidence sync   Automatically ingest discovered evidence.
trace           Follow one intent, behavior, or evidence ID end-to-end.
explain-score   Show why Bottleneck produced the current maturity score.
report          Generate a leadership-ready SDLC evidence report.
```

Advanced commands can include:

```text
validate, scorecard, diagnose, ingest, snapshot, seed-history, trends, explain
```

## Documentation structure recommendation

Create or update:

```text
readme.md
docs/quickstart-saas.md
docs/automated-evidence.md
docs/scoring-trust.md
docs/github-actions.md
```

Do not over-document every internal detail. Prioritize user journeys.

## Scoring trust documentation

Create `docs/scoring-trust.md` explaining:

- Maturity level
- Release recommendation
- Primary bottleneck
- Score confidence
- AI readiness
- Evidence provenance
- Tool-generated vs manual vs inferred evidence

Include this framing:

```text
Bottleneck does not ask teams to trust a black-box score. Every score can be explained with bottleneck explain-score.
```

## Implementation constraints

- Keep old commands and examples valid.
- Do not remove useful existing docs unless replacing with equivalent or clearer content.
- Update tests that assert README content.
- Keep examples runnable.
- Avoid making docs too theoretical. Lead with commands and examples.

## Tests to add or update

Required tests:

1. README test verifies the day-one path includes `init --template saas`, `assess`, and `trace`.
2. Root help test verifies `assess`, `discover`, `evidence sync`, and `explain-score` appear.
3. `bottleneck check` behaves like `validate`.
4. `bottleneck evidence sync` behaves like `ingest --auto`.
5. Optional: `bottleneck maturity` behaves like `assess` if implemented.
6. `go test ./...` passes.

## Acceptance criteria

- New README path is understandable in under five minutes.
- Existing users can still use current commands.
- Friendly aliases reduce cognitive load without changing internal architecture.
- Docs explain scoring trust plainly.

## Commit message suggestion

```text
Simplify onboarding docs and command aliases
```

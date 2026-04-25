# BIASED CLI

BIASED is a Go CLI for enforcing the BIASED SDLC framework through local, Git-based artifacts.

BIASED stands for:

- Behavior
- Intent
- Design
- Assurance
- Security
- Execution

The CLI answers one question:

> Is the system valid?

It does not run tests, execute BDD frameworks, call external services, or store data in a database. It validates the artifacts and outcomes produced by the team and by external tools.

## Purpose

BIASED defines how software systems are designed, validated, secured, and operated in the age of AI. It replaces ceremony-driven delivery with a capability-driven system that is measurable, enforceable, and observable.

The framework treats a system as invalid if any required capability artifact is missing, malformed, or below threshold.

## Core Model

```text
Behavior -> Intent -> Design -> Assurance -> Security -> Execution -> Behavior
```

Execution reveals truth. Truth updates Behavior and Intent.

## CLI Commands

### `biased init`

Creates the local BIASED artifact structure:

```text
biased/
  behavior/behavior-spec.md
  intent/intent.md
  design/architecture.md
  assurance/features/sample.feature
  assurance/results.json
  security/guardrails.json
  execution/telemetry.json
  docs/validation.md
```

### `biased validate`

Validates all six capabilities and prints system status.

Select thresholds with an environment:

```sh
biased validate --env=production
```

If `--env` is omitted, `default` is used. Environment-specific config inherits missing values from `default`.

Example:

```text
Behavior: PASS
Intent: PASS
Design: PASS
Assurance: PASS
Security: PASS
Execution: PASS

System Status: PASS
Primary Bottleneck: None
Environment: default
```

Exit behavior:

- Exit `0` when all capabilities are `PASS` or `WARNING`
- Exit `1` when any capability is `FAIL`

### `biased status`

Checks whether the local `biased/` artifact directory exists.

Output:

```text
System initialized
```

or:

```text
System incomplete
```

## Validation Rules

### Behavior

Artifact: `biased/behavior/behavior-spec.md`

Must exist, must not be empty, and must contain:

- `## Expected Behavior`
- `## Unacceptable Behavior`

### Intent

Artifact: `biased/intent/intent.md`

Must exist and contain:

- `## Outcomes`
- `## Constraints`
- `## Success Criteria`

### Design

Artifact: `biased/design/architecture.md`

Must exist, must not be empty, and must contain at least one Markdown section header.

### Assurance

Artifact: `biased/assurance/results.json`

BIASED does not run tests. BDD tests are executed externally by tools such as Cucumber, SpecFlow, Cucumber.js, or Behave. BIASED validates their output.

Required schema:

```json
{
  "scenarios_total": 1,
  "scenarios_passed": 1,
  "scenarios_failed": 0,
  "failures": []
}
```

BIASED computes `accuracy` as `scenarios_passed / scenarios_total`. Fails when JSON is invalid, required fields are missing, `scenarios_failed` exceeds the configured `max_failures`, or calculated accuracy is below the configured `min_accuracy`.

### Configuration

Artifact: `biased/config.yaml`

Defines inherited environment thresholds:

```yaml
environments:
  default:
    assurance:
      min_accuracy: 0.90
      max_failures: 0
    execution:
      max_error_rate: 0.05
      min_adoption: 0.5
  production:
    assurance:
      min_accuracy: 0.95
```

### Security

Artifact: `biased/security/guardrails.json`

Required schema:

```json
{
  "violations": 0
}
```

Fails when `violations > 0`.

### Execution

Artifact: `biased/execution/telemetry.json`

Required schema:

```json
{
  "adoption_rate": 0.9,
  "error_rate": 0.01
}
```

Fails when `error_rate` exceeds the configured `max_error_rate`. Returns `WARNING` when `adoption_rate` is below the configured `min_adoption`.

## Continuous Validation Loops

BIASED treats release as the beginning of continuous validation, not the end of delivery.

Core loops:

- Assurance Loop: validates behavior against expectations over time.
- Execution Loop: uses real-world telemetry to refine behavior and intent.
- Cost Loop: treats cost as a continuous operating constraint.

> Shipping is not completion. Continuous validation is completion.

## Development

Build the CLI:

```sh
go build ./...
```

Run locally:

```sh
go run . init
go run . validate
go run . status
```

## Documentation

Additional framework and validation references:

- `BIASED-strategy-v1.md`
- `biased/docs/validation.md`
- `GettingStarted.md`

# Getting Started With bottleneck

This guide helps you initialize and validate a project with the `bottleneck` CLI.

## Prerequisites

- Go installed
- A terminal
- This repository cloned or available locally

Check Go:

```sh
go version
```

## 1. Build Or Run The CLI

From the project root:

```sh
go build ./...
```

For local development, you can run commands without installing the binary:

```sh
go run . status
go run . init
go run . validate
```

## 2. Initialize Framework Artifacts

Run:

```sh
go run . init
```

This creates:

```text
bottleneck/
  config.yaml
  behavior/
  intent/
  design/
  assurance/
    features/
    results.json
  security/
  execution/
  docs/
```

It also creates starter artifacts for each capability.

## 3. Fill Out Capability Artifacts

### Behavior

Edit:

```text
bottleneck/behavior/behavior-spec.md
```

Required sections:

```md
## Expected Behavior
## Unacceptable Behavior
```

Use this file to describe what the system must do and what it must never do.

### Intent

Edit:

```text
bottleneck/intent/intent.md
```

Required sections:

```md
## Outcomes
## Constraints
## Success Criteria
```

Use this file to define why the system exists and how success is measured.

### Design

Edit:

```text
bottleneck/design/architecture.md
```

The file must not be empty and must contain at least one Markdown heading.

### Assurance

BDD scenarios live in:

```text
bottleneck/assurance/features/
```

bottleneck does not run these tests. Run them with your preferred BDD framework, then write the result summary to:

```text
bottleneck/assurance/results.json
```

Expected schema:

```json
{
  "scenarios_total": 1,
  "scenarios_passed": 1,
  "scenarios_failed": 0,
  "failures": []
}
```

bottleneck computes accuracy automatically as `scenarios_passed / scenarios_total`.

### Configuration

Environment thresholds live in:

```text
bottleneck/config.yaml
```

Use the default environment or select one explicitly:

```sh
go run . validate --env=dev
go run . validate --env=production
```

Environment-specific values inherit missing thresholds from `default`.

Current built-in environments:

- `default`
- `dev`
- `test`
- `stage`
- `production`

### Security

Edit:

```text
bottleneck/security/guardrails.json
```

Expected schema:

```json
{
  "violations": 0
}
```

### Execution

Edit:

```text
bottleneck/execution/telemetry.json
```

Expected schema:

```json
{
  "adoption_rate": 0.9,
  "error_rate": 0.01
}
```

## 4. Validate The System

Run:

```sh
go run . validate
```

Or validate against a stricter environment:

```sh
go run . validate --env=production
```

Passing output:

```text
Behavior: PASS
Intent: PASS
Design: PASS
Assurance: PASS
  accuracy: 1.00 (threshold: 0.90)
  scenarios_failed: 0 (allowed: 0)
Security: PASS
Execution: PASS

System Status: PASS
Primary Bottleneck: None
Environment: default
```

Failing output example:

```text
Behavior: PASS
Intent: PASS
Design: PASS
Assurance: FAIL (accuracy below threshold)
  accuracy: 0.90 (threshold: 0.95)
  scenarios_failed: 0 (allowed: 0)
Security: PASS
Execution: PASS

System Status: FAIL
Primary Bottleneck: Assurance
Environment: production
```

## 5. Check Initialization Status

Run:

```sh
go run . status
```

Output:

```text
System initialized
```

or:

```text
System incomplete
```

## 6. Explain The Current State

Run:

```sh
go run . explain
```

Or focus on one capability:

```sh
go run . explain --env=production --capability=Assurance
```

Use `explain` when you want:

- ownership
- mapped bottlenecks
- evidence from validation
- recommended next actions

## 7. Generate A Scorecard

Run:

```sh
go run . scorecard
```

For machine-readable output:

```sh
go run . scorecard --env=production --format=json
```

Use `scorecard` when you want a compact summary of:

- environment
- system status
- primary bottleneck
- owner per capability
- bottleneck per capability
- evidence per capability

## Recommended Workflow

1. Run `bottleneck init`.
2. Define Intent, Behavior, and Design artifacts.
3. Write BDD scenarios under `bottleneck/assurance/features/`.
4. Run your external BDD framework.
5. Export BDD results to `bottleneck/assurance/results.json`.
6. Select the target environment in `bottleneck/config.yaml` or with `--env`.
7. Update security and execution artifacts.
8. Run `bottleneck validate`.
9. Run `bottleneck explain` or `bottleneck scorecard` to understand ownership and bottlenecks.
10. Fix the primary bottleneck if the system fails.

## Principle

bottleneck validates outcomes, not implementation.

It enforces:

> Is the system valid?

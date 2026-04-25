# Getting Started With BIASED

This guide helps you initialize and validate a project with the `biased` CLI.

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

## 2. Initialize BIASED Artifacts

Run:

```sh
go run . init
```

This creates:

```text
biased/
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
biased/behavior/behavior-spec.md
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
biased/intent/intent.md
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
biased/design/architecture.md
```

The file must not be empty and must contain at least one Markdown heading.

### Assurance

BDD scenarios live in:

```text
biased/assurance/features/
```

BIASED does not run these tests. Run them with your preferred BDD framework, then write the result summary to:

```text
biased/assurance/results.json
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

BIASED computes accuracy automatically as `scenarios_passed / scenarios_total`.

### Configuration

Environment thresholds live in:

```text
biased/config.yaml
```

Use the default environment or select one explicitly:

```sh
go run . validate --env=dev
go run . validate --env=production
```

Environment-specific values inherit missing thresholds from `default`.

### Security

Edit:

```text
biased/security/guardrails.json
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
biased/execution/telemetry.json
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

Passing output:

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

Failing output example:

```text
Behavior: PASS
Intent: PASS
Design: PASS
Assurance: FAIL (scenarios_failed > 0)
Security: PASS
Execution: WARNING (low adoption)

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

## Recommended Workflow

1. Run `biased init`.
2. Define Intent, Behavior, and Design artifacts.
3. Write BDD scenarios under `biased/assurance/features/`.
4. Run your external BDD framework.
5. Export BDD results to `biased/assurance/results.json`.
6. Update security and execution artifacts.
7. Run `biased validate`.
8. Fix the primary bottleneck if the system fails.

## Principle

BIASED validates outcomes, not implementation.

It enforces:

> Is the system valid?

# SaaS Billing Bottleneck Example

This example shows Bottleneck evaluating a Subscription Billing Release for a SaaS product. It is small enough to inspect by hand and realistic enough to demo payment-method updates, failed invoice retry, duplicate-charge prevention, test evidence, security scan evidence, and telemetry.

## What Is Intentionally Broken

`BEHAVIOR-003` proves that payment retry does not create duplicate charges. The initial `bottleneck/` evidence describes the behavior and links it to intent, design, security, and execution telemetry, but it intentionally omits mapped assurance evidence for `BEHAVIOR-003`.

That means the first diagnosis should identify Assurance as the primary bottleneck. After ingesting the sample Cucumber report, Bottleneck writes `ASSURANCE-003` and the missing assurance gap is removed.

## Run The First Scorecard

```sh
cd examples/saas-billing
bottleneck validate
bottleneck scorecard
```

Expected first-run posture:

- Release recommendation is conditional.
- Primary bottleneck is Assurance.
- The reason names `BEHAVIOR-003` or payment retry test coverage.

## Diagnose The Bottleneck

```sh
bottleneck diagnose
```

The diagnosis explains why the missing payment retry assurance matters and suggests inspecting `BEHAVIOR-003`.

## Trace BEHAVIOR-003

```sh
bottleneck trace BEHAVIOR-003
```

The trace shows the duplicate-charge prevention behavior and the missing mapped test evidence before ingestion.

## Ingest Sample Evidence

```sh
bottleneck ingest cucumber --file reports/cucumber.json
bottleneck ingest sarif --file reports/codeql.sarif
```

The Cucumber report includes:

```text
@BEHAVIOR-003
Scenario: Payment retry uses idempotency to avoid duplicate charges
```

After ingestion, `bottleneck/assurance/results.json` includes `ASSURANCE-003` mapped to `BEHAVIOR-003`. The SARIF report writes low-risk security evidence linked to `BEHAVIOR-003`.

## Re-run The Production Scorecard

```sh
bottleneck scorecard --env=production
```

The production scorecard should improve after ingestion because the duplicate-charge prevention behavior now has mapped assurance evidence.

## Run In GitHub Actions

This example includes `.github/workflows/bottleneck.yml`, which mirrors the Day-One workflow pattern: validate evidence, write a Markdown scorecard to the step summary, emit diagnosis annotations, and run the release gate.

To use it from this repository, copy or adapt the workflow into `.github/workflows/`. To use it in another repository, ensure the `bottleneck` CLI is available before the validation steps.

## File Map

```text
examples/saas-billing/
  bottleneck/
    config.yaml
    intent/intent.md
    behavior/behavior-spec.md
    design/architecture.md
    assurance/results.json
    security/guardrails.json
    execution/telemetry.json
  reports/
    cucumber.json
    codeql.sarif
    test-summary.json
    telemetry.json
  .github/workflows/bottleneck.yml
```

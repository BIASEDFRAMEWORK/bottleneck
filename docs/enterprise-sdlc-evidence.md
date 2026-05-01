# Enterprise SDLC Evidence with Bottleneck

## What Bottleneck Does

Bottleneck is a local-first evidence system for understanding SDLC bottlenecks. It reads evidence files in a repository, validates whether the evidence is complete enough to support a release decision, renders a scorecard, captures scorecard snapshots, analyzes local history, explains the weakest evidence categories, and generates a leadership-ready report.

Bottleneck helps different audiences use the same evidence:

- Developers get a structured way to show what evidence exists and what is missing.
- Tech leads get a way to prioritize systemic fixes instead of treating every release issue as a one-off task.
- Product leaders get a way to connect release intent to behavior, validation, and adoption evidence.
- Security teams get evidence of guardrails, findings, and accepted risk.
- Leadership gets a decision-ready summary of the current constraint, trend direction, risk, owners, and next actions.

Reports are generated from evidence, not opinion. The command output is deterministic and local to the repository.

## What Bottleneck Does Not Do

Bottleneck is not a dashboard.
Bottleneck is not Jira replacement.
Bottleneck is not a project-management system.
Bottleneck is not a database-backed metrics warehouse.
Bottleneck is a local-first evidence system for understanding SDLC bottlenecks.

Bottleneck does not replace CI, security scanners, observability platforms, product analytics, release management, or engineering judgment. It connects those signals when they are exported into local evidence files. It does not require a SaaS backend, external storage, a hosted service, a dashboard, or a project-management integration; no external storage is required.

## Why Git Is the System of Record

Bottleneck treats local files as the source of evidence. The repository can contain intent, behavior, design, assurance, security, execution, configuration, snapshots, and generated reports.

Git is the system of record because it versions:

- evidence files under `bottleneck/`
- scorecard snapshots under `bottleneck/history/scorecards/`
- optional report artifacts under `bottleneck/reports/`
- the workflow and threshold changes that explain why a release decision changed

Trends are derived from committed or local snapshot history. No external storage is required. No dashboard is required. Teams can review evidence changes in pull requests, commit snapshot history intentionally, and compare delivery-system posture over time using normal Git review practices.

CI can generate snapshots and reports as artifacts without committing them. That is useful for PR and release review. Long-lived trend history across runs requires either artifact retention or intentionally committed snapshot files.

## Recommended Team Workflow

Start with local evidence and then build history over delivery cycles:

```sh
bottleneck init --template saas
bottleneck ingest cucumber --file reports/cucumber.json
bottleneck ingest sarif --file reports/codeql.sarif
bottleneck ingest telemetry --file reports/telemetry.json
bottleneck scorecard
bottleneck snapshot
bottleneck trends
bottleneck explain
bottleneck report
```

Command roles:

- `bottleneck init --template saas` creates a starter evidence layout with intent, behavior, design, assurance, security, execution, and configuration files.
- `bottleneck ingest cucumber --file reports/cucumber.json` converts Cucumber evidence into normalized assurance evidence.
- `bottleneck ingest sarif --file reports/codeql.sarif` converts SARIF or CodeQL findings into security evidence.
- `bottleneck ingest telemetry --file reports/telemetry.json` converts delivery, reliability, adoption, and operational signals into execution evidence.
- `bottleneck scorecard` summarizes release readiness, primary bottleneck, category status, and next action.
- `bottleneck snapshot` writes the current scorecard as a local JSON snapshot for trend history.
- `bottleneck trends` compares local snapshots and identifies improving, declining, recovered, regressed, stable, and persistent bottleneck signals.
- `bottleneck explain` explains how evidence affected category scores and what to inspect next.
- `bottleneck report` generates a leadership-ready SDLC evidence report from the current scorecard, trend history, and explanations.

Use `bottleneck seed-history` when you need deterministic demo history before a team has collected real delivery-cycle snapshots:

```sh
bottleneck seed-history
bottleneck trends
bottleneck report
```

## Recommended CI Workflow

A CI workflow can generate evidence artifacts for the current run:

```sh
bottleneck scorecard --env=production --format=json
bottleneck snapshot --env=production --label=ci
bottleneck trends --env=production --window=6
bottleneck report --env=production --format=markdown
```

The copyable GitHub Actions example is:

```text
examples/github-actions/bottleneck-evidence-report.yml
```

That workflow builds Bottleneck, runs validation, renders a Markdown scorecard, creates a CI snapshot, generates a trend summary, generates an SDLC evidence report, and uploads `bottleneck/history/` and `bottleneck/reports/` as artifacts.

CI-generated artifacts and committed snapshots serve different purposes:

- CI-generated artifacts show evidence for a specific workflow run. They are useful for PR and release review, but they disappear unless artifact retention is configured.
- Committed snapshots create Git-tracked trend history. Use this when leadership or teams need a durable view of SDLC evidence direction across delivery cycles.

The CI workflow does not commit generated files automatically. If your team wants Git to be the long-term trend history, review and commit snapshot files intentionally.

## Snapshot History

`bottleneck snapshot` writes the current scorecard to:

```text
bottleneck/history/scorecards/
```

It also updates the latest snapshot for the selected environment under:

```text
bottleneck/history/latest/
```

Snapshots use the same stable scorecard JSON contract as `bottleneck scorecard --format=json`. They include schema metadata, environment, release recommendation, primary bottleneck, category status and score data, and supporting scorecard details.

A single snapshot captures current posture. Multiple snapshots show delivery-system direction.

## Trend Analysis

`bottleneck trends` reads local snapshot files, filters them by environment, sorts them by `snapshot.created_at`, and analyzes the latest window.

Trend output can show:

- improving trends, where category scores are moving up overall
- declining trends, where category scores are moving down overall
- stable trends, where posture is not materially changing
- recovered categories, where a category moved from weak status to passing status
- regressed categories, where a category moved from passing status to warning or failing status
- persistent bottleneck, where the same category repeatedly appears as the primary constraint

If only one snapshot exists, trend output reports insufficient history and recommends creating snapshots over multiple delivery cycles.

## Evidence Explanation

`bottleneck explain` turns validation and scorecard data into evidence-backed explanations. It is useful when the team needs to understand why a category scored the way it did, what evidence was found, what evidence is missing, who should act, and what automation could prevent the issue from recurring.

Examples:

```sh
bottleneck explain
bottleneck explain --category=assurance
bottleneck explain --format=markdown --out=bottleneck/reports/bottleneck-explanation.md
```

Use explain output in engineering reviews when the scorecard identifies a weak category but the next action needs more context.

## Leadership Report

`bottleneck report` generates an SDLC evidence report from local evidence, the current scorecard, trend analysis, and explanations.

Examples:

```sh
bottleneck report
bottleneck report --format=markdown --out=bottleneck/reports/sdlc-evidence-report.md
bottleneck report --format=json
```

The report includes executive summary, current delivery-system status, primary bottleneck, category scorecard, trend summary, evidence found, evidence missing, risk to delivery, recommended actions, suggested owners, suggested automation, leadership decision needed, and snapshot metadata.

The leadership report is useful when leaders need to decide whether to slow down, invest in evidence, accept risk, block release, or continue with current controls.

## How to Interpret Results

Release recommendation values:

- `Proceed`: evidence is strong enough for the configured release decision.
- `Conditional`: the system has warnings or incomplete evidence that should be reviewed before release.
- `Block`: one or more evidence categories fail or release policy is not satisfied. Some teams may refer to this as blocked.
- `Unknown`: required scorecard evidence is unavailable or not assessed. Treat this as insufficient evidence until the missing artifacts are added.

Other interpretation terms:

- Insufficient Evidence: there is not enough evidence or history to support a confident decision. This can appear as missing evidence, `Unknown`, or insufficient trend history.
- Improving trends: the evidence posture is getting stronger across the selected snapshot window.
- Declining trends: the evidence posture is getting weaker and leadership should review the drivers before accelerating release.
- Stable trends: the evidence posture is not moving materially; this can be acceptable when healthy, or a concern when a bottleneck persists.
- Recovered categories: a previously weak category moved back to passing status.
- Regressed categories: a passing category moved to warning or failing status.
- Persistent bottleneck: the same category is repeatedly the primary constraint across snapshots.

Use scorecards for current posture, trends for direction, explain output for causes, and reports for decision-ready summaries.

## Common Enterprise Scenarios

### 1. Team is shipping quickly but validation is weak

Signal: `scorecard` or `report` identifies Assurance as the primary bottleneck, missing mapped test evidence, or weak validation score.

Run:

```sh
bottleneck scorecard
bottleneck explain --category=assurance
bottleneck report
```

Leadership decision: approve time to map critical behaviors to validation evidence before accelerating release.

### 2. Security is repeatedly discovered late

Signal: Security appears as the primary bottleneck, SARIF findings exceed configured thresholds, or trends show Security regressing across snapshots.

Run:

```sh
bottleneck ingest sarif --file reports/codeql.sarif
bottleneck scorecard --env=production
bottleneck trends --env=production --window=6
```

Leadership decision: decide whether release should be blocked until high-severity findings are resolved or formally accepted.

### 3. Requirements are implemented but not traceable

Signal: Traceability warnings show behavior IDs without assurance, security, or execution evidence; `trace` shows missing links.

Run:

```sh
bottleneck trace BEHAVIOR-003
bottleneck explain
bottleneck scorecard --details
```

Leadership decision: require evidence ID and refs repair before approving release expansion.

### 4. Production behavior does not match expected behavior

Signal: Execution evidence warns or fails, telemetry violates thresholds, or execution trends decline after behavior and assurance evidence had been passing.

Run:

```sh
bottleneck ingest telemetry --file reports/telemetry.json
bottleneck scorecard
bottleneck trends
```

Leadership decision: prioritize operational readiness, adoption feedback, and telemetry review before scaling usage.

### 5. Leadership wants faster delivery but the SDLC evidence is declining

Signal: `trends` reports declining or regressed categories, and `report` shows the persistent bottleneck and risk to delivery.

Run:

```sh
bottleneck trends --window=6
bottleneck report
```

Leadership decision: decide whether to invest in the persistent bottleneck, adjust release controls, or formally accept increased delivery risk.

## Frequently Asked Questions

### Does Bottleneck need a database?

No. Bottleneck reads local files and writes local snapshots and reports. Git can version those files when a team wants durable trend history.

### Does Bottleneck replace Jira?

No. Bottleneck is not a Jira replacement and not a project-management system. It can inform planning discussions by showing missing evidence and delivery constraints.

### Do we need to commit snapshots?

Not for local analysis or CI artifacts. Commit snapshots intentionally when you want Git to preserve long-term trend history across delivery cycles.

### Can CI generate reports without committing them?

Yes. The GitHub Actions example generates a snapshot, trend summary, and report, then uploads `bottleneck/history/` and `bottleneck/reports/` as artifacts. It does not auto-commit generated files.

### What if we only have one snapshot?

Trend analysis reports insufficient history. Create snapshots over multiple delivery cycles or use `bottleneck seed-history` for a deterministic demo.

### What if evidence is missing?

`scorecard`, `explain`, and `report` identify missing evidence and recommended actions. Add or ingest the missing evidence, then rerun the scorecard.

### How do teams use this in leadership reviews?

Use the scorecard to show current release posture, trends to show direction, explain output to show root evidence gaps, and the leadership report to summarize risk, actions, owners, automation, and the decision needed.

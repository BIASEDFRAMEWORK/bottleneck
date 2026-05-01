# AI Implementation Prompt: Enterprise SDLC Evidence Documentation

You are working in the Bottleneck Go CLI codebase.

Implement **Implementation Epic 8: Documentation for enterprise usage** from the `Enterprise SDLC Evidence Package` milestone.

## Milestone Context

Bottleneck should be understandable to enterprise teams as a local-first SDLC evidence system. It should explain current evidence, historical trends, bottlenecks, reports, and leadership decisions without requiring a database, SaaS backend, dashboard, or project-management system.

This epic is documentation-focused.

## Scope

Add:

```text
docs/enterprise-sdlc-evidence.md
```

Link it from:

```text
README or readme.md
```

Use the repository's actual README filename and style.

## Current Files To Inspect

Read before changing docs:

- `readme.md` or `README.md`
- existing files under `docs/`
- task list for this milestone
- existing command docs for `scorecard`, `snapshot`, `trends`, `explain`, `report`, and `seed-history`
- examples under `examples/`
- GitHub Actions examples if implemented

Do not document commands as fully available if they are not implemented, unless the documentation is explicitly for the completed milestone and clearly describes the expected workflow.

## Required Document

Create:

```markdown
# Enterprise SDLC Evidence with Bottleneck

## What Bottleneck Does
## What Bottleneck Does Not Do
## Why Git Is the System of Record
## Recommended Team Workflow
## Recommended CI Workflow
## Snapshot History
## Trend Analysis
## Evidence Explanation
## Leadership Report
## How to Interpret Results
## Common Enterprise Scenarios
## Frequently Asked Questions
```

## Positioning Requirements

Include this positioning clearly:

```text
Bottleneck is not a dashboard.
Bottleneck is not Jira replacement.
Bottleneck is not a project-management system.
Bottleneck is not a database-backed metrics warehouse.
Bottleneck is a local-first evidence system for understanding SDLC bottlenecks.
```

Explain:

- local files are the source of evidence
- Git versions snapshots and report artifacts
- trends are derived from committed or local snapshot history
- no external storage is required
- no dashboard is required
- reports are generated from evidence, not opinion

## Explain Team Value

Include practical value for different audiences:

- Developers get a structured way to show missing evidence.
- Tech leads get a way to prioritize systemic fixes.
- Product leaders get a way to connect intent to delivery.
- Security teams get evidence of guardrails and risk.
- Leadership gets a decision-ready summary.

## Recommended Team Workflow

Document a local workflow:

```sh
bottleneck init --template saas
bottleneck ingest --type cucumber --file reports/cucumber.json
bottleneck ingest --type sarif --file reports/codeql.sarif
bottleneck ingest --type telemetry --file reports/telemetry.json
bottleneck scorecard
bottleneck snapshot
bottleneck trends
bottleneck explain
bottleneck report
```

If actual ingestion syntax differs, use implemented command syntax. Do not invent a syntax that the CLI does not support.

Explain what each command does in one or two sentences.

## Recommended CI Workflow

Document a CI workflow:

```sh
bottleneck scorecard --env=production --format=json
bottleneck snapshot --env=production --label=ci
bottleneck trends --env=production --window=6
bottleneck report --env=production --format=markdown
```

If the GitHub Actions example exists, link to:

```text
examples/github-actions/bottleneck-evidence-report.yml
```

Explain the difference between:

- CI-generated artifacts uploaded from a workflow
- snapshots intentionally committed to Git for long-term trend history

## Common Enterprise Scenarios

Include these scenarios:

1. Team is shipping quickly but validation is weak.
2. Security is repeatedly discovered late.
3. Requirements are implemented but not traceable.
4. Production behavior does not match expected behavior.
5. Leadership wants faster delivery but the SDLC evidence is declining.

For each scenario, explain:

- What Bottleneck signal reveals the issue.
- Which command to run.
- What leadership decision might be needed.

## How To Interpret Results

Explain:

- `Proceed`
- `Conditional`
- `Blocked`
- `Insufficient Evidence`
- improving trends
- declining trends
- stable trends
- recovered categories
- regressed categories
- persistent bottleneck

If recommendation values are not yet standardized in code, avoid overpromising exact values or clearly say "when release recommendation standardization is implemented."

## Frequently Asked Questions

Include practical FAQs:

- Does Bottleneck need a database?
- Does Bottleneck replace Jira?
- Do we need to commit snapshots?
- Can CI generate reports without committing them?
- What if we only have one snapshot?
- What if evidence is missing?
- How do teams use this in leadership reviews?

## Acceptance Criteria

- Documentation explains snapshot.
- Documentation explains trends.
- Documentation explains report.
- Documentation explains no database requirement.
- Documentation includes local workflow.
- Documentation includes CI workflow.
- Documentation includes enterprise interpretation examples.
- Documentation is linked from README.

## Tests To Add Or Update

If the repository has documentation tests, add or update tests to verify:

- `docs/enterprise-sdlc-evidence.md` exists.
- README links to it.
- Documentation mentions no database.
- Documentation mentions snapshots.
- Documentation mentions trends.
- Documentation mentions reports.
- Documentation mentions Git as system of record.
- Documentation includes local workflow.
- Documentation includes CI workflow.

If no docs tests exist, add one small test only if that is consistent with local testing style. Do not introduce a heavy docs framework.

## Tone

Keep tone practical and enterprise-friendly.

Prefer:

- evidence-based language
- decision language
- local-first positioning
- clear command examples
- explicit non-goals

Avoid:

- marketing fluff
- unexplained framework jargon
- implying Bottleneck is a dashboard or PM system
- claiming unsupported integrations
- shaming language about teams or process

## Verification

Run:

```sh
go test ./...
```

If docs-only and no doc tests exist, still run tests if feasible to ensure no unrelated breakage.

Manually inspect:

```sh
sed -n '1,260p' docs/enterprise-sdlc-evidence.md
```

## Final Response Requirements

When finished, report:

1. Documentation file added.
2. README link added.
3. Key sections included.
4. Any docs tests added or changed.
5. Exact commands run and results.
6. Any acceptance criteria intentionally deferred and why.


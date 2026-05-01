# AI Implementation Prompt: Package A Demo-Ready Example

You are working in the `bottleneck` Go CLI repository.

Implement the tenth slice of the **SaaS Team Day-One Success** milestone.

## Milestone Goal

Make Bottleneck usable by a SaaS engineering team in the first 10 minutes.

A developer should be able to inspect a complete SaaS billing example, run Bottleneck against it, see an intentional release-readiness bottleneck, ingest sample evidence, and observe the bottleneck improve.

This implementation prompt is scoped to:

- Epic 10: Package A Demo-Ready Example
- Task 10.1: Add `examples/saas-billing`
- Task 10.2: Add example walkthrough
- Task 10.3: Add intentional failure scenario

## Definition Of Done For This Slice

A developer can run:

```sh
cd examples/saas-billing
bottleneck validate
bottleneck scorecard
bottleneck diagnose
bottleneck trace BEHAVIOR-003
bottleneck ingest cucumber --file reports/cucumber.json
bottleneck ingest sarif --file reports/codeql.sarif
bottleneck scorecard --env=production
```

And understand:

- What SaaS capability Bottleneck is checking.
- Which evidence exists.
- Which evidence is intentionally missing.
- Why Assurance is the primary bottleneck.
- How ingestion improves the bottleneck.
- How the same flow can run in GitHub Actions.

The example should be small enough to inspect, realistic enough to demo, and safe to copy into another repository.

## Current Architecture To Respect

Inspect the repository before changing files. Follow existing patterns for:

- examples under `examples/`
- SaaS starter template artifacts, if implemented
- sample ingestion files under `examples/saas/reports/`, if implemented
- GitHub Actions examples under `examples/github-actions/`, if implemented
- validator, scorecard, diagnosis, traceability, and ingestion tests
- fixture layout under `internal/**/testdata`

Preserve backwards compatibility:

- Do not rename public commands or flags.
- Do not change normalized evidence schemas just for the example.
- Do not make ingestion behavior weaker to make the demo pass.
- Do not overwrite or relocate existing examples unless they are clearly superseded and tests are updated.
- Do not document commands that are not implemented.
- Do not require secrets, network services, or external SaaS accounts to run the example.

Prefer adding a self-contained example and tests. Only change product behavior if the example exposes a clear bug in an already implemented feature.

## Primary Source Of Truth

Read:

- `tasks/DayOneExperience/saas-team-day-one-success-task-list.md`
- `readme.md`
- `docs/quickstart-saas.md`, if it exists
- existing `examples/`
- `cmd/init.go`
- `cmd/validate.go`
- `cmd/scorecard.go`
- `cmd/diagnose.go`
- `cmd/trace.go`
- `cmd/ingest*.go`
- `internal/scorecard/*`
- `internal/diagnosis/*`
- `internal/traceability/*`
- `internal/ingest/*`
- existing complete SaaS fixtures, if any

Keep the demo domain consistent with the Day-One milestone:

- Subscription Billing Release
- Users can update payment methods
- Users can retry failed invoices
- Payment retry behavior must avoid duplicate charges
- Payment details must not be exposed or stored directly
- `BEHAVIOR-003` is the intentional assurance gap before ingestion

## Epic 10: Package A Demo-Ready Example

### Task 10.1 - Add `examples/saas-billing`

Goal: Provide a complete example users can inspect and copy.

Create:

```text
examples/saas-billing/
  README.md
  bottleneck/
    config.yaml
    intent/intent.md
    behavior/behavior-spec.md
    design/architecture.md
    assurance/
    security/
    execution/
  reports/
    cucumber.json
    codeql.sarif
    test-summary.json
    telemetry.json
  .github/workflows/bottleneck.yml
```

Implementation guidance:

- Keep files small and readable.
- Use realistic SaaS billing language.
- Include stable evidence IDs:
  - `INTENT-001`
  - `BEHAVIOR-001`
  - `BEHAVIOR-002`
  - `BEHAVIOR-003`
  - `DESIGN-001`
  - `ASSURANCE-001`
  - `SECURITY-001`
  - `EXECUTION-001`
- Include enough evidence for `validate`, `scorecard`, `diagnose`, and `trace` to be useful.
- Intentionally leave assurance evidence missing for `BEHAVIOR-003` in the initial `bottleneck/` project.
- Keep sample report files under `reports/` able to fill or improve the missing evidence after ingestion, where ingestion commands are implemented.
- Include a workflow under `.github/workflows/bottleneck.yml` that mirrors the Day-One GitHub Actions example, if workflow support exists.
- Avoid duplicating large fixtures. Reuse content patterns from the SaaS starter or complete SaaS fixture where useful, but make this example independently understandable.

Recommended initial example posture:

- Intent and behavior are present.
- Design is present.
- Security evidence is acceptable or low-risk.
- Execution telemetry is present and healthy enough to avoid being the primary bottleneck.
- Assurance is intentionally incomplete for payment retry duplicate-charge prevention.
- `bottleneck diagnose` identifies Assurance as the primary bottleneck.

### Task 10.2 - Add Example Walkthrough

Goal: Let users reproduce the value quickly.

`examples/saas-billing/README.md` must include:

- `cd examples/saas-billing`
- `bottleneck validate`
- `bottleneck scorecard`
- `bottleneck diagnose`
- `bottleneck trace`
- `bottleneck ingest cucumber --file reports/cucumber.json`
- `bottleneck ingest sarif --file reports/codeql.sarif`
- `bottleneck scorecard --env=production`

README guidance:

- Start with what the example demonstrates.
- Explain the intentional failure before showing commands.
- Show the first-run workflow.
- Explain expected output at a high level.
- Show how to inspect the missing behavior evidence.
- Show how to ingest sample evidence.
- Show how to re-run the scorecard.
- Explain the GitHub Actions workflow.
- Keep examples copy/paste friendly.

Recommended README shape:

1. Title: `SaaS Billing Bottleneck Example`
2. Short purpose paragraph.
3. What is intentionally broken.
4. Run the first scorecard.
5. Diagnose the bottleneck.
6. Trace `BEHAVIOR-003`.
7. Ingest sample evidence.
8. Re-run production scorecard.
9. Run in GitHub Actions.
10. File map.

Recommended command block:

```sh
cd examples/saas-billing
bottleneck validate
bottleneck scorecard
bottleneck diagnose
bottleneck trace BEHAVIOR-003
bottleneck ingest cucumber --file reports/cucumber.json
bottleneck ingest sarif --file reports/codeql.sarif
bottleneck scorecard --env=production
```

Do not promise exact output that the CLI does not produce. Use short excerpts when needed.

### Task 10.3 - Add Intentional Failure Scenario

Goal: Show Bottleneck finding a real bottleneck.

Acceptance criteria:

- Example includes payment retry behavior.
- Intent and behavior are present.
- Assurance evidence is missing.
- `diagnose` identifies Assurance as the primary bottleneck.
- After ingesting test evidence, the bottleneck improves.

Implementation guidance:

- Make the failure specific:

```text
BEHAVIOR-003: Payment retry does not create duplicate charges.
```

- Initial behavior evidence should reference the intent.
- Initial assurance evidence should not cover `BEHAVIOR-003`.
- Cucumber report should include a passing scenario tagged with `@BEHAVIOR-003`.
- After `bottleneck ingest cucumber --file reports/cucumber.json`, normalized assurance evidence should include or map to `BEHAVIOR-003`.
- Scorecard or diagnosis should show improvement:
  - Assurance category moves from fail/block to warn/pass, or
  - missing `BEHAVIOR-003` assurance evidence is removed, or
  - primary bottleneck changes away from Assurance.
- If current ingestion cannot update the exact category yet, document the closest supported improvement and report the remaining gap.

Recommended Cucumber scenario:

```gherkin
@BEHAVIOR-003
Scenario: Payment retry uses idempotency to avoid duplicate charges
```

Do not fake the improvement in docs. The example and tests should prove it.

## Tests To Add Or Update

Add tests in the most appropriate package:

- example structure tests in a root or examples test file
- scorecard and diagnosis example tests in `cmd/*_test.go`, `internal/scorecard/*_test.go`, or `internal/diagnosis/*_test.go`
- ingestion smoke tests in `cmd/*_test.go` or `internal/ingest/*_test.go`
- workflow validation tests near other workflow tests, if workflow exists

Minimum test coverage:

- `examples/saas-billing/README.md` exists.
- Example README contains all required commands.
- Example project contains the required `bottleneck/` structure.
- Example contains `reports/cucumber.json`, `reports/codeql.sarif`, `reports/test-summary.json`, and `reports/telemetry.json`.
- Example includes payment retry behavior.
- Intent and behavior evidence are present.
- Initial diagnosis identifies Assurance as the primary bottleneck.
- Initial trace for `BEHAVIOR-003` shows missing assurance evidence.
- Cucumber ingestion against the sample report improves assurance coverage for `BEHAVIOR-003`, if cucumber ingestion is implemented.
- SARIF ingestion against the sample report succeeds, if SARIF ingestion is implemented.
- Production scorecard can be run against the example.
- Example workflow YAML exists and is valid, if workflow support exists.

If an ingestion command is not implemented, do not create a failing test that demands it unless this slice intentionally implements it. Document the deferred acceptance criterion in the final response.

## UX Requirements

The example should feel demo-ready.

Prefer:

- real SaaS billing names
- short files with obvious evidence IDs
- one intentional bottleneck
- copy/paste commands
- clear before/after ingestion flow
- no external dependencies
- no hidden secrets

Avoid:

- generic placeholder content
- multiple unrelated failures that obscure the demo
- giant report files
- exact-output docs that drift easily
- requiring users to understand the full BIASED framework
- examples that pass without demonstrating a bottleneck

## Documentation Updates

Update docs only where useful:

- README should link to `examples/saas-billing`.
- `docs/quickstart-saas.md`, if it exists, should point to the demo-ready example.
- If GitHub Actions docs exist, link to the example workflow.
- Keep the example README as the primary walkthrough for this slice.

## Verification Commands

Run:

```sh
go test ./...
```

If feasible, manually verify:

```sh
cd examples/saas-billing
bottleneck validate
bottleneck scorecard
bottleneck diagnose
bottleneck trace BEHAVIOR-003
bottleneck ingest cucumber --file reports/cucumber.json
bottleneck ingest sarif --file reports/codeql.sarif
bottleneck scorecard --env=production
```

For manual verification:

- Do not mutate tracked example files unless the ingestion flow is intentionally run on a copied temp version.
- Prefer copying `examples/saas-billing` to a temporary directory before running ingestion commands that write normalized evidence.
- Record whether commands pass or fail.
- If the initial example intentionally exits non-zero because Assurance is missing, document that as expected.

## Final Response Requirements

When finished, report:

1. Example files added.
2. Walkthrough commands documented.
3. Intentional failure behavior.
4. Before/after ingestion behavior.
5. Tests added or changed.
6. Documentation updates.
7. Exact commands run and results.
8. Any acceptance criteria intentionally deferred and why.


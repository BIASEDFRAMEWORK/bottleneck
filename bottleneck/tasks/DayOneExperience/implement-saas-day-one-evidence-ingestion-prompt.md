# AI Implementation Prompt: Simplify Evidence Ingestion For Common SaaS Teams

You are working in the `bottleneck` Go CLI repository.

Implement the fifth slice of the **SaaS Team Day-One Success** milestone.

## Milestone Goal

Make Bottleneck usable by a SaaS engineering team in the first 10 minutes.

A developer should be able to initialize Bottleneck, ingest common delivery evidence, run a scorecard, understand the primary bottleneck, and see exactly what to fix next without needing to understand the full BIASED framework first.

This implementation prompt is scoped to:

- Epic 5: Simplify Evidence Ingestion For Common SaaS Teams
- Task 5.1: Add one-command local evidence ingestion docs
- Task 5.2: Add sample SaaS evidence files
- Task 5.3: Add ingestion smoke tests

## Definition Of Done For This Slice

A SaaS developer can read the docs, run sample ingestion commands against local report files, and understand:

- What each input report represents.
- Where Bottleneck writes normalized evidence.
- Which scorecard category changes.
- How behavior and evidence IDs are linked.
- How to verify ingestion changed the scorecard.

The sample files should let a developer try ingestion immediately without needing real CI artifacts.

## Current Architecture To Respect

Inspect the repository before changing code. Follow existing patterns for:

- Cobra ingestion commands under `cmd/`
- ingestion implementations under `internal/ingest/`
- scorecard category evidence under `internal/scorecard/`
- traceability behavior under `internal/traceability/`
- existing ingestion fixtures under `internal/**/testdata`
- existing examples under `examples/`
- existing README and docs structure

Preserve backwards compatibility:

- Do not rename existing `ingest` subcommands.
- Do not rename public flags such as `--file` or `--dry-run`.
- Do not change normalized evidence schemas unless a failing test proves they are wrong.
- Do not weaken invalid-file handling.
- Do not make ingestion write files during `--dry-run`.
- Do not implement unrelated ingestion formats as part of this slice.

Prefer docs, fixtures, and smoke tests. Only change ingestion behavior if a sample-file test exposes a clear bug.

## Primary Source Of Truth

Read:

- `tasks/DayOneExperience/saas-team-day-one-success-task-list.md`
- `readme.md`
- `docs/quickstart-saas.md`, if it exists
- `cmd/ingest*.go`
- `cmd/*_test.go`
- `internal/ingest/*`
- `internal/scorecard/*`
- `internal/traceability/*`
- existing `examples/` and `internal/**/testdata` fixtures

Keep the sample domain consistent with the Day-One SaaS milestone:

- Subscription Billing Release
- Users can update payment methods and retry failed invoices
- `BEHAVIOR-001`: payment method update
- `BEHAVIOR-002`: failed invoice retry
- `BEHAVIOR-003`: duplicate-charge prevention during retry

## Epic 5: Simplify Evidence Ingestion For Common SaaS Teams

### Task 5.1 - Add One-Command Local Evidence Ingestion Docs

Goal: Show developers how to bring in test, security, and telemetry evidence.

Docs must include these commands:

```sh
bottleneck ingest cucumber --file reports/cucumber.json
bottleneck ingest sarif --file reports/codeql.sarif
bottleneck ingest test-summary --file reports/test-summary.json
bottleneck ingest telemetry --file reports/telemetry.json
```

For each example, explain:

- What the file represents.
- Where normalized evidence is written.
- Which scorecard category it updates.
- How evidence IDs are linked.

Recommended doc location:

- `docs/quickstart-saas.md`, if it exists
- README quickstart, if it already has an ingestion section
- a new focused docs page only if the repository already organizes docs that way

Recommended explanation shape:

```text
bottleneck ingest cucumber --file reports/cucumber.json

Represents: BDD/Cucumber scenario results from the SaaS billing test suite.
Writes: bottleneck/assurance/results.json.
Updates: Assurance scorecard category.
Links IDs: scenario tags such as @BEHAVIOR-003 map tests to behavior evidence.
```

```text
bottleneck ingest sarif --file reports/codeql.sarif

Represents: CodeQL or security scan findings in SARIF format.
Writes: bottleneck/security/guardrails.json.
Updates: Security scorecard category.
Links IDs: findings may include rules, locations, or tags that map to SECURITY-* or BEHAVIOR-* evidence where supported.
```

```text
bottleneck ingest test-summary --file reports/test-summary.json

Represents: summarized unit, integration, or end-to-end test results.
Writes: bottleneck/assurance/results.json or the repository's normalized assurance evidence path.
Updates: Assurance scorecard category.
Links IDs: test cases or suites can include BEHAVIOR-* references.
```

```text
bottleneck ingest telemetry --file reports/telemetry.json

Represents: execution signals such as deployment frequency, failure rate, error rate, override rate, adoption, and cost.
Writes: bottleneck/execution/telemetry.json.
Updates: Execution scorecard category.
Links IDs: telemetry signals can reference BEHAVIOR-* and EXECUTION-* evidence IDs.
```

If any command is not implemented, do not document it as working. Instead:

- Add a short "not yet implemented" note in the prompt final response.
- Add docs only for implemented commands.
- Add tests only for implemented commands.

### Task 5.2 - Add Sample SaaS Evidence Files

Goal: Let users run ingestion immediately.

Add sample files:

```text
examples/saas/reports/cucumber.json
examples/saas/reports/codeql.sarif
examples/saas/reports/test-summary.json
examples/saas/reports/telemetry.json
```

Each sample must map to SaaS billing behavior.

Sample file guidance:

- Keep files small and readable.
- Use realistic Subscription Billing Release names.
- Include behavior IDs where the ingestion format supports them.
- Include enough data to exercise scorecard changes.
- Include at least one sample that proves `BEHAVIOR-003` can become covered by ingested evidence.

Cucumber sample requirements:

- Include scenarios tagged with behavior IDs, for example:

```text
@BEHAVIOR-001
Scenario: Customer updates payment method for active subscription
```

- Include a scenario for `BEHAVIOR-003` payment retry duplicate-charge prevention.
- Include passing steps unless a failing fixture is specifically needed for an invalid or failure test.

SARIF sample requirements:

- Use valid SARIF JSON.
- Use CodeQL-like shape where practical.
- Include low or informational findings for the healthy sample.
- Include rule metadata or properties that represent severity.
- Avoid critical findings in the default sample unless the test is meant to prove blocking behavior.

Test summary sample requirements:

- Use the repository's supported test-summary schema if implemented.
- Include pass/fail counts.
- Include behavior or evidence references where supported.
- Include coverage or source information if supported.

Telemetry sample requirements:

- Use the repository's supported telemetry schema if implemented.
- Include deployment frequency.
- Include change failure rate.
- Include error rate.
- Include user override rate.
- Include adoption rate.
- Include cost signals.
- Include telemetry freshness/source fields where supported.

If a sample format is not implemented yet, create the file only if it is useful as a future fixture and clearly mark related tests/docs as deferred. Do not force product behavior to support an unimplemented format unless this epic explicitly owns that implementation and the existing code architecture already has a matching command stub.

### Task 5.3 - Add Ingestion Smoke Tests

Goal: Prove sample files work end-to-end.

Smoke tests must verify:

- ingestion runs against sample files
- evidence files are created
- evidence IDs are preserved or generated
- scorecard category changes after ingestion
- invalid files fail cleanly
- `--dry-run` does not write files

Implementation guidance:

- Prefer command-level smoke tests if the CLI command already supports temporary working directories.
- Otherwise, test `internal/ingest` package functions directly and add one CLI smoke test for flag wiring.
- Use temporary directories.
- Copy or initialize a small Bottleneck project before ingestion.
- Never write generated evidence to the repository root during tests.
- Keep invalid fixtures small.
- If ingestion writes normalized files, assert the exact expected path.
- If ingestion merges with existing evidence, assert IDs and counts instead of full snapshots.

Recommended smoke test shape:

1. Create temp directory.
2. Run or call equivalent of `bottleneck init --template saas`.
3. Confirm baseline scorecard category state.
4. Run ingestion against `examples/saas/reports/cucumber.json`.
5. Verify normalized assurance evidence exists.
6. Verify `BEHAVIOR-003` or related assurance IDs are present.
7. Re-run scorecard and verify Assurance category improves or missing evidence changes.
8. Repeat equivalent checks for SARIF and telemetry if commands are implemented.
9. Verify invalid input returns a useful error.
10. Verify `--dry-run` parses but does not create or modify normalized evidence.

Test requirements by input:

#### Cucumber

- Sample Cucumber file parses.
- Normalized assurance evidence is written.
- Behavior tags such as `@BEHAVIOR-003` are preserved or mapped.
- Scorecard Assurance category changes after ingestion.
- Invalid Cucumber JSON fails with a useful error.
- `--dry-run` does not write assurance evidence.

#### SARIF

- Sample SARIF file parses.
- Normalized security evidence is written.
- Severity values are preserved or normalized.
- Scorecard Security category reflects findings.
- Invalid SARIF fails with a useful error.
- `--dry-run` does not write security evidence.

#### Test Summary

- Only add tests if `ingest test-summary` is implemented.
- Sample test summary parses.
- Normalized assurance evidence is written.
- Test pass/fail counts and evidence source are preserved.
- Invalid test summary fails with a useful error.
- `--dry-run` does not write evidence.

#### Telemetry

- Sample telemetry file parses.
- Normalized execution evidence is written.
- Deployment frequency, change failure rate, error rate, override rate, adoption, and cost signals are preserved where supported.
- Scorecard Execution category reflects telemetry.
- Invalid telemetry fails with a useful error.
- `--dry-run` does not write execution evidence.

## UX Requirements

Ingestion should feel like a practical bridge from tools SaaS teams already use.

Prefer:

- one-command examples
- explicit file paths
- clear normalized output paths
- category names users see in scorecard
- behavior IDs such as `BEHAVIOR-003`
- small sample files developers can inspect

Avoid:

- unexplained framework jargon
- examples that require external services
- examples that require unavailable secrets
- large noisy report fixtures
- pretending unsupported commands work
- changing score math just to make samples look better

## Tests To Add Or Update

Add tests in the most appropriate package:

- ingestion parser and writer tests in `internal/ingest/*_test.go`
- CLI ingestion tests in `cmd/*_test.go`
- scorecard integration tests in `internal/scorecard/*_test.go` or an existing end-to-end test file
- docs/example path tests in a lightweight docs or examples test if the repository already has one

Minimum test coverage:

- Docs include all implemented ingestion commands.
- Sample SaaS report files exist at the required paths.
- Implemented ingestion commands can parse the sample files.
- Implemented ingestion commands write normalized evidence to the expected category path.
- Behavior/evidence IDs are preserved or generated.
- Scorecard category state changes after ingestion where the model supports it.
- Invalid files fail cleanly.
- `--dry-run` does not write files.

If a command is not implemented, do not add a failing test that demands new product behavior unless you are intentionally implementing that command in this slice. Report it as deferred in the final response.

## Documentation Updates

Update docs only where needed:

- Add one-command ingestion examples to `docs/quickstart-saas.md` if it exists.
- Add or update README links to the ingestion guide if useful.
- Mention `examples/saas/reports/*` sample files.
- Explain category impact:
  - Cucumber and test summary update Assurance.
  - SARIF updates Security.
  - Telemetry updates Execution.
- Explain ID linking:
  - Cucumber scenario tags such as `@BEHAVIOR-003`
  - SARIF rule IDs, properties, or evidence refs where supported
  - test summary behavior refs where supported
  - telemetry behavior and execution refs where supported

## Verification Commands

Run:

```sh
go test ./...
```

If feasible, manually verify from a temporary directory using the sample reports:

```sh
bottleneck init --template saas
bottleneck ingest cucumber --file /absolute/path/to/examples/saas/reports/cucumber.json
bottleneck ingest sarif --file /absolute/path/to/examples/saas/reports/codeql.sarif
bottleneck ingest test-summary --file /absolute/path/to/examples/saas/reports/test-summary.json
bottleneck ingest telemetry --file /absolute/path/to/examples/saas/reports/telemetry.json
bottleneck scorecard
bottleneck trace BEHAVIOR-003
```

Also verify dry runs:

```sh
bottleneck ingest cucumber --file /absolute/path/to/examples/saas/reports/cucumber.json --dry-run
bottleneck ingest sarif --file /absolute/path/to/examples/saas/reports/codeql.sarif --dry-run
bottleneck ingest test-summary --file /absolute/path/to/examples/saas/reports/test-summary.json --dry-run
bottleneck ingest telemetry --file /absolute/path/to/examples/saas/reports/telemetry.json --dry-run
```

For manual verification:

- Use a temporary directory.
- Do not create generated Bottleneck artifacts in the repository root.
- Record whether commands pass or fail.
- If any ingestion command is not implemented, document that clearly rather than treating it as a test failure.

## Final Response Requirements

When finished, report:

1. Ingestion docs changed.
2. Sample SaaS report files added.
3. Ingestion smoke tests added or changed.
4. Normalized evidence paths verified.
5. Scorecard category effects verified.
6. Invalid-file and dry-run behavior verified.
7. Exact commands run and results.
8. Any ingestion commands or acceptance criteria intentionally deferred and why.


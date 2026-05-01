# AI Implementation Prompt: Improve Error Messages And Help Text

You are working in the `bottleneck` Go CLI repository.

Implement the eighth slice of the **SaaS Team Day-One Success** milestone.

## Milestone Goal

Make Bottleneck usable by a SaaS engineering team in the first 10 minutes.

A new user should know what to run first, and error messages should explain the next useful action instead of leaving the user at a dead end.

This implementation prompt is scoped to:

- Epic 8: Improve Error Messages And Help Text
- Task 8.1: Rewrite top-level CLI help
- Task 8.2: Add helpful empty-state messages
- Task 8.3: Add command suggestion tests

## Definition Of Done For This Slice

A SaaS developer can run:

```sh
bottleneck --help
```

And immediately see the Day-One path:

```text
Start here:
  bottleneck init --template saas
  bottleneck scorecard
  bottleneck diagnose
```

When evidence is missing or invalid, Bottleneck tells the user what to do next with a concrete command, file path, or evidence ID.

## Current Architecture To Respect

Inspect the repository before changing code. Follow existing patterns for:

- Cobra root command and help text under `cmd/`
- command descriptions and examples
- validation errors under `internal/validator/`
- scorecard missing-evidence messages under `internal/scorecard/`
- diagnosis messages under `internal/diagnosis/`
- ingestion errors under `internal/ingest/`
- traceability findings under `internal/traceability/`
- GitHub annotation formatting, if errors surface there
- existing command and message tests

Preserve backwards compatibility:

- Do not rename commands or flags.
- Do not remove existing help content that is still accurate.
- Do not change exit codes unless a failing test proves they are wrong.
- Do not weaken validation, scorecard, diagnosis, gate, or ingestion behavior.
- Do not hide detailed errors; add next-step guidance alongside them.
- Do not add suggestions for commands that are not implemented.

Prefer focused message improvements and tests. Only change product behavior when needed to make an existing error actionable.

## Primary Source Of Truth

Read:

- `tasks/DayOneExperience/saas-team-day-one-success-task-list.md`
- `readme.md`
- `docs/quickstart-saas.md`, if it exists
- `cmd/root.go`
- `cmd/init.go`
- `cmd/validate.go`
- `cmd/scorecard.go`
- `cmd/diagnose.go`
- `cmd/trace.go`
- `cmd/ingest*.go`
- `cmd/explain.go`
- `cmd/*_test.go`
- `internal/validator/*`
- `internal/scorecard/*`
- `internal/diagnosis/*`
- `internal/traceability/*`
- `internal/ingest/*`

Keep the Day-One flow consistent with the SaaS milestone:

```sh
bottleneck init --template saas
bottleneck scorecard
bottleneck diagnose
bottleneck trace BEHAVIOR-003
```

## Epic 8: Improve Error Messages And Help Text

### Task 8.1 - Rewrite Top-Level CLI Help

Goal: A new user should know what to run first.

Top-level help must include:

```text
Start here:
  bottleneck init --template saas
  bottleneck scorecard
  bottleneck diagnose
```

Help must briefly explain:

- `validate`
- `scorecard`
- `diagnose`
- `trace`
- `ingest`
- `explain`

Implementation guidance:

- Update the root command long description, example section, or help template using existing Cobra patterns.
- Put start-here guidance near the top of `bottleneck --help`.
- Keep the command descriptions short and practical.
- Position `scorecard` as the primary value command.
- Keep `validate` as an evidence quality check, not the main product surface.
- Do not remove existing command help unless it is stale or misleading.

Recommended help language:

```text
Start here:
  bottleneck init --template saas
  bottleneck scorecard
  bottleneck diagnose

Common commands:
  validate    Check evidence files for missing, thin, or placeholder content.
  scorecard   Show release readiness, primary bottleneck, and next action.
  diagnose    Explain what is blocking delivery and what to inspect next.
  trace       Follow one intent, behavior, or evidence ID end-to-end.
  ingest      Convert test, security, and telemetry reports into Bottleneck evidence.
  explain     Show how evidence affected category scores.
```

Test requirements:

- `bottleneck --help` includes the exact `Start here:` section.
- Help briefly explains each required command.
- Help includes `bottleneck init --template saas`.
- Help includes `bottleneck scorecard`.
- Help includes `bottleneck diagnose`.

### Task 8.2 - Add Helpful Empty-State Messages

Goal: Avoid dead-end errors.

Missing evidence messages must say what to do next.

Target shape:

```text
No assurance evidence found.
Next action: Add test evidence manually or run:
  bottleneck ingest cucumber --file reports/cucumber.json
```

Implementation guidance:

- Add targeted next-step guidance to common missing or invalid states.
- Prefer concrete commands and paths.
- Preserve the original error reason.
- Keep suggestions short enough for terminal output.
- Avoid generic "see docs" messages unless paired with a concrete action.
- If an ingestion command is not implemented, do not suggest it.

Required empty-state guidance:

- Missing config:
  - explain that Bottleneck has not been initialized
  - suggest `bottleneck init --template saas`
- Missing evidence directory:
  - name the missing directory
  - suggest `bottleneck init --template saas` or adding the expected evidence files
- Invalid environment:
  - name the invalid environment
  - list or suggest supported environments
  - suggest `--env=dev` or `--env=production`
- Broken evidence reference:
  - name the broken reference
  - suggest `bottleneck trace <ID>` when an ID is available
  - suggest updating the referenced evidence file
- Missing assurance:
  - say no assurance evidence was found
  - suggest manual evidence or `bottleneck ingest cucumber --file reports/cucumber.json`
- Invalid ingestion file:
  - name the file
  - explain parse/schema issue where available
  - suggest checking the expected sample format
- Placeholder-heavy artifact:
  - name the file
  - explain that placeholder content does not support release readiness
  - suggest replacing it with real SaaS evidence or running `bottleneck init --template saas` in a new project for an example

Recommended message examples:

```text
No Bottleneck config found.
Next action: initialize a SaaS starter project:
  bottleneck init --template saas
```

```text
No assurance evidence found.
Next action: Add test evidence manually or run:
  bottleneck ingest cucumber --file reports/cucumber.json
```

```text
Unknown environment "prod-us".
Next action: choose one of: local, dev, test, stage, production.
Example:
  bottleneck scorecard --env=production
```

```text
Broken evidence reference: BEHAVIOR-003 references ASSURANCE-003, but ASSURANCE-003 was not found.
Next action: add ASSURANCE-003 or inspect the behavior trace:
  bottleneck trace BEHAVIOR-003
```

Do not over-normalize all errors into the same message. Specificity matters.

### Task 8.3 - Add Command Suggestion Tests

Goal: Keep UX intentional.

Tests must verify next-step guidance for:

- missing config
- missing evidence directory
- invalid environment
- broken evidence reference
- missing assurance
- invalid ingestion file
- placeholder-heavy artifact

Implementation guidance:

- Add tests near the package that owns the message.
- Use CLI tests when the message is surfaced by a command.
- Use package-level tests when the message is generated by validator, scorecard, diagnosis, traceability, or ingest code.
- Prefer small fixture projects and temp directories.
- Avoid brittle full-output snapshots; assert high-signal substrings.
- Ensure each test checks both:
  - the reason
  - the next action

Recommended test assertions:

```text
missing config:
- contains "No Bottleneck config"
- contains "bottleneck init --template saas"

missing assurance:
- contains "No assurance evidence found"
- contains "Next action"
- contains "bottleneck ingest cucumber --file reports/cucumber.json"

invalid environment:
- contains invalid environment name
- contains supported environment such as "production"
- contains "scorecard --env=production" or equivalent

broken reference:
- contains missing ID
- contains "bottleneck trace"

placeholder-heavy artifact:
- contains file path
- contains "placeholder"
- contains "replace" or "real evidence"
```

If a signal is not currently implemented, do not invent a broad new feature just for the message. Add the strongest test for the existing surfaced error and report remaining gaps.

## UX Requirements

Help and error messages should make Bottleneck feel usable without a framework deck.

Prefer:

- start-here commands
- concrete next actions
- one command the user can run
- file paths and evidence IDs
- SaaS-friendly examples
- short explanations of why the issue matters

Avoid:

- generic "validation failed" dead ends
- long framework descriptions
- suggestions that require unsupported commands
- hiding the original error
- changing output schemas without tests
- noisy walls of text for simple errors

## Tests To Add Or Update

Add tests in the most appropriate package:

- root help tests in `cmd/*_test.go`
- validator message tests in `internal/validator/*_test.go`
- scorecard or diagnosis message tests in `internal/scorecard/*_test.go` or `internal/diagnosis/*_test.go`
- traceability broken-reference tests in `internal/traceability/*_test.go`
- ingestion invalid-file tests in `internal/ingest/*_test.go` or command tests

Minimum test coverage:

- Top-level help includes start-here commands.
- Top-level help briefly explains `validate`, `scorecard`, `diagnose`, `trace`, `ingest`, and `explain`.
- Missing config includes next-step guidance.
- Missing evidence directory includes next-step guidance.
- Invalid environment includes next-step guidance.
- Broken evidence reference includes next-step guidance.
- Missing assurance includes next-step guidance.
- Invalid ingestion file includes next-step guidance.
- Placeholder-heavy artifact includes next-step guidance.
- Existing command help tests still pass.

## Documentation Updates

Update docs only where useful:

- README should not contradict top-level help.
- `docs/quickstart-saas.md`, if it exists, should use the same start-here commands.
- If new error guidance points to docs, ensure the linked docs exist.
- Do not add long docs just to explain short CLI messages.

## Verification Commands

Run:

```sh
go test ./...
```

If feasible, manually verify:

```sh
bottleneck --help
bottleneck scorecard
bottleneck scorecard --env=not-real
bottleneck ingest cucumber --file missing.json
```

For manual verification:

- Use a temporary directory for commands that create or inspect project evidence.
- Do not create generated Bottleneck artifacts in the repository root.
- Record whether commands pass or fail.
- Expected failures are acceptable when they include useful next-step guidance.

## Final Response Requirements

When finished, report:

1. Help text changes.
2. Empty-state or error message changes.
3. Command suggestion tests added or changed.
4. Any unsupported command suggestions intentionally avoided.
5. Exact commands run and results.
6. Any acceptance criteria intentionally deferred and why.


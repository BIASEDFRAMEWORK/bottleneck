# AI Implementation Prompt: Make Scorecard The Main Product Surface

You are working in the `bottleneck` Go CLI repository.

Implement the third slice of the **SaaS Team Day-One Success** milestone.

## Milestone Goal

Make Bottleneck usable by a SaaS engineering team in the first 10 minutes.

A developer should be able to initialize Bottleneck, run a scorecard, understand the primary bottleneck, and see exactly what to fix next without needing to understand the full BIASED framework first.

This implementation prompt is scoped to:

- Epic 3: Make Scorecard The Main Product Surface
- Task 3.1: Make `scorecard` the primary quickstart command
- Task 3.2: Improve plain-text scorecard readability
- Task 3.3: Add `--summary` or make concise output the default

## Definition Of Done For This Slice

A SaaS developer can run:

```sh
bottleneck init --template saas
bottleneck scorecard
```

And immediately see:

- A clear release recommendation.
- The primary bottleneck.
- Concise category results.
- Why the bottleneck matters.
- The next action to take.

The default scorecard output should be useful in a terminal without requiring users to read raw validation findings or understand the full BIASED framework.

## Current Architecture To Respect

Inspect the repository before changing code. Follow existing patterns for:

- Cobra command definitions under `cmd/`
- scorecard models and rendering under `internal/scorecard/`
- diagnosis or bottleneck selection under `internal/diagnosis/`
- config and threshold behavior
- existing text, JSON, and Markdown output formats
- existing command tests under `cmd/*_test.go`
- existing scorecard tests under `internal/scorecard/*_test.go`

Preserve backwards compatibility:

- Do not rename the `scorecard` command.
- Do not remove existing `--format` values.
- Do not rename JSON fields or change schemas unless a failing test proves they are wrong.
- Do not weaken release gate or score calculations to make output look nicer.
- Do not remove detailed evidence output if existing users rely on it.
- Do not change default `bottleneck validate` behavior.

Prefer small, reviewable changes. If you add a new flag, keep it focused and consistent with existing CLI conventions.

## Primary Source Of Truth

Read:

- `tasks/DayOneExperience/saas-team-day-one-success-task-list.md`
- `readme.md`
- `docs/quickstart-saas.md`, if it exists
- `cmd/scorecard.go`
- `cmd/root.go`
- `cmd/*_test.go`
- `internal/scorecard/*`
- `internal/diagnosis/*`
- existing SaaS starter template or fixture tests

Use the SaaS starter template as the primary UX test case if it exists:

```sh
bottleneck init --template saas
```

The expected Day-One bottleneck is:

- Primary bottleneck: `Assurance`
- Missing evidence: `BEHAVIOR-003` has no mapped test evidence
- Next action: add assurance evidence for payment retry behavior
- Release recommendation: `Conditional`

If current scorecard terminology differs, inspect existing release recommendation rules before changing behavior. If the roadmap requires `Conditional`, add a focused test for the SaaS starter and make the smallest safe mapping change.

## Epic 3: Make Scorecard The Main Product Surface

### Task 3.1 - Make Scorecard The Primary Quickstart Command

Goal: Developers should see value immediately.

Requirements:

- README positions `scorecard` as the main command after initialization.
- CLI help positions `scorecard` as the main command for seeing delivery risk and release readiness.
- Scorecard output is easier to understand than raw validation output.
- Scorecard includes a clear release recommendation.
- Scorecard includes the primary bottleneck.
- Scorecard includes the next recommended action.

Implementation guidance:

- Update README language only as needed to make `scorecard` the primary Day-One product surface.
- Update root command help or scorecard command help so new users know to run:

```sh
bottleneck init --template saas
bottleneck scorecard
bottleneck diagnose
```

- Keep `validate` documented as a useful evidence quality check, but do not make it the primary value moment.
- The scorecard should summarize the delivery decision before detailed evidence.

Test requirements:

- Test CLI help includes `scorecard` in a prominent start-here or quickstart section if help text is generated in code.
- Test README mentions `scorecard` as the primary command.
- Test plain-text scorecard includes release recommendation, primary bottleneck, and next action.

### Task 3.2 - Improve Plain-Text Scorecard Readability

Goal: Make terminal output instantly useful.

Plain-text scorecard output must include:

- `Bottleneck Scorecard`
- `Environment: dev`
- `Release Recommendation: Conditional`
- `Primary Bottleneck: Assurance`
- concise category results
- a clear `Why` section
- a clear `Next Action` section

Target shape:

```text
Bottleneck Scorecard
Environment: dev
Release Recommendation: Conditional
Primary Bottleneck: Assurance

Category Results:
- Intent: Pass
- Behavior: Pass
- Design: Pass
- Assurance: Warn
- Security: Pass
- Execution: Warn

Why:
Payment retry behavior has no mapped test evidence.

Next Action:
Add assurance evidence for BEHAVIOR-003.
```

Implementation guidance:

- Put the top-level decision first.
- Keep category results scan-friendly.
- Use plain language for `Why`; avoid generic framework descriptions.
- Include evidence IDs where they help the developer take action.
- Prefer deterministic ordering:
  - Intent
  - Behavior
  - Design
  - Assurance
  - Security
  - Execution
  - Governance, if implemented
- Keep text output stable enough for tests.
- Do not pollute JSON or Markdown output with terminal-only formatting.

If current text output has richer details, preserve those details behind an existing detailed mode or a new `--details` flag.

Test requirements:

- Add or update tests for plain-text scorecard rendering.
- Verify the SaaS starter output includes the target high-level fields.
- Verify category results are concise and ordered.
- Verify `Why` mentions payment retry or `BEHAVIOR-003`.
- Verify `Next Action` tells the user to add assurance evidence for `BEHAVIOR-003`.
- Verify JSON and Markdown output remain valid and deterministic.

### Task 3.3 - Add `--summary` Or Default Concise Mode

Goal: Avoid overwhelming new users.

Requirements:

- Default scorecard output is concise.
- Detailed evidence can still be shown with an existing or new flag, such as:

```sh
bottleneck scorecard --details
```

- Tests verify default output is not overly verbose.
- Tests verify detailed output still includes evidence-level reasoning.

Implementation options:

1. Make the default text scorecard concise and add `--details`.
2. Add `--summary` while preserving current default output.
3. If an equivalent flag already exists, reuse it rather than adding another flag.

Prefer option 1 if it can be done without breaking existing documented behavior. Prefer option 3 if the repository already has a clear detailed or summary mode.

Concise default output should include:

- heading
- environment
- release recommendation
- primary bottleneck
- category results
- why
- next action

Detailed output should include:

- evidence counts
- missing evidence
- reasons
- score impacts
- thresholds
- related evidence IDs
- category-level reasoning

Test requirements:

- Default `bottleneck scorecard` output should be bounded and scan-friendly.
- Use a pragmatic line-count or section-count assertion rather than exact full-output snapshots.
- `bottleneck scorecard --details`, or the chosen equivalent, should include evidence-level reasoning.
- Unsupported scorecard flags should still return useful errors.

## UX Requirements

The scorecard should feel like the product's main screen in terminal form.

Prefer:

- delivery risk language
- release readiness language
- direct bottleneck naming
- evidence IDs tied to next steps
- clear category pass/warn/fail summaries
- SaaS examples around billing, payment retry, duplicate charges, and telemetry

Avoid:

- long framework explanations
- walls of raw validation findings before the recommendation
- vague actions such as "improve evidence"
- hiding the release recommendation below detailed diagnostics
- changing score math only to improve wording

## Tests To Add Or Update

Add tests in the most appropriate package:

- CLI help behavior in `cmd/*_test.go`
- scorecard rendering in `internal/scorecard/*_test.go`
- end-to-end command output in `cmd/scorecard_test.go` or existing CLI command tests
- README checks in an existing docs test or a small new docs test

Minimum test coverage:

- README positions `scorecard` as the main quickstart command.
- Root or scorecard help positions `scorecard` as the command to see release readiness.
- Default text scorecard includes:
  - `Bottleneck Scorecard`
  - `Environment: dev`
  - `Release Recommendation: Conditional`
  - `Primary Bottleneck: Assurance`
  - `Category Results:`
  - `Why:`
  - `Next Action:`
- Default text scorecard includes concise status lines for Intent, Behavior, Design, Assurance, Security, and Execution.
- Default text scorecard output is not overly verbose.
- Detailed mode includes evidence-level reasoning.
- JSON and Markdown formats still work.
- Existing scorecard tests still pass.

If the SaaS starter template is not implemented yet, create tests using the smallest realistic fixture that produces:

- Assurance warning
- `BEHAVIOR-003` missing mapped test evidence
- conditional release recommendation

Do not implement unrelated SaaS template work as part of this slice.

## Documentation Updates

Update docs only where needed:

- README quickstart should point users to `scorecard` as the first command that explains delivery risk.
- If `docs/quickstart-saas.md` exists, update it to match the new default scorecard output.
- If you add `--details`, document it briefly.
- Keep validation documented as a supporting command for evidence quality.

## Verification Commands

Run:

```sh
go test ./...
```

If feasible, manually verify from a temporary directory:

```sh
bottleneck init --template saas
bottleneck scorecard
bottleneck scorecard --details
bottleneck scorecard --format=json
bottleneck scorecard --format=markdown
```

For manual verification:

- Use a temporary directory.
- Do not create generated Bottleneck artifacts in the repository root.
- Record whether commands pass or fail.
- If the SaaS starter intentionally exits non-zero because the recommendation is Conditional, document that as expected.

## Final Response Requirements

When finished, report:

1. Scorecard output changes.
2. CLI help and README changes.
3. Summary or details flag behavior.
4. Tests added or changed.
5. Any release recommendation mapping changes.
6. Exact commands run and results.
7. Any acceptance criteria intentionally deferred and why.


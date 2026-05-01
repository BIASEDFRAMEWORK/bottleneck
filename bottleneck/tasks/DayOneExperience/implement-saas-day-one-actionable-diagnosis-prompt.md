# AI Implementation Prompt: Make Diagnosis Actionable

You are working in the `bottleneck` Go CLI repository.

Implement the fourth slice of the **SaaS Team Day-One Success** milestone.

## Milestone Goal

Make Bottleneck usable by a SaaS engineering team in the first 10 minutes.

A developer should be able to initialize Bottleneck, run diagnosis, understand the primary bottleneck, and see exactly what to fix next without needing to understand the full BIASED framework first.

This implementation prompt is scoped to:

- Epic 4: Make Diagnosis Actionable
- Task 4.1: Add next action to `diagnose`
- Task 4.2: Add diagnosis rules for common SaaS bottlenecks
- Task 4.3: Add test cases for clear bottleneck prioritization

## Definition Of Done For This Slice

A SaaS developer can run:

```sh
bottleneck init --template saas
bottleneck diagnose
```

And immediately see:

- The single primary bottleneck.
- The reason it was selected.
- The release impact.
- The next action to take.
- Relevant evidence IDs.
- A suggested command to inspect the bottleneck further.
- Supporting issues without noisy or contradictory prioritization.

The diagnosis output should be action-oriented enough that a developer knows what to fix next without reading raw validation output.

## Current Architecture To Respect

Inspect the repository before changing code. Follow existing patterns for:

- Cobra command definitions under `cmd/`
- diagnosis models and rendering under `internal/diagnosis/`
- scorecard and release recommendation inputs under `internal/scorecard/`
- validator findings under `internal/validator/`
- traceability evidence and findings under `internal/traceability/`
- gate behavior under `internal/gate/`, if diagnosis uses it
- existing text, JSON, Markdown, and GitHub output formats
- existing command tests under `cmd/*_test.go`
- existing diagnosis tests under `internal/diagnosis/*_test.go`

Preserve backwards compatibility:

- Do not rename the `diagnose` command.
- Do not remove existing `--format` values.
- Do not rename JSON fields or output schemas unless a failing test proves they are wrong.
- Do not weaken release gate or scorecard rules to make diagnosis look cleaner.
- Do not hide existing findings; prioritize them and make the primary diagnosis clearer.
- Do not change `scorecard` or `validate` behavior unless a diagnosis test exposes a clear shared bug.

Prefer small, reviewable changes. Keep prioritization logic explicit and testable.

## Primary Source Of Truth

Read:

- `tasks/DayOneExperience/saas-team-day-one-success-task-list.md`
- `readme.md`
- `docs/quickstart-saas.md`, if it exists
- `cmd/diagnose.go`
- `cmd/*_test.go`
- `internal/diagnosis/*`
- `internal/scorecard/*`
- `internal/validator/*`
- `internal/traceability/*`
- `internal/gate/*`
- existing SaaS starter template or fixture tests

Use the SaaS starter template as the primary UX test case if it exists:

```sh
bottleneck init --template saas
```

The expected Day-One diagnosis is:

- Primary bottleneck: `Assurance`
- Reason: `BEHAVIOR-003` is not linked to any passing test evidence
- Impact: release confidence is reduced because payment retry behavior is unproven
- Next action: add or ingest test evidence mapped to `BEHAVIOR-003`
- Inspect command: `bottleneck trace BEHAVIOR-003`

## Epic 4: Make Diagnosis Actionable

### Task 4.1 - Add Next Action To Diagnose

Goal: Developers should know what to do next.

`bottleneck diagnose` output must include:

- primary bottleneck
- reason
- impact
- next action
- relevant evidence IDs
- suggested command to inspect further

Target shape:

```text
Primary Bottleneck: Assurance
Reason: BEHAVIOR-003 is not linked to any passing test evidence.
Impact: Release confidence is reduced because payment retry behavior is unproven.
Next Action: Add or ingest test evidence mapped to BEHAVIOR-003.
Inspect: bottleneck trace BEHAVIOR-003
```

Implementation guidance:

- Put the primary bottleneck and next action near the top of text output.
- Keep wording specific to the evidence gap; avoid generic descriptions such as "improve assurance."
- Include related IDs when known, especially `INTENT-*`, `BEHAVIOR-*`, `ASSURANCE-*`, `SECURITY-*`, and `EXECUTION-*`.
- Use the most specific inspect command available:
  - `bottleneck trace BEHAVIOR-003` for a behavior coverage gap
  - `bottleneck trace INTENT-001` for an intent-to-behavior gap
  - `bottleneck scorecard --details`, or existing equivalent, for category-level score evidence
- Preserve JSON and Markdown output. If diagnosis models do not currently carry these fields, add them carefully with tests.
- Do not invent fake IDs. If no ID is known, say what evidence category is missing and suggest the best supported command.

Test requirements:

- Text diagnosis includes `Primary Bottleneck`, `Reason`, `Impact`, `Next Action`, relevant IDs, and `Inspect`.
- JSON diagnosis includes equivalent structured fields if JSON output exists.
- Markdown diagnosis includes the same actionable sections if Markdown output exists.
- The SaaS starter diagnosis points at `BEHAVIOR-003`.
- The suggested inspect command uses `bottleneck trace BEHAVIOR-003`.

### Task 4.2 - Add Diagnosis Rules For Common SaaS Bottlenecks

Goal: Make diagnosis feel relevant to SaaS teams.

Diagnosis must identify and explain:

- missing intent
- behavior not mapped to intent
- behavior not covered by tests
- security findings blocking release
- stale telemetry
- missing production readiness evidence
- thin placeholder artifacts
- low traceability confidence

Implementation guidance:

- Prefer explicit, named diagnosis rules over hard-to-follow string matching.
- Keep rule inputs tied to existing scorecard, validator, traceability, telemetry, and security findings.
- Each rule should provide:
  - category or bottleneck name
  - reason
  - impact
  - next action
  - related evidence IDs, when available
  - suggested inspect command
  - severity or priority signal
- Use SaaS-specific language where evidence supports it:
  - payment retry
  - duplicate-charge prevention
  - payment method update
  - failed invoice retry
  - tokenized payment details
  - billing telemetry
- Keep generic fallback rules for non-SaaS projects.

Recommended rule examples:

```text
Missing Intent
Reason: No intent evidence describes the customer outcome.
Impact: The team cannot tell what release risk the evidence is meant to reduce.
Next Action: Add intent evidence with measurable SaaS outcome and related behavior IDs.
Inspect: bottleneck validate
```

```text
Behavior Not Mapped To Intent
Reason: BEHAVIOR-003 is not linked to INTENT-001.
Impact: The payment retry behavior is not traceable to a customer or release outcome.
Next Action: Add an INTENT-001 reference to BEHAVIOR-003 or update the intent evidence.
Inspect: bottleneck trace BEHAVIOR-003
```

```text
Behavior Not Covered By Tests
Reason: BEHAVIOR-003 is not linked to any passing assurance evidence.
Impact: Release confidence is reduced because payment retry duplicate-charge prevention is unproven.
Next Action: Add or ingest test evidence mapped to BEHAVIOR-003.
Inspect: bottleneck trace BEHAVIOR-003
```

```text
Security Blocker
Reason: Critical security findings are present.
Impact: Production release should not proceed while critical payment or account-security risk remains open.
Next Action: Resolve critical findings or add accepted-risk governance evidence.
Inspect: bottleneck scorecard --details
```

```text
Stale Telemetry
Reason: Execution telemetry is stale.
Impact: Release readiness is based on old production behavior and may miss current billing failures.
Next Action: Refresh telemetry evidence or ingest the latest execution metrics.
Inspect: bottleneck trace EXECUTION-001
```

Do not implement broad new roadmap features just to make a rule possible. If a signal is not modeled yet, add tests only for modeled signals and report the unimplemented signal in the final response.

Test requirements:

- Add focused tests for each implemented diagnosis rule.
- Use small fixture directories under package `testdata` where useful.
- Avoid massive fixture duplication; build helpers if the package already has fixture helpers.
- Verify each rule produces actionable language, not just category names.

### Task 4.3 - Add Test Cases For Clear Bottleneck Prioritization

Goal: Avoid noisy or contradictory diagnosis.

Prioritization tests must verify:

- security blocker outranks stale telemetry
- missing intent outranks missing design
- missing assurance for critical behavior outranks minor docs gaps
- production release blockers are clearer than development warnings
- diagnosis returns one primary bottleneck plus supporting issues

Implementation guidance:

- Define or document a deterministic priority order.
- Keep environment-sensitive behavior explicit:
  - production blockers should outrank dev warnings
  - security and assurance release blockers should be clearer than stale telemetry warnings when both exist
  - missing intent should outrank missing downstream design because design cannot be evaluated without intent
- Do not suppress supporting issues. The output should have one primary bottleneck and a supporting issue list.
- Ensure the same input produces the same primary bottleneck across runs.

Recommended priority shape:

1. Critical security blocker
2. Missing intent or no customer outcome
3. Critical behavior without assurance evidence
4. Broken traceability for required release evidence
5. Missing production readiness evidence
6. Stale or missing telemetry where required
7. Placeholder-heavy or thin evidence
8. Minor documentation gaps

Adjust the exact order to match existing release gate semantics, but make it explicit in tests.

Test requirements:

- Add a test where critical security findings and stale telemetry both exist; primary bottleneck should be Security.
- Add a test where intent and design are both missing; primary bottleneck should be Intent.
- Add a test where a critical behavior lacks assurance and docs are thin; primary bottleneck should be Assurance.
- Add a test showing production diagnosis uses stronger blocker language than dev diagnosis for the same evidence gap if environment-specific behavior exists.
- Add a test that supporting issues remain visible while only one primary bottleneck is selected.

## UX Requirements

Diagnosis should feel like a clear engineering triage note.

Prefer:

- direct statements of what is blocked
- concrete evidence IDs
- clear release impact
- one primary bottleneck
- supporting issues ordered by severity
- a command the developer can run next

Avoid:

- generic framework descriptions
- long lists before the primary bottleneck
- contradictory primary issues
- next actions with no evidence ID or command
- vague guidance such as "add more evidence"
- changing score calculations purely for wording

## Tests To Add Or Update

Add tests in the most appropriate package:

- diagnosis rule and prioritization tests in `internal/diagnosis/*_test.go`
- CLI output tests in `cmd/diagnose_test.go` or existing command tests
- traceability fixture tests in `internal/traceability/*_test.go` only if needed
- release gate interaction tests in `internal/gate/*_test.go` only if diagnosis uses gate outputs

Minimum test coverage:

- `bottleneck diagnose` text includes:
  - `Primary Bottleneck: Assurance`
  - `Reason:`
  - `Impact:`
  - `Next Action:`
  - `BEHAVIOR-003`
  - `Inspect: bottleneck trace BEHAVIOR-003`
- JSON and Markdown formats include equivalent actionable fields or sections, if those formats are supported.
- Diagnosis identifies each common SaaS bottleneck that has an implemented signal.
- Prioritization tests cover the required ordering cases.
- Diagnosis returns one primary bottleneck plus supporting issues.
- Existing diagnosis and scorecard tests still pass.

If a required signal is not currently implemented, do not invent it as part of this slice. Add the closest test for the implemented behavior and report the remaining gap.

## Documentation Updates

Update docs only where needed:

- README or quickstart docs should show the actionable diagnosis target shape.
- If `docs/quickstart-saas.md` exists, update the diagnosis step to include reason, impact, next action, and inspect command.
- Keep docs consistent with actual command output.
- Do not document non-existent flags or formats.

## Verification Commands

Run:

```sh
go test ./...
```

If feasible, manually verify from a temporary directory:

```sh
bottleneck init --template saas
bottleneck diagnose
bottleneck diagnose --format=json
bottleneck diagnose --format=markdown
bottleneck trace BEHAVIOR-003
```

For manual verification:

- Use a temporary directory.
- Do not create generated Bottleneck artifacts in the repository root.
- Record whether commands pass or fail.
- If the SaaS starter intentionally exits non-zero because the recommendation is Conditional or blocked by Assurance, document that as expected.

## Final Response Requirements

When finished, report:

1. Diagnosis output changes.
2. New or updated diagnosis rules.
3. Prioritization behavior.
4. Tests added or changed.
5. Any CLI output format changes.
6. Exact commands run and results.
7. Any acceptance criteria intentionally deferred and why.


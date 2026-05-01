# AI Implementation Prompt: SaaS Team Day-One Success Starter Template

You are working in the `bottleneck` Go CLI repository.

Implement the first slice of the **SaaS Team Day-One Success** milestone.

## Milestone Goal

Make Bottleneck usable by a SaaS engineering team in the first 10 minutes.

A developer should be able to initialize Bottleneck, ingest common delivery evidence, run a scorecard, understand the primary bottleneck, and see exactly what to fix next without needing to understand the full BIASED framework first.

This implementation prompt is scoped to:

- Epic 1: Create a SaaS Starter Template
- Task 1.1: Add `--template saas` to `bottleneck init`
- Task 1.2: Replace generic placeholders with realistic SaaS examples
- Task 1.3: Add a complete passing SaaS example fixture

## Definition Of Done For This Slice

A developer can run:

```sh
bottleneck init --template saas
bottleneck scorecard
bottleneck diagnose
bottleneck trace BEHAVIOR-001
```

And immediately understand:

- What SaaS capability Bottleneck is checking.
- What evidence exists.
- What evidence is missing or weak.
- What blocks release readiness.
- What should be fixed next.

Do not require users to understand the full BIASED framework before getting useful output.

## Current Architecture To Respect

Inspect the repository before changing code. Follow existing patterns for:

- Cobra command flags under `cmd/`
- `bottleneck init` file generation
- artifact paths under `bottleneck/`
- validator and scorecard tests
- traceability fixtures
- ingestion fixtures
- existing no-overwrite behavior

Preserve backwards compatibility:

- Existing `bottleneck init` behavior must remain unchanged.
- Existing generated default artifacts must not change unless a test proves they must.
- Existing commands, flags, output schemas, file paths, and evidence IDs must not be renamed.
- Do not overwrite existing user files.

Prefer small, reviewable changes.

## Epic 1: Create A SaaS Starter Template

### Task 1.1 - Add `--template saas` To `bottleneck init`

Goal: Let a SaaS team initialize realistic starter artifacts.

Required command:

```sh
bottleneck init --template saas
```

Requirements:

- Add a `--template` flag to `bottleneck init`.
- Supported values should include:
  - default template, preserving current behavior
  - `saas`
- The default `bottleneck init` command must behave exactly as it does today.
- `bottleneck init --template saas` should create a SaaS-oriented Bottleneck project structure.
- The SaaS template should use a recognizable SaaS capability.
- The init command must preserve existing safe no-overwrite behavior.
- Unsupported template values should fail with a useful error.

Suggested SaaS sample domain:

- Subscription Billing Release

Suggested feature:

- Users can update payment method and retry failed invoices.

Test requirements:

- Default `bottleneck init` still creates the existing default starter.
- `bottleneck init --template saas` creates SaaS-specific starter artifacts.
- Existing files are not overwritten by the SaaS template.
- Unsupported template values return a useful error.

### Task 1.2 - Replace Generic Placeholders With Realistic SaaS Examples

Goal: Make generated artifacts feel immediately understandable.

The SaaS template should generate realistic examples, not generic placeholder text.

Generated SaaS template must include:

- `bottleneck/config.yaml`
- `bottleneck/intent/intent.md`
- `bottleneck/behavior/behavior-spec.md`
- `bottleneck/design/architecture.md`
- `bottleneck/assurance/results.json` or equivalent normalized test evidence
- `bottleneck/security/guardrails.json` or equivalent normalized security evidence
- `bottleneck/execution/telemetry.json` or equivalent normalized telemetry evidence
- `bottleneck/docs/validation.md` if the default init currently creates it

Generated artifacts must include evidence IDs such as:

- `INTENT-001`
- `BEHAVIOR-001`
- `DESIGN-001`
- `ASSURANCE-001`
- `SECURITY-001`
- `EXECUTION-001`

Recommended SaaS evidence chain:

- `INTENT-001`: Customers can update payment methods without duplicate charges, lost billing state, or exposure of payment details.
- `BEHAVIOR-001`: Customer updates payment method for an active subscription.
- `BEHAVIOR-002`: System retries a failed invoice after payment method update.
- `BEHAVIOR-003`: System prevents duplicate charges during retry.
- `DESIGN-001`: Billing retry flow uses idempotency keys, payment provider tokens, and invoice state transitions.
- `ASSURANCE-001`: Passing test/evaluation evidence for payment method update.
- `ASSURANCE-002`: Passing test/evaluation evidence for invoice retry.
- `SECURITY-001`: Security evidence that payment details are tokenized and not stored directly.
- `EXECUTION-001`: Telemetry evidence for billing retry success, error rate, adoption, and cost.

Example intent text:

```text
INTENT-001: Customers must be able to update payment methods without duplicate charges, lost billing state, or exposure of payment details.
```

Recommended SaaS starter posture:

- The SaaS starter may intentionally leave one clear gap so `scorecard` and `diagnose` show value immediately.
- If a gap is intentional, make it specific and actionable, for example:
  - `BEHAVIOR-003 has no mapped assurance evidence`
  - primary bottleneck is `Assurance`
  - release recommendation is `Conditional` or equivalent warning state
- Do not generate a starter that looks production-ready if evidence is intentionally incomplete.
- Do not generate vague placeholder sections such as "Describe behavior here."

Test requirements:

- Generated SaaS intent contains subscription billing language.
- Generated SaaS behavior spec contains expected and unacceptable billing behaviors.
- Generated SaaS design artifact describes a concrete architecture flow.
- Generated SaaS assurance/security/execution artifacts parse as valid JSON if JSON is used.
- Generated SaaS evidence IDs can be traced.
- Generated SaaS scorecard output includes a primary bottleneck and next action.

### Task 1.3 - Add A Complete Passing SaaS Example Fixture

Goal: Give developers and tests a known-good reference project.

Add a complete passing SaaS fixture under the most appropriate existing location, such as:

```text
internal/traceability/testdata/complete-saas/
```

or:

```text
examples/saas-billing/
```

Choose the location that best matches existing repository conventions. If both are useful, keep the fixture small and avoid duplication.

Acceptance criteria:

- Fixture includes a realistic `bottleneck/` project structure.
- Running `bottleneck scorecard` against it returns a positive release recommendation.
- Running `bottleneck trace BEHAVIOR-001` shows a complete evidence chain.
- Running `bottleneck diagnose` shows no blocking bottleneck.
- Fixture is used in automated tests.

The complete passing fixture should include:

- valid intent evidence
- valid behavior evidence
- valid design evidence
- passing assurance evidence
- passing security evidence
- healthy execution telemetry
- complete refs between related IDs
- practical SaaS config thresholds

Test requirements:

- Add a test that validates the fixture passes `validator.NewEngine`.
- Add a test that renders scorecard output and verifies a positive release recommendation.
- Add a test that traces a key behavior ID and verifies related intent, design, assurance, security, and execution evidence.
- Add a test that diagnoses the fixture and verifies no blocking primary bottleneck.

## UX Expectations

The generated SaaS template should feel like a real SaaS team's first run.

Prefer language around:

- subscription billing
- payment method updates
- failed invoice retry
- duplicate charge prevention
- payment detail exposure prevention
- idempotency
- billing event telemetry
- payment provider tokenization

Avoid:

- generic framework descriptions
- abstract placeholder content
- unexplained BIASED jargon
- claims that evidence exists when related IDs are missing
- output that looks production-ready when a clear evidence gap remains

## Implementation Constraints

- Keep changes small and reviewable.
- Prefer adding tests before changing behavior.
- Do not remove or weaken existing tests.
- Do not rename existing public commands or output fields.
- Do not change default init behavior except to support an explicit template flag.
- Preserve safe file writing behavior.
- If adding helper functions, keep them close to init/template generation code.
- If adding fixtures, keep them small and readable.

## Suggested Implementation Plan

1. Inspect current `cmd/init.go` and init tests.
2. Add a template selector that defaults to current behavior.
3. Add SaaS template file contents as separate constants or a small template map.
4. Keep current no-overwrite behavior.
5. Add tests for default init, SaaS init, unsupported template, and no-overwrite behavior.
6. Add a complete passing SaaS fixture.
7. Add validator, scorecard, diagnose, and trace tests against the passing fixture.
8. Run the full Go test suite.

## Verification

Run:

```sh
go test ./...
```

Also manually exercise the new local flow from a temporary directory:

```sh
go run . init --template saas
go run . scorecard
go run . diagnose
go run . trace BEHAVIOR-001
```

If the SaaS starter intentionally includes an evidence gap, verify:

- `scorecard` clearly names the release recommendation.
- `diagnose` clearly names the primary bottleneck.
- the next action is specific to SaaS billing evidence.
- `trace` shows the missing link.

If the complete passing fixture is used, verify:

- scorecard returns a positive release recommendation.
- diagnose shows no blocking bottleneck.
- trace shows a complete evidence chain.

## Final Response Requirements

When finished, summarize:

- Files changed.
- Commands added or changed.
- SaaS template artifacts added.
- Fixture added.
- Tests added.
- Bugs found and fixed.
- Exact commands run and results.
- Any remaining gaps for later Day-One epics.

# AI Implementation Prompt: Add SaaS-Focused Release Gate Behavior

You are working in the `bottleneck` Go CLI repository.

Implement the ninth slice of the **SaaS Team Day-One Success** milestone.

## Milestone Goal

Make Bottleneck usable by a SaaS engineering team in the first 10 minutes.

A developer should be able to run a release gate and understand whether the release can proceed, is conditional, is blocked, or lacks enough evidence to decide.

This implementation prompt is scoped to:

- Epic 9: Add SaaS-Focused Release Gate Behavior
- Task 9.1: Define clear release recommendations
- Task 9.2: Make production blocking rules explicit
- Task 9.3: Add release gate fixture tests

## Definition Of Done For This Slice

A SaaS developer can run:

```sh
bottleneck scorecard --env=production
bottleneck diagnose --env=production --gate=release
```

And see a predictable release recommendation:

- `Proceed`
- `Conditional`
- `Blocked`
- `Insufficient Evidence`

Production gate behavior must be explicit, deterministic, and backed by fixture tests.

## Current Architecture To Respect

Inspect the repository before changing code. Follow existing patterns for:

- release gate behavior under `internal/gate/`
- scorecard recommendations under `internal/scorecard/`
- diagnosis gate behavior under `internal/diagnosis/`
- command flags under `cmd/scorecard.go` and `cmd/diagnose.go`
- environment thresholds and config loading
- validator and traceability findings
- existing fixture projects under `internal/**/testdata`
- existing release recommendation tests

Preserve backwards compatibility:

- Do not rename public commands or flags.
- Do not remove existing output formats.
- Do not rename JSON fields unless a failing test proves they are wrong.
- Do not weaken production blocking behavior.
- Do not make development behavior stricter than production.
- Do not hide supporting reasons when a release is blocked.
- Do not refactor scorecard, diagnosis, and gate packages broadly unless needed for clear testable rules.

Prefer small, explicit rule changes with focused tests.

## Primary Source Of Truth

Read:

- `tasks/DayOneExperience/saas-team-day-one-success-task-list.md`
- `readme.md`
- `docs/quickstart-saas.md`, if it exists
- `cmd/scorecard.go`
- `cmd/diagnose.go`
- `internal/gate/*`
- `internal/scorecard/*`
- `internal/diagnosis/*`
- `internal/validator/*`
- `internal/traceability/*`
- config loading and environment threshold code
- existing testdata fixtures

Keep SaaS examples consistent with:

- Subscription Billing Release
- payment method update
- failed invoice retry
- duplicate-charge prevention
- critical security findings
- stale telemetry
- required assurance evidence

## Epic 9: Add SaaS-Focused Release Gate Behavior

### Task 9.1 - Define Clear Release Recommendations

Goal: Make release guidance consistent.

Release recommendation values must be standardized as:

- `Proceed`
- `Conditional`
- `Blocked`
- `Insufficient Evidence`

Each recommendation must have clear rules.

Recommended rule semantics:

```text
Proceed
The release meets the effective environment threshold, required traceability is intact, no blocking security findings exist, required assurance exists, telemetry freshness requirements are satisfied, and there is enough evidence to make a release decision.

Conditional
The release has non-blocking gaps or warnings. It may proceed in lower environments or with accepted risk, but the scorecard must show what to fix next.

Blocked
The release has one or more blocking failures for the selected environment, especially production blockers.

Insufficient Evidence
Bottleneck cannot make a meaningful release recommendation because required evidence is missing or too thin across core categories.
```

Implementation guidance:

- Find the current recommendation values before changing them.
- If current values differ, add compatibility carefully:
  - prefer mapping internal statuses to the standardized user-facing values
  - preserve existing JSON fields
  - update tests and docs together
- Keep exact casing: `Proceed`, `Conditional`, `Blocked`, `Insufficient Evidence`.
- Keep recommendation rules deterministic.
- Ensure scorecard and diagnosis use the same recommendation vocabulary.
- Do not conflate `Conditional` with `Blocked`.
- Do not report `Proceed` when required production evidence is missing.

Test requirements:

- Tests verify each recommendation value can be produced.
- Tests verify each recommendation has a clear rule.
- Tests verify scorecard and diagnose agree on release recommendation where both expose it.
- Tests verify unsupported or legacy recommendation values do not leak into user-facing output unless retained for backwards compatibility in machine-readable fields.

### Task 9.2 - Make Production Blocking Rules Explicit

Goal: SaaS teams need predictable gates.

Production release is blocked when:

- critical security findings exist
- required behavior has no assurance evidence
- traceability is broken
- placeholder evidence is present in strict mode
- telemetry is stale or missing and required
- overall score is below production threshold

Implementation guidance:

- Implement or clarify production-specific gate checks in the package that already owns release decisions.
- Keep blocker reasons structured so scorecard and diagnosis can explain them.
- Blocking reasons should include relevant IDs when available:
  - `BEHAVIOR-*` for uncovered behavior
  - `SECURITY-*` or finding rule IDs for critical findings
  - `EXECUTION-*` for stale telemetry
  - broken reference IDs for traceability
- Respect environment configuration:
  - local/dev can warn for some gaps
  - production must block on critical release gaps
- Respect strict mode:
  - placeholder evidence in strict mode should block
  - placeholder evidence outside strict mode may warn if that is current behavior
- Use effective thresholds, not raw config fragments.

Recommended production blocker examples:

```text
Blocked: Critical security findings exist.
```

```text
Blocked: Required behavior BEHAVIOR-003 has no mapped assurance evidence.
```

```text
Blocked: Traceability is broken because BEHAVIOR-003 references ASSURANCE-003, but ASSURANCE-003 was not found.
```

```text
Blocked: Telemetry is stale and production requires fresh execution evidence.
```

```text
Blocked: Overall score 72 is below production threshold 85.
```

Do not add broad new telemetry, security, or traceability features if the signal is not modeled yet. Use existing modeled evidence and report any unsupported blocker as deferred.

Test requirements:

- Production blocks on critical security findings.
- Production blocks when required behavior has no assurance evidence.
- Production blocks when traceability is broken.
- Production blocks when placeholder evidence is present in strict mode.
- Production blocks when telemetry is stale or missing and required.
- Production blocks when overall score is below production threshold.
- Dev or local behavior is less strict where current config permits warnings.

### Task 9.3 - Add Release Gate Fixture Tests

Goal: Prove gate behavior.

Fixture tests must cover:

- passing release
- conditional release
- blocked by security
- blocked by assurance
- blocked by stale telemetry
- blocked by broken traceability
- insufficient evidence

Implementation guidance:

- Add small readable fixture projects under an appropriate testdata directory, for example:

```text
internal/gate/testdata/saas-release-gates/
```

- Reuse existing complete SaaS fixtures if they exist.
- Avoid huge duplicated fixture trees. Prefer fixture builder helpers if the package already uses them.
- Keep each fixture focused on one gate condition.
- Name fixtures by the behavior they prove:
  - `passing-release`
  - `conditional-release`
  - `blocked-security`
  - `blocked-assurance`
  - `blocked-stale-telemetry`
  - `blocked-traceability`
  - `insufficient-evidence`

Each fixture should contain the minimum realistic Bottleneck structure needed:

```text
bottleneck/
  config.yaml
  intent/intent.md
  behavior/behavior-spec.md
  design/architecture.md
  assurance/results.json
  security/guardrails.json
  execution/telemetry.json
```

Fixture expectations:

- Passing release:
  - complete evidence chain
  - no critical findings
  - fresh telemetry
  - production score meets threshold
  - recommendation `Proceed`
- Conditional release:
  - non-blocking warnings
  - no production blockers for the selected environment
  - recommendation `Conditional`
- Blocked by security:
  - critical security finding
  - recommendation `Blocked`
- Blocked by assurance:
  - required behavior lacks mapped assurance evidence
  - recommendation `Blocked`
- Blocked by stale telemetry:
  - telemetry stale or missing when required
  - recommendation `Blocked`
- Blocked by broken traceability:
  - broken ID reference or missing required traceability link
  - recommendation `Blocked`
- Insufficient evidence:
  - core evidence missing or too thin for a meaningful release decision
  - recommendation `Insufficient Evidence`

Test requirements:

- Use the same public APIs that scorecard or diagnose commands rely on where practical.
- Assert recommendation value.
- Assert at least one human-readable reason.
- Assert relevant evidence IDs when available.
- Assert production behavior explicitly for production blockers.
- Assert no fixture relies on placeholder text unless the fixture is testing placeholder behavior.

## UX Requirements

Release gate output should feel predictable and explainable.

Prefer:

- standardized recommendation words
- explicit blocker reasons
- relevant IDs
- environment-aware wording
- one primary recommendation with supporting reasons

Avoid:

- multiple competing recommendation names
- generic "failed gate" output without reason
- production passing with missing critical evidence
- broad refactors unrelated to release recommendations
- score changes made only to force desired wording

## Tests To Add Or Update

Add tests in the most appropriate package:

- release recommendation rule tests under `internal/gate/*_test.go` or `internal/scorecard/*_test.go`
- scorecard output tests under `internal/scorecard/*_test.go`
- diagnosis gate tests under `internal/diagnosis/*_test.go`
- CLI release gate tests under `cmd/*_test.go` if command output changes
- fixture tests under the package that owns gate evaluation

Minimum test coverage:

- `Proceed`, `Conditional`, `Blocked`, and `Insufficient Evidence` are produced by clear test cases.
- Production blocks for each required blocking rule.
- Release gate fixture tests cover all required fixture scenarios.
- Scorecard and diagnosis expose consistent recommendation language.
- Blocking reasons include relevant IDs where available.
- Existing release gate and scorecard tests still pass.

## Documentation Updates

Update docs only where needed:

- README or quickstart docs should use the standardized recommendation values.
- CI docs should explain that production release gates block on `Blocked`.
- If `docs/quickstart-saas.md` exists, explain:
  - `Proceed`
  - `Conditional`
  - `Blocked`
  - `Insufficient Evidence`
- Do not document rules that the code does not enforce.

## Verification Commands

Run:

```sh
go test ./...
```

If feasible, manually verify from a temporary directory:

```sh
bottleneck init --template saas
bottleneck scorecard --env=production
bottleneck diagnose --env=production --gate=release
```

For manual verification:

- Use a temporary directory.
- Do not create generated Bottleneck artifacts in the repository root.
- Record whether commands pass or fail.
- If production intentionally exits non-zero because the SaaS starter has an assurance gap, document that as expected.

## Final Response Requirements

When finished, report:

1. Release recommendation values and rules.
2. Production blocking rules implemented or clarified.
3. Fixture tests added or changed.
4. Scorecard and diagnosis output changes.
5. Documentation updates.
6. Exact commands run and results.
7. Any acceptance criteria intentionally deferred and why.


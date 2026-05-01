# AI Implementation Prompt: Improve Environment Defaults

You are working in the `bottleneck` Go CLI repository.

Implement the seventh slice of the **SaaS Team Day-One Success** milestone.

## Milestone Goal

Make Bottleneck usable by a SaaS engineering team in the first 10 minutes.

A developer should be able to run Bottleneck in local, dev, test, stage, and production contexts and understand why a scorecard passed, warned, or blocked release readiness.

This implementation prompt is scoped to:

- Epic 7: Improve Environment Defaults
- Task 7.1: Add practical SaaS environment thresholds
- Task 7.2: Show effective thresholds in scorecard
- Task 7.3: Add environment inheritance tests

## Definition Of Done For This Slice

A SaaS developer can run:

```sh
bottleneck scorecard --env=dev
bottleneck scorecard --env=production
bottleneck diagnose --env=production --gate=release
```

And immediately understand:

- Which environment is being evaluated.
- Which thresholds are effectively applied.
- Why development may warn while production blocks.
- How environment-specific config overrides inherited defaults.
- What to change when an environment name is invalid.

Environment behavior should be practical for SaaS delivery:

- local/dev should help developers without excessive blocking.
- test/stage should require stronger assurance and security evidence.
- production should block on critical release gaps.

## Current Architecture To Respect

Inspect the repository before changing code. Follow existing patterns for:

- config loading and default config generation
- environment flags under `cmd/`
- threshold models under `internal/scorecard/`, `internal/gate/`, or related packages
- release gate behavior
- scorecard rendering in text, JSON, and Markdown
- existing config fixtures and tests
- SaaS starter template config generation, if implemented

Preserve backwards compatibility:

- Do not rename `--env`.
- Do not remove existing environment names if users may already rely on them.
- Do not rename existing config keys or JSON fields unless a failing test proves they are wrong.
- Do not weaken production gate behavior.
- Do not make default local development behavior unexpectedly fail when it previously warned, unless the roadmap requires it and tests clearly cover it.
- Do not hide effective thresholds from detailed/JSON output if they already exist there.

Prefer small, explicit changes. Keep threshold resolution testable and deterministic.

## Primary Source Of Truth

Read:

- `tasks/DayOneExperience/saas-team-day-one-success-task-list.md`
- `readme.md`
- `docs/quickstart-saas.md`, if it exists
- `cmd/scorecard.go`
- `cmd/diagnose.go`
- `cmd/init.go`
- `internal/scorecard/*`
- `internal/gate/*`
- `internal/validator/*`
- config loading code and tests
- existing fixture projects under `internal/**/testdata`

Use the SaaS starter template as the primary configuration test case if it exists:

```sh
bottleneck init --template saas
```

## Epic 7: Improve Environment Defaults

### Task 7.1 - Add Practical SaaS Environment Thresholds

Goal: Make dev/test/stage/prod behavior understandable.

Default config must include:

- `local`
- `dev`
- `test`
- `stage`
- `production`

Behavior requirements:

- local/dev behavior produces helpful warnings.
- test/stage behavior uses stronger assurance and security requirements.
- production release gate blocks on critical gaps.

Implementation guidance:

- Update default config generation only where appropriate:
  - base/default init config
  - SaaS template config
  - example configs
- If the repository already has a default threshold section, extend it rather than replacing it.
- Use inheritance from defaults where supported.
- Keep threshold names consistent with existing config schema.
- Avoid changing score math unless threshold resolution already supports it.
- Production should be stricter than dev for:
  - minimum score
  - required traceability
  - critical security findings allowed
  - stale telemetry allowed
  - missing assurance or release evidence

Recommended conceptual thresholds:

```yaml
thresholds:
  default:
    minimum_score: 70
    required_traceability: false
    critical_security_findings_allowed: 0
    stale_telemetry_allowed: true
  local:
    minimum_score: 60
    required_traceability: false
    critical_security_findings_allowed: 1
    stale_telemetry_allowed: true
  dev:
    minimum_score: 70
    required_traceability: false
    critical_security_findings_allowed: 0
    stale_telemetry_allowed: true
  test:
    minimum_score: 75
    required_traceability: true
    critical_security_findings_allowed: 0
    stale_telemetry_allowed: true
  stage:
    minimum_score: 80
    required_traceability: true
    critical_security_findings_allowed: 0
    stale_telemetry_allowed: false
  production:
    minimum_score: 85
    required_traceability: true
    critical_security_findings_allowed: 0
    stale_telemetry_allowed: false
```

Adjust field names and values to match the existing config model. Do not invent schema keys if the code cannot read them; instead add narrowly scoped support with tests.

Test requirements:

- Default generated config includes local, dev, test, stage, and production.
- SaaS template config includes practical environment thresholds if the SaaS template exists.
- local/dev produces warnings for incomplete evidence where production would block.
- test/stage uses stricter assurance and security behavior than dev.
- production release gate blocks on critical gaps.

### Task 7.2 - Show Effective Thresholds In Scorecard

Goal: Make users understand why something passed or failed.

Scorecard output must show:

- `Environment: production`
- `Effective Thresholds`
- minimum score
- required traceability
- critical security findings allowed
- whether stale telemetry is allowed

Target shape:

```text
Environment: production
Effective Thresholds:
- Minimum score: 85
- Required traceability: true
- Critical security findings allowed: 0
- Stale telemetry allowed: false
```

Implementation guidance:

- Show effective thresholds after environment and before or near category results.
- Render resolved thresholds, not only raw environment overrides.
- Include inherited values so users do not need to manually merge config.
- Keep text output concise.
- Include equivalent threshold data in JSON output if scorecard JSON already exposes thresholds.
- Include equivalent threshold data in Markdown output if Markdown scorecard is used in CI summaries.
- Preserve deterministic field ordering.

If scorecard already has a detailed mode:

- Show thresholds in default output only if concise enough.
- Otherwise show a compact threshold block in default output and richer details in `--details`.

Test requirements:

- Text scorecard for `--env=production` includes the target threshold labels and values.
- JSON scorecard includes resolved/effective thresholds where supported.
- Markdown scorecard includes effective thresholds where supported.
- Output uses inherited defaults when an environment omits values.

### Task 7.3 - Add Environment Inheritance Tests

Goal: Prevent confusing config behavior.

Tests must verify:

- environment-specific values override defaults
- missing values inherit from defaults
- unknown environments fail with a helpful error
- production has stricter release behavior than dev

Implementation guidance:

- Add tests close to threshold resolution logic.
- If threshold resolution currently lives inside scorecard generation, consider extracting a small helper only if it reduces test complexity and follows local style.
- Avoid a broad config refactor.
- Use small in-memory config structs where possible.
- Use fixture config files only when file loading behavior matters.
- Helpful unknown environment error should name the invalid environment and list or hint at supported environments.

Recommended test cases:

1. `dev` overrides `minimum_score` but inherits `required_traceability`.
2. `production` overrides `required_traceability` and `stale_telemetry_allowed`.
3. `stage` inherits default critical security threshold when omitted.
4. `--env=not-real` fails with a useful error.
5. The same evidence produces warning or conditional behavior in dev and blocked behavior in production.

Do not implement unrelated environment features. Keep the focus on default environment behavior, threshold display, and inheritance correctness.

## UX Requirements

Environment behavior should be easy to explain:

- local/dev are for fast feedback.
- test/stage increase confidence before release.
- production enforces release readiness.

Prefer:

- concrete threshold labels
- clear environment names
- inherited effective values
- production gate wording that explains blockers
- examples in SaaS terms, such as payment retry assurance and critical security findings

Avoid:

- hidden defaults that only appear in config files
- ambiguous terms such as "strict enough"
- environment behavior that changes silently
- production passing with critical missing evidence
- noisy full config dumps in scorecard output

## Tests To Add Or Update

Add tests in the most appropriate package:

- config threshold resolution tests near config or scorecard code
- scorecard output tests under `internal/scorecard/*_test.go`
- CLI environment flag tests under `cmd/*_test.go`
- release gate environment behavior tests under `internal/gate/*_test.go`, if gate owns blocking behavior
- init config generation tests under `cmd/init_test.go`, if defaults are generated by init

Minimum test coverage:

- Generated default config includes `local`, `dev`, `test`, `stage`, and `production`.
- Effective thresholds resolve inheritance correctly.
- Environment-specific values override defaults.
- Missing values inherit from defaults.
- Unknown environments fail with a helpful error.
- Scorecard text output shows effective thresholds for production.
- Production release gate blocks on critical gaps.
- Dev behavior is less strict than production for the same non-critical evidence gap.
- Existing environment and scorecard tests still pass.

## Documentation Updates

Update docs only where needed:

- README or quickstart docs should mention `--env`.
- `docs/quickstart-saas.md`, if it exists, should explain local/dev/test/stage/production behavior briefly.
- CI docs should show production usage:

```sh
bottleneck scorecard --env=production --format=markdown
bottleneck diagnose --env=production --gate=release
```

- Avoid documenting config keys that the code cannot read.

## Verification Commands

Run:

```sh
go test ./...
```

If feasible, manually verify from a temporary directory:

```sh
bottleneck init --template saas
bottleneck scorecard --env=dev
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

1. Environment defaults changed.
2. Effective threshold display changes.
3. Environment inheritance behavior.
4. Tests added or changed.
5. Documentation updates.
6. Exact commands run and results.
7. Any acceptance criteria intentionally deferred and why.


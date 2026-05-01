# AI Implementation Prompt: SaaS Team Day-One Quickstart Flow

You are working in the `bottleneck` Go CLI repository.

Implement the second slice of the **SaaS Team Day-One Success** milestone.

## Milestone Goal

Make Bottleneck usable by a SaaS engineering team in the first 10 minutes.

A developer should be able to initialize Bottleneck, ingest common delivery evidence, run a scorecard, understand the primary bottleneck, and see exactly what to fix next without needing to understand the full BIASED framework first.

This implementation prompt is scoped to:

- Epic 2: Build The 10-Minute Quickstart Flow
- Task 2.1: Rewrite README quickstart around a SaaS team
- Task 2.2: Add a First 10 Minutes guide
- Task 2.3: Add expected output examples

## Definition Of Done For This Slice

A SaaS developer can open the repository README or the SaaS quickstart guide and immediately understand how to run:

```sh
bottleneck init --template saas
bottleneck validate
bottleneck scorecard
bottleneck diagnose
bottleneck trace
```

The docs must answer:

- What Bottleneck checks.
- What files Bottleneck creates.
- What each command tells the developer.
- What a good scorecard looks like.
- What a bad scorecard looks like.
- What to do when the primary bottleneck is Assurance.
- How Bottleneck would run in GitHub Actions.

Success test: a SaaS developer should not need the BIASED deck to understand why Bottleneck matters.

## Current Architecture To Respect

Inspect the repository before changing files. Follow existing conventions for:

- README structure and tone
- docs pages under `docs/`
- command names, flags, and output formats under `cmd/`
- existing examples under `examples/`
- existing tests for docs, CLI help, examples, or workflows

Preserve backwards compatibility:

- Do not rename public commands or flags.
- Do not change JSON schema fields or output field names just to make the docs easier.
- Do not invent commands that are not implemented.
- Do not claim ingestion or GitHub Actions behavior works unless the repository already supports it or this slice adds only documentation for an existing capability.
- If examples reference future milestone work, clearly label them as optional or upcoming.

Prefer documentation and tests over product behavior changes. Only change CLI behavior if a test exposes a clear bug in already implemented behavior and the smallest safe fix is necessary.

## Primary Source Of Truth

Read:

- `tasks/DayOneExperience/saas-team-day-one-success-task-list.md`
- `readme.md`
- existing files under `docs/`
- existing examples under `examples/`
- command implementations under `cmd/`
- relevant command tests under `cmd/*_test.go`

If Epic 1 has already been implemented, use the SaaS starter template behavior as the documented happy path:

```sh
bottleneck init --template saas
```

The SaaS sample domain should stay consistent with Epic 1:

- Subscription Billing Release
- Users can update payment methods and retry failed invoices
- `BEHAVIOR-003` represents payment retry duplicate-charge prevention
- The intentional quickstart bottleneck is missing mapped assurance evidence for `BEHAVIOR-003`

## Epic 2: Build The 10-Minute Quickstart Flow

### Task 2.1 - Rewrite README Quickstart Around A SaaS Team

Goal: Make the first page answer: "How do I use this on my app today?"

Update `readme.md` so the top-level quickstart is centered on a SaaS engineering team using Bottleneck on a real delivery capability.

Required quickstart commands:

```sh
bottleneck init --template saas
bottleneck validate
bottleneck scorecard
bottleneck diagnose
bottleneck trace BEHAVIOR-003
```

README requirements:

- Include `bottleneck init --template saas`.
- Include `bottleneck validate`.
- Include `bottleneck scorecard`.
- Include `bottleneck diagnose`.
- Include `bottleneck trace`.
- Explain what Bottleneck does in one short paragraph.
- Explain what files `init --template saas` creates.
- Explain what each command tells the developer.
- Show what a good scorecard looks like.
- Show what a bad scorecard looks like.
- Explain how to use Bottleneck in CI.

Recommended README shape:

1. One-paragraph product explanation.
2. "Start With A SaaS App" quickstart.
3. Generated files summary.
4. Command-by-command explanation.
5. Good scorecard example.
6. Bad scorecard example.
7. CI usage example.
8. Link to `docs/quickstart-saas.md` for the guided walkthrough.

Product explanation should be direct and practical. For example:

```text
Bottleneck is a CLI that diagnoses delivery risk from evidence in your repository: intent, behavior specs, design notes, test results, security findings, and execution telemetry. It turns those artifacts into a scorecard, release recommendation, primary bottleneck, and next action so a team can see what is blocking release readiness.
```

Avoid:

- Leading with abstract BIASED framework explanations.
- Describing Bottleneck as only a framework validator.
- Long conceptual setup before the user sees commands.
- Claiming a release is safe when evidence is missing.

### Task 2.2 - Add A First 10 Minutes Guide

Goal: Create a guided onboarding page separate from the full README.

Create:

```text
docs/quickstart-saas.md
```

The guide should be written as a hands-on walkthrough. Keep it concise enough that a developer can complete it in roughly 10 minutes.

Guide requirements:

- Walk through initializing the SaaS template.
- Walk through reviewing generated evidence.
- Walk through running validation.
- Walk through running scorecard.
- Walk through running diagnosis.
- Walk through breaking one evidence link.
- Walk through re-running diagnosis.
- Walk through fixing the evidence gap.
- Walk through adding or using the GitHub Actions workflow.
- Explain how to interpret release recommendation.

Recommended guide flow:

1. Create a temporary working folder.
2. Run `bottleneck init --template saas`.
3. Inspect generated evidence files:
   - `bottleneck/config.yaml`
   - `bottleneck/intent/intent.md`
   - `bottleneck/behavior/behavior-spec.md`
   - `bottleneck/design/architecture.md`
   - `bottleneck/assurance/results.json`
   - `bottleneck/security/guardrails.json`
   - `bottleneck/execution/telemetry.json`
4. Run `bottleneck validate`.
5. Run `bottleneck scorecard`.
6. Run `bottleneck diagnose`.
7. Run `bottleneck trace BEHAVIOR-003`.
8. Break one link, for example remove an assurance reference from a behavior or remove a `BEHAVIOR-003` result from assurance evidence.
9. Re-run `bottleneck diagnose`.
10. Fix the evidence gap by restoring the missing reference or adding mapped assurance evidence.
11. Re-run `bottleneck scorecard`.
12. Add or copy a GitHub Actions workflow if an example exists.
13. Interpret the release recommendation.

If a GitHub Actions SaaS workflow is not implemented yet, include a short CI section that uses currently supported commands and label a copy/paste SaaS workflow as part of the next milestone slice. Do not document non-existent files as if they already exist.

### Task 2.3 - Add Expected Output Examples

Goal: Developers should know whether the tool is working.

Add sample output blocks to the README and/or `docs/quickstart-saas.md` for:

- `bottleneck scorecard`
- `bottleneck scorecard --format=json`
- `bottleneck diagnose`
- `bottleneck trace BEHAVIOR-003`

The examples must clearly show:

- `Primary Bottleneck: Assurance`
- `BEHAVIOR-003 has no mapped test evidence` or equivalent clear wording
- `Next Action: Add assurance evidence for payment retry behavior`
- `Release Recommendation: Conditional`

If the current product output uses a different exact release recommendation term, choose one of these approaches:

- Prefer updating only docs to match actual current output if the command is already intentionally standardized.
- If the roadmap explicitly requires `Conditional` and the product has a small, safe mapping bug, add a focused failing test and the smallest fix.
- Do not rename broader release recommendation values without tests and without checking existing expectations.

Recommended text scorecard example:

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
BEHAVIOR-003 has no mapped test evidence for payment retry duplicate-charge prevention.

Next Action:
Add assurance evidence for payment retry behavior.
```

Recommended JSON scorecard example:

```json
{
  "schema_version": "scorecard.v1",
  "environment": "dev",
  "release_recommendation": "Conditional",
  "primary_bottleneck": "Assurance",
  "next_action": "Add assurance evidence for payment retry behavior.",
  "categories": [
    {
      "name": "Assurance",
      "status": "Warn",
      "missing_evidence": [
        "BEHAVIOR-003 has no mapped test evidence"
      ]
    }
  ]
}
```

Recommended diagnosis example:

```text
Primary Bottleneck: Assurance
Reason: BEHAVIOR-003 has no mapped test evidence for payment retry duplicate-charge prevention.
Impact: Release confidence is reduced because payment retry behavior is unproven.
Next Action: Add assurance evidence for payment retry behavior.
Inspect: bottleneck trace BEHAVIOR-003
```

Recommended trace example:

```text
Trace: BEHAVIOR-003

Related intent:
- INTENT-001

Related behavior evidence:
- BEHAVIOR-003 Payment retry duplicate-charge prevention

Missing links:
- No mapped assurance evidence for BEHAVIOR-003

Next Action:
Add assurance evidence for payment retry behavior.
```

Keep output examples realistic. They do not need to include every line from actual command output, but they must not contradict actual command behavior.

## Tests To Add Or Update

Add automated coverage for the documentation and quickstart flow where practical.

Prefer small tests in the most appropriate package:

- README and docs content tests can live in an existing docs-oriented test file, or a new root/package test if that is the repository convention.
- CLI quickstart behavior belongs in `cmd/*_test.go` if command helpers already support this.
- End-to-end quickstart flow tests can use temporary directories if existing tests already do this.

Minimum test coverage:

- README contains all required quickstart commands.
- README contains one-paragraph product language with delivery risk, evidence, scorecard, bottleneck diagnosis, and release readiness.
- README explains generated SaaS files.
- README includes good and bad scorecard examples.
- README includes CI usage guidance.
- `docs/quickstart-saas.md` exists.
- `docs/quickstart-saas.md` includes each required walkthrough step.
- Docs include sample output for:
  - `bottleneck scorecard`
  - `bottleneck scorecard --format=json`
  - `bottleneck diagnose`
  - `bottleneck trace`
- Output examples mention:
  - `Primary Bottleneck: Assurance`
  - `BEHAVIOR-003`
  - no mapped test evidence
  - `Next Action`
  - payment retry behavior
  - `Release Recommendation: Conditional`

If the repository does not currently have documentation tests, add one small test file rather than introducing a large testing framework.

Do not make tests brittle around full prose paragraphs unless the exact phrase is part of the acceptance criteria. Prefer checking required commands, headings, and high-signal phrases.

## CI Guidance To Document

The docs should show a minimal CI example using currently available commands.

Recommended shape:

```yaml
- name: Validate Bottleneck evidence
  run: bottleneck validate

- name: Generate Bottleneck scorecard
  run: bottleneck scorecard --format=markdown >> "$GITHUB_STEP_SUMMARY"

- name: Check release readiness
  run: bottleneck diagnose --gate=release
```

If GitHub annotation output is already implemented, mention:

```sh
bottleneck diagnose --format=github
```

If it is not implemented, do not claim it works.

## Implementation Constraints

- Keep changes focused on Epic 2.
- Prefer docs and tests over product behavior changes.
- Do not rewrite unrelated README sections unless needed for clarity.
- Do not remove existing commands or examples.
- Do not weaken existing claims that are still true.
- Do not add a full marketing landing page.
- Avoid unexplained framework jargon.
- Make the quickstart practical for a SaaS engineering team.
- Keep examples consistent with Subscription Billing Release.

## Verification Commands

Run:

```sh
go test ./...
```

If feasible, manually verify the documented flow from a temporary directory:

```sh
bottleneck init --template saas
bottleneck validate
bottleneck scorecard
bottleneck scorecard --format=json
bottleneck diagnose
bottleneck trace BEHAVIOR-003
```

For manual verification:

- Use a temporary directory.
- Do not create generated Bottleneck artifacts in the repository root.
- Record whether commands pass or fail.
- If a command intentionally exits non-zero because the starter has an Assurance gap, document that as expected.

## Final Response Requirements

When finished, report:

1. README sections changed.
2. New guide path.
3. Tests added or changed.
4. Any CLI behavior changes, if any.
5. Exact commands run and results.
6. Any acceptance criteria intentionally deferred and why.


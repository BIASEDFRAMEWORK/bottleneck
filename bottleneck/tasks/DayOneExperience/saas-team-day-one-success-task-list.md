# Next Milestone: SaaS Team Day-One Success

## Milestone Goal

Make Bottleneck usable by a SaaS engineering team in the first 10 minutes.

A developer should be able to initialize Bottleneck, ingest common delivery evidence, run a scorecard, understand the primary bottleneck, and see exactly what to fix next without needing to understand the full BIASED framework first.

## Target Folder

`tasks/DayOneExperience`

## Definition Of Done

This milestone is done when a new developer can run:

```sh
bottleneck init --template saas
bottleneck scorecard
bottleneck diagnose
bottleneck trace
```

And immediately understand:

- What Bottleneck is checking.
- What evidence is missing.
- What is blocking release readiness.
- What should be fixed next.
- How this would run in GitHub Actions.

Success test: a SaaS developer should not need the BIASED deck to understand why Bottleneck matters.

---

## Epic 1: Create A SaaS Starter Template

### Task 1.1 - Add `--template saas` To `bottleneck init`

Goal: Let a SaaS team initialize realistic starter artifacts.

Command:

```sh
bottleneck init --template saas
```

Acceptance criteria:

- [ ] Creates a SaaS-oriented Bottleneck project structure.
- [ ] Includes starter evidence for a recognizable SaaS product capability.
- [ ] Does not overwrite existing user files without confirmation or safe behavior.
- [ ] Existing `bottleneck init` behavior remains unchanged.
- [ ] Includes tests for default init and SaaS template init.

Suggested sample domain:

- Subscription Billing Release

Example feature:

- Users can update payment method and retry failed invoices.

### Task 1.2 - Replace Generic Placeholders With Realistic SaaS Examples

Goal: Make generated artifacts feel immediately understandable.

Acceptance criteria:

- [ ] Generated SaaS template includes `intent.md`.
- [ ] Generated SaaS template includes `behavior-spec.md`.
- [ ] Generated SaaS template includes `architecture.md`.
- [ ] Generated SaaS template includes assurance evidence, either `assurance.md` or normalized test evidence.
- [ ] Generated SaaS template includes security evidence, either `security.md` or normalized security evidence.
- [ ] Generated SaaS template includes execution evidence, either `execution.md` or normalized telemetry evidence.
- [ ] Generated SaaS template includes `config.yaml`.
- [ ] Generated artifacts include IDs such as `INTENT-001`, `BEHAVIOR-001`, `DESIGN-001`, `ASSURANCE-001`, `SECURITY-001`, and `EXECUTION-001`.

Example intent:

```text
INTENT-001: Customers must be able to update payment methods without duplicate charges, lost billing state, or exposure of payment details.
```

### Task 1.3 - Add A Complete Passing SaaS Example Fixture

Goal: Give developers and tests a known-good reference project.

Acceptance criteria:

- [ ] Adds a fixture project under `testdata`, `examples`, or `samples`.
- [ ] Running `bottleneck scorecard` against it returns a positive release recommendation.
- [ ] Running `bottleneck trace` shows a complete evidence chain.
- [ ] Running `bottleneck diagnose` shows no blocking bottleneck.
- [ ] Fixture is used in automated tests.

---

## Epic 2: Build The 10-Minute Quickstart Flow

### Task 2.1 - Rewrite README Quickstart Around A SaaS Team

Goal: Make the first page answer: "How do I use this on my app today?"

Acceptance criteria:

- [ ] README includes `bottleneck init --template saas`.
- [ ] README includes `bottleneck validate`.
- [ ] README includes `bottleneck scorecard`.
- [ ] README includes `bottleneck diagnose`.
- [ ] README includes `bottleneck trace`.
- [ ] README explains what Bottleneck does in one paragraph.
- [ ] README explains what files it creates.
- [ ] README explains what each command tells you.
- [ ] README shows what a good scorecard looks like.
- [ ] README shows what a bad scorecard looks like.
- [ ] README explains how to use it in CI.

### Task 2.2 - Add A First 10 Minutes Guide

Goal: Create a guided onboarding page separate from the full README.

Suggested file:

```text
docs/quickstart-saas.md
```

Acceptance criteria:

- [ ] Guide walks through initializing the SaaS template.
- [ ] Guide walks through reviewing generated evidence.
- [ ] Guide walks through running validation.
- [ ] Guide walks through running scorecard.
- [ ] Guide walks through running diagnosis.
- [ ] Guide walks through breaking one evidence link.
- [ ] Guide walks through re-running diagnosis.
- [ ] Guide walks through fixing the evidence gap.
- [ ] Guide walks through adding the GitHub Actions workflow.
- [ ] Guide explains how to interpret release recommendation.

### Task 2.3 - Add Expected Output Examples

Goal: Developers should know whether the tool is working.

Acceptance criteria:

- [ ] Docs include sample output for `bottleneck scorecard`.
- [ ] Docs include sample output for `bottleneck scorecard --format=json`.
- [ ] Docs include sample output for `bottleneck diagnose`.
- [ ] Docs include sample output for `bottleneck trace`.
- [ ] Output clearly shows `Primary Bottleneck: Assurance`.
- [ ] Output clearly explains `BEHAVIOR-003 has no mapped test evidence`.
- [ ] Output clearly shows `Next Action: Add assurance evidence for payment retry behavior`.
- [ ] Output clearly shows `Release Recommendation: Conditional`.

---

## Epic 3: Make Scorecard The Main Product Surface

### Task 3.1 - Make Scorecard The Primary Quickstart Command

Goal: Developers should see value immediately.

Acceptance criteria:

- [ ] README positions `scorecard` as the main command.
- [ ] CLI help positions `scorecard` as the main command.
- [ ] Scorecard output is easier to understand than raw validation output.
- [ ] Scorecard includes a clear release recommendation.
- [ ] Scorecard includes the primary bottleneck.
- [ ] Scorecard includes the next recommended action.

### Task 3.2 - Improve Plain-Text Scorecard Readability

Goal: Make terminal output instantly useful.

Acceptance criteria:

- [ ] Plain-text scorecard includes `Bottleneck Scorecard`.
- [ ] Plain-text scorecard includes `Environment: dev`.
- [ ] Plain-text scorecard includes `Release Recommendation: Conditional`.
- [ ] Plain-text scorecard includes `Primary Bottleneck: Assurance`.
- [ ] Plain-text scorecard includes concise category results.
- [ ] Plain-text scorecard includes a clear `Why` section.
- [ ] Plain-text scorecard includes a clear `Next Action` section.

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

### Task 3.3 - Add `--summary` Or Default Concise Mode

Goal: Avoid overwhelming new users.

Acceptance criteria:

- [ ] Default scorecard output is concise.
- [ ] Detailed evidence can still be shown with an existing or new flag, such as `bottleneck scorecard --details`.
- [ ] Tests verify default output is not overly verbose.
- [ ] Tests verify detailed output still includes evidence-level reasoning.

---

## Epic 4: Make Diagnosis Actionable

### Task 4.1 - Add Next Action To Diagnose

Goal: Developers should know what to do next.

Acceptance criteria:

- [ ] `bottleneck diagnose` output includes primary bottleneck.
- [ ] `bottleneck diagnose` output includes reason.
- [ ] `bottleneck diagnose` output includes impact.
- [ ] `bottleneck diagnose` output includes next action.
- [ ] `bottleneck diagnose` output includes relevant evidence IDs.
- [ ] `bottleneck diagnose` output includes a suggested command to inspect further.

Target shape:

```text
Primary Bottleneck: Assurance
Reason: BEHAVIOR-003 is not linked to any passing test evidence.
Impact: Release confidence is reduced because payment retry behavior is unproven.
Next Action: Add or ingest test evidence mapped to BEHAVIOR-003.
Inspect: bottleneck trace BEHAVIOR-003
```

### Task 4.2 - Add Diagnosis Rules For Common SaaS Bottlenecks

Goal: Make diagnosis feel relevant to SaaS teams.

Acceptance criteria:

- [ ] Diagnosis identifies and explains missing intent.
- [ ] Diagnosis identifies and explains behavior not mapped to intent.
- [ ] Diagnosis identifies and explains behavior not covered by tests.
- [ ] Diagnosis identifies and explains security findings blocking release.
- [ ] Diagnosis identifies and explains stale telemetry.
- [ ] Diagnosis identifies and explains missing production readiness evidence.
- [ ] Diagnosis identifies and explains thin placeholder artifacts.
- [ ] Diagnosis identifies and explains low traceability confidence.

### Task 4.3 - Add Test Cases For Clear Bottleneck Prioritization

Goal: Avoid noisy or contradictory diagnosis.

Acceptance criteria:

- [ ] Tests verify security blocker outranks stale telemetry.
- [ ] Tests verify missing intent outranks missing design.
- [ ] Tests verify missing assurance for critical behavior outranks minor docs gaps.
- [ ] Tests verify production release blockers are clearer than development warnings.
- [ ] Diagnosis returns one primary bottleneck plus supporting issues.

---

## Epic 5: Simplify Evidence Ingestion For Common SaaS Teams

### Task 5.1 - Add One-Command Local Evidence Ingestion Docs

Goal: Show developers how to bring in test, security, and telemetry evidence.

Acceptance criteria:

- [ ] Docs include `bottleneck ingest cucumber --file reports/cucumber.json`.
- [ ] Docs include `bottleneck ingest sarif --file reports/codeql.sarif`.
- [ ] Docs include `bottleneck ingest test-summary --file reports/test-summary.json`.
- [ ] Docs include `bottleneck ingest telemetry --file reports/telemetry.json`.
- [ ] Each example explains what the file represents.
- [ ] Each example explains where normalized evidence is written.
- [ ] Each example explains which scorecard category it updates.
- [ ] Each example explains how evidence IDs are linked.

### Task 5.2 - Add Sample SaaS Evidence Files

Goal: Let users run ingestion immediately.

Acceptance criteria:

- [ ] Add `examples/saas/reports/cucumber.json`.
- [ ] Add `examples/saas/reports/codeql.sarif`.
- [ ] Add `examples/saas/reports/test-summary.json`.
- [ ] Add `examples/saas/reports/telemetry.json`.
- [ ] Each sample maps to SaaS billing behavior.

### Task 5.3 - Add Ingestion Smoke Tests

Goal: Prove sample files work end-to-end.

Acceptance criteria:

- [ ] Tests run ingestion against sample files.
- [ ] Tests verify evidence files are created.
- [ ] Tests verify evidence IDs are preserved or generated.
- [ ] Tests verify scorecard category changes after ingestion.
- [ ] Tests verify invalid files fail cleanly.
- [ ] Tests verify `--dry-run` does not write files.

---

## Epic 6: Add A GitHub Actions Day-One Workflow

### Task 6.1 - Create A Copy/Paste SaaS Workflow

Goal: Make Bottleneck easy to try in CI.

Suggested file:

```text
examples/github-actions/bottleneck-saas-scorecard.yml
```

Acceptance criteria:

- [ ] Workflow runs `bottleneck validate`.
- [ ] Workflow runs `bottleneck scorecard --format=markdown`.
- [ ] Workflow runs `bottleneck diagnose --gate=release`.
- [ ] Workflow writes scorecard output to `$GITHUB_STEP_SUMMARY`.
- [ ] Workflow supports annotations if available through `bottleneck diagnose --format=github`.

### Task 6.2 - Document PR Comment And Step Summary Behavior

Goal: Show how this fits into pull requests.

Acceptance criteria:

- [ ] Docs explain how Bottleneck appears in GitHub Actions.
- [ ] Docs explain what developers see in a PR.
- [ ] Docs explain what blocks a release.
- [ ] Docs explain what only warns.
- [ ] Docs explain how to tune behavior by environment.

### Task 6.3 - Add Workflow Validation Tests

Goal: Prevent broken example workflows.

Acceptance criteria:

- [ ] Tests verify workflow YAML parses.
- [ ] Tests verify commands reference valid Bottleneck commands.
- [ ] Tests verify workflow does not require unavailable secrets.
- [ ] Tests verify workflow uses temporary or repo-safe paths.
- [ ] Tests verify workflow includes scorecard and diagnosis steps.

---

## Epic 7: Improve Environment Defaults

### Task 7.1 - Add Practical SaaS Environment Thresholds

Goal: Make dev/test/stage/prod behavior understandable.

Acceptance criteria:

- [ ] Default config includes `local`.
- [ ] Default config includes `dev`.
- [ ] Default config includes `test`.
- [ ] Default config includes `stage`.
- [ ] Default config includes `production`.
- [ ] Local/dev behavior produces helpful warnings.
- [ ] Test/stage behavior uses stronger assurance and security requirements.
- [ ] Production release gate blocks on critical gaps.

### Task 7.2 - Show Effective Thresholds In Scorecard

Goal: Make users understand why something passed or failed.

Acceptance criteria:

- [ ] Scorecard shows `Environment: production`.
- [ ] Scorecard shows `Effective Thresholds`.
- [ ] Scorecard shows minimum score.
- [ ] Scorecard shows required traceability.
- [ ] Scorecard shows critical security findings allowed.
- [ ] Scorecard shows whether stale telemetry is allowed.

Target shape:

```text
Environment: production
Effective Thresholds:
- Minimum score: 85
- Required traceability: true
- Critical security findings allowed: 0
- Stale telemetry allowed: false
```

### Task 7.3 - Add Environment Inheritance Tests

Goal: Prevent confusing config behavior.

Acceptance criteria:

- [ ] Tests verify environment-specific values override defaults.
- [ ] Tests verify missing values inherit from defaults.
- [ ] Tests verify unknown environments fail with a helpful error.
- [ ] Tests verify production has stricter release behavior than dev.

---

## Epic 8: Improve Error Messages And Help Text

### Task 8.1 - Rewrite Top-Level CLI Help

Goal: A new user should know what to run first.

Acceptance criteria:

- [ ] `bottleneck --help` includes start-here commands.
- [ ] Help briefly explains `validate`.
- [ ] Help briefly explains `scorecard`.
- [ ] Help briefly explains `diagnose`.
- [ ] Help briefly explains `trace`.
- [ ] Help briefly explains `ingest`.
- [ ] Help briefly explains `explain`.

Target start-here section:

```text
Start here:
  bottleneck init --template saas
  bottleneck scorecard
  bottleneck diagnose
```

### Task 8.2 - Add Helpful Empty-State Messages

Goal: Avoid dead-end errors.

Acceptance criteria:

- [ ] Missing evidence messages say what to do next.
- [ ] Missing assurance evidence suggests manual evidence or ingestion.

Target shape:

```text
No assurance evidence found.
Next action: Add test evidence manually or run:
  bottleneck ingest cucumber --file reports/cucumber.json
```

### Task 8.3 - Add Command Suggestion Tests

Goal: Keep UX intentional.

Acceptance criteria:

- [ ] Tests verify missing config includes next-step guidance.
- [ ] Tests verify missing evidence directory includes next-step guidance.
- [ ] Tests verify invalid environment includes next-step guidance.
- [ ] Tests verify broken evidence reference includes next-step guidance.
- [ ] Tests verify missing assurance includes next-step guidance.
- [ ] Tests verify invalid ingestion file includes next-step guidance.
- [ ] Tests verify placeholder-heavy artifact includes next-step guidance.

---

## Epic 9: Add SaaS-Focused Release Gate Behavior

### Task 9.1 - Define Clear Release Recommendations

Goal: Make release guidance consistent.

Acceptance criteria:

- [ ] Release recommendation values are standardized as `Proceed`, `Conditional`, `Blocked`, and `Insufficient Evidence`.
- [ ] Each recommendation has clear rules.
- [ ] Tests verify recommendation rules.

### Task 9.2 - Make Production Blocking Rules Explicit

Goal: SaaS teams need predictable gates.

Acceptance criteria:

- [ ] Production release is blocked when critical security findings exist.
- [ ] Production release is blocked when required behavior has no assurance evidence.
- [ ] Production release is blocked when traceability is broken.
- [ ] Production release is blocked when placeholder evidence is present in strict mode.
- [ ] Production release is blocked when telemetry is stale or missing and required.
- [ ] Production release is blocked when overall score is below production threshold.

### Task 9.3 - Add Release Gate Fixture Tests

Goal: Prove gate behavior.

Acceptance criteria:

- [ ] Fixture tests cover passing release.
- [ ] Fixture tests cover conditional release.
- [ ] Fixture tests cover blocked by security.
- [ ] Fixture tests cover blocked by assurance.
- [ ] Fixture tests cover blocked by stale telemetry.
- [ ] Fixture tests cover blocked by broken traceability.
- [ ] Fixture tests cover insufficient evidence.

---

## Epic 10: Package A Demo-Ready Example

### Task 10.1 - Add `examples/saas-billing`

Goal: Provide a complete example users can inspect and copy.

Suggested structure:

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

### Task 10.2 - Add Example Walkthrough

Goal: Let users reproduce the value quickly.

Acceptance criteria:

- [ ] Example README includes `cd examples/saas-billing`.
- [ ] Example README includes `bottleneck validate`.
- [ ] Example README includes `bottleneck scorecard`.
- [ ] Example README includes `bottleneck diagnose`.
- [ ] Example README includes `bottleneck trace`.
- [ ] Example README includes `bottleneck ingest cucumber --file reports/cucumber.json`.
- [ ] Example README includes `bottleneck ingest sarif --file reports/codeql.sarif`.
- [ ] Example README includes `bottleneck scorecard --env=production`.

### Task 10.3 - Add Intentional Failure Scenario

Goal: Show Bottleneck finding a real bottleneck.

Acceptance criteria:

- [ ] Example includes payment retry behavior.
- [ ] Intent and behavior are present.
- [ ] Assurance evidence is missing.
- [ ] `diagnose` identifies Assurance as the primary bottleneck.
- [ ] After ingesting test evidence, the bottleneck improves.

---

## Recommended Build Order

### Phase 1 - Make It Usable Locally

- [ ] Add `bottleneck init --template saas`.
- [ ] Add SaaS starter artifacts.
- [ ] Add complete passing SaaS fixture.
- [ ] Improve README quickstart.
- [ ] Improve default scorecard output.
- [ ] Add diagnosis next action.

### Phase 2 - Make It Useful In CI

- [ ] Add GitHub Actions SaaS workflow.
- [ ] Add Markdown step summary example.
- [ ] Add GitHub annotations documentation.
- [ ] Add workflow validation tests.

### Phase 3 - Make It Convincing

- [ ] Add `examples/saas-billing`.
- [ ] Add sample ingestion files.
- [ ] Add intentional failure scenario.
- [ ] Add end-to-end tests for init -> ingest -> scorecard -> diagnose -> trace.

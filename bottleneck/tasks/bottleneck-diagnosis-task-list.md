# Bottleneck Task List: Move From Framework Validation To Bottleneck Diagnosis

## Product Target

Move Bottleneck from checking whether framework files exist to diagnosing the primary delivery bottleneck.

The target first-run experience:

```sh
bottleneck init
bottleneck diagnose
```

Expected diagnosis:

```text
Primary Bottleneck: Assurance

Why:
You have defined intent and behavior, but no test or evaluation evidence proves the system behaves as intended.

Next action:
Add one assurance result mapped to BEHAVIOR-001.
```

## Epic 1: Make The CLI Diagnose The Primary Bottleneck

### Task 1.1: Add Primary Bottleneck Detection

**Goal:** `bottleneck scorecard` identifies the weakest category and explains why it matters.

**Work items:**

- [ ] Add scoring logic for each BIASED category.
- [ ] Score Behavior.
- [ ] Score Intent.
- [ ] Score Design.
- [ ] Score Assurance.
- [ ] Score Security.
- [ ] Score Execution.
- [ ] Calculate the lowest-scoring category.
- [ ] Return a clear primary bottleneck message.
- [ ] Handle ties when multiple categories are equally weak.
- [ ] Add tests for a single weakest category.
- [ ] Add tests for multiple tied bottlenecks.
- [ ] Add tests for all categories passing.
- [ ] Add tests for missing evidence files.

**Example output:**

```text
Primary Bottleneck: Assurance

Why:
Your system has defined intent and behavior, but no evidence proves that behavior was tested.

Next action:
Add assurance evidence that maps test or evaluation results to BEHAVIOR-001.
```

### Task 1.2: Add Why-This-Matters Explanations

**Goal:** Each bottleneck explains the delivery risk in plain language.

**Work items:**

- [ ] Create category explanation map.
- [ ] Add diagnosis copy for Behavior.
- [ ] Add diagnosis copy for Intent.
- [ ] Add diagnosis copy for Design.
- [ ] Add diagnosis copy for Assurance.
- [ ] Add diagnosis copy for Security.
- [ ] Add diagnosis copy for Execution.
- [ ] Include explanations in `scorecard`.
- [ ] Include explanations in `explain`.
- [ ] Keep language developer-friendly and light on framework jargon.

**Example:**

```text
Intent bottleneck:
The team has not clearly defined what good looks like. This creates downstream ambiguity in design, testing, security, and release decisions.
```

### Task 1.3: Add Recommended Next Action Per Diagnosis

**Goal:** Every finding produces one actionable next step.

**Work items:**

- [ ] Create recommendation rules per category.
- [ ] Add recommendations for missing evidence.
- [ ] Add recommendations for weak evidence.
- [ ] Add recommendations for stale evidence.
- [ ] Add recommendations for disconnected evidence.
- [ ] Prioritize one top action instead of listing every possible fix.
- [ ] Add tests for recommendation selection.

**Example:**

```text
Next action:
Replace placeholder intent statements with 1-3 measurable outcomes.
```

## Epic 2: Improve Scoring Beyond File Existence

### Task 2.1: Detect Placeholder Content

**Goal:** Prevent starter files from passing as meaningful evidence.

**Work items:**

- [ ] Detect placeholder phrase: `Describe required outcomes`.
- [ ] Detect placeholder phrase: `Describe system constraints`.
- [ ] Detect placeholder phrase: `TODO`.
- [ ] Detect placeholder phrase: `TBD`.
- [ ] Detect placeholder phrase: `Add measurable success criteria`.
- [ ] Penalize files that contain mostly placeholder text.
- [ ] Add warning output explaining the issue.
- [ ] Add tests for placeholder-heavy files.

**Example output:**

```text
Intent: Weak

Reason:
intent.md exists, but still contains starter placeholder content.
```

### Task 2.2: Add Thin-Evidence Scoring

**Goal:** A file should not pass simply because it exists.

**Work items:**

- [ ] Add minimum evidence requirements per file.
- [ ] Score based on content depth, not just presence.
- [ ] Check for expected sections.
- [ ] Check for IDs such as `INTENT-001`, `BEHAVIOR-001`, and related evidence IDs.
- [ ] Check for measurable language where appropriate.
- [ ] Add tests for empty files.
- [ ] Add tests for header-only files.
- [ ] Add tests for placeholder files.
- [ ] Add tests for meaningful files.

### Task 2.3: Add Traceability Scoring

**Goal:** Bottleneck detects when intent, behavior, assurance, security, and execution are disconnected.

**Work items:**

- [ ] Parse evidence IDs from Markdown.
- [ ] Parse evidence IDs from JSON.
- [ ] Validate references across files.
- [ ] Detect orphaned intent statements.
- [ ] Detect behavior specs with no assurance evidence.
- [ ] Detect security guardrails not mapped to behavior or intent.
- [ ] Detect execution metrics not tied to release readiness.
- [ ] Add score penalties for broken traceability.
- [ ] Add tests for missing relationships.
- [ ] Add tests for valid relationships.

**Example output:**

```text
Traceability Gap:
BEHAVIOR-001 exists, but no assurance result references it.
```

## Epic 3: Make The Scorecard Feel Like A Gauge

### Task 3.1: Redesign Terminal Scorecard Output

**Goal:** The CLI immediately shows where the system is weakest.

**Work items:**

- [ ] Add category bars or simple visual indicators.
- [ ] Sort categories by weakest first or highlight the weakest category.
- [ ] Add overall diagnosis.
- [ ] Add primary bottleneck section.
- [ ] Add next action section.
- [ ] Keep output readable in plain terminals.
- [ ] Keep output readable in CI logs.

**Example:**

```text
Bottleneck Scorecard

Primary Bottleneck: Assurance

Intent      [########--] 80
Behavior    [######----] 60
Design      [#####-----] 50
Assurance   [##--------] 20
Security    [#######---] 70
Execution   [######----] 60

Next action:
Map BEHAVIOR-001 to a passing BDD or evaluation result.
```

### Task 3.2: Add `bottleneck diagnose`

**Goal:** Create a command focused on diagnosis, not validation.

**Work items:**

- [ ] Add `bottleneck diagnose`.
- [ ] Output primary bottleneck.
- [ ] Output top 3 contributing findings.
- [ ] Output recommended next action.
- [ ] Output confidence level.
- [ ] Reuse existing validation and scorecard logic.
- [ ] Add tests for command behavior.

**Example:**

```text
Primary Bottleneck: Behavior

Contributing findings:
1. No unacceptable behaviors defined.
2. Behavior spec does not reference INTENT-001.
3. No evaluation evidence found for expected behavior.

Recommended next action:
Define expected and unacceptable behavior for each intent statement.
```

### Task 3.3: Add Confidence Level To Diagnosis

**Goal:** Show whether the diagnosis is based on strong evidence or limited evidence.

**Work items:**

- [ ] Add confidence level `High`.
- [ ] Add confidence level `Medium`.
- [ ] Add confidence level `Low`.
- [ ] Base confidence on number of evidence files present.
- [ ] Base confidence on traceability completeness.
- [ ] Base confidence on amount of meaningful content.
- [ ] Base confidence on evidence recency when timestamps exist.
- [ ] Add tests for confidence calculation.

**Example:**

```text
Diagnosis Confidence: Low

Reason:
Only 2 of 6 evidence categories contain meaningful content.
```

## Epic 4: Improve The Hello World Experience

### Task 4.1: Replace Generic Sample Files With A Realistic Sample App

**Goal:** The starter example demonstrates diagnosis value.

**Work items:**

- [ ] Replace generic placeholder samples with a tiny use case.
- [ ] Use `AI PDF Risk Summarizer` as the recommended sample unless another sample is chosen.
- [ ] Include realistic intent evidence.
- [ ] Include realistic behavior evidence.
- [ ] Include realistic assurance evidence.
- [ ] Include realistic security evidence.
- [ ] Include realistic execution evidence.
- [ ] Make one category intentionally weak to demonstrate diagnosis.
- [ ] Update README walkthrough.

**Recommended sample:**

```text
AI PDF Risk Summarizer

Intent:
The system must summarize financial PDFs without omitting material risk clauses.

Behavior:
The system must flag uncertainty when financial risk language is ambiguous.

Assurance:
One evaluation fails because ambiguous risk language was summarized as fact.

Primary Bottleneck:
Assurance
```

### Task 4.2: Update `bottleneck init` To Create Richer Starter Artifacts

**Goal:** `init` creates useful starter evidence, not empty framework templates.

**Work items:**

- [ ] Generate `INTENT-001`.
- [ ] Generate `BEHAVIOR-001`.
- [ ] Generate `DESIGN-001`.
- [ ] Generate `ASSURANCE-001`.
- [ ] Generate `SECURITY-001`.
- [ ] Generate `EXECUTION-001`.
- [ ] Add comments explaining what each section should contain.
- [ ] Add one intentionally incomplete section so `diagnose` has something to find.
- [ ] Ensure checked-in sample files match generated files.

### Task 4.3: Add A Guided First-Run Message

**Goal:** Help developers understand what to do after `init`.

**Work items:**

- [ ] After `bottleneck init`, print `bottleneck validate`.
- [ ] After `bottleneck init`, print `bottleneck scorecard`.
- [ ] After `bottleneck init`, print `bottleneck diagnose`.
- [ ] Explain that the initial project is intentionally incomplete.
- [ ] Tell the developer which file to edit first.

**Example:**

```text
Bottleneck initialized.

Next:
1. Run: bottleneck diagnose
2. Review the primary bottleneck
3. Replace placeholder intent with measurable outcomes in bottleneck/intent/intent.md
```

## Epic 5: Clarify Product Scope And Language

### Task 5.1: Clarify Framework Vs Product In README

**Goal:** Make the separation obvious.

**Work items:**

- [ ] Add near the top: `BIASED is the evidence model.`
- [ ] Add near the top: `Bottleneck is the CLI that diagnoses delivery risk using that model.`
- [ ] Remove language that makes Bottleneck sound like only a framework validator.
- [ ] Emphasize diagnosis.
- [ ] Emphasize release readiness.
- [ ] Emphasize hidden delivery risk.

### Task 5.2: Clarify What Bottleneck Evaluates

**Goal:** Reduce confusion about repo, team, release, and organization scope.

**Work items:**

- [ ] Add scope statement: Bottleneck evaluates a repo or release using local evidence artifacts.
- [ ] Explain that team and organization views can come later by aggregating repo scorecards.
- [ ] Add example: single application repo.
- [ ] Add example: service repo.
- [ ] Add example: AI feature repo.
- [ ] Add example: platform repo.

### Task 5.3: Clarify AI Vs Non-AI Positioning

**Goal:** Make the tool useful beyond AI without losing the AI urgency.

**Work items:**

- [ ] Add positioning statement for any software system.
- [ ] Explain why AI-enabled systems especially benefit from behavior, drift, evaluation, and governance evidence.
- [ ] Add one AI example.
- [ ] Add one non-AI example.
- [ ] Avoid positioning the product as limited to LLM apps only.

## Epic 6: Strengthen GitHub Actions Integration

### Task 6.1: Add PR Comment Output Mode

**Goal:** Make Bottleneck useful inside pull requests.

**Work items:**

- [ ] Add `--format markdown` to `scorecard` or `diagnose`.
- [ ] Generate PR-friendly Markdown.
- [ ] Include primary bottleneck.
- [ ] Include category scores.
- [ ] Include top findings.
- [ ] Include recommended next action.
- [ ] Add snapshot test for Markdown output.

**Example command:**

```sh
bottleneck diagnose --format markdown
```

### Task 6.2: Add GitHub Annotation Output

**Goal:** Surface findings directly in CI.

**Work items:**

- [ ] Add `--format github`.
- [ ] Emit GitHub Actions warning annotations.
- [ ] Emit GitHub Actions error annotations.
- [ ] Include file path when known.
- [ ] Include line number when known.
- [ ] Support warning vs error severity.
- [ ] Add tests for annotation format.

**Example annotation:**

```text
::warning file=bottleneck/intent/intent.md,line=1::Intent evidence contains placeholder content
```

### Task 6.3: Add Release Gate Mode

**Goal:** Allow teams to block releases based on diagnosis severity.

**Work items:**

- [ ] Add `bottleneck diagnose --gate release`.
- [ ] Add configurable thresholds in `config.yaml`.
- [ ] Fail build when primary bottleneck score is below threshold.
- [ ] Fail build when a required category is missing.
- [ ] Fail build when traceability is broken.
- [ ] Fail build when security evidence fails.
- [ ] Fail build when governance evidence fails.
- [ ] Add tests for passing release gates.
- [ ] Add tests for failing release gates.

## Epic 7: Add Evidence Ingestion From Existing Tools

### Task 7.1: Support BDD/Cucumber Result Ingestion

**Goal:** Assurance scores come from actual test output.

**Work items:**

- [ ] Support Cucumber JSON result files.
- [ ] Map scenarios to behavior IDs through tags such as `@BEHAVIOR-001`.
- [ ] Update Assurance score based on pass/fail results.
- [ ] Detect behavior specs with no matching scenario.
- [ ] Add sample Cucumber result fixture.
- [ ] Add tests.

**Example scenario tag:**

```gherkin
@BEHAVIOR-001
Scenario: Ambiguous risk clause is flagged
```

### Task 7.2: Support CodeQL Or Security Scan Ingestion

**Goal:** Security scores reflect real scan evidence.

**Work items:**

- [ ] Support SARIF input.
- [ ] Count findings by severity.
- [ ] Map findings to security score.
- [ ] Add configuration for severity thresholds.
- [ ] Add tests with sample SARIF files.

### Task 7.3: Support Telemetry Evidence Ingestion

**Goal:** Execution scores become more meaningful.

**Work items:**

- [ ] Define basic telemetry schema.
- [ ] Support deployment frequency.
- [ ] Support change failure rate.
- [ ] Support error rate.
- [ ] Support user override rate.
- [ ] Support adoption rate.
- [ ] Support cost signals.
- [ ] Penalize missing telemetry.
- [ ] Penalize stale telemetry.
- [ ] Add tests for telemetry scoring.

## Epic 8: Improve Explainability

### Task 8.1: Make `bottleneck explain` Evidence-Driven

**Goal:** `explain` shows how the score was derived.

**Work items:**

- [ ] Show evidence found for each category.
- [ ] Show evidence missing for each category.
- [ ] Show score impact for each category.
- [ ] Show related IDs for each category.
- [ ] Show recommendation for each category.
- [ ] Avoid generic framework descriptions.
- [ ] Add tests for explanation output.

**Example:**

```text
Assurance Score: 20

Evidence found:
- assurance/results.json exists

Evidence missing:
- No result references BEHAVIOR-001
- No failed scenario explanation found
- No drift or regression evidence found

Score impact:
-40 broken traceability
-20 missing evaluation evidence
-20 thin assurance evidence
```

### Task 8.2: Add `bottleneck trace --id`

**Goal:** Developers can inspect one intent, behavior, or finding end-to-end.

**Work items:**

- [ ] Support `bottleneck trace --id INTENT-001`.
- [ ] Support `bottleneck trace --id BEHAVIOR-001`.
- [ ] Show related behavior evidence.
- [ ] Show related design evidence.
- [ ] Show related assurance evidence.
- [ ] Show related security evidence.
- [ ] Show related execution evidence.
- [ ] Highlight missing links.
- [ ] Add tests.

## Suggested Implementation Order

### Sprint 1: Diagnosis Core

- [ ] Add primary bottleneck detection.
- [ ] Add why-this-matters explanations.
- [ ] Add next-action recommendations.
- [ ] Redesign scorecard output.
- [ ] Add tests for diagnosis logic.

### Sprint 2: Evidence Quality

- [ ] Detect placeholder content.
- [ ] Add thin-evidence scoring.
- [ ] Add traceability scoring.
- [ ] Improve `explain`.
- [ ] Update sample files.

### Sprint 3: Developer Experience

- [ ] Improve `bottleneck init`.
- [ ] Replace Hello World sample with realistic use case.
- [ ] Add `bottleneck diagnose`.
- [ ] Add guided first-run messaging.
- [ ] Update README.

### Sprint 4: CI/CD Value

- [ ] Add Markdown output.
- [ ] Add GitHub annotation output.
- [ ] Add release gate mode.
- [ ] Add PR comment example workflow.

### Sprint 5: Real Evidence Integrations

- [ ] Add Cucumber/BDD ingestion.
- [ ] Add SARIF/CodeQL ingestion.
- [ ] Add telemetry ingestion.
- [ ] Add trace-by-ID command.

## Definition Of Done For This Iteration

- [ ] `bottleneck init` creates a meaningful starter project with an intentional diagnosis target.
- [ ] `bottleneck diagnose` returns a primary bottleneck.
- [ ] Diagnosis includes a plain-language `Why`.
- [ ] Diagnosis includes one recommended next action.
- [ ] Scorecard shows category scores and the weakest category.
- [ ] Explain output shows evidence found, evidence missing, score impact, related IDs, and recommendations.
- [ ] Placeholder, thin evidence, and broken traceability affect scores.
- [ ] GitHub-friendly Markdown and annotation output exist.
- [ ] Release gate mode can fail CI based on diagnosis severity.
- [ ] Cucumber, SARIF, and telemetry ingestion produce normalized artifacts.
- [ ] Tests cover scoring, diagnosis, evidence quality, GitHub output, gates, ingestion, and traceability.
- [ ] Existing `validate`, `explain`, and `scorecard` behavior remains backward compatible unless strict mode is enabled.
- [ ] Every new score or warning links back to an artifact, threshold, or ingested evidence source.

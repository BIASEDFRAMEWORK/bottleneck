# AI Implementation Prompt: Epic 4 Hello World Experience

You are working in the `bottleneck` Go CLI repository.

Implement **Epic 4: Improve The Hello World Experience**.

This epic covers:

- Task 4.1: Replace generic sample files with a realistic sample app
- Task 4.2: Update `bottleneck init` to create richer starter artifacts
- Task 4.3: Add a guided first-run message

## Product Goal

Make the first run demonstrate Bottleneck's value immediately.

After a developer runs `bottleneck init`, the generated starter evidence should be realistic enough for `bottleneck diagnose` to identify a primary bottleneck. The starter should not feel like empty framework scaffolding.

Recommended starter use case:

```text
AI PDF Risk Summarizer
```

Target diagnosis:

```text
Primary Bottleneck: Assurance

Why:
The system has defined intent and behavior, but assurance evidence shows ambiguous risk language was summarized as fact.

Next action:
Add or fix evaluation evidence for BEHAVIOR-001 so ambiguous financial risk language is flagged as uncertain.
```

## Current Architecture To Respect

Use the existing `init` command and artifact layout.

Relevant files:

- `cmd/init.go`
  - Defines `initDirectories`, `initFiles`, starter artifact constants, and post-init output.
- `cmd/diagnose.go`
  - If present, generated starter artifacts should produce a useful diagnosis.
- `internal/validator/*`
  - Generated artifacts must remain valid or intentionally weak in a controlled way.
- `internal/traceability` or `internal/validator/traceability.go`
  - Generated IDs and refs should be coherent.
- `readme.md`
  - Update walkthrough.
- `bottleneck/docs/validation.md`
  - Update generated docs if starter behavior or artifact examples change.
- checked-in sample artifacts under the current workspace, if they are intended to mirror `init`.

The current `init` may already generate IDs such as `INTENT-001`, `BEHAVIOR-001`, and related refs. Replace generic placeholder copy with realistic sample evidence while preserving those ID conventions.

## Required Behavior

### 1. Replace Generic Starter Files With A Realistic Sample App

Use this sample unless there is a strong existing product reason to choose another:

```text
AI PDF Risk Summarizer
```

The sample should describe a tiny AI-assisted workflow:

- A user uploads or provides a financial PDF.
- The system summarizes material risk clauses.
- The system must flag uncertainty when risk language is ambiguous.
- The system should not present ambiguous risk language as fact.

Generated starter artifacts should include realistic evidence for:

- Intent
- Behavior
- Design
- Assurance
- Security
- Execution

Make one category intentionally weak so diagnosis has something useful to find.

Recommended weak category:

```text
Assurance
```

Use this failure example:

```text
One evaluation fails because ambiguous risk language was summarized as fact.
```

### 2. Generate Richer Intent Evidence

Update generated `bottleneck/intent/intent.md`.

Requirements:

- Include `INTENT-001`.
- Include realistic outcomes.
- Include constraints.
- Include measurable success criteria.
- Link to `BEHAVIOR-001`.
- Avoid empty placeholder phrases such as `Describe required outcomes`.

Recommended content direction:

```markdown
# Intent

## Outcomes

### INTENT-001: Summarize financial PDF risk without hiding uncertainty
Refs:
- BEHAVIOR-001

The system must summarize material risk clauses from financial PDFs while preserving uncertainty and caveats that affect release or investment decisions.

## Constraints

- The system must not invent risk facts that are not present in the source PDF.
- The system must flag ambiguous or qualified risk language instead of rewriting it as certainty.

## Success Criteria

- At least 95% of evaluated summaries preserve material risk caveats.
- 100% of ambiguous risk clauses in the evaluation set are flagged as uncertain.
```

### 3. Generate Richer Behavior Evidence

Update generated `bottleneck/behavior/behavior-spec.md`.

Requirements:

- Include `BEHAVIOR-001`.
- Mark it critical.
- Reference `INTENT-001`.
- Reference `ASSURANCE-001`.
- Include expected behavior.
- Include unacceptable behavior.

Recommended behavior:

```markdown
# Behavior Specification

## Expected Behavior

### BEHAVIOR-001: Flag ambiguous financial risk language
Critical: true
Refs:
- INTENT-001
- ASSURANCE-001

When a PDF contains qualified risk language such as "may", "could", "subject to", or "material uncertainty", the summary must preserve that uncertainty and flag it for review.

## Unacceptable Behavior

- The system must not summarize ambiguous risk language as a confirmed fact.
- The system must not omit material caveats from the risk summary.
```

### 4. Generate Richer Design Evidence

Update generated `bottleneck/design/architecture.md`.

Requirements:

- Include `DESIGN-001`.
- Reference intent and behavior.
- Describe a small architecture.
- Keep it realistic but brief.

Recommended design:

```markdown
# Architecture

### DESIGN-001: Local PDF risk summarization flow
Refs:
- INTENT-001
- BEHAVIOR-001

The workflow extracts PDF text, identifies candidate risk clauses, asks the summarizer to produce a short risk summary, and runs a post-summary uncertainty check before showing the output to a reviewer.

Key components:
- PDF text extraction
- Risk clause detector
- Summary generator
- Uncertainty flagging check
- Reviewer-facing output
```

### 5. Generate Intentionally Weak Assurance Evidence

Update generated `bottleneck/assurance/results.json` and `bottleneck/assurance/features/sample.feature`.

Requirements:

- Include `ASSURANCE-001`.
- Reference `BEHAVIOR-001`.
- Include at least one failing scenario.
- Keep the failure realistic and aligned with the sample.
- Ensure `bottleneck diagnose` can identify Assurance as the primary bottleneck.

Recommended JSON direction:

```json
{
  "scenarios_total": 2,
  "scenarios_passed": 1,
  "scenarios_failed": 1,
  "failures": [
    "Ambiguous risk clause was summarized as confirmed exposure"
  ],
  "evidence": [
    {
      "id": "ASSURANCE-001",
      "refs": ["BEHAVIOR-001"],
      "source": "sample evaluation",
      "status": "fail",
      "summary": "One evaluation failed because ambiguous risk language was summarized as fact."
    }
  ]
}
```

Recommended feature direction:

```gherkin
Feature: AI PDF risk summarization

  @BEHAVIOR-001
  Scenario: Ambiguous risk clause is flagged
    Given a financial PDF says exposure "may be material subject to market conditions"
    When the system summarizes the risk clause
    Then the summary should flag the exposure as uncertain

  @BEHAVIOR-001
  Scenario: Ambiguous risk clause is not stated as fact
    Given a financial PDF uses qualified risk language
    When the system produces a risk summary
    Then the summary should not state the risk as confirmed exposure
```

### 6. Generate Realistic Security Evidence

Update generated `bottleneck/security/guardrails.json`.

Requirements:

- Include `SECURITY-001`.
- Reference behavior or intent.
- Keep violations at `0` unless intentionally choosing Security as the weak category.
- Mention practical guardrails such as source-grounded summaries and no unsupported claims.

Recommended shape:

```json
{
  "violations": 0,
  "evidence": [
    {
      "id": "SECURITY-001",
      "refs": ["INTENT-001", "BEHAVIOR-001"],
      "source": "sample guardrail review",
      "status": "pass",
      "summary": "Guardrails require source-grounded summaries and prohibit unsupported risk claims."
    }
  ]
}
```

### 7. Generate Realistic Execution Evidence

Update generated `bottleneck/execution/telemetry.json`.

Requirements:

- Include `EXECUTION-001`.
- Reference behavior and assurance.
- Include adoption and error rate.
- Keep values realistic.
- Do not make Execution the primary bottleneck unless intentionally chosen.

Recommended shape:

```json
{
  "adoption_rate": 0.72,
  "error_rate": 0.02,
  "source_environment": "sample",
  "evidence": [
    {
      "id": "EXECUTION-001",
      "refs": ["BEHAVIOR-001", "ASSURANCE-001"],
      "source": "sample telemetry",
      "status": "pass",
      "summary": "Sample pilot telemetry shows reviewers used the summary flow with low processing errors."
    }
  ]
}
```

### 8. Add Comments Explaining What Each Section Should Contain

Add concise Markdown comments where useful in generated Markdown files.

Guidelines:

- Comments should help users edit the starter artifacts.
- Comments should not pollute diagnosis scoring as placeholder content.
- Prefer short HTML comments.

Example:

```markdown
<!-- Replace this sample outcome with the release outcome your system must prove. -->
```

Do not over-comment JSON files if comments would make them invalid.

### 9. Ensure Checked-In Sample Files Match Generated Files

If this repository includes checked-in sample artifacts under `bottleneck/`, update them to match the new `init` output.

This keeps:

- README walkthrough
- local examples
- generated starter project

aligned.

### 10. Add Guided First-Run Message

After `bottleneck init`, print clear next steps.

Required output:

```text
Bottleneck initialized.

Next:
1. Run: bottleneck diagnose
2. Review the primary bottleneck
3. Replace sample intent and behavior with evidence from your own system
```

Also include a direct pointer to the first file to edit.

Recommended:

```text
Start with: bottleneck/intent/intent.md
```

Mention that the starter project is intentionally incomplete or intentionally weak so the diagnosis has something to show.

Example:

```text
Bottleneck initialized.

This starter uses the AI PDF Risk Summarizer sample and intentionally leaves Assurance weak so diagnose can show a real bottleneck.

Next:
1. Run: bottleneck diagnose
2. Review the primary bottleneck
3. Replace sample intent and behavior with evidence from your own system

Start with: bottleneck/intent/intent.md
```

### 11. Update README Walkthrough

Update `readme.md` with a short first-run walkthrough:

```sh
bottleneck init
bottleneck diagnose
bottleneck scorecard
```

Explain:

- The generated sample is `AI PDF Risk Summarizer`.
- Assurance is intentionally weak.
- The user should replace sample artifacts with project-specific evidence.
- The first file to edit is `bottleneck/intent/intent.md`.

Keep README language focused on diagnosis and release readiness, not only framework validation.

## Backward Compatibility

- `bottleneck init` should still avoid overwriting existing files unless the existing command already supports overwrite.
- Existing artifact paths should remain under `bottleneck/`.
- Generated JSON must remain valid.
- Generated Markdown must still satisfy structural validation.
- The starter should not require network access or external tools.
- Existing `validate`, `scorecard`, `explain`, and `diagnose` commands should work on the starter project.

## Testing Requirements

Add focused Go tests.

Required test cases:

1. `initializeProject` creates realistic AI PDF Risk Summarizer artifacts.
2. Generated intent includes `INTENT-001`.
3. Generated behavior includes `BEHAVIOR-001`.
4. Generated design includes `DESIGN-001`.
5. Generated assurance includes `ASSURANCE-001`.
6. Generated security includes `SECURITY-001`.
7. Generated execution includes `EXECUTION-001`.
8. Generated assurance has one intentional failure.
9. Generated JSON files parse successfully.
10. Generated Markdown files satisfy required sections.
11. Existing files are not overwritten.
12. Init output includes `bottleneck validate`, `bottleneck scorecard`, and `bottleneck diagnose`.
13. Init output explains the starter is intentionally weak.
14. Running validation or diagnosis against the generated starter identifies Assurance as weak or failing.

Run:

```sh
go test ./...
```

## Implementation Guidance

Recommended approach:

1. Replace starter constants in `cmd/init.go` with AI PDF Risk Summarizer sample artifacts.
2. Keep artifact IDs and refs consistent across Markdown and JSON.
3. Make Assurance intentionally weak through `scenarios_failed: 1`.
4. Update `initializeProject` tests or add command tests around generated files.
5. Update checked-in sample artifacts if the repository uses them as examples.
6. Update README first-run walkthrough.
7. Run formatting and tests.

Keep the sample small. It should be realistic enough to demonstrate diagnosis, but not so large that users have to read a full application spec before understanding the product.

## Acceptance Criteria

- `bottleneck init` creates a realistic AI PDF Risk Summarizer starter.
- Starter artifacts include `INTENT-001`, `BEHAVIOR-001`, `DESIGN-001`, `ASSURANCE-001`, `SECURITY-001`, and `EXECUTION-001`.
- Starter artifacts include realistic intent, behavior, design, assurance, security, and execution evidence.
- Assurance is intentionally weak so diagnosis has a useful bottleneck to report.
- The first-run message tells users to run `bottleneck validate`, `bottleneck scorecard`, and `bottleneck diagnose`.
- The first-run message explains the sample is intentionally incomplete or weak.
- README includes a first-run walkthrough.
- Checked-in sample files match generated starter artifacts where applicable.
- Tests cover generated artifacts, IDs, JSON validity, non-overwrite behavior, and first-run output.
- `go test ./...` passes.

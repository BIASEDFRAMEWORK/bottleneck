# AI Implementation Prompt: Epic 5 Product Scope And Language

You are working in the `bottleneck` Go CLI repository.

Implement **Epic 5: Clarify Product Scope And Language**.

Also include the related CI-facing task:

- Epic 6, Task 6.1: Add PR comment output mode

## Product Goal

Make Bottleneck's positioning obvious in the README and user-facing Markdown output.

Readers should quickly understand:

- BIASED is the evidence model.
- Bottleneck is the CLI that diagnoses delivery risk using that model.
- Bottleneck evaluates a repo or release using local evidence artifacts.
- Bottleneck works for any software system, but AI-enabled systems especially need this kind of evidence because behavior, drift, evaluation, and governance cannot be inferred from code alone.

## Current Architecture To Respect

This epic is mostly documentation and output-language work. Keep changes focused and avoid broad product rewrites.

Relevant files:

- `readme.md`
  - Main product positioning and CLI walkthrough.
- `internal/scorecard/scorecard.go`
  - Markdown output for scorecard, if PR-friendly output needs refinement.
- `cmd/scorecard.go`
  - Existing scorecard flags and formats.
- `cmd/diagnose.go`
  - Existing diagnose command, if present.
- `internal/diagnosis`
  - Diagnosis data used by scorecard or diagnose.
- `tasks/bottleneck-diagnosis-task-list.md`
  - Source roadmap context.

If `--format markdown` already exists for `scorecard` or `diagnose`, improve it. Do not add duplicate formats or commands.

## Epic 5 Required Behavior

### 1. Clarify Framework Vs Product In README

Near the top of `readme.md`, make the separation explicit.

Required statements:

```text
BIASED is the evidence model.
Bottleneck is the CLI that diagnoses delivery risk using that model.
```

Rewrite surrounding copy so Bottleneck does not sound like only a framework validator.

Emphasize:

- Diagnosis
- Release readiness
- Hidden delivery risk
- Evidence-backed decisions

Avoid overclaiming maturity. If a capability is future direction, describe it as direction or roadmap, not as complete product behavior.

### 2. Clarify What Bottleneck Evaluates

Add a scope statement:

```text
Bottleneck evaluates a repo or release using local evidence artifacts.
```

Also explain:

```text
Team and organization-level views can come later by aggregating repo scorecards.
```

Add concise examples of evaluation scope:

- Single application repo
- Service repo
- AI feature repo
- Platform repo

Keep this section short. The goal is to remove ambiguity, not add a product manifesto.

### 3. Clarify AI Vs Non-AI Positioning

Add a positioning statement:

```text
Bottleneck works for any software system, but it is especially useful for AI-enabled systems where behavior, drift, evaluation, and governance cannot be inferred from code alone.
```

Add one AI example.

Recommended AI example:

```text
An AI PDF Risk Summarizer needs evidence that ambiguous financial risk language is flagged instead of summarized as fact.
```

Add one non-AI example.

Recommended non-AI example:

```text
A payments service needs evidence that checkout behavior, test results, security checks, and production telemetry are connected before release.
```

Avoid positioning Bottleneck as limited to LLM apps.

### 4. Update Product Language Across README

Review `readme.md` for language that over-emphasizes framework validation.

Replace weak framing:

```text
validates framework artifacts
```

With stronger, accurate framing:

```text
diagnoses delivery risk from local evidence artifacts
```

Use "validation" when describing the specific `validate` command. Use "diagnosis", "release readiness", and "evidence" when describing the product.

### 5. Preserve Existing Useful Documentation

Do not remove important command docs.

Preserve or update:

- `bottleneck init`
- `bottleneck validate`
- `bottleneck diagnose`
- `bottleneck scorecard`
- `bottleneck explain`
- `bottleneck trace`
- `bottleneck ingest`
- GitHub Actions guidance, if present

The README should stay practical and command-oriented.

## Related Epic 6 Task 6.1: Add PR Comment Output Mode

### 6. Add Or Refine PR-Friendly Markdown Output

Make Bottleneck useful inside pull requests.

Required command:

```sh
bottleneck diagnose --format markdown
```

If `diagnose --format markdown` already exists, verify and improve the output.

If only `scorecard --format markdown` exists, either:

- Add `diagnose --format markdown`, or
- Clearly document `scorecard --format markdown` as the PR comment output while implementing `diagnose --format markdown` if the command exists.

Markdown output should include:

- Primary bottleneck
- Category scores
- Top findings
- Recommended next action

Recommended Markdown shape:

```markdown
## Bottleneck Diagnosis

| Field | Value |
| --- | --- |
| Primary Bottleneck | Assurance |
| Confidence | Low |

### Category Scores

| Category | Score | Status |
| --- | ---: | --- |
| Assurance | 20 | FAIL |
| Intent | 80 | PASS |

### Top Findings

1. No assurance result references BEHAVIOR-001.
2. Assurance accuracy is below the selected environment threshold.

### Recommended Next Action

Map BEHAVIOR-001 to a passing BDD or evaluation result.
```

Keep Markdown readable in:

- GitHub PR comments
- GitHub Actions Step Summary
- Release notes

### 7. Add Snapshot Or Stable Output Tests

Add tests for Markdown output.

Required assertions:

- Markdown includes `Primary Bottleneck`.
- Markdown includes category scores.
- Markdown includes top findings.
- Markdown includes recommended next action.
- Markdown contains stable headings and table columns.
- Markdown does not depend on terminal color or ANSI formatting.

Prefer exact snapshot-style assertions only if the repo already uses snapshots. Otherwise use focused string assertions.

## Backward Compatibility

- Do not remove existing CLI command docs.
- Do not remove existing `scorecard --format markdown` support if it exists.
- Do not change exit behavior for `scorecard`, `diagnose`, or `validate`.
- Do not require GitHub to use Markdown output.
- Keep README accurate to current capabilities. Mark roadmap items as future direction.

## Testing Requirements

Run tests after implementation:

```sh
go test ./...
```

Required test coverage:

1. README contains `BIASED is the evidence model`.
2. README contains `Bottleneck is the CLI that diagnoses delivery risk using that model`.
3. README contains the repo-or-release scope statement.
4. README includes single application repo, service repo, AI feature repo, and platform repo examples.
5. README includes one AI example and one non-AI example.
6. Markdown diagnosis output includes primary bottleneck.
7. Markdown diagnosis output includes category scores.
8. Markdown diagnosis output includes top findings.
9. Markdown diagnosis output includes recommended next action.

Documentation assertions can be simple file-content tests if the repo already tests docs. If not, keep docs manually verified and focus automated tests on Markdown rendering.

## Implementation Guidance

Recommended approach:

1. Edit the top of `readme.md` first so the product position is clear.
2. Add a short "What Bottleneck Evaluates" section.
3. Add a short "AI And Non-AI Systems" section.
4. Review the rest of README for framework-validator wording and tighten it.
5. Check whether `diagnose --format markdown` exists.
6. If needed, add or improve the Markdown renderer using existing diagnosis data.
7. Add focused tests for Markdown output.
8. Run `go test ./...`.

Keep the writing concise. The README should help a developer understand what the product does in the first minute.

## Acceptance Criteria

- README clearly separates BIASED as the evidence model and Bottleneck as the CLI product.
- README says Bottleneck diagnoses delivery risk, release readiness, and hidden evidence gaps.
- README states Bottleneck evaluates a repo or release using local evidence artifacts.
- README explains team and organization views can be built later by aggregating repo scorecards.
- README includes examples for a single application repo, service repo, AI feature repo, and platform repo.
- README positions Bottleneck as useful for any software system and especially valuable for AI-enabled systems.
- README includes one AI example and one non-AI example.
- `bottleneck diagnose --format markdown` produces PR-friendly Markdown, or existing Markdown output is refined and documented if command support already exists.
- Markdown output includes primary bottleneck, category scores, top findings, and recommended next action.
- Tests cover Markdown output.
- `go test ./...` passes.

# AI Implementation Prompt: Epic 8 Improve Explainability

You are working in the `bottleneck` Go CLI repository.

Implement **Epic 8: Improve Explainability**.

This epic covers:

- Task 8.1: Make `bottleneck explain` evidence-driven
- Task 8.2: Add `bottleneck trace --id`

## Product Goal

Developers should be able to understand exactly why Bottleneck assigned a score and how evidence connects across intent, behavior, design, assurance, security, and execution.

`bottleneck explain` should explain score derivation from concrete evidence, not generic framework concepts. `bottleneck trace --id` should let a developer inspect one intent, behavior, or finding end-to-end and immediately see related evidence and missing links.

## Current Architecture To Respect

Inspect the repository before changing code. Use existing diagnosis, scorecard, validation, traceability, configuration, and output rendering patterns.

Relevant artifacts in the current framework include:

- `intent/intent.md`
- `behavior/behavior-spec.md`
- `design/architecture.md`
- `assurance/results.json`
- `security/guardrails.json`
- `execution/telemetry.json`
- `config.yaml`
- `docs/validation.md`

Likely implementation areas may include:

- CLI commands under `cmd/`
- explanation or trace packages under `internal/`
- validation and scoring packages under `internal/`
- shared models under `internal/models/`
- tests and fixtures following the existing test layout

Do not add a database, service, daemon, or network dependency. Explainability should be derived from local artifacts and existing scoring outputs.

## Required Behavior

### Task 8.1: Make `bottleneck explain` Evidence-Driven

Goal: `explain` shows how the score was derived.

Requirements:

- Show evidence found for each category.
- Show evidence missing for each category.
- Show score impact for each category.
- Show related IDs for each category.
- Show recommendation for each category.
- Avoid generic framework descriptions.
- Add tests for explanation output.

The output should answer these questions for every scored category:

- Which concrete files, IDs, results, or telemetry signals were found?
- Which expected evidence is missing?
- Which missing or weak evidence affected the score?
- Which IDs connect this category to other categories?
- What should the developer do next?

Required categories should follow the repository's existing model names. Expected categories likely include:

- Intent
- Behavior
- Design
- Assurance
- Security
- Execution

Recommended text output shape:

```text
Assurance Score: 20

Evidence found:
- assurance/results.json exists
- 1 assurance result references BEHAVIOR-002

Evidence missing:
- No result references BEHAVIOR-001
- No failed scenario explanation found
- No drift or regression evidence found

Related IDs:
- BEHAVIOR-001
- BEHAVIOR-002

Score impact:
-40 broken traceability
-20 missing evaluation evidence
-20 thin assurance evidence

Recommendation:
Add a Cucumber, evaluation, or test result that references BEHAVIOR-001 and explains pass/fail behavior.
```

Implementation expectations:

- Prefer using existing validation findings and score details rather than duplicating scoring rules.
- If current scoring does not expose score impacts, add a small structured explanation model near the scoring logic.
- Keep score impact values deterministic and testable.
- Do not print generic descriptions such as "Assurance measures confidence in implementation" unless tied to concrete evidence.
- Output should be stable enough for focused string or snapshot-style tests.
- If JSON output exists for `explain`, extend it with equivalent structured fields.
- If JSON output does not exist, only add it if consistent with existing CLI patterns.

Recommended internal explanation model, adjusted to match existing conventions:

```go
type CategoryExplanation struct {
    Category        string
    Score           int
    EvidenceFound   []EvidenceFact
    EvidenceMissing []EvidenceGap
    RelatedIDs      []string
    ScoreImpacts    []ScoreImpact
    Recommendation  string
}
```

Recommended behavior for evidence found:

- Include artifact existence, such as `assurance/results.json exists`.
- Include meaningful parsed facts, such as result counts, referenced IDs, SARIF severity counts, or telemetry freshness.
- Include source paths where possible.

Recommended behavior for evidence missing:

- Include missing files.
- Include missing references between IDs.
- Include missing required evidence inside existing files.
- Include stale telemetry or incomplete telemetry signals.

Recommended behavior for score impact:

- Use the existing scoring penalties if already modeled.
- If scoring is currently implicit, introduce named impact reasons that align with validation findings.
- Keep impact labels short and actionable.
- Use negative numbers for penalties and positive numbers only if the existing score model already supports boosts.

Recommended behavior for recommendations:

- Provide one category-specific next action.
- Prefer concrete artifact or ID references.
- Avoid generic framework guidance.

Test coverage must include:

- Explanation includes evidence found for each category.
- Explanation includes evidence missing for each category.
- Explanation includes score impact details.
- Explanation includes related IDs.
- Explanation includes category-specific recommendations.
- Explanation avoids generic framework descriptions.
- JSON output, if supported, contains equivalent structured fields.
- Missing or malformed artifacts produce clear explanations without panics.

### Task 8.2: Add `bottleneck trace --id`

Goal: Developers can inspect one intent, behavior, or finding end-to-end.

Add support for:

```sh
bottleneck trace --id INTENT-001
bottleneck trace --id BEHAVIOR-001
```

Requirements:

- Support tracing an intent ID.
- Support tracing a behavior ID.
- Support tracing a finding or evidence ID if the repository already models those IDs.
- Show related behavior evidence.
- Show related design evidence.
- Show related assurance evidence.
- Show related security evidence.
- Show related execution evidence.
- Highlight missing links.
- Add tests.

The trace output should show the chain of evidence connected to the requested ID.

Recommended text output shape for a behavior ID:

```text
Trace: BEHAVIOR-001

Behavior:
- Found in behavior/behavior-spec.md
- Ambiguous risk clause is flagged

Related intent:
- INTENT-001 found in intent/intent.md

Design evidence:
- Missing: no design evidence references BEHAVIOR-001

Assurance evidence:
- Missing: no assurance result references BEHAVIOR-001

Security evidence:
- Missing: no security evidence references BEHAVIOR-001

Execution evidence:
- Missing: no telemetry or execution signal references BEHAVIOR-001

Missing links:
- BEHAVIOR-001 has no assurance result
- BEHAVIOR-001 has no design reference

Recommendation:
Add design and assurance evidence that reference BEHAVIOR-001.
```

Recommended text output shape for an intent ID:

```text
Trace: INTENT-001

Intent:
- Found in intent/intent.md

Related behavior:
- BEHAVIOR-001 found in behavior/behavior-spec.md
- BEHAVIOR-002 found in behavior/behavior-spec.md

Design evidence:
- DESIGN-001 references INTENT-001

Assurance evidence:
- ASSURANCE-002 references BEHAVIOR-002
- Missing: no assurance result references BEHAVIOR-001

Security evidence:
- SECURITY-001 references INTENT-001

Execution evidence:
- Missing: no execution signal references INTENT-001

Missing links:
- BEHAVIOR-001 has no assurance result

Recommendation:
Add assurance evidence for BEHAVIOR-001 or revise traceability if the behavior is no longer in scope.
```

Implementation expectations:

- Reuse existing traceability parsing if available.
- If traceability is currently embedded in validation, extract shared helpers rather than duplicating parsing.
- Match IDs case-sensitively unless existing ID handling is case-insensitive.
- Preserve deterministic ordering in output.
- Return a non-zero exit code for an unknown ID if existing CLI conventions support non-zero validation failures.
- Print a clear message for unknown IDs.
- Include file paths and line numbers where existing parsers can provide them.
- If line numbers are not available, do not invent them.

Recommended trace model, adjusted to match existing conventions:

```go
type TraceResult struct {
    ID                string
    Kind              string
    Found             bool
    Source            EvidenceLocation
    RelatedIntent     []EvidenceLink
    RelatedBehavior   []EvidenceLink
    RelatedDesign     []EvidenceLink
    RelatedAssurance  []EvidenceLink
    RelatedSecurity   []EvidenceLink
    RelatedExecution  []EvidenceLink
    MissingLinks      []string
    Recommendation    string
}
```

Test coverage must include:

- `trace --id INTENT-001` shows related behaviors.
- `trace --id BEHAVIOR-001` shows related intent when present.
- Trace includes design evidence when present.
- Trace includes assurance evidence when present.
- Trace includes security evidence when present.
- Trace includes execution evidence when present.
- Trace highlights missing links.
- Unknown ID produces a clear message and appropriate exit behavior.
- Output ordering is deterministic.

## CLI Expectations

Preserve existing command behavior unless this epic explicitly changes it.

Expected commands:

```sh
bottleneck explain
bottleneck trace --id INTENT-001
bottleneck trace --id BEHAVIOR-001
```

If `trace` already exists, extend it rather than replacing it.

Recommended shared output flags if existing commands already use them:

- `--format text`
- `--format json`
- `--env <name>`

Do not add new flags unless they fit the existing CLI style and are needed for this epic.

## Output Quality Requirements

Outputs must be:

- Evidence-specific.
- Deterministic.
- Short enough for CLI use.
- Stable enough for tests.
- Useful when pasted into issues or pull requests.
- Free of ANSI color codes unless the repository already has explicit color handling.

Avoid:

- Generic framework descriptions.
- Repeating the same recommendation for every category.
- Claiming evidence exists when only a file exists but relevant IDs are missing.
- Hiding missing links behind aggregate scores.
- Emitting noisy debug details.

## Fixture Expectations

Add representative fixtures in the repository's existing test fixture location. If no fixture location exists, create one that matches Go test conventions.

Required fixture scenarios:

- Complete evidence chain for one behavior.
- Missing assurance evidence for one behavior.
- Missing design evidence for one behavior.
- Intent with multiple related behaviors.
- Unknown or orphaned ID.
- Category with file present but required ID reference missing.

## Acceptance Criteria

The implementation is complete when:

- `bottleneck explain` shows evidence found, evidence missing, score impact, related IDs, and recommendations per category.
- `bottleneck explain` avoids generic framework descriptions.
- `bottleneck trace --id INTENT-001` traces intent to related behavior and downstream evidence.
- `bottleneck trace --id BEHAVIOR-001` traces behavior to intent and downstream evidence.
- Missing links are visible and actionable.
- Unknown IDs are handled clearly.
- Tests cover explanation output and trace output.
- Existing tests still pass.

## Verification

Run the relevant test suite before finishing. Prefer the full Go test suite:

```sh
go test ./...
```

Manually exercise the relevant commands against sample artifacts:

```sh
go run . explain
go run . trace --id INTENT-001
go run . trace --id BEHAVIOR-001
```

## Final Response Requirements

When finished, summarize:

- Files changed.
- Commands added or changed.
- Explainability behavior added.
- Trace behavior added.
- Fixtures added.
- Tests run and results.
- Any intentional compatibility decisions or remaining limitations.

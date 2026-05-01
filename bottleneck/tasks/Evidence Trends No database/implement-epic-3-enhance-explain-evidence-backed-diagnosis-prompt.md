# AI Implementation Prompt: Enhance Explain Into Evidence-Backed Diagnosis

You are working in the Bottleneck Go CLI codebase.

Implement **Implementation Epic 3: Enhance explain into evidence-backed diagnosis** from the `Enterprise SDLC Evidence Package` milestone.

## Milestone Context

Bottleneck should help enterprise teams explain SDLC bottlenecks using local evidence, not opinions and not external metrics systems. This epic improves the existing `bottleneck explain` command so it can support leadership conversations.

This is an enhancement to an existing command, not a replacement.

## Scope

Enhance:

```sh
bottleneck explain
```

Add or confirm support for:

```sh
bottleneck explain
bottleneck explain --category=assurance
bottleneck explain --category=security
bottleneck explain --capability=billing
bottleneck explain --format=text
bottleneck explain --format=markdown
bottleneck explain --format=json
bottleneck explain --out=bottleneck/reports/bottleneck-explanation.md
bottleneck explain --env=default
bottleneck explain --strict
```

Current capability behavior should continue to work.

## Current Code To Inspect

Read before changing code:

- `cmd/explain.go`
- `cmd/explain_test.go` or related command tests
- `internal/diagnosis/*`
- `internal/scorecard/*`
- `internal/validator/*`
- `internal/traceability/*`
- `internal/ingest/*`
- existing render helpers
- README and docs references to `explain`

Preserve existing behavior unless intentionally updated by tests.

## Output Model

For each category, generate:

- Category
- Status
- Score, if available
- Why this matters
- Evidence found
- Evidence missing
- Risk to delivery
- Recommended next actions
- Suggested owner roles
- Suggested GitHub Actions or automation hooks

The explanation must be deterministic and local-only. Do not call an LLM or external service.

## Example Text Output

```text
Bottleneck Explanation

Primary bottleneck: Assurance
Status: FAIL

Why this matters:
The team can describe intended behavior, but cannot prove enough behavior before release.

Evidence found:
- behavior-spec.md exists
- Cucumber results were ingested
- Some test evidence references behavior IDs

Evidence missing:
- 5 expected behaviors have no mapped tests
- 2 tests do not reference behavior IDs
- No release validation evidence was found

Risk to delivery:
The team may ship functionality that appears complete but cannot be proven against intended behavior.

Recommended actions:
1. Add behavior IDs to Cucumber scenarios.
2. Map each critical behavior to at least one automated test.
3. Add release validation evidence before production approval.

Suggested owner roles:
- Assurance Engineer
- Developer
- Product/Domain Expert

Suggested automation:
- Run Cucumber in GitHub Actions
- Upload test output to bottleneck/assurance/results.json
- Fail release gate when critical behaviors lack mapped tests
```

## Rule-Based Explanation Engine

Build deterministic rules. Do not use AI generation.

### Intent Rules

If intent file is missing:

```text
Missing intent evidence.
Recommendation: Create bottleneck/intent/intent.md with measurable outcomes and constraints.
```

If intent exists but lacks measurable outcomes:

```text
Intent exists but does not clearly define measurable outcomes.
Recommendation: Add observable outcomes, business constraints, and unacceptable outcomes.
```

If intent has placeholders:

```text
Intent contains placeholder or thin content.
Recommendation: Replace template text with product-specific intent.
```

### Behavior Rules

If behavior spec is missing:

```text
Missing behavior specification.
Recommendation: Create behavior-spec.md with expected and unacceptable behaviors.
```

If behaviors lack IDs:

```text
Behavior expectations are not traceable.
Recommendation: Add stable behavior IDs such as BEHAVIOR-001.
```

If behavior has no tests:

```text
Behavior is not validated.
Recommendation: Map each critical behavior to test evidence.
```

### Design Rules

If architecture doc is missing:

```text
Missing architecture evidence.
Recommendation: Document major components, boundaries, dependencies, and risk decisions.
```

If design lacks tradeoffs:

```text
Architecture exists but does not explain tradeoffs.
Recommendation: Add decision records for key constraints and design choices.
```

If design lacks operational or failure-mode content:

```text
Architecture does not describe failure modes.
Recommendation: Add fallback, monitoring, and operational assumptions.
```

### Assurance Rules

If Cucumber or test evidence is missing:

```text
Missing automated validation evidence.
Recommendation: Add test output or BDD evidence under bottleneck/assurance/.
```

If tests exist but are not mapped:

```text
Tests exist but are not linked to behavior IDs.
Recommendation: Add traceability references from test evidence to behavior expectations.
```

If coverage is low:

```text
Critical behaviors lack validation.
Recommendation: Prioritize tests for high-risk behaviors before expanding feature scope.
```

### Security Rules

If SARIF or security evidence is missing:

```text
Missing security evidence.
Recommendation: Add CodeQL, dependency review, secret scanning, or SARIF evidence.
```

If high severity finding exists:

```text
High severity security findings exist.
Recommendation: Block release until findings are triaged or resolved.
```

If guardrails are missing:

```text
Security guardrails are not documented.
Recommendation: Add security/guardrails.json or equivalent evidence.
```

### Execution Rules

If telemetry is missing:

```text
Missing execution evidence.
Recommendation: Add telemetry or production-readiness evidence.
```

If adoption/usage signals are weak:

```text
Execution evidence suggests weak adoption or user trust.
Recommendation: Review user workflow, training, and feedback loops.
```

If reliability metrics are weak:

```text
Execution evidence suggests operational instability.
Recommendation: Address error rate, latency, or incident signals before accelerating release.
```

## Owner Roles

Add deterministic owner suggestions:

- Intent: Product Lead, Domain Expert, Technical Lead
- Behavior: Product Lead, Domain Expert, QA/Assurance Engineer
- Design: Architect, Technical Lead, Platform Engineer
- Assurance: QA/Assurance Engineer, Developer, Product Lead
- Security: Security Engineer, Platform Engineer, Technical Lead
- Execution: SRE/Operations, Product Lead, Customer Success/Adoption Lead

## Automation Suggestions

Add deterministic automation suggestions:

- Intent: PR template requiring intent reference, Markdown quality checks, commit hook for required intent IDs
- Behavior: behavior spec linting, traceability checks
- Design: architecture decision record check, diagram/doc freshness check
- Assurance: Cucumber in GitHub Actions, test result ingestion, behavior-to-test traceability gate
- Security: CodeQL, dependency review, secret scanning, SARIF ingestion
- Execution: telemetry JSON ingestion, release health check, production signal review

## Rendering

Support:

- `--format=text`
- `--format=markdown`
- `--format=json`
- `--out=<path>`

If `--out` is provided, write output to the path and create parent directories as needed.

Markdown output should be leadership-readable. JSON output should include structured explanation objects.

## Acceptance Criteria

- `bottleneck explain` continues to work with existing behavior.
- `bottleneck explain --category=assurance` explains only Assurance.
- Explanation includes evidence found.
- Explanation includes evidence missing.
- Explanation includes risk to delivery.
- Explanation includes recommended actions.
- Explanation includes suggested owner roles.
- Explanation includes suggested automations.
- Explanation supports text output.
- Explanation supports Markdown output.
- Explanation supports JSON output.
- Explanation can write to `bottleneck/reports/bottleneck-explanation.md`.
- Explanation is deterministic and does not require an LLM.
- Existing explain tests continue passing or are updated intentionally.

## Tests To Add

Add tests near existing explain code:

- `TestExplainCategoryFilter`
- `TestExplainIncludesEvidenceFound`
- `TestExplainIncludesEvidenceMissing`
- `TestExplainIncludesRisk`
- `TestExplainIncludesRecommendations`
- `TestExplainIncludesOwnerRoles`
- `TestExplainIncludesAutomationSuggestions`
- `TestExplainMarkdownOutput`
- `TestExplainJSONOutput`
- `TestExplainWritesOutputFile`
- `TestExplainIntentRules`
- `TestExplainBehaviorRules`
- `TestExplainDesignRules`
- `TestExplainAssuranceRules`
- `TestExplainSecurityRules`
- `TestExplainExecutionRules`

Use fixtures that target one rule at a time. Avoid brittle full-output snapshots; assert high-signal sections and phrases.

## Implementation Constraints

- Do not replace explain with LLM output.
- Do not call external services.
- Do not break existing `--capability` behavior.
- Do not change scorecard math unless a test exposes a shared bug.
- Keep rules deterministic and easy to test.
- Preserve existing commands and output formats.

## Verification

Run:

```sh
go test ./...
```

If feasible, manually verify:

```sh
bottleneck explain
bottleneck explain --category=assurance
bottleneck explain --category=assurance --format=markdown --out=bottleneck/reports/bottleneck-explanation.md
bottleneck explain --format=json
```

Use a temporary project for manual output files.

## Final Response Requirements

When finished, report:

1. Explain command changes.
2. Rule engine behavior.
3. Output formats supported.
4. Tests added or changed.
5. Documentation/help updates.
6. Exact commands run and results.
7. Any acceptance criteria intentionally deferred and why.


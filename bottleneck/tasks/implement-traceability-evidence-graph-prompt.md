# AI Implementation Prompt: Traceability Evidence Graph

You are working in the `bottleneck` Go CLI repository.

Implement feature 3 from the roadmap: **Traceability Across Intent, Behavior, Tests, Security, and Telemetry**.

## Product Goal

Turn bottleneck from artifact validation into a release-readiness evidence graph.

The CLI should not only confirm that artifacts exist. It should understand how intent, behavior, assurance, security, and execution evidence connect, then report orphaned or broken links before release.

## Current Architecture To Respect

Use the existing Go CLI patterns and validation engine. Add traceability as a deterministic local feature. Do not add network calls, LLM calls, databases, or external services.

Relevant files:

- `cmd/root.go`
  - Registers Cobra commands.
- `cmd/validate.go`
  - Prints validation results from `validator.Engine`.
- `cmd/explain.go`
  - Renders validation results through `internal/explainer`.
- `cmd/scorecard.go`
  - Renders validation results through `internal/scorecard`.
- `cmd/init.go`
  - Creates default artifacts.
- `internal/validator/engine.go`
  - Builds validators and calculates system status.
- `internal/validator/behavior.go`
- `internal/validator/intent.go`
- `internal/validator/design.go`
- `internal/validator/assurance.go`
- `internal/validator/security.go`
- `internal/validator/execution.go`
- `internal/validator/markdown.go`
- `internal/models/result.go`
  - Defines validation results and statuses.
- `internal/explainer/explainer.go`
  - Already surfaces validation details as evidence.
- `internal/scorecard/scorecard.go`
  - Already surfaces validation details in scorecard output.
- `readme.md`
  - Update command and artifact documentation.

## Required Behavior

### 1. Add Stable Evidence IDs

Support stable IDs across framework evidence:

- `INTENT-001`
- `BEHAVIOR-001`
- `DESIGN-001`
- `ASSURANCE-001`
- `SECURITY-001`
- `EXECUTION-001`

Use a deterministic parser and strict ID format:

```text
^(INTENT|BEHAVIOR|DESIGN|ASSURANCE|SECURITY|EXECUTION)-[0-9]{3,}$
```

IDs must be unique across the project. Duplicate IDs should fail validation because they make trace output ambiguous.

### 2. Define A Backward-Compatible Evidence Syntax

Add support for Markdown evidence headings in intent, behavior, and design artifacts.

Recommended Markdown convention:

```markdown
### INTENT-001: Reduce release risk
Refs:
- BEHAVIOR-001

Release decisions must be backed by test, security, and execution evidence.
```

```markdown
### BEHAVIOR-001: Block production release when assurance fails
Critical: true
Refs:
- INTENT-001
- ASSURANCE-001

When production assurance is below threshold, the release recommendation is Block.
```

Rules:

- Evidence headings start with a Markdown heading marker followed by a valid ID.
- `Refs:` introduces zero or more referenced IDs in list form.
- `References:` should be accepted as an alias for `Refs:`.
- `Critical: true` marks behavior as critical.
- `Critical: false` or no `Critical` line means not critical.
- Body text continues until the next Markdown heading at the same or higher level.

For JSON artifacts, support an optional `evidence` array while preserving current simple schemas.

Recommended JSON convention:

```json
{
  "scenarios_total": 1,
  "scenarios_passed": 1,
  "scenarios_failed": 0,
  "failures": [],
  "evidence": [
    {
      "id": "ASSURANCE-001",
      "refs": ["BEHAVIOR-001"],
      "source": "cucumber",
      "status": "pass"
    }
  ]
}
```

Apply the same optional `evidence` convention to:

- `bottleneck/assurance/results.json`
- `bottleneck/security/guardrails.json`
- `bottleneck/execution/telemetry.json`

Existing artifacts without `evidence` must still parse under existing validators. Traceability validation may warn or fail separately based on environment and strictness.

### 3. Validate References

Add traceability validation that checks:

- Every ID is unique.
- Every referenced ID exists.
- References use valid ID syntax.
- Intent can reference behavior.
- Behavior can reference intent and assurance.
- Assurance can reference behavior.
- Security can reference behavior, assurance, or release-relevant evidence.
- Execution can reference behavior and assurance.

Keep this flexible enough that valid cross-links are not rejected too aggressively, but strict enough to catch typos and missing evidence.

Broken references should fail validation because they are deterministic data errors.

Example detail:

```text
bottleneck/behavior/behavior-spec.md BEHAVIOR-001 references missing ASSURANCE-009
```

### 4. Map Intent To Behavior

Traceability must identify behavior that is not tied to intent.

Rules:

- A behavior is mapped to intent when:
  - The behavior references an `INTENT-*` ID, or
  - An intent references the behavior's `BEHAVIOR-*` ID.
- Behavior without intent creates a warning by default.
- In strict mode or production mode, behavior without intent should fail.

If strict mode does not exist yet, add it consistently with the validation engine pattern:

- `bottleneck validate --strict`
- `bottleneck explain --strict`
- `bottleneck scorecard --strict`
- `bottleneck trace --strict` if trace validates as part of output.

### 5. Map Behavior To Assurance Evidence

Traceability must identify critical behavior without assurance evidence.

Rules:

- A behavior is linked to assurance when:
  - The behavior references an `ASSURANCE-*` ID, or
  - An assurance evidence entry references the behavior's `BEHAVIOR-*` ID.
- Critical behavior without assurance evidence creates a warning by default.
- In strict mode or production mode, critical behavior without assurance evidence should fail.

If a behavior is not marked critical, missing assurance may warn only when the selected environment is `production`, or remain informational if the existing status model cannot represent info.

### 6. Report Orphaned Evidence

Report orphaned evidence in validation details:

- Intent not connected to any behavior.
- Behavior not connected to intent.
- Critical behavior not connected to assurance.
- Assurance evidence not connected to behavior.
- Security evidence not connected to any release-relevant artifact.
- Execution telemetry not connected to behavior or assurance.

Use actionable detail messages with file path and ID.

Example:

```text
bottleneck/intent/intent.md INTENT-002 is not linked to any behavior
bottleneck/assurance/results.json ASSURANCE-003 is not linked to any behavior
bottleneck/execution/telemetry.json EXECUTION-001 is not linked to behavior or assurance evidence
```

### 7. Add `bottleneck trace <id>`

Add a new Cobra command:

```sh
bottleneck trace INTENT-001
bottleneck trace BEHAVIOR-001
bottleneck trace ASSURANCE-001 --format=json
```

Supported flags:

- `--env`, default `default`
- `--format`, values `text` and `json`
- `--strict`, if strict mode exists after implementation

Trace output should show linked evidence in both directions:

- The selected ID.
- Artifact path.
- Title or summary.
- Type.
- Status if available.
- Outbound references.
- Inbound references.
- Connected chain.
- Broken references.
- Orphan warnings.

Text example:

```text
Trace: BEHAVIOR-001
Type: Behavior
Artifact: bottleneck/behavior/behavior-spec.md
Title: Block production release when assurance fails
Critical: true

Outbound References:
- INTENT-001
- ASSURANCE-001

Inbound References:
- INTENT-001
- ASSURANCE-001

Evidence Chain:
INTENT-001 -> BEHAVIOR-001 -> ASSURANCE-001 -> EXECUTION-001

Warnings:
- No security evidence linked to this behavior.
```

JSON output must be stable and automation-friendly.

Recommended JSON shape:

```json
{
  "schema_version": "trace.v1",
  "environment": "production",
  "query": "BEHAVIOR-001",
  "node": {
    "id": "BEHAVIOR-001",
    "type": "Behavior",
    "title": "Block production release when assurance fails",
    "artifact_path": "bottleneck/behavior/behavior-spec.md",
    "critical": true
  },
  "outbound_refs": ["INTENT-001", "ASSURANCE-001"],
  "inbound_refs": ["INTENT-001", "ASSURANCE-001"],
  "chains": [
    ["INTENT-001", "BEHAVIOR-001", "ASSURANCE-001", "EXECUTION-001"]
  ],
  "warnings": [],
  "broken_refs": []
}
```

Unknown trace IDs should return a useful error and non-zero exit code.

### 8. Integrate Traceability Into Validation, Explain, And Scorecard

Add traceability as a first-class validation result. Choose one of these approaches:

- Add a new capability named `Traceability`.
- Or add traceability details under the affected capabilities.

Prefer adding `Traceability` as a separate capability so broken references and orphaned evidence are visible without overloading Behavior or Assurance.

Required integration:

- `bottleneck validate` shows traceability warnings/failures.
- `bottleneck explain` includes traceability evidence and next actions.
- `bottleneck scorecard` includes traceability status and evidence.
- `bottleneck trace <id>` uses the same parsed graph as validation.

### 9. Update Initialized Templates

Update `bottleneck init` templates to include minimal example IDs and references without making sample evidence look production-grade.

Example direction:

- `intent.md` includes `INTENT-001`.
- `behavior-spec.md` includes `BEHAVIOR-001` referencing `INTENT-001`.
- `assurance/results.json` includes `ASSURANCE-001` referencing `BEHAVIOR-001`.
- `security/guardrails.json` can include `SECURITY-001`.
- `execution/telemetry.json` can include `EXECUTION-001`.

If placeholder/content-quality detection already exists, keep sample trace IDs but still allow placeholder warnings.

## Backward Compatibility

Do not break existing simple artifacts:

- Existing Markdown files without IDs should still pass their structural validators if they passed before.
- Traceability should warn or fail separately based on selected environment and strict mode.
- Existing JSON schemas for assurance, security, and execution must remain valid.
- Optional `evidence` arrays must not interfere with existing threshold calculations.

## Testing Requirements

Add focused Go tests using temp directories and real artifact files.

Required test cases:

1. Valid IDs across intent, behavior, assurance, security, and execution parse into a graph.
2. Duplicate IDs fail validation.
3. Invalid ID syntax fails validation.
4. References to missing IDs fail validation and identify the file, ID, and missing reference.
5. Behavior without intent creates warning by default.
6. Behavior without intent fails in strict mode or production mode.
7. Critical behavior without assurance creates warning by default.
8. Critical behavior without assurance fails in strict mode or production mode.
9. Orphaned intent, assurance, security, and telemetry evidence are reported.
10. `bottleneck trace <id>` text output shows outbound refs, inbound refs, and an evidence chain.
11. `bottleneck trace <id> --format=json` returns stable JSON with `schema_version: trace.v1`.
12. Unknown trace ID returns a useful error.
13. Existing simple JSON artifacts without `evidence` still validate under the original validators.

Run:

```sh
go test ./...
```

## Implementation Guidance

Recommended approach:

1. Create `internal/traceability` or `internal/validator/traceability` for graph parsing and analysis.
2. Define explicit structs, such as:
   - `EvidenceNode`
   - `EvidenceRef`
   - `EvidenceGraph`
   - `TraceResult`
   - `TraceFinding`
3. Build a parser for Markdown evidence headings and optional JSON `evidence` arrays.
4. Add graph validation helpers for duplicate IDs, missing references, orphaned nodes, and required mappings.
5. Add `TraceabilityValidator` and register it in `validator.NewEngine`.
6. Add `cmd/trace.go` for `bottleneck trace`.
7. Reuse the graph builder from both validation and trace command.
8. Update `explainer` and `scorecard` metadata for the new `Traceability` capability.
9. Update `init` templates and docs.
10. Add tests before or alongside implementation.

Keep the feature deterministic and small. Avoid broad semantic interpretation of prose. Traceability should come from explicit IDs and references, not inferred meaning.

## Acceptance Criteria

- Stable IDs such as `INTENT-001`, `BEHAVIOR-001`, `ASSURANCE-001`, `SECURITY-001`, and `EXECUTION-001` are supported.
- Duplicate IDs fail validation.
- Broken references fail validation and identify the exact source.
- Behavior without intent creates a warning or failure based on environment or strict mode.
- Critical behavior without assurance evidence creates a warning or failure based on environment or strict mode.
- Orphaned intent, behavior, tests, security, and telemetry are reported.
- `bottleneck trace <id>` shows linked evidence in both directions.
- Trace output shows the chain from intent to behavior to validation and runtime evidence.
- Existing artifact validation remains backward compatible.
- `go test ./...` passes.

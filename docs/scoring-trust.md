# Scoring Trust

Bottleneck does not ask teams to trust a black-box score. Every score can be explained with `bottleneck explain-score`.

## What The Score Means

`bottleneck assess` reports five related signals:

- **Maturity**: the SDLC evidence level for the repo or release. Level 0 is Ad Hoc, Level 1 is Documented, Level 2 is Managed, Level 3 is Measured, and Level 4 is Optimized.
- **Release recommendation**: the current release posture from the scorecard: Proceed, Conditional, Block, or Unknown.
- **Primary bottleneck**: the weakest evidence category blocking release confidence.
- **Score confidence**: Bottleneck's confidence in the score based on evidence coverage, provenance, and freshness.
- **AI readiness**: whether the repo has enough behavior, assurance, security, and execution evidence to support AI-assisted release work.

These signals are related, but they are not the same. A team can have a Managed maturity level and still receive a Block recommendation if the current assurance or security evidence fails.

## Evidence Provenance

Automatic ingestion adds provenance fields to normalized evidence:

- `source`: the original report path.
- `generated_by`: the tool or format that produced the evidence, such as `junit`, `lcov`, `sarif`, `cucumber`, or `telemetry`.
- `generated_at`: the source file modification time when available.
- `ingested_at`: when Bottleneck normalized the evidence.
- `confidence`: high when evidence is mapped to explicit refs, medium when the tool output is useful but unmapped.
- `provenance`: a short explanation of how the evidence entered Bottleneck.

Refs are only inferred from explicit identifiers such as `BEHAVIOR-003`, `ASSURANCE-001`, or `SECURITY-001`. Bottleneck does not invent refs from vague test names.

## Freshness Defaults

Freshness affects score confidence:

- Assurance evidence is expected within 14 days.
- Security evidence is expected within 14 days.
- Execution telemetry is expected within 7 days.

Freshness is intentionally simple in the first release. Teams can still use `scorecard --details` for configured threshold details and `explain-score` for score rationale.

## Useful Commands

```sh
bottleneck discover
bottleneck ingest --auto
bottleneck assess
bottleneck explain-score
bottleneck trace BEHAVIOR-003
```

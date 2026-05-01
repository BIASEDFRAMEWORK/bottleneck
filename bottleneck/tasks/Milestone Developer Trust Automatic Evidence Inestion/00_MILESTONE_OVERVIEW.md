# Bottleneck Milestone: Developer Trust + Automatic Evidence Ingestion

## Purpose

This milestone moves Bottleneck from a manually maintained SDLC evidence model toward an automatic SDLC maturity assessment tool.

The product belief is:

> The more mature your SDLC is, the easier your releases become. The easier your releases become, the safer it is to enable AI-assisted development.

The implementation goal is:

> Developers should maintain intent and behavior. Bottleneck should derive as much maturity evidence as possible from tools developers already use.

## Current repo assumptions

The existing Go CLI already includes commands such as:

- `bottleneck init --template saas`
- `bottleneck validate`
- `bottleneck scorecard`
- `bottleneck diagnose`
- `bottleneck trace`
- `bottleneck ingest cucumber --file ...`
- `bottleneck ingest sarif --file ...`
- `bottleneck ingest test-summary --file ...`
- `bottleneck ingest telemetry --file ...`
- `bottleneck snapshot`
- `bottleneck seed-history`
- `bottleneck trends`
- `bottleneck report`
- `bottleneck explain`

Do not remove or break any existing command behavior. All new functionality should be additive and backward compatible.

## Target user experience

A new SaaS team should be able to run:

```bash
bottleneck init --template saas
bottleneck assess
```

And receive a useful SDLC maturity assessment that includes:

- Overall maturity level
- AI readiness
- Release friction
- Primary bottleneck
- Detected evidence sources
- Missing or weak evidence
- Suggested next action
- Score confidence

## Milestone outputs

This milestone is split into sequential Codex implementation prompts:

1. `01_IMPLEMENT_DISCOVER_COMMAND.md`
2. `02_IMPLEMENT_AUTO_INGEST_AND_PROVENANCE.md`
3. `03_IMPLEMENT_ASSESS_COMMAND.md`
4. `04_IMPLEMENT_SCORE_TRUST_MATURITY_AND_EXPLAIN_SCORE.md`
5. `05_IMPLEMENT_DOCS_AND_COMMAND_ALIASES.md`
6. `06_IMPLEMENT_GITHUB_ACTIONS_SUMMARY.md`

Each prompt assumes the previous prompt may already have changed the codebase. Run them in order.

## Milestone design principles

### 1. Prefer generated evidence over manual evidence

Manual files should define intent, behavior, thresholds, and exceptions. Tool outputs should generate most of the evidence used for scoring.

### 2. Every score must explain itself

A maturity score without rationale will not be trusted by developers. Every score should be traceable to evidence, missing evidence, stale evidence, or configured thresholds.

### 3. Do not create onboarding overload

The day-one path should teach only:

```bash
bottleneck init --template saas
bottleneck assess
bottleneck trace BEHAVIOR-003
```

Everything else is advanced usage.

### 4. Maturity is not the same as release approval

Keep these concepts separate:

- SDLC maturity level
- Release recommendation
- Primary bottleneck
- Score confidence
- AI readiness

### 5. Inference is allowed, but must be labeled

Explicit evidence is stronger than inferred evidence. If Bottleneck guesses a relationship from file names, test names, paths, or tags, mark it as inferred and lower confidence.

## Definition of done for the milestone

The milestone is complete when:

- `bottleneck discover` detects common evidence sources.
- `bottleneck ingest --auto` ingests known evidence from configured/default paths.
- Ingested evidence includes provenance metadata.
- `bottleneck assess` gives a clear maturity-oriented day-one output.
- Scores include confidence and explainable rationale.
- `bottleneck explain-score` explains scoring logic in detail.
- Existing commands still pass their tests.
- README/docs explain the 5-minute onboarding path.
- GitHub Actions usage can publish a useful assessment summary.


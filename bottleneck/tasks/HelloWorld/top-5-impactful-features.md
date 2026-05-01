# Task: Add or Improve the 5 Most Impactful bottleneck Features

## Objective

Prioritize the next set of bottleneck product work around features that most directly improve release-readiness confidence, enterprise usefulness, and evidence quality.

## Why This Matters

bottleneck currently has a useful CLI foundation for validating framework artifacts, but the project can still produce false confidence when artifacts are placeholders, shallow, manually edited, or disconnected from real delivery signals. The next features should make the scorecard evidence-backed, CI-ready, traceable, and useful for release decisions.

## Top 5 Features

### 1. Placeholder and Content Quality Detection

**Impact:** Prevents a newly initialized or weakly documented project from appearing production-ready.

**Feature requirements:**

- Detect default template text from `bottleneck init`.
- Warn when required sections exist but contain placeholder or insufficient content.
- Add strict mode that fails validation when placeholder content remains.
- Show placeholder status in `validate`, `explain`, and `scorecard`.
- Add tests for unchanged templates, partially completed artifacts, and valid artifacts.

**Acceptance criteria:**

- A freshly initialized project does not receive a clean production-ready result.
- Warnings identify the exact file and section that need real evidence.
- Strict mode can turn placeholder warnings into failures.

### 2. Evidence-Backed Scorecard Depth

**Impact:** Makes `bottleneck scorecard` the primary product surface instead of a thin validation summary.

**Feature requirements:**

- Add evidence counts and missing evidence per capability.
- Add status levels: `PASS`, `WARN`, `FAIL`, and `UNKNOWN`.
- Add release recommendation: `Proceed`, `Conditional`, `Block`, or `Unknown`.
- Display effective environment thresholds.
- Add `--view executive`, `--view engineering`, and `--view governance`.
- Support Markdown output for GitHub Step Summary and PR comments.

**Acceptance criteria:**

- Scorecard explains why every category passed, warned, failed, or is unknown.
- JSON output is stable enough for CI/CD automation.
- Markdown output is readable in GitHub Actions and pull request comments.

### 3. Traceability Across Intent, Behavior, Tests, Security, and Telemetry

**Impact:** Turns bottleneck from artifact validation into a release-readiness evidence graph.

**Feature requirements:**

- Add stable evidence IDs such as `INTENT-001`, `BEHAVIOR-001`, `ASSURANCE-001`, `SECURITY-001`, and `EXECUTION-001`.
- Validate uniqueness of IDs.
- Validate that references point to existing artifacts.
- Map intent to behavior.
- Map behavior to assurance evidence.
- Report orphaned intent, behavior, tests, and telemetry.
- Add `bottleneck trace <id>` to show linked evidence.

**Acceptance criteria:**

- Behavior without intent creates a warning or failure based on environment.
- Critical behavior without assurance evidence creates a warning or failure.
- Trace output shows the chain from intent to behavior to validation and runtime evidence.

### 4. GitHub Actions and Pull Request Integration

**Impact:** Places bottleneck in the delivery workflow where release decisions happen.

**Feature requirements:**

- Add example GitHub Actions workflows for `validate` and `scorecard`.
- Output Markdown to GitHub Step Summary.
- Support PR comment output.
- Detect PR metadata when running in GitHub Actions.
- Include PR risk signals such as changed files, reviewers, approvals, checks, and large AI-generated PRs.
- Support annotations for missing or weak artifacts.

**Acceptance criteria:**

- A user can copy a workflow into `.github/workflows/`.
- Pull requests can be blocked based on configured evidence thresholds.
- Scorecard findings appear in CI output with links to affected files where possible.

### 5. Evidence Ingestion for Tests, Security, and Telemetry

**Impact:** Reduces manual artifact editing and grounds the scorecard in real tool outputs.

**Feature requirements:**

- Add `bottleneck ingest cucumber --file <path>` for BDD results.
- Add `bottleneck ingest codeql --file <path>` for SARIF security results.
- Add generic test summary ingestion for test count, failures, skips, and coverage.
- Add generic telemetry ingestion for adoption rate, error rate, latency, rollback rate, and source environment.
- Store normalized evidence in framework artifacts without replacing source tools.

**Acceptance criteria:**

- Failed tests reduce the Assurance score.
- CodeQL findings affect the Security score based on severity and environment.
- Production telemetry can warn or fail Execution even when tests pass.
- Ingestion commands are covered by parsing and threshold tests.

## Suggested Delivery Order

1. Placeholder and content quality detection.
2. Evidence-backed scorecard depth.
3. GitHub Actions and PR integration.
4. Traceability IDs and coverage mapping.
5. Evidence ingestion for Cucumber, CodeQL, tests, and telemetry.

## Definition of Done

- Each feature includes CLI behavior, tests, documentation, and example artifacts.
- Existing `validate`, `explain`, and `scorecard` behavior remains backward compatible unless strict mode is enabled.
- Every new score or warning links back to an artifact, threshold, or ingested evidence source.

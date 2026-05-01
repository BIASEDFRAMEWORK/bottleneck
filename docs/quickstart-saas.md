# SaaS Day-One Quickstart

Use this walkthrough from a SaaS application repository when you want a fast release-readiness check for a Subscription Billing Release.

For a demo-ready copy of this flow with Bottleneck evidence, sample reports, and a workflow already checked in, use `examples/saas-billing`.

## 1. Create The Starter

```sh
bottleneck init --template saas
```

Generated files:

```text
bottleneck/config.yaml
bottleneck/intent/intent.md
bottleneck/behavior/behavior-spec.md
bottleneck/design/architecture.md
bottleneck/assurance/features/sample.feature
bottleneck/assurance/results.json
bottleneck/security/guardrails.json
bottleneck/execution/telemetry.json
bottleneck/docs/validation.md
```

The sample release covers payment method updates, failed invoice retry, and duplicate-charge prevention. `BEHAVIOR-003` is the intentional Day One gap: payment retry behavior exists, but no mapped test evidence proves it yet.

## 2. Read The Scorecard

```sh
bottleneck scorecard
```

Expected starter output:

```text
Bottleneck Scorecard
Environment: default
Release Recommendation: Conditional
Primary Bottleneck: Assurance

Category Results:
- Intent: Pass
- Behavior: Pass
- Design: Pass
- Assurance: Warn
- Security: Pass
- Execution: Pass

Why:
BEHAVIOR-003 payment retry behavior has no mapped test evidence.

Next Action:
Add assurance evidence for payment retry behavior. Map it to BEHAVIOR-003.
```

The scorecard is the main terminal surface. It summarizes the delivery decision before raw validation findings.

## 3. Inspect Details When Needed

```sh
bottleneck scorecard --details
```

Detailed mode includes thresholds, capability evidence, missing evidence, and score impacts:

```text
Traceability:
  Status: WARN
  Score: 60
  Evidence:
    - bottleneck/behavior/behavior-spec.md BEHAVIOR-003 has no mapped test evidence
  Missing Evidence:
    - bottleneck/behavior/behavior-spec.md BEHAVIOR-003 has no mapped test evidence
  Score Impacts:
    - bottleneck/behavior/behavior-spec.md BEHAVIOR-003 has no mapped test evidence (-25)
```

## 4. Validate Evidence Quality

```sh
bottleneck validate
```

Expected starter posture:

```text
Traceability: WARNING (traceability warnings detected)
  bottleneck/behavior/behavior-spec.md BEHAVIOR-003 has no mapped test evidence

System Status: WARNING
Primary Bottleneck: Traceability
Environment: default
```

`validate` reports the raw validation status. Warning-only output exits `0`; failures exit `1`.

## 5. Capture JSON Output

```sh
bottleneck scorecard --format=json
```

Excerpt:

```json
{
  "schema_version": "scorecard.v2",
  "environment": "default",
  "system_status": "WARN",
  "release_recommendation": "Conditional",
  "primary_bottleneck": "Assurance",
  "diagnosis": {
    "primary_bottleneck": "Assurance",
    "recommended_action": "Add assurance evidence for payment retry behavior."
  }
}
```

## 6. Choose An Environment

The SaaS starter includes inherited thresholds for `local`, `dev`, `test`, `stage`, and `production`.

```sh
bottleneck scorecard --env=dev
bottleneck scorecard --env=production
bottleneck diagnose --env=production --gate=release
```

Local and dev are tuned for fast feedback: incomplete release evidence can warn without blocking the release gate. Test and stage increase assurance and security thresholds. Production requires traceability and uses a higher release-gate minimum score, so critical gaps such as missing mapped assurance for `BEHAVIOR-003`, broken traceability, critical security findings, or stale telemetry can block release readiness.

The scorecard shows the resolved values, including inherited defaults:

```text
Environment: production
Effective Thresholds:
- Minimum score: 85
- Required traceability: true
- Critical security findings allowed: 0
- Stale telemetry allowed: false
```

If an environment name is invalid, Bottleneck fails with the unknown name and the supported environment list. Fix the `--env` value or add that environment under `bottleneck/config.yaml`.

## 7. Diagnose The Bottleneck

```sh
bottleneck diagnose
```

Expected excerpt:

```text
Primary Bottleneck: Assurance
Reason: BEHAVIOR-003 is not linked to any passing test evidence.
Impact: Release confidence is reduced because payment retry behavior is unproven.
Next Action: Add or ingest test evidence mapped to BEHAVIOR-003.
Inspect: bottleneck trace BEHAVIOR-003
Relevant Evidence: BEHAVIOR-003
```

## 8. Trace BEHAVIOR-003

```sh
bottleneck trace BEHAVIOR-003
```

Expected excerpt:

```text
Trace: BEHAVIOR-003

Behavior:
- Found in bottleneck/behavior/behavior-spec.md
- Duplicate charges are prevented during retry

Missing links:
- BEHAVIOR-003 has no mapped test evidence
- BEHAVIOR-003 has no execution signal

Recommendation:
Add assurance evidence for payment retry behavior.
```

`bottleneck trace --id BEHAVIOR-003` is also supported.

## 9. Ingest Sample Evidence

The repository includes runnable sample reports under `examples/saas/reports/`. In a real SaaS app these files usually come from CI, test runners, security scanners, and telemetry exports. For a local trial, put the sample files under `reports/` in the project you initialized:

```sh
mkdir -p reports
cp examples/saas/reports/* reports/
```

### Cucumber

```sh
bottleneck ingest cucumber --file reports/cucumber.json
```

Represents: BDD/Cucumber scenario results from the Subscription Billing Release test suite.
Writes: `bottleneck/assurance/results.json`.
Updates: Assurance scorecard category.
Links IDs: scenario tags such as `@BEHAVIOR-003` map tests to behavior evidence. The sample generates `ASSURANCE-003` with `refs: ["BEHAVIOR-003"]`, which covers duplicate-charge prevention.

### SARIF

```sh
bottleneck ingest sarif --file reports/codeql.sarif
```

Represents: CodeQL or security scan findings in SARIF format.
Writes: `bottleneck/security/guardrails.json`.
Updates: Security scorecard category.
Links IDs: SARIF `properties.refs`, `properties.bottleneck.refs`, or `tags` can include `BEHAVIOR-*`, `INTENT-*`, or other Bottleneck evidence IDs. The sample creates `SECURITY-001` linked to `BEHAVIOR-003` and `INTENT-001` with a low-severity finding that stays within default thresholds.

### Test Summary

```sh
bottleneck ingest test-summary --file reports/test-summary.json
```

Represents: summarized unit, integration, or end-to-end test results.
Writes: `bottleneck/assurance/results.json`.
Updates: Assurance scorecard category.
Links IDs: evidence entries preserve their `ASSURANCE-*` IDs and `BEHAVIOR-*` refs. The sample preserves `ASSURANCE-001`, `ASSURANCE-002`, and `ASSURANCE-003`.

`cucumber` and `test-summary` both write normalized assurance evidence. Running the second command replaces the same file unless you pass `--merge`.

### Telemetry

```sh
bottleneck ingest telemetry --file reports/telemetry.json
```

Represents: execution signals such as deployment frequency, failure rate, error rate, override rate, adoption, and cost.
Writes: `bottleneck/execution/telemetry.json`.
Updates: Execution scorecard category.
Links IDs: telemetry evidence can reference `BEHAVIOR-*`, `ASSURANCE-*`, and `EXECUTION-*` evidence IDs. The sample preserves `EXECUTION-001` and links execution signals to all three billing behaviors.

Verify the effect:

```sh
bottleneck scorecard
bottleneck trace BEHAVIOR-003
```

After ingesting the Cucumber or test-summary sample, Assurance changes from `Warn` to `Pass` because `BEHAVIOR-003` now has mapped passing test evidence. SARIF updates Security, and telemetry updates Execution. Use `--dry-run --format=json` to parse a report without writing normalized evidence; invalid JSON fails cleanly with a parse error.

## 10. Fix The Evidence Gap Manually

Add mapped assurance evidence for `BEHAVIOR-003` in `bottleneck/assurance/results.json`:

```json
{
  "id": "ASSURANCE-003",
  "refs": ["BEHAVIOR-003"],
  "source": "billing retry duplicate-charge test",
  "status": "pass",
  "summary": "Duplicate retry requests reuse the original idempotency key and do not create a second charge."
}
```

Then add the assurance reference to `BEHAVIOR-003` in `bottleneck/behavior/behavior-spec.md`:

```markdown
Refs:
- INTENT-001
- ASSURANCE-003
```

Re-run:

```sh
bottleneck scorecard
```

Expected fixed posture:

```text
Bottleneck Scorecard
Environment: default
Release Recommendation: Proceed
Primary Bottleneck: None
```

## 11. Break And Restore The Link

To see the diagnosis return, remove `BEHAVIOR-003` from `ASSURANCE-003.refs` or remove `ASSURANCE-003` from the behavior refs, then run:

```sh
bottleneck diagnose
```

Restore the mapping and rerun:

```sh
bottleneck validate
bottleneck scorecard
```

## 12. Add CI

Use the Day-One SaaS workflow example in `examples/github-actions/`:

```sh
mkdir -p .github/workflows
cp examples/github-actions/bottleneck-saas-scorecard.yml .github/workflows/bottleneck.yml
```

The demo-ready project also includes a repository-local workflow at `examples/saas-billing/.github/workflows/bottleneck.yml`.

The workflow builds Bottleneck from source, so it does not require a published action, unavailable secrets, or hardcoded local paths.

Core CI steps:

```yaml
- name: Validate Bottleneck evidence
  run: ./bin/bottleneck validate --env="$BOTTLENECK_ENV" --github-annotations

- name: Generate Bottleneck scorecard
  run: ./bin/bottleneck scorecard --env="$BOTTLENECK_ENV" --format=markdown >> "$GITHUB_STEP_SUMMARY"

- name: Check release readiness
  run: ./bin/bottleneck diagnose --env="$BOTTLENECK_ENV" --gate=release
```

In a pull request, Bottleneck writes the scorecard to the GitHub Actions step summary. Developers see the release recommendation, primary bottleneck, category results, and next action in the check run. The workflow does not post a PR comment; use `examples/github-actions/bottleneck-pr-gate.yml` only when you want the comment workflow.

The release gate step controls whether CI blocks. Warnings can appear in the scorecard without failing the workflow when the selected environment and gate thresholds treat them as non-blocking. Blocking release reasons include missing required assurance, broken traceability, critical security findings, missing required categories, and production gate failures where configured.

GitHub annotation output is already supported and the workflow emits it:

```sh
bottleneck diagnose --format=github
bottleneck validate --github-annotations
```

Tune environment behavior with the workflow input or by editing commands:

```sh
bottleneck scorecard --env=stage --format=markdown
bottleneck diagnose --env=production --gate=release
```

Other workflow examples remain available:

- `bottleneck-validate.yml`
- `bottleneck-scorecard.yml`
- `bottleneck-pr-gate.yml`

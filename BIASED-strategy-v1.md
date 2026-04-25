# BIASED Strategy (v1)

## 1. Purpose

BIASED defines how software systems are designed, validated, secured, and operated in the age of AI.

BIASED replaces ceremony-driven delivery with a capability-driven system that is measurable, enforceable, and observable.

---

## 2. Core Principles

- Intent must be explicit and enforceable
- Behavior must be defined before system design
- Validation is continuous, not a phase
- Security is embedded, not reviewed
- Execution determines truth
- The system is the product

---

## 3. Capability Model

Behavior -> Intent -> Design -> Assurance -> Security -> Execution -> (feedback to Behavior)

A system is considered INVALID if any capability is missing or below threshold.

---

## 4. Capabilities

### B — Behavior (Understanding)

**Definition**
Defines what the system must do and how it is expected to behave.

**Measurements**
- Behavior Coverage (>90%)
- Testability (>95%)
- Ambiguity Rate (<10%)

**Artifacts**
`/biased/behavior/behavior-spec.md`

**Owner**
Product + Engineering

**Triggers**
- New feature
- Unclear requirements
- Validation failures

---

### I — Intent

**Definition**
Defines why the system exists, the outcomes it must achieve, and constraints.

**Measurements**
- Outcome Alignment (100%)
- Success Criteria Coverage (100%)
- Constraint Coverage (>90%)

**Artifacts**
`/biased/intent/intent.md`

**Owner**
Product + Leadership

**Triggers**
- New initiative
- Misalignment with business outcomes
- Low adoption

---

### D — Design

**Definition**
Defines how the system works (architecture, interactions, system composition).

**Measurements**
- Design Coverage (>90%)
- Interaction Coverage (>85%)
- Change Impact Visibility (>80%)

**Artifacts**
`/biased/design/architecture.md`

**Owner**
Engineering

**Triggers**
- Architecture change
- Scaling issues
- Missing traceability

---

### A — Assurance (Validation)

**Definition**
Continuously proves system behavior matches intent using externally executed BDD scenarios and a single results artifact.

**Measurements**
- Behavioral Accuracy (computed from `scenarios_passed / scenarios_total`)
- Defect Escape Rate (<5%)
- Edge Case Coverage (>90%)
- Drift Detection (within threshold)

**Artifacts**
`/biased/assurance/features/*.feature`
`/biased/assurance/results.json`

**Owner**
Engineering + QA

**Triggers**
- Failing BDD results
- New edge cases
- Production defects

---

### S — Security

**Definition**
Ensures system safety, compliance, and governance.

**Measurements**
- Policy Violations (0)
- Guardrail Coverage (>95%)
- Audit Completeness (100%)

**Artifacts**
`/biased/security/guardrails.json`

**Owner**
Security + Engineering

**Triggers**
- Policy updates
- Security incidents
- Pre-release checks

---

### E — Execution (Operations)

**Definition**
Measures real-world performance and feeds results back into the system.

**Measurements**
- Adoption Rate (context-specific)
- Task Success Rate (>90%)
- Reliability (SLA)
- Feedback Loop Speed (fast)
- Production Drift (minimal)

**Artifacts**
`/biased/execution/telemetry.json`

**Owner**
Operations + Product + Engineering

**Triggers**
- Production release
- Performance issues
- User feedback

---

## 5. Configuration Model

BIASED resolves validation thresholds from a single configuration file:

`/biased/config.yaml`

The model is environment-aware:

- `default` defines baseline thresholds
- `dev`, `test`, `stage`, and `production` override only the values they need
- Missing environment values inherit from `default`

Example:

```yaml
environments:
  default:
    assurance:
      min_accuracy: 0.90
      max_failures: 0
    execution:
      max_error_rate: 0.05
      min_adoption: 0.5
  production:
    assurance:
      min_accuracy: 0.95
```

This keeps developer friction low:

- Developers produce one assurance artifact: `/biased/assurance/results.json`
- The system computes metrics from that file
- The selected environment determines the thresholds that apply

---

## 6. Measurement Example

Input artifact:

```json
{
  "scenarios_total": 10,
  "scenarios_passed": 9,
  "scenarios_failed": 1,
  "failures": [
    "user cannot complete checkout"
  ]
}
```

Computed metrics:

- `accuracy = scenarios_passed / scenarios_total`
- `scenarios_failed = scenarios_failed`

For `production`, the effective assurance threshold is currently:

- `min_accuracy = 0.95`
- `max_failures = 0`

## 7. Validation Example

A system FAILS if:

- Missing `/biased/behavior/behavior-spec.md`
- Missing `/biased/intent/intent.md`
- Missing or invalid `/biased/config.yaml`
- Assurance failed scenarios exceed the resolved `max_failures`
- Assurance computed accuracy falls below the resolved `min_accuracy`
- Execution `error_rate` exceeds the resolved `max_error_rate`

Example CLI output for `production`:

```text
Behavior: PASS
Intent: PASS
Design: PASS
Assurance: FAIL (accuracy below threshold)
  accuracy: 0.90 (threshold: 0.95)
  scenarios_failed: 0 (allowed: 0)
Security: PASS
Execution: PASS

System Status: FAIL
Primary Bottleneck: Assurance
Environment: production
```

## 8. Enforcement Model

BIASED is enforced through:

- Git-based artifacts
- Environment-aware CLI validation
- Read-only interpretation commands (`biased explain`, `biased scorecard`)
- CI/CD integration
- External BDD runners writing `/biased/assurance/results.json`
- Threshold inheritance from `default` to the selected environment

## 9. Operating Model

Behavior -> Intent -> Design -> Assurance -> Security -> Execution -> Behavior

`biased validate --env=<environment>` resolves configuration, computes metrics from artifacts, and determines whether the system remains valid.

`biased explain` and `biased scorecard` do not change system state. They interpret the current validation state so teams can understand ownership, bottlenecks, and next actions without mutating artifacts.

Execution reveals truth. Truth updates Behavior and Intent.

## 10. Loops (Continuous System Validation)

### Purpose

Loops define how the system continuously validates and improves itself after deployment.

BIASED does not consider a feature complete at release.

> A feature is valid only while it continues to pass its loops.

This aligns with modern software practices where systems are **continuous, cyclical, and feedback-driven rather than linear** :contentReference[oaicite:0]{index=0}, with feedback loops enabling ongoing improvement and risk reduction :contentReference[oaicite:1]{index=1}.

---

## Core Principle

- Workflows define how features are created  
- Capabilities define what must exist  
- **Loops define how systems remain valid over time**

---

## Feature Lifecycle

A feature progresses through two phases:

### 1. Creation Phase (Linear)

Intent → Behavior → Design → Assurance → Security → Execution

---

### 2. Operational Phase (Continuous)

After deployment, the feature enters continuous loops:

- Assurance Loop
- Execution Loop
- Cost Loop

---

## 10.1 Assurance Loop (Primary Loop)

### Definition

Continuously validates that system behavior matches defined expectations.

### Flow

Execution → Assurance → Behavior/Design update → redeploy → repeat

### Rules

- BDD scenarios must continuously pass
- Any failure indicates system drift
- Failures require immediate correction

### Validation

- FAIL if `scenarios_failed` exceeds the resolved `max_failures`
- FAIL if computed accuracy falls below the resolved `min_accuracy`
- Thresholds are resolved from `/biased/config.yaml` for the selected environment

### Principle

> If Assurance fails in production, the system is invalid.

---

## 10.2 Execution Loop

### Definition

Continuously measures real-world system performance and usage.

### Flow

Execution → Metrics → Analysis → Behavior/Intent refinement → repeat

### Measurements

- Adoption rate
- Task success rate
- Error rate
- Latency

### Validation

- FAIL if `error_rate` exceeds the resolved `max_error_rate`
- WARNING if `adoption_rate` falls below the resolved `min_adoption`

### Principle

> Execution reveals truth.

---

## 10.3 Cost Loop

### Definition

Continuously evaluates and optimizes system cost.

### Flow

Execution → Cost measurement → Design adjustment → Assurance → repeat

### Measurements

- Cost per request
- Total system cost
- Cost trends

### Rules

- FAIL if cost exceeds defined thresholds
- WARNING if cost trends upward

### Principle

> Cost is a constraint applied continuously, not a one-time decision.

---

## 10.4 Loop Enforcement

Loops are enforced through:

- Continuous validation (`biased validate`)
- CI/CD integration
- Execution telemetry
- Environment-aware thresholds from `/biased/config.yaml`

---

## 10.5 System Validity

A system remains valid only if:

- All capabilities pass validation
- All loops continue to pass over time

---

## Final Principle

> Shipping is not completion.  
> Continuous validation is completion.

## Final Statement

BIASED does not define how teams work.
It defines what the system must produce--and enforces it.

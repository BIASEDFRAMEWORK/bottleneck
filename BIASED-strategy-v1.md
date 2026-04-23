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
`/biased/design/interactions.md`

**Owner**
Engineering

**Triggers**
- Architecture change
- Scaling issues
- Missing traceability

---

### A — Assurance (Validation)

**Definition**
Continuously proves system behavior matches intent.

**Measurements**
- Behavioral Accuracy (>95%)
- Defect Escape Rate (<5%)
- Edge Case Coverage (>90%)
- Drift Detection (within threshold)

**Artifacts**
`/biased/assurance/test-cases.json`
`/biased/assurance/results.json`

**Owner**
Engineering + QA

**Triggers**
- Failing tests
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
`/biased/security/policies.md`
`/biased/security/guardrails.json`
`/biased/security/audit-log.json`

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
`/biased/execution/adoption.json`
`/biased/execution/incidents.md`

**Owner**
Operations + Product + Engineering

**Triggers**
- Production release
- Performance issues
- User feedback

---

## 5. Measurement Example

```json
{
  "behavior": 0.88,
  "intent": 0.92,
  "design": 0.85,
  "assurance": 0.91,
  "security": 0.97,
  "execution": 0.76
}
```

## 6. Validation Example

A system FAILS if:

- Missing `/biased/behavior/behavior-spec.md`
- Missing `/biased/intent/intent.md`
- Assurance accuracy < 0.90
- Any required directory missing

Example CLI output:

```text
Behavior: FAIL (missing behavior-spec.md)
Intent: PASS
Design: PASS
Assurance: FAIL (accuracy 0.82 < 0.90)
Security: PASS
Execution: WARNING (low adoption)

System Status: FAIL
Primary Bottleneck: Assurance
```

## 7. Enforcement Model

BIASED is enforced through:

- Git-based artifacts
- CLI validation
- CI/CD integration

## 8. Operating Model

Behavior -> Intent -> Design -> Assurance -> Security -> Execution -> Behavior

Execution reveals truth. Truth updates Behavior and Intent.

## Final Statement

BIASED does not define how teams work.
It defines what the system must produce--and enforces it.

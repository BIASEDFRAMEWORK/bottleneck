# BIASED Validation

## 1. Capability Schemas

### Behavior

Artifact: /biased/behavior/behavior-spec.md

Required structure:

- Must not be empty
- Must contain ## Expected Behavior
- Must contain ## Unacceptable Behavior

### Intent

Artifact: /biased/intent/intent.md

Required structure:

- Must contain ## Outcomes
- Must contain ## Constraints
- Must contain ## Success Criteria

### Design

Artifact: /biased/design/architecture.md

Required structure:

- Must not be empty
- Must contain at least one Markdown section header

### Assurance

Artifact: /biased/assurance/results.json

Required JSON structure. Developers produce only this file; BIASED computes metrics from it:

~~~json
{
  "scenarios_total": 1,
  "scenarios_passed": 1,
  "scenarios_failed": 0,
  "failures": []
}
~~~

### Security

Artifact: /biased/security/guardrails.json

Required JSON structure:

~~~json
{
  "violations": 0
}
~~~

### Execution

Artifact: /biased/execution/telemetry.json

Required JSON structure:

~~~json
{
  "adoption_rate": 0.9,
  "error_rate": 0.01
}
~~~

## 2. Validation Rules

Behavior passes only when behavior-spec.md exists, is not empty, and includes both required behavior sections.

Intent passes only when intent.md exists and includes Outcomes, Constraints, and Success Criteria sections.

Design passes only when architecture.md exists, is not empty, and includes at least one Markdown section header.

Assurance passes only when results.json exists, parses as JSON, includes all required fields, has failed scenarios at or below the configured max_failures threshold, and has calculated accuracy greater than or equal to the configured min_accuracy threshold.

Security passes only when guardrails.json exists, parses as JSON, includes violations, and violations equals 0.

Execution passes when telemetry.json exists, parses as JSON, includes adoption_rate and error_rate, and error_rate is less than or equal to the configured max_error_rate threshold. Execution returns WARNING when adoption_rate is below the configured min_adoption threshold.

## 3. CLI Mapping

biased validate maps each capability to a dedicated validator. Use --env to select environment thresholds:

~~~sh
biased validate --env=production
~~~

- Behavior -> validateBehavior()
- Intent -> validateIntent()
- Design -> validateDesign()
- Assurance -> validateAssurance()
- Security -> validateSecurity()
- Execution -> validateExecution()

The CLI enforces presence checks for required artifacts, schema checks for Markdown and JSON structure, environment-specific threshold checks for assurance accuracy and failures, security violations, execution error rate, and execution adoption.

## 4. Example Output

~~~text
Behavior: PASS
Intent: PASS
Design: PASS
Assurance: FAIL (scenarios_failed > 0)
Security: PASS
Execution: WARNING (low adoption)

System Status: FAIL
Primary Bottleneck: Assurance
~~~

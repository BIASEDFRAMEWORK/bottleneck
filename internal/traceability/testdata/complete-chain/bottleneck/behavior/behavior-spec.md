# Behavior Specification

## Expected Behavior

### BEHAVIOR-001: Complete evidence chain
Critical: true
Refs:
- INTENT-001
- ASSURANCE-001

The CLI shows complete related evidence for this behavior.

### BEHAVIOR-002: Missing assurance chain
Critical: true
Refs:
- INTENT-001

The CLI clearly reports when a behavior has no assurance evidence.

## Unacceptable Behavior

- The CLI must not hide missing downstream evidence behind an aggregate score.

# Behavior Specification

## Expected Behavior

<!-- Replace this sample behavior with the behavior your release must prove. -->

### BEHAVIOR-001: Flag ambiguous financial risk language
Critical: true
Refs:
- INTENT-001
- ASSURANCE-001

When a PDF contains qualified risk language such as "may", "could", "subject to", or "material uncertainty", the summary must preserve that uncertainty and flag it for review.

## Unacceptable Behavior

<!-- Replace these examples with failures your system must prevent. -->

- The system must not summarize ambiguous risk language as a confirmed fact.
- The system must not omit material caveats from the risk summary.

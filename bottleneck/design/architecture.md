# Architecture

<!-- Replace this sample design with the architecture evidence reviewers need. -->

### DESIGN-001: Local PDF risk summarization flow
Refs:
- INTENT-001
- BEHAVIOR-001

The workflow extracts PDF text, identifies candidate risk clauses, asks the summarizer to produce a short risk summary, and runs a post-summary uncertainty check before showing the output to a reviewer.

Key components:
- PDF text extraction
- Risk clause detector
- Summary generator
- Uncertainty flagging check
- Reviewer-facing output

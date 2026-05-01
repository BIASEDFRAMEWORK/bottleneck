# Intent

<!-- Sample app: AI PDF Risk Summarizer. -->

## Outcomes

<!-- Replace this sample outcome with the release outcome your system must prove. -->

### INTENT-001: Summarize financial PDF risk without hiding uncertainty
Refs:
- BEHAVIOR-001

The system must summarize material risk clauses from financial PDFs while preserving uncertainty and caveats that affect release or investment decisions.

## Constraints

<!-- Replace these sample constraints with the boundaries your system must honor. -->

- The system must not invent risk facts that are not present in the source PDF.
- The system must flag ambiguous or qualified risk language instead of rewriting it as certainty.

## Success Criteria

<!-- Replace these sample criteria with measurable release checks for your system. -->

- At least 95% of evaluated summaries preserve material risk caveats.
- 100% of ambiguous risk clauses in the evaluation set are flagged as uncertain.

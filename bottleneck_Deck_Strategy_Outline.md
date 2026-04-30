# bottleneck Deck Strategy Outline

## Purpose
Create a refreshed bottleneck strategy deck guided by the BIASED framework that explains the theory, framework, and operating model for AI-era software delivery. The deck should move away from sounding like an attack on Scrum or Agile and instead position BIASED as the missing evidence layer and operating model organizations need when AI makes code generation cheap, fast, and less deterministic.

## Core Thesis
AI did not simply make software teams faster. AI changed the constraint.

For decades, organizations optimized the SDLC around the assumption that writing code was hard, slow, and expensive. AI changes that. Code generation is becoming cheap. The scarce resources are now context, verification, governance, human judgment, production feedback, and organizational adoption.

BIASED is an evidence framework and operating model for AI-era software delivery. It helps organizations prove that generated software is aligned to business intent, validated against expected behavior, governed by policy, secured before release, and measured in production.

## Recommended Positioning

### Avoid this framing
- Scrum is broken.
- Agile failed.
- Developers are going away.
- AI means we need to replace everything.

### Use this framing instead
- Agile and Scrum gave organizations a useful sense of rhythm, control, and coordination for human-speed software development.
- AI changes the bottleneck from code generation to verification, context, governance, and adoption.
- Existing ceremonies, roles, and metrics were not designed for machine-speed creation and non-deterministic outputs.
- Organizations need a new evidence layer, scorecard, and operating model to manage AI software development responsibly.
- BIASED extends the SDLC with opinionated artifacts, roles, ceremonies, and metrics for the AI era.

## Suggested Deck Title Options

1. **BIASED: The Evidence Framework for AI-Era Software Delivery**
2. **When Code Gets Cheap: A New Operating Model for AI Software Development**
3. **BIASED: Turning AI-Generated Code Into Accountable Outcomes**
4. **The SDLC After Code Generation Stops Being the Bottleneck**
5. **From Agile Control to AI Accountability: Introducing BIASED**

## Target Audience
- Technology executives
- Product leaders
- Engineering leaders
- Security and governance leaders
- Consulting clients evaluating AI adoption
- Non-technology organizations trying to manage AI-enabled software teams
- Enterprise transformation leaders

## Desired Audience Takeaway
By the end of the deck, the audience should believe:

1. AI changes the economics and risk profile of software delivery.
2. More code does not automatically produce more value.
3. Existing SDLC controls are too code-centric and too Jira-centric.
4. Organizations need scorecards, ceremonies, roles, and evidence to manage AI development.
5. BIASED provides a practical operating model and implementation path.
6. The bottleneck CLI is the first implementation of the framework.

---

# Narrative Arc

## Act 1: The Constraint Changed
Explain why AI changes the software delivery system.

## Act 2: The Old Management Model Is Not Enough
Explain why existing roles, ceremonies, and metrics do not provide enough control in AI-enabled delivery.

## Act 3: BIASED Creates an Evidence Layer
Introduce the framework: Behavior, Intent, Assurance, Security, Execution, and Design.

## Act 4: BIASED Becomes Operational
Show roles, ceremonies, artifacts, scorecards, and Git-based evidence.

## Act 5: The CLI Implements the Framework
Position the application as the test, validation mechanism, and implementation layer for the theory.

---

# Proposed Slide Outline

## Slide 1 — Title
**BIASED: The Evidence Framework for AI-Era Software Delivery**

Subtitle: **When code generation becomes cheap, accountability becomes the product.**

Visual direction:
- Minimal black or white background.
- BIASED acronym centered.
- Small subtitle underneath.
- Use existing BIASED color palette and clean Excella-style visual restraint.

Speaker point:
- This is not a talk about replacing developers or replacing Agile.
- This is about what organizations need once AI changes the constraint in software delivery.

---

## Slide 2 — The Old Constraint
**For decades, software delivery was organized around one assumption: code is expensive to create.**

Key points:
- Roles, ceremonies, estimates, backlogs, and controls were shaped around human-speed development.
- Scrum and Agile helped organizations coordinate work, inspect progress, and create management visibility.
- That structure gave many organizations a sense of control, especially non-technology companies managing software delivery.

Visual direction:
- Layered SDLC stack: Customer, Operations, Policy, Product, Code.
- Highlight Code as the historic bottleneck.

---

## Slide 3 — The New Constraint
**AI made code generation faster. It did not make delivery automatically faster.**

Key points:
- Code volume increases.
- PRs get larger.
- Review gets harder.
- Validation becomes more important.
- Security risk scales with generated output.
- Teams need more context, not less.

Visual direction:
- Funnel or pipeline motif: high AI/code input creates pressure on validation, review, security, and operations.

---

## Slide 4 — The Bottleneck Moved
**The bottleneck is no longer just code generation. It is verification.**

Key points:
- Can we prove the generated output is correct?
- Can we prove it matches business intent?
- Can we prove it is secure?
- Can we prove it works in production?
- Can we prove users trust and adopt it?

Visual direction:
- Five checkpoints: Context, Validation, Security, Governance, Production Feedback.

---

## Slide 5 — The Deeper Problem
**Most organizations do not have an evidence layer across the SDLC.**

Key points:
- Requirements live in tickets.
- Policies live in documents.
- Architecture lives in diagrams.
- Security lives in separate tools.
- Test results live in CI.
- Production behavior lives in observability tools.
- Adoption feedback lives in meetings or support channels.
- No single system connects intent to production evidence.

Visual direction:
- Disconnected islands: Jira, GitHub, Security, CI/CD, Docs, Observability, Support.

---

## Slide 6 — The Risk for Enterprises
**Without a new operating model, AI adoption stalls.**

Key points:
- Non-technology companies still need management visibility.
- Leaders need scorecards, roles, ceremonies, and release confidence.
- Developers need guardrails and feedback loops.
- Security needs evidence.
- Product needs intent alignment.
- Governance needs release accountability.

Visual direction:
- Executive dashboard with missing signals.
- Status: unclear, blocked, unknown risk.

---

## Slide 7 — Reframe Agile Without Attacking It
**Agile gave us rhythm. AI requires evidence.**

Key points:
- Scrum and Agile were valuable coordination systems.
- They were not designed for agentic code generation, non-deterministic behavior, or AI-scale output.
- Story points, velocity, and sprint rituals do not prove correctness, safety, adoption, or production value.
- BIASED complements delivery methods by defining the evidence modern teams need.

Visual direction:
- Two-column comparison:
  - Agile/Scrum: cadence, planning, coordination, visibility.
  - BIASED: intent, evidence, verification, governance, production feedback.

---

## Slide 8 — Introduce BIASED
**BIASED is an evidence framework for AI-era software delivery.**

Acronym:
- **B — Behavior**: What the system should and should not do.
- **I — Intent**: Why the system exists and what outcome matters.
- **A — Assurance**: How correctness is tested and continuously evaluated.
- **S — Security**: How risk, policy, and protection are enforced.
- **E — Execution**: How the system performs, operates, and is adopted.
- **D — Design**: How human experience, architecture, and workflow shape outcomes.

Visual direction:
- Six-pillar framework model.
- Use consistent icons.
- Keep it clean and memorable.

---

## Slide 9 — The Framework Promise
**BIASED connects generated work back to intent, risk, and evidence.**

Key points:
- Not just “did the build pass?”
- Did the behavior match intent?
- Did tests validate expected and unacceptable outcomes?
- Did security checks pass?
- Did governance approve the release?
- Did production telemetry confirm the expected result?
- Did adoption data show users trust the system?

Visual direction:
- Traceability chain:
  Intent → Behavior → Design → Code → Tests → Security → Release → Telemetry → Learning.

---

## Slide 10 — The Framework Evidence Loop
**Intent becomes behavior. Behavior becomes tests. Tests become evidence. Evidence informs release. Production informs intent.**

Key loop:
1. Define intent.
2. Specify expected behavior.
3. Design human and system workflows.
4. Generate/build code.
5. Validate with tests and evaluations.
6. Secure and govern release.
7. Observe production behavior.
8. Feed learning back into intent and behavior.

Visual direction:
- Circular loop or infinity loop.
- Emphasize continuous learning.

---

## Slide 11 — Roles in an AI-Era Delivery Model
**AI does not eliminate roles. It changes what roles must own.**

Suggested role model:
- **Intent Owner**: owns outcomes, constraints, and value definition.
- **Behavior Engineer**: translates intent into measurable behavior and edge cases.
- **Design Engineer**: owns workflow, human interaction, adoption signals, and experience.
- **Assurance Engineer**: owns tests, evaluations, drift detection, and evidence quality.
- **Security/Governance Engineer**: owns policy, risk, controls, and release gates.
- **Execution Engineer**: owns CI/CD, telemetry, operations, and production feedback.

Visual direction:
- Modern team map, not hierarchy.
- Avoid making this look like rigid org design; show capabilities.

---

## Slide 12 — Ceremonies for AI Software Delivery
**When output accelerates, rituals must shift from status to evidence.**

Recommended ceremonies:
- Intent Workshop
- Behavior Design Review
- Data & Evaluation Sync
- Behavioral Demo
- Security, Governance & Risk Checkpoint
- Adoption & Change Review
- Release Decision
- Production Learning Review

Visual direction:
- Calendar-style or lifecycle-style view.
- Each ceremony produces artifacts, not just discussion.

---

## Slide 13 — Artifacts Become the System of Control
**The new control surface is Git-based evidence.**

Artifacts:
- intent.md
- behavior-spec.md
- design.md
- assurance results
- security results
- execution telemetry
- governance-card.md
- release decision log
- adoption feedback
- scorecard

Key point:
- Metrics should be derived from Git commits, CI/CD, tests, scans, telemetry, and human review — not manually maintained in Jira.

Visual direction:
- Repository tree or evidence map.

---

## Slide 14 — The bottleneck Scorecard
**Executives and teams need a shared confidence model.**

Scorecard dimensions:
- Behavior
- Intent
- Design
- Assurance
- Security
- Execution
- Governance
- Adoption
- Cost
- Drift

Example output:
- Release Recommendation: Proceed / Block / Conditional
- Primary Constraint: Security / Assurance / Adoption / Governance
- Evidence: linked to tests, scans, telemetry, and approvals

Visual direction:
- Dashboard mockup.
- Keep simple enough for executives.

---

## Slide 15 — Where bottleneck Fits With Existing Tools
**bottleneck does not replace the tools. It connects their signals through the BIASED framework.**

Map:
- Copilot / Codex / Cursor / Claude Code generate code.
- GitHub Actions executes workflows.
- Cucumber validates behavior.
- CodeQL and security tools inspect risk.
- Observability tools measure production.
- bottleneck connects the evidence back to intent, behavior, governance, and release readiness.

Visual direction:
- Hub-and-spoke model with bottleneck as the evidence application.

---

## Slide 16 — bottleneck CLI: The Framework Implementation
**The application validates the theory.**

Key points:
- The framework defines what must be true.
- The CLI checks whether the evidence exists.
- Integrations collect signals from GitHub, CI/CD, tests, security, and telemetry.
- Agents can derive, explain, and update artifacts.
- Scorecards make organizational readiness visible.

Visual direction:
- Framework layer above, CLI implementation below.

---

## Slide 17 — Why This Matters for Enterprises
**AI adoption will stall without management confidence.**

Key points:
- Leaders need confidence that generated software is safe and valuable.
- Teams need a clear process for AI-generated work.
- Governance needs auditable evidence.
- Non-tech companies need operating models they can understand and manage.
- bottleneck gives organizations a way to adopt AI development without losing control.

Visual direction:
- Enterprise operating model diagram.

---

## Slide 18 — The Opinionated Point of View
**The BIASED framework is intentionally opinionated.**

Principles:
- Intent must be explicit.
- Behavior must be measurable.
- Evidence must live close to code.
- Security and governance must be continuous.
- Production feedback must update the system.
- Scorecards must be explainable.
- AI-generated output must never outrun accountability.

Visual direction:
- Manifesto-style slide.

---

## Slide 19 — The Future State
**The SDLC becomes a learning system.**

Key points:
- Production incidents update behavior specs.
- Failed tests update prompts and guardrails.
- Security findings update policy rules.
- Adoption friction updates design.
- Cost trends update execution thresholds.
- AI agents help maintain the evidence layer.

Visual direction:
- Closed-loop operating system.

---

## Slide 20 — Call to Action
**Do not use AI only to generate more code. Use AI to fix the system that turns code into outcomes.**

Closing statement:
- When code gets cheap, correctness gets expensive.
- When output accelerates, accountability must become explicit.
- The future of software delivery is not just AI-generated code.
- The future is evidence-driven delivery.

Visual direction:
- Return to the SDLC stack or framework loop.
- Make it visually consistent with the opening slide.

---

# Appendix Slide Ideas

## Appendix A — Framework vs Traditional Agile Metrics
| Traditional Metric | Limitation | Framework Alternative |
|---|---|---|
| Velocity | Measures output, not value | Intent and behavior alignment |
| Story points | Subjective and often gamed | Evidence-backed readiness |
| Sprint completion | Does not prove production success | Release confidence score |
| Defect counts | Reactive | Continuous assurance signals |
| Status reports | Manual and stale | Git-derived scorecards |

## Appendix B — Example Scorecard
Show an example scorecard for a feature moving from dev to test to production.

## Appendix C — Example Artifacts
Show the relationship between:
- intent.md
- behavior-spec.md
- cucumber results
- CodeQL results
- telemetry.json
- governance-card.md

## Appendix D — Ceremonies and Outputs
Show each ceremony with its artifact output.

## Appendix E — Enterprise Adoption Model
Show maturity levels:
1. AI coding experiments
2. Team-level guardrails
3. Git-based evidence
4. Cross-SDLC scorecards
5. Enterprise operating model

---

# Messaging Changes From Current Deck

## Reduce
- Direct criticism of Scrum Masters.
- “Scrum failed” framing.
- Broad claims that Agile is obsolete.
- Implying bottleneck replaces all existing delivery methods.

## Increase
- Evidence layer framing.
- Verification as the new bottleneck.
- Context engineering.
- Production telemetry.
- Human judgment as scarce resource.
- Enterprise management confidence.
- Scorecards and ceremonies for non-tech organizations.
- GitHub-native implementation path.

## Keep
- “Code got faster. Everything else did not.”
- “AI exposed the system.”
- “The system is the product.”
- “When generation becomes cheap, accountability becomes expensive.”
- Funnel/pipeline visual motif.
- BIASED framework as opinionated and practical.

---

# Recommended Speaker Framing

## Short version
AI does not remove the need for process. It changes what the process must prove.

## Executive version
Organizations used Agile, Scrum, and delivery metrics to create visibility and control over software teams. AI changes the speed and risk profile of software delivery. The BIASED framework gives leaders a new control surface: evidence that connects intent, behavior, security, governance, adoption, and production performance.

## Technical version
The BIASED framework is a Git-native evidence framework. It ingests signals from tests, scans, hooks, CI/CD, telemetry, and agent-generated artifacts, then produces explainable scorecards that help teams decide whether generated software is ready to release.

## Consulting version
bottleneck gives enterprises an opinionated approach for adopting AI software development without losing governance, security, or management visibility.

---

# Visual Design Guidance

## Style
- Minimalist.
- High contrast.
- Few words per slide.
- Strong recurring motifs.
- Use existing bottleneck brand colors.
- Avoid cluttered framework diagrams early in the deck.

## Recommended visual motifs
- SDLC stack.
- Pipeline/funnel showing bottlenecks.
- Evidence loop.
- Six-pillar framework model.
- Git-based artifact tree.
- Scorecard dashboard.
- Connected toolchain map.

## Tone
- Confident.
- Executive-ready.
- Opinionated but not combative.
- Practical, not academic.
- Framework-first, product-enabled.


# bottleneck Product Roadmap Work Items

## Purpose
This roadmap translates the BIASED framework into bottleneck application work items. bottleneck is the implementation and test of the framework. It should validate whether the framework can give enterprise teams an opinionated, Git-native, evidence-driven approach for AI-era software delivery.

## Product Thesis
AI makes code generation cheap. Organizations still need control, visibility, governance, roles, ceremonies, scorecards, and evidence to manage software delivery. Agile and Scrum gave organizations a management model for human-speed development. The BIASED framework gives organizations a management and evidence model for AI-speed development.

bottleneck should prove this thesis by using GitHub hooks, GitHub Actions, CI/CD outputs, security tools, BDD results, telemetry, and agents to derive scorecards and explain whether a feature, service, or system is ready to release.

## Product Positioning
The BIASED framework is not another ticketing tool. It is not another static analysis tool. It is not a replacement for GitHub, Jira, observability, or security tooling.

The BIASED strategy defines the evidence framework. bottleneck is the application that connects those tools back to intent, behavior, design, assurance, security, execution, governance, adoption, cost, and drift.

## Product Layers

1. **Framework Layer**
   - Defines the framework capabilities, ceremonies, roles, and scorecard categories.

2. **Artifact Layer**
   - Stores intent, behavior, design, assurance, security, execution, governance, and adoption artifacts in the codebase.

3. **Integration Layer**
   - Ingests signals from GitHub, tests, security scans, CI/CD, telemetry, and external systems.

4. **Agent Layer**
   - Helps create, update, explain, and reconcile framework artifacts.

5. **Validation Layer**
   - Checks whether evidence exists, thresholds are met, traceability is intact, and release gates are satisfied.

6. **Scorecard Layer**
   - Produces human-readable and machine-readable readiness, risk, and maturity scores.

7. **Operating Model Layer**
   - Supports enterprise ceremonies, roles, release decisions, and organizational adoption.

---

# Roadmap Principles

## Principle 1 — Evidence over opinion
bottleneck should prefer evidence derived from Git, tests, actions, scans, telemetry, and review events over manually entered status.

## Principle 2 — Git-native by default
Artifacts should live with the codebase whenever possible. Git history becomes the audit trail.

## Principle 3 — Integrate, do not replace
bottleneck should ingest from GitHub Actions, Cucumber, CodeQL, dependency review, security scans, telemetry tools, and agent outputs rather than trying to own every workflow.

## Principle 4 — Agents assist, evidence decides
Agents may generate explanations, suggest fixes, draft artifacts, and reconcile metrics. They should not be treated as authoritative without evidence.

## Principle 5 — Scorecards must be explainable
Every score should be traceable to evidence, thresholds, rules, and source artifacts.

## Principle 6 — Enterprise adoption requires rituals
bottleneck should support ceremonies, scorecards, and roles because organizations need process to adopt AI development safely.

---

# Phase 0 — Stabilize the Current CLI Foundation

## Epic 0.1 — Clarify Product Language in the Codebase
**Goal:** Align application terminology with the evolved framework.

### Work items
- [x] Update README to describe the BIASED strategy as an evidence framework for AI-era software delivery.
- [x] Explain the relationship between framework, artifacts, CLI, integrations, agents, and scorecards.
- [x] Clarify that the CLI is the implementation layer, not the entire framework.
- [ ] Add a glossary for Behavior, Intent, Assurance, Security, Execution, and Design.
- [ ] Add example enterprise use cases.

### Acceptance criteria
- [x] README clearly distinguishes strategy/framework from CLI implementation.
- [x] New users understand why the framework exists before seeing commands.
- [ ] Documentation avoids overclaiming maturity.

---

## Epic 0.2 — Prevent False Confidence From Placeholder Artifacts
**Goal:** Ensure initialized templates do not create fake PASS results.

### Work items
- [ ] Detect default placeholder text created by `bottleneck init`.
- [ ] Add warnings when placeholder content remains unchanged.
- [ ] Add a strict mode that fails validation when placeholder content is detected.
- [ ] Include placeholder status in `bottleneck explain`.
- [ ] Include placeholder status in `bottleneck scorecard`.

### Acceptance criteria
- [ ] A newly initialized project should not appear production-ready.
- [ ] Placeholder warnings identify the exact file and section.
- [ ] Tests prove that default templates warn or fail as expected.

---

## Epic 0.3 — Improve Validator Depth
**Goal:** Move from structural validation toward meaningful validation.

### Work items
- [ ] Add minimum content quality rules for intent, behavior, design, assurance, security, and execution artifacts.
- [ ] Validate that intent includes measurable outcomes.
- [ ] Validate that behavior includes expected and unacceptable behaviors.
- [ ] Validate that assurance includes evidence sources, not just pass/fail values.
- [ ] Validate that security includes evidence sources, not just violation counts.
- [ ] Validate that execution includes adoption, error, latency, and telemetry source information.

### Acceptance criteria
- [ ] Validation distinguishes between present, complete, and evidence-backed artifacts.
- [ ] `bottleneck explain` tells users what makes an artifact weak.
- [ ] Tests cover incomplete, placeholder, and evidence-backed artifacts.

---

# Phase 1 — Build the bottleneck Scorecard as the Main Product Surface

## Epic 1.1 — Implement `bottleneck scorecard`
**Goal:** Create the primary human-readable output for bottleneck.

### Work items
- [ ] Add `bottleneck scorecard` command.
- [ ] Display scores for Behavior, Intent, Design, Assurance, Security, and Execution.
- [ ] Add status levels: PASS, WARN, FAIL, UNKNOWN.
- [ ] Show primary constraint blocking release readiness.
- [ ] Show release recommendation: Proceed, Conditional, Block, Unknown.
- [ ] Show evidence count and missing evidence per category.
- [ ] Support JSON output for CI/CD usage.

### Acceptance criteria
- [ ] Users can run `bottleneck scorecard` locally.
- [ ] Scorecard can be consumed in GitHub Actions.
- [ ] Scorecard explains why a category passed, warned, or failed.
- [ ] Tests confirm scorecard output for passing and failing projects.

---

## Epic 1.2 — Add Environment-Aware Scorecards
**Goal:** Support different thresholds by environment without forcing duplicate metrics files.

### Work items
- [ ] Support environments: local, dev, test, stage, production.
- [ ] Implement inheritance so lower environments can inherit thresholds from higher-level configuration.
- [ ] Allow environment-specific overrides.
- [ ] Display effective thresholds in scorecard output.
- [ ] Add tests for inheritance and override behavior.

### Acceptance criteria
- [ ] Users can run `bottleneck scorecard --env dev`.
- [ ] Users can run `bottleneck scorecard --env production`.
- [ ] Missing environment settings inherit from defaults.
- [ ] Scorecard explains which thresholds were inherited and which were overridden.

---

## Epic 1.3 — Create Executive and Engineer Views
**Goal:** Support different audiences without changing the evidence model.

### Work items
- [ ] Add `--view executive` for summarized release readiness.
- [ ] Add `--view engineering` for detailed evidence and remediation.
- [ ] Add `--view governance` for policy, approval, and release gating.
- [ ] Add Markdown output for status reports.
- [ ] Add GitHub Step Summary output support.

### Acceptance criteria
- [ ] Executive view is short and decision-oriented.
- [ ] Engineering view includes remediation guidance.
- [ ] Governance view includes approval and policy evidence.
- [ ] Markdown output can be pasted into a PR comment or release note.

---

# Phase 2 — Add Traceability Across the Framework

## Epic 2.1 — Introduce Framework Evidence IDs
**Goal:** Create stable identifiers for framework artifacts.

### Work items
- [ ] Define ID patterns: INTENT-001, BEHAVIOR-001, DESIGN-001, ASSURANCE-001, SECURITY-001, EXECUTION-001, GOVERNANCE-001.
- [ ] Add optional front matter support to Markdown artifacts.
- [ ] Add validation that IDs are unique.
- [ ] Add validation that references point to existing artifacts.
- [ ] Add `bottleneck trace` command to show relationships.

### Acceptance criteria
- [ ] Every major artifact can have a stable ID.
- [ ] Broken references are detected.
- [ ] `bottleneck trace INTENT-001` shows linked behavior, tests, security, and telemetry.

---

## Epic 2.2 — Map Intent to Behavior
**Goal:** Ensure behavior is explicitly tied to business intent.

### Work items
- [ ] Add schema for intent-to-behavior references.
- [ ] Validate that every expected behavior supports an intent.
- [ ] Validate that every unacceptable behavior is tied to risk or constraint.
- [ ] Add explain output for orphaned behaviors.
- [ ] Add tests for complete and incomplete mappings.

### Acceptance criteria
- [ ] Scorecard detects behavior without intent.
- [ ] Scorecard detects intent without validating behavior.
- [ ] Trace output shows intent-to-behavior coverage.

---

## Epic 2.3 — Map Behavior to Assurance Evidence
**Goal:** Ensure behavior is backed by tests or evaluations.

### Work items
- [ ] Add support for linking behavior IDs to Cucumber scenarios, test IDs, or evaluation cases.
- [ ] Validate that critical behaviors have assurance evidence.
- [ ] Validate that unacceptable behaviors have negative tests or safety checks.
- [ ] Add coverage percentage for behavior-to-test mapping.
- [ ] Add scorecard display for behavior coverage.

### Acceptance criteria
- [ ] Scorecard reports behavior coverage.
- [ ] Critical behavior without tests creates WARN or FAIL based on environment.
- [ ] Tests prove coverage calculations work.

---

## Epic 2.4 — Map Security and Governance to Release Decisions
**Goal:** Connect security findings and governance approvals to release readiness.

### Work items
- [ ] Add release decision artifact.
- [ ] Add governance card artifact.
- [ ] Link security evidence to governance decisions.
- [ ] Validate approval status before production release.
- [ ] Add release recommendation logic based on governance state.

### Acceptance criteria
- [ ] Production scorecard fails or warns when approval is missing.
- [ ] Security findings can block release.
- [ ] Release decision shows supporting evidence.

---

# Phase 3 — GitHub-Native Integrations

## Epic 3.1 — GitHub Actions Integration
**Goal:** Make bottleneck easy to run in CI/CD.

### Work items
- [ ] Create example GitHub Actions workflow for `bottleneck validate`.
- [ ] Create example workflow for `bottleneck scorecard`.
- [ ] Output Markdown to GitHub Step Summary.
- [ ] Support PR comment output.
- [ ] Support build failure based on environment thresholds.
- [ ] Add documentation for local, dev, test, stage, and production gates.

### Acceptance criteria
- [ ] A user can copy a workflow into `.github/workflows/bottleneck.yml`.
- [ ] Scorecard appears in GitHub Actions summary.
- [ ] Pull request can be blocked based on configured evidence thresholds.
- [ ] Docs explain recommended workflow triggers.

---

## Epic 3.2 — GitHub Pull Request Hook Model
**Goal:** Use PR events as evidence sources.

### Work items
- [ ] Detect PR metadata when running inside GitHub Actions.
- [ ] Capture PR size, changed files, labels, reviewers, approvals, and checks.
- [ ] Add optional PR risk signals to scorecard.
- [ ] Add warning for large AI-generated PRs.
- [ ] Add support for requiring framework artifact changes when code changes affect behavior.

### Acceptance criteria
- [ ] Scorecard can include PR context.
- [ ] PR risk can influence warning status.
- [ ] Tests mock GitHub event payloads.

---

## Epic 3.3 — Pre-Commit Hook Support
**Goal:** Shift lightweight checks left without making developers update multiple files manually.

### Work items
- [ ] Provide recommended pre-commit hook configuration.
- [ ] Run placeholder checks locally.
- [ ] Run artifact schema checks locally.
- [ ] Run secret-scan ingestion or validation locally when available.
- [ ] Add docs for using bottleneck with pre-commit frameworks.

### Acceptance criteria
- [ ] Developers can install local bottleneck checks.
- [ ] Local checks are fast.
- [ ] Local checks do not require network access by default.

---

## Epic 3.4 — GitHub Checks API / PR Annotation Support
**Goal:** Make bottleneck findings visible where developers work.

### Work items
- [ ] Output annotations for missing or weak artifacts.
- [ ] Output annotations for broken traceability.
- [ ] Output annotations for failed score categories.
- [ ] Document how to publish annotations from GitHub Actions.

### Acceptance criteria
- [ ] bottleneck findings appear in PR checks.
- [ ] Findings link to affected files where possible.
- [ ] CI output remains readable without annotations.

---

# Phase 4 — Evidence Ingestion

## Epic 4.1 — Cucumber / BDD Ingestion
**Goal:** Use BDD results to update Assurance and Behavior scores.

### Work items
- [ ] Add `bottleneck ingest cucumber --file <path>`.
- [ ] Parse Cucumber JSON results.
- [ ] Map scenarios to behavior IDs using tags.
- [ ] Calculate pass rate, failed scenarios, and behavior coverage.
- [ ] Update or generate assurance evidence file.
- [ ] Add scorecard support for BDD evidence.

### Acceptance criteria
- [ ] Cucumber tags can map to framework behavior IDs.
- [ ] Failed scenarios reduce Assurance score.
- [ ] Missing scenario mappings create warnings.
- [ ] Tests cover successful and failed Cucumber imports.

---

## Epic 4.2 — CodeQL SARIF Ingestion
**Goal:** Use CodeQL results to update Security score.

### Work items
- [ ] Add `bottleneck ingest codeql --file <path>`.
- [ ] Parse SARIF results.
- [ ] Count findings by severity.
- [ ] Support threshold rules by environment.
- [ ] Link security findings to scorecard.
- [ ] Add remediation hints in `bottleneck explain`.

### Acceptance criteria
- [ ] CodeQL findings affect Security score.
- [ ] High-severity findings can block production.
- [ ] Tests cover SARIF parsing and threshold behavior.

---

## Epic 4.3 — Dependency Review / Supply Chain Ingestion
**Goal:** Include dependency and software supply chain risk.

### Work items
- [ ] Add ingestion for dependency review output.
- [ ] Add support for vulnerability severity counts.
- [ ] Add support for license risk counts.
- [ ] Add support for new dependency risk warnings.
- [ ] Include dependency risk in Security and Governance score.

### Acceptance criteria
- [ ] Dependency findings are visible in scorecard.
- [ ] Production thresholds can block high-risk dependency changes.
- [ ] Tests cover dependency ingestion.

---

## Epic 4.4 — Test Coverage and Unit Test Ingestion
**Goal:** Use standard test outputs as assurance signals.

### Work items
- [ ] Support common coverage formats where practical.
- [ ] Add generic JSON ingestion for test summary data.
- [ ] Track total tests, failed tests, skipped tests, and coverage percentage.
- [ ] Support minimum thresholds by environment.
- [ ] Add scorecard display for test health.

### Acceptance criteria
- [ ] Test evidence can be ingested without manual editing.
- [ ] Failed tests affect Assurance score.
- [ ] Coverage thresholds are environment-aware.

---

## Epic 4.5 — Telemetry Ingestion
**Goal:** Connect production behavior back to Execution and Adoption scores.

### Work items
- [ ] Add generic telemetry ingestion format.
- [ ] Support adoption rate, error rate, latency, rollback rate, override rate, and support friction.
- [ ] Add threshold configuration by environment.
- [ ] Add production feedback section to scorecard.
- [ ] Add ability to mark telemetry as synthetic, staging, or production.

### Acceptance criteria
- [ ] Execution score is derived from telemetry evidence.
- [ ] Adoption data can create warnings even when tests pass.
- [ ] Production telemetry can trigger learning-loop recommendations.

---

# Phase 5 — Governance, Release, and Enterprise Control

## Epic 5.1 — Governance Card Artifact
**Goal:** Add a first-class governance artifact.

### Work items
- [ ] Create `bottleneck/governance/governance-card.md` template.
- [ ] Include policy constraints, approval status, risk level, exceptions, and evidence links.
- [ ] Validate governance card completeness.
- [ ] Show governance status in scorecard.
- [ ] Allow governance to block production release.

### Acceptance criteria
- [ ] Governance card exists after init or upgrade.
- [ ] Missing governance approval affects production readiness.
- [ ] Governance evidence is explainable.

---

## Epic 5.2 — Release Decision Log
**Goal:** Capture release decisions as auditable evidence.

### Work items
- [ ] Create `bottleneck/release/release-decision.md` or JSON equivalent.
- [ ] Include release recommendation, approver, date, environment, risks, and exceptions.
- [ ] Link release decision to scorecard snapshot.
- [ ] Add validation for production release decision.
- [ ] Add command to generate draft release decision.

### Acceptance criteria
- [ ] Production release requires release decision evidence.
- [ ] Release decision references scorecard status.
- [ ] Exceptions are visible and auditable.

---

## Epic 5.3 — Exception and Waiver Management
**Goal:** Allow enterprises to manage risk without hiding it.

### Work items
- [ ] Add waiver schema.
- [ ] Include expiration dates and owners.
- [ ] Link waivers to scorecard warnings or failures.
- [ ] Prevent expired waivers from suppressing failures.
- [ ] Add waiver display in governance view.

### Acceptance criteria
- [ ] Waivers can downgrade or explain failures only when valid.
- [ ] Expired waivers fail validation.
- [ ] Scorecard clearly shows accepted risk.

---

# Phase 6 — Agent Capabilities

## Epic 6.1 — `bottleneck explain` Agent Enhancement
**Goal:** Use agents to explain failures and recommend fixes without changing evidence automatically.

### Work items
- [ ] Generate plain-language explanation of scorecard failures.
- [ ] Suggest artifact updates.
- [ ] Suggest missing tests, security checks, or telemetry signals.
- [ ] Suggest GitHub Actions configuration improvements.
- [ ] Preserve deterministic non-agent fallback.

### Acceptance criteria
- [ ] Users can understand why a score failed.
- [ ] Agent output is clearly advisory.
- [ ] Existing non-agent behavior remains stable.

---

## Epic 6.2 — Artifact Drafting Agent
**Goal:** Help teams create initial framework artifacts from existing project context.

### Work items
- [ ] Generate draft intent.md from README, issues, or project docs.
- [ ] Generate draft behavior-spec.md from acceptance criteria, user stories, or BDD files.
- [ ] Generate draft design.md from architecture docs and repository structure.
- [ ] Generate draft governance-card.md from security policies and scans.
- [ ] Require human review before accepting generated artifacts.

### Acceptance criteria
- [ ] Agent-generated artifacts are marked as draft.
- [ ] Human review is required before artifacts count as approved evidence.
- [ ] Generated artifacts include source references.

---

## Epic 6.3 — Traceability Agent
**Goal:** Suggest missing links between intent, behavior, tests, security, and telemetry.

### Work items
- [ ] Analyze existing artifacts for likely relationships.
- [ ] Recommend traceability links.
- [ ] Detect orphaned tests, behaviors, and requirements.
- [ ] Suggest Cucumber tag mappings.
- [ ] Suggest missing telemetry signals.

### Acceptance criteria
- [ ] Agent can propose traceability without applying it automatically.
- [ ] User can review and accept recommendations.
- [ ] Broken traceability is explained in scorecard and explain output.

---

## Epic 6.4 — Release Readiness Agent
**Goal:** Provide enterprise-friendly release readiness summaries.

### Work items
- [ ] Summarize scorecard results for executives.
- [ ] Summarize blockers for engineering teams.
- [ ] Summarize governance risks for approvers.
- [ ] Draft release notes based on evidence.
- [ ] Draft PR comments with scorecard summary.

### Acceptance criteria
- [ ] Release readiness summaries are evidence-backed.
- [ ] Summary links to source categories and artifacts.
- [ ] Agent does not override scorecard logic.

---

# Phase 7 — Ceremonies and Operating Model Support

## Epic 7.1 — Ceremony Templates
**Goal:** Convert the framework ceremonies into usable artifacts.

### Work items
- [ ] Create templates for Intent Workshop.
- [ ] Create templates for Behavior Design Review.
- [ ] Create templates for Data & Evaluation Sync.
- [ ] Create templates for Behavioral Demo.
- [ ] Create templates for Security, Governance & Risk Checkpoint.
- [ ] Create templates for Adoption & Change Review.
- [ ] Create templates for Release Decision.
- [ ] Create templates for Production Learning Review.

### Acceptance criteria
- [ ] `bottleneck init` can include ceremony templates.
- [ ] Users understand what each ceremony produces.
- [ ] Templates map to scorecard categories.

---

## Epic 7.2 — Role-Based Guidance
**Goal:** Help organizations understand who owns what in AI-era delivery.

### Work items
- [ ] Define role guidance for Intent Owner.
- [ ] Define role guidance for Behavior Engineer.
- [ ] Define role guidance for Design Engineer.
- [ ] Define role guidance for Assurance Engineer.
- [ ] Define role guidance for Security/Governance Engineer.
- [ ] Define role guidance for Execution Engineer.
- [ ] Map roles to artifacts, ceremonies, and scorecard sections.

### Acceptance criteria
- [ ] Documentation explains ownership without requiring rigid titles.
- [ ] Roles map cleanly to framework capabilities.
- [ ] Enterprise clients can use role guidance for transformation planning.

---

## Epic 7.3 — Enterprise Adoption Maturity Model
**Goal:** Provide a staged path for clients adopting bottleneck.

### Work items
- [ ] Define maturity levels from basic AI coding to enterprise evidence-driven delivery.
- [ ] Create scorecard maturity view.
- [ ] Add recommended next steps per maturity level.
- [ ] Add sample adoption roadmap for enterprises.

### Acceptance criteria
- [ ] bottleneck can explain where a team is on the adoption curve.
- [ ] Scorecard can recommend next maturity steps.
- [ ] Maturity model aligns with framework and product capabilities.

---

# Phase 8 — Cost, Drift, and Production Learning

## Epic 8.1 — Cost Score
**Goal:** Treat cost as a first-class AI-era delivery signal.

### Work items
- [ ] Add cost artifact or telemetry fields.
- [ ] Track cost per task, token cost, model cost, infrastructure cost, and rework cost where available.
- [ ] Add cost thresholds by environment.
- [ ] Add cost trend to scorecard.
- [ ] Add explain output for cost warnings.

### Acceptance criteria
- [ ] Cost can affect Execution or Financial Stability score.
- [ ] Cost thresholds are configurable.
- [ ] Scorecard shows cost evidence source.

---

## Epic 8.2 — Drift Score
**Goal:** Detect when system behavior changes over time.

### Work items
- [ ] Add drift artifact or telemetry fields.
- [ ] Support model drift, data drift, behavior drift, and prompt drift concepts.
- [ ] Track drift history over time.
- [ ] Add environment-aware drift thresholds.
- [ ] Add drift findings to scorecard and release recommendations.

### Acceptance criteria
- [ ] Drift can trigger WARN or FAIL.
- [ ] Drift evidence is linked to source data.
- [ ] Production drift can trigger a learning-loop recommendation.

---

## Epic 8.3 — Production Learning Loop
**Goal:** Ensure production findings update the framework artifacts.

### Work items
- [ ] Detect production issues that should update behavior specs.
- [ ] Detect user overrides that should update design or intent.
- [ ] Detect failed edge cases that should update eval sets.
- [ ] Detect security findings that should update governance rules.
- [ ] Add `bottleneck learn` command to propose artifact updates from production evidence.

### Acceptance criteria
- [ ] Production feedback can generate recommended artifact updates.
- [ ] Recommendations are reviewable before commit.
- [ ] Scorecard shows stale artifacts when production evidence has not been reconciled.

---

# Phase 9 — GitHub App / Marketplace Direction

## Epic 9.1 — Package bottleneck for GitHub Adoption
**Goal:** Make bottleneck easy for enterprise teams to adopt inside GitHub.

### Work items
- [ ] Publish reusable GitHub Action.
- [ ] Provide starter workflow templates.
- [ ] Provide sample repository demonstrating framework and CLI.
- [ ] Provide example PR comments and Step Summary outputs.
- [ ] Provide enterprise configuration examples.

### Acceptance criteria
- [ ] New team can adopt bottleneck by copying one workflow and running init.
- [ ] GitHub-native usage is documented clearly.
- [ ] Sample repo demonstrates the full evidence loop.

---

## Epic 9.2 — GitHub App Concept
**Goal:** Explore a GitHub App as the enterprise product surface.

### Work items
- [ ] Define GitHub App capabilities.
- [ ] Support PR checks and annotations.
- [ ] Support scorecard comments.
- [ ] Support repository-level bottleneck status.
- [ ] Support organization-level scorecard aggregation.
- [ ] Explore permissions model.

### Acceptance criteria
- [ ] GitHub App concept has architecture and security model.
- [ ] Enterprise use cases are documented.
- [ ] Roadmap identifies what belongs in CLI vs GitHub App.

---

# Phase 10 — Product Demonstration Scenarios

## Epic 10.1 — Demo Scenario: AI-Generated Feature With Missing Evidence
**Goal:** Show why generated code alone is not enough.

### Work items
- [ ] Create demo repo with generated feature.
- [ ] Leave behavior evidence incomplete.
- [ ] Run scorecard to show warning/failure.
- [ ] Use explain to show remediation.
- [ ] Add BDD evidence and rerun scorecard.

### Acceptance criteria
- [ ] Demo clearly shows bottleneck improving release confidence.
- [ ] Audience understands that passing build is not enough.

---

## Epic 10.2 — Demo Scenario: Security Blocks AI-Generated PR
**Goal:** Show security and governance as continuous evidence.

### Work items
- [ ] Create PR with AI-generated dependency or insecure code example.
- [ ] Run CodeQL or dependency review.
- [ ] Ingest findings into bottleneck.
- [ ] Show Security score failure.
- [ ] Show governance release block.

### Acceptance criteria
- [ ] Demo shows integration with GitHub-native security tools.
- [ ] Release recommendation is evidence-backed.

---

## Epic 10.3 — Demo Scenario: Production Telemetry Changes Release Confidence
**Goal:** Show that validation continues after deployment.

### Work items
- [ ] Create telemetry sample with high error rate or low adoption.
- [ ] Ingest telemetry.
- [ ] Show Execution warning.
- [ ] Use learning loop to recommend behavior/design updates.

### Acceptance criteria
- [ ] Demo proves bottleneck is not just pre-commit or CI validation.
- [ ] Production feedback updates the system.

---

# Near-Term Recommended Backlog

## Must do next
- [ ] Implement placeholder detection.
- [ ] Implement `bottleneck scorecard` MVP.
- [ ] Add environment-aware thresholds.
- [ ] Add traceability IDs.
- [ ] Add Cucumber ingestion.
- [ ] Add CodeQL SARIF ingestion.
- [ ] Add GitHub Actions Step Summary output.
- [ ] Add governance-card artifact.

## Should do soon
- [ ] Add dependency review ingestion.
- [ ] Add release decision log.
- [ ] Add PR comment output.
- [ ] Add telemetry ingestion.
- [ ] Add adoption artifact.
- [ ] Add agent-assisted explain output.

## Later
- [ ] Add GitHub App concept.
- [ ] Add organization-level dashboards.
- [ ] Add enterprise maturity model automation.
- [ ] Add cost and drift history.
- [ ] Add production learning loop.

---

# Product Success Metrics

## Framework validation metrics
- Teams can map intent to behavior.
- Behavior can be traced to tests and evidence.
- Security and governance can block release.
- Production telemetry can update scorecards.
- Scorecards are explainable to both executives and engineers.

## Application adoption metrics
- Time to initialize bottleneck in a repo.
- Time to first scorecard.
- Number of evidence sources ingested.
- Number of scorecard categories populated automatically.
- Reduction in manually maintained status artifacts.
- Number of PRs with bottleneck scorecard output.

## Enterprise value metrics
- Increased release confidence.
- Reduced unreviewed AI-generated risk.
- Improved traceability from intent to production.
- Faster governance review.
- Better visibility into AI software delivery readiness.
- Higher adoption of AI-enabled development practices.

---

# Strategic Product Statement
bottleneck should become the opinionated evidence and management layer that enterprises need when AI changes software development from human-speed code creation to machine-speed output plus human-accountable verification, guided by the BIASED strategy.

The application validates the framework by proving that scorecards, ceremonies, roles, and artifacts can be derived from real development signals instead of manually maintained status reports.

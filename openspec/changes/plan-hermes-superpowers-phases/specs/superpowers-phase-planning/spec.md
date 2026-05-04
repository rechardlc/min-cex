## ADDED Requirements

### Requirement: Plan Phase 1 Superpowers
The roadmap SHALL define Phase 1 superpowers as foundational backend capabilities required before CEX trading complexity.

#### Scenario: Phase 1 capability tree is reviewed
- **WHEN** a developer reviews the Phase 1 plan
- **THEN** it includes go-zero API service wiring, RPC service wiring, typed config, response/error conventions, basic account model, local Docker Compose runtime, tests, and learning documentation

### Requirement: Plan Phase 2 Superpowers
The roadmap SHALL define Phase 2 superpowers as CEX business-chain capabilities built on top of Phase 1.

#### Scenario: Phase 2 capability tree is reviewed
- **WHEN** a developer reviews the Phase 2 plan
- **THEN** it includes order creation/cancel/query, fund freeze/release, ledger records, matching-core integration, trade settlement, market data publishing, wallet deposit/withdraw simulation, and event consumers

### Requirement: Specify Tasks Per Phase
The OpenSpec tasks SHALL group implementation work by Phase 1 and Phase 2 and preserve dependency order within each phase.

#### Scenario: Implementation starts from tasks
- **WHEN** implementation begins from `tasks.md`
- **THEN** Phase 1 can be executed before Phase 2 without needing hidden prerequisite work

### Requirement: Include Learning Outputs
Each phase SHALL include documentation and progress-tracking tasks that explain the implemented go-zero boundaries with frontend-friendly mental models.

#### Scenario: A frontend-first developer reviews progress
- **WHEN** the developer reads updated progress docs
- **THEN** they can map gateway handlers, logic, svc, config, pkg, and RPC boundaries to familiar frontend concepts

### Requirement: Keep Future Implementation Testable
Each phase SHALL include verification tasks for service behavior, domain invariants, and documentation completeness.

#### Scenario: Phase task completion is audited
- **WHEN** a maintainer checks the completed tasks
- **THEN** there are tests or runnable verification steps for core behavior and docs showing what changed

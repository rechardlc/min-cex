## ADDED Requirements

### Requirement: Summarize Current Project Requirements
The planning artifacts SHALL summarize the current mini CEX backend training requirements from existing roadmap, go-zero demo, and matching-engine notes.

#### Scenario: Existing materials are consolidated
- **WHEN** a maintainer reads the OpenSpec change
- **THEN** they can identify the target user, learning goal, current repository baseline, Phase 1 scope, Phase 2 scope, and explicit non-goals without reading every source document first

### Requirement: Define Phase Boundaries
The planning artifacts SHALL separate Phase 1 and Phase 2 by learning risk and implementation dependency.

#### Scenario: Phase 1 is understood
- **WHEN** Phase 1 tasks are reviewed
- **THEN** they focus on go-zero engineering foundation, gateway/account RPC flow, account basics, local runtime, and documentation

#### Scenario: Phase 2 is understood
- **WHEN** Phase 2 tasks are reviewed
- **THEN** they focus on CEX business flow, including orders, balance freeze/release, matching integration, settlement, market data, wallet simulation, and event processing

### Requirement: Preserve Scope Control
The planning artifacts SHALL define non-goals so advanced exchange topics do not enter early implementation by accident.

#### Scenario: Advanced topics are requested early
- **WHEN** a future task tries to add production-grade matching recovery, real chain signing, Kubernetes, or service mesh during Phase 1
- **THEN** the task is treated as out of scope unless a new OpenSpec change explicitly expands the phase

### Requirement: Provide Acceptance Criteria
The planning artifacts SHALL provide observable acceptance criteria for each phase.

#### Scenario: Phase completion is checked
- **WHEN** a phase is marked complete
- **THEN** its runnable behavior, documented learning outcome, and project progress update can be verified from the repository

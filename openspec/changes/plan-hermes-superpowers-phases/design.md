## Context

The current repository contains three useful but separate threads:

- `docs/GOLANG_CEX_BACKEND_ROADMAP.md` defines the long-range goal: build a mini CEX backend that can support interview storytelling within 3 to 4 months.
- `go-zero-demo/` contains an SME-style go-zero learning skeleton with `gateway`, `account-rpc`, shared `pkg`, deployment docs, and progress tracking.
- `docs/撮合引擎.md` captures deeper matching-engine concepts that should inform later phases without pulling them too early into the first implementation milestone.

The planning model is:

- Hermes: requirement synthesis, scope control, acceptance criteria.
- Superpowers: capability growth map, split into Phase 1 and Phase 2.
- OpenSpec: executable task list for each phase.

## Goals / Non-Goals

**Goals:**

- Convert scattered learning and architecture notes into one phased OpenSpec plan.
- Make Phase 1 small enough to finish and explain: go-zero API/RPC, account service, typed responses, basic local runtime, and docs.
- Make Phase 2 meaningful for CEX backend interviews: order flow, balance freeze/release, ledger events, matching-core integration, market data, wallet simulation, and event-driven processing.
- Preserve the current `go-zero-demo` service boundaries and learning-first documentation style.

**Non-Goals:**

- Do not build a production exchange in this change.
- Do not implement low-latency matching, WAL/snapshot recovery, multi-symbol sharding, real chain signing, or compliance workflows in Phase 1.
- Do not introduce Kubernetes, service mesh, full observability stacks, or heavy distributed architecture before the core business path is visible.

## Decisions

### Decision 1: Use one OpenSpec change with two phases

Use a single change named `plan-hermes-superpowers-phases` instead of two independent OpenSpec changes.

Rationale: the user asked for one flow from Hermes to Superpowers to OpenSpec. A single change keeps requirements, design, and tasks connected while still separating implementation tasks by phase.

Alternative considered: create `phase-1` and `phase-2` changes. That would make implementation tracking stricter later, but it would duplicate context before the roadmap is stable.

### Decision 2: Treat Phase 1 as the engineering foundation

Phase 1 focuses on `go-zero-demo` basics:

- gateway API structure
- account-rpc contract and logic
- response/error package usage
- local Docker Compose baseline
- PostgreSQL schema plan and minimal persistence
- progress docs and learning notes

Rationale: this matches the current repository state and gives the user a backend mental model comparable to frontend routing, state containers, shared libs, and environment configuration.

Alternative considered: start with matching engine integration. That is tempting for CEX identity, but it creates too much business complexity before service boundaries, database transactions, and account assets are clear.

### Decision 3: Treat Phase 2 as the CEX business chain

Phase 2 adds the interview-relevant trading chain:

- trade-rpc for orders
- account freezing/releasing
- internal ledger entries
- orderbook/matching integration
- trade settlement
- market data and WebSocket publishing
- wallet deposit/withdraw simulation
- MQ/job/consumer structure

Rationale: the long-term roadmap emphasizes being able to explain the core business chain, not only write isolated Go examples.

Alternative considered: split wallet and matching into later phases. The plan keeps them in Phase 2 but scopes them as training-grade simulations, not production-grade infrastructure.

### Decision 4: Keep the go-zero SME structure

New services should follow the existing `apps/<service>/{cmd,api|desc,internal}` and shared `pkg/` conventions.

Rationale: the repository already documents this pattern, and it maps cleanly to small and medium enterprise teams.

Alternative considered: merge everything into one monolith for speed. That would be faster at first but would teach fewer service-boundary lessons.

## Risks / Trade-offs

- [Risk] Phase 2 may become too broad. → Mitigation: keep every Phase 2 task tied to one business chain scenario: place order, freeze funds, match, settle, publish result.
- [Risk] Learning docs can drift from implementation. → Mitigation: every phase includes explicit documentation and progress tracker tasks.
- [Risk] Matching-engine depth can overshadow account correctness. → Mitigation: Phase 2 acceptance requires account freeze/release and ledger invariants before advanced matching optimizations.
- [Risk] Introducing MQ too early can add ceremony. → Mitigation: Phase 2 starts with simple in-process or local broker events and only documents NATS/Kafka as swappable infrastructure.

## Migration Plan

1. Land this OpenSpec plan as the roadmap baseline.
2. During implementation, complete Phase 1 tasks first and update `go-zero-demo/PROGRESS_TRACKER.md`.
3. Start Phase 2 only after Phase 1 can be run locally and explained from gateway to account-rpc.
4. If scope needs to change, split Phase 2 into a later OpenSpec change before implementation begins.

Rollback is documentation-only for this change: remove the change directory or archive/supersede it with a clearer plan.

## Open Questions

- Should Phase 1 persist real account data immediately, or first use an in-memory repository to teach service wiring?
- Should Phase 2 use NATS or Kafka for the first event-driven pass?
- Should `exchange-demo/orderbook` be reused directly, adapted into a package, or reimplemented in `go-zero-demo` as a learning exercise?

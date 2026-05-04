## 1. Hermes Requirement Summary

- [ ] 1.1 Read `docs/GOLANG_CEX_BACKEND_ROADMAP.md`, `go-zero-demo/DIRECTORY_RULES.md`, `go-zero-demo/docs/LEARNING_PATH.md`, and `docs/撮合引擎.md` before implementation starts.
- [ ] 1.2 Create or update a requirements summary doc that records target user, project goal, current baseline, Phase 1 scope, Phase 2 scope, and non-goals.
- [ ] 1.3 Define phase-level acceptance criteria for runnable behavior, learning output, and progress documentation.
- [ ] 1.4 Update `go-zero-demo/PROGRESS_TRACKER.md` with the selected phase plan and current next step.

## 2. Phase 1: Go-Zero Engineering Foundation

- [ ] 2.1 Verify the existing `gateway` API service can start locally and document the run command.
- [ ] 2.2 Verify the existing `account-rpc` service can start locally and document the run command.
- [ ] 2.3 Connect `gateway` to `account-rpc` through typed config and `svc` dependency wiring.
- [ ] 2.4 Add a gateway account query endpoint that calls `account-rpc` instead of returning isolated mock data.
- [ ] 2.5 Standardize API success and error responses through `pkg/response` and `pkg/xerr`.
- [ ] 2.6 Add a minimal account domain model covering user id, currency, available balance, frozen balance, and account status.
- [ ] 2.7 Add a PostgreSQL migration plan or migration files for users/accounts using the repository's chosen migration style.
- [ ] 2.8 Add repository or data-access code for account read operations.
- [ ] 2.9 Add local Docker Compose services for PostgreSQL and any required service configuration.
- [ ] 2.10 Add focused tests or runnable verification steps for gateway-to-account-rpc account query behavior.
- [ ] 2.11 Update learning docs to explain handler, logic, svc, config, pkg, and RPC boundaries with frontend-friendly analogies.
- [ ] 2.12 Update `go-zero-demo/PROGRESS_TRACKER.md` when Phase 1 behavior is runnable and documented.

## 3. Phase 1 Acceptance Check

- [ ] 3.1 Confirm a developer can run gateway and account-rpc locally from documented commands.
- [ ] 3.2 Confirm a gateway endpoint can return account data through the RPC boundary.
- [ ] 3.3 Confirm account data structure and persistence choices are documented.
- [ ] 3.4 Confirm progress docs show Phase 1 completion status and Phase 2 next step.

## 4. Phase 2: CEX Business Chain

- [ ] 4.1 Add a `trade-rpc` service skeleton following the existing go-zero directory rules.
- [ ] 4.2 Define order create, cancel, and query RPC contracts for `trade-rpc`.
- [ ] 4.3 Add order domain models covering symbol, side, type, price, quantity, filled quantity, status, and timestamps.
- [ ] 4.4 Add account freeze/release RPC operations needed by order creation, cancellation, and settlement.
- [ ] 4.5 Add ledger records for freeze, release, trade debit, trade credit, deposit, and withdrawal events.
- [ ] 4.6 Implement the order creation flow: validate request, freeze funds, create order, and emit an order event.
- [ ] 4.7 Decide whether to reuse `exchange-demo/orderbook` or create a learning-focused matching package in `go-zero-demo`.
- [ ] 4.8 Integrate a minimal matching flow that accepts order events and produces trade events.
- [ ] 4.9 Implement settlement logic that consumes trade events and updates account balances plus ledger entries.
- [ ] 4.10 Add market data projection for latest trades and orderbook snapshot query.
- [ ] 4.11 Add a WebSocket or streaming endpoint for market data updates if it fits the selected go-zero structure.
- [ ] 4.12 Add a wallet simulation service or job for deposit confirmation and withdrawal status changes.
- [ ] 4.13 Add local MQ/job/consumer structure using NATS, Kafka, or an intentionally documented local substitute.
- [ ] 4.14 Add tests or runnable verification for the full training path: place order, freeze funds, match, settle, and publish result.
- [ ] 4.15 Update learning docs to explain the CEX business chain from gateway request to ledger and market data event.
- [ ] 4.16 Update `go-zero-demo/PROGRESS_TRACKER.md` when Phase 2 behavior is runnable and documented.

## 5. Phase 2 Acceptance Check

- [ ] 5.1 Confirm order creation freezes the correct asset and records an auditable ledger event.
- [ ] 5.2 Confirm cancellation releases the correct frozen balance and updates order status.
- [ ] 5.3 Confirm a matched trade settles both sides without negative available or frozen balances.
- [ ] 5.4 Confirm market data can expose either latest trade data, orderbook snapshot data, or both.
- [ ] 5.5 Confirm wallet simulation records deposit or withdrawal state transitions without real chain dependencies.
- [ ] 5.6 Confirm docs explain the implemented chain well enough for interview storytelling.

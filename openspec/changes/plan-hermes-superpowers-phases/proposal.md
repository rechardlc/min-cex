## Why

当前仓库已经有 CEX 后端学习路线、go-zero 目录骨架、gateway/account-rpc 示例和撮合引擎笔记，但它们还没有被整理成可执行的阶段计划。现在需要用 Hermes 先沉淀需求，再用 Superpowers 把能力成长拆成一期/二期，最后通过 OpenSpec 指定每期任务，避免后续实现时范围漂移。

## What Changes

- 新增一份需求整理能力：把现有文档与项目状态归纳为 mini CEX 后端训练项目的目标、非目标、用户画像、业务范围和验收标准。
- 新增一份 Superpowers 分期规划能力：将学习与实现拆成一期“go-zero 工程化与账户基础”和二期“CEX 交易链路与事件化扩展”。
- 新增每期 OpenSpec 任务：一期任务聚焦可运行、可讲清楚的 API/RPC/账户/数据库基础；二期任务聚焦订单、冻结、撮合、成交、行情、钱包模拟与事件链路。
- 保持现有 `go-zero-demo` 中小型企业目录规则，不引入过早的重型微服务复杂度。

## Capabilities

### New Capabilities
- `hermes-requirement-summary`: 汇总 mini CEX 后端训练项目的一期/二期需求、范围边界、交付物和验收标准。
- `superpowers-phase-planning`: 将后端工程化、账户资产、订单撮合、事件链路和面试表达拆成分期能力树与任务计划。

### Modified Capabilities
- None.

## Impact

- Affected docs: `openspec/changes/plan-hermes-superpowers-phases/**`, `go-zero-demo/docs/**`, `go-zero-demo/PROGRESS_TRACKER.md`.
- Affected code later: `go-zero-demo/apps/gateway`, `go-zero-demo/apps/account-rpc`, future `trade-rpc`, `wallet-rpc`, jobs/consumers, shared `pkg`, and `deploy`.
- APIs later: gateway REST APIs, account RPC contracts, future trade/wallet RPC contracts, possible WebSocket market data endpoint.
- Dependencies later: PostgreSQL, Redis, NATS or Kafka, Docker Compose, and go-zero service tooling.

# Hermes / Superpowers 分期计划

## Hermes 需求总结

当前目标是把 `go-zero-demo` 从学习骨架推进成一个可面试讲述的 mini CEX 后端训练项目。

目标用户：

- 有前端经验，正在转 Go / go-zero 后端。
- 已经理解基础 Go，但需要补齐后端工程化、数据库、RPC、账户资产、订单交易链路。
- 希望最终能讲清楚一个 CEX 后端从 API 请求到资产变更、订单撮合、成交结算、行情推送的主链路。

当前基线：

- 已有 `gateway` API 服务骨架。
- 已有 `account-rpc` RPC 服务骨架。
- 已有 `pkg/response`、`pkg/xerr`、`pkg/types` 公共包示例。
- 已有 `deploy/docker-compose.yml` 和学习进度文档。
- 已有 CEX 后端路线图和撮合引擎笔记。

非目标：

- 一期不做生产级撮合引擎。
- 一期不做真实链上签名、提现广播和合规系统。
- 一期不引入 Kubernetes、服务网格或复杂监控体系。
- 二期仍以训练项目为主，不承诺生产交易所级别的性能和容灾。

## Superpowers 一期：go-zero 工程化与账户基础

一期目标是把基础后端能力跑通。前端类比：这一步像先把路由、页面级 action、全局 provider、共享 utils 和环境配置串起来，而不是一上来写复杂交易页面。

核心能力：

- 看懂并运行 gateway API 服务。
- 看懂并运行 account-rpc 服务。
- 让 gateway 通过 RPC 调 account-rpc。
- 建立统一响应和错误结构。
- 建立账户资产的最小领域模型。
- 接入 PostgreSQL 的最小数据结构或迁移计划。
- 补齐本地运行和验证文档。

一期验收：

- 本地能启动 gateway 和 account-rpc。
- gateway 有一个接口可以通过 RPC 返回账户数据。
- 文档能解释 handler、logic、svc、config、pkg、RPC 的职责。
- `PROGRESS_TRACKER.md` 更新一期状态。

## Superpowers 二期：CEX 业务链路

二期目标是让项目具备 CEX 后端面试表达价值。前端类比：这一步像从“页面能请求接口”升级到“完整业务流”，包含下单、状态流转、事件通知、数据投影和用户可见反馈。

核心能力：

- 增加 trade-rpc，支持订单创建、撤销、查询。
- 增加账户冻结和释放能力。
- 增加内部账本流水。
- 接入或改造撮合核心，产生成交事件。
- 根据成交事件完成结算。
- 提供行情查询或 WebSocket 推送。
- 增加钱包充值/提现模拟。
- 增加 MQ/job/consumer 结构，串起事件驱动链路。

二期验收：

- 下单会冻结正确资产。
- 撤单会释放正确冻结资产。
- 成交会更新双方余额和账本流水。
- 行情接口或流式端点能暴露成交或订单簿信息。
- 钱包模拟能记录充值/提现状态变化。
- 文档能讲清楚从 gateway 到 ledger/market data 的完整业务链路。

## OpenSpec 执行入口

本计划已经落到 OpenSpec change：

```text
openspec/changes/plan-hermes-superpowers-phases/
```

执行任务时优先看：

```text
openspec/changes/plan-hermes-superpowers-phases/tasks.md
```


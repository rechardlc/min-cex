# go-zero 学习进度记录

## 当前目标

按 Hermes / Superpowers 分期计划推进 mini CEX 后端训练项目：一期先跑通 `go-zero` 工程化与账户基础，二期再进入订单、冻结、撮合、结算、行情和钱包模拟链路。

## 阶段进度

| 阶段 | 内容 | 状态 | 备注 |
| --- | --- | --- | --- |
| 01 | 理解顶层目录设计 | 已完成 | 已生成学习骨架 |
| 02 | 理解 API 服务结构 | 进行中 | 先从 `apps/gateway` 开始 |
| 03 | 理解 RPC 服务结构 | 未开始 | 学完 API 再进入 |
| 04 | 理解公共包拆分 | 未开始 | 重点看 `pkg/` |
| 05 | 理解本地运行与部署 | 未开始 | 重点看 `deploy/` |
| 06 | Hermes / Superpowers 分期规划 | 已完成 | 已落到 `openspec/changes/plan-hermes-superpowers-phases` |
| 07 | 一期：go-zero 工程化与账户基础 | 未开始 | 先跑通 gateway -> account-rpc |
| 08 | 二期：CEX 业务链路 | 未开始 | 一期完成后再进入 |

## 本周学习记录

### Day 1

- [x] 生成 go-zero 学习目录骨架
- [x] 生成 go-zero 专用 skill：`go-zero-sme-frontend-bridge`
- [x] 将 skill 正式放入仓库根级 `.agents/skills/`
- [ ] 阅读 `DIRECTORY_RULES.md`
- [ ] 阅读 `docs/LEARNING_PATH.md`
- [ ] 看懂 `apps/gateway` 每个目录做什么

### Day 2

- [ ] 看懂 `gateway.api`
- [ ] 看懂 `main.go`
- [ ] 看懂 `ServiceContext`
- [ ] 看懂 `handler` 和 `logic` 的关系

### Day 3

- [ ] 看懂 `account.proto`
- [ ] 看懂 `account-rpc` 的 `main.go`
- [ ] 看懂 `zrpc` 服务的基本职责

## 学习心得

### 已理解

- 待补充

### 仍然模糊

- 待补充

### 下一步

- 先看 `docs/HERMES_SUPERPOWERS_PHASE_PLAN.md`
- 再按 `../openspec/changes/plan-hermes-superpowers-phases/tasks.md` 执行一期任务
- 用 `go-zero-sme-frontend-bridge` 约束后续新增内容都同步更新进度

# go-zero 学习路径

## 分期执行入口

当前学习路线已经整理成 Hermes / Superpowers 分期计划：

- 需求总结与一期/二期说明：`docs/HERMES_SUPERPOWERS_PHASE_PLAN.md`
- OpenSpec 任务入口：`../../openspec/changes/plan-hermes-superpowers-phases/tasks.md`

建议先完成一期：跑通 `gateway -> account-rpc`。这条链路可以类比前端里的 route action 调用内部 service，再由 app context/provider 注入依赖。

## 第 1 步：先理解 go-zero 在项目中的位置

你可以先把 `go-zero` 理解成一套 Go 后端脚手架与工程组织方案，它帮助你更快搭 REST API、RPC 服务和项目分层。

## 第 2 步：先学 API 服务

重点看：

- `apps/gateway/api/gateway.api`
- `apps/gateway/cmd/api/main.go`
- `apps/gateway/internal/handler`
- `apps/gateway/internal/logic`
- `apps/gateway/internal/svc`

## 第 3 步：再学 RPC 服务

重点看：

- `apps/account-rpc/desc/account.proto`
- `apps/account-rpc/cmd/rpc/main.go`
- `apps/account-rpc/internal/logic`
- `apps/account-rpc/internal/svc`

## 第 4 步：补公共能力

重点看：

- `pkg/xerr`
- `pkg/response`
- `pkg/types`

## 第 5 步：开始自己加业务

建议顺序：

1. 健康检查接口
2. 用户查询接口
3. 用户 RPC 查询
4. 登录态或 token
5. 数据库接入
6. Redis 接入

## 第 6 步：使用项目 skill 辅助学习

项目内已经加入技能：

- `.agents/skills/go-zero-sme-frontend-bridge`

这个 skill 会约束后续 go-zero 相关工作遵守三件事：

- 新增内容时同步更新进度文档
- 用前端类比去解释 Go 和 go-zero 分层
- 尽量贴近中小型企业的 go-zero 工程规范

## 学习提醒

- 先看目录职责，再看代码细节
- 先跑通样例，再加自己的业务
- 不要一开始就把所有中间件都接进来

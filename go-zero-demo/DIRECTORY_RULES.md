# go-zero-demo 目录规则

## 1. 顶层职责

- `apps/`：业务服务入口，一般一个目录对应一个服务
- `pkg/`：跨服务复用代码
- `deploy/`：本地运行、容器、部署编排
- `scripts/`：初始化、生成、辅助脚本
- `docs/`：学习笔记、补充说明、阶段规划

## 2. 服务内通用规则

以 `go-zero` 服务为例，推荐保持下面结构：

```text
service-name/
  cmd/                # 程序入口
  api/ 或 desc/       # API / Proto 定义
  internal/
    config/           # 配置结构和配置文件
    handler/          # API handler，仅做参数接收和响应返回
    logic/            # 业务逻辑
    svc/              # ServiceContext，统一注入依赖
```

## 3. 分层约束

- `handler` 不写复杂业务
- `logic` 聚焦单个用例
- `svc` 只做依赖组织
- `config` 只定义配置，不写业务逻辑
- `pkg` 只放公共能力，不放具体业务流程

## 4. 中小型企业项目演进建议

### 第一阶段

- 一个 `gateway` API 服务
- 一个 `account-rpc` 用户/账户服务

### 第二阶段

- 增加 `trade-rpc`
- 增加 `wallet-rpc`
- 增加 `job` 或 `consumer` 目录

### 第三阶段

- 增加 `mq/` 事件驱动
- 增加 `observability/` 监控与链路追踪
- 增加 `test/` 集成测试与压测

## 5. 学习时重点关注

- go-zero 如何组织 API 服务
- go-zero 如何组织 RPC 服务
- `handler -> logic -> svc` 的依赖流向
- 配置如何加载
- 公共包和业务包如何拆分


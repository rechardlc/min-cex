# Golang CEX 后端转型学习计划

适用背景：

- 8 年前端开发经验。
- 已具备 Golang 基础能力。
- 目标方向：Web3 / CEX 后端开发。
- 当前短板：后端工程化、数据库、缓存、消息队列、交易所业务、链上交互、真实项目经验。

这份计划的目标不是“学完全部 Web3”，而是帮助你在 3 到 4 个月内做出一个可面试的 Go CEX 后端训练项目，并能讲清楚核心业务链路。

## 1. 职业目标定位

你的目标岗位可以先定位为：

```text
Go Web3 后端工程师
CEX 后端工程师
交易所业务后端工程师
钱包 / 资产服务后端工程师
订单 / 账户 / 行情服务后端工程师
```

第一阶段不建议直接把目标定成“顶级撮合引擎工程师”。撮合引擎是交易系统里很深的一块，需要大量数据结构、低延迟、并发模型和交易规则积累。

更现实的切入路径是：

```text
前端开发者
  -> Go 后端工程师
  -> Web3 链上交互
  -> CEX 业务后端
  -> 交易系统方向深挖
```

## 2. 总体学习路线

完整路线：

```text
Go 后端工程化
  -> PostgreSQL / Redis / MQ
  -> CEX 账户与资产模型
  -> 订单系统与资金冻结
  -> 订单簿与撮合引擎
  -> 成交结算与内部账本
  -> 行情服务与 WebSocket
  -> 链上充值扫描与提现广播
  -> Docker Compose 模拟真实环境
  -> 测试、压测、故障演练
  -> 简历项目与面试表达
```

最终作品建议命名：

```text
mini-cex-backend
```

它应该包含：

```text
用户系统
账户资产
内部账本
订单管理
资金冻结
撮合引擎
成交记录
行情查询
WebSocket 推送
充值扫描
提现广播
PostgreSQL
Redis
NATS / Kafka
Docker Compose
测试和 README
```

## 3. 推荐技术栈

你的目标是用 go-zero 学习最小可面试的 Go Web3 / CEX 后端，所以技术栈建议围绕“能做出真实业务链路”选择，而不是追求大而全。

主线技术栈：

```text
Go + go-zero + PostgreSQL + Redis + NATS + go-ethereum + Anvil/Ganache + Docker Compose
```

这套技术栈能覆盖最小 CEX 后端项目需要的关键能力：

```text
API 服务
RPC 服务
用户与账户
订单与账本
撮合事件
行情推送
链上充值扫描
提现广播
本地环境模拟
测试与压测
```

### 3.1 服务框架

主选：

```text
go-zero
```

Web3 / CEX 核心作用：

```text
构建 API 服务
构建内部 RPC 服务
统一项目结构
生成 handler / logic / svc
管理配置
内置中间件和服务治理能力
适合拆分 account、order、wallet、marketdata 等服务
```

go-zero 在你的训练项目中可以这样使用：

```text
api-service：
  对外 REST API，下单、查账户、查订单、充值提现查询

account-rpc：
  账户余额、冻结、释放、账本流水

order-rpc：
  创建订单、撤单、订单状态

wallet-rpc：
  充值扫描结果入账、提现状态更新
```

备选：

```text
Gin / Echo
```

区别：

```text
go-zero：
  偏微服务框架，带代码生成、API 描述、RPC 体系和较强工程约束。

Gin / Echo：
  偏轻量 HTTP 框架，更自由、更通用，适合单体 API 或小服务。
```

你的选择：

```text
学习主线用 go-zero。
Gin / Echo 只需要能看懂和简单使用即可。
```

### 3.2 服务通信

主选：

```text
go-zero zrpc
```

Web3 / CEX 核心作用：

```text
API 服务调用账户服务
订单服务调用账户冻结接口
结算服务调用账本更新接口
钱包服务调用账户入账接口
```

备选：

```text
gRPC 原生
```

区别：

```text
go-zero zrpc：
  在 gRPC 上封装了 go-zero 的工程体系，和 goctl、配置、服务结构结合更紧。

gRPC 原生：
  更通用、更底层，跨框架使用更自由，但项目结构和治理能力要自己搭。
```

建议：

```text
训练项目先用 go-zero zrpc。
理解底层仍然是 gRPC，这样以后切换到原生 gRPC 不困难。
```

### 3.3 数据库

主选：

```text
PostgreSQL
```

Web3 / CEX 核心作用：

```text
存用户
存账户余额
存订单
存成交
存账本流水
存充值提现记录
存链上事件
通过事务保证资产一致性
通过唯一索引保证幂等
```

备选：

```text
MySQL
```

区别：

```text
PostgreSQL：
  类型能力强，事务和复杂查询能力好，适合训练账本、事件、订单等核心模型。

MySQL：
  国内业务系统使用也非常广，生态成熟，很多交易所业务团队也会用。
```

建议：

```text
学习项目优先 PostgreSQL。
如果目标公司明确使用 MySQL，再补 MySQL 事务、索引和锁。
```

### 3.4 数据访问

主选：

```text
sqlc + pgx
```

Web3 / CEX 核心作用：

```text
手写 SQL
生成类型安全 Go 代码
精确控制事务
精确控制行锁
适合账户余额、订单、账本这种高一致性业务
```

备选：

```text
GORM
```

区别：

```text
sqlc + pgx：
  SQL 可控，适合核心交易业务，面试时也更容易讲清楚事务和锁。

GORM：
  开发快，适合后台 CRUD，但核心账本和订单逻辑容易隐藏 SQL 细节。
```

建议：

```text
账户、订单、账本、充值提现：sqlc + pgx。
管理后台或简单配置表：可以了解 GORM。
```

### 3.5 数据库迁移

主选：

```text
golang-migrate
```

Web3 / CEX 核心作用：

```text
管理 users、accounts、orders、trades、deposits、withdrawals 等表结构变更
保证本地、测试、生产环境数据库结构一致
方便 Docker Compose 一键初始化环境
```

备选：

```text
goose
```

区别：

```text
golang-migrate：
  使用广，CLI 和 Go library 都成熟，适合项目标准化迁移。

goose：
  也很常见，使用简单，适合中小项目。
```

### 3.6 缓存与限流

主选：

```text
Redis + go-redis
```

Web3 / CEX 核心作用：

```text
缓存交易对配置
缓存行情快照
API 限流
登录 token / session 辅助
提现幂等 key
WebSocket 订阅状态
轻量 Pub/Sub
```

备选：

```text
Redis Streams
```

区别：

```text
Redis 普通缓存 / PubSub：
  适合缓存、限流、临时广播，但 PubSub 消息不持久。

Redis Streams：
  可以作为轻量消息队列，支持消费组和消息保留，但不如 Kafka/NATS 专门。
```

建议：

```text
缓存和限流用 Redis。
消息事件主线优先用 NATS。
```

### 3.7 消息队列 / 事件流

主选：

```text
NATS
```

Web3 / CEX 核心作用：

```text
订单创建事件
撤单事件
成交事件
结算事件
行情事件
充值确认事件
提现状态事件
```

典型链路：

```text
order-service
  -> publish OrderCommand

matching-service
  -> consume OrderCommand
  -> publish TradeEvent / BookEvent

settlement-service
  -> consume TradeEvent
  -> update account ledger

marketdata-service
  -> consume BookEvent
  -> push WebSocket
```

备选：

```text
Kafka
```

区别：

```text
NATS：
  轻量，部署简单，适合训练项目和中小型事件驱动服务。

Kafka：
  更重，适合高吞吐、可回放、长期保存的大规模事件流。
```

建议：

```text
最小可面试项目用 NATS。
面试前了解 Kafka 的 topic、partition、consumer group、offset、重复消费和顺序性。
```

### 3.8 链上交互

主选：

```text
go-ethereum
```

Web3 / CEX 核心作用：

```text
连接 Ethereum RPC
查询区块高度
查询 ETH / ERC20 余额
解析 ERC20 Transfer 事件
构造提现交易
签名交易
广播交易
查询 receipt
处理 nonce
处理 gas
处理确认数
```

备选：

```text
abigen 生成的合约 binding
```

区别：

```text
go-ethereum 原生 ABI / ethclient：
  更底层，能理解 logs、topics、交易、receipt 的细节。

abigen binding：
  根据 ABI 生成类型安全调用代码，写合约调用更方便，但需要先理解底层概念。
```

建议：

```text
先用 go-ethereum 原生方式实现 ERC20 Transfer 扫描。
再用 abigen 调 ERC20 balanceOf / transfer。
```

### 3.9 本地链与合约开发辅助

主选：

```text
Anvil
```

Web3 / CEX 核心作用：

```text
本地启动 EVM 链
部署测试 ERC20
模拟充值
模拟提现
快速出块
可控测试账户
```

备选：

```text
Ganache
```

区别：

```text
Anvil：
  Foundry 生态的一部分，速度快，适合现代 Solidity / EVM 开发测试。

Ganache：
  老牌本地链工具，资料多，很多教学项目使用。
```

建议：

```text
新项目优先 Anvil。
当前仓库如果已有 Ganache 使用方式，也可以继续保留。
```

### 3.10 WebSocket 行情

主选：

```text
go-zero API + gorilla/websocket
```

Web3 / CEX 核心作用：

```text
推送订单簿 depth
推送最新成交 trades
推送 ticker
推送用户订单状态
推送用户资产变动
```

备选：

```text
nhooyr/websocket
```

区别：

```text
gorilla/websocket：
  资料多，使用广，适合快速实现训练项目。

nhooyr/websocket：
  API 更现代，context 支持更自然，也适合生产服务。
```

建议：

```text
训练项目用 gorilla/websocket 即可。
重点放在订阅模型、心跳、断线、消息序列号。
```

### 3.11 日志、配置、监控

主选：

```text
go-zero 内置日志与配置
Prometheus
```

Web3 / CEX 核心作用：

```text
记录订单请求
记录撮合事件
记录结算失败
记录充值扫描高度
记录提现广播状态
监控 RPC 错误率
监控 MQ 堆积
监控 WebSocket 连接数
```

备选：

```text
zap
OpenTelemetry
```

区别：

```text
go-zero 内置日志：
  和框架结合好，训练项目够用。

zap：
  高性能结构化日志库，适合更精细的日志控制。

Prometheus：
  偏指标监控，例如 QPS、延迟、错误数。

OpenTelemetry：
  偏链路追踪和统一可观测性，适合多服务调用链。
```

建议：

```text
第一版使用 go-zero 内置日志。
第二版补 Prometheus 指标。
OpenTelemetry 放到进阶阶段。
```

### 3.12 测试、压测与本地环境

主选：

```text
testing + testify
Docker Compose
k6
```

Web3 / CEX 核心作用：

```text
测试订单簿撮合
测试余额冻结
测试成交结算幂等
测试充值事件去重
测试提现状态机
一键启动 PostgreSQL / Redis / NATS / Anvil
压测下单接口和行情推送
```

备选：

```text
testcontainers-go
hey
```

区别：

```text
Docker Compose：
  适合本地完整环境模拟，直观，方便写 README。

testcontainers-go：
  适合集成测试时自动拉起 PostgreSQL、Redis 等依赖。

k6：
  更适合写脚本化压测场景。

hey：
  更轻量，适合快速测 HTTP QPS。
```

### 3.13 最小可面试技术栈清单

如果只选一套，使用：

```text
Go
go-zero
PostgreSQL
sqlc + pgx
golang-migrate
Redis + go-redis
NATS
go-ethereum
Anvil
Docker Compose
gorilla/websocket
testing + testify
k6
```

这套最小技术栈对应的项目模块：

```text
go-zero API：
  用户、账户、订单、充值提现查询

go-zero zrpc：
  account-rpc、order-rpc、wallet-rpc

PostgreSQL：
  订单、账户、账本、充值提现、链上事件

Redis：
  限流、缓存、短期幂等 key、行情快照

NATS：
  订单事件、成交事件、结算事件、行情事件

go-ethereum：
  ERC20 充值扫描、提现交易广播、receipt 查询

Anvil：
  本地链和测试 ERC20

Docker Compose：
  一键启动本地真实环境

WebSocket：
  行情推送和用户订单状态推送
```

第二套备选技术栈：

```text
Go
Gin / Echo
PostgreSQL
pgx
Redis
Kafka
go-ethereum
Ganache
Docker Compose
```

两套区别：

```text
go-zero + NATS：
  更适合你学习微服务拆分、RPC、事件驱动和工程规范。

Gin/Echo + Kafka：
  更自由、更通用，Kafka 更贴近大型事件流系统，但初期复杂度更高。
```

你的学习建议：

```text
主线只做 go-zero + NATS。
面试前了解 Gin/Echo 和 Kafka 的基本概念即可。
```

### 3.14 本项目技术栈与企业常见替换关系

这份路线使用的是“最小可面试、适合个人训练”的技术栈。真实企业会根据团队历史、规模、性能要求、安全要求和基础设施成熟度做替换。

你需要记住：训练项目的重点不是证明“企业一定这么用”，而是证明你理解每个组件在 CEX / Web3 后端里的职责，并且知道生产环境可能如何升级。

| 本项目技术栈 | 企业常见替换 | 替换原因 | Web3 / CEX 核心作用 |
| --- | --- | --- | --- |
| `go-zero` | `Gin` / `Echo` / `Kratos` / 原生 `net/http` / 自研框架 | 不同团队有不同 Go 服务框架规范；大厂常有内部框架 | API、RPC、服务分层、配置、日志、中间件 |
| `go-zero zrpc` | 原生 `gRPC` / `Connect RPC` / 自研 RPC | 跨语言、服务治理、内部基础设施兼容 | 订单、账户、钱包、结算等内部服务调用 |
| `PostgreSQL` | `MySQL` / 云数据库 | 国内业务系统 MySQL 很常见；企业会沿用已有 DBA 和运维体系 | 用户、账户、订单、账本、充值提现等强一致数据 |
| `sqlc + pgx` | `GORM` / `sqlx` / `database/sql` / 自研 DAO | 团队开发习惯不同；CRUD 系统可能更偏 ORM | 核心交易业务需要可控 SQL、事务、行锁、幂等 |
| `golang-migrate` | `goose` / `Atlas` / 企业数据库发布平台 | 企业可能有统一数据库变更流程 | 管理表结构变更，保证环境一致 |
| `Redis` | `KeyDB` / 云 Redis / 企业缓存平台 | 性能、成本、运维和高可用要求不同 | 缓存、限流、行情快照、短期幂等 key |
| `NATS` | `Kafka` / `RocketMQ` / `Pulsar` / `RabbitMQ` | 大规模事件流、持久化、回放、生态兼容需求 | 订单事件、成交事件、结算事件、行情事件 |
| `go-ethereum` | 链厂商 SDK / 第三方 RPC SDK / 自研链交互库 | 多链支持、企业封装、节点服务商差异 | ERC20 充值扫描、提现广播、receipt 查询、链上状态同步 |
| `Anvil` | `Ganache` / Hardhat Network / 测试网 / 自建 devnet | 团队合约工具链不同 | 本地链测试、模拟充值提现、部署测试 ERC20 |
| `Docker Compose` | `Kubernetes` / Helm / Docker Swarm / 云容器平台 | 生产环境需要弹性伸缩、服务发现、滚动发布、运维治理 | 本地模拟多服务，生产部署服务集群 |
| `gorilla/websocket` | `nhooyr/websocket` / 自研推送网关 / 长连接平台 | 高并发连接、统一接入层、内部推送平台 | 行情推送、订单状态推送、资产变动推送 |
| 本地私钥 `.env` | `KMS` / `HSM` / `Vault` / `MPC` / 多签 | 生产私钥安全要求极高，不能明文保存 | 提现签名、热钱包管理、资产安全 |
| 单节点 RPC | 自建 RPC 集群 / 第三方 RPC / 负载均衡 / 归档节点 | 高可用、限流、容灾、链上历史查询 | 扫链、发交易、查余额、查 receipt |
| `testing + testify` | 企业测试平台 / `testcontainers-go` / mock 平台 | 自动化测试、集成测试、CI/CD 规范 | 验证撮合、账本、充值、提现的正确性 |
| `k6` | `JMeter` / `wrk` / `hey` / 企业压测平台 | 团队工具习惯和压测体系不同 | 压测下单接口、行情推送、钱包扫描吞吐 |
| go-zero 内置日志 | `zap` / `zerolog` / ELK / Loki | 结构化日志、日志聚合、查询和告警 | 排查订单、结算、充值提现和链上交易问题 |
| 简单 Prometheus 指标 | Prometheus + Grafana / OpenTelemetry / Jaeger / Tempo | 多服务链路追踪、指标监控、性能分析 | 监控撮合延迟、结算积压、RPC 错误、提现 pending |

面试时可以这样表达：

```text
我的训练项目使用 go-zero + PostgreSQL + Redis + NATS + go-ethereum，是为了在个人环境里模拟 CEX 后端的 API、RPC、事件流、账本和链上交互。

真实企业里，这些组件都可能根据团队基础设施替换：
go-zero 可以替换为 Gin、Kratos 或自研框架；
NATS 可以替换为 Kafka、RocketMQ 或 Pulsar；
PostgreSQL 可以替换为 MySQL；
本地 Anvil 可以替换为自建节点或第三方 RPC；
本地私钥会替换为 KMS、HSM、MPC 或多签方案。

我关注的不是某个框架本身，而是账户、订单、撮合、结算、钱包、行情这些模块的职责边界，以及事务、幂等、事件驱动和链上确认这些核心问题。
```

如果你只记最重要的替换关系，优先记这 5 个：

```text
go-zero -> Gin / Kratos / 自研框架
NATS -> Kafka / RocketMQ / Pulsar
PostgreSQL -> MySQL
Anvil -> 自建 RPC / 第三方 RPC / 测试网
本地私钥 -> KMS / HSM / MPC / 多签
```

## 4. 16 周学习计划

### 第 1-2 周：Go 后端工程化

目标：

- 从“会 Go 语法”升级到“能写标准后端服务”。

重点内容：

```text
HTTP API
路由
中间件
参数校验
统一错误返回
日志
配置管理
context
优雅关闭
PostgreSQL 连接
项目分层
单元测试
Dockerfile
```

训练项目模块：

```text
用户服务
```

实现接口：

```text
POST /users/register
POST /users/login
GET /users/me
GET /health
```

要求：

```text
JWT 鉴权
密码 hash
PostgreSQL 存用户
统一错误码
请求日志
Docker 启动
基础单元测试
```

面试表达目标：

```text
我能用 Go 搭建标准后端服务，包含路由、中间件、配置、日志、数据库、鉴权和测试。
```

### 第 3-4 周：账户资产与数据库事务

CEX 后端最重要的基础之一是资产一致性。

重点内容：

```text
PostgreSQL 表设计
事务
行锁
唯一索引
幂等
状态机
资产流水
可用余额
冻结余额
```

核心表：

```text
users
assets
accounts
account_ledgers
```

账户模型：

```text
user_id
asset
available_balance
frozen_balance
created_at
updated_at
```

流水模型：

```text
id
user_id
asset
change_amount
balance_after
business_type
business_id
created_at
```

实现接口：

```text
GET /accounts
POST /admin/deposit/mock
POST /accounts/transfer/mock
GET /account-ledgers
```

必须做到：

```text
余额变更必须使用事务
每次余额变化必须写流水
同一个 business_id 不能重复入账
不能出现负余额
```

面试表达目标：

```text
我理解 CEX 的账户资产模型，可用余额和冻结余额分离，余额变化必须通过账本流水记录，并用事务和幂等保证一致性。
```

### 第 5-6 周：订单系统与资金冻结

目标：

- 实现 CEX 下单前的核心业务检查。

重点内容：

```text
交易对
订单表
订单状态机
下单参数校验
余额校验
买单冻结 quote asset
卖单冻结 base asset
撤单释放冻结资产
```

订单状态：

```text
NEW
PARTIALLY_FILLED
FILLED
CANCELED
REJECTED
```

订单表字段：

```text
id
user_id
symbol
side
type
price
quantity
filled_quantity
status
created_at
updated_at
```

注意：

```text
不要使用 float64 表示价格和数量。
使用整数或定点数模型。
```

示例：

```text
1 ETH = 100000000 units
1 USDT = 1000000 units
```

实现接口：

```text
POST /orders
DELETE /orders/:id
GET /orders
GET /orders/:id
```

面试表达目标：

```text
我知道下单不是直接进入撮合，前面要做参数校验、余额校验、资金冻结、订单落库，然后再进入撮合流程。
```

### 第 7-8 周：订单簿与撮合引擎

目标：

- 掌握 CEX 撮合核心。

当前仓库的 `orderbook/orderbook.go` 可以作为入门材料，但需要继续改造。

重点内容：

```text
订单簿
买盘 bids
卖盘 asks
价格档 price level
价格优先
时间优先
限价单主动撮合
市价单逐档成交
部分成交
完全成交
跨价格档成交
成交事件
```

关键改造：

```text
买入限价单：
  如果 best ask <= buy price，先撮合
  剩余未成交数量再挂入买单簿

卖出限价单：
  如果 best bid >= sell price，先撮合
  剩余未成交数量再挂入卖单簿
```

建议设计：

```text
一个 symbol 一个 orderbook
撮合核心单线程处理命令
输入 OrderCommand
输出 TradeEvent / OrderUpdatedEvent
```

必须测试：

```text
限价单直接成交
限价单部分成交后挂单
市价单跨价格档成交
完全成交后价格档移除
同价格订单时间优先
```

面试表达目标：

```text
我理解价格优先、时间优先；限价单可能立即成交，也可能部分成交后挂单；市价单会按最优对手盘逐档吃单。
```

### 第 9-10 周：成交结算与内部账本

目标：

- 理解撮合和结算的分离。

撮合引擎只负责：

```text
订单匹配
生成成交
更新订单剩余数量
输出成交事件
```

结算服务负责：

```text
更新买卖双方余额
释放冻结资产
写成交记录
写账户流水
更新订单状态
保证幂等
```

示例：

```text
交易对：ETH/USDT
成交：买方买入 1 ETH，价格 2000 USDT

买方：
  frozen USDT 减少 2000
  available ETH 增加 1

卖方：
  frozen ETH 减少 1
  available USDT 增加 2000
```

必须做到：

```text
trade_id 不能重复结算
订单更新、余额更新、流水写入必须在事务内
结算失败可重试
重复消费不会导致重复加钱
```

面试表达目标：

```text
我知道撮合和结算是两个阶段，撮合产生成交事件，结算服务消费成交事件并用事务更新订单、余额和流水。
```

### 第 11 周：行情服务与 WebSocket

目标：

- 理解撮合事件如何变成行情数据。

实现接口：

```text
GET /markets
GET /orderbook/:symbol
GET /trades/:symbol
WS /ws
```

WebSocket 推送内容：

```text
orderbook depth
recent trades
ticker
```

推荐链路：

```text
matching engine
  -> publish TradeEvent / BookEvent
  -> market data service
  -> websocket clients
```

可以使用：

```text
Redis Pub/Sub
NATS
Kafka
```

面试表达目标：

```text
我知道撮合事件不仅用于结算，也用于行情服务。行情服务通常会维护快照和增量推送。
```

### 第 12 周：链上充值与提现

目标：

- 补齐 Web3 后端特色能力。

重点内容：

```text
go-ethereum
RPC
ERC20 Transfer 事件
ABI 解析
充值扫描
确认数
交易签名
nonce
gas
receipt
提现广播
交易状态跟踪
断点续扫
事件去重
```

充值状态：

```text
SCANNED
CONFIRMED
CREDITED
```

提现状态：

```text
CREATED
APPROVED
BROADCASTED
CONFIRMED
FAILED
```

实现模块：

```text
充值地址绑定
ERC20 Transfer 事件扫描
充值确认数
充值入账
提现申请
提现审核 mock
构造 ERC20 transfer
广播交易
监听 receipt
提现状态更新
```

面试表达目标：

```text
我能用 Go 监听链上 ERC20 Transfer 事件做充值入账，也能构造、签名、广播提现交易，并跟踪链上确认状态。
```

### 第 13 周：Docker Compose 模拟真实环境

目标：

- 把项目从“能跑”变成“接近工作环境”。

本地环境：

```text
PostgreSQL
Redis
NATS / Kafka
Anvil / Ganache
API Service
Matching Service
Settlement Service
Market Data Service
Wallet Scanner
Wallet Worker
```

推荐目录：

```text
cmd/api
cmd/matcher
cmd/settlement
cmd/marketdata
cmd/wallet-scanner
cmd/wallet-worker

internal/account
internal/order
internal/matching
internal/marketdata
internal/wallet
internal/db
internal/event
internal/config
```

完整流程：

```text
1. 创建用户
2. mock 充值 USDT / ETH
3. 用户 A 挂卖单
4. 用户 B 挂买单
5. 撮合成交
6. 订单状态变化
7. 买卖双方余额变化
8. 成交记录生成
9. WebSocket 推送成交
10. 提现申请
11. 链上广播交易
12. receipt 确认后更新提现状态
```

面试表达目标：

```text
我用 Docker Compose 搭建了本地 CEX 后端模拟环境，可以完整跑通用户、账户、订单、撮合、结算、行情、充值和提现流程。
```

### 第 14 周：故障演练与可靠性

目标：

- 从 happy path 升级到真实后端思维。

需要模拟的问题：

```text
服务重启后 scanner 能否从上次区块继续扫？
重复收到同一个链上事件会不会重复入账？
结算服务重复消费同一个 trade 会不会重复加钱？
提现广播失败怎么办？
交易 pending 太久怎么办？
数据库事务中途失败怎么办？
Redis / MQ 暂时不可用怎么办？
撮合服务重启后订单簿如何恢复？
```

建议实现：

```text
幂等表
唯一索引
last_scanned_block
event_log 表
trade_settlement 表
withdraw 状态机
重试次数
错误日志
```

面试表达目标：

```text
我考虑过重复消费、服务重启、链上事件重复、交易 pending、RPC 失败等真实问题，并在项目里做了幂等和状态机处理。
```

### 第 15 周：测试、压测与性能观察

目标：

- 让项目更像工程作品，而不是功能 demo。

测试类型：

```text
单元测试
集成测试
API 测试
压测
```

重点测试：

```text
订单簿排序
限价单主动撮合
市价单跨价格档成交
余额冻结
撤单释放冻结
成交结算
重复 trade 不重复结算
重复充值 event 不重复入账
提现状态机
```

压测内容：

```text
下单接口 QPS
撮合吞吐
WebSocket 推送
充值扫描速度
```

工具：

```text
go test
testify
k6
hey
pprof
```

面试表达目标：

```text
我不仅实现了功能，还做了订单簿单测、账本幂等测试、API 集成测试和基础压测。
```

### 第 16 周：简历项目与面试材料

目标：

- 把学习成果包装成可面试项目。

README 必须包含：

```text
项目目标
系统架构图
模块说明
核心业务流程
数据库表设计
订单生命周期
账本模型
撮合规则
充值提现流程
本地启动方式
测试方式
压测结果
已知不足
下一步优化
```

准备 5 张图：

```text
1. 系统架构图
2. 下单到撮合流程图
3. 成交到账本结算流程图
4. 充值扫描流程图
5. 提现状态机图
```

简历描述示例：

```text
基于 Go 实现 Mini CEX 后端系统，包含账户资产、订单管理、撮合引擎、成交结算、行情推送、ERC20 充值扫描和提现广播。使用 PostgreSQL 维护账户与订单状态，Redis/NATS 进行事件分发，go-ethereum 完成链上事件解析和交易发送，支持 Docker Compose 本地一键启动，并针对撮合、账本幂等和充值入账编写测试。
```

## 5. 如何模拟真实 CEX 环境

真实 CEX 后端不是一个单体函数，而是一组服务协作。

你可以用本地 Docker Compose 模拟：

```text
api-service
  -> 接收用户请求

order-service
  -> 参数校验、余额冻结、订单落库

matching-service
  -> 维护订单簿、执行撮合

settlement-service
  -> 消费成交事件、更新账本

marketdata-service
  -> 维护行情快照、WebSocket 推送

wallet-scanner
  -> 扫描链上充值事件

wallet-worker
  -> 处理提现广播和确认

PostgreSQL
  -> 持久化用户、订单、账本、充值提现

Redis
  -> 缓存、限流、Pub/Sub

NATS / Kafka
  -> 事件流

Anvil / Ganache
  -> 本地链环境
```

最小真实链路：

```text
用户下单
  -> API 校验
  -> 账户服务冻结资金
  -> 订单写入数据库
  -> 订单命令进入撮合服务
  -> 撮合服务生成 trade
  -> settlement 服务更新余额和流水
  -> marketdata 服务推送行情
```

充值链路：

```text
链上 ERC20 Transfer
  -> wallet-scanner 扫描事件
  -> event_log 去重
  -> 满足确认数
  -> 入账账户余额
  -> 写充值记录和账户流水
```

提现链路：

```text
用户申请提现
  -> 冻结提现资产
  -> 审核通过
  -> wallet-worker 构造交易
  -> 签名并广播
  -> 监听 receipt
  -> 确认后扣减冻结资产
  -> 写提现流水
```

## 6. 面试必须能讲清楚的问题

账户与资产：

```text
CEX 为什么要区分 available_balance 和 frozen_balance？
买单冻结什么资产？
卖单冻结什么资产？
撤单时如何释放冻结资产？
为什么余额变化必须写流水？
如何防止重复入账？
```

订单与撮合：

```text
限价单什么时候挂单，什么时候立即成交？
市价单如何跨价格档成交？
什么是价格优先、时间优先？
部分成交后订单状态如何变化？
订单簿重启后如何恢复？
为什么撮合和结算要拆开？
```

账本与结算：

```text
成交后买卖双方余额如何变化？
如何保证 trade 不重复结算？
结算失败如何重试？
为什么不能用 float64 处理金额？
数据库事务应该包住哪些操作？
```

链上交互：

```text
ERC20 充值如何扫描？
什么是确认数？
如何处理链上事件重复？
提现交易 pending 怎么办？
nonce 冲突怎么办？
RPC 节点挂了怎么办？
```

工程能力：

```text
如何做幂等？
如何设计状态机？
如何做 API 错误码？
如何做压测？
如何做服务重启恢复？
如何用 Docker Compose 启动整套环境？
```

## 7. 最小可面试版本

如果时间紧，先做这个版本：

```text
go-zero API 服务
go-zero zrpc 内部服务
PostgreSQL 用户、账户、订单、账本
可用余额 / 冻结余额
限价单下单和撤单
简单撮合引擎
成交记录
成交后账本结算
ERC20 Transfer 充值扫描
提现申请和 mock 广播
Redis 限流和行情缓存
NATS 成交事件和结算事件
WebSocket 行情推送
Docker Compose 一键启动
核心测试
README 架构说明
```

这个版本已经足够支撑 Web3 Go 后端初中级岗位面试。

最小技术栈对应关系：

```text
go-zero API：
  对外接口，下单、查订单、查账户、查充值提现。

go-zero zrpc：
  account-rpc、order-rpc、wallet-rpc，模拟真实服务拆分。

PostgreSQL + sqlc/pgx：
  核心业务数据、事务、行锁、幂等、账本流水。

Redis：
  限流、缓存交易对配置、缓存行情快照。

NATS：
  OrderCreated、TradeCreated、SettlementCompleted 等事件。

go-ethereum：
  ERC20 Transfer 充值扫描、提现交易构造和 receipt 查询。

Anvil / Ganache：
  本地链，模拟充值和提现。

Docker Compose：
  一键启动 API、RPC、PostgreSQL、Redis、NATS、本地链。
```

## 8. 不要优先学的内容

当前阶段不要被这些分散：

```text
Kubernetes 深度运维
微服务治理全家桶
高频撮合极致优化
复杂 Solidity 合约
MEV
跨链桥
期权
永续合约爆仓系统
零知识证明
```

它们都重要，但不是你从前端转 Go CEX 后端的第一阶段重点。

你的第一阶段基本盘是：

```text
Go 服务
数据库事务
账户余额
订单状态
资金冻结
撮合逻辑
成交结算
链上充值提现
幂等和重试
```

## 9. 每日学习节奏

如果每天有 2 到 3 小时：

```text
60 分钟：学习一个主题
90 分钟：写项目代码
30 分钟：写笔记 / README / 流程图
```

每周必须产出：

```text
一个可运行功能
一组测试
一段 README 说明
一个流程图或状态机图
```

不要只看教程。你需要把项目做出来，因为你是从前端转后端，面试官会更关心你是否真的写过后端业务链路。

## 10. 当前仓库怎么用

当前仓库适合用来学习：

```text
订单簿
限价单
市价单
撮合
做市机器人
链上转账演示
```

建议使用方式：

```text
第一步：读懂 orderbook/orderbook.go
第二步：读懂 server/server.go 如何调用订单簿
第三步：读懂 mm/maker.go 如何提供流动性
第四步：为订单簿补充测试
第五步：实现限价单主动撮合
第六步：把学到的撮合思想迁移到 mini-cex-backend
```

建议不要直接把当前仓库改成完整 CEX。更好的方式是：

```text
当前仓库：
  用作撮合学习和改造实验

新仓库 mini-cex-backend：
  用作完整 CEX 后端作品集
```

## 11. 你的优势

你的 8 年前端经验不是负担，而是优势。

你天然更容易理解：

```text
交易页面需要什么 API
订单状态如何展示
充值提现流程用户如何感知
WebSocket 行情如何驱动 UI
前端钱包交互怎么发生
用户体验和后端状态之间的关系
```

你需要补的是：

```text
后端工程化
数据库事务
资产账本
消息队列
链上事件
撮合业务
幂等和可靠性
```

把前端经验和 Go CEX 后端能力组合起来，你的定位可以是：

```text
懂交易产品体验的 Go Web3 后端工程师
```

这个标签比“刚学 Go 的后端新人”更有竞争力。

## 12. 最终目标

最终你要能做到：

```text
能写 Go 后端 API
能设计账户和账本表
能处理余额冻结和释放
能实现基础订单簿和撮合
能处理成交后的账本结算
能做行情 WebSocket 推送
能监听 ERC20 充值事件
能广播提现交易
能用 Docker Compose 启动完整环境
能写测试证明核心链路可靠
能在面试中讲清楚 CEX 后端核心流程
```

一句话路线：

```text
用 Go 做出一个能解释账户、订单、撮合、账本、充值、提现、行情的 Mini CEX 后端系统。
```

这比泛泛地说“我学过 Web3 后端”更有说服力。

# Crypto Exchange 学习链路

这份学习计划的目标，是通过当前这个 Go 项目，系统理解一个简化中心化交易所的核心组成：订单簿、撮合引擎、HTTP API、做市逻辑，以及成交后的链上模拟结算。

这个仓库更适合作为“交易所后端核心原理”的学习材料，而不是生产级交易系统模板。学习时要重点关注思想和数据流，不要把里面的安全、资金、并发和精度处理直接当成真实交易所实现。

## 0. 先建立整体地图

先从这些文件建立全局认识：

- `README.md`：了解项目目标。
- `main.go`：看程序如何启动，以及如何模拟交易流。
- `orderbook/orderbook.go`：项目核心，订单薄和撮合。
- `server/server.go`：HTTP API 及内部 PostgreSQL 记账。
- `db/` 和 `cmd/`：数据库连接、迁移、Mock 流水线。
- `client/client.go`：调用 HTTP API 的封装。
- `mm/maker.go`：自动做市脚本。

建议先画出这条主链路：

```text
docker compose up (启动 PG 数据库)
  -> 运行迁移及 seed
  -> main.go 启动 server
  -> server 连接 DB 并常驻内存对账本镜像
  -> market maker & 散户持续发 request
  -> orderbook 撮合出结果 -> 写入 PG 结案
```

学完这一阶段，你应该能回答：

- 为什么 CEX 速度快？订单簿、撮合引擎和传统金融数据库的关系是？

## 1. 订单簿数据结构

重点文件：`orderbook/orderbook.go`

先读这些类型：

- `Order`：一笔订单，包含方向、数量、用户、时间戳、所属价格档。
- `Limit`：一个价格档位，里面放同一价格的多笔订单。
- `Orderbook`：维护买单簿 `bids`、卖单簿 `asks`、所有订单索引和成交记录。
- `Match`：一次撮合结果。
- `Trade`：对外记录的成交数据。

重点理解：

- 买单 `bid` 和卖单 `ask` 为什么要分开存。
- `Limit` 为什么代表一个价格档，而不是一笔订单。
- 同一价格档里的订单为什么要按时间优先。
- 买单簿为什么按价格从高到低排。
- 卖单簿为什么按价格从低到高排。

建议阅读顺序：

1. `NewOrder`
2. `NewLimit`
3. `Limit.AddOrder`
4. `Limit.DeleteOrder`
5. `Orderbook.NewOrderbook`
6. `Orderbook.Asks`
7. `Orderbook.Bids`

练习：

- 手动画一个订单簿，放入三个卖单价格 `1000`、`1010`、`990`，排序后最佳卖价是多少？
- 手动画一个订单簿，放入三个买单价格 `980`、`1000`、`970`，排序后最佳买价是多少？
- 思考：如果用 `float64` 表示价格，在真实金融系统里会有什么风险？

## 2. 限价单逻辑

重点函数：`Orderbook.PlaceLimitOrder`

限价单的意思是：用户指定价格，订单不会立刻无条件成交，而是挂到订单簿上等待别人来吃。

阅读时重点看：

- 如果对应价格档不存在，如何创建新的 `Limit`。
- 买单进入 `BidLimits` 和 `bids`。
- 卖单进入 `AskLimits` 和 `asks`。
- 订单如何进入全局索引 `Orders`。
- 订单如何挂到价格档 `limit.AddOrder(o)`。

你应该能说清楚：

```text
PlaceLimitOrder(price, order)
  -> 找 price 对应的 Limit
  -> 没有就创建
  -> 挂到 bid 或 ask 一侧
  -> 写入订单索引
  -> 加入价格档订单队列
```

练习：

- 增加一个测试：同一个价格连续挂两笔卖单，检查它们是否进入同一个 `Limit`。
- 增加一个测试：挂不同价格的买单，检查 `Bids()` 返回的第一档是否是最高价格。

## 3. 市价单和撮合逻辑

重点函数：

- `Orderbook.PlaceMarketOrder`
- `Limit.Fill`
- `Limit.fillOrder`

市价单的意思是：不指定价格，直接按当前订单簿中最优价格尽快成交。

阅读买入市价单时，可以按这条线理解：

```text
买入 market order
  -> 检查 ask 总量是否足够
  -> 从最低卖价开始吃
  -> 一个价格档吃完后继续下一个价格档
  -> 直到市价单数量填满
  -> 生成 matches
  -> 生成 trades
```

阅读卖出市价单时，可以按这条线理解：

```text
卖出 market order
  -> 检查 bid 总量是否足够
  -> 从最高买价开始吃
  -> 一个价格档吃完后继续下一个价格档
  -> 直到市价单数量填满
  -> 生成 matches
  -> 生成 trades
```

重点理解：

- 为什么买入市价单吃的是 asks。
- 为什么卖出市价单吃的是 bids。
- 部分成交和完全成交分别怎么修改订单数量。
- 一个市价单为什么可能生成多个 `Match`。
- 一个价格档被吃空后为什么要 `clearLimit`。

练习：

- 手写一个场景：卖单 `1000@5`、`1010@8`，买入市价单数量为 `10`，最终会产生几笔 match？
- 增加一个测试：市价单跨两个价格档成交。
- 增加一个测试：市价单刚好吃空某个价格档后，该价格档是否从订单簿移除。

## 4. 成交记录和订单状态

重点字段：

- `Order.Size`
- `Order.IsFilled`
- `Orderbook.Trades`

这个项目中，成交会直接修改订单剩余数量：

```text
订单原始 Size = 10
成交 Size = 4
订单剩余 Size = 6
```

当 `Size == 0` 时，订单被认为已经完全成交。

你需要注意：

- `Match` 是一次撮合结果。
- `Trade` 是记录给外部查询的成交历史。
- `Orderbook.Trades` 是内存数组，没有数据库持久化。
- 被完全成交的限价单会从价格档中删除。

练习：

- 查看 `TestLastMarketTrades`，理解 match 和 trade 的关系。
- 增加测试：一笔限价单被部分成交后，订单簿剩余数量是否正确。

## 5. HTTP API 层

重点文件：`server/server.go`

API 层负责把外部请求转换成订单簿操作。

核心入口：

- `StartServer`
- `handlePlaceOrder`
- `handleGetBook`
- `handleGetBestBid`
- `handleGetBestAsk`
- `handleGetTrades`
- `cancelOrder`

重点链路：

```text
POST /order
  -> handlePlaceOrder
  -> 解析 PlaceOrderRequest
  -> 创建 orderbook.Order
  -> 如果是 LIMIT，调用 handlePlaceLimitOrder
  -> 如果是 MARKET，调用 handlePlaceMarketOrder
  -> 返回 OrderID
```

建议用 curl 或 Postman 手动调用：

```bash
curl -X POST http://localhost:3000/order \
  -H "Content-Type: application/json" \
  -d '{"UserID":8,"Type":"LIMIT","Bid":true,"Size":10,"Price":960,"Market":"ETH"}'
```

```bash
curl http://localhost:3000/book/ETH
```

学完这一阶段，你应该能回答：

- server 层是否真正做撮合？
- `/order` 这个接口如何区分市价单和限价单？
- API 返回的数据和订单簿内部结构有什么差别？

## 6. Client 封装

重点文件：`client/client.go`

这个模块只是把 HTTP 请求封装成 Go 函数，方便 `main.go` 和 `mm/maker.go` 调用。

重点函数：

- `PlaceLimitOrder`
- `PlaceMarketOrder`
- `GetBestBid`
- `GetBestAsk`
- `GetOrders`
- `GetTrades`
- `CancelOrder`

你可以把它理解成：

```text
业务代码不直接写 HTTP 请求
而是调用 client.PlaceLimitOrder(...)
client 内部负责组装 JSON 和请求路径
```

练习：

- 给 `client` 增加一个 `GetBook(market string)` 方法。
- 给所有 HTTP response 增加 `defer resp.Body.Close()`。
- 如果响应状态码不是 2xx，返回错误。

## 7. 做市机器人

重点文件：`mm/maker.go`

Market maker 的作用是给订单簿提供流动性。没有流动性时，市价单无法成交。

核心逻辑：

```text
makerLoop
  -> 查询 best bid 和 best ask
  -> 如果订单簿为空，seedMarket
  -> 如果 spread 足够大，在中间继续挂买卖单
  -> 周期性重复
```

重点理解：

- `seedMarket` 为什么要在当前价格上下各挂一笔。
- `MinSpread` 控制什么。
- `PriceOffset` 控制什么。
- 做市机器人为什么同时挂买单和卖单。

练习：

- 修改 `OrderSize`，观察成交频率和订单簿变化。
- 修改 `MinSpread`，观察 maker 是否还继续挂单。
- 把 `simulateFetchCurrentETHPrice` 改成随机价格，观察 seed 行为。

## 8. 内部账本与 PostgreSQL 持久化结算

重点文件：`server/server.go` 与 `db/` 目录

重点函数：

- `handleMatches` (内存账户扣血加血并原子入库)
- GORM 模型: `UserModel`, `TradeModel`
- `db.MigrateUp`

这部分是成交之后的清结算。真实 CEX 的资产交割是在自己服务器内部完成的，而不是上链。我们使用了 PostgreSQL + GORM 来模拟这套极速记账机制。

当前核心结算流程：

```text
订单撮合成功
  -> 得到 matches，遍历每一个 match 寻找买卖双边
  -> 内存账户操作：卖方余额增加，买方余额扣除
  -> 开启数据库事务 (DB.Transaction)
  -> DB UPDATE user 账本记录落盘
  -> DB INSERT trade 成交记录落盘
```

学习重点：

- `gorm.Open` 建立连接以及配置 Logger。
- `golang-migrate` 与 `go:embed` 搭配：将由 SQL 组成的 Schema 版本化构建进可执行二进制文件内。
- 事务 (`tx *gorm.DB`) 回滚机制：确保买卖双边以及成交记录对齐进出一致。
- 内存镜像：为什么 `ex.Users` 在内存里先扣一笔？以空间换时间。

重要提醒：

- 这极大提高了响应速度并降低 Gas 费用，但这也带来了资金全权受项目方服务器控制的问题。
- 真实的交易所会引入极其严格的风控引擎以及多级资金对账，不是简单的余额加减。

练习：

- 使用 Docker 运行环境内的 Postgres ，并使用 SQL 查询观察 `trades` 与 `users` 表的变化。
- 理解项目中 `Makefile` (`make build`, `make dev`) 和 `cmd/migrate` CLI 的作用。
- 思考：如果在向 `users` 发起 DB update 扣取余额的过程中进程崩溃了，但是上层撮合引擎已经被消化了，该如何恢复数据一致性？


## 9. 从 CEX 到 DEX 的概念对照

这个项目更像中心化交易所的撮合核心。

CEX 风格：

```text
用户下单
  -> 中心化服务器接收订单
  -> 中心化订单簿维护挂单
  -> 中心化撮合引擎生成成交
  -> 内部账本或链上系统结算
```

DEX 常见风格：

```text
用户连接钱包
  -> 与智能合约交互
  -> AMM 池子或链上订单簿完成交换
  -> 状态直接记录在链上
```

通过这个项目，你能优先学到：

- CEX 订单簿模型。
- CEX 撮合模型。
- maker/taker 的基础概念。
- API 如何包裹撮合引擎。
- 链上结算和撮合之间的区别。

但这个项目不能直接教完整 DEX，因为它没有：

- 智能合约交易池。
- AMM 定价公式。
- 钱包签名下单。
- 链上流动性池。
- 链上订单状态。

## 10. 推荐学习顺序

建议按 8 天学习。

### Day 1：跑通项目结构

目标：

- 看懂目录结构。
- 看懂 `main.go` 启动流程。
- 知道 server、client、orderbook、mm 的关系。

产出：

- 画一张项目调用链路图。
- 写下你理解的“一笔订单从提交到成交”的过程。

### Day 2：订单簿基础

目标：

- 看懂 `Order`、`Limit`、`Orderbook`。
- 知道 bids 和 asks 的排序规则。

产出：

- 手动画一个订单簿。
- 写两个小测试验证 bids/asks 排序。

### Day 3：限价单

目标：

- 看懂 `PlaceLimitOrder`。
- 理解价格档的创建和复用。

产出：

- 写测试验证同价格订单进入同一个价格档。
- 写测试验证订单索引 `Orders` 正确记录订单。

### Day 4：市价单和撮合

目标：

- 看懂 `PlaceMarketOrder`、`Fill`、`fillOrder`。
- 理解完全成交、部分成交、跨价格档成交。

产出：

- 写一个跨两个价格档成交的测试。
- 手动推演成交前后每笔订单剩余数量。

### Day 5：API 层

目标：

- 看懂 `handlePlaceOrder`。
- 知道 HTTP 请求如何转换成订单簿操作。

产出：

- 用 curl 提交限价单和市价单。
- 查询 `/book/ETH` 和 `/trades/ETH`。

### Day 6：做市机器人

目标：

- 看懂 `makerLoop` 和 `seedMarket`。
- 理解 spread、best bid、best ask。

产出：

- 修改做市参数，观察订单簿变化。
- 写一段笔记解释 maker 为什么提供流动性。

### Day 7：数据库存储与结算

目标：

- 看懂 `handleMatches` 中的 GORM 事务运作
- 看懂 `db/migrate.go` 以及 `embed` 的原生静态绑定用法
- 理解核心撮合引擎（内存动作）和 数据库持久化结算 的业务切分

产出：

- 用 docker-compose 后台启动测试数据库。
- 观察终端内做市跑出成交后，用类似 DataGrip 的客户端去查验数据库内 balance 的变化。

### Day 8：重构和扩展

目标：

- 识别这个教学项目和真实系统的差距。
- 尝试做小型改进。

推荐改进：

- 用整数表示价格和数量。
- 给 client 关闭 response body。
- 处理市价单流动性不足时的错误，不要 panic。
- 增加 market 参数，不只支持 ETH。
- 增加更完整的测试。
- 把订单和成交记录持久化到数据库。

## 11. 深挖问题清单

学完基础后，可以继续追问这些问题：

- 如果两个用户同时下单，当前代码是否线程安全？
- 限价单是否应该主动撮合已经存在的对手单？
- 市价单没有足够流动性时应该全部拒绝，还是部分成交？
- 订单取消时，如何避免取消已经成交的订单？
- 真实交易所如何处理用户余额冻结？
- 如果扣取用户余额时 DB 写失败怎么办？
- 撮合引擎崩盘，内存索引丢失可以靠 DB 还原吗？
- 为什么真实系统通常不用 `float64` 处理金额？
- 订单 ID 应该如何生成才可靠？
- 成交记录为什么需要持久化和审计？

## 12. 最终学习目标

如果你完整走完这条链路，应该能够独立解释：

- 什么是订单簿。
- 什么是限价单和市价单。
- 什么是 best bid 和 best ask。
- 什么是 spread。
- 市价单如何吃掉限价单。
- 一笔订单如何部分成交。
- 一次市价单为什么可能产生多笔成交。
- 做市机器人为什么要同时挂买单和卖单。
- API 层和撮合引擎之间如何分工。
- CEX 撮合核心和 DEX 链上交换有什么区别。
- 这个项目离真实交易所还缺哪些关键能力。

最终你可以尝试自己实现一个最小版本：

```text
1. 支持挂买单和卖单
2. 支持市价单撮合
3. 支持查询订单簿
4. 支持查询成交记录
5. 支持取消订单
6. 支持一个简单做市机器人
```

做到这里，你就不是只“看过交易所代码”，而是真的摸到交易所撮合系统的骨架了。

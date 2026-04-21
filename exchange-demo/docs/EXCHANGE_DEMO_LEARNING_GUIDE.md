# exchange-demo 学习指南与链路分析

这份文档面向想学习 `exchange-demo` 的同学：先看整体结构，再沿着“启动链路 -> 下单链路 -> 撮合链路 -> 结算链路”把 demo 跑通和读懂。

> 说明：`exchange-demo` 内已有 `docs/architecture.md` 和 `docs/LEARNING_PLAN.md`，但当前文件内容存在明显编码乱码。本文按实际 Go 代码重新整理。

## 1. Demo 定位

`exchange-demo` 是一个极简中心化交易所后端 demo，核心目标是演示：

- 订单簿模型：`bid / ask`、价格档位、限价单、市价单。
- 撮合逻辑：按最优价格档位消耗对手盘，生成 `Match` 和 `Trade`。
- HTTP API：用 Echo 暴露下单、查询订单簿、查询成交、撤单等接口。
- 模拟做市：`mm` 模块持续放置买卖限价单，给订单簿提供流动性。
- 模拟交易者：`main.go` 中的 `marketOrderPlacer` 持续发市价单。
- 简单结算：撮合完成后更新内存用户余额，并把余额和成交记录写入 PostgreSQL。

它更适合作为“理解 CEX 基础链路”的教学项目，不适合直接对标生产级交易所。

## 2. 目录结构

```text
exchange-demo/
  main.go                 # 程序入口：启动 server、market maker、随机市价单
  client/                 # Go HTTP Client SDK，封装调用 server API
  server/                 # Echo HTTP 服务、Exchange 聚合对象、API handler、结算逻辑
  orderbook/              # 订单簿核心：Order、Limit、Orderbook、Match、Trade
  mm/                     # Market Maker：查询盘口并持续挂买卖限价单
  db/                     # 数据库连接、迁移执行、嵌入式 SQL migrations
  db/migrations/          # users / trades 表结构
  cmd/migrate/            # 手动执行 migration 的 CLI
  cmd/seed/               # 批量创建模拟用户 100-199
  docker-compose.yml      # PostgreSQL 13
  README.md               # 原项目说明，内容较少
```

建议按这条顺序阅读：

1. `orderbook/orderbook.go`
2. `server/server.go`
3. `client/client.go`
4. `mm/maker.go`
5. `main.go`
6. `db/migrate.go` 与 `db/migrations/*.sql`
7. `cmd/seed/main.go` 与 `cmd/migrate/main.go`

## 3. 核心对象关系

```mermaid
classDiagram
    class Order {
      int64 ID
      int64 UserID
      float64 Size
      bool Bid
      *Limit Limit
      int64 Timestamp
    }

    class Limit {
      float64 Price
      Orders Orders
      float64 TotalVolume
      AddOrder()
      DeleteOrder()
      Fill()
    }

    class Orderbook {
      []*Limit asks
      []*Limit bids
      map AskLimits
      map BidLimits
      map Orders
      []*Trade Trades
      PlaceLimitOrder()
      PlaceMarketOrder()
      CancelOrder()
      Asks()
      Bids()
    }

    class Exchange {
      *gorm.DB DB
      map Users
      map Orders
      map orderbooks
      handlePlaceOrder()
      handleMatches()
    }

    Orderbook "1" o-- "*" Limit
    Limit "1" o-- "*" Order
    Exchange "1" o-- "*" Orderbook
```

关键理解点：

- `Order` 是单个订单，`Bid=true` 表示买单，`Bid=false` 表示卖单。
- `Limit` 是同一价格档位，里面挂着多个订单。
- `Orderbook` 同时维护切片和 map：切片用于排序和遍历最优价格，map 用于快速定位价格档位或订单。
- `Exchange` 是服务层聚合对象：持有 DB、用户余额、用户订单索引和不同市场的订单簿。

## 4. 启动链路

`main.go` 是 demo 的总控入口：

```text
main()
  -> go server.StartServer()
  -> sleep 1s
  -> client.NewClient()
  -> mm.NewMakerMaker(cfg)
  -> maker.Start()
  -> sleep 2s
  -> go marketOrderPlacer(client)
  -> select{} 阻塞主 goroutine
```

`server.StartServer()` 内部会：

```text
StartServer()
  -> 读取 DATABASE_URL，缺省使用本地 PostgreSQL
  -> db.MigrateUp(dsn)
  -> db.Connect(dsn)
  -> NewExchange(gormDB)
  -> registerUser(8, 10000), registerUser(7, 10000), registerUser(99, 10000)
  -> loadUsersFromDB()
  -> 注册 HTTP 路由
  -> e.Start(":3000")
```

`market maker` 启动后先检查订单簿是否为空。如果没有 best bid / best ask，就用模拟 ETH 价格 `1000` 加减 `SeedOffset` 建立初始盘口。

## 5. HTTP API 链路

主要 API：

```text
POST   /order              下限价单或市价单
GET    /trades/:market     查询最近成交
GET    /order/:userID      查询用户当前挂单
GET    /book/:market       查询订单簿
GET    /book/:market/bid   查询 best bid
GET    /book/:market/ask   查询 best ask
DELETE /order/:id          撤单
GET    /user/:userID/balance 查询余额
```

统一入口是 `server.Exchange.handlePlaceOrder`：

```text
POST /order
  -> JSON decode 为 PlaceOrderRequest
  -> orderbook.NewOrder(req.Bid, req.Size, req.UserID)
  -> Type == LIMIT  -> handlePlaceLimitOrder()
  -> Type == MARKET -> handlePlaceMarketOrder() -> handleMatches()
  -> 返回 OrderID
```

`client/client.go` 相当于一个简易 SDK，把 Go 方法转成 HTTP 调用：

```text
mm / main
  -> client.PlaceLimitOrder() 或 client.PlaceMarketOrder()
  -> HTTP POST /order
  -> server handler
  -> orderbook
```

建议学习时先用 `curl` 直接请求 API，再回到 `client` 看它如何封装。

## 6. 限价单链路

限价单是“挂到订单簿上等待成交”的订单。

```text
client.PlaceLimitOrder()
  -> POST /order Type=LIMIT
  -> Exchange.handlePlaceOrder()
  -> Exchange.handlePlaceLimitOrder()
  -> Orderbook.PlaceLimitOrder(price, order)
  -> 找到或创建 price 对应的 Limit
  -> 加入 bids 或 asks
  -> 写入 Orderbook.Orders
  -> limit.AddOrder(order)
  -> 写入 Exchange.Orders[userID]
```

`Orderbook.PlaceLimitOrder` 的核心是价格档位管理：

- 买单进入 `BidLimits` 和 `bids`。
- 卖单进入 `AskLimits` 和 `asks`。
- `Orders` map 用订单 ID 快速索引订单，便于撤单。
- `Limit.AddOrder` 会设置 `order.Limit = limit`，并累加档位总量。

这里没有实现“可成交限价单立即撮合”的逻辑。也就是说，即使一个买入限价价高于 best ask，当前代码仍会直接挂到订单簿，而不是立刻成交。这是教学 demo 的简化点。

## 7. 市价单与撮合链路

市价单是“立即吃掉对手盘”的订单。

```text
client.PlaceMarketOrder()
  -> POST /order Type=MARKET
  -> Exchange.handlePlaceOrder()
  -> Exchange.handlePlaceMarketOrder()
  -> Orderbook.PlaceMarketOrder(order)
  -> Limit.Fill(order)
  -> Limit.fillOrder(restingOrder, marketOrder)
  -> 返回 []Match
  -> Exchange.handleMatches()
  -> DB transaction 更新余额并插入成交
```

买入市价单会吃 `asks`：

```text
market buy
  -> 检查 AskTotalVolume 是否足够
  -> 从 best ask 开始遍历
  -> 逐个 Limit.Fill()
  -> 被吃空的价格档位 clearLimit(false, limit)
  -> 生成 Trade
```

卖出市价单会吃 `bids`：

```text
market sell
  -> 检查 BidTotalVolume 是否足够
  -> 从 best bid 开始遍历
  -> 逐个 Limit.Fill()
  -> 被吃空的价格档位 clearLimit(true, limit)
  -> 生成 Trade
```

`Limit.fillOrder` 会修改双方订单的 `Size`：

```text
如果 restingOrder.Size >= marketOrder.Size:
  restingOrder.Size -= marketOrder.Size
  marketOrder.Size = 0

否则:
  marketOrder.Size -= restingOrder.Size
  restingOrder.Size = 0
```

撮合产物有两个层次：

- `Match`：一次订单与订单之间的撮合结果，含买方订单、卖方订单、成交数量和价格。
- `Trade`：订单簿内保存的成交记录，记录价格、数量、方向、时间。

## 8. 做市链路

`mm/maker.go` 的 `makerLoop` 周期性运行：

```text
makerLoop()
  -> GetBestBid()
  -> GetBestAsk()
  -> 如果盘口为空：seedMarket()
  -> 如果单边为空：用另一边价格补出一个参考价格
  -> spread = bestAsk.Price - bestBid.Price
  -> spread <= MinSpread 则跳过
  -> placeOrder(true,  bestBid + PriceOffset)
  -> placeOrder(false, bestAsk - PriceOffset)
```

`seedMarket()` 用模拟价格 `1000` 建立初始盘口：

```text
bid price = currentPrice - SeedOffset
ask price = currentPrice + SeedOffset
```

按 `main.go` 当前配置，初始盘口大致是：

```text
bid: 960
ask: 1040
spread: 80
```

然后做市器会在 spread 足够大时把买价往上贴、卖价往下贴。

## 9. 随机市价单链路

`main.go` 的 `marketOrderPlacer` 每 500ms 发一个模拟用户市价单：

```text
marketOrderPlacer()
  -> 随机 Bid 方向，当前逻辑约 30% 买入、70% 卖出
  -> 随机 UserID: 100-199
  -> Size = 1
  -> client.PlaceMarketOrder()
```

这些用户需要先由 `cmd/seed` 写入数据库，或者由其他逻辑注册。否则市价单成交后 `handleMatches` 可能找不到用户。

## 10. 数据库与结算链路

数据库表：

- `users`：`user_id`、`balance`，用于模拟用户余额。
- `trades`：`market`、`price`、`size`、`bid_user_id`、`ask_user_id`，用于保存成交记录。

迁移执行链路：

```text
server.StartServer()
  -> db.MigrateUp(dsn)
  -> go:embed db/migrations/*.sql
  -> golang-migrate 执行 users / trades 建表
```

成交后的结算链路：

```text
Exchange.handleMatches(market, matches)
  -> DB.Transaction()
  -> 找到 askUser 和 bidUser
  -> askUser.Balance += SizeFilled
  -> bidUser.Balance -= SizeFilled
  -> UPDATE users SET balance = ...
  -> INSERT trades (...)
```

注意：当前余额语义非常简化，代码里用成交数量直接加减 `balance`，并没有按 `price * size` 结算，也没有区分资产账户与现金账户。这一点适合学习链路，但不是正确的交易所资产模型。

## 11. 如何运行

推荐步骤：

```powershell
cd C:\Users\33002\Desktop\min-cex\exchange-demo
docker compose up -d
go run ./cmd/migrate -action=up
go run ./cmd/seed
go run .
```

如果没有设置环境变量，程序默认使用：

```text
postgres://cex:cex_secret@localhost:5432/cex_db?sslmode=disable&TimeZone=Asia/Shanghai
```

可以用这些请求观察系统：

```bash
curl http://localhost:3000/book/ETH
curl http://localhost:3000/book/ETH/bid
curl http://localhost:3000/book/ETH/ask
curl http://localhost:3000/trades/ETH
curl http://localhost:3000/user/8/balance
```

手动挂单示例：

```bash
curl -X POST http://localhost:3000/order \
  -H "Content-Type: application/json" \
  -d '{"UserID":8,"Type":"LIMIT","Bid":true,"Size":10,"Price":960,"Market":"ETH"}'
```

## 12. 当前已知问题

学习时建议带着这些问题读代码，它们正好是很好的练习题：

- 源码和已有文档中的中文注释有编码乱码，建议后续统一转成 UTF-8 后重写注释。
- `orderbook/orderbook_test.go` 仍按旧签名调用 `PlaceMarketOrder`，但当前函数返回 `([]Match, error)`，测试需要改成接收两个返回值。
- `client` 没有关闭 `resp.Body`，也没有检查非 2xx 状态码。
- `handlePlaceMarketOrder` 里 `avgPrice := sumPrice / float64(len(matches))`，如果没有 match 会除以 0；虽然 `Orderbook.PlaceMarketOrder` 目前会先检查总量，但服务层最好仍做防御。
- `cancelOrder` 没有检查订单 ID 不存在、订单已成交或 `order.Limit == nil` 的情况。
- `float64` 不适合真实金融金额和价格，生产系统通常使用定点数、整数最小单位或 decimal。
- `Order.ID` 使用 `rand.Intn`，可能碰撞；真实系统需要全局唯一 ID。
- 用户余额和订单簿状态分散在内存和 DB，重启后订单簿不会恢复。
- 当前仅支持 `ETH` 一个市场，`client` 里也写死了 `ETH`。
- 结算模型没有区分 base asset / quote asset，也没有冻结资金、手续费、风控、账户流水。
- 做市器没有撤单策略，会持续堆积挂单。
- 限价单不会主动 crossing match，与真实撮合引擎差异较大。

## 13. 学习建议

第一阶段：只看订单簿

- 自己画出 `Order -> Limit -> Orderbook` 的对象图。
- 手写几个测试：挂两个 ask、一个 market buy，观察 `Size` 如何变化。
- 重点理解 `Asks()` 按低价优先、`Bids()` 按高价优先。

第二阶段：看 HTTP 到撮合

- 从 `POST /order` 开始跟到 `Orderbook.PlaceLimitOrder`。
- 再从 `POST /order Type=MARKET` 跟到 `handleMatches`。
- 用 `curl` 造一组订单，观察 `/book/ETH` 与 `/trades/ETH` 的变化。

第三阶段：看自动化交易行为

- 读 `mm/maker.go`，理解 best bid、best ask、spread。
- 修改 `MinSpread`、`PriceOffset`、`OrderSize`，观察订单簿变化。
- 暂时注释掉 `marketOrderPlacer`，先只观察做市器如何铺盘口。

第四阶段：看数据库一致性

- 跟踪 `handleMatches` 的 transaction。
- 在数据库里查 `users` 和 `trades`。
- 思考哪些状态只在内存里，哪些状态落库了，重启后会丢什么。

第五阶段：尝试修复和增强

- 修复 `orderbook_test.go` 的返回值问题，让测试重新通过。
- 给 `client` 增加 `defer resp.Body.Close()` 和 HTTP 状态码检查。
- 给撤单、空成交、未知市场、未知用户补充错误处理。
- 把 `Order.ID` 改成更可靠的生成方式。
- 增加 `GetBook` client 方法。
- 把限价 crossing 逻辑补上：可成交部分立即吃单，剩余部分再挂单。
- 把结算改成 base / quote 双资产账户模型。

## 14. 一句话总链路

```text
main 启动服务和模拟交易者
  -> client 发 HTTP 请求
  -> server 解析请求并调用 Exchange
  -> Exchange 把订单交给 Orderbook
  -> Orderbook 维护价格档位、执行撮合、生成 matches/trades
  -> Exchange 根据 matches 做简化结算
  -> PostgreSQL 保存用户余额和成交记录
```

把这条链路能从头到尾讲清楚，这个 demo 的学习价值就吃到大半了。

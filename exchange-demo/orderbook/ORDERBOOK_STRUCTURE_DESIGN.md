# orderbook.go 结构设计说明

这份文档专门解释 `exchange-demo/orderbook/orderbook.go`。它不从交易所理论开始，而是从代码里的结构、字段、方法和状态变化讲清楚：一个订单进来后，到底放在哪里，撮合时到底删了什么、改了什么。

## 1. 这个文件解决什么问题

`orderbook.go` 实现的是一个极简订单簿。

它要做三件事：

1. 保存限价单。
2. 用市价单吃掉对手盘。
3. 记录撮合结果。

代码里没有账户、余额、HTTP、数据库，这些都在 `server` 层。这里关心的只有订单簿内存结构。

```text
限价单：挂在订单簿里，等别人来成交。
市价单：不挂单，直接吃掉订单簿里已有的对手单。
```

## 2. 核心结构一眼看懂

订单簿分成买盘和卖盘。

```text
Orderbook
  asks: 卖盘价格档位列表，价格从低到高排序后，asks[0] 是 best ask
  bids: 买盘价格档位列表，价格从高到低排序后，bids[0] 是 best bid

  AskLimits: price -> 卖盘价格档位
  BidLimits: price -> 买盘价格档位
  Orders: orderID -> order
  Trades: 成交记录
```

用一个具体例子看：

```text
卖盘 asks
  1010: ask order A, size 2
  1020: ask order B, size 5

买盘 bids
  990: bid order C, size 3
  980: bid order D, size 4
```

在代码里的形态大概是：

```text
Orderbook{
  asks: [
    Limit{Price: 1010, Orders: [A], TotalVolume: 2},
    Limit{Price: 1020, Orders: [B], TotalVolume: 5},
  ],
  bids: [
    Limit{Price: 990, Orders: [C], TotalVolume: 3},
    Limit{Price: 980, Orders: [D], TotalVolume: 4},
  ],
  AskLimits: {
    1010: 指向 asks 里的 1010 档位,
    1020: 指向 asks 里的 1020 档位,
  },
  BidLimits: {
    990: 指向 bids 里的 990 档位,
    980: 指向 bids 里的 980 档位,
  },
  Orders: {
    A.ID: A,
    B.ID: B,
    C.ID: C,
    D.ID: D,
  },
}
```

注意：`asks` / `bids` 是价格档位列表，不是订单列表。真正的订单在每个 `Limit.Orders` 里面。

## 3. 数据结构拆解

### 3.1 Order：一张订单

```go
type Order struct {
    ID        int64
    UserID    int64
    Size      float64
    Bid       bool
    Limit     *Limit
    Timestamp int64
}
```

字段含义：

```text
ID: 订单 ID，当前用 rand.Intn 随机生成。
UserID: 下单用户。
Size: 剩余未成交数量。不是原始数量。
Bid: true 表示买单，false 表示卖单。
Limit: 如果这是挂在订单簿里的限价单，指向所属价格档位。
Timestamp: 创建时间，用来表达先来后到。
```

`Size` 是最容易误解的字段。它会随着撮合不断减少：

```text
原始订单 size = 10
成交 4
剩余 Size = 6
再成交 6
剩余 Size = 0，订单完成
```

`IsFilled()` 判断的就是 `Size == 0.0`。

### 3.2 Limit：一个价格档位

```go
type Limit struct {
    Price       float64
    Orders      Orders
    TotalVolume float64
}
```

一个 `Limit` 表示同一个价格下的所有订单。

例如卖盘 1010 这个价格有三个卖单：

```text
Limit{
  Price: 1010,
  Orders: [
    ask order A, size 2,
    ask order B, size 3,
    ask order C, size 5,
  ],
  TotalVolume: 10,
}
```

`TotalVolume` 是这个价格档位的总剩余数量。

```text
TotalVolume = A.Size + B.Size + C.Size
```

### 3.3 Orderbook：完整订单簿

```go
type Orderbook struct {
    asks []*Limit
    bids []*Limit
    Trades []*Trade

    mu        sync.RWMutex
    AskLimits map[float64]*Limit
    BidLimits map[float64]*Limit
    Orders    map[int64]*Order
}
```

为什么同时有 slice 和 map？

```text
asks / bids:
  用来排序，找到 best ask / best bid，撮合时按价格顺序遍历。

AskLimits / BidLimits:
  用价格快速找到档位。否则每次挂单都要遍历 asks / bids。

Orders:
  用订单 ID 快速找到订单，撤单时使用。
```

这是一种“读写方便但需要维护一致性”的结构。每次新增、成交、撤单，都要同时维护多个容器。

## 4. 排序规则

代码里有两个排序器：

```go
type ByBestAsk struct{ Limits }
func (a ByBestAsk) Less(i, j int) bool {
    return a.Limits[i].Price < a.Limits[j].Price
}
```

卖盘越便宜越优先：

```text
asks: 100, 101, 102
best ask = asks[0] = 100
```

```go
type ByBestBid struct{ Limits }
func (b ByBestBid) Less(i, j int) bool {
    return b.Limits[i].Price > b.Limits[j].Price
}
```

买盘越贵越优先：

```text
bids: 99, 98, 97
best bid = bids[0] = 99
```

`Asks()` 和 `Bids()` 每次调用都会现场排序：

```go
func (ob *Orderbook) Asks() []*Limit {
    sort.Sort(ByBestAsk{ob.asks})
    return ob.asks
}

func (ob *Orderbook) Bids() []*Limit {
    sort.Sort(ByBestBid{ob.bids})
    return ob.bids
}
```

这意味着 `asks` / `bids` 平时不保证有序，只有调用 `Asks()` / `Bids()` 后才变成按最优价格排列。

## 5. 挂限价单：PlaceLimitOrder

入口：

```go
func (ob *Orderbook) PlaceLimitOrder(price float64, o *Order)
```

它做的事可以拆成五步：

```text
1. 判断订单方向。
2. 根据 price 找已有 Limit。
3. 如果这个价格档位不存在，就创建新 Limit。
4. 把订单放进 Orders map。
5. 把订单放进对应 Limit.Orders。
```

### 5.1 买入限价单

比如用户挂买单：

```text
buy limit: price = 990, size = 3
```

代码逻辑：

```text
o.Bid == true
  -> limit = ob.BidLimits[990]
  -> 如果没有这个价格档位:
       limit = NewLimit(990)
       ob.bids = append(ob.bids, limit)
       ob.BidLimits[990] = limit
  -> ob.Orders[o.ID] = o
  -> limit.AddOrder(o)
```

状态变化：

```text
新增前:
  bids = []
  BidLimits = {}
  Orders = {}

新增后:
  bids = [
    Limit{Price: 990, Orders: [order], TotalVolume: 3}
  ]
  BidLimits[990] = 上面这个 Limit
  Orders[order.ID] = order
  order.Limit = 上面这个 Limit
```

### 5.2 卖出限价单

比如用户挂卖单：

```text
sell limit: price = 1010, size = 2
```

代码逻辑：

```text
o.Bid == false
  -> limit = ob.AskLimits[1010]
  -> 如果没有这个价格档位:
       limit = NewLimit(1010)
       ob.asks = append(ob.asks, limit)
       ob.AskLimits[1010] = limit
  -> ob.Orders[o.ID] = o
  -> limit.AddOrder(o)
```

### 5.3 当前代码的简化点

真实撮合系统中，如果你挂一个买入限价单 `price=1100`，而当前 best ask 是 `1010`，这张单应该立即成交。

但当前代码不会这么做。

```text
当前实现：
  所有限价单都直接进入订单簿。

缺失能力：
  crossing limit order 立即撮合。
```

所以这个 demo 的限价单更像是“纯挂单”，不负责主动撮合。

## 6. 市价单：PlaceMarketOrder

入口：

```go
func (ob *Orderbook) PlaceMarketOrder(o *Order) ([]Match, error)
```

市价单不保存到 `ob.Orders`，它只是来吃对手盘。

```text
市价买单：吃 asks。
市价卖单：吃 bids。
```

### 6.1 市价买单吃 asks

假设当前卖盘：

```text
asks:
  1010: A size 2
  1020: B size 5
  1030: C size 10
```

来了一个市价买单：

```text
market buy: size 6
```

执行顺序：

```text
1. 检查 AskTotalVolume >= 6。
2. ob.Asks() 排序，得到 1010 -> 1020 -> 1030。
3. 先吃 1010 档位。
4. 如果还没吃满，再吃 1020 档位。
5. 生成 Match。
6. 被完全吃空的订单从 Limit.Orders 删除。
7. 被完全吃空的价格档位从 asks 和 AskLimits 删除。
8. 根据 Match 生成 Trade。
```

撮合结果：

```text
吃 A: 2 @ 1010
吃 B: 4 @ 1020

market buy 剩余 size = 0
A 剩余 size = 0，被删除
B 剩余 size = 1，继续留在 1020 档位
C 不动
```

生成的 `matches`：

```text
[
  Match{Ask: A, Bid: marketBuy, SizeFilled: 2, Price: 1010},
  Match{Ask: B, Bid: marketBuy, SizeFilled: 4, Price: 1020},
]
```

### 6.2 市价卖单吃 bids

假设当前买盘：

```text
bids:
  990: A size 3
  980: B size 4
  970: C size 5
```

来了一个市价卖单：

```text
market sell: size 5
```

执行顺序：

```text
1. 检查 BidTotalVolume >= 5。
2. ob.Bids() 排序，得到 990 -> 980 -> 970。
3. 先吃 990 档位。
4. 如果还没吃满，再吃 980 档位。
5. 生成 Match。
6. 清理吃空的订单和价格档位。
7. 生成 Trade。
```

撮合结果：

```text
吃 A: 3 @ 990
吃 B: 2 @ 980

market sell 剩余 size = 0
A 剩余 size = 0，被删除
B 剩余 size = 2，继续留在 980 档位
C 不动
```

## 7. 一个价格档位内部怎么撮合：Limit.Fill

入口：

```go
func (l *Limit) Fill(o *Order) []Match
```

`o` 是主动进来的市价单。`l.Orders` 是这个价格档位上原本挂着的限价单。

逻辑：

```text
for 每一个挂单 order in l.Orders:
  如果市价单 o 已经完全成交:
    break

  match = l.fillOrder(order, o)
  matches append match
  l.TotalVolume -= match.SizeFilled

  如果挂单 order 已经完全成交:
    先记到 ordersToDelete

循环结束后:
  删除所有已经完全成交的挂单
```

为什么不在循环中立刻删除？

因为正在遍历 `l.Orders`，边遍历边改 slice 容易跳过元素或造成混乱。所以它先收集到 `ordersToDelete`，遍历结束后再删。

## 8. 两张订单怎么扣数量：fillOrder

入口：

```go
func (l *Limit) fillOrder(a, b *Order) Match
```

这里：

```text
a = 价格档位里已有的挂单
b = 新来的市价单
```

代码只按数量大小扣减：

```text
如果 a.Size >= b.Size:
  a.Size -= b.Size
  sizeFilled = b.Size
  b.Size = 0

否则:
  b.Size -= a.Size
  sizeFilled = a.Size
  a.Size = 0
```

例子 1：挂单比市价单大

```text
a = ask limit, size 10
b = market buy, size 4

成交 4
a 剩余 6
b 剩余 0
```

例子 2：市价单比挂单大

```text
a = ask limit, size 3
b = market buy, size 8

成交 3
a 剩余 0
b 剩余 5
```

`Match.Price` 使用的是当前价格档位的价格：

```go
Price: l.Price
```

这符合“主动单吃被动单，按被动挂单价格成交”的简化模型。

## 9. 撤单：CancelOrder

入口：

```go
func (ob *Orderbook) CancelOrder(o *Order)
```

撤单依赖 `Order.Limit`：

```text
1. limit := o.Limit
2. limit.DeleteOrder(o)
3. delete(ob.Orders, o.ID)
4. 如果 limit.Orders 空了:
     ob.clearLimit(o.Bid, limit)
```

也就是说，订单必须还挂在某个 `Limit` 上，才能被撤。

如果传进来的订单是 nil，或者订单已经成交、`o.Limit == nil`，当前代码会 panic。这个地方需要业务层提前保证，或者后续补防御。

## 10. 删除订单：Limit.DeleteOrder

入口：

```go
func (l *Limit) DeleteOrder(o *Order)
```

它从当前价格档位里删掉指定订单：

```text
for i := 0; i < len(l.Orders); i++ {
  如果 l.Orders[i] == o:
    用最后一个订单覆盖当前位置
    l.Orders 缩短一格
}

o.Limit = nil
l.TotalVolume -= o.Size
sort.Sort(l.Orders)
```

这里用了一个常见的 O(1) 删除技巧：

```text
原数组: [A, B, C, D]
删除 B:
  用 D 覆盖 B -> [A, D, C, D]
  缩短长度 -> [A, D, C]
```

然后再按时间排序，试图恢复 FIFO。

注意：如果同一个 `Limit` 有多个订单，这种“尾部覆盖再排序”的方式可以工作，但不如链表或稳定队列清晰。

## 11. 删除价格档位：clearLimit

入口：

```go
func (ob *Orderbook) clearLimit(bid bool, l *Limit)
```

当某个价格档位没有订单了，就要从订单簿里删除这个档位。

如果删除买盘档位：

```text
delete(ob.BidLimits, l.Price)
从 ob.bids 里移除 l
```

如果删除卖盘档位：

```text
delete(ob.AskLimits, l.Price)
从 ob.asks 里移除 l
```

这里也用了“最后一个元素覆盖当前位置”的删除方式。

## 12. Trade 与 Match 的区别

代码里有两个容易混的结构：

```go
type Match struct {
    Ask        *Order
    Bid        *Order
    SizeFilled float64
    Price      float64
}

type Trade struct {
    Price     float64
    Size      float64
    Bid       bool
    Timestamp int64
}
```

区别：

```text
Match:
  撮合过程的详细结果。
  知道是哪张买单、哪张卖单成交。

Trade:
  订单簿内保存的简化成交记录。
  只保存价格、数量、方向、时间。
```

`PlaceMarketOrder` 先生成 `matches`，然后把每个 `Match` 转成一个 `Trade`：

```text
Match{Ask: A, Bid: B, SizeFilled: 2, Price: 1010}
  -> Trade{Price: 1010, Size: 2, Bid: marketOrder.Bid, Timestamp: now}
```

## 13. 完整例子：从空订单簿开始

### 第一步：挂两个卖单

```go
ob := NewOrderbook()

ask1 := NewOrder(false, 2, 1)
ob.PlaceLimitOrder(1010, ask1)

ask2 := NewOrder(false, 5, 2)
ob.PlaceLimitOrder(1020, ask2)
```

状态：

```text
asks:
  1010: ask1 size 2
  1020: ask2 size 5

AskLimits:
  1010 -> Limit(1010)
  1020 -> Limit(1020)

Orders:
  ask1.ID -> ask1
  ask2.ID -> ask2
```

### 第二步：来一个市价买单 size 4

```go
buy := NewOrder(true, 4, 3)
matches, err := ob.PlaceMarketOrder(buy)
```

执行过程：

```text
buy size 4，先吃最低卖价 1010

1010 档位:
  ask1 size 2
  成交 2
  ask1 size -> 0
  buy size -> 2
  1010 档位被清空，删除

继续吃 1020 档位:
  ask2 size 5
  成交 2
  ask2 size -> 3
  buy size -> 0
  1020 档位保留
```

最终状态：

```text
asks:
  1020: ask2 size 3

AskLimits:
  1020 -> Limit(1020)

Orders:
  ask1.ID 仍然可能还在 ob.Orders 里
  ask2.ID -> ask2

Trades:
  2 @ 1010
  2 @ 1020
```

这里有一个重要问题：`Limit.DeleteOrder` 只从 `Limit.Orders` 删除订单，但市价单撮合吃空挂单时，并没有同步从 `ob.Orders` 删除已成交订单。撤单或查询时可能遇到旧订单引用。这是当前设计的一个不一致点。

## 14. 当前设计的状态一致性表

| 操作 | asks/bids | AskLimits/BidLimits | Orders map | Limit.Orders | Trades |
|---|---|---|---|---|---|
| PlaceLimitOrder | 可能新增档位 | 可能新增 price -> Limit | 新增 orderID -> Order | 新增订单 | 不变 |
| PlaceMarketOrder | 可能删除空档位 | 可能删除空档位 | 当前未删除吃空订单 | 删除吃空订单 | 新增成交 |
| CancelOrder | 可能删除空档位 | 可能删除空档位 | 删除订单 | 删除订单 | 不变 |
| clearLimit | 删除空档位 | 删除 price -> Limit | 不变 | 不变 | 不变 |

最值得注意的是：

```text
撮合成交后，已成交的挂单会从 Limit.Orders 删除，
但代码没有从 ob.Orders 删除。
```

如果后续要改进，这是优先级很高的点。

## 15. 主要问题与改进建议

### 15.1 订单 ID 可能重复

当前：

```go
ID: int64(rand.Intn(10000000))
```

问题：

```text
随机数可能碰撞。
```

建议：

```text
使用自增序列、snowflake、UUID，或者由上层服务生成唯一 ID。
```

### 15.2 使用 float64 表示价格和数量

问题：

```text
float64 有精度误差，不适合金融金额。
```

建议：

```text
用整数最小单位，例如 price = cents，size = satoshi-like unit。
或使用 decimal 库。
```

### 15.3 Limit.DeleteOrder 删除方式不直观

当前是尾部覆盖再排序。

问题：

```text
逻辑绕，不利于理解。
如果订单量不大，直接 append(l.Orders[:i], l.Orders[i+1:]...) 更清楚。
```

### 15.4 成交后 Orders map 未同步清理

问题：

```text
撮合吃空的挂单从 Limit.Orders 删除了，
但 ob.Orders 里还可能保留这个订单 ID。
```

建议：

```text
PlaceMarketOrder 在删除已成交挂单时，同时 delete(ob.Orders, order.ID)。
```

### 15.5 市价单没有部分成交

当前逻辑：

```text
如果订单簿总量不足，直接返回错误。
```

真实系统中可能允许：

```text
能成交多少成交多少，剩余取消。
```

这取决于业务设计，但需要明确。

### 15.6 限价单不会 crossing match

当前逻辑：

```text
所有限价单都直接挂入订单簿。
```

真实系统：

```text
买入限价 >= best ask，应立即吃卖盘。
卖出限价 <= best bid，应立即吃买盘。
剩余未成交部分再挂单。
```

### 15.7 并发读写锁使用不完整

`PlaceLimitOrder` 和 `PlaceMarketOrder` 加了写锁，但 `Asks()`、`Bids()`、`AskTotalVolume()`、`BidTotalVolume()` 没有单独加读锁。

在当前调用路径里，多数时候它们在写锁内部被调用，但如果外部直接并发调用这些方法，仍然可能有数据竞争。

## 16. 推荐重构方向

如果要让这个文件更容易学，可以按下面顺序重构：

1. 把乱码注释全部删除或重写。
2. 把 `Order.Side` 从 `bool` 改成明确枚举：`Buy` / `Sell`。
3. 把 `PlaceMarketOrder` 拆成 `matchBuyMarketOrder` 和 `matchSellMarketOrder`。
4. 把 `Limit.Fill` 改名为 `FillMarketOrder`。
5. 成交删除订单时同步维护 `ob.Orders`。
6. 给 `CancelOrder` 增加 nil 和不存在订单保护。
7. 加一组更小、更直观的单元测试，每个测试只验证一个行为。

## 17. 最小心智模型

只记这几句话就够了：

```text
Order 是一张订单。
Limit 是一个价格档位。
Orderbook 是很多价格档位组成的买卖盘。

买单放 bids，卖单放 asks。
买市价单吃 asks，从最低卖价开始吃。
卖市价单吃 bids，从最高买价开始吃。

Limit.Fill 负责在一个价格档位里扣数量。
PlaceMarketOrder 负责跨多个价格档位撮合。
```

这就是当前 `orderbook.go` 的核心结构。

// Package orderbook 实现了一个简化版的加密货币交易所订单薄（Order Book）。
//
// 核心概念：
//   - Order（订单）：用户提交的买/卖请求，包含数量、方向（买/卖）等信息
//   - Limit（价格档位）：同一价格的所有订单聚合在一个 Limit 中，类似于"挂单队列"
//   - Orderbook（订单薄）：管理所有买单（bids）和卖单（asks）的核心数据结构
//   - Match（撮合结果）：当市价单与限价单成交时产生的匹配记录
//   - Trade（成交记录）：每次撮合后记录的历史交易信息
//
// 订单薄的工作原理（简单比喻）：
//   想象一个菜市场的摊位。卖家在左边的黑板上标价（asks），买家在右边标价（bids）。
//   - 限价单（Limit Order）= 在黑板上写价格排队等待
//   - 市价单（Market Order）= 直接按黑板上的最优价格成交
//   - 买方最高出价（Best Bid）和卖方最低要价（Best Ask）之间的差就是"价差"（Spread）
package orderbook

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Trade 代表一笔已完成的成交记录。
// 每当市价单与限价单撮合成功，就会生成一个 Trade。
// 这些记录会被保存下来，供外部查询历史成交信息。
type Trade struct {
	Price     float64 // 成交价格
	Size      float64 // 成交数量
	Bid       bool    // true=这笔成交由买方市价单触发, false=由卖方市价单触发
	Timestamp int64   // 成交时间戳（纳秒级 Unix 时间）
}

// Match 代表一次订单撮合的结果。
// 当一个市价单进入订单薄时，它会与对手方的限价单逐一匹配，
// 每匹配一个限价单就会产生一个 Match。
//
// 例如：一个买入市价单（size=15）可能匹配到：
//   - 卖单A（size=10）→ Match{SizeFilled: 10, Price: 100}
//   - 卖单B（size=5）→ Match{SizeFilled: 5, Price: 101}
type Match struct {
	Ask        *Order  // 卖方订单（提供流动性的一方，或者市价卖单）
	Bid        *Order  // 买方订单（提供流动性的一方，或者市价买单）
	SizeFilled float64 // 本次撮合成交的数量
	Price      float64 // 成交价格（= 限价单所在 Limit 的价格）
}

// Order 代表一笔订单。
// 这是订单薄中最基本的单元，每个用户提交的买/卖请求都是一个 Order。
type Order struct {
	ID        int64   // 订单唯一标识（随机生成）
	UserID    int64   // 下单用户的 ID
	Size      float64 // 订单数量（会随着部分成交而减少，变为 0 表示完全成交）
	Bid       bool    // true=买单(Bid), false=卖单(Ask)
	Limit     *Limit  // 指向该订单所属的价格档位（市价单没有 Limit，为 nil）
	Timestamp int64   // 下单时间戳（纳秒级），用于同价格内按时间优先排序（FIFO）
}

// Orders 是 Order 指针的切片，实现了 sort.Interface 接口。
// 目的：让同一价格档位内的订单按时间先后排序（先来先成交，即 FIFO 原则）。
type Orders []*Order

func (o Orders) Len() int           { return len(o) }
func (o Orders) Swap(i, j int)      { o[i], o[j] = o[j], o[i] }
func (o Orders) Less(i, j int) bool { return o[i].Timestamp < o[j].Timestamp } // 时间小的排前面

// NewOrder 创建一个新订单。
// 参数：
//   - bid: true=买单, false=卖单
//   - size: 下单数量（比如要买/卖多少个 ETH）
//   - userID: 下单用户的 ID
//
// 注意：ID 通过 rand.Intn 随机生成，在生产环境中应该用分布式 ID 生成器（如 snowflake）。
func NewOrder(bid bool, size float64, userID int64) *Order {
	return &Order{
		UserID:    userID,
		ID:        int64(rand.Intn(10000000)),
		Size:      size,
		Bid:       bid,
		Timestamp: time.Now().UnixNano(),
	}
}

// String 返回订单的可读字符串表示，用于调试输出。
func (o *Order) String() string {
	return fmt.Sprintf("[size: %.2f] | [id: %d]", o.Size, o.ID)
}

// Type 返回订单方向的字符串表示："BID"（买）或 "ASK"（卖）。
func (o *Order) Type() string {
	if o.Bid {
		return "BID"
	}
	return "ASK"
}

// IsFilled 判断订单是否已完全成交（Size 变为 0）。
func (o *Order) IsFilled() bool {
	return o.Size == 0.0
}

// Limit 代表一个"价格档位"。
// 同一个价格的所有订单被组织在同一个 Limit 中。
//
// 可以把 Limit 想象成超市里的一个货架：
//   - Price = 货架上标的价格
//   - Orders = 这个价格上排队的所有订单（按时间 FIFO 排列）
//   - TotalVolume = 这个价格上所有订单的总量
//
// 例如：价格 10000 的 Limit 可能有 3 个卖单，总量 25 ETH
type Limit struct {
	Price       float64 // 该档位的价格
	Orders      Orders  // 该价格上所有挂着的订单（按时间排序）
	TotalVolume float64 // 该价格上所有订单的总数量
}

// Limits 是 Limit 指针的切片，下面的 ByBestAsk 和 ByBestBid 用它来排序。
type Limits []*Limit

// ByBestAsk 实现了 sort.Interface，按价格从低到高排序。
// 原因：对于卖单（Ask），价格越低越优先成交（卖家出价越低，买家越愿意买）。
// 排序后 asks[0] 就是最优卖价（Best Ask）。
type ByBestAsk struct{ Limits }

func (a ByBestAsk) Len() int           { return len(a.Limits) }
func (a ByBestAsk) Swap(i, j int)      { a.Limits[i], a.Limits[j] = a.Limits[j], a.Limits[i] }
func (a ByBestAsk) Less(i, j int) bool { return a.Limits[i].Price < a.Limits[j].Price }

// ByBestBid 实现了 sort.Interface，按价格从高到低排序。
// 原因：对于买单（Bid），价格越高越优先成交（买家出价越高，卖家越愿意卖）。
// 排序后 bids[0] 就是最优买价（Best Bid）。
type ByBestBid struct{ Limits }

func (b ByBestBid) Len() int           { return len(b.Limits) }
func (b ByBestBid) Swap(i, j int)      { b.Limits[i], b.Limits[j] = b.Limits[j], b.Limits[i] }
func (b ByBestBid) Less(i, j int) bool { return b.Limits[i].Price > b.Limits[j].Price }

// NewLimit 创建一个新的价格档位。
func NewLimit(price float64) *Limit {
	return &Limit{
		Price:  price,
		Orders: []*Order{},
	}
}

// AddOrder 将一个订单添加到该价格档位中。
// 同时建立双向引用：Order.Limit 指向 Limit，Limit.Orders 包含 Order。
func (l *Limit) AddOrder(o *Order) {
	o.Limit = l                // 让订单知道自己属于哪个价格档位
	l.Orders = append(l.Orders, o)
	l.TotalVolume += o.Size    // 累加该价格档位的总量
}

// DeleteOrder 从该价格档位中移除指定订单。
// 使用"将最后一个元素替换到被删除位置"的技巧来避免数组移动（O(1) 删除）。
// 删除后重新排序以保持 FIFO 顺序。
func (l *Limit) DeleteOrder(o *Order) {
	for i := 0; i < len(l.Orders); i++ {
		if l.Orders[i] == o {
			// 技巧：用最后一个元素覆盖当前位置，然后截断切片
			// 比 append(l.Orders[:i], l.Orders[i+1:]...) 更高效
			l.Orders[i] = l.Orders[len(l.Orders)-1]
			l.Orders = l.Orders[:len(l.Orders)-1]
		}
	}

	o.Limit = nil              // 断开订单与价格档位的引用
	l.TotalVolume -= o.Size    // 减去该订单的数量

	sort.Sort(l.Orders)        // 因为替换打乱了顺序，需要重新按时间排序
}

// Fill 用传入的订单 o（通常是市价单）来"吃掉"该价格档位上的限价单。
// 按 FIFO 顺序逐个匹配，直到市价单被完全填满或该档位的所有订单都被消耗。
//
// 返回所有撮合结果（[]Match）。
//
// 流程图：
//   市价买单(size=15) → 价格档位(price=100, 3个卖单)
//     → 卖单1(size=5): 成交5, 卖单1被填满 ✓
//     → 卖单2(size=8): 成交8, 卖单2被填满 ✓
//     → 卖单3(size=10): 成交2, 市价单被填满, 卖单3剩余8 ✓
func (l *Limit) Fill(o *Order) []Match {
	var (
		matches        []Match
		ordersToDelete []*Order
	)

	for _, order := range l.Orders {
		if o.IsFilled() { // 如果传入的市价单已经全部成交，停止匹配
			break
		}

		match := l.fillOrder(order, o) // 两两匹配
		matches = append(matches, match)

		l.TotalVolume -= match.SizeFilled // 减少该价格档位的总量

		if order.IsFilled() { // 如果限价单被完全消耗，标记为待删除
			ordersToDelete = append(ordersToDelete, order)
		}
	}

	// 批量删除已完全成交的限价单
	// （不在上面的循环中直接删除，是因为修改正在遍历的切片会导致问题）
	for _, order := range ordersToDelete {
		l.DeleteOrder(order)
	}

	return matches
}

// fillOrder 执行两个订单之间的实际数量交换。
// 参数 a 是限价单（挂单方），b 是市价单（吃单方）。
//
// 核心逻辑：
//   - 如果 a 的数量 >= b 的数量：b 被完全填满，a 剩余一部分
//   - 如果 a 的数量 < b 的数量：a 被完全填满，b 还需要继续匹配下一个订单
func (l *Limit) fillOrder(a, b *Order) Match {
	var (
		bid        *Order
		ask        *Order
		sizeFilled float64
	)

	// 确定哪个是买单、哪个是卖单（因为参数可能是任意顺序传入）
	if a.Bid {
		bid = a
		ask = b
	} else {
		bid = b
		ask = a
	}

	// 执行数量交换
	if a.Size >= b.Size {
		a.Size -= b.Size   // a 还有剩余
		sizeFilled = b.Size // 成交量 = b 的全部数量
		b.Size = 0.0        // b 被完全填满
	} else {
		b.Size -= a.Size   // b 还有剩余，需要继续匹配
		sizeFilled = a.Size // 成交量 = a 的全部数量
		a.Size = 0.0        // a 被完全填满
	}

	return Match{
		Bid:        bid,
		Ask:        ask,
		SizeFilled: sizeFilled,
		Price:      l.Price, // 成交价格 = 限价单所在的价格档位
	}
}

// Orderbook 是整个订单薄的核心数据结构。
// 它管理所有的买单（bids）和卖单（asks），并提供下单、撮合、取消等操作。
//
// 数据结构设计：
//   - asks/bids: 切片形式的价格档位列表，用于排序和遍历
//   - AskLimits/BidLimits: map 形式的价格到档位的映射，用于 O(1) 查找
//   - Orders: 订单 ID 到订单的映射，用于快速按 ID 查找/取消订单
//   - Trades: 成交记录的历史列表
//
// 为什么同时用切片和 map？
//   - 切片方便排序（找最优价格）
//   - map 方便按价格快速查找（下限价单时检查该价格是否已有档位）
//   这是典型的"空间换时间"策略。
type Orderbook struct {
	asks []*Limit // 所有卖单价格档位（排序后 asks[0] = 最低卖价 = Best Ask）
	bids []*Limit // 所有买单价格档位（排序后 bids[0] = 最高买价 = Best Bid）

	Trades []*Trade // 历史成交记录

	mu        sync.RWMutex       // 读写锁，保证并发安全
	AskLimits map[float64]*Limit // 价格 → 卖单档位 的映射（O(1) 查找）
	BidLimits map[float64]*Limit // 价格 → 买单档位 的映射（O(1) 查找）
	Orders    map[int64]*Order   // 订单ID → 订单 的映射（O(1) 查找/取消）
}

// NewOrderbook 创建一个空的订单薄。
func NewOrderbook() *Orderbook {
	return &Orderbook{
		asks:      []*Limit{},
		bids:      []*Limit{},
		Trades:    []*Trade{},
		AskLimits: make(map[float64]*Limit),
		BidLimits: make(map[float64]*Limit),
		Orders:    make(map[int64]*Order),
	}
}

// PlaceMarketOrder 处理市价单。
// 市价单不指定价格，而是按当前订单薄中最优价格立即成交。
//
// 流程：
//   1. 检查对手方是否有足够的挂单量来填满这个市价单
//   2. 按价格优先顺序遍历对手方的价格档位，逐一撮合
//   3. 如果某个价格档位的订单全部被消耗，清除该档位
//   4. 记录成交历史（Trade）
//
// 例如：一个买入市价单会从最低卖价开始"吃单"，逐步向高价吃，
// 这就是为什么大额市价单会导致价格滑点（slippage）。
func (ob *Orderbook) PlaceMarketOrder(o *Order) []Match {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	matches := []Match{}

	if o.Bid {
		// 买入市价单 → 消耗卖单（asks）
		if o.Size > ob.AskTotalVolume() {
			// 安全检查：如果要买的数量超过所有卖单的总量，无法成交
			panic(fmt.Errorf("not enough volume [size: %.2f] for market order [size: %.2f]", ob.AskTotalVolume(), o.Size))
		}

		// 从最低卖价开始，逐个价格档位撮合
		for _, limit := range ob.Asks() {
			limitMatches := limit.Fill(o) // 在该价格档位内逐个订单撮合
			matches = append(matches, limitMatches...)

			// 如果该价格档位的所有订单都被吃光了，清除该档位
			if len(limit.Orders) == 0 {
				ob.clearLimit(false, limit) // false = 不是 bid 档位，是 ask 档位
			}
		}
	} else {
		// 卖出市价单 → 消耗买单（bids）
		if o.Size > ob.BidTotalVolume() {
			panic(fmt.Errorf("not enough volume [size: %.2f] for market order [size: %.2f]", ob.BidTotalVolume(), o.Size))
		}

		// 从最高买价开始，逐个价格档位撮合
		for _, limit := range ob.Bids() {
			limitMatches := limit.Fill(o)
			matches = append(matches, limitMatches...)

			if len(limit.Orders) == 0 {
				ob.clearLimit(true, limit) // true = 是 bid 档位
			}
		}
	}

	// 将所有撮合结果记录为成交历史
	for _, match := range matches {
		trade := &Trade{
			Price:     match.Price,
			Size:      match.SizeFilled,
			Timestamp: time.Now().UnixNano(),
			Bid:       o.Bid,
		}
		ob.Trades = append(ob.Trades, trade)
	}

	// 打印当前最新成交价格（最后一笔 Trade 的价格）
	logrus.WithFields(logrus.Fields{
		"currentPrice": ob.Trades[len(ob.Trades)-1].Price,
	}).Info()

	return matches
}

// PlaceLimitOrder 处理限价单。
// 限价单指定一个价格，挂在订单薄中等待被市价单匹配。
//
// 流程：
//   1. 检查该价格是否已有对应的 Limit（价格档位）
//   2. 如果没有，创建一个新的 Limit 并加入到 asks 或 bids 列表中
//   3. 将订单添加到对应的 Limit 中
//   4. 在全局 Orders map 中注册该订单（方便后续按 ID 查找/取消）
//
// 注意：这个简化版没有实现"限价单立即匹配"的逻辑。
// 在真实交易所中，如果限价买单的价格 >= 当前最优卖价，应该立即成交（taker），
// 而不是挂在订单薄里。这里简化了，所有限价单都直接挂单。
func (ob *Orderbook) PlaceLimitOrder(price float64, o *Order) {
	var limit *Limit

	ob.mu.Lock()
	defer ob.mu.Unlock()

	// 查找该价格是否已有 Limit
	if o.Bid {
		limit = ob.BidLimits[price]
	} else {
		limit = ob.AskLimits[price]
	}

	// 如果该价格还没有 Limit，创建一个新的
	if limit == nil {
		limit = NewLimit(price)

		if o.Bid {
			ob.bids = append(ob.bids, limit)  // 加入买单列表
			ob.BidLimits[price] = limit        // 加入买单 map
		} else {
			ob.asks = append(ob.asks, limit)  // 加入卖单列表
			ob.AskLimits[price] = limit        // 加入卖单 map
		}
	}

	logrus.WithFields(logrus.Fields{
		"price":  limit.Price,
		"type":   o.Type(),
		"size":   o.Size,
		"userID": o.UserID,
	}).Info("new limit order")

	ob.Orders[o.ID] = o // 在全局订单 map 中注册
	limit.AddOrder(o)   // 将订单挂到该价格档位上
}

// clearLimit 清除一个已空的价格档位。
// 当一个价格档位的所有订单都被成交完毕，需要从订单薄中移除该档位。
//
// 同时从 map 和切片中移除：
//   - map: delete(ob.BidLimits, l.Price) → O(1)
//   - 切片: 用"末尾替换"技巧 → O(1)
func (ob *Orderbook) clearLimit(bid bool, l *Limit) {
	if bid {
		delete(ob.BidLimits, l.Price)
		for i := 0; i < len(ob.bids); i++ {
			if ob.bids[i] == l {
				ob.bids[i] = ob.bids[len(ob.bids)-1]
				ob.bids = ob.bids[:len(ob.bids)-1]
			}
		}
	} else {
		delete(ob.AskLimits, l.Price)
		for i := 0; i < len(ob.asks); i++ {
			if ob.asks[i] == l {
				ob.asks[i] = ob.asks[len(ob.asks)-1]
				ob.asks = ob.asks[:len(ob.asks)-1]
			}
		}
	}

	fmt.Printf("clearing limit price level [%.2f]\n", l.Price)
}

// CancelOrder 取消一个订单。
// 从其所属的 Limit 中移除，并从全局 Orders map 中删除。
// 如果该 Limit 变空了，也一并清除该价格档位。
func (ob *Orderbook) CancelOrder(o *Order) {
	limit := o.Limit
	limit.DeleteOrder(o)
	delete(ob.Orders, o.ID)

	// 如果这个价格档位再也没有订单了，整个档位都清掉
	if len(limit.Orders) == 0 {
		ob.clearLimit(o.Bid, limit)
	}
}

// BidTotalVolume 计算所有买单价格档位的总量。
// 用于检查卖出市价单时是否有足够的买方流动性。
func (ob *Orderbook) BidTotalVolume() float64 {
	totalVolume := 0.0

	for i := 0; i < len(ob.bids); i++ {
		totalVolume += ob.bids[i].TotalVolume
	}

	return totalVolume
}

// AskTotalVolume 计算所有卖单价格档位的总量。
// 用于检查买入市价单时是否有足够的卖方流动性。
func (ob *Orderbook) AskTotalVolume() float64 {
	totalVolume := 0.0

	for i := 0; i < len(ob.asks); i++ {
		totalVolume += ob.asks[i].TotalVolume
	}

	return totalVolume
}

// Asks 返回所有卖单价格档位，按价格从低到高排序。
// asks[0] = 最低卖价 = Best Ask（最优卖价）。
// 每次调用都会重新排序，确保返回最新顺序。
func (ob *Orderbook) Asks() []*Limit {
	sort.Sort(ByBestAsk{ob.asks})
	return ob.asks
}

// Bids 返回所有买单价格档位，按价格从高到低排序。
// bids[0] = 最高买价 = Best Bid（最优买价）。
// 每次调用都会重新排序，确保返回最新顺序。
func (ob *Orderbook) Bids() []*Limit {
	sort.Sort(ByBestBid{ob.bids})
	return ob.bids
}

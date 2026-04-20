// orderbook_test.go 是 orderbook 包的单元测试文件。
//
// Go 的测试约定：
//   - 测试文件以 _test.go 结尾
//   - 测试函数以 Test 开头，参数是 *testing.T
//   - 运行测试：go test ./orderbook/ 或在项目根目录 go test ./...
//   - 测试文件和被测代码在同一个包中（这里都是 package orderbook），
//     可以直接访问包内的私有成员（如 ob.asks、ob.bids）
package orderbook

import (
	"fmt"
	"reflect"
	"testing"
)

// assert 是一个辅助断言函数，比较两个值是否深度相等。
// 使用 reflect.DeepEqual 进行比较，支持任意类型（包括切片、结构体等）。
// 如果不相等，调用 t.Errorf 报告错误（但不会中止测试，会继续执行后面的断言）。
//
// 注意：生产项目通常使用第三方断言库（如 testify）而不是手写 assert。
func assert(t *testing.T, a, b any) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("%+v != %+v", a, b)
	}
}

// TestLastMarketTrades 测试市价单成交后是否正确记录了 Trade（成交记录）。
//
// 场景：
//   1. 在价格 10000 挂一个卖单（size=10）
//   2. 下一个买入市价单（size=10）
//   3. 验证：
//      - 只产生了 1 个 Match（因为 1 个卖单就足够填满整个市价单）
//      - Trade 记录数量为 1
//      - Trade 的价格 = 10000（限价单的价格）
//      - Trade 的方向 = 买入（市价单是买入方向）
//      - Trade 的成交量 = Match 的 SizeFilled
func TestLastMarketTrades(t *testing.T) {
	ob := NewOrderbook()
	price := 10000.0

	// 先挂一个卖单作为流动性
	sellOrder := NewOrder(false, 10, 0)
	ob.PlaceLimitOrder(price, sellOrder)

	// 下一个买入市价单来"吃掉"卖单
	marketOrder := NewOrder(true, 10, 0)
	matches := ob.PlaceMarketOrder(marketOrder)
	assert(t, len(matches), 1) // 应该只有 1 个 Match
	match := matches[0]

	// 验证成交记录
	assert(t, len(ob.Trades), 1)
	trade := ob.Trades[0]
	assert(t, trade.Price, price)          // 成交价格 = 限价单的价格
	assert(t, trade.Bid, marketOrder.Bid)  // 成交方向 = 市价单的方向
	assert(t, trade.Size, match.SizeFilled) // 成交量 = 撮合量
}

// TestLimit 测试 Limit（价格档位）的基本增删操作。
//
// 场景：
//   1. 创建一个价格为 10000 的 Limit
//   2. 添加 3 个买单
//   3. 删除中间一个（buyOrderB）
//   4. 通过打印验证剩余 2 个订单
//
// 这主要验证 Limit.AddOrder 和 Limit.DeleteOrder 的正确性。
func TestLimit(t *testing.T) {
	l := NewLimit(10_000)
	buyOrderA := NewOrder(true, 5, 0)
	buyOrderB := NewOrder(true, 8, 0)
	buyOrderC := NewOrder(true, 10, 0)

	l.AddOrder(buyOrderA)
	l.AddOrder(buyOrderB)
	l.AddOrder(buyOrderC)

	l.DeleteOrder(buyOrderB) // 删除中间的订单

	fmt.Println(l) // 打印查看结果（应只剩 A 和 C）
}

// TestPlaceLimitOrder 测试在订单薄中放置限价单。
//
// 场景：
//   1. 在两个不同价格（10000 和 9000）各放一个卖单
//   2. 验证：
//      - Orders map 中总共 2 个订单
//      - 可以通过 ID 找到对应的订单
//      - asks 列表中有 2 个价格档位（因为两个订单价格不同）
func TestPlaceLimitOrder(t *testing.T) {
	ob := NewOrderbook()

	sellOrderA := NewOrder(false, 10, 0) // 卖单 A: 10个ETH
	sellOrderB := NewOrder(false, 5, 0)  // 卖单 B: 5个ETH
	ob.PlaceLimitOrder(10_000, sellOrderA) // 挂在价格 10000
	ob.PlaceLimitOrder(9_000, sellOrderB)  // 挂在价格 9000

	assert(t, len(ob.Orders), 2)                     // 总共 2 个订单
	assert(t, ob.Orders[sellOrderA.ID], sellOrderA) // 可以通过 ID 找到 A
	assert(t, ob.Orders[sellOrderB.ID], sellOrderB) // 可以通过 ID 找到 B
	assert(t, len(ob.asks), 2)                       // 2 个价格档位
}

// TestPlaceMarketOrder 测试单个市价单的撮合（部分成交场景）。
//
// 场景：
//   1. 挂一个卖单：price=10000, size=20
//   2. 下一个买入市价单：size=10
//   3. 预期结果：
//      - 产生 1 个 Match（市价单在第一个卖单上就被全部消耗了）
//      - 卖单还剩 10（20 - 10）
//      - 仍有 1 个价格档位（卖单没被完全消耗）
//      - 卖方总量从 20 变成 10
//      - 买单被完全填满（IsFilled = true）
func TestPlaceMarketOrder(t *testing.T) {
	ob := NewOrderbook()

	sellOrder := NewOrder(false, 20, 0)  // 卖 20 个
	ob.PlaceLimitOrder(10_000, sellOrder)

	buyOrder := NewOrder(true, 10, 0)    // 买 10 个
	matches := ob.PlaceMarketOrder(buyOrder)

	assert(t, len(matches), 1)                 // 1 个 Match
	assert(t, len(ob.asks), 1)                 // 卖单没被吃光，还有 1 个档位
	assert(t, ob.AskTotalVolume(), 10.0)       // 还剩 10
	assert(t, matches[0].Ask, sellOrder)       // 匹配的卖单 = sellOrder
	assert(t, matches[0].Bid, buyOrder)        // 匹配的买单 = buyOrder
	assert(t, matches[0].SizeFilled, 10.0)     // 成交 10
	assert(t, matches[0].Price, 10_000.0)      // 成交价 = 10000
	assert(t, buyOrder.IsFilled(), true)       // 买单完全成交
}

// TestPlaceMarketOrderMultiFill 测试市价单跨多个价格档位的多次撮合。
//
// 场景（重点！这是理解撮合引擎的核心测试）：
//
//   订单薄初始状态（买方）：
//     价格 10000: buyOrderA (size=5)
//     价格  9000: buyOrderB (size=8)
//     价格  5000: buyOrderC (size=1), buyOrderD (size=1)
//     买方总量 = 5 + 8 + 1 + 1 = 15
//
//   下一个卖出市价单 (size=10)：
//     → 先吃最高买价 10000 的 buyOrderA (size=5)，成交 5，buyOrderA 被填满 ✓
//     → 再吃次高价 9000 的 buyOrderB (size=8)，成交 5（市价单剩余5），buyOrderB 未填满
//     → 市价单已经全部成交，停止
//
//   最终状态：
//     买方总量 = 0 + 3 + 1 + 1 = 5（从 15 减少了 10）
//     买方档位 = 2（价格 10000 的档位被清空，只剩 9000 和 5000）
//     Match 数量 = 2（跨了 2 个价格档位）
func TestPlaceMarketOrderMultiFill(t *testing.T) {
	ob := NewOrderbook()

	buyOrderA := NewOrder(true, 5, 0) // 将被完全填满
	buyOrderB := NewOrder(true, 8, 0) // 将被部分填满（剩余 3）
	buyOrderC := NewOrder(true, 1, 0) // 不受影响
	buyOrderD := NewOrder(true, 1, 0) // 不受影响

	// 注意挂单顺序：先挂低价，再挂高价
	// 但 Bids() 排序后会按价格从高到低，所以成交顺序是 A → B → C/D
	ob.PlaceLimitOrder(5_000, buyOrderC)
	ob.PlaceLimitOrder(5_000, buyOrderD)
	ob.PlaceLimitOrder(9_000, buyOrderB)
	ob.PlaceLimitOrder(10_000, buyOrderA)

	assert(t, ob.BidTotalVolume(), 15.00) // 买方总量 = 15

	// 下卖出市价单，size=10
	sellOrder := NewOrder(false, 10, 0)
	matches := ob.PlaceMarketOrder(sellOrder)

	assert(t, ob.BidTotalVolume(), 5.00) // 还剩 5（= 3 + 1 + 1）
	assert(t, len(ob.bids), 2)           // 还剩 2 个价格档位（9000 和 5000）
	assert(t, len(matches), 2)           // 跨了 2 个价格档位撮合
}

// TestCancelOrderAsk 测试取消一个卖单。
//
// 场景：
//   1. 挂一个卖单：price=10000, size=4
//   2. 取消该订单
//   3. 验证：
//      - 卖方总量变为 0
//      - 订单从 Orders map 中被移除
//      - 对应的价格档位也被清除（因为没有其他订单了）
func TestCancelOrderAsk(t *testing.T) {
	ob := NewOrderbook()
	sellOrder := NewOrder(false, 4, 0)
	price := 10_000.0
	ob.PlaceLimitOrder(price, sellOrder)

	assert(t, ob.AskTotalVolume(), 4.0) // 挂单后总量 = 4

	ob.CancelOrder(sellOrder) // 取消订单
	assert(t, ob.AskTotalVolume(), 0.0) // 总量归零

	// 验证订单已从 map 中移除
	_, ok := ob.Orders[sellOrder.ID]
	assert(t, ok, false)

	// 验证价格档位也被清除
	_, ok = ob.AskLimits[price]
	assert(t, ok, false)
}

// TestCancelOrderBid 测试取消一个买单（同上，但是买方向）。
func TestCancelOrderBid(t *testing.T) {
	ob := NewOrderbook()
	buyOrder := NewOrder(true, 4, 0)
	price := 10_000.0
	ob.PlaceLimitOrder(price, buyOrder)

	assert(t, ob.BidTotalVolume(), 4.0)

	ob.CancelOrder(buyOrder)
	assert(t, ob.BidTotalVolume(), 0.0)

	_, ok := ob.Orders[buyOrder.ID]
	assert(t, ok, false)

	_, ok = ob.BidLimits[price]
	assert(t, ok, false)
}

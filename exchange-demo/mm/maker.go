// Package mm 实现了一个简化版的做市商（Market Maker）策略。
//
// ====== 什么是做市商？======
// 做市商是金融市场中的"流动性提供者"。
// 它同时在买方（Bid）和卖方（Ask）挂限价单，赚取买卖价差（Spread）。
//
// 简单比喻：做市商就像街边的换汇小摊贩。
//   - 他以 6.98 元/美元的价格"买入"（Bid），标在一块牌子上
//   - 他以 7.02 元/美元的价格"卖出"（Ask），标在另一块牌子上
//   - 每成交一笔，他赚 0.04 元的差价（Spread = 7.02 - 6.98）
//   - 如果没人来换，他就一直挂着等；如果有人来换，他立刻补挂新的牌子
//
// 做市商的核心目标：
//   1. 提供流动性（让市场上总有买单和卖单）
//   2. 赚取买卖价差（Spread）
//   3. 控制风险（不要持有太多单方向的仓位）
//
// ====== 本文件实现的策略 ======
// 每隔 makeInterval 时间：
//   1. 查询当前最优买价和最优卖价
//   2. 如果订单薄为空，先"播种"（seedMarket）
//   3. 如果 Spread > MinSpread，在买卖双方各挂一个限价单
//      - 买单价格 = bestBid + priceOffset（稍微提高出价来抢位）
//      - 卖单价格 = bestAsk - priceOffset（稍微降低要价来抢位）
package mm

import (
	"time"

	"github.com/anthdm/crypto-exchange/client"
	"github.com/sirupsen/logrus"
)

// Config 是做市商的配置参数。
type Config struct {
	UserID         int64          // 做市商账户的用户 ID
	OrderSize      float64        // 每次挂单的数量（如 10 ETH）
	MinSpread      float64        // 最小价差阈值：只有当 Spread > MinSpread 时才挂单
	SeedOffset     float64        // 播种时买卖单距离"当前价格"的偏移量
	ExchangeClient *client.Client // 交易所 HTTP 客户端
	MakeInterval   time.Duration  // 做市循环的间隔时间（如每 1 秒执行一次）
	PriceOffset    float64        // 挂单时相对最优价的偏移量
}

// MarketMaker 是做市商的核心结构体。
// 所有配置字段都是私有的（小写），只能通过 NewMakerMaker 构造。
type MarketMaker struct {
	userID         int64
	orderSize      float64
	minSpread      float64
	seedOffset     float64
	priceOffset    float64
	exchangeClient *client.Client
	makeInterval   time.Duration
}

// NewMakerMaker 根据配置创建一个新的做市商实例。
// （注意：函数名叫 NewMakerMaker 而不是 NewMarketMaker，可能是个笔误 😄）
func NewMakerMaker(cfg Config) *MarketMaker {
	return &MarketMaker{
		userID:         cfg.UserID,
		orderSize:      cfg.OrderSize,
		minSpread:      cfg.MinSpread,
		seedOffset:     cfg.SeedOffset,
		exchangeClient: cfg.ExchangeClient,
		makeInterval:   cfg.MakeInterval,
		priceOffset:    cfg.PriceOffset,
	}
}

// Start 启动做市商。在一个新的 goroutine 中运行做市循环。
// 使用 go mm.makerLoop() 使做市逻辑异步执行，不阻塞调用方。
func (mm *MarketMaker) Start() {
	logrus.WithFields(logrus.Fields{
		"id":           mm.userID,
		"orderSize":    mm.orderSize,
		"makeInterval": mm.makeInterval,
		"minSpread":    mm.minSpread,
		"priceOffset":  mm.priceOffset,
	}).Info("starting market maker")

	go mm.makerLoop()
}

// makerLoop 是做市商的核心循环。
// 每隔 makeInterval 执行一次做市逻辑。
//
// 策略逻辑：
//
//	┌─────────────────────────────────────────┐
//	│              查询 Best Bid / Ask         │
//	└───────────────┬─────────────────────────┘
//	                ↓
//	     ┌────── 都为 0？──────┐
//	     │ Yes                 │ No
//	     ↓                     ↓
//	 seedMarket()        计算 Spread
//	 （播种初始订单）     Spread = BestAsk - BestBid
//	                           ↓
//	              ┌─── Spread <= MinSpread? ───┐
//	              │ Yes                        │ No
//	              ↓                            ↓
//	          跳过本轮                    挂买单 + 挂卖单
//	         （价差太窄，              买: BestBid + Offset
//	          不值得做市）              卖: BestAsk - Offset
func (mm *MarketMaker) makerLoop() {
	ticker := time.NewTicker(mm.makeInterval)

	for {
		// 第一步：获取当前最优买价
		bestBid, err := mm.exchangeClient.GetBestBid()
		if err != nil {
			logrus.Error(err)
			break
		}

		// 第二步：获取当前最优卖价
		bestAsk, err := mm.exchangeClient.GetBestAsk()
		if err != nil {
			logrus.Error(err)
			break
		}

		// 第三步：如果订单薄完全为空（买卖双方都没有挂单），
		// 做市商需要"播种"——先挂上初始的买单和卖单，给市场提供初始流动性
		if bestAsk.Price == 0 && bestBid.Price == 0 {
			if err := mm.seedMarket(); err != nil {
				logrus.Error(err)
				break
			}
			continue // 播种后跳过本轮，等下次循环再检查
		}

		// 边界情况处理：如果只有一方有挂单
		// 如果没有买单，则用卖价减去一个偏移量作为"虚拟买价"
		if bestBid.Price == 0 {
			bestBid.Price = bestAsk.Price - mm.priceOffset*2
		}

		// 如果没有卖单，则用买价加上一个偏移量作为"虚拟卖价"
		if bestAsk.Price == 0 {
			bestAsk.Price = bestBid.Price + mm.priceOffset*2
		}

		// 第四步：计算当前价差
		// Spread = 最低卖价 - 最高买价
		// 例如：Best Ask=1050, Best Bid=950, Spread=100
		spread := bestAsk.Price - bestBid.Price

		// 如果价差太小（<= MinSpread），跳过本轮。
		// 原因：如果价差太窄，做市商挂单后几乎没有利润空间，
		// 反而可能因为价格波动而亏损。
		if spread <= mm.minSpread {
			continue
		}

		// 第五步：挂买单和卖单
		// 买单价格 = 当前最优买价 + priceOffset
		// 含义：比当前最高买价稍高一点，"插队"排在最前面
		// 这样做市商的买单更容易被成交
		if err := mm.placeOrder(true, bestBid.Price+mm.priceOffset); err != nil {
			logrus.Error(err)
			break
		}

		// 卖单价格 = 当前最优卖价 - priceOffset
		// 含义：比当前最低卖价稍低一点，同样是"插队"
		if err := mm.placeOrder(false, bestAsk.Price-mm.priceOffset); err != nil {
			logrus.Error(err)
			break
		}

		// 等待 ticker 触发下一次循环
		<-ticker.C
	}
}

// placeOrder 挂一个限价单到交易所。
// 这是对 client.PlaceLimitOrder 的简单封装。
func (mm *MarketMaker) placeOrder(bid bool, price float64) error {
	bidOrder := &client.PlaceOrderParams{
		UserID: mm.userID,
		Size:   mm.orderSize,
		Bid:    bid,
		Price:  price,
	}
	_, err := mm.exchangeClient.PlaceLimitOrder(bidOrder)
	return err
}

// seedMarket "播种"市场，在空的订单薄中放入初始的买单和卖单。
//
// 当交易所刚启动时，订单薄是空的，没有任何流动性。
// 做市商需要先放入一对买/卖单来"引导"市场。
//
// 播种逻辑：
//   1. 获取当前 ETH 的"参考价格"（这里用 simulateFetchCurrentETHPrice 模拟）
//   2. 在参考价格的上下各偏移 seedOffset 的位置挂单：
//      - 买单: currentPrice - seedOffset (比如 1000 - 40 = 960)
//      - 卖单: currentPrice + seedOffset (比如 1000 + 40 = 1040)
//   3. 这样就建立了一个初始的 Spread = seedOffset * 2 = 80
func (mm *MarketMaker) seedMarket() error {
	// 获取当前 ETH 参考价格（模拟从其他交易所获取价格）
	currentPrice := simulateFetchCurrentETHPrice()

	logrus.WithFields(logrus.Fields{
		"currentETHPrice": currentPrice,
		"seedOffset":      mm.seedOffset,
	}).Info("orderbooks empty => seeding market!")

	// 挂一个买单：在参考价格下方
	bidOrder := &client.PlaceOrderParams{
		UserID: mm.userID,
		Size:   mm.orderSize,
		Bid:    true,
		Price:  currentPrice - mm.seedOffset,
	}
	_, err := mm.exchangeClient.PlaceLimitOrder(bidOrder)
	if err != nil {
		return err
	}

	// 挂一个卖单：在参考价格上方
	askOrder := &client.PlaceOrderParams{
		UserID: mm.userID,
		Size:   mm.orderSize,
		Bid:    false,
		Price:  currentPrice + mm.seedOffset,
	}
	_, err = mm.exchangeClient.PlaceLimitOrder(askOrder)

	return err
}

// simulateFetchCurrentETHPrice 模拟从外部交易所（如 Binance、Coinbase）获取当前 ETH 价格。
//
// 在真实场景中，做市商需要参考其他交易所的实时价格来确定自己的报价。
// 这里简化为等 80ms 后返回固定值 1000.0。
//
// 80ms 的延迟模拟了网络请求的耗时（真实 API 调用通常需要几十到几百毫秒）。
func simulateFetchCurrentETHPrice() float64 {
	time.Sleep(80 * time.Millisecond)

	return 1000.0
}

// Package server 实现了加密货币交易所的 HTTP 服务端。
//
// 架构概览：
//
//	Client (HTTP) → Echo Router → Exchange (业务逻辑) → Orderbook (内存撮合引擎)
//	                                                  ↘ PostgreSQL (账本/成交记录持久化)
package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/anthdm/crypto-exchange/db"
	"github.com/anthdm/crypto-exchange/orderbook"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	MarketETH Market = "ETH" // 当前交易所唯一支持的交易对：ETH

	MarketOrder OrderType = "MARKET" // 市价单：不指定价格，按照当前订单薄里的最优价格立刻成交
	LimitOrder  OrderType = "LIMIT"  // 限价单：指定预期价格，满足该价格时才成交，否则挂在订单薄里等待
)

type (
	OrderType string
	Market    string

	// PlaceOrderRequest 定义了来自客户端的下单请求体格式（JSON）。
	PlaceOrderRequest struct {
		UserID int64     `json:"UserID"` // 谁下的单
		Type   OrderType `json:"Type"`   // 订单类型：MARKET市价单 或 LIMIT限价单
		Bid    bool      `json:"Bid"`    // 交易方向：true表示买入，false表示卖出
		Size   float64   `json:"Size"`   // 数量：比如要交易 10 个 ETH
		Price  float64   `json:"Price"`  // 限价单的预期价格（市价单会被忽略）
		Market Market    `json:"Market"` // 目标交易对（目前固定传 "ETH"）
	}

	// Order 用于给客户端返回单条订单信息。
	Order struct {
		UserID    int64   `json:"UserID"`
		ID        int64   `json:"ID"`
		Price     float64 `json:"Price"`
		Size      float64 `json:"Size"`
		Bid       bool    `json:"Bid"`
		Timestamp int64   `json:"Timestamp"`
	}

	// OrderbookData 用于展示当前市场深度的全景视图。
	// 返回了所有的未成交买单、未成交卖单，以及买卖双方挂单的总量。
	OrderbookData struct {
		TotalBidVolume float64  `json:"TotalBidVolume"` // 挂在订单薄上等待买入的总量
		TotalAskVolume float64  `json:"TotalAskVolume"` // 挂在订单薄上等待卖出的总量
		Asks           []*Order `json:"Asks"`           // 卖单深度队列
		Bids           []*Order `json:"Bids"`           // 买单深度队列
	}

	// MatchedOrder 用于内部流转：表示一笔已经被撮合的（部分或全部完成的）订单摘要信息。
	MatchedOrder struct {
		UserID int64
		Price  float64
		Size   float64
		ID     int64
	}

	// APIError 是标准的 API 错误返回格式（JSON）。
	APIError struct {
		Error string `json:"Error"`
	}
)

// GetOrdersResponse 是 GET /order/:userID 的响应结构（导出供 client 包使用）。
type GetOrdersResponse struct {
	Asks []Order `json:"Asks"`
	Bids []Order `json:"Bids"`
}

// ─── GORM 模型 ───────────────────────────────────────────────────────────────

// UserModel 是持久化到 PostgreSQL 的用户账本记录。
type UserModel struct {
	gorm.Model
	UserID  int64   `gorm:"uniqueIndex;not null"`
	Balance float64 `gorm:"default:0"`
}

func (UserModel) TableName() string { return "users" }

// TradeModel 是持久化到 PostgreSQL 的成交记录。
type TradeModel struct {
	gorm.Model
	Market    string  `gorm:"not null"`
	Price     float64 `gorm:"not null"`
	Size      float64 `gorm:"not null"`
	Bid       bool    `gorm:"not null"`
	BidUserID int64   `gorm:"not null"`
	AskUserID int64   `gorm:"not null"`
}

func (TradeModel) TableName() string { return "trades" }

// ─── 内存用户（撮合层使用，不持久化私钥）────────────────────────────────────

// User 是内存中的用户对象，撮合引擎使用。
type User struct {
	ID      int64
	Balance float64
}

// ─── 启动 ──────────────────────────────────────────────────────────────────

// StartServer 初始化并启动交易所 HTTP 服务。
//
// 启动顺序：
//  1. 执行数据库迁移（MigrateUp）
//  2. 建立 GORM 连接
//  3. 初始化 Exchange，注册用户
//  4. 挂载路由，启动 HTTP
func StartServer() {
	e := echo.New()
	e.HTTPErrorHandler = httpErrorHandler

	// 从环境变量读取 DSN，默认指向本地 Docker Compose 的 PG
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://cex:cex_secret@localhost:5432/cex_db?sslmode=disable&TimeZone=Asia/Shanghai"
	}

	// 第一步：执行版本化 SQL 迁移（幂等，已是最新则跳过）
	if err := db.MigrateUp(dsn); err != nil {
		logrus.WithError(err).Fatal("migration failed")
	}

	// 第二步：建立 GORM 连接（迁移完成后再连接，确保表存在）
	gormDB, err := db.Connect(dsn)
	if err != nil {
		logrus.WithError(err).Fatal("failed to connect postgres")
	}

	ex := NewExchange(gormDB)

	// 确保系统用户存在（做市商/测试用户）
	ex.registerUser(8, 10000)
	ex.registerUser(7, 10000)
	ex.registerUser(666, 10000)

	// 从 DB 加载所有用户到内存（包含 seed 写入的模拟用户）
	if err := ex.loadUsersFromDB(); err != nil {
		logrus.WithError(err).Fatal("load users failed")
	}

	e.POST("/order", ex.handlePlaceOrder)
	e.GET("/trades/:market", ex.handleGetTrades)
	e.GET("/order/:userID", ex.handleGetOrders)
	e.GET("/book/:market", ex.handleGetBook)
	e.GET("/book/:market/bid", ex.handleGetBestBid)
	e.GET("/book/:market/ask", ex.handleGetBestAsk)
	e.DELETE("/order/:id", ex.cancelOrder)

	// 查询用户余额 GET /user/:userID/balance
	e.GET("/user/:userID/balance", ex.handleGetBalance)

	e.Start(":3000")
}

// ─── Exchange ──────────────────────────────────────────────────────────────

// Exchange 是交易所的核心结构体。
type Exchange struct {
	DB         *gorm.DB
	mu         sync.RWMutex
	Users      map[int64]*User              // 内存用户（余额镜像，撮合时使用）
	Orders     map[int64][]*orderbook.Order // UserID → 活跃订单
	orderbooks map[Market]*orderbook.Orderbook
}

func NewExchange(db *gorm.DB) *Exchange {
	obs := make(map[Market]*orderbook.Orderbook)
	obs[MarketETH] = orderbook.NewOrderbook()

	return &Exchange{
		DB:         db,
		Users:      make(map[int64]*User),
		Orders:     make(map[int64][]*orderbook.Order),
		orderbooks: obs,
	}
}

// registerUser 注册用户到内存和数据库（幂等：已存在则跳过，余额从 DB 读取）。
func (ex *Exchange) registerUser(userID int64, initialBalance float64) {
	var um UserModel
	result := ex.DB.Where("user_id = ?", userID).First(&um)
	if result.Error == gorm.ErrRecordNotFound {
		um = UserModel{UserID: userID, Balance: initialBalance}
		ex.DB.Create(&um)
	}

	ex.Users[userID] = &User{ID: userID, Balance: um.Balance}

	logrus.WithFields(logrus.Fields{
		"id":      userID,
		"balance": um.Balance,
	}).Info("registered user")
}

// loadUsersFromDB 将数据库中所有用户加载到内存 map。
// 避免每次撮合时都要查 DB。
func (ex *Exchange) loadUsersFromDB() error {
	var users []UserModel
	if err := ex.DB.Where("deleted_at IS NULL").Find(&users).Error; err != nil {
		return err
	}

	for _, u := range users {
		// 已经在 registerUser 里加载过的系统用户直接跳过（以 registerUser 的为准）
		if _, exists := ex.Users[u.UserID]; exists {
			continue
		}
		ex.Users[u.UserID] = &User{ID: u.UserID, Balance: u.Balance}
	}

	logrus.WithField("total", len(ex.Users)).Info("users loaded from DB")
	return nil
}

// ─── HTTP 错误处理 ─────────────────────────────────────────────────────────

func httpErrorHandler(err error, c echo.Context) {
	fmt.Println(err)
}

// ─── Handlers ──────────────────────────────────────────────────────────────

// handleGetTrades 处理 GET /trades/:market
// 作用：提供给前端或客户端展示"最新成交历史"，按时间倒序查询数据库里的前 200 条成交记录。
func (ex *Exchange) handleGetTrades(c echo.Context) error {
	market := c.Param("market")

	var trades []TradeModel
	if err := ex.DB.Where("market = ?", market).
		Order("created_at desc").
		Limit(200).
		Find(&trades).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, APIError{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, trades)
}

// handleGetOrders 处理 GET /order/:userID
// 作用：查询某个用户当前在订单薄里所有挂着（尚未完全成交）的活跃订单。
// 注意：只返回挂着的限价单。
func (ex *Exchange) handleGetOrders(c echo.Context) error {
	userIDStr := c.Param("userID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return err
	}

	ex.mu.RLock()         // 并发控制：保护内存 map `ex.Orders` 不在读取时被修改
	orderbookOrders := ex.Orders[int64(userID)]
	resp := &GetOrdersResponse{Asks: []Order{}, Bids: []Order{}}

	for _, o := range orderbookOrders {
		if o.Limit == nil {
			continue // 如果 Limit 为 nil 说明是市价单，或者是已经完全成交被踢出订单薄的，跳过不用返回
		}
		order := Order{
			ID:        o.ID,
			UserID:    o.UserID,
			Price:     o.Limit.Price,
			Size:      o.Size,
			Bid:       o.Bid,
			Timestamp: o.Timestamp,
		}
		if order.Bid {
			resp.Bids = append(resp.Bids, order) // 分类存放
		} else {
			resp.Asks = append(resp.Asks, order) // 分类存放
		}
	}
	ex.mu.RUnlock()

	return c.JSON(http.StatusOK, resp)
}

// handleGetBook 处理 GET /book/:market
// 作用：获取某市场完整的订单薄深度信息，前端用它来绘制红绿深度的"挂单列表"。
func (ex *Exchange) handleGetBook(c echo.Context) error {
	market := Market(c.Param("market"))
	ob, ok := ex.orderbooks[market]
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]any{"msg": "market not found"})
	}

	data := OrderbookData{
		TotalBidVolume: ob.BidTotalVolume(),
		TotalAskVolume: ob.AskTotalVolume(),
		Asks:           []*Order{},
		Bids:           []*Order{},
	}

	// 遍历卖方的价格档位（Asks 从低到高排好序）
	for _, limit := range ob.Asks() {
		for _, o := range limit.Orders {
			data.Asks = append(data.Asks, &Order{
				UserID: o.UserID, ID: o.ID,
				Price: limit.Price, Size: o.Size,
				Bid: o.Bid, Timestamp: o.Timestamp,
			})
		}
	}
	// 遍历买方的价格档位（Bids 从高到低排好序）
	for _, limit := range ob.Bids() {
		for _, o := range limit.Orders {
			data.Bids = append(data.Bids, &Order{
				UserID: o.UserID, ID: o.ID,
				Price: limit.Price, Size: o.Size,
				Bid: o.Bid, Timestamp: o.Timestamp,
			})
		}
	}

	return c.JSON(http.StatusOK, data)
}

type PriceResponse struct {
	Price float64
}

// handleGetBestBid 处理 GET /book/:market/bid
// 作用：提供当前最优买价（最高出的买价）。做市商会频繁调用拉取。
func (ex *Exchange) handleGetBestBid(c echo.Context) error {
	market := Market(c.Param("market"))
	ob := ex.orderbooks[market]
	order := Order{}
	if len(ob.Bids()) == 0 { // 如果订单薄空了，买方无挂单
		return c.JSON(http.StatusOK, order)
	}
	best := ob.Bids()[0] // 预排序过的：第0项就是最佳价位
	order.Price = best.Price
	order.UserID = best.Orders[0].UserID
	return c.JSON(http.StatusOK, order)
}

// handleGetBestAsk 处理 GET /book/:market/ask
// 作用：提供当前最优卖价（最低出的卖价）。做市商会频繁调用拉取。
func (ex *Exchange) handleGetBestAsk(c echo.Context) error {
	market := Market(c.Param("market"))
	ob := ex.orderbooks[market]
	order := Order{}
	if len(ob.Asks()) == 0 { // 如果订单薄空了，卖方无挂单
		return c.JSON(http.StatusOK, order)
	}
	best := ob.Asks()[0] // 预排序过的：第0项就是最佳价位
	order.Price = best.Price
	order.UserID = best.Orders[0].UserID
	return c.JSON(http.StatusOK, order)
}

// cancelOrder 处理 DELETE /order/:id
// 作用：用户主动撤销未成交的限价单。
func (ex *Exchange) cancelOrder(c echo.Context) error {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	ob := ex.orderbooks[MarketETH]
	order := ob.Orders[int64(id)]
	
	// 从内存撮合引擎的订单薄里移除该挂单
	ob.CancelOrder(order) 

	logrus.WithField("orderID", id).Info("order canceled")
	return c.JSON(200, map[string]any{"msg": "order deleted"})
}

// handleGetBalance 处理 GET /user/:userID/balance
// 作用：查询任意用户的最新 PostgreSQL 持久化后的余额及基本信息。
func (ex *Exchange) handleGetBalance(c echo.Context) error {
	userIDStr := c.Param("userID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, APIError{Error: "invalid userID"})
	}

	var um UserModel
	if err := ex.DB.Where("user_id = ?", userID).First(&um).Error; err != nil {
		return c.JSON(http.StatusNotFound, APIError{Error: "user not found"})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"userID":  um.UserID,
		"balance": um.Balance,
	})
}

type PlaceOrderResponse struct {
	OrderID int64 // 返回系统自动分配的该单号
}

// handlePlaceOrder 处理 POST /order 下单接口
// 作用：接收一切用户发过来的挂单或吃单请求，分发给核心业务。
func (ex *Exchange) handlePlaceOrder(c echo.Context) error {
	var req PlaceOrderRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return err
	}

	market := Market(req.Market)
	// 将传入的对象转换给订单薄底层识别的结构：方向(Bid)、数量(Size)、下单者身份(UserID)
	order := orderbook.NewOrder(req.Bid, req.Size, req.UserID)

	// 分支一：处理限价单
	if req.Type == LimitOrder {
		if err := ex.handlePlaceLimitOrder(market, req.Price, order); err != nil {
			return err
		}
	}

	// 分支二：处理市价单（不管多少钱立刻吃单，因此不需要透传 Price）
	if req.Type == MarketOrder {
		matches, _ := ex.handlePlaceMarketOrder(market, order)
		// 如果吃单成功，会产生撮合记录匹配(matches)，执行后续结算链下处理
		if err := ex.handleMatches(string(market), matches); err != nil {
			return err
		}
	}

	return c.JSON(200, &PlaceOrderResponse{OrderID: order.ID})
}

// ─── 核心业务 ──────────────────────────────────────────────────────────────

// handlePlaceLimitOrder 核心逻辑：用户下限价单
func (ex *Exchange) handlePlaceLimitOrder(market Market, price float64, order *orderbook.Order) error {
	ob := ex.orderbooks[market]
	
	// 在底层的撮合引擎记录单子位置。限价单是"Maker"行为。
	ob.PlaceLimitOrder(price, order)

	// 把该挂单记录在交易所的上层内存，并用互斥锁保证并发安全。方便 /order/:userid 接口查询挂单。
	ex.mu.Lock()
	ex.Orders[order.UserID] = append(ex.Orders[order.UserID], order)
	ex.mu.Unlock()

	return nil
}

// handlePlaceMarketOrder 核心逻辑：用户下市价单
func (ex *Exchange) handlePlaceMarketOrder(market Market, order *orderbook.Order) ([]orderbook.Match, []*MatchedOrder) {
	ob := ex.orderbooks[market]
	matches := ob.PlaceMarketOrder(order) // "Taker"过程：执行连续多次匹配，吃掉深度订单薄上的单
	matchedOrders := make([]*MatchedOrder, len(matches))

	isBid := order.Bid
	totalSizeFilled := 0.0
	sumPrice := 0.0

	// 解析出每笔子交易详情
	for i, match := range matches {
		id := match.Bid.ID
		limitUserID := match.Bid.UserID // 挂单方(Maker方)是买单的话是谁被吃
		if isBid { // 自己是市价头寸，且自己是买方，这意味着对方（挂单方Maker）就是被卖吃单
			limitUserID = match.Ask.UserID 
			id = match.Ask.ID
		}
		// 记录内部对账流转的抽象模型
		matchedOrders[i] = &MatchedOrder{
			UserID: limitUserID,
			ID:     id,
			Size:   match.SizeFilled,
			Price:  match.Price,
		}
		totalSizeFilled += match.SizeFilled
		sumPrice += match.Price  // 累加总花了多少钱
	}

	// avgPrice是本笔大市价单打出来的"滑点均价"
	avgPrice := sumPrice / float64(len(matches))
	logrus.WithFields(logrus.Fields{
		"type":     order.Type(),
		"size":     totalSizeFilled,
		"avgPrice": avgPrice, // 此处日志印出该单把系统吃出来了什么样的整体滑点
	}).Info("filled market order")

	// 清理阶段：因为订单有被整个吃掉的情况，所以需要清理事先在 ex.Orders 里存过的挂单标记，防止内存泄漏或者查挂单数据不准
	newOrderMap := make(map[int64][]*orderbook.Order) // 因为直接过滤麻烦，索性重建map清理
	ex.mu.Lock()
	for userID, orders := range ex.Orders {
		for _, o := range orders {
			if !o.IsFilled() { // 如果 o.Size != 0.0，说明没全填完或者只部分吃掉了。
				newOrderMap[userID] = append(newOrderMap[userID], o)
			}
		}
	}
	ex.Orders = newOrderMap // 将干净的、尚未撮合彻底的订单集合重新绑回
	ex.mu.Unlock()

	return matches, matchedOrders
}

// handleMatches 处理撮合后的链下账本结算，并持久化成交记录到 PostgreSQL。
//
// 流程：
//  1. Ask（卖方）余额 += SizeFilled（卖出收款）
//  2. Bid（买方）余额 -= SizeFilled（买入付款）
//  3. 写 TradeModel 到 DB
//  4. 更新 UserModel 余额到 DB
func (ex *Exchange) handleMatches(market string, matches []orderbook.Match) error {
	return ex.DB.Transaction(func(tx *gorm.DB) error {
		for _, match := range matches {
			askUser, ok := ex.Users[match.Ask.UserID]
			if !ok {
				return fmt.Errorf("ask user not found: %d", match.Ask.UserID)
			}
			bidUser, ok := ex.Users[match.Bid.UserID]
			if !ok {
				return fmt.Errorf("bid user not found: %d", match.Bid.UserID)
			}

			amount := match.SizeFilled

			// 内存余额更新
			askUser.Balance += amount // 卖方收到钱
			bidUser.Balance -= amount // 买方付出钱

			// DB：更新卖方余额
			if err := tx.Model(&UserModel{}).
				Where("user_id = ?", askUser.ID).
				Update("balance", askUser.Balance).Error; err != nil {
				return err
			}
			// DB：更新买方余额
			if err := tx.Model(&UserModel{}).
				Where("user_id = ?", bidUser.ID).
				Update("balance", bidUser.Balance).Error; err != nil {
				return err
			}

			// DB：写成交记录
			trade := TradeModel{
				Market:    market,
				Price:     match.Price,
				Size:      match.SizeFilled,
				Bid:       true,
				BidUserID: match.Bid.UserID,
				AskUserID: match.Ask.UserID,
			}
			if err := tx.Create(&trade).Error; err != nil {
				return err
			}

			logrus.WithFields(logrus.Fields{
				"market":    market,
				"askUser":   askUser.ID,
				"bidUser":   bidUser.ID,
				"amount":    amount,
				"price":     match.Price,
				"newAskBal": askUser.Balance,
				"newBidBal": bidUser.Balance,
			}).Info("internal settlement")
		}
		return nil
	})
}

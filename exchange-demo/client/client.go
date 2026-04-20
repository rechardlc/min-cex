// Package client 提供了与交易所 HTTP API 交互的客户端 SDK。
//
// 这个包将底层的 HTTP 请求/响应封装为类型安全的 Go 函数调用，
// 使得其他模块（如做市商 mm 包、main 中的模拟交易者）无需关心 HTTP 细节。
//
// 设计模式：这是一个典型的 "API Client SDK" 模式，在微服务架构中很常见。
// 前端同学可以类比为前端项目中的 api.js / request.ts：
//   前端: api.getUsers() → axios.get('/users')
//   这里: client.GetBestBid() → http.Get('/book/ETH/bid')
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/anthdm/crypto-exchange/orderbook"
	"github.com/anthdm/crypto-exchange/server"
)

// Endpoint 是交易所服务端的基础 URL。
// 在生产环境中通常从配置文件/环境变量读取，这里硬编码为本地地址。
const Endpoint = "http://localhost:3000"

// PlaceOrderParams 是下单时需要提供的参数。
// 这是客户端侧的参数结构，与 server.PlaceOrderRequest 不同：
//   - 客户端不需要指定 OrderType（由调用 PlaceLimitOrder 还是 PlaceMarketOrder 决定）
//   - 客户端不需要指定 Market（默认 ETH）
type PlaceOrderParams struct {
	UserID int64   // 下单用户 ID
	Bid    bool    // true=买入, false=卖出
	Price  float64 // 挂单价格（仅限价单需要，市价单忽略此字段）
	Size   float64 // 下单数量
}

// Client 是交易所的 HTTP 客户端。
// 通过嵌入 *http.Client，继承了标准库 HTTP 客户端的所有方法（如 Do, Get, Post 等）。
//
// 嵌入（Embedding）是 Go 的组合（composition）特性，类似于前端中的 extends：
//   class Client extends http.Client { ... }
// 但 Go 不是继承，而是将 http.Client 的方法"提升"到 Client 上直接使用。
type Client struct {
	*http.Client
}

// NewClient 创建一个新的交易所客户端实例。
// 使用 http.DefaultClient 作为底层 HTTP 客户端（带默认的超时和连接池设置）。
func NewClient() *Client {
	return &Client{
		Client: http.DefaultClient,
	}
}

// GetTrades 获取指定市场的历史成交记录。
// GET /trades/:market
//
// 返回 Trade 切片，每个 Trade 包含成交价格、数量、时间等信息。
func (c *Client) GetTrades(market string) ([]*orderbook.Trade, error) {
	e := fmt.Sprintf("%s/trades/%s", Endpoint, market)
	req, err := http.NewRequest(http.MethodGet, e, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req) // 发送 HTTP 请求
	if err != nil {
		return nil, err
	}

	trades := []*orderbook.Trade{}
	// json.NewDecoder 从 resp.Body（io.Reader）直接流式解码 JSON，
	// 比先 ioutil.ReadAll 再 json.Unmarshal 更内存友好。
	if err := json.NewDecoder(resp.Body).Decode(&trades); err != nil {
		return nil, err
	}

	return trades, nil
}

// GetOrders 获取指定用户的所有活跃订单。
// GET /order/:userID
//
// 返回 GetOrdersResponse，包含该用户的买单列表和卖单列表。
func (c *Client) GetOrders(userID int64) (*server.GetOrdersResponse, error) {
	e := fmt.Sprintf("%s/order/%d", Endpoint, userID)
	req, err := http.NewRequest(http.MethodGet, e, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	orders := server.GetOrdersResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&orders); err != nil {
		return nil, err
	}

	return &orders, nil
}

// PlaceMarketOrder 提交一个市价单。
// POST /order (Type=MARKET)
//
// 市价单会立即按当前订单薄中的最优价格成交。
// 不需要指定 Price，因为市价单的核心就是"不管什么价格，立即成交"。
//
// 注意事项：
//   - 如果订单薄中没有足够的流动性（挂单量不足），服务端会 panic
//   - Market 硬编码为 MarketETH（因为当前只有一个交易对）
func (c *Client) PlaceMarketOrder(p *PlaceOrderParams) (*server.PlaceOrderResponse, error) {
	// 将客户端参数转换为服务端需要的请求格式
	params := &server.PlaceOrderRequest{
		UserID: p.UserID,
		Type:   server.MarketOrder, // 标记为市价单
		Bid:    p.Bid,
		Size:   p.Size,
		Market: server.MarketETH,   // 默认 ETH 市场
	}

	body, err := json.Marshal(params) // 序列化为 JSON
	if err != nil {
		return nil, err
	}

	e := Endpoint + "/order"
	req, err := http.NewRequest(http.MethodPost, e, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	placeOrderResponse := &server.PlaceOrderResponse{}
	if err := json.NewDecoder(resp.Body).Decode(placeOrderResponse); err != nil {
		return nil, err
	}

	return placeOrderResponse, nil
}

// GetBestAsk 获取当前最优卖价（最低的卖单价格）。
// GET /book/ETH/ask
//
// 做市商策略用这个价格来决定自己挂卖单的位置。
// 例如：如果最优卖价是 1050，做市商可能会在 1045 挂一个卖单来"抢位"。
func (c *Client) GetBestAsk() (*server.Order, error) {
	e := fmt.Sprintf("%s/book/ETH/ask", Endpoint)
	req, err := http.NewRequest(http.MethodGet, e, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	order := &server.Order{}
	if err := json.NewDecoder(resp.Body).Decode(order); err != nil {
		return nil, err
	}

	return order, err
}

// GetBestBid 获取当前最优买价（最高的买单价格）。
// GET /book/ETH/bid
//
// 做市商策略用这个价格来决定自己挂买单的位置。
// 例如：如果最优买价是 950，做市商可能会在 955 挂一个买单。
func (c *Client) GetBestBid() (*server.Order, error) {
	e := fmt.Sprintf("%s/book/ETH/bid", Endpoint)
	req, err := http.NewRequest(http.MethodGet, e, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	order := &server.Order{}
	if err := json.NewDecoder(resp.Body).Decode(order); err != nil {
		return nil, err
	}

	return order, err
}

// CancelOrder 取消一个挂单。
// DELETE /order/:id
//
// 注意：取消订单只影响限价单（因为市价单是立即成交的，不会挂在订单薄上）。
func (c *Client) CancelOrder(orderID int64) error {
	e := fmt.Sprintf("%s/order/%d", Endpoint, orderID)
	req, err := http.NewRequest(http.MethodDelete, e, nil)
	if err != nil {
		return err
	}

	_, err = c.Do(req)
	if err != nil {
		return err
	}

	return nil
}

// PlaceLimitOrder 提交一个限价单。
// POST /order (Type=LIMIT)
//
// 限价单不会立即成交，而是"挂"在订单薄的某个价格档位上，
// 等待有人下市价单来"吃掉"它。
//
// 例如：挂一个买入限价单 Price=950, Size=10
//   → 意思是"我愿意以 950 的价格买入 10 个 ETH"
//   → 当有人下卖出市价单时，如果 950 是当前最高买价，就会成交
func (c *Client) PlaceLimitOrder(p *PlaceOrderParams) (*server.PlaceOrderResponse, error) {
	// 限价单必须指定数量
	if p.Size == 0.0 {
		return nil, fmt.Errorf("size cannot be 0 when placing a limit order")
	}

	params := &server.PlaceOrderRequest{
		UserID: p.UserID,
		Type:   server.LimitOrder, // 标记为限价单
		Bid:    p.Bid,
		Size:   p.Size,
		Price:  p.Price,           // 限价单必须指定价格
		Market: server.MarketETH,
	}

	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	e := Endpoint + "/order"
	req, err := http.NewRequest(http.MethodPost, e, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	placeOrderResponse := &server.PlaceOrderResponse{}
	if err := json.NewDecoder(resp.Body).Decode(placeOrderResponse); err != nil {
		return nil, err
	}

	return placeOrderResponse, nil
}

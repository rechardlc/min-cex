// Package main 是整个加密货币交易所演示程序的入口。
//
// ====== 系统架构概览 ======
//
// 这个项目模拟了一个最简化的中心化加密货币交易所（CEX），包含以下组件：
//
//   ┌─────────────────────────────────────────────────────┐
//   │                     main.go（本文件）                │
//   │  1. 启动交易所服务器（server）                        │
//   │  2. 启动做市商（market maker）                       │
//   │  3. 启动模拟交易者（marketOrderPlacer）               │
//   └──────┬────────────────┬────────────────┬────────────┘
//          ↓                ↓                ↓
//   ┌──────────┐   ┌───────────────┐   ┌──────────────┐
//   │  server  │   │ mm (做市商)   │   │ 模拟交易者    │
//   │ (HTTP API│   │ 提供流动性    │   │ 随机下市价单  │
//   │  + 撮合) │   │ 赚取价差      │   │ 模拟真实交易  │
//   └──────────┘   └───────────────┘   └──────────────┘
//         ↓
//   ┌──────────┐
//   │orderbook │
//   │ 订单薄   │
//   │ 撮合引擎 │
//   └──────────┘
//
// ====== 启动顺序 ======
//
//   1. server.StartServer()    - 在 goroutine 中启动 HTTP 服务
//   2. time.Sleep(1s)          - 等待服务器就绪
//   3. 创建 Client             - 用于与服务器通信
//   4. 创建 MarketMaker        - 做市商，在订单薄中挂买卖单
//   5. maker.Start()           - 开始做市循环（通过 goroutine 异步执行）
//   6. time.Sleep(2s)          - 等待做市商完成初始播种
//   7. marketOrderPlacer()     - 在 goroutine 中模拟随机市价单交易
//   8. select{}                - 阻塞主 goroutine，防止程序退出
package main

import (
	"math/rand"
	"time"

	"github.com/anthdm/crypto-exchange/client"
	"github.com/anthdm/crypto-exchange/mm"
	"github.com/anthdm/crypto-exchange/server"
)

func main() {
	// 第一步：在后台 goroutine 中启动交易所 HTTP 服务器。
	// 使用 go 关键字启动一个新的 goroutine，使服务器在后台运行。
	// 如果不用 go，main 函数会阻塞在 StartServer() 中（因为 e.Start 是阻塞的），
	// 后面的代码永远不会执行。
	go server.StartServer()

	// 等待 1 秒，确保 HTTP 服务器已完全启动并开始监听端口。
	// 这是一种简单但不够健壮的做法，生产环境应使用健康检查（health check）。
	time.Sleep(1 * time.Second)

	// 第二步：创建一个 HTTP 客户端，用于与交易所 API 通信。
	// 做市商和模拟交易者都通过这个客户端来下单。
	c := client.NewClient()

	// 第三步：配置并创建做市商。
	cfg := mm.Config{
		UserID:         8,              // 做市商使用 UserID=8（在 server 中已注册）
		OrderSize:      10,             // 每次挂单 10 个 ETH
		MinSpread:      20,             // 当买卖价差 > 20 时才挂单（否则利润太低）
		MakeInterval:   1 * time.Second, // 每 1 秒执行一次做市策略
		SeedOffset:     40,             // 播种时买卖单距参考价格的偏移量
		ExchangeClient: c,              // 使用上面创建的 HTTP 客户端
		PriceOffset:    10,             // 挂单时相对最优价的偏移量
	}
	maker := mm.NewMakerMaker(cfg)

	// 启动做市商（内部通过 goroutine 异步运行做市循环）
	maker.Start()

	// 等待 2 秒，让做市商完成"播种"——在空的订单薄中放入初始的买卖单。
	// 如果不等待，下面的市价单可能因为没有流动性而 panic。
	time.Sleep(2 * time.Second)

	// 第四步：启动一个模拟交易者，在后台随机下市价单。
	// 这模拟了真实市场中的普通用户交易行为。
	go marketOrderPlacer(c)

	// select{} 是 Go 中让主 goroutine 永久阻塞的惯用写法。
	// 如果 main 函数返回，所有 goroutine 都会被强制终止。
	// 所以需要 select{} 来保持 main 活着。
	// 等效于 for {} 但更符合 Go 惯例且不消耗 CPU。
	select {}
}

// marketOrderPlacer 模拟多个普通交易者，每 500ms 随机选一个用户下一个市价单。
//
// 模拟用户池：UserID 100-199（由 make seed 写入数据库）
// 策略：70% 概率卖出，30% 概率买入（模拟轻微卖压）
func marketOrderPlacer(c *client.Client) {
	ticker := time.NewTicker(500 * time.Millisecond)

	// 模拟用户池：seed 写入的 UserID 100-199
	const (
		userPoolStart = 100
		userPoolSize  = 100
	)

	for {
		randint := rand.Intn(10)
		bid := true
		if randint < 7 {
			bid = false
		}

		// 从用户池中随机选一个用户
		userID := int64(userPoolStart + rand.Intn(userPoolSize))

		order := client.PlaceOrderParams{
			UserID: userID,
			Bid:    bid,
			Size:   1,
		}

		_, err := c.PlaceMarketOrder(&order)
		if err != nil {
			panic(err)
		}

		<-ticker.C
	}
}


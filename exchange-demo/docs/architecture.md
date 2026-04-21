# min-cex 交易所架构与调用链路

本项目模拟了一个最简化的中心化加密货币交易所（CEX）。文档包含了核心的系统架构图和调用链路图。

## 1. 系统架构图

系统主要由三大模块组成：提供流动性的做市商（MM）、模拟真实用户的市价单交易者（Trader），以及核心的后端交易所服务（API层 + 撮合引擎 + 数据库）。

```mermaid
graph TD
    subgraph "min-cex Exchange Demo 运行环境"
        Main("main.go (入口启动程序)")
        
        subgraph "客户端模拟"
            MM["做市商 (Market Maker)<br>定时监控并挂限价单"]
            Trader["模拟交易者 (Trader)<br>每500ms下市价单"]
            ClientPkg["Client SDK<br>(HTTP 封装)"]
        end
        
        subgraph "后端核心服务 (Server)"
            API["HTTP API (Server)<br>路由、请求校验、业务逻辑"]
            OB["Orderbook (撮合引擎)<br>内存价格优先-时间优先撮合"]
            DB[("PostgreSQL<br>(用户/资产/订单/交易数据持久化)")]
        end

        Main -->|1. 异步启动| API
        Main -->|2. 实例化并启动| MM
        Main -->|3. 异步启动| Trader
        
        MM --> |调用| ClientPkg
        Trader --> |调用| ClientPkg
        ClientPkg -->|HTTP API 通信| API
        
        API -->|处理订单| OB
        API -->|数据存取| DB
        OB -->|撮合结果/清算落库| DB
    end
```

## 2. 调用链路 / 启动与生命周期图

该时序图展示了程序执行 `main()` 时的启动顺序，以及做市商和普通用户的并发模拟交互流：

```mermaid
sequenceDiagram
    participant Main as main.go (入口)
    participant API as Server (交易所主服务)
    participant Client as Client (HTTP SDK)
    participant MM as mm (做市商 Goroutine)
    participant Trader as trader (模拟用户 Goroutine)
    participant OB as Orderbook (内存撮合引擎)

    Main->>API: 1. go StartServer() 后台启动 HTTP 服务
    activate API
    Main->>Main: 等待 1s 确保服务器就绪
    Main->>Client: 2. NewClient() 初始化客户端 SDK
    
    Main->>MM: 3. maker.Start() 初始化并启动做市商
    activate MM
    loop 定时做市 (1s 间隔)
        MM->>Client: PlaceLimitOrder(提供流动性)
        Client->>API: HTTP POST /order (限价单)
        API->>OB: PlaceOrder() 挂入买卖盘
        OB-->>API: 返回撮合记录 (无成交或部分成交)
        API-->>Client: 200 OK
    end
    
    Main->>Main: 等待 2s 确保做市商完成初次"播种"
    
    Main->>Trader: 4. go marketOrderPlacer() 启动模拟交易
    activate Trader
    loop 定时模拟交易 (500ms 间隔)
        Trader->>Trader: 从随机池(ID:100~199)随机选买/卖单
        Trader->>Client: PlaceMarketOrder(吃单操作)
        Client->>API: HTTP POST /order (市价单)
        API->>OB: PlaceOrder()
        OB-->>API: 撮合成交 (Match trades)
        API-->>Client: 200 OK (订单完成)
    end
    
    Main->>Main: 5. select{} 挂起主线程，维持各个 Goroutine 运行
```

## 简要说明
* **入口程序**：`main.go` 充当调度者，它依次拉起 Server、建立 Client，接着分别挂载做市商(`mm`)和散户交易模块(`marketOrderPlacer`)。
* **数据流向**：做市商负责给原本空空的撮合引擎（`Orderbook`）建立深度池，随后散户进来根据设定的比例通过市价单“吃掉”这些挂单，从而实现整个闭环 Demo。撮合引擎处理好的明细再经 Server 同步持久化。

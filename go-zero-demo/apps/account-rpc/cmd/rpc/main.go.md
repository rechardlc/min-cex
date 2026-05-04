# account-rpc main.go 做什么用

这是 RPC 服务启动入口。

## 真实项目里一般会做什么

- 加载 RPC 配置
- 初始化 service context
- 注册 gRPC / zrpc 服务
- 启动监听

你可以把它理解成 RPC 版本的服务入口。


# account-rpc servicecontext.go 做什么用

和 API 服务一样，这里是 RPC 服务的依赖注入中心。

## 常见内容

- 配置
- 数据库访问对象
- 缓存客户端
- 下游 RPC 客户端

学习时你可以重点对比 `gateway` 和 `account-rpc` 的 `svc`，理解 go-zero 在不同服务形态下的共性。


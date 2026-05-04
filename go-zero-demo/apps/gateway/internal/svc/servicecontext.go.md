# servicecontext.go 做什么用

`ServiceContext` 是 go-zero 项目里的依赖注入中心。

## 一般会放什么

- 配置对象
- 数据库连接
- Redis 客户端
- RPC 客户端
- MQ Producer / Consumer

## 为什么它很关键

因为它把“依赖初始化”和“业务逻辑”分开了。`logic` 不需要自己 new 一堆对象，只需要从 `svc` 拿依赖。


# go.work 做什么用

这个文件用于把多个 Go 模块组合到一个工作区里，方便本地联调。

## 在这个学习骨架里的作用

- 让 `gateway` 和 `account-rpc` 可以一起开发
- 以后新增 `trade-rpc`、`wallet-rpc` 时也可以继续挂进来

## 学习重点

- 理解 Go workspace 和单模块的区别
- 理解为什么微服务项目常常需要 `go.work`


## RPC
- net/rpc没有router概念，是直接调用方法的，通过restful api需要经过hander层，rpc直接到service层
- 标准包的net/rpc方法签名规范是硬性规定的
```golang
    type Calc struct {}
    // 这个是硬性规定的：调用方需要调用Calc.Add方法
    func (c *Calc) Add(args *Args, reply *Reply) error {}
``


### 真实的环境中
- 微服务是动态IP
- 存在上下文超时控制
- 浏览器调用只能通过新增网关去调用，RPC是服务与服务的调用


### rpc核心的缺陷
- 缺乏跨语言的支持
- 缺乏错误处理机制
- 缺乏服务发现机制
- 没有清晰的接口语义
- 没有可拔插中间件支持


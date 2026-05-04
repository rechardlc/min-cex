## RPC
- net/rpc没有router概念，是直接调用方法的，通过restful api需要经过hander层，rpc直接到service层
- 标准包的net/rpc方法签名规范是硬性规定的
```golang
    type Calc struct {}
    // 这个是硬性规定的：调用方需要调用Calc.Add方法
    func (c *Calc) Add(args *Args, reply *Reply) error {}
``
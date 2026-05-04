package main

import (
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
)

type Args struct {
	A,B int
}
type Reply struct {
	Sum int
}


type Calculator struct{}

func (c *Calculator) Add(args Args, reply *Reply) error {
	reply.Sum = args.A + args.B
	log.Printf("收到请求: %d + %d = %d", args.A, args.B, reply.Sum)
	return nil
}
func main() {
	// new: 创建一个Calculator实例
	calc := new(Calculator)
	// Register: 注册Calculator实例
	rpc.Register(calc)
	// Listen: 监听端口
	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Printf("监听端口失败: %v", err)
		return
	}
	log.Printf("监听端口成功: %s", listener.Addr().String())
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		// 使用jsonrpc.ServeConn()来处理请求
		go jsonrpc.ServeConn(conn)
	}
}

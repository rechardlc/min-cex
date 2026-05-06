

package main

import (
	"log"
	"net/rpc/jsonrpc"
	"fmt"
)

type Args struct {
	A,B int
}
type Reply struct {
	Sum int
}

func main() {
	// Dial: 连接到RPC服务
	conn, err := jsonrpc.Dial("tcp", "localhost:1234")
	if err  != nil {
		log.Printf("连接失败: %v", err)
		return
	}
	// defer: 延迟关闭连接
	defer conn.Close()
	// 创建一个Args实例
	args := Args{
		A: 1,
		B: 1,
	}
	// 创建一个Reply实例
	var reply Reply
	// Call: 调用RPC方法
	err = conn.Call("Calculator.Add", args, &reply)
	if err != nil {
		log.Printf("调用失败: %v", err)
		return
	}
	log.Printf("RPC调用结果: %d", reply.Sum)
	fmt.Println("RPC调用结果: ", reply.Sum)
}
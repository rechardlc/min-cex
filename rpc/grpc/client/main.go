package main

import (
	"context"
	"fmt"
	"grpc/pb_user"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// WithTransportCredentials: 加密，这里是本地环境所以是明文传输
	start := time.Now()
	conn, err := grpc.NewClient("localhost:1234", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if  err != nil {
		log.Fatal("链接失败")
	}
	// defer: 延迟关闭连接
	defer conn.Close()
	// 创建一个grpc客户端
	client := pb_user.NewUserServiceClient(conn)
	// 创建一个上下文
	ctx, cancle := context.WithTimeout(context.Background(), time.Second * 100)
	defer cancle()
	// 创建一个请求
	result, err := client.GetUser(ctx, &pb_user.GetUserRequest{
		UserId: "a87",
		A: 1,
		B: 2,
	})
	if err != nil {
		log.Fatal("请求失败")
	}
	// 打印结果
	fmt.Println(result)
	// 打印时间: time.Since(start) 返回的是一个时间差
	elapsed := time.Since(start)
	fmt.Println(elapsed.Seconds()) // 打印时间差
}
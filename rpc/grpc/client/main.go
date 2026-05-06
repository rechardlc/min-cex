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
	defer conn.Close()
	client := pb_user.NewUserServiceClient(conn)
	ctx, cancle := context.WithTimeout(context.Background(), time.Second * 1)
	defer cancle()
	result, err := client.GetUser(ctx, &pb_user.GetUserRequest{
		UserId: "a87",
	})
	time.Sleep(2 * time.Second)
	if err != nil {
		log.Fatal("请求失败")
	}
	fmt.Println(result)
	elapsed := time.Since(start)
	fmt.Println(elapsed.Seconds())
}
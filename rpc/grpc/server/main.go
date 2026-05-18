package main

import (
	"context"
	"errors"
	"fmt"
	"grpc/pb_user"
	"log"
	"net"
	"strings"

	"google.golang.org/grpc" // 引入grpc包
)

type UserService struct {
	pb_user.UnimplementedUserServiceServer
}

// 实现UserService接口
func (s *UserService) GetUser(ctx context.Context, req *pb_user.GetUserRequest) (*pb_user.GetUserResponse, error) {
	fmt.Println("服务进来了")
	if id := req.UserId; id != "" {
		// 使用strings.Builder来拼接字符串
		var builder strings.Builder
		builder.WriteString("user_")
		builder.WriteString(id)
		result := builder.String()
		msg := fmt.Sprintf("请求的id: %s", id)
		fmt.Println(msg)
		return &pb_user.GetUserResponse{
			UserId:   id,
			Nickname: result,
			Sum: sum(req.A, req.B),
		}, nil
	}
	return nil, errors.New("user id is required")	
}

func main() {
	// 监听TCP端口
	listen, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal("监听失败:", err)
	}
	// 创建一个grpc服务器
	server := grpc.NewServer()
	// new(UserService) 与 &UserService{}是一个意思，golang社区更加偏向第二种写法
	pb_user.RegisterUserServiceServer(server, &UserService{})
	log.Println("服务启动")
	server.Serve(listen)
}

func sum(a float64, b float64) float64{
	return a + b
}
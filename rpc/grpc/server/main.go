package main

import (
	"context"
	"errors"
	"fmt"
	"grpc/pb_user"
	"log"
	"net"
	"strings"

	"google.golang.org/grpc"
)

type UserService struct {
	pb_user.UnimplementedUserServiceServer
}

// 实现UserService接口
func (s *UserService) GetUser(ctx context.Context, req *pb_user.GetUserRequest) (*pb_user.GetUserResponse, error) {
	fmt.Println("服务进来了")
	if id := req.UserId; id != "" {
		var builder strings.Builder
		builder.WriteString("user_")
		builder.WriteString(id)
		result := builder.String()
		msg := fmt.Sprintf("请求的id: %s", id)
		fmt.Println(msg)
		return &pb_user.GetUserResponse{
			UserId:   id,
			Nickname: result,
		}, nil
	}
	return nil, errors.New("user id is required")	
}

func main() {
	listen, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal("监听失败:", err)
	}
	server := grpc.NewServer()
	// new(UserService) 与 &UserService{}是一个意思，golang社区更加偏向第二种写法
	pb_user.RegisterUserServiceServer(server, &UserService{})
	log.Println("服务启动")
	server.Serve(listen)
}
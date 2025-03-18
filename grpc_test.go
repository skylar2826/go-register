package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	gprc2 "micro/grpc"
	"micro/proto/gen"
	"net"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	userService := &gprc2.UserServiceServer{}
	server := grpc.NewServer()
	gen.RegisterUserServiceServer(server, userService)

	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		t.Fatal(err)
	}
	err = server.Serve(l)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("server end")
}

func TestClient(t *testing.T) {
	cc, err := grpc.Dial("my-scheme:///localhost:8080", grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	defer cc.Close()
	client := gen.NewUserServiceClient(cc)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	req := &gen.GetByIdReq{
		Id: 2,
	}
	var resp *gen.GetByIdResp
	resp, err = client.GetById(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(resp)
}

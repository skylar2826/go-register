package main

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	grpc2 "micro/grpc"
	"micro/proto/gen"
	"micro/register"
	"micro/register/etcd"
	"testing"
	"time"
)

func TestRegisterServer(t *testing.T) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"127.0.0.1:2379"},
	})
	if err != nil {
		t.Fatal(err)
		return
	}
	defer client.Close()
	var r register.Register
	r, err = etcd.NewRegister(client)
	if err != nil {
		t.Fatal(err)
		return
	}
	s := NewServer("user_service", WithServerRegister(r, time.Second*10))

	userServiceServer := &grpc2.UserServiceServer{}
	gen.RegisterUserServiceServer(s, userServiceServer)

	err = s.Start("127.0.0.1:8080")
	if err != nil {
		t.Fatal(err)
		return
	}
}

func TestRegisterClient(t *testing.T) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2379"},
	})
	var r register.Register
	r, err = etcd.NewRegister(client)
	if err != nil {
		fmt.Println(err)
		return
	}
	c := NewClient(WithInSecure(), WithClientRegister(r))
	cc, err := c.Dial("user_service", time.Minute)
	if err != nil {
		t.Fatal(err)
		return
	}
	defer cc.Close()
	userServiceClient := gen.NewUserServiceClient(cc)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	req := &gen.GetByIdReq{}
	var resp *gen.GetByIdResp
	resp, err = userServiceClient.GetById(ctx, req)
	if err != nil {
		t.Fatal(err)
		return
	}
	fmt.Printf("resp:%v\n", resp)
	cancel()
}

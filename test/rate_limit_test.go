package test

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	grpc2 "micro/grpc"
	"micro/load_balancer/round_robin"
	"micro/proto/gen"
	redis2 "micro/rate_limit/redis"
	"micro/register"
	"micro/register/etcd"
	"sync"
	"testing"
	"time"
)

func TestRateLimitServer(t *testing.T) {
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

	//tokenBucketInterceptor := rate_limit.NewTokenBucket(1, time.Second)
	//config := grpc2.GrpcServerConfig{Interceptor: tokenBucketInterceptor.Build()}

	//slideWindow := rate_limit.NewSlideWindow(2, time.Second)
	//config := grpc2.GrpcServerConfig{Interceptor: slideWindow.Build()}
	redisClient := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	//redisFixWindow := redis2.NewFixWindow(redisClient, time.Second, 3, "user_service")
	//config := grpc2.GrpcServerConfig{Interceptor: redisFixWindow.Build()}

	redisSlideWinow := redis2.NewSlideWindow(redisClient, time.Second, 3, "user_service")
	config := grpc2.GrpcServerConfig{Interceptor: redisSlideWinow.Build()}

	s := grpc2.NewServer("user_service", config, grpc2.WithServerRegister(r, time.Second*10))

	userServiceServer := &grpc2.UserServiceServer{}
	gen.RegisterUserServiceServer(s, userServiceServer)

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		err = s.Start("127.0.0.1:8080", &grpc2.ServerConfig{Weight: 10, Group: "A"})
		wg.Done()
		if err != nil {
			t.Fatal(err)
			return
		}

	}()

	go func() {
		err = s.Start("127.0.0.1:8081", &grpc2.ServerConfig{Weight: 12, Group: "A"})
		wg.Done()
		if err != nil {
			t.Fatal(err)
			return
		}

	}()

	go func() {
		err = s.Start("127.0.0.1:8082", &grpc2.ServerConfig{Weight: 13, Group: "B"})
		wg.Done()
		if err != nil {
			t.Fatal(err)
			return
		}
	}()
	wg.Wait()
}

func TestRateLimitClient(t *testing.T) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2379"},
	})
	var r register.Register
	r, err = etcd.NewRegister(client)
	if err != nil {
		fmt.Println(err)
		return
	}

	builder := base.NewBalancerBuilder("TEST_DEMO_ROUND_ROBIN", &round_robin.BalancerBuilder{}, base.Config{HealthCheck: true})
	balancer.Register(builder)

	c := grpc2.NewClient(grpc2.WithInSecure(), grpc2.WithClientRegister(r), grpc2.WithBalancer("TEST_DEMO_ROUND_ROBIN"))
	cc, err := c.Dial("user_service", time.Minute)
	if err != nil {
		t.Fatal(err)
		return
	}
	defer cc.Close()
	userServiceClient := gen.NewUserServiceClient(cc)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	ctx = context.WithValue(ctx, "group", "A")
	for i := 0; i < 4; i++ {
		//if i == 3 {
		//	time.Sleep(2 * time.Second)
		//}
		req := &gen.GetByIdReq{}
		var resp *gen.GetByIdResp
		resp, err = userServiceClient.GetById(ctx, req)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("resp:%v\n", resp)
	}
	cancel()

	select {
	case <-ctx.Done():
		fmt.Println(ctx.Err())
		return
	}
}

package test

//
//func TestRoundRobinServer(t *testing.T) {
//	client, err := clientv3.New(clientv3.Config{
//		Endpoints: []string{"127.0.0.1:2379"},
//	})
//	if err != nil {
//		t.Fatal(err)
//		return
//	}
//	defer client.Close()
//	var r register.Register
//	r, err = etcd.NewRegister(client)
//	if err != nil {
//		t.Fatal(err)
//		return
//	}
//	s := main.NewServer("user_service", main.WithServerRegister(r, time.Second*10))
//
//	userServiceServer := &grpc2.UserServiceServer{}
//	gen.RegisterUserServiceServer(s, userServiceServer)
//
//	var wg sync.WaitGroup
//	wg.Add(3)
//	go func() {
//		err = s.Start("127.0.0.1:8080", &main.ServerConfig{Weight: 10})
//		wg.Done()
//		if err != nil {
//			t.Fatal(err)
//			return
//		}
//
//	}()
//
//	go func() {
//		err = s.Start("127.0.0.1:8081", &main.ServerConfig{Weight: 12})
//		wg.Done()
//		if err != nil {
//			t.Fatal(err)
//			return
//		}
//
//	}()
//
//	go func() {
//		err = s.Start("127.0.0.1:8082", &main.ServerConfig{Weight: 13})
//		wg.Done()
//		if err != nil {
//			t.Fatal(err)
//			return
//		}
//	}()
//	wg.Wait()
//}
//
//func TestRoundRobinClient(t *testing.T) {
//	client, err := clientv3.New(clientv3.Config{
//		Endpoints: []string{"localhost:2379"},
//	})
//	var r register.Register
//	r, err = etcd.NewRegister(client)
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//	builder := base.NewBalancerBuilder("TEST_DEMO_WEIGHT_ROUND_ROBIN", &round_robin.WeightBalancerBuilder{}, base.Config{HealthCheck: true})
//	//builder := base.NewBalancerBuilder("TEST_DEMO_ROUND_ROBIN", &round_robin.BalancerBuilder{}, base.Config{HealthCheck: true})
//	balancer.Register(builder)
//	c := main.NewClient(main.WithInSecure(), main.WithClientRegister(r), main.WithBalancer("TEST_DEMO_WEIGHT_ROUND_ROBIN"))
//	//c := NewClient(WithInSecure(), WithClientRegister(r), WithBalancer("TEST_DEMO_ROUND_ROBIN"))
//	cc, err := c.Dial("user_service", time.Minute)
//	if err != nil {
//		t.Fatal(err)
//		return
//	}
//	defer cc.Close()
//	userServiceClient := gen.NewUserServiceClient(cc)
//	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
//	for i := 0; i < 4; i++ {
//		req := &gen.GetByIdReq{}
//		var resp *gen.GetByIdResp
//		resp, err = userServiceClient.GetById(ctx, req)
//		if err != nil {
//			t.Fatal(err)
//			return
//		}
//		fmt.Printf("resp:%v\n", resp)
//	}
//	cancel()
//}

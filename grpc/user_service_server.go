package grpc

import (
	"context"
	"fmt"
	"micro/proto/gen"
)

type UserServiceServer struct {
	gen.UnimplementedUserServiceServer
}

func (s *UserServiceServer) GetById(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error) {
	fmt.Println(req)
	return &gen.GetByIdResp{
		User: &gen.User{
			Id:   1,
			Name: "zly",
		},
	}, nil
}

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
	fmt.Println("user_service调用: ", req)
	return &gen.GetByIdResp{
		User: &gen.User{
			Id:   req.Id,
			Name: "zly",
		},
	}, nil
}

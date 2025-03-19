package route_strategy

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/resolver"
)

type Filter func(info balancer.PickInfo, address resolver.Address) bool

type GroupFilterBuilder struct {
}

func (g GroupFilterBuilder) Build() Filter {
	return func(info balancer.PickInfo, address resolver.Address) bool {
		clientGroup := info.Ctx.Value("group").(string)
		serverGroup := address.Attributes.Value("group").(string)

		return clientGroup == serverGroup
	}
}

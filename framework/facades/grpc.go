package facades

import "github.com/zhoudm1743/go-fast/framework/contracts"

// GRPC 返回 gRPC Server 服务实例。
// 使用方式与 facades.Route() 完全一致。
func GRPC() contracts.GRPCServer {
	return App().MustMake("grpc").(contracts.GRPCServer)
}

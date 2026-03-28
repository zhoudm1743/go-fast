package routes

import (
	userpb "github.com/zhoudm1743/go-fast/app/grpc/proto/user"
	"github.com/zhoudm1743/go-fast/app/grpc/services"
	"github.com/zhoudm1743/go-fast/framework/facades"
)

// RegisterGRPC 向 gRPC Server 注册所有服务。
// 调用方式与 RegisterAdmin / RegisterApp 完全一致。
func RegisterGRPC() {
	grpcServer := facades.GRPC()

	// 注册用户服务
	grpcServer.RegisterService(
		&userpb.UserService_ServiceDesc,
		&services.UserServiceServer{},
	)

	// 在此追加更多服务注册：
	// grpcServer.RegisterService(&orderpb.OrderService_ServiceDesc, &services.OrderServiceServer{})
}

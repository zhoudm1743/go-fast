package contracts

import "google.golang.org/grpc"

// GRPCServer gRPC 服务器契约。
// 应用层通过此接口注册服务、启动/关闭服务器，
// 不直接依赖 google.golang.org/grpc 包。
type GRPCServer interface {
	// RegisterService 向 gRPC Server 注册一个服务描述和实现。
	// desc 为 protoc 生成的 *grpc.ServiceDesc，impl 为业务实现。
	RegisterService(desc *grpc.ServiceDesc, impl any)

	// Run 启动 gRPC 监听（阻塞），addr 可选（默认读取配置）。
	Run(addr ...string) error

	// Shutdown 优雅关闭 gRPC Server，等待进行中的 RPC 完成。
	Shutdown()

	// RawServer 返回底层的 *grpc.Server，用于高级场景（反射、健康检查等）。
	RawServer() *grpc.Server
}

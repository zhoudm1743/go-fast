package grpc

import (
	"github.com/zhoudm1743/go-fast/framework/contracts"
	"github.com/zhoudm1743/go-fast/framework/foundation"
)

// ServiceProvider gRPC 服务提供者。
type ServiceProvider struct{}

func (sp *ServiceProvider) Register(app foundation.Application) {
	app.Singleton("grpc", func(app foundation.Application) (any, error) {
		cfg := app.MustMake("config").(contracts.Config)
		log := app.MustMake("log").(contracts.Log)
		return NewServer(cfg, log)
	})
}

func (sp *ServiceProvider) Boot(app foundation.Application) error {
	// 注册优雅关闭钩子
	app.OnShutdown(func() {
		if s, err := app.Make("grpc"); err == nil {
			s.(contracts.GRPCServer).Shutdown()
		}
	})
	return nil
}

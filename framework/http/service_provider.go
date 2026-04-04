package http

import (
	"github.com/zhoudm1743/go-fast/framework/contracts"
	fiberdriver "github.com/zhoudm1743/go-fast/framework/http/fiber"
	gindriver "github.com/zhoudm1743/go-fast/framework/http/gin"
	"github.com/zhoudm1743/go-fast/framework/foundation"
)

// ServiceProvider HTTP 路由服务提供者。
// 通过配置 server.driver（fiber | gin）选择底层框架，默认为 fiber。
type ServiceProvider struct{}

func (sp *ServiceProvider) Register(app foundation.Application) {
	app.Singleton("route", func(app foundation.Application) (any, error) {
		cfg := app.MustMake("config").(contracts.Config)
		validator := app.MustMake("validator").(contracts.Validation)
		storage := app.MustMake("storage").(contracts.Storage)
		log := app.MustMake("log").(contracts.Log)

		driver := cfg.GetString("server.driver", "fiber")
		switch driver {
		case "gin":
			return gindriver.NewRoute(cfg, validator, storage, log)
		default: // fiber
			return fiberdriver.NewRoute(cfg, validator, storage, log)
		}
	})
}

func (sp *ServiceProvider) Boot(app foundation.Application) error {
	app.OnShutdown(func() {
		if r, err := app.Make("route"); err == nil {
			if closer, ok := r.(contracts.Route); ok {
				_ = closer.Shutdown()
			}
		}
	})
	return nil
}



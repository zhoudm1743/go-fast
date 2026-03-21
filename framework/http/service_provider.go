package http

import (
	"go-fast/framework/contracts"
	"go-fast/framework/foundation"
)

// ServiceProvider HTTP 路由服务提供者。
type ServiceProvider struct{}

func (sp *ServiceProvider) Register(app foundation.Application) {
	app.Singleton("route", func(app foundation.Application) (any, error) {
		cfg := app.MustMake("config").(contracts.Config)
		return NewRoute(cfg)
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

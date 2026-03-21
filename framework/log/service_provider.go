package log

import (
	"go-fast/framework/contracts"
	"go-fast/framework/foundation"
)

// ServiceProvider Log 服务提供者。
type ServiceProvider struct{}

func (sp *ServiceProvider) Register(app foundation.Application) {
	app.Singleton("log", func(app foundation.Application) (any, error) {
		cfg := app.MustMake("config").(contracts.Config)
		return NewLogger(cfg)
	})
}

func (sp *ServiceProvider) Boot(app foundation.Application) error {
	return nil
}

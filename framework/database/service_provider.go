package database

import (
	"go-fast/framework/contracts"
	"go-fast/framework/foundation"
)

// ServiceProvider Database 服务提供者。
type ServiceProvider struct{}

func (sp *ServiceProvider) Register(app foundation.Application) {
	app.Singleton("orm", func(app foundation.Application) (any, error) {
		cfg := app.MustMake("config").(contracts.Config)
		log := app.MustMake("log").(contracts.Log)
		return NewOrm(cfg, log)
	})
}

func (sp *ServiceProvider) Boot(app foundation.Application) error {
	app.OnShutdown(func() {
		if o, err := app.Make("orm"); err == nil {
			if closer, ok := o.(contracts.Orm); ok {
				_ = closer.Close()
			}
		}
	})
	return nil
}

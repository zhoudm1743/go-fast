package config

import "github.com/zhoudm1743/go-fast/framework/foundation"

// ServiceProvider Config 服务提供者。
type ServiceProvider struct{}

func (sp *ServiceProvider) Register(app foundation.Application) {
	app.Singleton("config", func(app foundation.Application) (any, error) {
		return NewConfig(app.BasePath("config", "config.yaml"))
	})
}

func (sp *ServiceProvider) Boot(app foundation.Application) error {
	return nil
}

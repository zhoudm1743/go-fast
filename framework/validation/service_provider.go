package validation

import "github.com/zhoudm1743/go-fast/framework/foundation"

// ServiceProvider Validation 服务提供者。
type ServiceProvider struct{}

func (sp *ServiceProvider) Register(app foundation.Application) {
	app.Singleton("validator", func(app foundation.Application) (any, error) {
		return NewValidator()
	})
}

func (sp *ServiceProvider) Boot(app foundation.Application) error {
	return nil
}

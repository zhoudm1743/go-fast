package test

import (
	"time"

	"github.com/zhoudm1743/go-fast-framework/contracts"
	"github.com/zhoudm1743/go-fast-framework/foundation"
)

// ServiceProvider Test 示例服务提供者。
type ServiceProvider struct{}

func (sp *ServiceProvider) Register(app foundation.Application) {
	app.Singleton("test", func(app foundation.Application) (any, error) {
		cfg := app.MustMake("config").(contracts.Config)
		cache := app.MustMake("cache").(contracts.Cache)
		log := app.MustMake("log").(contracts.Log)
		return NewTestService(cfg, cache, log)
	})
}

func (sp *ServiceProvider) Boot(app foundation.Application) error {
	svc := app.MustMake("test").(*testService)
	svc.Log().Info("[GoFast] Test service booted")

	// 预热缓存示例：启动时写入一条状态键
	_ = svc.Cache().Put("test:booted_at", time.Now().Unix(), time.Hour)

	return nil
}

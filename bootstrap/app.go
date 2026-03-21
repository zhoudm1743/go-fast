package bootstrap

import (
	"go-fast/framework/cache"
	"go-fast/framework/config"
	"go-fast/framework/database"
	"go-fast/framework/facades"
	"go-fast/framework/filesystem"
	"go-fast/framework/foundation"
	gohttp "go-fast/framework/http"
	"go-fast/framework/log"
	"go-fast/framework/validation"
)

// Boot 创建并引导 GoFast 应用。
// 按声明顺序注册所有内置 ServiceProvider，然后执行 Boot。
func Boot() foundation.Application {
	app := foundation.NewApplication(".")

	app.SetProviders(providers())

	app.Boot()

	facades.SetApp(app)

	return app
}

// providers 返回内置服务提供者列表。
// 顺序即 Register → Boot 的执行顺序，请确保依赖在前。
func providers() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&config.ServiceProvider{},     // 1. 配置
		&log.ServiceProvider{},        // 2. 日志
		&cache.ServiceProvider{},      // 3. 缓存
		&database.ServiceProvider{},   // 4. 数据库
		&filesystem.ServiceProvider{}, // 5. 文件系统
		&validation.ServiceProvider{}, // 6. 验证器
		&gohttp.ServiceProvider{},     // 7. HTTP 路由
		// 业务方可在此追加自定义 Provider
	}
}

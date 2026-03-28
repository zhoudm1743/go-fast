package bootstrap

import (
	"github.com/zhoudm1743/go-fast/framework/cache"
	"github.com/zhoudm1743/go-fast/framework/config"
	"github.com/zhoudm1743/go-fast/framework/database"
	"github.com/zhoudm1743/go-fast/framework/facades"
	"github.com/zhoudm1743/go-fast/framework/filesystem"
	"github.com/zhoudm1743/go-fast/framework/foundation"
	gogrpc "github.com/zhoudm1743/go-fast/framework/gRPC"
	gohttp "github.com/zhoudm1743/go-fast/framework/http"
	"github.com/zhoudm1743/go-fast/framework/log"
	"github.com/zhoudm1743/go-fast/framework/validation"
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
		&gogrpc.ServiceProvider{},     // 8. gRPC 服务器
	}
}

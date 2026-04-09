package bootstrap

import (
	"github.com/zhoudm1743/go-fast/framework/cache"
	"github.com/zhoudm1743/go-fast/framework/config"
	"github.com/zhoudm1743/go-fast/framework/database"
	"github.com/zhoudm1743/go-fast/framework/facades"
	"github.com/zhoudm1743/go-fast/framework/fast"
	"github.com/zhoudm1743/go-fast/framework/filesystem"
	"github.com/zhoudm1743/go-fast/framework/foundation"
	gogrpc "github.com/zhoudm1743/go-fast/framework/grpc"
	gohttp "github.com/zhoudm1743/go-fast/framework/http"
	gojwt "github.com/zhoudm1743/go-fast/framework/jwt"
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

	// 注册所有控制台命令到 Fast 内核
	facades.Fast().Register(Commands())

	return app
}

// providers 返回服务提供者列表。
// 顺序即 Register → Boot 的执行顺序，请确保依赖在前。
func providers() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&config.ServiceProvider{},     // 1. 配置
		&log.ServiceProvider{},        // 2. 日志
		&cache.ServiceProvider{},      // 3. 缓存
		&database.ServiceProvider{},   // 4. 数据库
		&filesystem.ServiceProvider{}, // 5. 文件系统
		&validation.ServiceProvider{}, // 6. 验证器
		&gojwt.ServiceProvider{},      // 7. JWT
		&gohttp.ServiceProvider{},     // 8. HTTP 路由
		&gogrpc.ServiceProvider{},     // 9. gRPC 服务器
		&fast.ServiceProvider{},       // 10. 控制台
	}
}

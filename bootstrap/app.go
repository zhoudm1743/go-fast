package bootstrap

import (
	_ "github.com/zhoudm1743/go-fast/config" // 触发 Go 配置文件的 init() 注册

	"github.com/zhoudm1743/go-fast/framework/cache"
	"github.com/zhoudm1743/go-fast/framework/config"
	"github.com/zhoudm1743/go-fast/framework/database"
	goevent "github.com/zhoudm1743/go-fast/framework/event"
	"github.com/zhoudm1743/go-fast/framework/facades"
	"github.com/zhoudm1743/go-fast/framework/fast"
	"github.com/zhoudm1743/go-fast/framework/filesystem"
	"github.com/zhoudm1743/go-fast/framework/foundation"
	gohttp "github.com/zhoudm1743/go-fast/framework/http"
	gojwt "github.com/zhoudm1743/go-fast/framework/jwt"
	"github.com/zhoudm1743/go-fast/framework/log"
	goqueue "github.com/zhoudm1743/go-fast/framework/queue"
	goschedule "github.com/zhoudm1743/go-fast/framework/schedule"
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
		&gojwt.ServiceProvider{},      // 6. JWT
		&gohttp.ServiceProvider{},     // 7. HTTP 路由（含验证器）
		&fast.ServiceProvider{},       // 8. 控制台
		&goevent.ServiceProvider{},    // 9. 事件系统
		&goqueue.ServiceProvider{},    // 10. 队列系统
		&goschedule.ServiceProvider{}, // 11. 任务调度
	}
}

package bootstrap

import (
	_ "github.com/zhoudm1743/go-fast/config" // 触发 Go 配置文件的 init() 注册

	"github.com/zhoudm1743/go-fast/app/events"
	"github.com/zhoudm1743/go-fast/app/listeners"
	testsvc "github.com/zhoudm1743/go-fast/services/test"
	"github.com/zhoudm1743/go-fast-framework/cache"
	"github.com/zhoudm1743/go-fast-framework/config"
	"github.com/zhoudm1743/go-fast-framework/contracts"
	"github.com/zhoudm1743/go-fast-framework/database"
	goevent "github.com/zhoudm1743/go-fast-framework/event"
	"github.com/zhoudm1743/go-fast-framework/facades"
	"github.com/zhoudm1743/go-fast-framework/fast"
	"github.com/zhoudm1743/go-fast-framework/filesystem"
	"github.com/zhoudm1743/go-fast-framework/foundation"
	gohttp "github.com/zhoudm1743/go-fast-framework/http"
	gojwt "github.com/zhoudm1743/go-fast-framework/jwt"
	"github.com/zhoudm1743/go-fast-framework/log"
	goqueue "github.com/zhoudm1743/go-fast-framework/queue"
	goschedule "github.com/zhoudm1743/go-fast-framework/schedule"
	"github.com/zhoudm1743/go-fast-framework/tenant"
)

// Boot 创建并引导 GoFast 应用。
// 按声明顺序注册所有内置 ServiceProvider，然后执行 Boot。
func Boot() foundation.Application {
	app := foundation.NewApplication(".")

	app.SetProviders(providers())

	app.Boot()

	facades.SetApp(app)

	// 注册事件 → 监听器映射
	goevent.RegisterEvents(app, map[contracts.Eventer][]contracts.EventListener{
		&events.OrderShipped{}: {
			&listeners.SendShipmentNotification{},
		},
	})

	// 注册调度任务并启动调度器（示例：每天执行 example 命令）
	_ = goschedule.RegisterSchedule(app, []contracts.ScheduleEvent{
		facades.Schedule().Command("example").Daily(),
	})

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
		&tenant.ServiceProvider{},     // 4. 租户上下文（database 之前）
		&database.ServiceProvider{},   // 5. 数据库
		&filesystem.ServiceProvider{}, // 6. 文件系统
		&gojwt.ServiceProvider{},      // 7. JWT
		&gohttp.ServiceProvider{},     // 8. HTTP 路由（含验证器）
		&fast.ServiceProvider{},       // 9. 控制台
		&goevent.ServiceProvider{},    // 10. 事件系统
		&goqueue.ServiceProvider{},    // 11. 队列系统
		&goschedule.ServiceProvider{}, // 12. 任务调度
		&testsvc.ServiceProvider{},    // 13. 自定义示例服务
	}
}

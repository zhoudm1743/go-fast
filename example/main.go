package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/zhoudm1743/go-fast/example/app/models"
	"github.com/zhoudm1743/go-fast/example/routes"
	"github.com/zhoudm1743/go-fast-framework/cache"
	"github.com/zhoudm1743/go-fast-framework/config"
	"github.com/zhoudm1743/go-fast-framework/database"
	"github.com/zhoudm1743/go-fast-framework/facades"
	"github.com/zhoudm1743/go-fast-framework/filesystem"
	"github.com/zhoudm1743/go-fast-framework/foundation"
	gohttp "github.com/zhoudm1743/go-fast-framework/http"
	"github.com/zhoudm1743/go-fast-framework/log"
)

func main() {
	app := foundation.NewApplication(".")

	app.SetProviders([]foundation.ServiceProvider{
		&config.ServiceProvider{},
		&log.ServiceProvider{},
		&cache.ServiceProvider{},
		&database.ServiceProvider{},
		&filesystem.ServiceProvider{},
		&gohttp.ServiceProvider{},
	})
	app.Boot()
	facades.SetApp(app)

	// 自动迁移
	if err := facades.DB().AutoMigrate(&models.User{}); err != nil {
		facades.Log().Fatalf("数据库迁移失败: %v", err)
	}

	// 注册路由
	routes.Register()

	// 启动 HTTP 服务
	go func() {
		if err := facades.Http.Route().Run(); err != nil {
			facades.Log().Errorf("服务启动失败: %v", err)
		}
	}()

	facades.Log().Info("GoFast 示例服务已启动: http://localhost:8000")

	// 等待退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	facades.Log().Info("正在关闭...")
	app.Shutdown()
}

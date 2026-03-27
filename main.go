package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/zhoudm1743/go-fast/bootstrap"
	"github.com/zhoudm1743/go-fast/framework/facades"
	"github.com/zhoudm1743/go-fast/routes"
)

func main() {
	// 1. 创建并引导应用（注册所有 Provider → Register → Boot）
	app := bootstrap.Boot()

	fmt.Printf("[GoFast] v%s booted\n", app.Version())

	// 2. 注册路由
	routes.Register()

	// 3. 启动 HTTP 服务器（协程，非阻塞）
	go func() {
		if err := facades.Route().Run(); err != nil {
			facades.Log().Errorf("server error: %v", err)
		}
	}()

	// 4. 等待退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\n[GoFast] shutting down...")

	// 5. 优雅关闭
	app.Shutdown()
}

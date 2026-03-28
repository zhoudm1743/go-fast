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

	// 2. Fast 控制台模式：go run . fast [command] [args]
	if len(os.Args) > 1 && os.Args[1] == "fast" {
		if err := facades.Fast().Run(os.Args[2:]); err != nil {
			os.Exit(1)
		}
		return
	}

	// 3. 注册路由（HTTP + gRPC）
	routes.Register()
	routes.RegisterGRPC()

	// 4. 启动 HTTP 服务器（协程，非阻塞）
	go func() {
		if err := facades.Route().Run(); err != nil {
			facades.Log().Errorf("server error: %v", err)
		}
	}()

	// 5. 启动 gRPC 服务器（协程，非阻塞）
	// go func() {
	// 	if err := facades.GRPC().Run(); err != nil {
	// 		facades.Log().Errorf("grpc server error: %v", err)
	// 	}
	// }()

	// 6. 等待退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\n[GoFast] shutting down...")

	// 7. 优雅关闭（HTTP + gRPC 均通过 OnShutdown 钩子关闭）
	app.Shutdown()
}

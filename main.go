package main

import (
	"os"
	"runtime"

	"github.com/zhoudm1743/go-fast-framework/facades"
	"github.com/zhoudm1743/go-fast/bootstrap"
	"github.com/zhoudm1743/go-fast/routes"
	gofastfyne "github.com/zhoudm1743/gofast-fyne"
)

func main() {
	// Fyne/GLFW 要求 UI 操作发生在主线程；
	// 先锁定，避免 bootstrap 启动的后台 goroutine 把主 goroutine 挤离主线程。
	runtime.LockOSThread()

	// 1. 创建并引导应用（注册所有 Provider → Register → Boot）
	app := bootstrap.Boot()

	// fmt.Printf("[GoFast] v%s booted\n", app.Version())

	// 2. Fast 控制台模式：go run . fast [command] [args]
	//    （GUI 驱动为延迟加载，CLI 场景完全不初始化 Fyne）
	if len(os.Args) > 1 && os.Args[1] == "fast" {
		if err := facades.Fast().Run(os.Args[2:]); err != nil {
			os.Exit(1)
		}
		return
	}

	// 3. 桌面 GUI 模式：装配窗口页面并阻塞运行事件循环。
	//    窗口关闭后 gui.Run 返回，自动触发框架优雅关闭（数据库/队列等逆序回收）。
	gui := app.MustMake("gui").(gofastfyne.Gui)
	routes.RegisterWindow(gui)
	facades.Log().Info("[GoFast] desktop gui starting...")
	if err := gui.Run(); err != nil {
		facades.Log().Errorf("gui error: %v", err)
		os.Exit(1)
	}
}

package routes

import (
	"fyne.io/fyne/v2/container"

	maincontroller "github.com/zhoudm1743/go-fast/app/window/main/controller"
	mainmiddleware "github.com/zhoudm1743/go-fast/app/window/main/middleware"
	operatecontroller "github.com/zhoudm1743/go-fast/app/window/operate/controller"
	gofastfyne "github.com/zhoudm1743/gofast-fyne"
)

// RegisterWindow 注册桌面窗口：
// 动态菜单 → 次级窗口事件 → 登录守卫 → 装配主窗口页面。
// 由 main.go 在获取 gui 服务后调用（主线程）。
func RegisterWindow(g gofastfyne.Gui) {
	// 菜单控制器：登录成功后动态重建菜单（gofastfyne.SetMenu）
	(&maincontroller.MenuController{}).Register(g)

	// 订单操作窗口（次级窗口，经事件打开，演示多窗口管理）
	g.On("window.order.open", func(map[string]any) {
		g.OpenWindow("order", "订单操作", (&operatecontroller.OrderController{}).View(g))
	})

	// 登录守卫：未登录先弹登录窗口（login 域），成功后装配主界面
	mainmiddleware.Auth(g, func() { mountMain(g) })()
}

// mountMain 装配主窗口页面（main 域控制器 → AppTabs 导航）。
func mountMain(g gofastfyne.Gui) {
	ctrls := []gofastfyne.WindowController{
		&maincontroller.HomeController{},
		&maincontroller.UserController{},
	}

	tabs := make([]*container.TabItem, 0, len(ctrls))
	for _, c := range ctrls {
		tabs = append(tabs, container.NewTabItem(c.Title(), c.View(g)))
	}
	appTabs := container.NewAppTabs(tabs...)
	appTabs.SetTabLocation(container.TabLocationLeading)
	g.SetContent(appTabs)
}

package routes

import (
	"fyne.io/fyne/v2/container"

	windowControllers "github.com/zhoudm1743/go-fast/app/window/controllers"
	gofastfyne "github.com/zhoudm1743/gofast-fyne"
)

// RegisterWindow 注册桌面窗口页面（MVC：控制器装配进主窗口导航）。
// 由 main.go 在获取 gui 服务后调用。
func RegisterWindow(g gofastfyne.Gui) {
	ctrls := []gofastfyne.WindowController{
		&windowControllers.HomeController{},
		&windowControllers.UserController{},
	}

	tabs := make([]*container.TabItem, 0, len(ctrls))
	for _, c := range ctrls {
		tabs = append(tabs, container.NewTabItem(c.Title(), c.View(g)))
	}
	appTabs := container.NewAppTabs(tabs...)
	appTabs.SetTabLocation(container.TabLocationLeading)
	g.SetContent(appTabs)
}

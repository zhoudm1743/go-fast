package controller

import (
	"fmt"

	gofastfyne "github.com/zhoudm1743/gofast-fyne"
)

// MenuController 主窗口菜单控制器。
// 启动时的菜单来自 config 的 fyne.menu（配置驱动）；
// 这里演示运行期动态重建：登录成功后把当前用户写进菜单。
type MenuController struct{}

// Register 绑定菜单相关事件。
func (c *MenuController) Register(g gofastfyne.Gui) {
	g.On("app.logged_in", func(p map[string]any) {
		user, _ := p["user"].(string)
		g.SetMenu([]gofastfyne.MenuItem{
			{
				Label: "应用",
				Items: []gofastfyne.MenuItem{
					{Label: "打开订单窗口", Event: "window.order.open"},
					{Separator: true},
					{Label: fmt.Sprintf("当前用户: %s", user), Event: "app.about"},
					{Label: "退出", Action: gofastfyne.ActionQuit},
				},
			},
		})
	})
}

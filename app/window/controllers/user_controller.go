package controllers

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/zhoudm1743/go-fast-framework/facades"
	"github.com/zhoudm1743/go-fast/app/models"
	gofastfyne "github.com/zhoudm1743/gofast-fyne"
)

// UserController 用户列表页：Model（models.User）→ View（widget.List），
// 数据访问直接复用框架数据库服务。
type UserController struct{}

// Title 页面标题（实现 gofastfyne.WindowController）。
func (c *UserController) Title() string { return "用户" }

// View 构建用户列表页。
func (c *UserController) View(g gofastfyne.Gui) fyne.CanvasObject {
	var users []models.User

	status := widget.NewLabel("加载中 …")
	list := widget.NewList(
		func() int { return len(users) },
		func() fyne.CanvasObject { return widget.NewLabel("template") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(fmt.Sprintf("%s <%s>", users[i].Name, users[i].Email))
		},
	)

	reload := func() {
		var fresh []models.User
		if err := facades.DB().Query().Find(&fresh); err != nil {
			status.SetText("查询失败: " + err.Error())
			return
		}
		users = fresh
		status.SetText(fmt.Sprintf("共 %d 个用户", len(users)))
		list.Refresh()
	}

	// 菜单「刷新用户列表」（配置 fyne.menu → event: window.users.refresh）走同一入口
	g.On("window.users.refresh", func(map[string]any) { reload() })

	reload()
	return container.NewBorder(nil, status, nil, nil, list)
}

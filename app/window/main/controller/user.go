package controller

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"github.com/zhoudm1743/go-fast-framework/facades"
	"github.com/zhoudm1743/go-fast/app/models"
	"github.com/zhoudm1743/go-fast/app/window/main/view"
	gofastfyne "github.com/zhoudm1743/gofast-fyne"
)

// UserController 主窗口用户管理页：Model（models.User）→ View（view.UserList），
// 数据访问复用框架数据库服务。
type UserController struct{}

// Title 页面标题（实现 gofastfyne.WindowController）。
func (c *UserController) Title() string { return "用户" }

// View 构建用户列表页。
func (c *UserController) View(g gofastfyne.Gui) fyne.CanvasObject {
	var users []models.User

	// 视图层装配界面，数据绑定闭包由控制器提供
	v := view.NewUserList(
		func() int { return len(users) },
		func() fyne.CanvasObject { return widget.NewLabel("template") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(fmt.Sprintf("%s <%s>", users[i].Name, users[i].Email))
		},
	)

	reload := func() {
		var fresh []models.User
		if err := facades.DB().Query().Find(&fresh); err != nil {
			v.Status.SetText("查询失败: " + err.Error())
			return
		}
		users = fresh
		v.Status.SetText(fmt.Sprintf("共 %d 个用户", len(users)))
		v.List.Refresh()
	}

	v.Refresh.OnTapped = reload
	// 菜单「刷新用户列表」（配置 fyne.menu → event: window.users.refresh）走同一入口
	g.On("window.users.refresh", func(map[string]any) { reload() })

	reload()
	return v.Root
}

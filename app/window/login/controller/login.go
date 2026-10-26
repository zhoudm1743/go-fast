package controller

import (
	"fyne.io/fyne/v2"

	"github.com/zhoudm1743/go-fast/app/window/login/view"
	gofastfyne "github.com/zhoudm1743/gofast-fyne"
)

// LoginController 登录窗口控制器（次级窗口，非主窗口页面）。
// 演示任意账号即可登录；真实场景应校验框架数据库 / 鉴权服务。
type LoginController struct {
	onSuccess func(user string)
}

// New 创建登录控制器；登录成功后回调 onSuccess（由中间件注入后续动作）。
func New(onSuccess func(user string)) *LoginController {
	return &LoginController{onSuccess: onSuccess}
}

// View 构建登录窗口内容。
func (c *LoginController) View(g gofastfyne.Gui) fyne.CanvasObject {
	v := view.NewLogin()
	v.Submit.OnTapped = func() {
		if v.Name.Text == "" {
			v.Hint.SetText("请输入用户名")
			return
		}
		if v.Password.Text == "" {
			v.Hint.SetText("请输入密码")
			return
		}
		g.CloseWindow("login")
		c.onSuccess(v.Name.Text)
	}
	return v.Root
}

package view

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Login 登录窗口视图。
type Login struct {
	Root fyne.CanvasObject

	Name     *widget.Entry
	Password *widget.Entry
	Hint     *widget.Label
	Submit   *widget.Button
}

// NewLogin 构建登录表单。
func NewLogin() *Login {
	v := &Login{
		Name:     widget.NewEntry(),
		Password: widget.NewPasswordEntry(),
		Hint:     widget.NewLabel(""),
		Submit:   widget.NewButton("登录", nil),
	}
	v.Name.SetPlaceHolder("用户名")
	v.Password.SetPlaceHolder("密码")

	v.Root = container.NewVBox(
		widget.NewLabelWithStyle("登录", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		v.Name,
		v.Password,
		v.Hint,
		v.Submit,
	)
	return v
}

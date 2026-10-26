package view

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Home 首页视图：纯界面装配，事件由 controller 绑定。
type Home struct {
	Root fyne.CanvasObject

	Ping   *widget.Button
	Pong   *widget.Label
	Show   *widget.Button
	Hide   *widget.Button
	Toggle *widget.Button
	Quit   *widget.Button
}

// NewHome 构建首页界面。
func NewHome() *Home {
	v := &Home{
		Ping:   widget.NewButton("Ping 框架事件", nil),
		Pong:   widget.NewLabel("等待 ping …"),
		Show:   widget.NewButton("显示", nil),
		Hide:   widget.NewButton("隐藏", nil),
		Toggle: widget.NewButton("切换", nil),
		Quit:   widget.NewButton("退出", nil),
	}
	v.Root = container.NewVBox(
		widget.NewLabelWithStyle("GoFast Desktop", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		v.Ping,
		v.Pong,
		widget.NewSeparator(),
		container.NewHBox(v.Show, v.Hide, v.Toggle, v.Quit),
	)
	return v
}

package controllers

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/zhoudm1743/go-fast-framework/facades"
	gofastfyne "github.com/zhoudm1743/gofast-fyne"
)

// HomeController 首页：窗口级事件与事件总线闭环演示、窗口显隐控制。
type HomeController struct{}

// Title 页面标题（实现 gofastfyne.WindowController）。
func (c *HomeController) Title() string { return "首页" }

// View 构建首页视图；g 提供窗口控制与事件桥接能力。
func (c *HomeController) View(g gofastfyne.Gui) fyne.CanvasObject {
	// ── 窗口级事件 → 框架日志（回调已在 UI 线程）──────────────────
	g.On(gofastfyne.EventWindowShown, func(map[string]any) {
		facades.Log().Info("[window] shown")
	})
	g.On(gofastfyne.EventWindowHidden, func(map[string]any) {
		facades.Log().Info("[window] hidden")
	})
	g.On(gofastfyne.EventWindowClosed, func(map[string]any) {
		facades.Log().Info("[window] closed")
	})

	// ── 事件总线闭环演示：按钮 Emit → 框架总线 → On 监听回到 UI ──
	pong := widget.NewLabel("等待 ping …")
	g.On("demo.pong", func(p map[string]any) {
		pong.SetText(fmt.Sprintf("收到 pong: %v", p["msg"]))
	})
	// 模拟框架侧对 demo.ping 的处理（真实场景应注册到 app/listeners）
	g.On("demo.ping", func(p map[string]any) {
		_ = g.Emit("demo.pong", map[string]any{"msg": p["msg"]})
	})
	ping := widget.NewButton("Ping 框架事件", func() {
		if err := g.Emit("demo.ping", map[string]any{"msg": "hi"}); err != nil {
			facades.Log().Errorf("[gui] emit demo.ping failed: %v", err)
		}
	})

	// ── 窗口显隐控制（等价于托盘/菜单内置动作）────────────────────
	buttons := container.NewHBox(
		widget.NewButton("显示", g.ShowMainWindow),
		widget.NewButton("隐藏", g.HideMainWindow),
		widget.NewButton("切换", g.ToggleMainWindow),
		widget.NewButton("退出", g.Quit),
	)

	return container.NewVBox(
		widget.NewLabelWithStyle("GoFast Desktop", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		ping,
		pong,
		widget.NewSeparator(),
		buttons,
	)
}

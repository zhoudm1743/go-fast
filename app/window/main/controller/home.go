package controller

import (
	"fmt"

	"fyne.io/fyne/v2"

	"github.com/zhoudm1743/go-fast-framework/facades"
	"github.com/zhoudm1743/go-fast/app/window/main/view"
	gofastfyne "github.com/zhoudm1743/gofast-fyne"
)

// HomeController 主窗口首页：事件总线闭环演示、窗口显隐控制、跨窗口事件展示。
type HomeController struct{}

// Title 页面标题（实现 gofastfyne.WindowController）。
func (c *HomeController) Title() string { return "首页" }

// View 构建首页：视图装配在 view 包，事件绑定在控制器。
func (c *HomeController) View(g gofastfyne.Gui) fyne.CanvasObject {
	v := view.NewHome()

	// ── 窗口级事件 → 框架日志（回调已在 UI 线程）──────────────────
	g.On(gofastfyne.EventWindowShown, func(map[string]any) {
		facades.Log().Info("[window] shown")
	})
	g.On(gofastfyne.EventWindowHidden, func(map[string]any) {
		facades.Log().Info("[window] hidden")
	})
	g.On(gofastfyne.EventWindowClosed, func(p map[string]any) {
		facades.Log().Infof("[window] closed: %v", p["window"])
	})

	// ── 事件总线闭环：Ping → 框架总线 → 回到 UI ──────────────────
	g.On("demo.pong", func(p map[string]any) {
		v.Pong.SetText(fmt.Sprintf("收到 pong: %v", p["msg"]))
	})
	// 模拟框架侧对 demo.ping 的处理（真实场景应注册到 app/listeners）
	g.On("demo.ping", func(p map[string]any) {
		_ = g.Emit("demo.pong", map[string]any{"msg": p["msg"]})
	})
	v.Ping.OnTapped = func() {
		if err := g.Emit("demo.ping", map[string]any{"msg": "hi"}); err != nil {
			facades.Log().Errorf("[gui] emit demo.ping failed: %v", err)
		}
	}

	// ── 跨窗口事件：订单窗口创建成功 → 首页展示 ──────────────────
	g.On("order.created", func(p map[string]any) {
		v.Pong.SetText(fmt.Sprintf("订单已创建: %v", p["no"]))
	})

	// ── 窗口显隐控制（等价于托盘/菜单内置动作）────────────────────
	v.Show.OnTapped = g.ShowMainWindow
	v.Hide.OnTapped = g.HideMainWindow
	v.Toggle.OnTapped = g.ToggleMainWindow
	v.Quit.OnTapped = g.Quit

	return v.Root
}

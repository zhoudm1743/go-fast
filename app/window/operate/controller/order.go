package controller

import (
	"fyne.io/fyne/v2"

	"github.com/zhoudm1743/go-fast-framework/facades"
	"github.com/zhoudm1743/go-fast/app/window/operate/view"
	gofastfyne "github.com/zhoudm1743/gofast-fyne"
)

// OrderController 订单操作窗口控制器（次级窗口，经 window.order.open 打开）。
type OrderController struct{}

// View 构建订单窗口内容。
func (c *OrderController) View(g gofastfyne.Gui) fyne.CanvasObject {
	v := view.NewOrder()
	v.Create.OnTapped = func() {
		if v.No.Text == "" {
			v.Hint.SetText("请输入订单号")
			return
		}
		if err := g.Emit("order.created", map[string]any{"no": v.No.Text}); err != nil {
			facades.Log().Errorf("[gui] emit order.created failed: %v", err)
			return
		}
		v.Hint.SetText("已创建: " + v.No.Text)
	}
	return v.Root
}

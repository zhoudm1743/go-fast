package view

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Order 订单操作窗口视图。
type Order struct {
	Root fyne.CanvasObject

	No     *widget.Entry
	Create *widget.Button
	Hint   *widget.Label
}

// NewOrder 构建订单表单。
func NewOrder() *Order {
	v := &Order{
		No:     widget.NewEntry(),
		Create: widget.NewButton("创建订单", nil),
		Hint:   widget.NewLabel(""),
	}
	v.No.SetPlaceHolder("订单号")

	v.Root = container.NewVBox(
		widget.NewLabelWithStyle("订单操作", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		v.No,
		v.Create,
		v.Hint,
	)
	return v
}

package view

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// UserList 用户列表视图：数据绑定函数由 controller 注入。
type UserList struct {
	Root    fyne.CanvasObject
	List    *widget.List
	Status  *widget.Label
	Refresh *widget.Button
}

// NewUserList 构建用户列表界面；length/create/update 为列表数据绑定。
func NewUserList(
	length func() int,
	create func() fyne.CanvasObject,
	update func(widget.ListItemID, fyne.CanvasObject),
) *UserList {
	v := &UserList{
		List:    widget.NewList(length, create, update),
		Status:  widget.NewLabel("加载中 …"),
		Refresh: widget.NewButton("刷新", nil),
	}
	v.Root = container.NewBorder(nil, container.NewHBox(v.Refresh, v.Status), nil, nil, v.List)
	return v
}

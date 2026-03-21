package routes

// Register 注册所有路由，入口函数由 main.go 调用。
func Register() {
	RegisterApp()   // 前台路由：/api/...
	RegisterAdmin() // 后台路由：/admin/api/...
}

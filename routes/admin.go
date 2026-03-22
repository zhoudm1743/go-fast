package routes

import (
	adminControllers "go-fast/app/http/admin/controllers"
	adminMiddleware "go-fast/app/http/admin/middleware"
	"go-fast/framework/contracts"
	"go-fast/framework/facades"
)

// RegisterAdmin 注册后台管理路由，统一前缀 /admin。
func RegisterAdmin() {
	facades.Route().Group("/admin", adminMiddleware.AdminAuth, func(admin contracts.Route) {
		admin.Register(
			&adminControllers.UserController{},
		)
	})
}

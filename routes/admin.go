package routes

import (
	adminControllers "github.com/zhoudm1743/go-fast/app/http/admin/controllers"
	adminMiddleware "github.com/zhoudm1743/go-fast/app/http/admin/middleware"
	"github.com/zhoudm1743/go-fast/framework/contracts"
	"github.com/zhoudm1743/go-fast/framework/facades"
)

// RegisterAdmin 注册后台管理路由，统一前缀 /admin。
func RegisterAdmin() {
	facades.Route().Group("/admin", adminMiddleware.AdminAuth, func(admin contracts.Route) {
		admin.Register(
			&adminControllers.UserController{},
		)
	})
}

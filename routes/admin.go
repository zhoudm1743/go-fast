package routes

import (
	adminControllers "go-fast/app/http/admin/controllers"
	adminMiddleware "go-fast/app/http/admin/middleware"
	"go-fast/framework/facades"
)

// RegisterAdmin 注册后台管理路由，统一前缀 /admin。
func RegisterAdmin() {
	user := adminControllers.UserController{}

	admin := facades.Route().Group("/admin")
	admin.Use(adminMiddleware.AdminAuth) // 整组路由都需要后台鉴权

	admin.Get("/users", user.Index)
	admin.Get("/users/:id", user.Show)
	admin.Post("/users", user.Store)
	admin.Put("/users/:id", user.Update)
	admin.Delete("/users/:id", user.Destroy)
}

package routes

import (
	appControllers "go-fast/app/http/app/controllers"
	appMiddleware "go-fast/app/http/app/middleware"
	"go-fast/framework/contracts"
	"go-fast/framework/facades"
)

// RegisterApp 注册前台路由，统一前缀 /api/v1。
func RegisterApp() {
	r := facades.Route()
	user := appControllers.UserController{}

	// 公开接口（无需登录）
	r.Get("/api/ping", func(ctx contracts.Context) error {
		return ctx.JSON(200, map[string]string{"message": "pong"})
	})

	// 需要登录的接口
	api := r.Group("/api/v1")
	api.Use(appMiddleware.Auth)

	api.Get("/user/profile", user.Profile)
	api.Put("/user/profile", user.UpdateProfile)
}

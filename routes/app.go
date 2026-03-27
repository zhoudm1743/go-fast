package routes

import (
	appControllers "github.com/zhoudm1743/go-fast/app/http/app/controllers"
	appMiddleware "github.com/zhoudm1743/go-fast/app/http/app/middleware"
	"github.com/zhoudm1743/go-fast/framework/contracts"
	"github.com/zhoudm1743/go-fast/framework/facades"
)

// RegisterApp 注册前台路由，统一前缀 /api/v1。
func RegisterApp() {
	r := facades.Route()

	// 公开接口（无需登录）
	r.Get("/api/ping", func(ctx contracts.Context) error {
		return ctx.JSON(200, map[string]string{"message": "pong"})
	})

	// 需要登录的接口
	r.Group("/api/v1", appMiddleware.Auth, func(v1 contracts.Route) {
		v1.Register(
			&appControllers.UserController{},
		)
	})
}

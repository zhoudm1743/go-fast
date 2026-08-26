package routes

import (
	appControllers "github.com/zhoudm1743/go-fast/app/http/app/controllers"
	appMiddleware "github.com/zhoudm1743/go-fast/app/http/app/middleware"
	appcontracts "github.com/zhoudm1743/go-fast/services/contracts"
	"github.com/zhoudm1743/go-fast-framework/contracts"
	"github.com/zhoudm1743/go-fast-framework/facades"
)

// RegisterApp 注册前台路由，统一前缀 /api/v1。
func RegisterApp() {
	r := facades.Http.Route()

	// 公开接口（无需登录）
	r.Get("/api/ping", func(ctx contracts.Context) error {
		test := facades.App().MustMake("test").(appcontracts.Test)
		return ctx.JSON(200, map[string]any{
			"message": test.Greet("GoFast"),
			"status":  test.Status(),
		})
	})

	// 需要登录的接口
	r.Group("/api/v1", appMiddleware.Auth, func(v1 contracts.Route) {
		v1.Register(
			&appControllers.UserController{},
		)
	})
}

package routes

import (
	"github.com/zhoudm1743/go-fast/example/app/http/controllers"
	"github.com/zhoudm1743/go-fast-framework/contracts"
	"github.com/zhoudm1743/go-fast-framework/facades"
)

// Register 注册所有路由。
func Register() {
	r := facades.Http.Route()

	// 健康检查
	r.Get("/api/ping", func(ctx contracts.Context) error {
		return ctx.Response().Success(map[string]string{"message": "pong"})
	})

	// 用户 CRUD 接口
	r.Group("/api/v1", func(v1 contracts.Route) {
		v1.Register(&controllers.UserController{})
	})
}

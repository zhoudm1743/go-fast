package routes

import (
	"go-fast/app/http/controllers"
	"go-fast/framework/contracts"
	"go-fast/framework/facades"
)

// Register 注册所有路由。在 main.go 中 bootstrap.Boot() 之后调用。
func Register() {
	r := facades.Route()

	// 基础示例
	r.Get("/api/ping", func(ctx contracts.Context) error {
		return ctx.JSON(200, map[string]string{"message": "pong"})
	})

	// 用户资源路由
	user := controllers.UserController{}
	v1 := r.Group("/api/v1")
	v1.Get("/users", user.Index)
	v1.Get("/users/:id", user.Show)
	v1.Post("/users", user.Store)
	v1.Put("/users/:id", user.Update)
	v1.Delete("/users/:id", user.Destroy)
}

package routes

import (
	"go-fast/framework/facades"

	"github.com/gofiber/fiber/v2"
)

// Register 注册所有路由。在 main.go 中 bootstrap.Boot() 之后调用。
func Register() {
	r := facades.Route()

	// 示例路由
	r.Get("/api/ping", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "pong",
		})
	})

	// api/v1 路由组
	// v1 := r.Group("/api/v1")
	// v1.Get("/users", controllers.UserController{}.Index)
}

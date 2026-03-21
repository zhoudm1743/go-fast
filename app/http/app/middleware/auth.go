package middleware

import (
	"strings"

	"go-fast/framework/contracts"
)

// Auth 前台用户鉴权中间件（示例）。
// 验证 JWT token，将 user_id 注入 ctx 供后续 Handler 读取。
func Auth(ctx contracts.Context) error {
	token := strings.TrimPrefix(ctx.Header("Authorization"), "Bearer ")
	if token == "" {
		return ctx.Response().Unauthorized("请先登录")
	}

	// TODO: 验证 token，获取用户 ID
	// userID, err := parseJWT(token)
	// if err != nil { return ctx.JSON(401, ...) }
	// ctx.WithValue("user_id", userID)

	return ctx.Next()
}

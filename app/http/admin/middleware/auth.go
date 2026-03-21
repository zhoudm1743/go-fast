package middleware

import (
	"strings"

	"go-fast/framework/contracts"
)

// AdminAuth 后台管理鉴权中间件（示例）。
// 实际项目中应验证 JWT 或 Session，并确认用户具有 admin 角色。
func AdminAuth(ctx contracts.Context) error {
	token := strings.TrimPrefix(ctx.Header("Authorization"), "Bearer ")
	if token == "" {
		return ctx.Response().Unauthorized("未授权")
	}

	// TODO: 验证 token，获取 admin 用户 ID
	// adminID, err := parseAdminJWT(token)
	// if err != nil { return ctx.Response().Unauthorized("token 无效") }
	// if !isAdmin(adminID) { return ctx.Response().Forbidden("无权限访问后台") }
	// ctx.WithValue("admin_id", adminID)

	return ctx.Next()
}

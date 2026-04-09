package contracts

import "github.com/golang-jwt/jwt/v5"

// JWT JWT 服务契约。
type JWT interface {
	// GenerateToken 根据 MapClaims 生成签名 Token 字符串。
	// claims 中可包含任意业务字段，框架层会自动注入 exp（过期时间）。
	GenerateToken(claims jwt.MapClaims) (string, error)

	// ParseToken 解析并验证 Token，返回 MapClaims。
	// Token 无效、过期或签名错误时返回 error。
	ParseToken(tokenStr string) (jwt.MapClaims, error)

	// RefreshToken 在 Token 仍有效时刷新过期时间，返回新 Token。
	RefreshToken(tokenStr string) (string, error)
}

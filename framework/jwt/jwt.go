package jwt

import (
	"errors"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/zhoudm1743/go-fast/framework/contracts"
)

type jwtService struct {
	secret     []byte
	ttl        time.Duration // token 有效期
	signingAlg gojwt.SigningMethod
}

var _ contracts.JWT = (*jwtService)(nil)

// New 根据配置创建 JWT 服务实例。
// 读取配置键：
//
//	jwt.secret  — 签名密钥（必填）
//	jwt.ttl     — 有效期（分钟），默认 60
//	jwt.alg     — 签名算法，支持 HS256 / HS384 / HS512，默认 HS256
func New(cfg contracts.Config) (contracts.JWT, error) {
	secret := cfg.GetString("jwt.secret", "")
	if secret == "" {
		return nil, errors.New("jwt: secret is required, please set jwt.secret in config")
	}

	ttlMin := cfg.GetInt("jwt.ttl", 60)
	algName := cfg.GetString("jwt.alg", "HS256")

	var alg gojwt.SigningMethod
	switch algName {
	case "HS384":
		alg = gojwt.SigningMethodHS384
	case "HS512":
		alg = gojwt.SigningMethodHS512
	default:
		alg = gojwt.SigningMethodHS256
	}

	return &jwtService{
		secret:     []byte(secret),
		ttl:        time.Duration(ttlMin) * time.Minute,
		signingAlg: alg,
	}, nil
}

// GenerateToken 根据 MapClaims 生成签名 Token，自动注入 exp 字段。
func (j *jwtService) GenerateToken(claims gojwt.MapClaims) (string, error) {
	// 自动设置过期时间（调用方传入的 exp 会被覆盖）
	claims["exp"] = gojwt.NewNumericDate(time.Now().Add(j.ttl))
	if _, ok := claims["iat"]; !ok {
		claims["iat"] = gojwt.NewNumericDate(time.Now())
	}

	token := gojwt.NewWithClaims(j.signingAlg, claims)
	return token.SignedString(j.secret)
}

// ParseToken 解析并验证 Token，返回 MapClaims。
func (j *jwtService) ParseToken(tokenStr string) (gojwt.MapClaims, error) {
	token, err := gojwt.ParseWithClaims(
		tokenStr,
		gojwt.MapClaims{},
		func(t *gojwt.Token) (any, error) {
			if t.Method.Alg() != j.signingAlg.Alg() {
				return nil, errors.New("jwt: unexpected signing algorithm")
			}
			return j.secret, nil
		},
	)
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("jwt: token is invalid")
	}

	claims, ok := token.Claims.(gojwt.MapClaims)
	if !ok {
		return nil, errors.New("jwt: failed to parse claims")
	}
	return claims, nil
}

// RefreshToken 在 Token 仍有效时刷新过期时间，返回新 Token。
func (j *jwtService) RefreshToken(tokenStr string) (string, error) {
	claims, err := j.ParseToken(tokenStr)
	if err != nil {
		return "", err
	}

	// 删除旧时间戳，由 GenerateToken 重新注入
	delete(claims, "exp")
	delete(claims, "iat")
	return j.GenerateToken(claims)
}

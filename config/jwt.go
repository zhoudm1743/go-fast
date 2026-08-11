package config

import fwconfig "github.com/zhoudm1743/go-fast-framework/config"

func init() {
	fwconfig.Add("jwt", map[string]any{
		// JWT 签名密钥（生产环境请更换为随机字符串）
		"secret": "change-me-to-a-random-secret",
		// Token 有效期（分钟）
		"ttl": 60,
		// 签名算法：HS256 / HS384 / HS512
		"alg": "HS256",
		// 命名 Guard（可选），每个 Guard 独立配置密钥、有效期、算法
		// "guards": map[string]any{
		//     "platform": map[string]any{
		//         "secret": "platform-guard-secret",
		//         "ttl":    120,
		//         "alg":    "HS512",
		//     },
		// },
	})
}

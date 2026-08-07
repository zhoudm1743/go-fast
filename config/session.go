package config

import fwconfig "github.com/zhoudm1743/go-fast/framework/config"

func init() {
	fwconfig.Add("session", map[string]any{
		// 会话有效期（秒），默认 2 小时
		"lifetime": 7200,
		// 保存 Session ID 的 Cookie 名称
		"cookie": "go_fast_session",
	})
}

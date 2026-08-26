package config

import fwconfig "github.com/zhoudm1743/go-fast-framework/config"

func init() {
	fwconfig.Add("test", map[string]any{
		// 问候语前缀（services/test 示例服务使用）
		"greeting_prefix": "Hello",
	})
}

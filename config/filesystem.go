package config

import fwconfig "github.com/zhoudm1743/go-fast/framework/config"

func init() {
	fwconfig.Add("filesystem", map[string]any{
		// 默认磁盘
		"default": "local",
		// 磁盘配置
		"disks": map[string]any{
			"local": map[string]any{
				// 驱动类型：local
				"driver": "local",
				// 存储根目录（相对于项目根）
				"root": "storage/app",
				// 访问 URL 前缀
				"url": "/storage",
			},
		},
	})
}

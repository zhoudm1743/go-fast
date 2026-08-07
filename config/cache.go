package config

import fwconfig "github.com/zhoudm1743/go-fast/framework/config"

func init() {
	fwconfig.Add("cache", map[string]any{
		// 缓存驱动：memory / redis
		"driver": "memory",

		// 内存缓存配置
		"memory": map[string]any{
			// 分片数量
			"shard_count": 32,
			// 清理间隔（秒）
			"clean_interval": 60,
		},

		// Redis 缓存配置（driver=redis 时使用）
		"redis": map[string]any{
			"host":     "127.0.0.1",
			"port":     6379,
			"password": "",
			"db":       0,
			"prefix":   "",
		},
	})
}

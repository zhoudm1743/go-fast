package config

import fwconfig "github.com/zhoudm1743/go-fast-framework/config"

func init() {
	fwconfig.Add("cache", map[string]any{
		// 缓存驱动：memory / file / redis
		// 降级链：redis → file → memory（连接或目录不可用时自动降级）
		"driver": "file",

		// 内存缓存配置（最终兜底，总是可用）
		"memory": map[string]any{
			// 分片数量
			"shard_count": 32,
			// 清理间隔（秒）
			"clean_interval": 60,
		},

		// 文件缓存配置（driver=file 时使用，跨重启持久化，无外部依赖）
		"file": map[string]any{
			// 缓存目录（相对于项目根）
			"path": "storage/cache",
			// 过期文件清理间隔（秒）
			"clean_interval": 600,
		},

		// Redis 缓存配置（driver=redis 时在 config.yaml 中配置）
		// 注意：此处不要预填 host，否则框架会认为已启用 Redis 并尝试连接
		"redis": map[string]any{
			"host":     "",
			"port":     6379,
			"password": "",
			"db":       0,
			"prefix":   "",
		},
	})
}

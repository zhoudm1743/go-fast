package config

import fwconfig "github.com/zhoudm1743/go-fast/framework/config"

func init() {
	fwconfig.Add("log", map[string]any{
		// 输出模式：console / file / hybrid（控制台 + 文件）
		"mode": "hybrid",
		// 日志级别：debug / info / warn / error / fatal / panic
		"level": "debug",
		// 控制台格式：color / json / text
		"format": "color",
		// 文件格式：json / text
		"file_format": "json",
		// 日志输出路径（mode=file 或 hybrid 时生效）
		"output_path": "storage/logs/app.log",
		// 时间戳格式
		"timestamp_format": "2006-01-02 15:04:05",
		// 单个日志文件最大大小（MB）
		"max_size": 100,
		// 保留的旧日志文件数量
		"max_backups": 5,
		// 旧日志文件保留天数
		"max_age": 30,
		// 是否压缩旧日志文件
		"compress": false,
	})
}

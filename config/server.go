package config

import fwconfig "github.com/zhoudm1743/go-fast/framework/config"

func init() {
	fwconfig.Add("server", map[string]any{
		// 应用名称
		"name": "GoFast",
		// HTTP 引擎：gin / fiber
		"driver": "gin",
		// 监听地址
		"host": "0.0.0.0",
		// 监听端口
		"port": 3000,
		// 运行模式：debug / release
		"mode": "debug",
		// 读超时（秒）
		"read_timeout_sec": 30,
		// 写超时（秒）
		"write_timeout_sec": 30,
		// 空闲超时（秒）
		"idle_timeout_sec": 120,
		// 优雅关闭超时（秒）
		"shutdown_timeout_sec": 10,
		// 是否启用 Prefork（仅 fiber 支持）
		"prefork": false,
		// 请求体大小限制（MB）
		"body_limit_mb": 10,
		// CORS 允许的来源列表
		"cors_allow_origins": []string{"*"},
	})
}

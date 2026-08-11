package config

import fwconfig "github.com/zhoudm1743/go-fast-framework/config"

func init() {
	fwconfig.Add("database", map[string]any{
		// 默认连接名称
		"default": "main",

		// 数据库连接配置（每个连接一个独立 map）
		"connections": map[string]any{
			"main": map[string]any{
				// 驱动注册名（默认 "gormdriver"，对应 GORM 实现）
				"driver": "gormdriver",
				// 数据库引擎：sqlite / mysql / postgres / mssql
				"engine":   "sqlite",
				"database": "database/gofast.db",

				// MySQL / PostgreSQL 连接参数（SQLite 忽略）
				"host":     "localhost",
				"port":     3306,
				"username": "root",
				"password": "",
				"charset":  "utf8mb4",
				"loc":      "Local",
				"ssl_mode": "",

				// 连接池
				"max_idle_conns":     10,
				"max_open_conns":     100,
				"conn_max_lifetime":  60,  // 分钟
				"conn_max_idle_time": 30,  // 分钟

				// GORM 日志级别：silent / error / warn / info
				"log_level": "info",

				// 慢查询阈值（毫秒），0 表示不记录
				"slow_threshold": 200,
			},

			// 只读副本示例
			// "replica": map[string]any{
			//     "driver":   "gormdriver",
			//     "engine":   "mysql",
			//     "host":     "10.0.0.2",
			//     "port":     3306,
			//     "username": "readonly",
			//     "password": "secret",
			//     "database": "gofast",
			//     "charset":  "utf8mb4",
			// },
		},
	})
}

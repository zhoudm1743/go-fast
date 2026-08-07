package config

import fwconfig "github.com/zhoudm1743/go-fast/framework/config"

func init() {
	fwconfig.Add("database", map[string]any{
		// 数据库驱动：sqlite / mysql / postgres / mssql
		"driver": "sqlite",
		// SQLite 数据库文件路径
		"database": "database/gofast.db",
		// MySQL / PostgreSQL 连接参数（使用 SQLite 时忽略）
		// "host":     "127.0.0.1",
		// "port":     3306,
		// "username": "root",
		// "password": "secret",
		// "charset":  "utf8mb4",
		// 连接池配置
		"max_idle_conns":     10,
		"max_open_conns":     100,
		"conn_max_lifetime":  60,  // 分钟
		"conn_max_idle_time": 30,  // 分钟
	})
}

package config

import fwconfig "github.com/zhoudm1743/go-fast/framework/config"

func init() {
	fwconfig.Add("grpc", map[string]any{
		// gRPC 监听地址
		"host": "0.0.0.0",
		// gRPC 监听端口
		"port": 9000,
		// 运行模式：debug（开启 reflection）/ release
		"mode": "debug",
		// 单次接收消息最大大小（MB）
		"max_recv_msg_size_mb": 4,
		// 单次发送消息最大大小（MB）
		"max_send_msg_size_mb": 4,
		// 单个连接最大存活时间（秒）
		"max_conn_age_sec": 300,
		// 超过 max_conn_age 后的宽限期（秒）
		"max_conn_age_grace_sec": 5,
		// Keepalive ping 间隔（秒）
		"keepalive_time_sec": 60,
		// Keepalive 超时（秒）
		"keepalive_timeout_sec": 20,
		// TLS 配置（可选，留空则不启用）
		"tls": map[string]any{
			"cert_file": "",
			"key_file":  "",
		},
	})
}

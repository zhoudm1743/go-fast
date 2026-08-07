// Package config 提供 Go 方式的应用程序配置。
// 每个文件通过 init() 函数注册一个配置命名空间，配置值作为默认值层，
// 可被 config/config.yaml 中的同名键覆盖。
package config

import fwconfig "github.com/zhoudm1743/go-fast/framework/config"

func init() {
	fwconfig.Add("app", map[string]any{
		// 应用名称
		"name": "GoFast",
		// 运行环境：production / staging / local
		"env": "production",
		// 调试模式
		"debug": false,
	})
}

package facades

import "go-fast/framework/contracts"

// Config 获取配置服务实例。
func Config() contracts.Config {
	return App().MustMake("config").(contracts.Config)
}

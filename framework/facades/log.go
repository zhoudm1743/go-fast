package facades

import "go-fast/framework/contracts"

// Log 获取日志服务实例。
func Log() contracts.Log {
	return App().MustMake("log").(contracts.Log)
}

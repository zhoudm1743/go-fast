package contracts

// Config 配置服务契约。
type Config interface {
	// Env 读取环境变量，支持默认值。
	Env(key string, defaultValue ...any) any
	// Get 读取配置值，支持点号路径（如 "database.host"），支持默认值。
	Get(key string, defaultValue ...any) any
	// GetString 读取字符串配置。
	GetString(key string, defaultValue ...string) string
	// GetInt 读取整数配置。
	GetInt(key string, defaultValue ...int) int
	// GetBool 读取布尔配置。
	GetBool(key string, defaultValue ...bool) bool
	// Set 运行时设置配置值（不持久化到文件）。
	Set(key string, value any)
}

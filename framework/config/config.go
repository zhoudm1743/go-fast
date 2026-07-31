package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/spf13/viper"

	"github.com/zhoudm1743/go-fast/framework/contracts"
)

// 编译期保证 configImpl 实现了 contracts.Config 接口。
var _ contracts.Config = (*configImpl)(nil)

// configImpl 实现 contracts.Config 接口，包装 viper。
type configImpl struct {
	viper *viper.Viper
	mu    sync.RWMutex // 保护 Set/SetDefaults 与 Get 系列之间的并发读写
}

// NewConfig 从配置文件创建 Config 实例。
func NewConfig(path string) (contracts.Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("[GoFast] 读取配置文件失败: %w", err)
	}
	return &configImpl{viper: v}, nil
}

// Env 读取操作系统环境变量，支持默认值。
// 与 Get 系列不同，Env 直接读取 os.Getenv，不经过配置文件。
func (c *configImpl) Env(key string, defaultValue ...any) any {
	val := os.Getenv(key)
	if val != "" {
		return val
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return nil
}

func (c *configImpl) Get(key string, defaultValue ...any) any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if !c.viper.IsSet(key) && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return c.viper.Get(key)
}

func (c *configImpl) GetString(key string, defaultValue ...string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if !c.viper.IsSet(key) && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return c.viper.GetString(key)
}

func (c *configImpl) GetInt(key string, defaultValue ...int) int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if !c.viper.IsSet(key) && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return c.viper.GetInt(key)
}

func (c *configImpl) GetBool(key string, defaultValue ...bool) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if !c.viper.IsSet(key) && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return c.viper.GetBool(key)
}

func (c *configImpl) GetFloat64(key string, defaultValue ...float64) float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if !c.viper.IsSet(key) && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return c.viper.GetFloat64(key)
}

func (c *configImpl) GetStringSlice(key string, defaultValue ...[]string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if !c.viper.IsSet(key) && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return c.viper.GetStringSlice(key)
}

func (c *configImpl) GetStringMap(key string, defaultValue ...map[string]any) map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if !c.viper.IsSet(key) {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return nil
	}
	return c.viper.GetStringMap(key)
}

// Set 运行时设置配置值（不持久化到文件）。
func (c *configImpl) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.viper.Set(key, value)
}

// SetDefaults 批量设置默认值，底层调用 viper.SetDefault。
// 仅在用户未通过配置文件或 Set() 明确设置时生效，不会覆盖已有配置。
func (c *configImpl) SetDefaults(defaults map[string]any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key, val := range defaults {
		c.viper.SetDefault(key, val)
	}
}

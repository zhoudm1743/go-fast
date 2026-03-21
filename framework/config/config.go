package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// configImpl 实现 contracts.Config 接口，包装 viper。
type configImpl struct {
	viper *viper.Viper
}

// NewConfig 从配置文件创建 Config 实例。
func NewConfig(path string) (*configImpl, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("[GoFast] read config failed: %w", err)
	}
	return &configImpl{viper: v}, nil
}

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
	if !c.viper.IsSet(key) && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return c.viper.Get(key)
}

func (c *configImpl) GetString(key string, defaultValue ...string) string {
	if !c.viper.IsSet(key) && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return c.viper.GetString(key)
}

func (c *configImpl) GetInt(key string, defaultValue ...int) int {
	if !c.viper.IsSet(key) && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return c.viper.GetInt(key)
}

func (c *configImpl) GetBool(key string, defaultValue ...bool) bool {
	if !c.viper.IsSet(key) && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return c.viper.GetBool(key)
}

func (c *configImpl) Set(key string, value any) {
	c.viper.Set(key, value)
}

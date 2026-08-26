package test

import (
	"fmt"

	appcontracts "github.com/zhoudm1743/go-fast/services/contracts"
	"github.com/zhoudm1743/go-fast-framework/contracts"
)

type testService struct {
	prefix string
	cache  contracts.Cache
	log    contracts.Log
}

// NewTestService 创建 Test 服务实例。
func NewTestService(cfg contracts.Config, cache contracts.Cache, log contracts.Log) (appcontracts.Test, error) {
	return &testService{
		prefix: cfg.GetString("test.greeting_prefix", "Hello"),
		cache:  cache,
		log:    log,
	}, nil
}

func (s *testService) Greet(name string) string {
	if name == "" {
		name = "World"
	}
	return fmt.Sprintf("%s, %s!", s.prefix, name)
}

func (s *testService) Status() map[string]any {
	return map[string]any{
		"service": "test",
		"prefix":  s.prefix,
	}
}

// Cache contracts.Cache — 供 Boot 阶段预热或后续扩展使用。
func (s *testService) Cache() contracts.Cache { return s.cache }

// Log contracts.Log — 供 Boot 阶段日志输出。
func (s *testService) Log() contracts.Log { return s.log }

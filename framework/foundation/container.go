package foundation

import (
	"fmt"
	"sync"
)

// Container 服务容器接口（IoC）。
type Container interface {
	// Bind 绑定工厂函数，每次 Make 都会调用工厂创建新实例。
	Bind(key string, factory func(app Application) (any, error))
	// Singleton 绑定单例工厂，仅首次 Make 时调用工厂，后续返回缓存实例。
	Singleton(key string, factory func(app Application) (any, error))
	// Instance 直接绑定一个已创建的实例到容器。
	Instance(key string, instance any)
	// Make 根据 key 解析服务，返回实例和可能的错误。
	Make(key string) (any, error)
	// MustMake 解析服务，失败时 panic。
	MustMake(key string) any
	// Bound 检查 key 是否已绑定。
	Bound(key string) bool
	// Flush 清空所有绑定和缓存（测试用）。
	Flush()
}

// bindingType 绑定类型
type bindingType int

const (
	bindFactory   bindingType = iota // 每次创建新实例
	bindSingleton                    // 首次创建后缓存
	bindInstance                     // 直接绑定实例
)

// binding 单个绑定项
type binding struct {
	typ      bindingType
	factory  func(app Application) (any, error)
	instance any
	once     sync.Once
	err      error // singleton 首次构建时的错误
}

// container Container 接口的默认实现
type container struct {
	mu       sync.RWMutex
	bindings map[string]*binding
	app      Application // 延迟设置，由 application 在创建后注入
}

// newContainer 创建一个空容器
func newContainer() *container {
	return &container{
		bindings: make(map[string]*binding),
	}
}

// setApp 由 application 创建后调用，将 app 引用注入容器（供 factory 回调使用）
func (c *container) setApp(app Application) {
	c.app = app
}

func (c *container) Bind(key string, factory func(app Application) (any, error)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.bindings[key] = &binding{
		typ:     bindFactory,
		factory: factory,
	}
}

func (c *container) Singleton(key string, factory func(app Application) (any, error)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.bindings[key] = &binding{
		typ:     bindSingleton,
		factory: factory,
	}
}

func (c *container) Instance(key string, instance any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.bindings[key] = &binding{
		typ:      bindInstance,
		instance: instance,
	}
}

func (c *container) Make(key string) (any, error) {
	c.mu.RLock()
	b, ok := c.bindings[key]
	c.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("[GoFast] service not found: %s", key)
	}

	switch b.typ {
	case bindInstance:
		return b.instance, nil

	case bindFactory:
		return b.factory(c.app)

	case bindSingleton:
		b.once.Do(func() {
			b.instance, b.err = b.factory(c.app)
		})
		if b.err != nil {
			return nil, fmt.Errorf("[GoFast] singleton %q init failed: %w", key, b.err)
		}
		return b.instance, nil

	default:
		return nil, fmt.Errorf("[GoFast] unknown binding type for key: %s", key)
	}
}

func (c *container) MustMake(key string) any {
	instance, err := c.Make(key)
	if err != nil {
		panic(err)
	}
	return instance
}

func (c *container) Bound(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.bindings[key]
	return ok
}

func (c *container) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.bindings = make(map[string]*binding)
}

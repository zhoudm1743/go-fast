package foundation

import (
	"fmt"
	"path/filepath"
	"sync"
)

const Version = "0.1.0"

// Application 应用实例接口，嵌入 Container。
type Application interface {
	Container

	// Boot 按声明顺序依次执行所有 Provider 的 Register，然后依次执行 Boot。
	Boot()
	// SetProviders 设置服务提供者列表（应在 Boot 之前调用）。
	SetProviders(providers []ServiceProvider)
	// BasePath 返回应用根目录；传入子路径时拼接返回。
	BasePath(path ...string) string
	// StoragePath 返回 storage 目录；传入子路径时拼接返回。
	StoragePath(path ...string) string
	// Version 返回框架版本号。
	Version() string
	// IsBooted 是否已引导完成。
	IsBooted() bool
	// Shutdown 优雅关闭，按注册逆序执行 shutdown hooks。
	Shutdown()
	// OnShutdown 注册关闭回调。
	OnShutdown(hook func())
}

// deferredEntry 将一个 DeferredProvider 与 sync.Once 绑定，保证线程安全地只初始化一次。
type deferredEntry struct {
	provider DeferredProvider
	once     sync.Once
	err      error
}

// application Application 接口的默认实现
type application struct {
	*container
	basePath      string
	providers     []ServiceProvider
	booted        bool
	shutdownHooks []func()
	mu            sync.Mutex
	deferredMap   map[string]*deferredEntry // service key → deferred entry
}

// NewApplication 创建一个新的 Application 实例。
// basePath 为应用根目录（通常为 "."）。
func NewApplication(basePath string) Application {
	c := newContainer()
	app := &application{
		container: c,
		basePath:  basePath,
	}
	// 将 app 自身注入容器，供 factory 回调中使用
	c.setApp(app)
	// 把 app 自己也注册到容器，方便 facades.App() 之外的场景使用
	app.Instance("app", app)
	return app
}

func (a *application) SetProviders(providers []ServiceProvider) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.providers = providers
}

func (a *application) Boot() {
	// ── 幂等检查：已引导则直接返回 ──────────────────
	a.mu.Lock()
	if a.booted {
		a.mu.Unlock()
		return
	}

	a.deferredMap = make(map[string]*deferredEntry)

	// 分离即时 Provider 和延迟 Provider（在锁内完成，避免竞态）
	var immediate []ServiceProvider
	for _, p := range a.providers {
		if dp, ok := p.(DeferredProvider); ok {
			entry := &deferredEntry{provider: dp}
			for _, key := range dp.DeferredServices() {
				a.deferredMap[key] = entry
			}
		} else {
			immediate = append(immediate, p)
		}
	}
	// 释放锁：provider 的 Register/Boot 可能回调 Make、OnShutdown 等方法，
	// 若持锁调用会导致死锁。
	a.mu.Unlock()

	// Phase 1: Register 所有即时 Provider
	for _, p := range immediate {
		p.Register(a)
	}

	// Phase 2: Boot 所有即时 Provider
	for _, p := range immediate {
		if err := p.Boot(a); err != nil {
			panic(fmt.Sprintf("[GoFast] boot provider failed: %v", err))
		}
	}

	// 标记引导完成
	a.mu.Lock()
	a.booted = true
	a.mu.Unlock()
}

// bootDeferredIfNeeded 在首次 Make 延迟服务时触发其 Provider 的 Register + Boot。
// 使用 sync.Once 确保同一 Provider 只初始化一次，并发安全。
func (a *application) bootDeferredIfNeeded(key string) {
	a.mu.Lock()
	entry, ok := a.deferredMap[key]
	a.mu.Unlock()
	if !ok {
		return
	}

	entry.once.Do(func() {
		entry.provider.Register(a)
		entry.err = entry.provider.Boot(a)
	})
	if entry.err != nil {
		panic(fmt.Sprintf("[GoFast] boot deferred provider failed: %v", entry.err))
	}
}

// Make 解析服务。若 key 属于延迟 Provider，则先触发其初始化。
func (a *application) Make(key string) (any, error) {
	a.bootDeferredIfNeeded(key)
	return a.container.Make(key)
}

// MustMake 解析服务，失败时 panic。覆写以确保走 application.Make 的延迟逻辑。
func (a *application) MustMake(key string) any {
	v, err := a.Make(key)
	if err != nil {
		panic(err)
	}
	return v
}

// Bound 检查 key 是否已绑定或由延迟 Provider 声明。
func (a *application) Bound(key string) bool {
	a.mu.Lock()
	_, deferred := a.deferredMap[key]
	a.mu.Unlock()
	if deferred {
		return true
	}
	return a.container.Bound(key)
}

func (a *application) BasePath(path ...string) string {
	if len(path) == 0 {
		return a.basePath
	}
	return filepath.Join(a.basePath, filepath.Join(path...))
}

func (a *application) StoragePath(path ...string) string {
	base := filepath.Join(a.basePath, "storage")
	if len(path) == 0 {
		return base
	}
	return filepath.Join(base, filepath.Join(path...))
}

func (a *application) Version() string {
	return Version
}

func (a *application) IsBooted() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.booted
}

func (a *application) OnShutdown(hook func()) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.shutdownHooks = append(a.shutdownHooks, hook)
}

func (a *application) Shutdown() {
	a.mu.Lock()
	hooks := make([]func(), len(a.shutdownHooks))
	copy(hooks, a.shutdownHooks)
	a.mu.Unlock()

	// 按注册逆序执行
	for i := len(hooks) - 1; i >= 0; i-- {
		hooks[i]()
	}
}

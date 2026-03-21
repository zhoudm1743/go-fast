package foundation

import (
	"sync"
	"sync/atomic"
	"testing"
)

// ── Container 测试 ──────────────────────────────────────────────────

func TestContainer_Bind(t *testing.T) {
	app := NewApplication(".")
	callCount := 0
	app.Bind("svc", func(a Application) (any, error) {
		callCount++
		return callCount, nil
	})

	v1, _ := app.Make("svc")
	v2, _ := app.Make("svc")
	if v1.(int) != 1 || v2.(int) != 2 {
		t.Fatalf("Bind should create new instance each time, got %v %v", v1, v2)
	}
}

func TestContainer_Singleton(t *testing.T) {
	app := NewApplication(".")
	callCount := 0
	app.Singleton("svc", func(a Application) (any, error) {
		callCount++
		return "instance", nil
	})

	v1, _ := app.Make("svc")
	v2, _ := app.Make("svc")
	if v1 != v2 {
		t.Fatal("Singleton should return same instance")
	}
	if callCount != 1 {
		t.Fatalf("Singleton factory should be called once, got %d", callCount)
	}
}

func TestContainer_Instance(t *testing.T) {
	app := NewApplication(".")
	obj := &struct{ Name string }{Name: "test"}
	app.Instance("svc", obj)

	v, err := app.Make("svc")
	if err != nil {
		t.Fatal(err)
	}
	if v != obj {
		t.Fatal("Instance should return exact same pointer")
	}
}

func TestContainer_MustMake_Panic(t *testing.T) {
	app := NewApplication(".")
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("MustMake should panic for unknown key")
		}
	}()
	app.MustMake("nonexistent")
}

func TestContainer_Bound(t *testing.T) {
	app := NewApplication(".")
	if app.Bound("svc") {
		t.Fatal("should not be bound")
	}
	app.Instance("svc", 1)
	if !app.Bound("svc") {
		t.Fatal("should be bound")
	}
}

func TestContainer_Flush(t *testing.T) {
	app := NewApplication(".")
	app.Instance("svc", 1)
	app.Flush()
	if app.Bound("svc") {
		t.Fatal("Flush should clear all bindings")
	}
}

func TestContainer_Singleton_ConcurrentSafe(t *testing.T) {
	app := NewApplication(".")
	var count int64
	app.Singleton("svc", func(a Application) (any, error) {
		atomic.AddInt64(&count, 1)
		return "ok", nil
	})

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v := app.MustMake("svc")
			if v != "ok" {
				t.Errorf("unexpected value: %v", v)
			}
		}()
	}
	wg.Wait()

	if atomic.LoadInt64(&count) != 1 {
		t.Fatalf("singleton factory called %d times, expected 1", count)
	}
}

// ── Application 测试 ────────────────────────────────────────────────

type testProvider struct {
	registerOrder *[]string
	bootOrder     *[]string
	name          string
}

func (p *testProvider) Register(app Application) {
	*p.registerOrder = append(*p.registerOrder, p.name)
}

func (p *testProvider) Boot(app Application) error {
	*p.bootOrder = append(*p.bootOrder, p.name)
	return nil
}

func TestApplication_Boot_Order(t *testing.T) {
	var regOrder, bootOrder []string

	providers := []ServiceProvider{
		&testProvider{registerOrder: &regOrder, bootOrder: &bootOrder, name: "config"},
		&testProvider{registerOrder: &regOrder, bootOrder: &bootOrder, name: "log"},
		&testProvider{registerOrder: &regOrder, bootOrder: &bootOrder, name: "db"},
	}

	app := NewApplication(".")
	app.SetProviders(providers)
	app.Boot()

	// Register 全部先于 Boot
	expectedReg := []string{"config", "log", "db"}
	expectedBoot := []string{"config", "log", "db"}

	for i, v := range expectedReg {
		if regOrder[i] != v {
			t.Fatalf("Register order mismatch at %d: got %s, want %s", i, regOrder[i], v)
		}
	}
	for i, v := range expectedBoot {
		if bootOrder[i] != v {
			t.Fatalf("Boot order mismatch at %d: got %s, want %s", i, bootOrder[i], v)
		}
	}
}

func TestApplication_Boot_Idempotent(t *testing.T) {
	count := 0
	p := &countProvider{count: &count}
	app := NewApplication(".")
	app.SetProviders([]ServiceProvider{p})
	app.Boot()
	app.Boot() // 重复调用
	if count != 1 {
		t.Fatalf("Boot should be idempotent, register called %d times", count)
	}
}

type countProvider struct {
	count *int
}

func (p *countProvider) Register(app Application)   { *p.count++ }
func (p *countProvider) Boot(app Application) error { return nil }

func TestApplication_Shutdown_ReverseOrder(t *testing.T) {
	var order []string
	app := NewApplication(".")
	app.OnShutdown(func() { order = append(order, "first") })
	app.OnShutdown(func() { order = append(order, "second") })
	app.OnShutdown(func() { order = append(order, "third") })
	app.Shutdown()

	expected := []string{"third", "second", "first"}
	for i, v := range expected {
		if order[i] != v {
			t.Fatalf("Shutdown order mismatch at %d: got %s, want %s", i, order[i], v)
		}
	}
}

func TestApplication_BasePath(t *testing.T) {
	app := NewApplication("/app")
	if app.BasePath() != "/app" {
		t.Fatalf("unexpected: %s", app.BasePath())
	}
	got := app.BasePath("config.yaml")
	// filepath.Join 在不同 OS 下分隔符不同，只检查包含关系
	if got != "/app/config.yaml" && got != "\\app\\config.yaml" && got != "/app\\config.yaml" {
		t.Fatalf("unexpected: %s", got)
	}
}

func TestApplication_Version(t *testing.T) {
	app := NewApplication(".")
	if app.Version() == "" {
		t.Fatal("version should not be empty")
	}
}

// ── DeferredProvider 测试 ───────────────────────────────────────────

// deferredTestProvider 模拟延迟服务提供者
type deferredTestProvider struct {
	registered bool
	booted     bool
	keys       []string
	value      any
}

func (p *deferredTestProvider) Register(app Application) {
	p.registered = true
	for _, key := range p.keys {
		val := p.value
		app.Singleton(key, func(a Application) (any, error) {
			return val, nil
		})
	}
}

func (p *deferredTestProvider) Boot(app Application) error {
	p.booted = true
	return nil
}

func (p *deferredTestProvider) DeferredServices() []string {
	return p.keys
}

func TestDeferredProvider_NotBootedDuringBoot(t *testing.T) {
	dp := &deferredTestProvider{keys: []string{"lazy"}, value: "hello"}
	app := NewApplication(".")
	app.SetProviders([]ServiceProvider{dp})
	app.Boot()

	if dp.registered {
		t.Fatal("deferred provider should NOT be registered during Boot")
	}
	if dp.booted {
		t.Fatal("deferred provider should NOT be booted during Boot")
	}
}

func TestDeferredProvider_BootedOnFirstMake(t *testing.T) {
	dp := &deferredTestProvider{keys: []string{"lazy"}, value: "hello"}
	app := NewApplication(".")
	app.SetProviders([]ServiceProvider{dp})
	app.Boot()

	v, err := app.Make("lazy")
	if err != nil {
		t.Fatalf("Make failed: %v", err)
	}
	if v != "hello" {
		t.Fatalf("expected 'hello', got %v", v)
	}
	if !dp.registered {
		t.Fatal("deferred provider should be registered after Make")
	}
	if !dp.booted {
		t.Fatal("deferred provider should be booted after Make")
	}
}

func TestDeferredProvider_MustMakeTriggersDeferred(t *testing.T) {
	dp := &deferredTestProvider{keys: []string{"lazy"}, value: 42}
	app := NewApplication(".")
	app.SetProviders([]ServiceProvider{dp})
	app.Boot()

	v := app.MustMake("lazy")
	if v != 42 {
		t.Fatalf("expected 42, got %v", v)
	}
	if !dp.registered || !dp.booted {
		t.Fatal("deferred provider should be fully initialized after MustMake")
	}
}

func TestDeferredProvider_MultipleKeys(t *testing.T) {
	dp := &deferredTestProvider{keys: []string{"svcA", "svcB"}, value: "shared"}
	app := NewApplication(".")
	app.SetProviders([]ServiceProvider{dp})
	app.Boot()

	// 通过第一个 key 触发
	v1, _ := app.Make("svcA")
	if v1 != "shared" {
		t.Fatalf("expected 'shared', got %v", v1)
	}

	// 第二个 key 也应该可用（同一 Provider 只初始化一次）
	v2, _ := app.Make("svcB")
	if v2 != "shared" {
		t.Fatalf("expected 'shared', got %v", v2)
	}
}

func TestDeferredProvider_Bound(t *testing.T) {
	dp := &deferredTestProvider{keys: []string{"lazy"}, value: "x"}
	app := NewApplication(".")
	app.SetProviders([]ServiceProvider{dp})
	app.Boot()

	// 未 Make 前也应返回 true
	if !app.Bound("lazy") {
		t.Fatal("Bound should return true for deferred key before Make")
	}

	// Make 后依然 true
	app.MustMake("lazy")
	if !app.Bound("lazy") {
		t.Fatal("Bound should return true for deferred key after Make")
	}
}

func TestDeferredProvider_OnlyBootedOnce(t *testing.T) {
	var regCount, bootCount int
	dp := &onceCountDeferredProvider{
		keys:      []string{"once"},
		regCount:  &regCount,
		bootCount: &bootCount,
	}
	app := NewApplication(".")
	app.SetProviders([]ServiceProvider{dp})
	app.Boot()

	app.MustMake("once")
	app.MustMake("once")
	app.MustMake("once")

	if regCount != 1 {
		t.Fatalf("Register called %d times, expected 1", regCount)
	}
	if bootCount != 1 {
		t.Fatalf("Boot called %d times, expected 1", bootCount)
	}
}

type onceCountDeferredProvider struct {
	keys      []string
	regCount  *int
	bootCount *int
}

func (p *onceCountDeferredProvider) Register(app Application) {
	*p.regCount++
	for _, key := range p.keys {
		app.Singleton(key, func(a Application) (any, error) {
			return "ok", nil
		})
	}
}
func (p *onceCountDeferredProvider) Boot(app Application) error {
	*p.bootCount++
	return nil
}
func (p *onceCountDeferredProvider) DeferredServices() []string {
	return p.keys
}

func TestDeferredProvider_MixedWithImmediate(t *testing.T) {
	var regOrder []string

	immediate := &testProvider{registerOrder: &regOrder, bootOrder: &regOrder, name: "config"}
	deferred := &deferredTestProvider{keys: []string{"lazy"}, value: "deferred_val"}

	app := NewApplication(".")
	app.SetProviders([]ServiceProvider{immediate, deferred})
	app.Boot()

	// immediate 已注册，deferred 未注册
	if len(regOrder) != 2 { // Register + Boot for config
		t.Fatalf("expected 2 order entries (reg+boot for config), got %d", len(regOrder))
	}
	if deferred.registered {
		t.Fatal("deferred should not be registered yet")
	}

	// 触发 deferred
	v := app.MustMake("lazy")
	if v != "deferred_val" {
		t.Fatalf("expected 'deferred_val', got %v", v)
	}
	if !deferred.registered || !deferred.booted {
		t.Fatal("deferred should be fully initialized after MustMake")
	}
}

func TestDeferredProvider_ConcurrentMake(t *testing.T) {
	var regCount int64
	dp := &atomicCountDeferredProvider{
		keys:     []string{"concurrent"},
		regCount: &regCount,
	}
	app := NewApplication(".")
	app.SetProviders([]ServiceProvider{dp})
	app.Boot()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v := app.MustMake("concurrent")
			if v != "ok" {
				t.Errorf("unexpected value: %v", v)
			}
		}()
	}
	wg.Wait()

	if atomic.LoadInt64(&regCount) != 1 {
		t.Fatalf("Register called %d times, expected 1", atomic.LoadInt64(&regCount))
	}
}

type atomicCountDeferredProvider struct {
	keys     []string
	regCount *int64
}

func (p *atomicCountDeferredProvider) Register(app Application) {
	atomic.AddInt64(p.regCount, 1)
	for _, key := range p.keys {
		app.Singleton(key, func(a Application) (any, error) {
			return "ok", nil
		})
	}
}
func (p *atomicCountDeferredProvider) Boot(app Application) error { return nil }
func (p *atomicCountDeferredProvider) DeferredServices() []string { return p.keys }

# GoFast 服务容器 API

> 服务容器（IoC Container）是 GoFast 的核心引擎，负责管理所有服务的创建、缓存与解析。
> `Application` 嵌入了 `Container`，因此 `facades.App()` 即可直接调用所有容器方法。

---

## 一、核心接口

### 1.1 Container

```go
type Container interface {
    Bind(key string, factory func(app Application) (any, error))
    Singleton(key string, factory func(app Application) (any, error))
    Instance(key string, instance any)
    Make(key string) (any, error)
    MustMake(key string) any
    Bound(key string) bool
    Flush()
}
```

### 1.2 Application

```go
type Application interface {
    Container   // 嵌入容器，直接拥有全部容器方法

    Boot()
    SetProviders(providers []ServiceProvider)
    BasePath(path ...string) string
    StoragePath(path ...string) string
    Version() string
    IsBooted() bool
    Shutdown()
    OnShutdown(hook func())
}
```

---

## 二、绑定服务

### 2.1 Bind — 每次创建新实例

```go
app.Bind("mailer", func(app foundation.Application) (any, error) {
    cfg := app.MustMake("config").(contracts.Config)
    return NewMailer(cfg), nil
})
```

每次调用 `app.Make("mailer")` 都会执行工厂函数，返回**全新的实例**。

适用场景：
- 有状态的服务（如请求级别的上下文对象）
- 每次使用需要独立配置的服务

### 2.2 Singleton — 单例（懒加载）

```go
app.Singleton("config", func(app foundation.Application) (any, error) {
    return NewConfig(app.BasePath("config.yaml"))
})
```

首次 `Make` 时调用工厂函数创建实例，后续返回**同一个缓存实例**。底层使用 `sync.Once` 保证并发安全。

适用场景：
- 全局共享的服务（配置、日志、数据库连接、缓存等）
- 创建成本较高的资源

### 2.3 Instance — 直接绑定已有实例

```go
app.Instance("app", app)           // 绑定自身
app.Instance("custom", myObject)   // 绑定已创建的对象
```

直接将一个已存在的实例注册到容器中，无需工厂函数。

适用场景：
- 在容器外部已经创建好的对象
- 测试时注入 Mock 对象

---

## 三、解析服务

### 3.1 Make — 安全解析

```go
instance, err := app.Make("config")
if err != nil {
    // key 不存在或工厂函数返回错误
    log.Fatal(err)
}
cfg := instance.(contracts.Config)
```

返回 `(any, error)`，需要自行类型断言。

### 3.2 MustMake — 解析或 Panic

```go
cfg := app.MustMake("config").(contracts.Config)
```

内部调用 `Make`，如果出错则 **panic**。适合在确定服务一定存在时使用（如 Facade 内部）。

### 3.3 Bound — 检查是否已绑定

```go
if app.Bound("redis") {
    redis := app.MustMake("redis")
    // ...
}
```

返回 `true` 表示 key 已通过 `Bind`、`Singleton` 或 `Instance` 注册，或者由某个 `DeferredProvider` 声明。

---

## 四、延迟服务提供者

实现 `DeferredProvider` 接口的 Provider 不会在 `Boot()` 阶段立即执行，而是等到首次 `Make` 其声明的 key 时才自动触发 `Register + Boot`：

```go
type DeferredProvider interface {
    ServiceProvider
    DeferredServices() []string   // 该 Provider 提供的服务 key 列表
}
```

示例：

```go
type RedisServiceProvider struct{}

func (p *RedisServiceProvider) DeferredServices() []string {
    return []string{"redis"}
}

func (p *RedisServiceProvider) Register(app foundation.Application) {
    app.Singleton("redis", func(app foundation.Application) (any, error) {
        cfg := app.MustMake("config").(contracts.Config)
        return NewRedisClient(cfg)
    })
}

func (p *RedisServiceProvider) Boot(app foundation.Application) error {
    return nil
}
```

注册后，只有当代码中首次执行 `app.Make("redis")` 或 `facades.App().MustMake("redis")` 时，才会触发 `Register → Boot`，实现按需加载。

---

## 五、路径辅助方法

```go
app.BasePath()                  // 返回应用根目录，如 "."
app.BasePath("config.yaml")    // 返回 "./config.yaml"
app.BasePath("storage", "logs") // 返回 "./storage/logs"

app.StoragePath()               // 返回 "./storage"
app.StoragePath("logs", "app.log") // 返回 "./storage/logs/app.log"
```

---

## 六、生命周期管理

### 6.1 OnShutdown — 注册关闭钩子

```go
app.OnShutdown(func() {
    fmt.Println("cleaning up...")
    db.Close()
})
```

支持注册多个钩子。钩子在 `Shutdown()` 时按**注册逆序**执行（后注册的先执行），确保依赖关系正确：

```
注册顺序: db.OnShutdown → http.OnShutdown → cache.OnShutdown
执行顺序: cache 关闭 → http 关闭 → db 关闭
```

### 6.2 Shutdown — 优雅关闭

```go
app.Shutdown()
```

依次（逆序）执行所有通过 `OnShutdown` 注册的钩子。内置的 Provider 已自动注册关闭钩子：

| Provider | 关闭行为 |
|----------|---------|
| `database.ServiceProvider` | 关闭数据库连接 |
| `http.ServiceProvider` | 优雅关闭 HTTP 服务器 |
| `cache.ServiceProvider` | 停止内存缓存 GC |

### 6.3 Flush — 清空容器（测试用）

```go
app.Flush()     // 清空所有绑定和缓存
```

---

## 七、并发安全

- 所有容器操作均通过 `sync.RWMutex` 保护
- `Singleton` 使用 `sync.Once` 确保工厂函数只执行一次
- `DeferredProvider` 的 `Register + Boot` 使用 `sync.Once` 确保只初始化一次
- 100 个 goroutine 并发 `MustMake` 同一个 Singleton，工厂函数只会被调用 **1 次**

---

## 八、完整示例

```go
package main

import (
    "fmt"
    "github.com/zhoudm1743/go-fast/framework/facades"
    "github.com/zhoudm1743/go-fast/framework/foundation"
)

func main() {
    app := facades.App()

    // 绑定自定义服务
    app.Singleton("greeting", func(app foundation.Application) (any, error) {
        name := app.MustMake("config").(contracts.Config).GetString("app.name", "GoFast")
        return fmt.Sprintf("Hello from %s!", name), nil
    })

    // 解析
    msg := app.MustMake("greeting").(string)
    fmt.Println(msg) // Hello from GoFast!

    // 检查
    fmt.Println(app.Bound("greeting"))  // true
    fmt.Println(app.Bound("unknown"))   // false
}
```

---

## 九、内置服务 Key 一览

| Key | 类型 | Provider |
|-----|------|----------|
| `app` | `foundation.Application` | 自动注册 |
| `config` | `contracts.Config` | `config.ServiceProvider` |
| `log` | `contracts.Log` | `log.ServiceProvider` |
| `cache` | `contracts.Cache` | `cache.ServiceProvider` |
| `orm` | `contracts.Orm` | `database.ServiceProvider` |
| `storage` | `contracts.Storage` | `filesystem.ServiceProvider` |
| `validator` | `contracts.Validation` | `validation.ServiceProvider` |
| `route` | `contracts.Route` | `http.ServiceProvider` |


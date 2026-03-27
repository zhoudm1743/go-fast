# GoFast 编写自定义 ServiceProvider

> ServiceProvider（服务提供者）是 GoFast 组织服务注册与引导的核心机制。
> 无论是框架内置服务还是业务自定义服务，都通过 Provider 注册到容器中。

---

## 一、核心接口

```go
// foundation/provider.go

// ServiceProvider 服务提供者接口。
type ServiceProvider interface {
    // Register 将服务绑定到容器。
    // 此时其他服务可能尚未就绪，不可调用 MustMake。
    Register(app Application)
    // Boot 引导服务。
    // 所有 Provider 的 Register 均已执行完毕，可放心使用容器中的服务。
    Boot(app Application) error
}
```

### 执行顺序

```
providers 列表：[ConfigProvider, LogProvider, CacheProvider, MyProvider]

Phase 1 — Register（按声明顺序）
  ConfigProvider.Register(app)
  LogProvider.Register(app)
  CacheProvider.Register(app)
  MyProvider.Register(app)

Phase 2 — Boot（按声明顺序）
  ConfigProvider.Boot(app)
  LogProvider.Boot(app)
  CacheProvider.Boot(app)
  MyProvider.Boot(app)
```

**关键原则**：
- `Register` 中只做 **绑定**（`Bind` / `Singleton` / `Instance`），不要使用其他服务
- `Boot` 中可以安全地 **使用任何已注册的服务**（`MustMake` / `Make`）

---

## 二、编写一个 Provider — 完整示例

假设我们要开发一个 SMS 短信服务。

### 2.1 定义契约

```go
// contracts/sms.go
package contracts

// Sms 短信服务契约。
type Sms interface {
    Send(phone string, content string) error
    SendCode(phone string, code string) error
}
```

### 2.2 实现服务

```go
// sms/sms.go
package sms

import (
    "fmt"
    "github.com/zhoudm1743/go-fast/framework/contracts"
)

type smsService struct {
    apiKey   string
    endpoint string
}

func NewSmsService(cfg contracts.Config) (contracts.Sms, error) {
    apiKey := cfg.GetString("sms.api_key")
    if apiKey == "" {
        return nil, fmt.Errorf("[GoFast] sms.api_key is required")
    }
    return &smsService{
        apiKey:   apiKey,
        endpoint: cfg.GetString("sms.endpoint", "https://sms-api.example.com"),
    }, nil
}

func (s *smsService) Send(phone string, content string) error {
    // 调用短信 API ...
    return nil
}

func (s *smsService) SendCode(phone string, code string) error {
    return s.Send(phone, fmt.Sprintf("您的验证码是：%s", code))
}
```

### 2.3 编写 ServiceProvider

```go
// sms/service_provider.go
package sms

import (
    "github.com/zhoudm1743/go-fast/framework/contracts"
    "github.com/zhoudm1743/go-fast/framework/foundation"
)

// ServiceProvider SMS 服务提供者。
type ServiceProvider struct{}

func (sp *ServiceProvider) Register(app foundation.Application) {
    // 使用 Singleton 绑定——全局只需一个 SMS 客户端实例
    app.Singleton("sms", func(app foundation.Application) (any, error) {
        cfg := app.MustMake("config").(contracts.Config)
        return NewSmsService(cfg)
    })
}

func (sp *ServiceProvider) Boot(app foundation.Application) error {
    // 如果需要在引导阶段做额外工作（如预热连接），可在此处理
    // 大部分 Provider 的 Boot 只需返回 nil
    return nil
}
```

### 2.4 添加 Facade（可选）

```go
// facades/sms.go
package facades

import "github.com/zhoudm1743/go-fast/framework/contracts"

// Sms 获取短信服务实例。
func Sms() contracts.Sms {
    return App().MustMake("sms").(contracts.Sms)
}
```

### 2.5 注册到应用

在 `bootstrap/app.go` 的 `providers()` 函数中追加：

```go
func providers() []foundation.ServiceProvider {
    return []foundation.ServiceProvider{
        &config.ServiceProvider{},
        &log.ServiceProvider{},
        &cache.ServiceProvider{},
        &database.ServiceProvider{},
        &filesystem.ServiceProvider{},
        &validation.ServiceProvider{},
        &gohttp.ServiceProvider{},
        // ↓ 追加自定义 Provider
        &sms.ServiceProvider{},
    }
}
```

### 2.6 配置文件

```yaml
# config.yaml
sms:
  api_key: "your-api-key"
  endpoint: "https://sms-api.example.com"
```

### 2.7 使用

```go
// 通过 Facade
facades.Sms().SendCode("13800138000", "123456")

// 或直接从容器解析
svc := facades.App().MustMake("sms").(contracts.Sms)
svc.Send("13800138000", "Hello!")
```

---

## 三、绑定方式选择

| 方法 | 行为 | 适用场景 |
|------|------|---------|
| `Singleton` | 懒加载，首次 Make 时创建，后续缓存 | 全局共享（推荐默认方式） |
| `Bind` | 每次 Make 都创建新实例 | 有状态、请求级别的服务 |
| `Instance` | 直接绑定已创建的实例 | 测试 Mock、外部创建的对象 |

```go
// Singleton — 最常用
app.Singleton("sms", func(app foundation.Application) (any, error) {
    return NewSmsService(cfg)
})

// Bind — 每次获取新实例
app.Bind("request-context", func(app foundation.Application) (any, error) {
    return NewRequestContext()
})

// Instance — 直接绑定
app.Instance("build-info", &BuildInfo{Version: "1.0.0"})
```

---

## 四、关闭钩子

如果服务持有需要释放的资源（连接池、定时器等），应在 `Boot` 中注册关闭钩子：

```go
func (sp *ServiceProvider) Boot(app foundation.Application) error {
    app.OnShutdown(func() {
        if svc, err := app.Make("sms"); err == nil {
            if closer, ok := svc.(io.Closer); ok {
                _ = closer.Close()
            }
        }
    })
    return nil
}
```

关闭钩子按注册逆序执行，确保依赖正确释放：

```
注册顺序: config → log → db → sms
关闭顺序: sms → db → log → config
```

---

## 五、延迟 Provider（DeferredProvider）

对于不是每次请求都会用到的服务，可以实现 `DeferredProvider` 接口实现按需加载：

```go
// foundation/provider.go
type DeferredProvider interface {
    ServiceProvider
    DeferredServices() []string
}
```

### 示例

```go
type ServiceProvider struct{}

// DeferredServices 声明此 Provider 提供的服务 key。
func (sp *ServiceProvider) DeferredServices() []string {
    return []string{"sms"}
}

func (sp *ServiceProvider) Register(app foundation.Application) {
    app.Singleton("sms", func(app foundation.Application) (any, error) {
        cfg := app.MustMake("config").(contracts.Config)
        return NewSmsService(cfg)
    })
}

func (sp *ServiceProvider) Boot(app foundation.Application) error {
    return nil
}
```

**行为差异**：
- 普通 Provider：`Boot()` 时立即执行 `Register` + `Boot`
- DeferredProvider：`Boot()` 时跳过，首次 `Make("sms")` 时才触发 `Register` + `Boot`

**适用场景**：
- 服务初始化较重（如连接远程 API）
- 不是每个请求都会用到的服务
- 需要减少启动时间

---

## 六、Provider 依赖管理

Provider 的声明顺序决定了 `Register` 和 `Boot` 的执行顺序。**被依赖的 Provider 必须排在前面**：

```go
func providers() []foundation.ServiceProvider {
    return []foundation.ServiceProvider{
        &config.ServiceProvider{},      // 1. 无依赖
        &log.ServiceProvider{},         // 2. 依赖 config
        &database.ServiceProvider{},    // 3. 依赖 config + log
        &sms.ServiceProvider{},         // 4. 依赖 config
    }
}
```

### 安全规则

| 阶段 | 可以做 | 不可以做 |
|------|--------|---------|
| `Register` | `Bind` / `Singleton` / `Instance` | `Make` / `MustMake`（其他服务可能未注册） |
| `Boot` | 任何容器操作、`OnShutdown`、初始化逻辑 | — |

---

## 七、内置 Provider 参考

以下是框架内置 Provider 的实现，供编写自定义 Provider 时参考：

### config.ServiceProvider

```go
func (sp *ServiceProvider) Register(app foundation.Application) {
    app.Singleton("config", func(app foundation.Application) (any, error) {
        return NewConfig(app.BasePath("config.yaml"))
    })
}

func (sp *ServiceProvider) Boot(app foundation.Application) error {
    return nil
}
```

### database.ServiceProvider（带关闭钩子）

```go
func (sp *ServiceProvider) Register(app foundation.Application) {
    app.Singleton("orm", func(app foundation.Application) (any, error) {
        cfg := app.MustMake("config").(contracts.Config)
        log := app.MustMake("log").(contracts.Log)
        return NewOrm(cfg, log)
    })
}

func (sp *ServiceProvider) Boot(app foundation.Application) error {
    app.OnShutdown(func() {
        if o, err := app.Make("orm"); err == nil {
            if closer, ok := o.(contracts.Orm); ok {
                _ = closer.Close()
            }
        }
    })
    return nil
}
```

---

## 八、最佳实践

1. **一个 Provider 注册一组相关服务**，不要把所有服务塞进一个 Provider
2. **优先使用 `Singleton`**，除非服务确实需要每次创建新实例
3. **`Register` 只做绑定**，`Boot` 做初始化 — 不要搞反
4. **注册关闭钩子** 释放资源（数据库连接、定时器、文件句柄等）
5. **声明依赖顺序**，被依赖的 Provider 排在前面
6. **低频服务用 `DeferredProvider`**，减少启动开销
7. **定义 `contracts` 接口**，实现依赖倒置，便于测试和替换

---

## 九、相关文档

- [容器 API](container.md) — Bind / Singleton / Make 等方法详解
- [Facade 使用说明](facade.md) — 为自定义服务添加 Facade
- [插件开发指南](plugins.md) — 将 Provider 打包为可复用插件


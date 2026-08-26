# GoFast 编写自定义 ServiceProvider

> ServiceProvider（服务提供者）是 GoFast 组织服务注册与引导的核心机制。
> 框架接口（`foundation.ServiceProvider` 等）来自 **`go-fast-framework`** module；实现代码写在本仓库 `services/`。
> 架构说明见 [README.md](README.md)。

> 框架内置服务与业务自定义服务，都通过 Provider 注册到 IoC 容器，再经 Facade 访问。

本仓库已有完整参考实现：**`services/test/`**（契约 → 实现 → Provider → 测试 → 配置 → 注册 → 路由调用）。

---

## 一、核心接口

```go
// go-fast-framework/foundation/provider.go

// ServiceProvider 服务提供者接口。
type ServiceProvider interface {
    // Register 将服务绑定到容器。
    // 此时其他服务可能尚未就绪，不可调用 MustMake。
    Register(app Application)
    // Boot 引导服务。
    // 所有 Provider 的 Register 均已执行完毕，可安全使用容器中的服务。
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

- `Register` 中只做 **绑定**（`Bind` / `Singleton` / `Instance`），不要 `MustMake` 其他服务
- `Boot` 中可以 **使用任何已注册的服务**（`MustMake` / `Make`）、注册 `OnShutdown` 钩子

---

## 二、目录约定

自定义业务服务统一放在 **`services/`** 下：

```
services/
├── contracts/           # 业务契约接口（本仓库专用，与框架 contracts 区分）
│   └── sms.go
└── sms/                 # 具体服务
    ├── sms.go           # 实现
    ├── service_provider.go
    └── service_provider_test.go
```

| 路径 | 说明 |
|------|------|
| `services/contracts/` | 业务 interface，便于测试 Mock 与依赖倒置 |
| `services/<name>/` | 实现 + ServiceProvider + 单元测试 |
| `config/<name>.go` | Go 配置默认值 |
| `bootstrap/app.go` | `providers()` 末尾注册 Provider |

> ❌ 不要在项目根目录散落 `contracts/`、`facades/` 包。
> ❌ 不要新建 `services/facades/` — 见 [2.4 访问服务](#24-访问服务facade)。

---

## 三、完整示例 — SMS 短信服务

以下按推荐步骤编写；可直接对照本仓库 **`services/test/`** 对照阅读。

### 3.1 定义契约

业务契约放在 **`services/contracts/`**（不是框架的 `contracts` 包）：

```go
// services/contracts/sms.go
package contracts

// Sms 短信服务契约。
type Sms interface {
    Send(phone string, content string) error
    SendCode(phone string, code string) error
}
```

### 3.2 实现服务

```go
// services/sms/sms.go
package sms

import (
    "fmt"

    appcontracts "github.com/zhoudm1743/go-fast/services/contracts"
    "github.com/zhoudm1743/go-fast-framework/contracts"
)

type smsService struct {
    apiKey   string
    endpoint string
}

func NewSmsService(cfg contracts.Config) (appcontracts.Sms, error) {
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

**import 说明**：

| 包 | 用途 |
|----|------|
| `go-fast/services/contracts` | 本仓库业务接口（`Sms`） |
| `go-fast-framework/contracts` | 框架接口（`Config`、`Log`、`Cache` 等） |

### 3.3 编写 ServiceProvider

```go
// services/sms/service_provider.go
package sms

import (
    "github.com/zhoudm1743/go-fast-framework/contracts"
    "github.com/zhoudm1743/go-fast-framework/foundation"
)

type ServiceProvider struct{}

func (sp *ServiceProvider) Register(app foundation.Application) {
    app.Singleton("sms", func(app foundation.Application) (any, error) {
        cfg := app.MustMake("config").(contracts.Config)
        return NewSmsService(cfg)
    })
}

func (sp *ServiceProvider) Boot(app foundation.Application) error {
    // 可选：预热连接、写启动日志、注册 OnShutdown 等
    return nil
}
```

### 3.4 访问服务（Facade）

自定义服务注册到容器后，通过 **框架 Facade 入口** `facades.App().MustMake()` 解析：

```go
import (
    appcontracts "github.com/zhoudm1743/go-fast/services/contracts"
    "github.com/zhoudm1743/go-fast-framework/facades"
)

// 推荐写法
sms := facades.App().MustMake("sms").(appcontracts.Sms)
sms.SendCode("13800138000", "123456")
```

#### 为什么不用 `facades.Sms()` 或 `services/facades/`？

| 方式 | 说明 |
|------|------|
| `facades.Log()` / `facades.DB()` | 框架内置，函数定义在 `go-fast-framework/facades/` |
| `facades.App().MustMake("sms")` | **应用自定义服务的标准访问方式** |
| `facades.Sms()` 顶级函数 | 仅当在 **framework/facades/** 源码中新增文件时可用（插件合入框架） |

Go 模块不允许从应用项目向外部包的 `facades` 追加函数，因此**不需要**也**不应**创建 `services/facades/` 包装层。

### 3.5 注册到应用

在 `bootstrap/app.go` 的 `providers()` **末尾**追加（框架 Provider 之后）：

```go
import (
    smssvc "github.com/zhoudm1743/go-fast/services/sms"
    // ...
)

func providers() []foundation.ServiceProvider {
    return []foundation.ServiceProvider{
        &config.ServiceProvider{},
        &log.ServiceProvider{},
        &cache.ServiceProvider{},
        &tenant.ServiceProvider{},
        &database.ServiceProvider{},
        &filesystem.ServiceProvider{},
        &gojwt.ServiceProvider{},
        &gohttp.ServiceProvider{},
        &fast.ServiceProvider{},
        &goevent.ServiceProvider{},
        &goqueue.ServiceProvider{},
        &goschedule.ServiceProvider{},
        // ↓ 自定义 Provider 追加在框架 Provider 之后
        &smssvc.ServiceProvider{},
    }
}
```

### 3.6 配置文件

**Go 默认值**（`config/sms.go`）：

```go
package config

import fwconfig "github.com/zhoudm1743/go-fast-framework/config"

func init() {
    fwconfig.Add("sms", map[string]any{
        "api_key":  "",
        "endpoint": "https://sms-api.example.com",
    })
}
```

**YAML 覆盖**（`config/config.yaml`）：

```yaml
sms:
  api_key: "your-api-key"
  endpoint: "https://sms-api.example.com"
```

优先级：`Go 默认值 < YAML < 运行时 Set()`。

### 3.7 单元测试

```go
// services/sms/service_provider_test.go
package sms

import (
    "testing"

    _ "github.com/zhoudm1743/go-fast/config"

    appcontracts "github.com/zhoudm1743/go-fast/services/contracts"
    "github.com/zhoudm1743/go-fast-framework/cache"
    "github.com/zhoudm1743/go-fast-framework/config"
    "github.com/zhoudm1743/go-fast-framework/facades"
    "github.com/zhoudm1743/go-fast-framework/foundation"
    "github.com/zhoudm1743/go-fast-framework/log"
)

func TestServiceProvider_RegisterAndBoot(t *testing.T) {
    app := foundation.NewApplication(".")
    app.SetProviders([]foundation.ServiceProvider{
        &config.ServiceProvider{},
        &log.ServiceProvider{},
        &cache.ServiceProvider{},
        &ServiceProvider{},
    })
    app.Boot()
    facades.SetApp(app)

    svc, err := app.Make("sms")
    if err != nil {
        t.Fatal(err)
    }
    _ = svc.(appcontracts.Sms)
}
```

---

## 四、本仓库参考实现 — services/test

无需从零编写时，直接复制并改名：

| 文件 | 说明 |
|------|------|
| `services/contracts/test.go` | 业务契约 |
| `services/test/test.go` | 服务实现 |
| `services/test/service_provider.go` | Provider（Register + Boot） |
| `services/test/service_provider_test.go` | 单元测试 |
| `config/test.go` | 配置默认值 |
| `bootstrap/app.go` | `&testsvc.ServiceProvider{}` 注册 |
| `routes/app.go` | `facades.App().MustMake("test")` 调用示例 |

---

## 五、绑定方式选择

| 方法 | 行为 | 适用场景 |
|------|------|---------|
| `Singleton` | 懒加载，首次 Make 时创建，后续缓存 | 全局共享（**推荐默认**） |
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

// Instance — 直接绑定（测试 Mock 常用）
app.Instance("build-info", &BuildInfo{Version: "1.0.0"})
```

---

## 六、关闭钩子

服务持有需释放的资源（连接池、定时器等）时，在 `Boot` 中注册：

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

关闭钩子按注册 **逆序** 执行：

```
注册顺序: config → log → db → sms
关闭顺序: sms → db → log → config
```

---

## 七、延迟 Provider（DeferredProvider）

低频、初始化重的服务可实现 `DeferredProvider`，首次 `Make` 时才加载：

```go
type DeferredProvider interface {
    ServiceProvider
    DeferredServices() []string
}
```

```go
func (sp *ServiceProvider) DeferredServices() []string {
    return []string{"sms"}
}
```

| 类型 | 行为 |
|------|------|
| 普通 Provider | `Boot()` 阶段立即 Register + Boot |
| DeferredProvider | 跳过 Boot，首次 `Make("sms")` 时才触发 |

适用：远程 API 客户端、非每次启动都用的服务等。

---

## 八、Provider 依赖管理

声明顺序决定 `Register` / `Boot` 执行顺序，**被依赖者在前**：

```go
func providers() []foundation.ServiceProvider {
    return []foundation.ServiceProvider{
        &config.ServiceProvider{},   // 1. 无依赖
        &log.ServiceProvider{},      // 2. 依赖 config
        &cache.ServiceProvider{},    // 3. 依赖 config
        &database.ServiceProvider{}, // 4. 依赖 config + log
        &sms.ServiceProvider{},    // 5. 依赖 config（自定义，放最后）
    }
}
```

| 阶段 | 可以做 | 不可以做 |
|------|--------|---------|
| `Register` | `Bind` / `Singleton` / `Instance` | `Make` / `MustMake` |
| `Boot` | 容器操作、`OnShutdown`、预热逻辑 | — |

---

## 九、内置 Provider 参考

### config.ServiceProvider

```go
func (sp *ServiceProvider) Register(app foundation.Application) {
    app.Singleton("config", func(app foundation.Application) (any, error) {
        return NewConfig(app.BasePath("config.yaml"))
    })
}
```

### database.ServiceProvider（带关闭钩子）

```go
func (sp *ServiceProvider) Register(app foundation.Application) {
    app.Singleton("db", func(app foundation.Application) (any, error) {
        cfg := app.MustMake("config").(contracts.Config)
        log := app.MustMake("log").(contracts.Log)
        return database.NewDBManager(cfg, log)
    })
}

func (sp *ServiceProvider) Boot(app foundation.Application) error {
    app.OnShutdown(func() {
        if db, err := app.Make("db"); err == nil {
            if closer, ok := db.(contracts.DB); ok {
                _ = closer.Close()
            }
        }
    })
    return nil
}
```

> `orm` 服务已 Deprecated，请使用 `db` / `facades.DB()`。

---

## 十、最佳实践

1. **一个 Provider 注册一组相关服务**，不要把所有服务塞进一个 Provider
2. **业务契约放 `services/contracts/`**，框架契约用 `go-fast-framework/contracts`
3. **优先 `Singleton`**，除非服务确实需要每次新建实例
4. **`Register` 只做绑定**，`Boot` 做初始化 — 不要搞反
5. **注册关闭钩子** 释放连接池、定时器、文件句柄等
6. **自定义 Provider 追加在框架 Provider 之后**
7. **低频服务用 `DeferredProvider`**，减少启动开销
8. **写 `_test.go`**，至少覆盖 Register + Boot + Make 路径
9. **通过 `facades.App().MustMake("key")` 访问**，不建 `services/facades/` 包装层

---

## 十一、相关文档

- [AGENT.md](../AGENT.md) — AI 协作指南（目录约定、容器 Key、排查表）
- [容器 API](container.md) — Bind / Singleton / Make 等方法详解
- [Facade 使用说明](facade.md) — 框架 Facade API 完整参考
- [配置说明](configuration.md) — Go 配置 + YAML 双模式
- [插件开发指南](plugins.md) — 将 Provider 打包为独立 Go module 插件

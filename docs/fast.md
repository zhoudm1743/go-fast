# GoFast 服务提供者架构方案（v2）

> **定位**：GoFast 是一个 Go 快速开发框架（类似 Goravel），不是业务项目。
> 框架本身 **不持有任何服务实例**，所有实例由服务容器统一管理。
>
> 参考：[goravel/goravel v1.17.x](https://github.com/goravel/goravel/tree/v1.17.x)

---

## 一、与 Goravel 启动流程对齐

### 1.1 Goravel 的启动方式

```
goravel/goravel（用户脚手架）          goravel/framework（框架核心）
├── main.go                           ├── foundation/application.go
│   └── bootstrap.Boot()              ├── contracts/
├── bootstrap/app.go                  ├── facades/
│   └── app = framework.NewApp()      ├── config/service_provider.go
│   └── app.Boot()                    ├── log/service_provider.go
├── config/app.go                     ├── database/service_provider.go
│   └── Providers = []ServiceProvider ├── filesystem/service_provider.go
│       ConfigSP, LogSP, DbSP ...    └── ...
```

**关键特征：**
1. 框架核心 **零全局变量**，不持有实例
2. 每个服务包内自带 `service_provider.go`，负责把自己注册到容器
3. Facade 是纯静态函数，内部从全局 `App` 容器 `MustMake()`
4. 用户在 `config/app.go` 中声明要启用哪些 Provider（可增删）
5. `bootstrap/app.go` 只做一件事：创建 App → 注册 Providers → Boot

### 1.2 GoFast 的对齐方案

```
GoFast（既是框架又是脚手架，一个仓库）
├── main.go                     ← bootstrap.Boot()
├── bootstrap/app.go            ← NewApp + registerProviders + Boot
├── config/app.go               ← 声明 Providers 列表
│
├── foundation/                 ← 框架核心（零业务依赖）
│   ├── application.go
│   ├── container.go
│   └── provider.go
│
├── contracts/                  ← 纯接口（零实现依赖）
├── facades/                    ← 静态门面
│
├── config/                     ← Config 服务（含 service_provider.go）
├── log/                        ← Log 服务（含 service_provider.go）
├── database/                   ← Database 服务（含 service_provider.go）
├── http/                       ← HTTP 服务（含 service_provider.go）
├── filesystem/                 ← Filesystem 服务（含 service_provider.go）
├── validation/                 ← Validation 服务（含 service_provider.go）
└── ...
```

---

## 二、现有代码问题与改造方向

### 2.1 必须去掉的

| 问题 | 说明 | 改造 |
|---|---|---|
| `services/di` | 旧版复制，不存在实际文件 | **删除引用**，由容器取代 |
| 包级全局 `var cfg *Config` | 框架不应持有实例 | 改为注入容器，Facade 解析 |
| 包级全局 `var db *gorm.DB` | 同上 | 同上 |
| 包级全局 `var log *Logger` | 同上 | 同上 |
| `database.WithSchema` / 多租户 | 框架核心不内置 | **移除**，后期作为插件包 |
| `server.go` 硬编码 `config.Get()` | 服务间直接 import | 改为从容器获取 config |
| `web.DistFS` 前端嵌入 | 业务层关注，不属于框架 | 移到应用层 |

### 2.2 保留的核心能力

| 能力 | 说明 |
|---|---|
| Fiber HTTP 引擎 | 保留，封装为 http 服务 |
| Context / Request / Response 接口 | 保留，已抽象良好 |
| filesystem 驱动注册表 | 保留，天然适合 Provider 模式 |
| validation 验证器 | 保留，封装为 Provider |
| UUID v7 主键回调 | 保留，作为 database 服务内部能力 |
| Migrate / Seed 注册 | 保留，在 database Provider Boot 阶段执行 |

---

## 三、核心架构设计

### 3.1 foundation（框架引擎）

#### Container —— 服务容器

```go
package foundation

type Container interface {
    // Bind 每次 Make 都创建新实例
    Bind(key string, factory func(app Application) (any, error))
    // Singleton 首次 Make 创建，后续返回缓存
    Singleton(key string, factory func(app Application) (any, error))
    // Instance 直接绑定已有实例
    Instance(key string, instance any)
    // Make 解析服务
    Make(key string) (any, error)
    // MustMake 解析服务（失败 panic）
    MustMake(key string) any
    // Bound 是否已绑定
    Bound(key string) bool
    // Flush 清空（测试用）
    Flush()
}
```

实现要点：
- `sync.RWMutex` 保护 map
- Singleton 用 `sync.Once` 确保并发安全懒加载
- `MustMake` 内部调 `Make`，error 则 panic

#### ServiceProvider —— 服务提供者

```go
package foundation

type ServiceProvider interface {
    // Register 绑定服务到容器（不可使用其他服务）
    Register(app Application)
    // Boot 引导服务（所有 Provider 已 Register，可使用其他服务）
    Boot(app Application) error
}
```

**注意**：对比上一版方案，去掉了 `Requires() / Key() / Shutdown()`：
- Goravel 也没有 Requires，它靠 **Provider 在 config/app.go 中的声明顺序** 保证引导顺序（简单直接）
- Shutdown 由 Application 统一管理，不是 Provider 的职责

#### Application —— 应用实例

```go
package foundation

type Application interface {
    Container

    // Boot 按顺序执行所有 Provider 的 Register → Boot
    Boot()

    // MakeWith 带参数解析（高级用法）
    MakeWith(key string, params map[string]any) (any, error)

    // BasePath / ConfigPath / StoragePath 路径辅助
    BasePath(path ...string) string

    // SetProviders 设置服务提供者列表
    SetProviders(providers []ServiceProvider)

    // Shutdown 优雅关闭
    Shutdown()
}
```

**关键**：`Application` **嵌入了 Container**（与 Goravel 一致），所以 `facades.App().MustMake("config")` 直接可用，不需要 `App().Container().MustMake()`。

### 3.2 启动流程（对齐 Goravel）

```
main.go
  └─→ bootstrap.Boot()
        │
        ├── 1. app = foundation.NewApplication(".")
        ├── 2. app.SetProviders(config.Providers())  // 从 config/app.go 读取
        ├── 3. app.Boot()
        │       ├── 遍历 providers → p.Register(app)  // 全部注册
        │       └── 遍历 providers → p.Boot(app)      // 全部引导
        ├── 4. facades.SetApp(app)                    // 设置全局入口
        └── 5. return app
```

#### config/app.go（用户声明 Providers）

```go
package config

import (
    "go-fast/framework/foundation"
    configSP  "go-fast/framework/config"
    logSP     "go-fast/framework/log"
    dbSP      "go-fast/framework/database"
    httpSP    "go-fast/framework/http"
    fsSP      "go-fast/framework/filesystem"
    validSP   "go-fast/framework/validation"
)

func Providers() []foundation.ServiceProvider {
    return []foundation.ServiceProvider{
        &configSP.ServiceProvider{},     // 最先：读配置
        &logSP.ServiceProvider{},        // 其次：初始化日志
        &dbSP.ServiceProvider{},         // 然后：连接数据库
        &fsSP.ServiceProvider{},         // 文件系统
        &validSP.ServiceProvider{},      // 验证器
        &httpSP.ServiceProvider{},       // HTTP 服务器（最后）
        // 业务方在此追加自己的 Provider
    }
}
```

#### bootstrap/app.go

```go
package bootstrap

import (
    "go-fast/framework/config"
    "go-fast/framework/facades"
    "go-fast/framework/foundation"
)

func Boot() foundation.Application {
    app := foundation.NewApplication(".")
    app.SetProviders(config.Providers())
    app.Boot()
    facades.SetApp(app)
    return app
}
```

#### main.go

```go
package main

import (
    "os"
    "os/signal"
    "syscall"

    "go-fast/bootstrap"
    "go-fast/framework/facades"
)

func main() {
    app := bootstrap.Boot()

    // 启动 HTTP（阻塞放协程）
    go func() {
        if err := facades.Route().Run(); err != nil {
            facades.Log().Errorf("server error: %v", err)
        }
    }()

    // 优雅关闭
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    app.Shutdown()
}
```

---

## 四、Contracts（服务契约）

纯接口包，**零外部依赖**（除标准库和 `gorm.io/gorm`）。

### 4.1 contracts/config.go

```go
package contracts

type Config interface {
    // Env 读取环境变量
    Env(key string, defaultValue ...any) any
    // Get 读取配置值（支持点号路径 "database.host"）
    Get(key string, defaultValue ...any) any
    GetString(key string, defaultValue ...string) string
    GetInt(key string, defaultValue ...int) int
    GetBool(key string, defaultValue ...bool) bool
    // Set 运行时设置（不持久化）
    Set(key string, value any)
}
```

### 4.2 contracts/log.go

```go
package contracts

type Log interface {
    Debug(args ...any)
    Debugf(format string, args ...any)
    Info(args ...any)
    Infof(format string, args ...any)
    Warn(args ...any)
    Warnf(format string, args ...any)
    Error(args ...any)
    Errorf(format string, args ...any)
    Fatal(args ...any)
    Fatalf(format string, args ...any)
    Panic(args ...any)
    Panicf(format string, args ...any)
    WithField(key string, value any) Log
    WithFields(fields map[string]any) Log
}
```

### 4.3 contracts/orm.go

```go
package contracts

import "gorm.io/gorm"

type Orm interface {
    DB() *gorm.DB
    Ping() error
    Close() error
}
```

### 4.4 contracts/route.go

```go
package contracts

type Route interface {
    // Run 启动 HTTP 服务器
    Run(addr ...string) error
    // Shutdown 优雅关闭
    Shutdown() error

    // 路由注册
    Get(path string, handler any) Route
    Post(path string, handler any) Route
    Put(path string, handler any) Route
    Delete(path string, handler any) Route
    Patch(path string, handler any) Route

    // 路由组
    Group(prefix string) Route

    // 中间件
    Use(middleware ...any) Route
}
```

### 4.5 contracts/storage.go

```go
package contracts

type Storage interface {
    Driver
    Disk(disk string) Driver
}

type Driver interface {
    Put(file, content string) error
    Get(file string) (string, error)
    Exists(file string) bool
    Delete(file ...string) error
    Url(file string) string
    // ... 其他方法同现有 types.Driver
}
```

### 4.6 contracts/validation.go

```go
package contracts

type Validation interface {
    Validate(obj any) error
    RegisterRule(rule any) error
}
```

---

## 五、Facades（静态门面）

**每个 Facade 就是一个函数**，从全局 App 容器解析实例。与 Goravel 完全一致。

### 5.1 facades/app.go

```go
package facades

import "go-fast/framework/foundation"

var app foundation.Application

func SetApp(a foundation.Application) { app = a }
func App() foundation.Application     { return app }
```

### 5.2 其他 Facades

```go
// facades/config.go
package facades

import "go-fast/framework/contracts"

func Config() contracts.Config {
    return App().MustMake("config").(contracts.Config)
}
```

```go
// facades/log.go
func Log() contracts.Log {
    return App().MustMake("log").(contracts.Log)
}
```

```go
// facades/orm.go
func Orm() contracts.Orm {
    return App().MustMake("orm").(contracts.Orm)
}
```

```go
// facades/route.go
func Route() contracts.Route {
    return App().MustMake("route").(contracts.Route)
}
```

```go
// facades/storage.go
func Storage() contracts.Storage {
    return App().MustMake("storage").(contracts.Storage)
}
```

```go
// facades/validator.go
func Validator() contracts.Validation {
    return App().MustMake("validator").(contracts.Validation)
}
```

**业务代码使用：**

```go
import "go-fast/framework/facades"

func CreateUser(ctx http.Context) error {
    db := facades.Orm().DB()
    facades.Log().Info("creating user")
    facades.Storage().Put("avatars/1.png", content)
    return ctx.OkWithData(user)
}
```

---

## 六、各服务包改造（每包自带 ServiceProvider）

### 6.1 config 包

```
config/
├── config.go              ← Config 接口实现（包装 viper）
├── service_provider.go    ← ServiceProvider：注册到容器
└── types.go               ← 配置结构体定义（保留）
```

```go
// config/service_provider.go
package config

import "go-fast/framework/foundation"

type ServiceProvider struct{}

func (sp *ServiceProvider) Register(app foundation.Application) {
    app.Singleton("config", func(app foundation.Application) (any, error) {
        return NewConfig(app.BasePath("config.yaml"))
    })
}

func (sp *ServiceProvider) Boot(app foundation.Application) error {
    return nil
}
```

```go
// config/config.go
package config

// configImpl 实现 contracts.Config
type configImpl struct {
    viper *viper.Viper
}

func NewConfig(path string) (*configImpl, error) {
    v := viper.New()
    v.SetConfigFile(path)
    v.SetConfigType("yaml")
    if err := v.ReadInConfig(); err != nil {
        return nil, fmt.Errorf("读取配置失败: %w", err)
    }
    return &configImpl{viper: v}, nil
}

func (c *configImpl) Get(key string, def ...any) any {
    if !c.viper.IsSet(key) && len(def) > 0 {
        return def[0]
    }
    return c.viper.Get(key)
}

func (c *configImpl) GetString(key string, def ...string) string {
    if !c.viper.IsSet(key) && len(def) > 0 {
        return def[0]
    }
    return c.viper.GetString(key)
}
// ... GetInt / GetBool / Set / Env
```

### 6.2 log 包

```go
// log/service_provider.go
package log

import "go-fast/framework/foundation"

type ServiceProvider struct{}

func (sp *ServiceProvider) Register(app foundation.Application) {
    app.Singleton("log", func(app foundation.Application) (any, error) {
        cfg := app.MustMake("config").(contracts.Config)
        return NewLogger(cfg)    // 用 config 构建 logger，不再 import config 包
    })
}

func (sp *ServiceProvider) Boot(app foundation.Application) error {
    return nil
}
```

**关键变化**：logger 不再 `import config`，而是通过容器获取 config 实例。

### 6.3 database 包

```go
// database/service_provider.go
package database

type ServiceProvider struct{}

func (sp *ServiceProvider) Register(app foundation.Application) {
    app.Singleton("orm", func(app foundation.Application) (any, error) {
        cfg := app.MustMake("config").(contracts.Config)
        log := app.MustMake("log").(contracts.Log)
        return NewOrm(cfg, log) // 从容器拿 config 和 log，不 import 其他服务包
    })
}

func (sp *ServiceProvider) Boot(app foundation.Application) error {
    // 自动迁移（若配置开启）
    cfg := app.MustMake("config").(contracts.Config)
    if cfg.GetBool("database.migrate") {
        return Migrate(app.MustMake("orm").(contracts.Orm).DB())
    }
    return nil
}
```

**关键变化**：
- 去掉 `var db *gorm.DB` 全局变量
- 去掉 `WithSchema` / 多租户（后期插件化）
- 去掉 `database.Init()` / `database.GetDB()`，全部走容器

### 6.4 http 包

```go
// http/service_provider.go
package http

type ServiceProvider struct{}

func (sp *ServiceProvider) Register(app foundation.Application) {
    app.Singleton("route", func(app foundation.Application) (any, error) {
        cfg := app.MustMake("config").(contracts.Config)
        return NewRoute(cfg)  // 创建 Fiber app，注册中间件
    })
}

func (sp *ServiceProvider) Boot(app foundation.Application) error {
    return nil
}
```

**关键变化**：
- 去掉 `server.New()` 中直接 `config.Get()`
- 去掉 `web.DistFS` 前端嵌入（移到应用层 Provider）
- Route 实现 `contracts.Route` 接口

### 6.5 filesystem 包

```go
// filesystem/service_provider.go
package filesystem

type ServiceProvider struct{}

func (sp *ServiceProvider) Register(app foundation.Application) {
    app.Singleton("storage", func(app foundation.Application) (any, error) {
        cfg := app.MustMake("config").(contracts.Config)
        return NewStorage(cfg) // 复用现有 RegisterDriver 机制
    })
}

func (sp *ServiceProvider) Boot(app foundation.Application) error {
    RegisterDefaultDrivers() // 注册内置云存储驱动
    return nil
}
```

### 6.6 validation 包

```go
// validation/service_provider.go
package validation

type ServiceProvider struct{}

func (sp *ServiceProvider) Register(app foundation.Application) {
    app.Singleton("validator", func(app foundation.Application) (any, error) {
        return NewValidator() // 包装 go-playground/validator
    })
}

func (sp *ServiceProvider) Boot(app foundation.Application) error {
    return nil
}
```

---

## 七、目录结构（最终）

```
GoFast/
├── main.go                         # bootstrap.Boot() → Run → Shutdown
├── bootstrap/
│   └── app.go                      # NewApp → SetProviders → Boot → SetFacade
├── config/
│   └── app.go                      # Providers() 列表（用户可编辑）
│
├── foundation/                     # 框架引擎（零业务依赖）
│   ├── application.go              # Application 接口 + 实现
│   ├── container.go                # Container 接口 + 实现
│   └── provider.go                 # ServiceProvider 接口
│
├── contracts/                      # 纯接口
│   ├── config.go
│   ├── log.go
│   ├── orm.go
│   ├── route.go
│   ├── storage.go
│   ├── validation.go
│   ├── cache.go                    # 预留
│   └── event.go                    # 预留
│
├── facades/                        # 静态门面
│   ├── app.go
│   ├── cache.go
│   ├── config.go
│   ├── log.go
│   ├── orm.go
│   ├── route.go
│   ├── storage.go
│   └── validator.go
│
├── config/                         # Config 服务
│   ├── config.go                   # contracts.Config 实现（包装 viper）
│   └── service_provider.go
│
├── log/                            # Log 服务
│   ├── logger.go                   # contracts.Log 实现（包装 logrus + lumberjack）
│   └── service_provider.go
│
├── cache/                          # Cache 服务
│   ├── cache.go                    # contracts.Cache 实现（CacheManager 多 Store 管理）
│   ├── memory_store.go             # 分片内存缓存（FNV-1a 哈希、标签、Hash、锁、GC）
│   ├── cache_test.go               # 13 个测试
│   └── service_provider.go
│
├── database/                       # Database 服务
│   ├── orm.go                      # contracts.Orm 实现
│   ├── model.go                    # Model 基类 + UUID v7 回调
│   └── service_provider.go
│
├── http/                           # HTTP 服务
│   ├── route.go                    # contracts.Route 实现（Fiber 封装）
│   └── service_provider.go
│
├── filesystem/                     # Filesystem 服务
│   ├── storage.go                  # contracts.Storage 实现（多磁盘管理）
│   ├── local.go                    # 本地文件系统驱动
│   ├── filesystem_test.go          # 13 个测试
│   └── service_provider.go
│
├── validation/                     # Validation 服务
│   ├── validator.go                # contracts.Validation 实现
│   └── service_provider.go
│
└── utils/                          # 工具函数（保留）
    ├── file.go
    ├── string.go
    └── tools.go
```

---

## 八、与 Goravel 对照

| 特性 | Goravel | GoFast |
|---|---|---|
| 容器 | Application 嵌入 Container | **同** |
| Provider 顺序 | config/app.go 声明顺序即 Boot 顺序 | **同** |
| Facade | `facades.Xxx()` 从 App MustMake | **同** |
| 路由 | `facades.Route()` | **同** |
| ORM | `facades.Orm().Query()` 自研 | `facades.Orm().DB()` 返回 *gorm.DB |
| 日志 | `facades.Log()` | **同** |
| 配置 | `facades.Config().Get()` | **同** |
| 文件 | `facades.Storage()` | **同** |
| 缓存 | `facades.Cache()` Redis 为主 | `facades.Cache()` 分片内存（可扩展 Redis） |
| HTTP | 基于 gin | 基于 Fiber v2 |
| 多租户 | 无内置 | 无内置（后期插件） |
| 验证 | 自研 | go-playground/validator |
| UUID 主键 | 无 | 内置 UUID v7 |
| 全局变量 | 零（仅 facades.app） | **同** |

---

## 九、TODO List

### Phase 1：框架引擎 ⭐ 最高优先级

- [x] **1.1** `foundation/container.go` — Container 接口 + 实现（Bind / Singleton / Instance / Make / MustMake / Flush）
- [x] **1.2** `foundation/provider.go` — ServiceProvider 接口（仅 Register + Boot）
- [x] **1.3** `foundation/application.go` — Application 接口 + 实现（嵌入 Container，Boot 按序 Register→Boot，Shutdown + OnShutdown）
- [x] **1.4** `foundation/` 单元测试 — 12 个测试全部通过（Bind/Singleton/Make 并发安全、Boot 顺序、Shutdown 逆序）

### Phase 2：服务契约

- [x] **2.1** `contracts/config.go`
- [x] **2.2** `contracts/log.go`
- [x] **2.3** `contracts/orm.go`
- [x] **2.4** `contracts/route.go`
- [x] **2.5** `contracts/storage.go`
- [x] **2.6** `contracts/validation.go`
- [x] **2.7** `contracts/cache.go`（预留）
- [x] **2.8** `contracts/event.go`（预留）

### Phase 3：静态门面

- [x] **3.1** `facades/app.go` — SetApp / App
- [x] **3.2** `facades/config.go`
- [x] **3.3** `facades/log.go`
- [x] **3.4** `facades/orm.go`
- [x] **3.5** `facades/route.go`
- [x] **3.6** `facades/storage.go`
- [x] **3.7** `facades/validator.go`

### Phase 4：各服务改造（每个包自带 service_provider.go）

- [x] **4.1** 改造 config 服务 — 实现 `contracts.Config`，包装 viper，零全局变量
- [x] **4.2** 改造 log 服务 — 实现 `contracts.Log`，包装 logrus + lumberjack，通过容器获取 config
- [x] **4.3** 改造 database 服务 — 实现 `contracts.Orm`，零全局变量，含 UUID v7 + Model 基类
- [x] **4.4** 改造 http 服务 — 实现 `contracts.Route`，Fiber 封装，零 config 直接 import
- [x] **4.5** 改造 filesystem 服务 — 实现 `contracts.Storage`，本地驱动 + Storage 管理器 + 13 个测试
- [x] **4.6** 改造 validation 服务 — 实现 `contracts.Validation`，包装 go-playground/validator

### Phase 5：启动流程

- [x] **5.1** 创建 `bootstrap/app.go` — Boot 函数（含 7 个 Providers：config → log → cache → database → filesystem → validation → http）
- [x] **5.2** 重写 `main.go` — `bootstrap.Boot()` → `Run` → signal → `Shutdown`

### Phase 6：清理

- [x] **6.1** 删除 `services/` 目录（含 di、router、server、config、database、filesystem、logger 全部旧代码）
- [x] **6.2** `go mod tidy` 清理未使用依赖（aliyun-oss / aws-sdk / cloudinary / minio / copier / swaggo 等已移除）

### Phase 7：高级特性

- [ ] **7.1** 支持 DeferredProvider（延迟加载，首次 Make 时才 Register+Boot）
- [x] **7.2** Application 注册 shutdown hook（`OnShutdown(func())`）— 已内置，database/http/cache 均注册了关闭钩子
- [ ] **7.3** 实现 i18n 服务 + Provider
- [x] **7.4** 实现 cache 服务 + Provider — 分片内存缓存（FNV-1a 哈希分片、标签分组、Hash 操作、原子自增、分布式锁、Remember、批量操作、GC）+ 13 个测试
- [x] **7.4.1** `facades/cache.go` — `facades.Cache()` 门面
- [ ] **7.5** 实现 event 服务 + Provider
- [ ] **7.6** 多租户插件包 `plugins/tenant`

### Phase 8：文档

- [ ] **8.1** `docs/getting-started.md` — 快速开始
- [ ] **8.2** `docs/service-provider.md` — 如何编写自定义 Provider
- [ ] **8.3** `docs/facade.md` — Facade 使用说明
- [ ] **8.4** `docs/container.md` — 容器 API
- [ ] **8.5** `docs/plugins.md` — 插件开发指南（多租户、缓存等）

# GoFast 插件开发指南

> 插件（Plugin）是 GoFast 中可复用、可分发的功能扩展单元。
> 一个插件本质上是一个独立的 Go module，内部包含 **实现 + ServiceProvider**，
> 业务方只需 `go get` 后在 `providers()` 列表中追加即可使用。

---

## 一、插件 vs. 自定义 Provider

| 维度 | 自定义 Provider | 插件 |
|------|----------------|------|
| 代码位置 | 项目内部（如 `sms/service_provider.go`） | 独立 Go module |
| 分发方式 | 拷贝 / mono-repo | `go get github.com/xxx/gofast-xxx` |
| 版本管理 | 跟随主项目 | 独立 semver |
| 适用场景 | 业务专属逻辑 | 通用能力（Redis、OSS、邮件、队列…） |

两者在技术实现上完全一致，都是实现 `foundation.ServiceProvider` 接口，
并可按需实现下面介绍的**可选扩展接口**。

---

## 二、Boot 生命周期

GoFast 的 `Boot()` 分 4 个阶段依次执行，插件可通过不同接口在合适的阶段介入：

```
Phase 1  Register         — 所有 Provider.Register()，仅绑定工厂，不实例化
Phase 1.5 ConfigDefaults  — ConfigProvider.ConfigDefaults() 写入配置默认值
Phase 2  Boot             — 所有 Provider.Boot()，可安全 MustMake 任意服务
Phase 3  Migrate          — Migrator.Migrate()，自动同步数据库表结构
Phase 4  RegisterRoutes   — RouteRegistrar.RegisterRoutes()，注册 HTTP 路由
```

---

## 三、可选扩展接口

框架在 `Boot` 过程中自动检测并调用以下可选接口，插件按需实现即可。

### 3.1 ConfigProvider — 声明默认配置

实现此接口后，框架在 Phase 1.5 自动将返回的 key-value 写入 Config 服务。
**已在配置文件或环境变量中设置的 key 不会被覆盖**。

```go
func (sp *ServiceProvider) ConfigDefaults() map[string]any {
    return map[string]any{
        "redis.host":     "127.0.0.1",
        "redis.port":     6379,
        "redis.password": "",
        "redis.db":       0,
        "redis.pool":     10,
    }
}
```

用户只需在 `config.yaml` 中覆盖自己想改的部分，其余值使用插件默认值。

### 3.2 Migrator — 自动数据库迁移

实现此接口后，框架在 Phase 3（所有 Boot 完成后）自动调用 `Migrate`，
无需在 `Boot()` 中手动获取 ORM 并调用 AutoMigrate。

```go
func (sp *ServiceProvider) Migrate(orm contracts.Orm) error {
    return orm.AutoMigrate(&Post{}, &Comment{}, &Tag{})
}
```

> 若应用未注册数据库服务（`"orm"` 未绑定），该方法不会被调用。

### 3.3 RouteRegistrar — 注册 HTTP 路由

实现此接口后，框架在 Phase 4（所有 Boot 完成后）自动调用 `RegisterRoutes`，
无需在 `Boot()` 中手动获取 Route 服务。

```go
func (sp *ServiceProvider) RegisterRoutes(r contracts.Route) {
    r.Group("/blog", func(g contracts.Route) {
        g.Get("/posts", sp.listPosts)
        g.Get("/posts/:id", sp.getPost)
    })
    r.Group("/blog", sp.authMiddleware, func(g contracts.Route) {
        g.Post("/posts", sp.createPost)
        g.Put("/posts/:id", sp.updatePost)
        g.Delete("/posts/:id", sp.deletePost)
    })
}
```

> 若应用未注册 HTTP 服务（`"route"` 未绑定），该方法不会被调用。

---

## 四、Application 类型化访问方法

在 `Boot()` 和可选接口方法中，可使用 `app` 的类型化快捷方法，
无需手写 `app.MustMake("config").(contracts.Config)` 等类型断言：

```go
app.Config()   // contracts.Config
app.Log()      // contracts.Log
app.Cache()    // contracts.Cache
app.Orm()      // contracts.Orm
app.Route()    // contracts.Route
app.Storage()  // contracts.Storage
```

---

## 五、完整示例 — Blog 插件

以下示例展示了一个同时使用了所有扩展接口的 Blog 插件。

### 5.1 目录结构

```
gofast-blog/
├── go.mod                  # module github.com/example/gofast-blog
├── models.go               # GORM 模型
├── handlers.go             # HTTP 处理函数
├── service_provider.go     # ServiceProvider + 所有可选接口
└── facade.go               # 便捷访问函数（可选）
```

### 5.2 模型定义

```go
// models.go
package blog

import "time"

type Post struct {
    ID        string    `gormdriver:"primaryKey"`
    Title     string    `gormdriver:"not null"`
    Content   string
    Published bool      `gormdriver:"default:false"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

type Comment struct {
    ID        string `gormdriver:"primaryKey"`
    PostID    string `gormdriver:"index;not null"`
    Body      string `gormdriver:"not null"`
    CreatedAt time.Time
}
```

### 5.3 ServiceProvider（含所有可选接口）

```go
// service_provider.go
package blog

import (
    "github.com/zhoudm1743/go-fast/framework/contracts"
    "github.com/zhoudm1743/go-fast/framework/foundation"
)

// ServiceProvider Blog 插件服务提供者。
// 同时实现了 ConfigProvider、Migrator、RouteRegistrar 三个可选接口。
type ServiceProvider struct{}

// ── 必须实现 ──────────────────────────────────────────────────────────

func (sp *ServiceProvider) Register(app foundation.Application) {
    // 绑定插件自己的服务到容器（可选，简单插件可为空）
    app.Singleton("blog", func(app foundation.Application) (any, error) {
        return NewBlogService(app.Orm(), app.Cache()), nil
    })
}

func (sp *ServiceProvider) Boot(app foundation.Application) error {
    // 注册关闭钩子（如有需要）
    app.OnShutdown(func() {
        app.Log().Info("[blog] shutting down")
    })
    return nil
}

// ── ConfigProvider：声明默认配置 ──────────────────────────────────────

func (sp *ServiceProvider) ConfigDefaults() map[string]any {
    return map[string]any{
        "blog.per_page":       10,
        "blog.cache_ttl_sec":  300,
        "blog.allow_comments": true,
    }
}

// ── Migrator：自动数据库迁移 ──────────────────────────────────────────

func (sp *ServiceProvider) Migrate(orm contracts.Orm) error {
    return orm.AutoMigrate(&Post{}, &Comment{})
}

// ── RouteRegistrar：注册 HTTP 路由 ────────────────────────────────────

func (sp *ServiceProvider) RegisterRoutes(r contracts.Route) {
    r.Group("/blog", func(g contracts.Route) {
        g.Get("/posts", sp.listPosts)
        g.Get("/posts/:id", sp.getPost)
    })
}

func (sp *ServiceProvider) listPosts(ctx contracts.Context) error {
    // 通过 facade 或直接从容器拿服务都可以
    return ctx.Response().Success(nil)
}

func (sp *ServiceProvider) getPost(ctx contracts.Context) error {
    return ctx.Response().Success(nil)
}
```

### 5.4 在 Boot 中直接读取配置与服务

```go
func (sp *ServiceProvider) Boot(app foundation.Application) error {
    // 类型化访问，无需类型断言
    perPage := app.Config().GetInt("blog.per_page", 10)
    app.Log().Infof("[blog] per_page = %d", perPage)

    // 读取缓存
    _ = app.Cache().Put("blog:init", true, 0)

    // 注册关闭钩子
    app.OnShutdown(func() {
        app.Log().Info("[blog] shutdown")
    })
    return nil
}
```

---

## 六、延迟加载（DeferredProvider）

如果插件的服务不是启动时必须初始化的，可实现 `DeferredProvider`，
在首次 `Make("xxx")` 时才触发 Register + Boot，减少启动开销：

```go
// 实现 DeferredProvider 接口（此时 Migrator/RouteRegistrar 不会自动执行）
func (sp *ServiceProvider) DeferredServices() []string {
    return []string{"blog"}
}
```

> ⚠️ 延迟 Provider **不参与** Phase 3/4，即 `Migrator.Migrate` 和
> `RouteRegistrar.RegisterRoutes` 不会被框架自动调用。
> 如需迁移或注册路由，请改为即时 Provider，或在 `Boot()` 中手动处理。

---

## 七、接入业务项目

### 7.1 安装

```bash
go get github.com/example/gofast-blog@latest
```

### 7.2 注册 Provider

```go
// bootstrap/app.go
import (
    // ...existing imports...
    blog "github.com/example/gofast-blog"
)

func providers() []foundation.ServiceProvider {
    return []foundation.ServiceProvider{
        &config.ServiceProvider{},
        &log.ServiceProvider{},
        &cache.ServiceProvider{},
        &database.ServiceProvider{},
        &filesystem.ServiceProvider{},
        &validation.ServiceProvider{},
        &gohttp.ServiceProvider{},
        // ↓ 追加 Blog 插件（放在 http 之后）
        &blog.ServiceProvider{},
    }
}
```

### 7.3 覆盖默认配置（可选）

插件通过 `ConfigDefaults` 声明了合理默认值，如需覆盖，在 `config.yaml` 中添加：

```yaml
blog:
  per_page: 20
  cache_ttl_sec: 600
  allow_comments: false
```

未配置的项自动使用插件默认值，**无需为每个配置项都设置**。

---

## 八、插件开发规范

### 8.1 命名约定

| 项目 | 约定 | 示例 |
|------|------|------|
| Module 名 | `gofast-<功能>` | `gofast-redis`、`gofast-oss`、`gofast-mail` |
| 包名 | 简短名词 | `redis`、`oss`、`mail` |
| 容器 Key | 小写名词 | `"redis"`、`"oss"`、`"mail"` |
| Provider 类型名 | `ServiceProvider` | `redis.ServiceProvider` |
| 配置前缀 | 插件名 | `redis.*`、`oss.aliyun.*` |

### 8.2 依赖原则

- 插件应 **仅依赖 `github.com/zhoudm1743/go-fast/framework/foundation` 和 `github.com/zhoudm1743/go-fast/framework/contracts`**
- **不要依赖** `github.com/zhoudm1743/go-fast/framework/facades`（避免循环依赖）
- 如需提供便捷函数，在插件包内自行实现（见 `facade.go`）

### 8.3 配置约定

- 通过 `ConfigProvider.ConfigDefaults()` 声明所有配置项及默认值
- 配置项以插件名为前缀：`redis.*`、`oss.*`
- 必填项应在 `Boot()` 或工厂函数中校验并返回明确 error

### 8.4 错误处理

- 工厂函数返回 `error` 而非 panic
- 错误消息以 `[gofast-<插件名>]` 为前缀

```go
return nil, fmt.Errorf("[gofast-blog] init failed: %w", err)
```

### 8.5 关闭钩子

持有资源的插件 **必须** 在 `Boot` 中注册 `OnShutdown` 钩子：

```go
func (sp *ServiceProvider) Boot(app foundation.Application) error {
    app.OnShutdown(func() {
        // 关闭连接池、停止定时器等
    })
    return nil
}
```

### 8.6 测试

```go
func TestBlogServiceProvider(t *testing.T) {
    app := foundation.NewApplication(".")

    // 注入 Mock 依赖
    app.Instance("config", &mockConfig{
        data: map[string]any{"blog.per_page": 5},
    })
    app.Instance("log",   &mockLog{})
    app.Instance("cache", &mockCache{})
    app.Instance("orm",   &mockOrm{})
    app.Instance("route", &mockRoute{})

    sp := &ServiceProvider{}
    sp.Register(app)
    _ = sp.Boot(app)

    svc, err := app.Make("blog")
    require.NoError(t, err)
    require.NotNil(t, svc)
}
```

---

## 九、可开发的插件方向

| 插件 | 容器 Key | 说明 |
|------|---------|------|
| `gofast-redis` | `redis` | Redis Store + 原生客户端 |
| `gofast-oss` | `oss` | 阿里云 OSS 文件存储驱动 |
| `gofast-s3` | `s3` | AWS S3 / MinIO 存储驱动 |
| `gofast-mail` | `mail` | 邮件发送（SMTP / SendGrid） |
| `gofast-queue` | `queue` | 消息队列（Redis / RabbitMQ） |
| `gofast-event` | `event` | 事件总线（同步 / 异步） |
| `gofast-scheduler` | `scheduler` | 定时任务调度 |
| `gofast-jwt` | `jwt` | JWT 认证中间件 |
| `gofast-i18n` | `i18n` | 国际化翻译 |

---

## 十、发布检查清单

- [ ] `go mod tidy` 无多余依赖
- [ ] 所有声明的 `contracts` 接口方法均已实现
- [ ] 通过 `ConfigDefaults()` 为所有配置项提供默认值
- [ ] `Migrator.Migrate()` 已实现（如有数据库模型）
- [ ] `RouteRegistrar.RegisterRoutes()` 已实现（如提供 HTTP 接口）
- [ ] `OnShutdown` 钩子已注册（如持有资源）
- [ ] `_test.go` 覆盖核心逻辑
- [ ] `README.md` 包含安装、配置说明与使用示例
- [ ] 错误消息带 `[gofast-<name>]` 前缀
- [ ] 版本号遵循 semver

---

## 十一、相关文档

- [快速开始](getting-started.md) — 框架安装与入门
- [容器 API](container.md) — Bind / Singleton / Make 详解
- [编写自定义 Provider](service-provider.md) — Provider 开发基础
- [Facade 使用说明](facade.md) — 门面模式详解


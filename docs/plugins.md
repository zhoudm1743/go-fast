# GoFast 插件开发指南

> 插件（Plugin）是 GoFast 中可复用、可分发的功能扩展单元。
> 一个插件本质上是一个独立的 Go module，内部包含 **contracts + 实现 + ServiceProvider**，
> 业务方只需 `go get` 后在 `providers()` 列表中追加即可使用。

---

## 一、插件 vs. 自定义 Provider

| 维度 | 自定义 Provider | 插件 |
|------|----------------|------|
| 代码位置 | 项目内部（如 `sms/service_provider.go`） | 独立 Go module |
| 分发方式 | 拷贝 / mono-repo | `go get github.com/xxx/gofast-xxx` |
| 版本管理 | 跟随主项目 | 独立 semver |
| 适用场景 | 业务专属逻辑 | 通用能力（Redis、OSS、邮件、队列…） |

两者在技术实现上完全一致，都是实现 `foundation.ServiceProvider` 接口。

---

## 二、插件目录结构

推荐的插件 module 目录结构：

```
gofast-redis/
├── go.mod                      # module github.com/example/gofast-redis
├── go.sum
├── README.md
├── contracts.go                # 对外暴露的接口（可选，也可复用主框架的 contracts）
├── redis.go                    # 核心实现
├── redis_test.go               # 单元测试
├── service_provider.go         # ServiceProvider 入口
└── facade.go                   # Facade 辅助函数（可选）
```

---

## 三、完整示例 — Redis 缓存插件

### 3.1 初始化 module

```bash
mkdir gofast-redis && cd gofast-redis
go mod init github.com/example/gofast-redis
go get github.com/redis/go-redis/v9
go get gitee.com/your-org/GoFast/backend  # 依赖框架核心
```

### 3.2 定义契约（可选）

如果插件需要暴露额外接口（超出框架 `contracts.CacheStore` 的部分），可自定义：

```go
// contracts.go
package redis

import "context"

// RedisClient 插件对外暴露的 Redis 原生客户端接口。
type RedisClient interface {
    Eval(ctx context.Context, script string, keys []string, args ...any) (any, error)
    Pipeline(ctx context.Context, fn func(pipe Pipeliner) error) error
    Subscribe(ctx context.Context, channels ...string) PubSub
}
```

### 3.3 实现服务

```go
// redis.go
package redis

import (
    "context"
    "fmt"
    "time"

    "go-fast/framework/contracts"

    goredis "github.com/redis/go-redis/v9"
)

type redisStore struct {
    client *goredis.Client
}

// NewRedisStore 根据配置创建 Redis Store。
func NewRedisStore(cfg contracts.Config) (*redisStore, error) {
    addr := cfg.GetString("cache.redis.host", "127.0.0.1") +
        ":" + cfg.GetString("cache.redis.port", "6379")

    client := goredis.NewClient(&goredis.Options{
        Addr:     addr,
        Password: cfg.GetString("cache.redis.password"),
        DB:       cfg.GetInt("cache.redis.db", 0),
    })

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := client.Ping(ctx).Err(); err != nil {
        return nil, fmt.Errorf("[gofast-redis] connection failed: %w", err)
    }

    return &redisStore{client: client}, nil
}

func (r *redisStore) Get(key string, def ...any) any {
    val, err := r.client.Get(context.Background(), key).Result()
    if err != nil {
        if len(def) > 0 {
            return def[0]
        }
        return nil
    }
    return val
}

func (r *redisStore) Put(key string, value any, ttl time.Duration) error {
    return r.client.Set(context.Background(), key, value, ttl).Err()
}

// ... 实现 contracts.CacheStore 的其余方法 ...

func (r *redisStore) Close() error {
    return r.client.Close()
}
```

### 3.4 编写 ServiceProvider

```go
// service_provider.go
package redis

import (
    "go-fast/framework/contracts"
    "go-fast/framework/foundation"
)

// ServiceProvider Redis 插件的服务提供者。
type ServiceProvider struct{}

func (sp *ServiceProvider) Register(app foundation.Application) {
    app.Singleton("redis", func(app foundation.Application) (any, error) {
        cfg := app.MustMake("config").(contracts.Config)
        return NewRedisStore(cfg)
    })
}

func (sp *ServiceProvider) Boot(app foundation.Application) error {
    // 注册关闭钩子
    app.OnShutdown(func() {
        if r, err := app.Make("redis"); err == nil {
            if store, ok := r.(*redisStore); ok {
                _ = store.Close()
            }
        }
    })
    return nil
}
```

### 3.5 延迟加载（可选）

如果希望仅在首次使用时才连接 Redis：

```go
// 实现 DeferredProvider 接口
func (sp *ServiceProvider) DeferredServices() []string {
    return []string{"redis"}
}
```

### 3.6 提供 Facade（可选）

```go
// facade.go
package redis

import "go-fast/framework/facades"

// Redis 获取 Redis Store 实例的便捷函数。
func Redis() *redisStore {
    return facades.App().MustMake("redis").(*redisStore)
}
```

---

## 四、业务方接入

### 4.1 安装插件

```bash
go get github.com/example/gofast-redis@latest
```

### 4.2 注册 Provider

```go
// bootstrap/app.go
import (
    // ...existing imports...
    goredis "github.com/example/gofast-redis"
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
        // ↓ 追加 Redis 插件
        &goredis.ServiceProvider{},
    }
}
```

### 4.3 添加配置

```yaml
# config.yaml
cache:
  redis:
    host: 127.0.0.1
    port: 6379
    password: ""
    db: 0
```

### 4.4 使用

```go
// 通过容器
store := facades.App().MustMake("redis").(*redis.RedisStore)
store.Put("key", "value", time.Hour)

// 通过插件提供的 Facade
redis.Redis().Put("key", "value", time.Hour)
```

---

## 五、插件开发规范

### 5.1 命名约定

| 项目 | 约定 | 示例 |
|------|------|------|
| Module 名 | `gofast-<功能>` | `gofast-redis`、`gofast-oss`、`gofast-mail` |
| 包名 | 简短名词 | `redis`、`oss`、`mail` |
| 容器 Key | 小写名词 | `"redis"`、`"oss"`、`"mail"` |
| Provider 类型名 | `ServiceProvider` | `redis.ServiceProvider` |

### 5.2 依赖原则

- 插件应 **仅依赖 `go-fast/framework/foundation` 和 `go-fast/framework/contracts`**
- **不要依赖** `go-fast/framework/facades`（避免循环依赖）
- 如需提供 Facade，在插件包内自行实现（如上面的 `redis.Redis()`）

### 5.3 配置约定

- 配置项以插件名为前缀：`cache.redis.*`、`oss.aliyun.*`
- 提供合理的默认值，减少必填项
- 在 README 中列出所有配置项及默认值

### 5.4 错误处理

- 工厂函数返回 `error` 而非 panic
- 错误消息以 `[gofast-<插件名>]` 为前缀

```go
return nil, fmt.Errorf("[gofast-redis] connection failed: %w", err)
```

### 5.5 关闭钩子

持有资源的插件 **必须** 在 `Boot` 中注册 `OnShutdown` 钩子：

```go
func (sp *ServiceProvider) Boot(app foundation.Application) error {
    app.OnShutdown(func() {
        // 释放连接池、停止定时器等
    })
    return nil
}
```

### 5.6 测试

- 提供 `_test.go` 单元测试
- 使用 `foundation.NewApplication(".")` 创建测试容器
- 通过 `Instance` 注入 Mock 依赖

```go
func TestRedisStore(t *testing.T) {
    app := foundation.NewApplication(".")

    // Mock config
    app.Instance("config", &mockConfig{
        values: map[string]any{
            "cache.redis.host": "127.0.0.1",
            "cache.redis.port": "6379",
        },
    })

    provider := &ServiceProvider{}
    provider.Register(app)
    _ = provider.Boot(app)

    store, err := app.Make("redis")
    // ...
}
```

---

## 六、可开发的插件方向

以下是推荐的插件开发方向，均可按上述规范实现：

| 插件 | 容器 Key | 说明 |
|------|---------|------|
| `gofast-redis` | `redis` | Redis 缓存 Store + 原生客户端 |
| `gofast-oss` | `oss` | 阿里云 OSS 文件存储驱动 |
| `gofast-s3` | `s3` | AWS S3 / MinIO 文件存储驱动 |
| `gofast-cos` | `cos` | 腾讯云 COS 文件存储驱动 |
| `gofast-mail` | `mail` | 邮件发送（SMTP / SendGrid / Mailgun） |
| `gofast-queue` | `queue` | 消息队列（Redis / RabbitMQ / Kafka） |
| `gofast-event` | `event` | 事件总线（同步 / 异步） |
| `gofast-scheduler` | `scheduler` | 定时任务调度 |
| `gofast-jwt` | `jwt` | JWT 认证中间件 |
| `gofast-i18n` | `i18n` | 国际化翻译 |

---

## 七、发布检查清单

发布插件前请确认：

- [ ] `go mod tidy` 无多余依赖
- [ ] 所有 `contracts.XxxInterface` 方法均已实现
- [ ] `_test.go` 覆盖核心逻辑
- [ ] `README.md` 包含安装、配置、使用示例
- [ ] 注册了 `OnShutdown` 钩子（如有资源需释放）
- [ ] 错误消息带 `[gofast-<name>]` 前缀
- [ ] 配置项有合理默认值
- [ ] 版本号遵循 semver

---

## 八、相关文档

- [快速开始](getting-started.md) — 框架安装与入门
- [容器 API](container.md) — Bind / Singleton / Make 详解
- [编写自定义 Provider](service-provider.md) — Provider 开发基础
- [Facade 使用说明](facade.md) — 门面模式详解


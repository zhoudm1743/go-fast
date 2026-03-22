# GoFast Facade 使用说明

> Facade（门面）为容器中的服务提供静态访问入口，让你在任何地方都可以直接调用 `facades.Xxx()` 而无需手动从容器解析。
> 所有 Facade 定义在 `facades/` 包下，每个函数内部均通过 `App().MustMake(key)` 解析并断言为对应的 `contracts` 接口。

---

## 一、Facade 一览

| Facade 函数 | 返回类型 | 容器 Key | 说明 |
|-------------|---------|---------|------|
| `facades.App()` | `foundation.Application` | — | 全局应用实例 |
| `facades.Config()` | `contracts.Config` | `config` | 配置服务 |
| `facades.Log()` | `contracts.Log` | `log` | 日志服务 |
| `facades.Cache()` | `contracts.Cache` | `cache` | 缓存服务 |
| `facades.Orm()` | `contracts.Orm` | `orm` | ORM 数据库服务 |
| `facades.Route()` | `contracts.Route` | `route` | HTTP 路由服务 |
| `facades.Storage()` | `contracts.Storage` | `storage` | 文件存储服务 |
| `facades.Validator()` | `contracts.Validation` | `validator` | 验证服务 |

---

## 二、facades.App()

返回全局 `foundation.Application` 实例，可直接调用容器方法和应用方法。

```go
import "go-fast/framework/facades"

app := facades.App()

app.Version()                     // "0.1.0"
app.BasePath("config.yaml")      // "./config.yaml"
app.StoragePath("logs")           // "./storage/logs"
app.IsBooted()                    // true
app.Bound("config")               // true

// 直接使用容器方法
svc, err := app.Make("config")
cfg := app.MustMake("config").(contracts.Config)
```

---

## 三、facades.Config()

返回 `contracts.Config`，基于 Viper 实现，支持 YAML 点号路径和环境变量。

### 接口

```go
type Config interface {
    Env(key string, defaultValue ...any) any
    Get(key string, defaultValue ...any) any
    GetString(key string, defaultValue ...string) string
    GetInt(key string, defaultValue ...int) int
    GetBool(key string, defaultValue ...bool) bool
    Set(key string, value any)
}
```

### 示例

```go
cfg := facades.Config()

// 读取配置（支持点号路径）
host := cfg.GetString("server.host", "0.0.0.0")
port := cfg.GetInt("server.port", 3000)
debug := cfg.GetBool("server.prefork", false)

// 读取嵌套配置
driver := cfg.GetString("database.driver", "sqlite")

// 读取环境变量
secret := cfg.Env("APP_SECRET", "default-secret")

// 运行时修改（不持久化到文件）
cfg.Set("server.port", 8080)

// 获取任意类型
val := cfg.Get("database.max_open_conns", 100)
```

---

## 四、facades.Log()

返回 `contracts.Log`，基于 Logrus 实现，支持 6 个日志级别 + 结构化字段。

### 接口

```go
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

### 示例

```go
log := facades.Log()

// 基本日志
log.Info("server started")
log.Debugf("listening on port %d", 3000)
log.Error("something went wrong")

// 结构化日志（链式调用，返回新 Log 实例）
log.WithField("user_id", 42).Info("user logged in")

log.WithFields(map[string]any{
    "method": "POST",
    "path":   "/api/users",
    "status": 201,
}).Info("request completed")

// Fatal / Panic 会终止程序
log.Fatal("unrecoverable error")
```

### 配置项

```yaml
log:
  level: debug        # debug / info / warn / error / fatal / panic
  format: color       # color（带颜色文本）/ json / text
  output_path: storage/logs/app.log   # 为空则仅输出到控制台
  max_size: 100       # 单个日志文件最大 MB
  max_backups: 5      # 保留旧文件数
  max_age: 30         # 保留天数
  compress: false     # 是否压缩旧文件
```

---

## 五、facades.Cache()

返回 `contracts.Cache`，支持多 Store、标签分组、原子操作、Hash 操作和分布式锁。

### 基础 CRUD

```go
c := facades.Cache()

// 写入（TTL 10 分钟）
c.Put("user:1:name", "Alice", 10*time.Minute)

// 永不过期
c.Forever("site:title", "GoFast")

// 读取
name := c.GetString("user:1:name", "unknown")
count := c.GetInt("visitor_count", 0)
exists := c.Has("user:1:name")

// 获取后删除
val := c.Pull("one_time_code")

// 删除
c.Forget("user:1:name")

// 清空全部
c.Flush()
```

### Remember（缓存穿透保护）

```go
// 存在则直接返回，不存在则调用回调生成值并缓存
result, err := c.Remember("user:1", time.Hour, func() (any, error) {
    return db.FindUser(1)  // 只在缓存未命中时执行
})

// 永不过期版本
result, err := c.RememberForever("config:site", func() (any, error) {
    return loadSiteConfig()
})
```

### 原子操作

```go
newVal, err := c.Increment("page_views")        // +1
newVal, err := c.Increment("page_views", 5)     // +5
newVal, err := c.Decrement("stock", 1)           // -1
```

### 批量操作

```go
// 批量获取
vals := c.Many([]string{"key1", "key2", "key3"})

// 批量设置
c.PutMany(map[string]any{
    "key1": "val1",
    "key2": "val2",
}, 30*time.Minute)
```

### Hash 操作

```go
c.HSet("user:1", "name", "Alice")
c.HSet("user:1", "age", 25)

name, _ := c.HGet("user:1", "name")      // "Alice"
all, _ := c.HGetAll("user:1")             // map[name:Alice age:25]
length := c.HLen("user:1")                // 2
keys, _ := c.HKeys("user:1")              // ["name", "age"]
exists := c.HExists("user:1", "name")     // true

c.HDel("user:1", "age")
```

### 标签分组

```go
// 按标签写入
c.Tags("users", "admin").Put("user:1", userData, time.Hour)
c.Tags("users").Put("user:2", userData2, time.Hour)

// 按标签清除
c.Tags("users").Flush()  // 清除所有带 "users" 标签的缓存
```

### 分布式锁

```go
lock := c.Lock("process:order:123", 30*time.Second)

if lock.Acquire() {
    defer lock.Release()
    // 临界区代码 ...
}

// 阻塞等待锁
lock.Block(10*time.Second, func() {
    // 获取锁后执行
})
```

### 多 Store

```go
// 使用指定的缓存驱动
memStore := c.Store("memory")
memStore.Put("key", "value", time.Hour)

// 默认 Store 由配置 cache.driver 决定
```

---

## 六、facades.Orm()

返回 `contracts.Orm`，基于 GORM 实现，支持 MySQL / PostgreSQL / SQLite / SQL Server。

### 接口

```go
type Orm interface {
    DB() *gorm.DB
    Ping() error
    Close() error
}
```

### 示例

```go
orm := facades.Orm()
db := orm.DB()

// 自动迁移
db.AutoMigrate(&User{})

// CRUD
db.Create(&User{Name: "Alice", Email: "alice@example.com"})

var user User
db.First(&user, "name = ?", "Alice")

db.Model(&user).Update("name", "Bob")

db.Delete(&user)

// 复杂查询
var users []User
db.Where("created_at > ?", yesterday).
    Order("created_at DESC").
    Limit(10).
    Find(&users)

// 事务
db.Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&order).Error; err != nil {
        return err
    }
    if err := tx.Create(&payment).Error; err != nil {
        return err
    }
    return nil
})

// 测试连接
if err := orm.Ping(); err != nil {
    log.Fatal("database unreachable")
}
```

### 配置项

```yaml
database:
  driver: sqlite          # sqlite / mysql / postgres / mssql
  # SQLite
  dsn: "gofast.db"
  # MySQL
  # host: 127.0.0.1
  # port: 3306
  # database: gofast
  # username: root
  # password: secret
  # charset: utf8mb4
  max_idle_conns: 10
  max_open_conns: 100
  conn_max_lifetime: 60   # 分钟
  conn_max_idle_time: 30  # 分钟
```

---

## 七、facades.Route()

返回 `contracts.Route`，基于 Fiber v2 实现，支持链式注册、路由组回调、中间件和控制器自注册。

### 接口

```go
type Route interface {
    Run(addr ...string) error
    Shutdown() error

    Get(path string, handler HandlerFunc) Route
    Post(path string, handler HandlerFunc) Route
    Put(path string, handler HandlerFunc) Route
    Delete(path string, handler HandlerFunc) Route
    Patch(path string, handler HandlerFunc) Route
    Head(path string, handler HandlerFunc) Route
    Options(path string, handler HandlerFunc) Route

    // Group 创建路由组。
    // args 支持 HandlerFunc（中间件）和 func(Route)（回调），可任意组合。
    Group(prefix string, args ...any) Route
    Use(middleware ...HandlerFunc) Route

    // Register 批量注册控制器（自动处理 Prefix / Middleware / Boot）。
    Register(controllers ...Controller) Route
}
```

### 控制器注册（推荐方式）

控制器实现 `contracts.Controller` 接口，通过 `Boot()` 声明自己的路由：

```go
type UserController struct{}

func (c *UserController) Prefix() string { return "/users" }

func (c *UserController) Boot(r contracts.Route) {
    r.Get("/", c.Index)
    r.Get("/:id", c.Show)
    r.Post("/", c.Store)
    r.Put("/:id", c.Update)
    r.Delete("/:id", c.Destroy)
}
```

路由文件只做编排：

```go
r := facades.Route()

r.Group("/admin", adminMiddleware.AdminAuth, func(admin contracts.Route) {
    admin.Register(
        &controllers.UserController{},
        &controllers.PostController{},
    )
})
```

### Group 回调 + 行内中间件

```go
r := facades.Route()

// 回调方式（推荐）
r.Group("/api/v1", appMiddleware.Auth, func(v1 contracts.Route) {
    v1.Register(&controllers.UserController{})
})

// 多层嵌套
r.Group("/api", func(api contracts.Route) {
    api.Get("/ping", pingHandler)

    api.Group("/v1", appMiddleware.Auth, func(v1 contracts.Route) {
        v1.Register(&controllers.OrderController{})
    })
})

// 链式调用（仍支持）
admin := r.Group("/admin")
admin.Use(adminMiddleware.AdminAuth)
admin.Get("/hello", helloHandler)
```

### 启动服务器

```go
r := facades.Route()
r.Run()                  // 使用配置中的 host:port
r.Run(":8080")           // 指定地址
```

### 内置特性

- **健康检查**：`GET /health` → `{"status": "ok"}`（自动注册）
- **Panic 恢复**：`recover` 中间件自动捕获 panic
- **请求 ID**：`requestid` 中间件自动为每个请求生成唯一 ID
- **CORS**：根据配置自动启用跨域支持

### 配置项

```yaml
server:
  name: GoFast
  host: 0.0.0.0
  port: 3000
  mode: debug             # debug 模式开启 stack trace
  read_timeout_sec: 30
  write_timeout_sec: 30
  idle_timeout_sec: 120
  shutdown_timeout_sec: 10
  prefork: false
  body_limit_mb: 10
  cors_allow_origins:
    - "*"
```

---

## 八、facades.Storage()

返回 `contracts.Storage`，支持多磁盘管理，默认提供本地文件系统驱动。

### 常用操作

```go
s := facades.Storage()

// 写入 / 读取
s.Put("hello.txt", "world")
content, err := s.Get("hello.txt")
bytes, err := s.GetBytes("hello.txt")

// 判断存在
s.Exists("hello.txt")   // true
s.Missing("hello.txt")  // false

// 文件信息
size, _ := s.Size("hello.txt")
mime, _ := s.MimeType("hello.txt")
mtime, _ := s.LastModified("hello.txt")
url := s.Url("hello.txt")
fullPath := s.Path("hello.txt")

// 复制 / 移动 / 删除
s.Copy("hello.txt", "backup/hello.txt")
s.Move("old.txt", "new.txt")
s.Delete("temp1.txt", "temp2.txt")

// 目录操作
s.MakeDirectory("uploads/images")
files, _ := s.Files("uploads")
allFiles, _ := s.AllFiles("uploads")  // 递归
dirs, _ := s.Directories("uploads")
s.DeleteDirectory("uploads/temp")
```

### 多磁盘切换

```go
// 使用指定磁盘
local := s.Disk("local")
local.Put("file.txt", "data")

// 配置多个磁盘
// filesystem:
//   default: local
//   disks:
//     local:
//       driver: local
//       root: storage/app
//       url: /storage
//     public:
//       driver: local
//       root: storage/public
//       url: /public
```

---

## 九、facades.Validator()

返回 `contracts.Validation`，基于 `go-playground/validator` 实现，使用 `binding` tag 定义规则。

### 接口

```go
type Validation interface {
    Validate(obj any) error
    RegisterRule(rule any) error
}
```

### 示例

```go
v := facades.Validator()

type CreateUserRequest struct {
    Name  string `json:"name"  binding:"required,min=2,max=50"`
    Email string `json:"email" binding:"required,email"`
    Age   int    `json:"age"   binding:"required,gte=1,lte=150"`
}

req := CreateUserRequest{
    Name:  "A",         // 太短，min=2
    Email: "invalid",   // 不是合法 email
    Age:   0,           // 不满足 gte=1
}

if err := v.Validate(req); err != nil {
    // err 包含所有验证失败的字段信息
    fmt.Println(err)
}
```

### 字段名解析

验证器优先使用 `json` → `form` → `query` tag 的值作为错误消息中的字段名，而非 Go 结构体字段名：

```go
type Input struct {
    UserName string `json:"user_name" binding:"required"`
}
// 错误消息中的字段名为 "user_name"，而非 "UserName"
```

---

## 十、Facade 内部原理

每个 Facade 本质上只是容器解析的语法糖，实现模式完全一致：

```go
// facades/xxx.go 的标准模式
package facades

import "go-fast/framework/contracts"

func Xxx() contracts.XxxInterface {
    return App().MustMake("xxx_key").(contracts.XxxInterface)
}
```

如果你新增了自定义 Provider 并注册到容器中，可以很方便地为其添加一个 Facade：

```go
// facades/sms.go
package facades

import "go-fast/framework/contracts"

func Sms() contracts.Sms {
    return App().MustMake("sms").(contracts.Sms)
}
```

---

## 十一、相关文档

- [路由设计文档](route.md) — Group 回调、控制器自注册、中间件策略
- [容器 API](container.md) — 了解 Bind / Singleton / Make 等底层方法
- [编写自定义 Provider](service-provider.md) — 将自定义服务注册到容器
- [插件开发指南](plugins.md) — 开发可复用的第三方插件


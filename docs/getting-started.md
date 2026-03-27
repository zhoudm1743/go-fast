# GoFast 快速开始

> GoFast 是一个基于 Go 的快速开发框架。

---

## 一、环境要求

| 依赖 | 最低版本 |
|------|---------|
| Go   | 1.25+   |
| 数据库（可选） | MySQL 5.7+ / PostgreSQL 12+ / SQLite 3 / SQL Server 2017+ |

---

## 二、获取代码

```bash
git clone https://github.com/zhoudm1743/go-fast.git
cd GoFast
```

---

## 三、配置文件

在 `config/config.yaml` 中配置：

```yaml
# ── 服务器 ──────────────────────────────────────
server:
  name: GoFast
  host: 0.0.0.0
  port: 3000
  mode: debug                # debug / release
  read_timeout_sec: 30
  write_timeout_sec: 30
  idle_timeout_sec: 120
  shutdown_timeout_sec: 10
  prefork: false
  body_limit_mb: 10
  cors_allow_origins:
    - "*"

# ── 数据库 ──────────────────────────────────────
database:
  driver: sqlite              # sqlite / mysql / postgres / mssql
  dsn: "gofast.db"            # SQLite 文件路径
  # MySQL 示例:
  # driver: mysql
  # host: 127.0.0.1
  # port: 3306
  # database: gofast
  # username: root
  # password: secret
  # charset: utf8mb4
  max_idle_conns: 10
  max_open_conns: 100
  conn_max_lifetime: 60       # 分钟
  conn_max_idle_time: 30      # 分钟

# ── 日志 ────────────────────────────────────────
log:
  level: debug                # debug / info / warn / error / fatal / panic
  format: color               # color / json / text
  output_path: storage/logs/app.log
  max_size: 100               # MB
  max_backups: 5
  max_age: 30                 # 天
  compress: false

# ── 文件系统 ────────────────────────────────────
filesystem:
  default: local
  disks:
    local:
      driver: local
      root: storage/app
      url: /storage

# ── 缓存 ────────────────────────────────────────
cache:
  driver: memory
  memory:
    shard_count: 32
    clean_interval: 60        # 秒
```

---

## 四、启动应用

```bash
go run main.go
```

启动成功后你将看到：

```
[GoFast] v0.1.0 booted

 ┌───────────────────────────────────────────────────┐
 │                    GoFast v2                       │
 │                   http://0.0.0.0:3000              │
 └───────────────────────────────────────────────────┘
```

访问 `http://localhost:3000/health` 即可看到：

```json
{"status": "ok"}
```

---

## 五、启动流程

GoFast 的启动流程与 Goravel 对齐，核心只有三步：

```
main.go
  └── bootstrap.Boot()
        ├── 1. foundation.NewApplication(".")     // 创建应用 & 容器
        ├── 2. app.SetProviders(providers())      // 设置 Provider 列表
        ├── 3. app.Boot()                         // Register → Boot 全部 Provider
        └── 4. facades.SetApp(app)                // 设置全局门面入口
```

内置 Provider 按顺序加载：

| 顺序 | Provider | 容器 Key | 说明 |
|------|----------|---------|------|
| 1 | `config.ServiceProvider` | `config` | 读取配置文件 |
| 2 | `log.ServiceProvider` | `log` | 初始化日志（依赖 config） |
| 3 | `cache.ServiceProvider` | `cache` | 初始化缓存（依赖 config） |
| 4 | `database.ServiceProvider` | `db` / ~~`orm`~~ | 连接数据库（依赖 config + log）；`orm` 已 Deprecated |
| 5 | `filesystem.ServiceProvider` | `storage` | 文件系统（依赖 config） |
| 6 | `validation.ServiceProvider` | `validator` | 验证器 |
| 7 | `http.ServiceProvider` | `route` | HTTP 路由（依赖 config） |

---

## 六、注册路由

GoFast 采用**控制器自注册**模式：控制器在 `Boot()` 中声明路由，路由文件只做编排。

路由文件按模块拆分：

- `routes/app.go` —— 前台 / 用户端路由
- `routes/admin.go` —— 后台管理路由
- `routes/api.go` —— 统一入口，仅负责调用 `RegisterApp()` / `RegisterAdmin()`

### 控制器示例

控制器实现 `contracts.Controller` 接口，可选实现 `contracts.Prefixer` 声明前缀：

```go
// app/http/app/controllers/user_controller.go
package controllers

import "github.com/zhoudm1743/go-fast/framework/contracts"

type UserController struct{}

func (c *UserController) Prefix() string { return "/user" }

func (c *UserController) Boot(r contracts.Route) {
    r.Get("/profile", c.Profile)
    r.Put("/profile", c.UpdateProfile)
}

func (c *UserController) Profile(ctx contracts.Context) error {
    return ctx.Response().Success(map[string]string{"name": "Alice"})
}

func (c *UserController) UpdateProfile(ctx contracts.Context) error {
    return ctx.Response().Success(nil)
}
```

### 前台路由示例 `routes/app.go`

```go
package routes

import (
    appControllers "github.com/zhoudm1743/go-fast/app/http/app/controllers"
    appMiddleware "github.com/zhoudm1743/go-fast/app/http/app/middleware"
    "github.com/zhoudm1743/go-fast/framework/contracts"
    "github.com/zhoudm1743/go-fast/framework/facades"
)

func RegisterApp() {
    r := facades.Route()

    // 公开接口（无需登录）
    r.Get("/api/ping", func(ctx contracts.Context) error {
        return ctx.JSON(200, map[string]string{"message": "pong"})
    })

    // 需要登录的接口
    r.Group("/api/v1", appMiddleware.Auth, func(v1 contracts.Route) {
        v1.Register(
            &appControllers.UserController{},
        )
    })
}
```

### 后台路由示例 `routes/admin.go`

```go
package routes

import (
    adminControllers "github.com/zhoudm1743/go-fast/app/http/admin/controllers"
    adminMiddleware "github.com/zhoudm1743/go-fast/app/http/admin/middleware"
    "github.com/zhoudm1743/go-fast/framework/contracts"
    "github.com/zhoudm1743/go-fast/framework/facades"
)

func RegisterAdmin() {
    facades.Route().Group("/admin", adminMiddleware.AdminAuth, func(admin contracts.Route) {
        admin.Register(
            &adminControllers.UserController{},
        )
    })
}
```

### 总入口 `routes/api.go`

```go
package routes

func Register() {
    RegisterApp()
    RegisterAdmin()
}
```

`main.go` 已内置路由注册调用：

```go
app := bootstrap.Boot()
routes.Register()       // 注册路由
facades.Route().Run()   // 启动服务器
```

---

## 七、使用 Facade 访问服务

GoFast 提供静态门面（Facade），让你无需手动从容器解析服务：

```go
import "github.com/zhoudm1743/go-fast/framework/facades"

// 配置
port := facades.Config().GetInt("server.port", 3000)

// 日志
facades.Log().Info("server started")
facades.Log().WithField("user_id", 1).Info("user login")

// 数据库（推荐）
q := facades.DB().Query()
q.Create(&User{Name: "Alice"})

var users []User
facades.DB().Query().Where("status = ?", 1).Order("created_at DESC").Find(&users)

// 缓存
facades.Cache().Put("key", "value", 10*time.Minute)
val := facades.Cache().GetString("key")

// 文件存储
facades.Storage().Put("hello.txt", "world")
content, _ := facades.Storage().Get("hello.txt")

// 验证
err := facades.Validator().Validate(input)
```

---

## 八、定义模型

GoFast 内置 UUID v7 主键自动生成：

```go
package models

import "github.com/zhoudm1743/go-fast/framework/database"

// User 业务模型，嵌入 database.Model 即自带 UUID v7 主键。
type User struct {
    database.Model
    Name  string `gorm:"size:100" json:"name"`
    Email string `gorm:"size:200;uniqueIndex" json:"email"`
}

// 带软删除
type Article struct {
    database.ModelWithSoftDelete
    Title   string `gorm:"size:255" json:"title"`
    Content string `gorm:"type:text" json:"content"`
}
```

`database.Model` 提供：
- `ID` — UUID v7 字符串主键（36 位，自动生成）
- `CreatedAt` — Unix 时间戳（自动设置）
- `UpdatedAt` — Unix 时间戳（自动更新）

**自动迁移**：实现 `DBMigrator` 接口，框架启动后自动调用：

```go
// 在 ServiceProvider 中实现
func (p *AppProvider) MigrateDB(db contracts.DB) error {
    return db.AutoMigrate(&models.User{}, &models.Article{})
}
```

---

## 九、优雅关闭

按 `Ctrl+C` 发送 SIGINT 信号后，GoFast 会：

1. 停止接收新请求
2. 按注册逆序执行 shutdown hooks：
   - HTTP 服务器优雅关闭
   - 缓存 GC 停止
   - 数据库连接关闭
3. 进程退出

---

## 十、下一步

- [数据库文档](database/README.md) — 完整的数据库操作指南（查询、事务、分页、多连接等）
- [路由设计文档](route.md) — Group 回调、控制器自注册、中间件策略
- [控制器开发指南](controller.md) — 控制器、验证、数据库、中间件完整示例
- [容器 API](container.md) — 了解服务容器的完整接口
- [Facade 使用说明](facade.md) — 每个 Facade 的详细用法
- [编写自定义 Provider](service-provider.md) — 扩展框架能力
- [插件开发指南](plugins.md) — 开发可复用插件


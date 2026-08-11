# GoFast 路由设计指南

> 本文介绍 GoFast 的 HTTP 路由体系：路由注册、控制器声明、路由组、中间件策略和静态资源。
> GoFast 支持 **Gin** 和 **Fiber** 双引擎驱动，通过配置文件切换，业务代码无需改动。

---

## 一、路由契约

`contracts.Route` 接口定义了统一的路由 API：

| 方法 | 说明 |
|------|------|
| `Get(path, handler)` | GET 路由 |
| `Post(path, handler)` | POST 路由 |
| `Put(path, handler)` | PUT 路由 |
| `Delete(path, handler)` | DELETE 路由 |
| `Patch(path, handler)` | PATCH 路由 |
| `Head(path, handler)` | HEAD 路由 |
| `Options(path, handler)` | OPTIONS 路由 |
| `Group(prefix, args...)` | 路由组（支持中间件和子组回调） |
| `Use(middleware...)` | 注册中间件 |
| `Register(controllers...)` | 批量注册控制器 |
| `Static(urlPrefix, dir)` | 静态文件服务 |
| `StaticFS(urlPrefix, fs)` | 嵌入静态文件服务 |
| `Run(addr...)` | 启动 HTTP 服务器 |
| `Shutdown()` | 优雅关闭 |

所有 Handler 签名统一为 `contracts.HandlerFunc`：

```go
type HandlerFunc func(ctx contracts.Context) error
```

---

## 二、引擎切换

```yaml
# config.yaml
http:
  driver: fiber  # 可选 gin | fiber，默认 fiber
  host: 0.0.0.0
  port: 8080
```

切换引擎后无需修改任何业务代码，Controller / Middleware 完全不变。

---

## 三、路由文件结构

GoFast 推荐按业务域拆分路由，总入口只做编排：

```
routes/
├── api.go      # 总入口，调用各模块 Register
├── app.go      # 前台/用户端路由
└── admin.go    # 后台管理路由
```

### 3.1 简单路由声明

```go
package routes

import (
    "net/http"
    "github.com/zhoudm1743/go-fast-framework/contracts"
    "github.com/zhoudm1743/go-fast-framework/facades"
)

func RegisterApp() {
    r := facades.Http.Route()

    // 健康检查
    r.Get("/api/ping", func(ctx contracts.Context) error {
        return ctx.Response().Success(map[string]string{"message": "pong"})
    })

    // 路径参数
    r.Get("/api/users/:id", func(ctx contracts.Context) error {
        id := ctx.Param("id")
        return ctx.Response().Success(id)
    })
}
```

### 3.2 路由组 (Group)

路由组提供共享前缀和中间件。`Group` 支持两种形式的参数：

```go
r.Group("/admin", middleware1, middleware2, func(admin contracts.Route) {
    // admin 自动携带 /admin 前缀
    admin.Get("/dashboard", handler)  // → GET /admin/dashboard
    admin.Get("/users", handler)      // → GET /admin/users
})
```

### 3.3 控制器自注册

控制器实现 `contracts.Controller` 接口，通过 `Boot()` 声明自己的路由：

```go
// app/http/admin/controllers/user_controller.go
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

路由文件中使用 `Register()` 批量注册：

```go
facades.Http.Route().Group("/api/v1", middleware.Auth, func(v1 contracts.Route) {
    v1.Register(
        &controllers.UserController{},
        &controllers.OrderController{},
    )
})
```

框架自动检测控制器是否实现 `Prefixer` / `Middlewarer` 接口，自动应用前缀和中间件。

---

## 四、中间件

### 4.1 中间件签名

```go
type HandlerFunc func(ctx contracts.Context) error
```

与 Handler 签名完全相同。通过调用 `ctx.Next()` 放行到下一个处理器。

### 4.2 中间件的三种应用范围

| 范围 | 方式 | 影响 |
|------|------|------|
| **全局** | `r.Use(middleware)` | 所有路由 |
| **路由组** | `r.Group(prefix, middleware, ...)` | 该组内所有路由 |
| **控制器** | 实现 `Middlewarer` 接口 | 该控制器的所有路由 |

### 4.3 编写中间件

```go
// JWT 鉴权中间件
func Auth(ctx contracts.Context) error {
    token := ctx.Header("Authorization")
    token = strings.TrimPrefix(token, "Bearer ")

    if token == "" {
        return ctx.Response().Unauthorized("请先登录")
    }

    claims, err := facades.JWT.Parse(token)
    if err != nil {
        return ctx.Response().Unauthorized("令牌无效")
    }

    ctx.WithValue("user_id", claims.UID)
    return ctx.Next()
}
```

### 4.4 CORS / Recovery

Gin 和 Fiber 引擎内置了 CORS 和 Panic Recovery 中间件。如需自定义，可使用 `Use()` 手动注册。

---

## 五、静态文件

### 5.1 本地目录

```go
r.Static("/static", "resources/static")
// → /static/logo.png → resources/static/logo.png
```

### 5.2 go:embed 嵌入

```go
//go:embed assets/*
var assets embed.FS

func init() {
    subFS, _ := fs.Sub(assets, "assets")
    facades.Http.Route().StaticFS("/assets", http.FS(subFS))
}
```

---

## 六、路由列表

调试时查看所有已注册路由：

```go
routes := facades.Http.Route().Routes()
for _, r := range routes {
    fmt.Printf("%-8s %s\n", r.Method, r.Path)
}
```

---

## 七、完整路由文件示例

```go
// routes/api.go
package routes

func Register() {
    RegisterApp()
    RegisterAdmin()
}
```

```go
// routes/admin.go
package routes

import (
    adminCtrl "go-fast/app/http/admin/controllers"
    "go-fast/app/http/admin/middleware"
    "github.com/zhoudm1743/go-fast-framework/contracts"
    "github.com/zhoudm1743/go-fast-framework/facades"
)

func RegisterAdmin() {
    facades.Http.Route().Group("/admin", middleware.AdminAuth, func(admin contracts.Route) {
        admin.Register(
            &adminCtrl.DashboardController{},
            &adminCtrl.UserController{},
        )
    })
}
```

```go
// routes/app.go
package routes

import (
    appCtrl "go-fast/app/http/app/controllers"
    "github.com/zhoudm1743/go-fast-framework/contracts"
    "github.com/zhoudm1743/go-fast-framework/facades"
)

func RegisterApp() {
    r := facades.Http.Route()

    r.Get("/api/ping", func(ctx contracts.Context) error {
        return ctx.Response().Success(map[string]string{"message": "ok"})
    })

    r.Group("/api/v1", func(v1 contracts.Route) {
        v1.Register(
            &appCtrl.AuthController{},    // /api/v1/login, /api/v1/register
            &appCtrl.UserController{},    // /api/v1/user/*  (通过 Prefix)
        )
    })
}
```

---

## 八、相关文档

- [控制器开发指南](controller.md) — 控制器编写、请求解析、统一响应
- [Facade 使用说明](facade.md) — HTTP / JWT / Log / Cache 等门面 API
- [快速开始](getting-started.md) — 项目初始化与启动

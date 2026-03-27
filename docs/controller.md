# GoFast 控制器开发指南

> 本文介绍如何在 GoFast 中编写控制器，涵盖：控制器自注册、请求解析、表单验证、数据库操作、统一响应结构和中间件。
> 所有控制器代码 **只依赖 `github.com/zhoudm1743/go-fast/framework/contracts` 和 `github.com/zhoudm1743/go-fast/framework/facades`**，不引入任何底层 HTTP 框架包。
>
> 控制器通过实现 `contracts.Controller` 接口在 `Boot()` 中声明自己的路由，路由文件只做编排。
> 详见 [路由设计文档](route.md)。

---

## 一、核心概念

### HandlerFunc

```go
// 所有控制器方法的统一签名
type HandlerFunc func(ctx contracts.Context) error
```

### contracts.Context

`Context` 是传给每个 Handler 的核心对象，封装了请求读取、响应发送和上下文存储：

| 方法 | 说明 |
|------|------|
| `ctx.Param("id")` | 读取 URL 路径参数 |
| `ctx.Query("page", "1")` | 读取查询字符串，支持默认值 |
| `ctx.Header("Authorization")` | 读取请求头 |
| `ctx.IP()` | 客户端 IP |
| `ctx.BodyRaw()` | 原始请求体字节 |
| `ctx.Bind(&req)` | 解析所有来源 + 自动验证（见下文）|
| `ctx.JSON(code, data)` | 发送 JSON 响应 |
| `ctx.String(code, s)` | 发送纯文本响应 |
| `ctx.Status(code)` | 设置状态码（链式） |
| `ctx.SetHeader(k, v)` | 设置响应头（链式） |
| `ctx.Value("key")` | 读取 Middleware 存储的值 |
| `ctx.WithValue("key", v)` | 向下游 Handler 传值 |
| `ctx.Next()` | 调用下一个 Middleware |

### ctx.Bind 多源解析

`ctx.Bind(&req)` 会按顺序从三个来源填充 struct 字段，最后统一验证：

| 优先级 | Tag | 数据来源 | 示例 |
|--------|-----|---------|------|
| 1 | `uri:"xxx"` | URL 路径参数 | `/users/:id` 中的 `id` |
| 2 | `query:"xxx"` | Query String | `?page=1&size=20` |
| 3 | `json:"xxx"` / `form:"xxx"` | 请求体（POST/PUT body） | `{"name":"Alice"}` |

**一个 struct 可以混合多种 tag**，框架自动从各来源填充：

```go
// PUT /api/v1/users/:id
type UpdateUserRequest struct {
    ID    string `uri:"id"     binding:"required"`           // ← 来自 URL 路径
    Name  string `json:"name"  binding:"omitempty,min=2"`    // ← 来自 JSON body
    Email string `json:"email" binding:"omitempty,email"`    // ← 来自 JSON body
}

func (c *UserController) Update(ctx contracts.Context) error {
    var req UpdateUserRequest
    if err := ctx.Bind(&req); err != nil {  // 一次调用填满所有字段并验证
        return fail(ctx, http.StatusUnprocessableEntity, err.Error())
    }
    // req.ID    = "abc-123"   (from /users/abc-123)
    // req.Name  = "Alice"     (from body)
    // req.Email = ""          (omitempty，未传则跳过验证)
    ...
}
```

**GET 请求（纯 query 参数）**：

```go
// GET /api/v1/users?page=2&size=10&email=alice
type ListUserRequest struct {
    Page  int    `query:"page"  binding:"omitempty,gte=1"`
    Size  int    `query:"size"  binding:"omitempty,gte=1,lte=100"`
    Email string `query:"email" binding:"omitempty,email"`
}

func (c *UserController) Index(ctx contracts.Context) error {
    var req ListUserRequest
    if err := ctx.Bind(&req); err != nil {
        return fail(ctx, http.StatusUnprocessableEntity, err.Error())
    }
    // req.Page = 2, req.Size = 10, req.Email = "alice"
    ...
}
```

---

## 二、目录结构

```
app/
├── http/
│   ├── admin/
│   │   ├── controllers/
│   │   │   └── user_controller.go   # 后台管理：完整 CRUD
│   │   └── middleware/
│   │       └── auth.go              # 后台鉴权
│   └── app/
│       ├── controllers/
│       │   └── user_controller.go   # 前台/用户端：个人中心、自服务
│       └── middleware/
│           └── auth.go              # 前台鉴权
└── models/
    └── user.go                      # 数据模型
```

---

## 三、定义模型

模型放在 `app/models/`，嵌入 `database.Model` 自动获得 UUID v7 主键：

```go
// app/models/user.go
package models

import "github.com/zhoudm1743/go-fast/framework/database"

type User struct {
    database.Model                                      // ID + CreatedAt + UpdatedAt
    Name     string `gorm:"size:100;not null"        json:"name"`
    Email    string `gorm:"size:200;uniqueIndex;not null" json:"email"`
    Password string `gorm:"size:255;not null"        json:"-"` // json:"-" 不序列化
}
```

`database.Model` 提供：
- `ID` — UUID v7 字符串（自动生成，无需手动赋值）
- `CreatedAt` — Unix 毫秒时间戳（自动填充）
- `UpdatedAt` — Unix 毫秒时间戳（自动更新）

**自动迁移**（建议放在 `bootstrap/app.go` 的 `Boot()` 之后或专用迁移脚本中）：

```go
facades.Orm().DB().AutoMigrate(&models.User{})
```

---

## 四、编写控制器

### 4.1 控制器分层建议

GoFast 推荐按业务入口分模块：

- `app/http/admin/controllers/`：后台管理控制器，通常是完整 CRUD、列表筛选、批量操作
- `app/http/app/controllers/`：前台/用户端控制器，通常是“当前用户自己的数据”

例如同样是“用户”能力：

- 后台：`/admin/api/v1/users` —— 管理任意用户
- 前台：`/api/v1/user/profile` —— 只操作当前登录用户自己的资料

### 4.2 后台控制器基本结构

每个控制器实现 `contracts.Controller` 接口（`Boot` 方法），可选实现 `contracts.Prefixer`（`Prefix` 方法）声明路由前缀：

```go
// app/http/admin/controllers/user_controller.go
package controllers

import (
    "net/http"
    "go-fast/app/models"
    "go-fast/framework/contracts"
    "go-fast/framework/facades"
)

type UserController struct{}

// Prefix 路由前缀（实现 contracts.Prefixer）。
func (c *UserController) Prefix() string { return "/users" }

// Boot 声明路由（实现 contracts.Controller）。
func (c *UserController) Boot(r contracts.Route) {
    r.Get("/", c.Index)          // GET    /users
    r.Get("/:id", c.Show)       // GET    /users/:id
    r.Post("/", c.Store)        // POST   /users
    r.Put("/:id", c.Update)     // PUT    /users/:id
    r.Delete("/:id", c.Destroy) // DELETE /users/:id
}
```

### 4.3 请求体与验证规则

使用 `binding` tag 同时定义 JSON 字段名和验证规则：

```go
type CreateUserRequest struct {
    Name     string `json:"name"     binding:"required,min=2,max=50"`
    Email    string `json:"email"    binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
}

type UpdateUserRequest struct {
    ID    string `uri:"id"     binding:"required"`
    Name  string `json:"name"  binding:"omitempty,min=2,max=50"`
    Email string `json:"email" binding:"omitempty,email"`
}
```

**常用 binding 规则**：

| 规则 | 说明 |
|------|------|
| `required` | 必填 |
| `omitempty` | 为空时跳过验证（用于更新） |
| `min=N,max=N` | 字符串最小/最大长度 |
| `gte=N,lte=N` | 数字范围 |
| `email` | 合法邮箱格式 |
| `url` | 合法 URL |
| `len=N` | 固定长度 |
| `oneof=a b c` | 枚举值 |

### 4.4 统一响应结构

GoFast 已在 `framework/http` 内置标准响应结构：

```go
type Response struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    any    `json:"data,omitempty"`
}
```

控制器无需再重复定义 `apiResponse` / `ok` / `fail`，直接使用：

```go
ctx.Response().Success(data)                          // HTTP 200, code=0
ctx.Response().Created(data)                          // HTTP 201, code=0
ctx.Response().Fail(http.StatusBadRequest, "参数错误") // 默认业务码 code=1
ctx.Response().Unauthorized()                         // HTTP 401
ctx.Response().Forbidden()                            // HTTP 403
ctx.Response().NotFound("用户不存在")                 // HTTP 404
ctx.Response().Validation(err)                        // HTTP 422
ctx.Response().Paginate(list, total, page, size)      // 标准分页结构
ctx.Response().Build(200, 10001, "自定义消息", data)   // 完全自定义
```

---

## 五、后台管理控制器示例（Admin 模块）

后台管理一般放在：`app/http/admin/controllers/`，路由前缀通常是 `/admin/api/v1/...`。

### 5.1 列表 GET /users

```go
func (c *UserController) Index(ctx contracts.Context) error {
    var req ListUserRequest
    if err := ctx.Bind(&req); err != nil { // query tag 自动填充
        return ctx.Response().Validation(err)
    }
    if req.Page == 0 { req.Page = 1 }
    if req.Size == 0 { req.Size = 20 }

    db := facades.Orm().DB().Model(&models.User{}).Order("created_at DESC")
    if req.Email != "" {
        db = db.Where("email LIKE ?", "%"+req.Email+"%")
    }

    var total int64
    db.Count(&total)

    var users []models.User
    err := db.Offset((req.Page-1)*req.Size).Limit(req.Size).Find(&users).Error

    if err != nil {
        facades.Log().Errorf("admin list users: %v", err)
        return ctx.Response().Fail(http.StatusInternalServerError, "查询失败")
    }

    return ctx.Response().Paginate(users, total, req.Page, req.Size)
}
```

### 5.2 详情 GET /users/:id

```go
func (c *UserController) Show(ctx contracts.Context) error {
    var req UserIDRequest
    if err := ctx.Bind(&req); err != nil { // uri tag 自动填充
        return ctx.Response().Validation(err)
    }

    var user models.User
    if err := facades.Orm().DB().First(&user, "id = ?", req.ID).Error; err != nil {
        return ctx.Response().NotFound("用户不存在")
    }

    return ctx.Response().Success(user)
}
```

### 5.3 创建 POST /users

```go
func (c *UserController) Store(ctx contracts.Context) error {
    var req CreateUserRequest

    // Bind = JSON解析 + binding验证，一步完成
    if err := ctx.Bind(&req); err != nil {
        return ctx.Response().Validation(err)
    }

    // 业务唯一性校验
    var count int64
    facades.Orm().DB().Model(&models.User{}).Where("email = ?", req.Email).Count(&count)
    if count > 0 {
        return ctx.Response().Fail(http.StatusConflict, "邮箱已存在")
    }

    user := models.User{
        Name:     req.Name,
        Email:    req.Email,
        Password: hashPassword(req.Password), // 生产环境必须哈希
    }

    if err := facades.Orm().DB().Create(&user).Error; err != nil {
        facades.Log().Errorf("admin create user: %v", err)
        return ctx.Response().Fail(http.StatusInternalServerError, "创建失败")
    }

    return ctx.Response().Created(user)
}
```

### 5.4 更新 PUT /users/:id

```go
func (c *UserController) Update(ctx contracts.Context) error {
    var req UpdateUserRequest
    if err := ctx.Bind(&req); err != nil { // uri + json body 一次绑定
        return ctx.Response().Validation(err)
    }

    var user models.User
    if err := facades.Orm().DB().First(&user, "id = ?", req.ID).Error; err != nil {
        return ctx.Response().NotFound("用户不存在")
    }

    updates := map[string]any{}
    if req.Name != ""  { updates["name"] = req.Name }
    if req.Email != "" { updates["email"] = req.Email }

    if len(updates) > 0 {
        if err := facades.Orm().DB().Model(&user).Updates(updates).Error; err != nil {
            return ctx.Response().Fail(http.StatusInternalServerError, "更新失败")
        }
    }

    return ctx.Response().Success(user)
}
```

### 5.5 删除 DELETE /users/:id

```go
func (c *UserController) Destroy(ctx contracts.Context) error {
    var req UserIDRequest
    if err := ctx.Bind(&req); err != nil {
        return ctx.Response().Validation(err)
    }

    var user models.User
    if err := facades.Orm().DB().First(&user, "id = ?", req.ID).Error; err != nil {
        return ctx.Response().NotFound("用户不存在")
    }

    if err := facades.Orm().DB().Delete(&user).Error; err != nil {
        return ctx.Response().Fail(http.StatusInternalServerError, "删除失败")
    }

    return ctx.Response().Success(nil)
}
```

---

## 六、注册路由

控制器通过 `Boot()` 声明自己的路由后，路由文件只需用 `Register()` 编排即可。

GoFast 推荐拆分为三份路由文件：

- `routes/app.go` —— 前台/用户端路由
- `routes/admin.go` —— 后台管理路由
- `routes/api.go` —— 统一入口，只负责调用两者

### 6.1 后台路由 `routes/admin.go`

```go
package routes

import (
    adminControllers "go-fast/app/http/admin/controllers"
    adminMiddleware "go-fast/app/http/admin/middleware"
    "go-fast/framework/contracts"
    "go-fast/framework/facades"
)

func RegisterAdmin() {
    facades.Route().Group("/admin", adminMiddleware.AdminAuth, func(admin contracts.Route) {
        admin.Register(
            &adminControllers.UserController{},
        )
    })
}
```

### 6.2 前台路由 `routes/app.go`

```go
package routes

import (
    appControllers "go-fast/app/http/app/controllers"
    appMiddleware "go-fast/app/http/app/middleware"
    "go-fast/framework/contracts"
    "go-fast/framework/facades"
)

func RegisterApp() {
    r := facades.Route()

    // 公开接口
    r.Get("/api/ping", func(ctx contracts.Context) error {
        return ctx.JSON(200, map[string]string{"message": "pong"})
    })

    // 需登录接口
    r.Group("/api/v1", appMiddleware.Auth, func(v1 contracts.Route) {
        v1.Register(
            &appControllers.UserController{},
        )
    })
}
```

### 6.3 总入口 `routes/api.go`

```go
package routes

func Register() {
    RegisterApp()
    RegisterAdmin()
}
```

---

## 七、中间件

中间件与控制器签名完全一致：`func(contracts.Context) error`，调用 `ctx.Next()` 继续执行后续 Handler。

### 7.1 前台认证中间件

```go
// app/http/app/middleware/auth.go
package middleware

import (
    "net/http"
    "strings"
    "go-fast/framework/contracts"
)

// Auth 前台用户鉴权中间件。
func Auth(ctx contracts.Context) error {
    token := ctx.Header("Authorization")
    token = strings.TrimPrefix(token, "Bearer ")

    if token == "" {
        return ctx.Response().Unauthorized("请先登录")
    }

    // TODO: 验证 JWT，成功后将 user_id 写入上下文
    // ctx.WithValue("user_id", userID)

    return ctx.Next() // 继续执行下一个 Handler
}
```

### 7.2 后台认证中间件

```go
// app/http/admin/middleware/auth.go
package middleware

import (
    "net/http"
    "strings"
    "go-fast/framework/contracts"
)

func AdminAuth(ctx contracts.Context) error {
    token := strings.TrimPrefix(ctx.Header("Authorization"), "Bearer ")
    if token == "" {
        return ctx.Response().Unauthorized("未授权")
    }

    // TODO: 验证 admin token，并确认其角色为管理员
    // if !isAdmin(adminID) { return ctx.Response().Forbidden("无权限访问后台") }
    // ctx.WithValue("admin_id", adminID)

    return ctx.Next()
}
```

### 7.3 中间件的三种使用方式

**组级中间件** —— 写在路由文件的 Group 上，整组共享：

```go
facades.Route().Group("/admin", adminMiddleware.AdminAuth, func(admin contracts.Route) {
    admin.Register(&adminControllers.UserController{})
})
```

**控制器级中间件** —— 实现 `contracts.Middlewarer` 接口，控制器独享：

```go
func (c *UserController) Middleware() []contracts.HandlerFunc {
    return []contracts.HandlerFunc{middleware.Auth}
}
```

**路由级中间件** —— 在 `Boot()` 里用 Group 包裹部分路由：

```go
func (c *AuthController) Boot(r contracts.Route) {
    r.Post("/login", c.Login)       // 公开

    r.Group("/", middleware.Auth, func(g contracts.Route) {
        g.Get("/me", c.Me)          // 需要鉴权
        g.Post("/logout", c.Logout) // 需要鉴权
    })
}
```

### 7.4 在 Handler 中读取中间件传递的值

```go
func (c *UserController) Profile(ctx contracts.Context) error {
    userID := ctx.Value("user_id").(string)   // 由 appMiddleware.Auth 写入

    var user models.User
    facades.Orm().DB().First(&user, "id = ?", userID)

    return ctx.Response().Success(user)
}
```

---

## 八、事务处理

```go
func (c OrderController) Store(ctx contracts.Context) error {
    var req CreateOrderRequest
    if err := ctx.Bind(&req); err != nil {
        return ctx.Response().Fail(http.StatusUnprocessableEntity, err.Error())
    }

    var order models.Order
    err := facades.Orm().DB().Transaction(func(tx *gorm.DB) error {
        order = models.Order{UserID: req.UserID, Amount: req.Amount}
        if err := tx.Create(&order).Error; err != nil {
            return err
        }
        // 扣减库存
        if err := tx.Model(&models.Product{}).
            Where("id = ?", req.ProductID).
            UpdateColumn("stock", gorm.Expr("stock - ?", 1)).Error; err != nil {
            return err
        }
        return nil
    })

    if err != nil {
        facades.Log().Errorf("create order failed: %v", err)
        return ctx.Response().Fail(http.StatusInternalServerError, "下单失败")
    }

    return ctx.Response().Created(order)
}
```

---

## 九、日志记录

```go
func (c *UserController) Store(ctx contracts.Context) error {
    // 结构化日志（带请求 IP）
    log := facades.Log().WithFields(map[string]any{
        "ip":     ctx.IP(),
        "method": ctx.Method(),
        "path":   ctx.Path(),
    })

    var req CreateUserRequest
    if err := ctx.Bind(&req); err != nil {
        log.Warnf("validation failed: %v", err)
        return ctx.Response().Fail(http.StatusUnprocessableEntity, err.Error())
    }

    // ... 业务逻辑 ...

    log.WithField("user_email", req.Email).Info("user created")
    return ctx.Response().Created(user)
}
```

---

## 十、相关文档

- [路由设计文档](route.md) — Group 回调、控制器自注册、中间件策略
- [Facade 使用说明](facade.md) — Config / Log / Cache / Orm 详细 API
- [编写自定义 Provider](service-provider.md) — 注册自定义服务
- [快速开始](getting-started.md) — 项目初始化与启动


# GoFast 控制器开发指南

> 本文介绍如何在 GoFast 中编写控制器，涵盖：路由绑定、请求解析、表单验证、数据库操作、统一响应结构和中间件。
> 所有控制器代码 **只依赖 `go-fast/framework/contracts` 和 `go-fast/framework/facades`**，不引入任何底层 HTTP 框架包。

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
| `ctx.Bind(&req)` | 解析 JSON/Form 体 + 自动验证 |
| `ctx.JSON(code, data)` | 发送 JSON 响应 |
| `ctx.String(code, s)` | 发送纯文本响应 |
| `ctx.Status(code)` | 设置状态码（链式） |
| `ctx.SetHeader(k, v)` | 设置响应头（链式） |
| `ctx.Value("key")` | 读取 Middleware 存储的值 |
| `ctx.WithValue("key", v)` | 向下游 Handler 传值 |
| `ctx.Next()` | 调用下一个 Middleware |

---

## 二、目录结构

```
app/
├── http/
│   ├── controllers/
│   │   ├── user_controller.go    # 用户控制器
│   │   └── ...
│   └── middleware/
│       ├── auth.go               # 认证中间件
│       └── ...
└── models/
    └── user.go                   # 数据模型
```

---

## 三、定义模型

模型放在 `app/models/`，嵌入 `database.Model` 自动获得 UUID v7 主键：

```go
// app/models/user.go
package models

import "go-fast/framework/database"

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

### 4.1 基本结构

```go
// app/http/controllers/user_controller.go
package controllers

import (
    "net/http"
    "go-fast/app/models"
    "go-fast/framework/contracts"
    "go-fast/framework/facades"
)

type UserController struct{}
```

### 4.2 请求体与验证规则

使用 `binding` tag 同时定义 JSON 字段名和验证规则：

```go
type CreateUserRequest struct {
    Name     string `json:"name"     binding:"required,min=2,max=50"`
    Email    string `json:"email"    binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
}

type UpdateUserRequest struct {
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

### 4.3 统一响应结构

```go
type apiResponse struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    any    `json:"data,omitempty"`
}

func ok(ctx contracts.Context, data any) error {
    return ctx.JSON(http.StatusOK, apiResponse{Code: 0, Message: "ok", Data: data})
}

func created(ctx contracts.Context, data any) error {
    return ctx.JSON(http.StatusCreated, apiResponse{Code: 0, Message: "ok", Data: data})
}

func fail(ctx contracts.Context, code int, msg string) error {
    return ctx.JSON(code, apiResponse{Code: 1, Message: msg})
}
```

---

## 五、完整 CRUD 示例

### 5.1 列表 GET /api/v1/users

```go
func (c UserController) Index(ctx contracts.Context) error {
    // 查询参数：分页
    page  := ctx.Query("page", "1")
    limit := 20

    var users []models.User
    result := facades.Orm().DB().
        Order("created_at DESC").
        Limit(limit).
        Find(&users)

    if result.Error != nil {
        facades.Log().Errorf("list users: %v", result.Error)
        return fail(ctx, http.StatusInternalServerError, "查询失败")
    }

    return ok(ctx, map[string]any{
        "list":  users,
        "total": result.RowsAffected,
        "page":  page,
    })
}
```

### 5.2 详情 GET /api/v1/users/:id

```go
func (c UserController) Show(ctx contracts.Context) error {
    id := ctx.Param("id")   // URL 路径参数

    var user models.User
    if err := facades.Orm().DB().First(&user, "id = ?", id).Error; err != nil {
        return fail(ctx, http.StatusNotFound, "用户不存在")
    }

    return ok(ctx, user)
}
```

### 5.3 创建 POST /api/v1/users

```go
func (c UserController) Store(ctx contracts.Context) error {
    var req CreateUserRequest

    // Bind = JSON解析 + binding验证，一步完成
    if err := ctx.Bind(&req); err != nil {
        return fail(ctx, http.StatusUnprocessableEntity, err.Error())
    }

    // 业务唯一性校验
    var count int64
    facades.Orm().DB().Model(&models.User{}).Where("email = ?", req.Email).Count(&count)
    if count > 0 {
        return fail(ctx, http.StatusConflict, "邮箱已存在")
    }

    user := models.User{
        Name:     req.Name,
        Email:    req.Email,
        Password: hashPassword(req.Password), // 生产环境必须哈希
    }

    if err := facades.Orm().DB().Create(&user).Error; err != nil {
        facades.Log().Errorf("create user: %v", err)
        return fail(ctx, http.StatusInternalServerError, "创建失败")
    }

    return created(ctx, user)
}
```

### 5.4 更新 PUT /api/v1/users/:id

```go
func (c UserController) Update(ctx contracts.Context) error {
    id := ctx.Param("id")

    var user models.User
    if err := facades.Orm().DB().First(&user, "id = ?", id).Error; err != nil {
        return fail(ctx, http.StatusNotFound, "用户不存在")
    }

    var req UpdateUserRequest
    if err := ctx.Bind(&req); err != nil {
        return fail(ctx, http.StatusUnprocessableEntity, err.Error())
    }

    updates := map[string]any{}
    if req.Name != ""  { updates["name"] = req.Name }
    if req.Email != "" { updates["email"] = req.Email }

    if len(updates) > 0 {
        if err := facades.Orm().DB().Model(&user).Updates(updates).Error; err != nil {
            return fail(ctx, http.StatusInternalServerError, "更新失败")
        }
    }

    return ok(ctx, user)
}
```

### 5.5 删除 DELETE /api/v1/users/:id

```go
func (c UserController) Destroy(ctx contracts.Context) error {
    id := ctx.Param("id")

    var user models.User
    if err := facades.Orm().DB().First(&user, "id = ?", id).Error; err != nil {
        return fail(ctx, http.StatusNotFound, "用户不存在")
    }

    if err := facades.Orm().DB().Delete(&user).Error; err != nil {
        return fail(ctx, http.StatusInternalServerError, "删除失败")
    }

    return ok(ctx, nil)
}
```

---

## 六、注册路由

在 `routes/api.go` 中绑定控制器：

```go
// routes/api.go
package routes

import (
    "go-fast/app/http/controllers"
    "go-fast/framework/contracts"
    "go-fast/framework/facades"
)

func Register() {
    r := facades.Route()

    // 简单闭包
    r.Get("/api/ping", func(ctx contracts.Context) error {
        return ctx.JSON(200, map[string]string{"message": "pong"})
    })

    // 控制器方法
    user := controllers.UserController{}
    v1 := r.Group("/api/v1")
    v1.Get("/users",       user.Index)
    v1.Get("/users/:id",   user.Show)
    v1.Post("/users",      user.Store)
    v1.Put("/users/:id",   user.Update)
    v1.Delete("/users/:id", user.Destroy)
}
```

---

## 七、中间件

中间件与控制器签名完全一致：`func(contracts.Context) error`，调用 `ctx.Next()` 继续执行后续 Handler。

### 7.1 编写认证中间件

```go
// app/http/middleware/auth.go
package middleware

import (
    "net/http"
    "strings"
    "go-fast/framework/contracts"
)

// Auth JWT 认证中间件。
func Auth(ctx contracts.Context) error {
    token := ctx.Header("Authorization")
    token = strings.TrimPrefix(token, "Bearer ")

    if token == "" {
        return ctx.JSON(http.StatusUnauthorized, map[string]string{
            "message": "未授权",
        })
    }

    // 解析 JWT，获取用户 ID
    userID, err := parseJWT(token)
    if err != nil {
        return ctx.JSON(http.StatusUnauthorized, map[string]string{
            "message": "token 无效",
        })
    }

    // 将用户 ID 存入上下文，供后续 Handler 读取
    ctx.WithValue("user_id", userID)

    return ctx.Next() // 继续执行下一个 Handler
}
```

### 7.2 在路由组中使用中间件

```go
// routes/api.go
import "go-fast/app/http/middleware"

func Register() {
    r := facades.Route()

    // 公开路由（无需认证）
    r.Post("/api/auth/login",  authController.Login)
    r.Post("/api/auth/register", authController.Register)

    // 需要认证的路由组
    api := r.Group("/api/v1")
    api.Use(middleware.Auth)  // 对该组所有路由生效
    api.Get("/users",     user.Index)
    api.Get("/profile",   profileController.Show)
}
```

### 7.3 在 Handler 中读取中间件传递的值

```go
func (c UserController) Profile(ctx contracts.Context) error {
    userID := ctx.Value("user_id").(string)   // 由 Auth 中间件写入

    var user models.User
    facades.Orm().DB().First(&user, "id = ?", userID)

    return ok(ctx, user)
}
```

---

## 八、事务处理

```go
func (c OrderController) Store(ctx contracts.Context) error {
    var req CreateOrderRequest
    if err := ctx.Bind(&req); err != nil {
        return fail(ctx, http.StatusUnprocessableEntity, err.Error())
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
        return fail(ctx, http.StatusInternalServerError, "下单失败")
    }

    return created(ctx, order)
}
```

---

## 九、日志记录

```go
func (c UserController) Store(ctx contracts.Context) error {
    // 结构化日志（带请求 IP）
    log := facades.Log().WithFields(map[string]any{
        "ip":     ctx.IP(),
        "method": ctx.Method(),
        "path":   ctx.Path(),
    })

    var req CreateUserRequest
    if err := ctx.Bind(&req); err != nil {
        log.Warnf("validation failed: %v", err)
        return fail(ctx, http.StatusUnprocessableEntity, err.Error())
    }

    // ... 业务逻辑 ...

    log.WithField("user_email", req.Email).Info("user created")
    return created(ctx, user)
}
```

---

## 十、相关文档

- [Facade 使用说明](facade.md) — Config / Log / Cache / Orm 详细 API
- [编写自定义 Provider](service-provider.md) — 注册自定义服务
- [快速开始](getting-started.md) — 项目初始化与启动


# 错误处理

## Sentinel Errors

GoFast 定义了一套与底层 ORM 无关的**标准错误**，无论底层使用哪种驱动，错误类型保持一致：

```go
// framework/contracts/query.go
var (
    ErrRecordNotFound    = errors.New("record not found")      // 查询无结果
    ErrDuplicatedKey     = errors.New("duplicated key")        // 唯一约束冲突
    ErrInvalidTransaction = errors.New("invalid transaction")  // 无效事务
    ErrDeadlock          = errors.New("deadlock detected")     // 死锁
    ErrQueryTimeout      = errors.New("query timeout")         // 查询超时
    ErrConnFailed        = errors.New("connection failed")     // 连接失败
    ErrUnsupported       = errors.New("operation not supported by driver")
)
```

---

## errors.Is — 精确判断错误类型

使用 `errors.Is()` 而非直接比较，以正确处理错误链：

```go
import (
    "errors"
    "github.com/zhoudm1743/go-fast/framework/contracts"
)

if err := facades.DB().Query().First(&user, "id = ?", id); err != nil {
    switch {
    case errors.Is(err, contracts.ErrRecordNotFound):
        return ctx.Response().NotFound("用户不存在")
    default:
        facades.Log().Errorf("查询用户失败: %v", err)
        return ctx.Response().Fail(http.StatusInternalServerError, "服务器错误")
    }
}
```

---

## 各操作错误处理模式

### 查询单条记录

```go
var user models.User
err := facades.DB().Query().First(&user, "id = ?", id)
if err != nil {
    if errors.Is(err, contracts.ErrRecordNotFound) {
        return ctx.Response().NotFound("用户不存在")
    }
    return ctx.Response().Fail(http.StatusInternalServerError, "查询失败")
}
```

### 创建记录（唯一约束冲突）

```go
if err := facades.DB().Query().Create(&user); err != nil {
    if errors.Is(err, contracts.ErrDuplicatedKey) {
        return ctx.Response().Fail(http.StatusConflict, "邮箱已被注册")
    }
    return ctx.Response().Fail(http.StatusInternalServerError, "创建失败")
}
```

### 更新记录（检查是否命中）

```go
result := facades.DB().Query().
    Model(&models.User{}).
    Where("id = ?", id).
    UpdateResult("status", 0)

if result.Error != nil {
    return ctx.Response().Fail(http.StatusInternalServerError, "更新失败")
}
if result.IsZeroRow() {
    return ctx.Response().NotFound("记录不存在或无变化")
}
```

### 事务错误

```go
err := facades.DB().Transaction(func(tx contracts.Query) error {
    if err := tx.Create(&order); err != nil {
        return err // 自动回滚，错误会传递到外层
    }
    return nil
})

if err != nil {
    if errors.Is(err, contracts.ErrDeadlock) {
        // 死锁，可以重试
        return retryTransaction()
    }
    return ctx.Response().Fail(http.StatusInternalServerError, "操作失败")
}
```

---

## Result 结构体

写操作的 `Result` 变体同时返回错误和影响行数：

```go
type Result struct {
    RowsAffected int64
    Error        error
}

// IsZeroRow 返回 true 当操作成功但没有行被影响
func (r Result) IsZeroRow() bool {
    return r.Error == nil && r.RowsAffected == 0
}
```

使用示例：

```go
result := facades.DB().Query().
    Where("user_id = ? AND status = ?", userID, "pending").
    DeleteResult(&models.Order{})

if result.Error != nil {
    return result.Error
}
fmt.Printf("清理了 %d 笔待处理订单\n", result.RowsAffected)
```

---

## 错误日志最佳实践

```go
func (c *UserController) Show(ctx contracts.Context) error {
    id := ctx.Param("id")
    var user models.User

    if err := facades.DB().Query().First(&user, "id = ?", id); err != nil {
        if errors.Is(err, contracts.ErrRecordNotFound) {
            // 预期错误，无需记录日志
            return ctx.Response().NotFound("用户不存在")
        }
        // 非预期错误，记录详细日志
        facades.Log().Errorf("[UserController.Show] 查询用户 %s 失败: %v", id, err)
        return ctx.Response().Fail(http.StatusInternalServerError, "服务器内部错误")
    }

    return ctx.Response().Success(user)
}
```

---

## 封装通用错误处理

```go
// app/support/db_helper.go
package support

import (
    "errors"
    "net/http"

    "github.com/zhoudm1743/go-fast/framework/contracts"
)

// HandleDBError 将数据库错误转换为 HTTP 响应
func HandleDBError(ctx contracts.Context, err error, resource string) error {
    if err == nil {
        return nil
    }
    switch {
    case errors.Is(err, contracts.ErrRecordNotFound):
        return ctx.Response().NotFound(resource + "不存在")
    case errors.Is(err, contracts.ErrDuplicatedKey):
        return ctx.Response().Fail(http.StatusConflict, resource+"已存在")
    case errors.Is(err, contracts.ErrDeadlock):
        return ctx.Response().Fail(http.StatusServiceUnavailable, "服务繁忙，请稍后重试")
    case errors.Is(err, contracts.ErrQueryTimeout):
        return ctx.Response().Fail(http.StatusGatewayTimeout, "查询超时")
    default:
        facades.Log().Errorf("database error: %v", err)
        return ctx.Response().Fail(http.StatusInternalServerError, "服务器内部错误")
    }
}

// 使用
func (c *UserController) Show(ctx contracts.Context) error {
    var user models.User
    if err := facades.DB().Query().First(&user, "id = ?", ctx.Param("id")); err != nil {
        return support.HandleDBError(ctx, err, "用户")
    }
    return ctx.Response().Success(user)
}
```

---

## 错误映射表

| 数据库错误 | Sentinel Error | HTTP 状态码建议 |
|-----------|----------------|----------------|
| 记录不存在 | `ErrRecordNotFound` | 404 |
| 唯一键冲突 | `ErrDuplicatedKey` | 409 Conflict |
| 事务无效 | `ErrInvalidTransaction` | 500 |
| 死锁 | `ErrDeadlock` | 503 / 重试 |
| 查询超时 | `ErrQueryTimeout` | 504 |
| 连接失败 | `ErrConnFailed` | 503 |
| 其他未知错误 | — | 500 |


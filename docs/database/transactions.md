# 事务

## 自动事务（推荐）

传入一个函数，返回 `nil` 自动提交，返回任意 `error` 自动回滚：

```go
err := facades.DB().Transaction(func(tx contracts.Query) error {
    // tx 是事务内的查询构建器，用法与 facades.DB().Query() 完全一致

    // 创建订单
    order := &models.Order{UserID: userID, Amount: 100}
    if err := tx.Create(&order); err != nil {
        return err // 触发回滚
    }

    // 扣减余额
    if err := tx.Model(&models.Wallet{}).
        Where("user_id = ?", userID).
        Update("balance", gorm.Expr("balance - ?", 100)); err != nil {
        return err // 触发回滚
    }

    return nil // 提交事务
})

if err != nil {
    // 事务已回滚，处理错误
}
```

---

## 手动事务

适合需要在事务过程中执行额外逻辑（如第三方 API 调用）的场景：

```go
// 开启事务
tx := facades.DB().Query().Begin()

// 执行操作
if err := tx.Create(&order); err != nil {
    tx.Rollback()
    return err
}

if err := tx.Model(&wallet).Update("balance", wallet.Balance-amount); err != nil {
    tx.Rollback()
    return err
}

// 提交
if err := tx.Commit(); err != nil {
    return err
}
```

---

## SavePoint — 部分回滚

在事务内设置保存点，可以只回滚到某个中间状态：

```go
err := facades.DB().Transaction(func(tx contracts.Query) error {
    // 第一步：创建用户（如果失败，整个事务回滚）
    if err := tx.Create(&user); err != nil {
        return err
    }

    // 设置保存点
    tx.SavePoint("after_user")

    // 第二步：创建用户配置（如果失败，只回滚到保存点）
    if err := tx.Create(&profile); err != nil {
        tx.RollbackTo("after_user") // 只回滚 profile 的创建
        // 继续执行，不影响 user 的创建
    }

    return nil
})
```

---

## 事务隔离级别

使用预定义选项：

```go
import "github.com/zhoudm1743/go-fast/framework/contracts"

// READ COMMITTED
err := facades.DB().Transaction(func(tx contracts.Query) error {
    // ...
}, contracts.TxReadCommitted)

// REPEATABLE READ（MySQL 默认）
err := facades.DB().Transaction(func(tx contracts.Query) error {
    // ...
}, contracts.TxRepeatableRead)

// SERIALIZABLE（最高隔离级别）
err := facades.DB().Transaction(func(tx contracts.Query) error {
    // ...
}, contracts.TxSerializable)

// 只读事务
err := facades.DB().Transaction(func(tx contracts.Query) error {
    // ...
}, contracts.TxReadOnly)
```

自定义隔离级别：

```go
import (
    "database/sql"
    "github.com/zhoudm1743/go-fast/framework/contracts"
)

opts := contracts.TxOpts(sql.LevelReadCommitted, false)
err := facades.DB().Transaction(func(tx contracts.Query) error {
    // ...
}, opts)
```

---

## 嵌套事务

GoFast 支持在事务中嵌套调用（使用 SavePoint 实现）：

```go
err := facades.DB().Transaction(func(tx contracts.Query) error {
    tx.Create(&user)

    // 嵌套事务（自动使用 SavePoint）
    if err := tx.Transaction(func(tx2 contracts.Query) error {
        return tx2.Create(&profile)
    }); err != nil {
        // 内层事务失败，外层仍可继续
        log.Println("profile 创建失败，跳过:", err)
    }

    return nil
})
```

---

## 在 Service 层传递事务

将 `tx` 作为参数传递给 Service 方法，保持代码分层清晰：

```go
// service/order_service.go
type OrderService struct{}

func (s *OrderService) CreateOrder(tx contracts.Query, userID string, amount int64) (*models.Order, error) {
    order := &models.Order{UserID: userID, Amount: amount}
    if err := tx.Create(&order); err != nil {
        return nil, err
    }
    return order, nil
}

func (s *OrderService) DeductBalance(tx contracts.Query, userID string, amount int64) error {
    result := tx.Model(&models.Wallet{}).
        Where("user_id = ? AND balance >= ?", userID, amount).
        UpdateResult("balance", gorm.Expr("balance - ?", amount))
    if result.Error != nil {
        return result.Error
    }
    if result.IsZeroRow() {
        return errors.New("余额不足")
    }
    return nil
}

// controller/order_controller.go
func (c *OrderController) Create(ctx contracts.Context) error {
    svc := &service.OrderService{}

    err := facades.DB().Transaction(func(tx contracts.Query) error {
        order, err := svc.CreateOrder(tx, userID, 100)
        if err != nil {
            return err
        }
        if err := svc.DeductBalance(tx, userID, order.Amount); err != nil {
            return err
        }
        return nil
    })

    if err != nil {
        return ctx.Response().Fail(http.StatusBadRequest, err.Error())
    }
    return ctx.Response().Success(nil)
}
```

---

## 最佳实践

| 场景 | 建议 |
|------|------|
| 多表写操作 | 始终使用事务，保证原子性 |
| 读操作 | 无需事务，使用 `TxReadOnly` 可提升性能 |
| 长事务 | 避免在事务中调用外部 API，减少锁持有时间 |
| 错误处理 | 自动事务中只需 `return err`，框架自动回滚 |
| Service 分层 | 将 `tx contracts.Query` 作为参数传递 |
| 并发写同一行 | 使用 `Lock(contracts.LockForUpdate)` 加悲观锁 |


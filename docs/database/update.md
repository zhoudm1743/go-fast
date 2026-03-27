# 更新记录

## Save — 保存所有字段

`Save` 会保存模型的**全部字段**（包括零值），相当于 `UPDATE ... SET all_columns`。

```go
var user models.User
facades.DB().Query().First(&user, "id = ?", id)

user.Name = "Alice Updated"
user.Email = "new@example.com"

if err := facades.DB().Query().Save(&user); err != nil {
    // ...
}
```

> ⚠️ `Save` 会更新所有字段，若字段为零值（空字符串、0 等）也会被写入。

---

## Update — 更新单列

```go
// 更新单个字段
if err := facades.DB().Query().
    Model(&user).
    Update("name", "Alice"); err != nil {
    // ...
}

// 带条件（不推荐，应先 First 再 Update）
if err := facades.DB().Query().
    Model(&models.User{}).
    Where("email = ?", "old@example.com").
    Update("email", "new@example.com"); err != nil {
    // ...
}
```

---

## Updates — 更新多列

### 使用 Map（推荐，支持零值）

```go
updates := map[string]any{
    "name":   "Alice",
    "status": 0,       // 零值也会被写入
    "score":  9.5,
}

if err := facades.DB().Query().
    Model(&user).
    Updates(updates); err != nil {
    // ...
}
```

### 使用 Struct（零值字段会被忽略）

```go
type UpdateUserInput struct {
    Name  string
    Email string
}

if err := facades.DB().Query().
    Model(&user).
    Updates(&UpdateUserInput{Name: "Alice", Email: ""}); err != nil {
    // Email 为空字符串，会被忽略，不会更新
}
```

### 完整控制器示例

```go
func (c *UserController) Update(ctx contracts.Context) error {
    var req UpdateUserRequest
    if err := ctx.Bind(&req); err != nil {
        return ctx.Response().Validation(err)
    }

    // 先查出记录
    var user models.User
    if err := facades.DB().Query().First(&user, "id = ?", req.ID); err != nil {
        if errors.Is(err, contracts.ErrRecordNotFound) {
            return ctx.Response().NotFound("用户不存在")
        }
        return ctx.Response().Fail(http.StatusInternalServerError, "查询失败")
    }

    // 只更新非空字段
    updates := map[string]any{}
    if req.Name != "" {
        updates["name"] = req.Name
    }
    if req.Email != "" {
        updates["email"] = req.Email
    }

    if len(updates) > 0 {
        if err := facades.DB().Query().Model(&user).Updates(updates); err != nil {
            if errors.Is(err, contracts.ErrDuplicatedKey) {
                return ctx.Response().Fail(http.StatusConflict, "邮箱已存在")
            }
            return ctx.Response().Fail(http.StatusInternalServerError, "更新失败")
        }
    }

    return ctx.Response().Success(user)
}
```

---

## Select / Omit 限制更新字段

```go
// 只更新 name 和 email
if err := facades.DB().Query().
    Model(&user).
    Select("name", "email").
    Updates(&user); err != nil {
    // ...
}

// 排除 password，更新其余字段
if err := facades.DB().Query().
    Model(&user).
    Omit("password").
    Save(&user); err != nil {
    // ...
}
```

---

## Result 变体 — 获取影响行数

```go
result := facades.DB().Query().
    Model(&user).
    UpdateResult("status", 1)

if result.Error != nil {
    // 处理错误
}
if result.IsZeroRow() {
    // 没有记录被更新（WHERE 没有命中）
}
fmt.Println("更新了", result.RowsAffected, "条")
```

```go
result := facades.DB().Query().
    Model(&models.User{}).
    Where("last_login_at < ?", thirtyDaysAgo).
    UpdatesResult(map[string]any{"status": 0})

fmt.Printf("批量禁用 %d 个用户\n", result.RowsAffected)
```

---

## 使用 SQL 表达式更新

```go
// 浏览数 +1
if err := facades.DB().Query().
    Model(&post).
    Update("views", gorm.Expr("views + ?", 1)); err != nil {
    // ...
}
```

> 需要导入 `gorm.io/gorm`，此为 GORM 驱动的特性，使用时注意驱动耦合。

推荐用原生 SQL 替代：

```go
if err := facades.DB().Query().
    Exec("UPDATE posts SET views = views + 1 WHERE id = ?", post.ID); err != nil {
    // ...
}
```

---

## 批量更新

```go
// 批量修改状态
if err := facades.DB().Query().
    Model(&models.User{}).
    Where("created_at < ?", cutoff).
    Updates(map[string]any{"status": 2}); err != nil {
    // ...
}
```

> ⚠️ 没有 `Where` 条件的批量更新会更新所有记录，务必谨慎。

---

## 在事务中更新

```go
err := facades.DB().Transaction(func(tx contracts.Query) error {
    if err := tx.Model(&order).Updates(map[string]any{"status": "paid"}); err != nil {
        return err
    }
    if err := tx.Model(&wallet).Update("balance", gorm.Expr("balance - ?", order.Amount)); err != nil {
        return err
    }
    return nil
})
```


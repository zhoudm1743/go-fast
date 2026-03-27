# 删除记录

## Delete — 删除记录

```go
// 先查出再删除（推荐，确保记录存在）
var user models.User
if err := facades.DB().Query().First(&user, "id = ?", id); err != nil {
    return err
}
if err := facades.DB().Query().Delete(&user); err != nil {
    return err
}

// 带条件直接删除（跳过查询）
if err := facades.DB().Query().
    Delete(&models.User{}, "id = ?", id); err != nil {
    // ...
}
```

---

## 软删除

嵌入 `database.ModelWithSoftDelete` 的模型，`Delete` 操作只设置 `deleted_at` 时间戳，**不物理删除数据**。

```go
type Article struct {
    database.ModelWithSoftDelete
    Title string `gorm:"size:200" json:"title"`
}

// 软删除：deleted_at = unix_now()
if err := facades.DB().Query().Delete(&article); err != nil {
    // ...
}

// 常规查询自动过滤软删除记录（WHERE deleted_at = 0）
var articles []models.Article
facades.DB().Query().Find(&articles) // 只返回未删除的
```

### Unscoped — 包含软删除记录的查询

```go
// 查询所有记录（含已软删除）
var all []models.Article
facades.DB().Query().Unscoped().Find(&all)

// 查询特定软删除记录
var article models.Article
facades.DB().Query().Unscoped().First(&article, "id = ?", id)
```

### OnlyTrashed — 只查询软删除记录

```go
// 只返回已软删除的记录
var trashed []models.Article
facades.DB().Query().OnlyTrashed().Find(&trashed)
```

### Restore — 恢复软删除记录

```go
// 先查出软删除记录
var article models.Article
if err := facades.DB().Query().Unscoped().First(&article, "id = ?", id); err != nil {
    return err
}

// 恢复：deleted_at = 0
if err := facades.DB().Query().Model(&article).Restore(); err != nil {
    return err
}
```

### ForceDelete — 物理删除

```go
// 物理删除软删除模型（绕过软删除机制）
if err := facades.DB().Query().ForceDelete(&article, "id = ?", id); err != nil {
    // ...
}

// 或先 Unscoped 再 Delete
if err := facades.DB().Query().Unscoped().Delete(&article); err != nil {
    // ...
}
```

---

## Result 变体 — 获取影响行数

```go
result := facades.DB().Query().DeleteResult(&user)
if result.Error != nil {
    return result.Error
}
if result.IsZeroRow() {
    // 没有记录被删除
    return ctx.Response().NotFound("记录不存在")
}
fmt.Println("已删除", result.RowsAffected, "条")
```

---

## 批量删除

```go
// 批量软删除
result := facades.DB().Query().
    Where("user_id = ? AND status = ?", userID, 0).
    DeleteResult(&models.Article{})

fmt.Printf("批量删除 %d 篇草稿\n", result.RowsAffected)
```

```go
// 批量硬删除（按 ID 列表）
ids := []string{"id1", "id2", "id3"}
if err := facades.DB().Query().
    Delete(&models.Tag{}, "id IN ?", ids); err != nil {
    // ...
}
```

---

## 控制器完整示例

```go
func (c *ArticleController) Destroy(ctx contracts.Context) error {
    id := ctx.Param("id")
    userID := ctx.Value("user_id").(string)

    // 查出记录，同时验证归属
    var article models.Article
    if err := facades.DB().Query().First(&article, "id = ? AND user_id = ?", id, userID); err != nil {
        if errors.Is(err, contracts.ErrRecordNotFound) {
            return ctx.Response().NotFound("文章不存在")
        }
        return ctx.Response().Fail(http.StatusInternalServerError, "查询失败")
    }

    // 软删除
    if err := facades.DB().Query().Delete(&article); err != nil {
        return ctx.Response().Fail(http.StatusInternalServerError, "删除失败")
    }

    return ctx.Response().Success(nil)
}

// 管理员接口：恢复被删除文章
func (c *ArticleController) Restore(ctx contracts.Context) error {
    id := ctx.Param("id")

    var article models.Article
    if err := facades.DB().Query().Unscoped().First(&article, "id = ?", id); err != nil {
        if errors.Is(err, contracts.ErrRecordNotFound) {
            return ctx.Response().NotFound("文章不存在")
        }
        return ctx.Response().Fail(http.StatusInternalServerError, "查询失败")
    }

    if err := facades.DB().Query().Model(&article).Restore(); err != nil {
        return ctx.Response().Fail(http.StatusInternalServerError, "恢复失败")
    }

    return ctx.Response().Success(article)
}
```

---

## 注意事项

| 场景 | 推荐做法 |
|------|---------|
| 删除前需验证 | 先 `First` 查出记录，确认存在且有权限后再删除 |
| 需要知道是否命中 | 使用 `DeleteResult` 检查 `RowsAffected` |
| 重要数据 | 使用软删除模型，保留数据可恢复 |
| 清理测试数据 | 使用 `ForceDelete` 或 `Unscoped().Delete()` |
| 批量删除 | 明确加 `Where` 条件，避免误删全表 |


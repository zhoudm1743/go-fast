# 高级查询

## Joins — 关联查询

### INNER JOIN

```go
type PostWithAuthor struct {
    models.Post
    AuthorName string `json:"author_name"`
}

var posts []PostWithAuthor
facades.DB().Query().
    Model(&models.Post{}).
    Select("posts.*, users.name as author_name").
    Joins("INNER JOIN users ON users.id = posts.user_id").
    Where("posts.status = ?", 1).
    Order("posts.created_at DESC").
    Find(&posts)
```

### LEFT JOIN

```go
var posts []PostWithAuthor
facades.DB().Query().
    Model(&models.Post{}).
    Select("posts.*, users.name as author_name").
    Joins("LEFT JOIN users ON users.id = posts.user_id").
    Find(&posts)
```

### 多表关联

```go
type PostDetail struct {
    ID           string `json:"id"`
    Title        string `json:"title"`
    AuthorName   string `json:"author_name"`
    CategoryName string `json:"category_name"`
}

var details []PostDetail
facades.DB().Query().
    Model(&models.Post{}).
    Select("posts.id, posts.title, users.name as author_name, categories.name as category_name").
    Joins("LEFT JOIN users ON users.id = posts.user_id").
    Joins("LEFT JOIN categories ON categories.id = posts.category_id").
    Where("posts.status = ?", 1).
    Scan(&details)
```

---

## Preload — 预加载关联

避免 N+1 问题，一次性加载关联数据。

```go
// 预加载单个关联
var posts []models.Post
facades.DB().Query().
    Preload("Author").
    Find(&posts)

// 多个关联
facades.DB().Query().
    Preload("Author").
    Preload("Category").
    Preload("Tags").
    Find(&posts)

// 带条件预加载（只加载已发布的评论）
facades.DB().Query().
    Preload("Comments", "status = ?", 1).
    Find(&posts)
```

### 嵌套预加载

```go
// 加载 Posts，同时加载每个 Post 的 Author 和 Tags
var users []models.User
facades.DB().Query().
    Preload("Posts.Author").
    Preload("Posts.Tags").
    Find(&users)
```

---

## Group — 分组

```go
type StatResult struct {
    UserID string `json:"user_id"`
    Count  int64  `json:"count"`
    Total  int64  `json:"total"`
}

var stats []StatResult
facades.DB().Query().
    Model(&models.Order{}).
    Select("user_id, COUNT(*) as count, SUM(amount) as total").
    Where("status = ?", "paid").
    Group("user_id").
    Scan(&stats)
```

---

## Having — 分组过滤

```go
// 查找发帖超过 10 篇的用户
type ActiveUser struct {
    UserID    string `json:"user_id"`
    PostCount int64  `json:"post_count"`
}

var active []ActiveUser
facades.DB().Query().
    Model(&models.Post{}).
    Select("user_id, COUNT(*) as post_count").
    Group("user_id").
    Having("COUNT(*) > ?", 10).
    Scan(&active)
```

---

## Distinct — 去重

```go
// 查询所有不同的城市
var cities []string
facades.DB().Query().
    Model(&models.User{}).
    Distinct("city").
    Pluck("city", &cities)

// 去重查询完整记录
var users []models.User
facades.DB().Query().
    Distinct("email").
    Find(&users)
```

---

## 子查询

使用 `Raw` 构建子查询：

```go
// 查询订单金额高于平均值的用户
var users []models.User
facades.DB().Query().
    Where("id IN (?)", facades.DB().Query().
        Model(&models.Order{}).
        Select("user_id").
        Where("amount > (?)",
            facades.DB().Query().Model(&models.Order{}).Select("AVG(amount)"),
        ),
    ).
    Find(&users)
```

> ⚠️ 复杂子查询推荐使用 [原生 SQL](./raw-sql.md) 以提高可读性。

---

## Lock — 悲观锁

### FOR UPDATE（写锁）

适用于先查后改的并发场景，防止幻读：

```go
err := facades.DB().Transaction(func(tx contracts.Query) error {
    var wallet models.Wallet
    // 加写锁，其他事务无法读取此行直到本事务提交
    if err := tx.Lock(contracts.LockForUpdate).First(&wallet, "user_id = ?", userID); err != nil {
        return err
    }

    if wallet.Balance < amount {
        return errors.New("余额不足")
    }

    return tx.Model(&wallet).Update("balance", wallet.Balance-amount)
})
```

### LOCK IN SHARE MODE（读锁）

```go
err := facades.DB().Transaction(func(tx contracts.Query) error {
    var stock models.Stock
    // 共享读锁，允许其他事务读，但不允许写
    tx.Lock(contracts.LockShareMode).First(&stock, "product_id = ?", productID)

    if stock.Quantity < quantity {
        return errors.New("库存不足")
    }
    return tx.Model(&stock).Update("quantity", stock.Quantity-quantity)
})
```

---

## ScanMap — 扫描为 Map 列表

无需预先定义结构体，直接扫描为 `[]map[string]any`：

```go
var rows []map[string]any
if err := facades.DB().Query().
    Model(&models.User{}).
    Select("id", "name", "email").
    Where("status = ?", 1).
    ScanMap(&rows); err != nil {
    // ...
}

for _, row := range rows {
    fmt.Println(row["name"], row["email"])
}
```

适合动态报表、数据导出等不确定结构的场景。

---

## WithContext — 携带上下文

结合请求超时控制数据库查询时长：

```go
func (c *PostController) Index(ctx contracts.Context) error {
    // 从 HTTP 请求上下文传入，超时自动取消
    q := facades.DB().Query().WithContext(ctx.Context())

    var posts []models.Post
    if err := q.Find(&posts); err != nil {
        return ctx.Response().Fail(http.StatusInternalServerError, "查询超时或失败")
    }
    return ctx.Response().Success(posts)
}
```

---

## Debug — 打印 SQL

```go
// 开启调试模式，输出完整 SQL 语句
facades.DB().Query().Debug().
    Where("status = ?", 1).
    Find(&users)

// 输出示例：
// [GoFast] [2026-03-27 10:00:00] [rows:10] SELECT * FROM `users` WHERE status = 1
```

---

## 综合示例：文章列表接口

```go
type ListPostRequest struct {
    Page       int    `query:"page"`
    Size       int    `query:"size"`
    Keyword    string `query:"keyword"`
    CategoryID string `query:"category_id"`
    Status     int    `query:"status"`
    UserID     string `query:"user_id"`
}

func (c *PostController) Index(ctx contracts.Context) error {
    var req ListPostRequest
    ctx.Bind(&req)
    if req.Page < 1 { req.Page = 1 }
    if req.Size < 1 { req.Size = 20 }

    q := facades.DB().Query().
        Model(&models.Post{}).
        Preload("Author").
        Preload("Category").
        Order("created_at DESC")

    if req.Keyword != "" {
        q = q.Where("title LIKE ? OR content LIKE ?",
            "%"+req.Keyword+"%", "%"+req.Keyword+"%")
    }
    if req.CategoryID != "" {
        q = q.Where("category_id = ?", req.CategoryID)
    }
    if req.Status != 0 {
        q = q.Where("status = ?", req.Status)
    }
    if req.UserID != "" {
        q = q.Where("user_id = ?", req.UserID)
    }

    var total int64
    q.Count(&total)

    var posts []models.Post
    if err := q.Paginate(req.Page, req.Size).Find(&posts); err != nil {
        return ctx.Response().Fail(http.StatusInternalServerError, "查询失败")
    }

    return ctx.Response().Paginate(posts, total, req.Page, req.Size)
}
```


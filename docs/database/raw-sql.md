# 原生 SQL

## Raw — 原生查询

当 Query Builder 无法表达复杂 SQL 时，使用 `Raw` 直接执行：

```go
// 查询到 Struct
var users []models.User
if err := facades.DB().Query().
    Raw("SELECT * FROM users WHERE status = ? AND created_at > ?", 1, cutoff).
    Scan(&users); err != nil {
    // ...
}

// 查询到自定义结构
type Stat struct {
    Date  string `json:"date"`
    Count int64  `json:"count"`
}

var stats []Stat
facades.DB().Query().
    Raw(`
        SELECT DATE(FROM_UNIXTIME(created_at/1000)) as date,
               COUNT(*) as count
        FROM users
        WHERE created_at >= ?
        GROUP BY date
        ORDER BY date DESC
    `, startTs).
    Scan(&stats)
```

---

## Exec — 执行写操作

```go
// 执行更新
if err := facades.DB().Query().
    Exec("UPDATE users SET status = ? WHERE last_login_at < ?", 0, cutoff); err != nil {
    // ...
}

// 执行删除
if err := facades.DB().Query().
    Exec("DELETE FROM logs WHERE created_at < ?", thirtyDaysAgo); err != nil {
    // ...
}

// 执行 DDL（建表、加索引等）
if err := facades.DB().Query().
    Exec(`CREATE INDEX IF NOT EXISTS idx_posts_status ON posts(status)`); err != nil {
    // ...
}
```

---

## Scan — 扫描到结构体

配合 `Raw` 将查询结果映射到任意 Struct：

```go
type UserSummary struct {
    ID         string `json:"id"`
    Name       string `json:"name"`
    PostCount  int64  `json:"post_count"`
    TotalViews int64  `json:"total_views"`
}

var summaries []UserSummary
facades.DB().Query().
    Raw(`
        SELECT u.id, u.name,
               COUNT(p.id)    AS post_count,
               SUM(p.views)   AS total_views
        FROM users u
        LEFT JOIN posts p ON p.user_id = u.id AND p.status = 1
        GROUP BY u.id, u.name
        ORDER BY post_count DESC
        LIMIT ?
    `, 10).
    Scan(&summaries)
```

---

## ScanMap — 扫描到 Map

无需预定义结构体，适合动态报表：

```go
var rows []map[string]any
if err := facades.DB().Query().
    Raw("SELECT id, name, email FROM users WHERE status = ?", 1).
    ScanMap(&rows); err != nil {
    // ...
}

for _, row := range rows {
    fmt.Printf("ID: %v, Name: %v\n", row["id"], row["name"])
}
```

---

## Row / Rows — 底层 SQL 扫描

### Row — 单行扫描

```go
var name string
var email string

row := facades.DB().Query().
    Raw("SELECT name, email FROM users WHERE id = ?", id).
    Row()

if err := row.Scan(&name, &email); err != nil {
    if errors.Is(err, sql.ErrNoRows) {
        // 未找到
    }
    // 其他错误
}
```

### Rows — 多行迭代

```go
rows, err := facades.DB().Query().
    Raw("SELECT id, name, score FROM leaderboard ORDER BY score DESC").
    Rows()
if err != nil {
    return err
}
defer rows.Close()

for rows.Next() {
    var id, name string
    var score float64
    if err := rows.Scan(&id, &name, &score); err != nil {
        return err
    }
    fmt.Printf("%s: %.2f\n", name, score)
}
```

---

## 防 SQL 注入

**永远使用占位符 `?`，不要拼接字符串：**

```go
// ✅ 正确（使用参数化查询）
facades.DB().Query().
    Raw("SELECT * FROM users WHERE name = ?", req.Name)

// ❌ 危险（SQL 注入风险）
facades.DB().Query().
    Raw("SELECT * FROM users WHERE name = '" + req.Name + "'")
```

使用占位符的其他形式：

```go
// 多个参数
facades.DB().Query().
    Raw("SELECT * FROM posts WHERE user_id = ? AND status = ?", userID, 1)

// IN 子句
ids := []string{"id1", "id2", "id3"}
facades.DB().Query().
    Raw("SELECT * FROM users WHERE id IN (?)", ids)
```

---

## 在事务中执行原生 SQL

```go
err := facades.DB().Transaction(func(tx contracts.Query) error {
    // 原生 SQL 也可以在 tx 上执行
    if err := tx.Exec("UPDATE inventory SET stock = stock - ? WHERE product_id = ?",
        quantity, productID); err != nil {
        return err
    }

    if err := tx.Exec("INSERT INTO order_items (order_id, product_id, quantity) VALUES (?, ?, ?)",
        orderID, productID, quantity); err != nil {
        return err
    }

    return nil
})
```

---

## 综合示例：数据报表接口

```go
type DailyReport struct {
    Date       string  `json:"date"`
    NewUsers   int64   `json:"new_users"`
    NewOrders  int64   `json:"new_orders"`
    Revenue    float64 `json:"revenue"`
}

func (c *ReportController) Daily(ctx contracts.Context) error {
    startDate := ctx.Query("start_date")
    endDate := ctx.Query("end_date")

    var reports []DailyReport
    err := facades.DB().Query().
        Raw(`
            SELECT
                DATE(FROM_UNIXTIME(u.created_at / 1000)) AS date,
                COUNT(DISTINCT u.id)                      AS new_users,
                COUNT(DISTINCT o.id)                      AS new_orders,
                COALESCE(SUM(o.amount), 0)                AS revenue
            FROM users u
            LEFT JOIN orders o
                ON DATE(FROM_UNIXTIME(o.created_at / 1000))
                 = DATE(FROM_UNIXTIME(u.created_at / 1000))
                AND o.status = 'paid'
            WHERE DATE(FROM_UNIXTIME(u.created_at / 1000)) BETWEEN ? AND ?
            GROUP BY date
            ORDER BY date DESC
        `, startDate, endDate).
        Scan(&reports)

    if err != nil {
        return ctx.Response().Fail(http.StatusInternalServerError, "生成报表失败")
    }
    return ctx.Response().Success(reports)
}
```


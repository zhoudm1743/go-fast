# 查询记录

## Find — 查询多条

```go
// 查询所有用户
var users []models.User
if err := facades.DB().Query().Find(&users); err != nil {
    // ...
}

// 带条件查询
var users []models.User
if err := facades.DB().Query().
    Where("status = ?", 1).
    Find(&users); err != nil {
    // ...
}

// 直接在 Find 中传条件（主键 / 字符串条件）
var users []models.User
facades.DB().Query().Find(&users, "name LIKE ?", "%Alice%")
```

---

## First — 按主键排序取第一条

```go
var user models.User

// 按主键查
if err := facades.DB().Query().First(&user, "id = ?", "some-uuid"); err != nil {
    if errors.Is(err, contracts.ErrRecordNotFound) {
        // 未找到
    }
}

// 带条件取第一条（ORDER BY primary_key ASC）
if err := facades.DB().Query().
    Where("email = ?", "alice@example.com").
    First(&user); err != nil {
    // ...
}
```

## Last — 按主键排序取最后一条

```go
var user models.User
if err := facades.DB().Query().Last(&user); err != nil {
    // ...
}
```

## Take — 不排序取一条

```go
var user models.User
if err := facades.DB().Query().Take(&user, "email = ?", "alice@example.com"); err != nil {
    // ...
}
```

---

## Where — 条件查询

### 字符串条件（推荐，防 SQL 注入）

```go
// = 等于
facades.DB().Query().Where("name = ?", "Alice")

// <> 不等于
facades.DB().Query().Where("age <> ?", 18)

// LIKE 模糊查询
facades.DB().Query().Where("name LIKE ?", "%Ali%")

// IN 集合查询
facades.DB().Query().Where("id IN ?", []string{"id1", "id2", "id3"})

// BETWEEN 范围查询
facades.DB().Query().Where("created_at BETWEEN ? AND ?", start, end)

// IS NULL
facades.DB().Query().Where("deleted_at IS NULL")

// 多条件（AND）
facades.DB().Query().Where("name = ? AND age >= ?", "Alice", 18)
```

### Struct / Map 条件

```go
// Struct（零值字段会被忽略）
facades.DB().Query().Where(&models.User{Name: "Alice", Email: "alice@example.com"})

// Map（支持零值）
facades.DB().Query().Where(map[string]any{"name": "Alice", "age": 0})
```

### OrWhere — OR 条件

```go
var users []models.User
facades.DB().Query().
    Where("name = ?", "Alice").
    OrWhere("name = ?", "Bob").
    Find(&users)
// WHERE name = 'Alice' OR name = 'Bob'
```

### Not — 排除条件

```go
var users []models.User
facades.DB().Query().
    Not("name = ?", "Alice").
    Find(&users)
// WHERE NOT name = 'Alice'

// Not with Struct
facades.DB().Query().Not(&models.User{Name: "Alice"}).Find(&users)
```

---

## Select — 选择字段

```go
// 只查询指定字段
var users []models.User
facades.DB().Query().
    Select("id", "name", "email").
    Find(&users)

// 使用 SQL 表达式
facades.DB().Query().
    Select("name, COUNT(*) as post_count").
    Joins("LEFT JOIN posts ON posts.user_id = users.id").
    Group("users.id").
    Find(&users)
```

### 查询到 Struct（DTO）

```go
type UserDTO struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

var dtos []UserDTO
facades.DB().Query().
    Model(&models.User{}).
    Select("id", "name", "email").
    Scan(&dtos)
```

---

## Order — 排序

```go
// 单字段排序
facades.DB().Query().Order("created_at DESC").Find(&users)

// 多字段排序
facades.DB().Query().
    Order("status ASC").
    Order("created_at DESC").
    Find(&users)
```

---

## Limit / Offset — 限制与偏移

```go
// 查前 10 条
facades.DB().Query().Limit(10).Find(&users)

// 跳过 20 条，取 10 条（手动分页）
facades.DB().Query().Offset(20).Limit(10).Find(&users)

// 取消 Limit（传 -1）
facades.DB().Query().Limit(10).Limit(-1).Find(&users)
```

> 推荐使用 [`Paginate`](./pagination.md) 替代手动 Offset/Limit。

---

## Distinct — 去重

```go
var emails []string
facades.DB().Query().
    Model(&models.User{}).
    Distinct("email").
    Pluck("email", &emails)
```

---

## Count — 统计数量

```go
var total int64
facades.DB().Query().Model(&models.User{}).Count(&total)

// 带条件统计
facades.DB().Query().
    Model(&models.User{}).
    Where("status = ?", 1).
    Count(&total)
```

---

## Pluck — 提取单列

```go
// 提取所有用户名
var names []string
facades.DB().Query().Model(&models.User{}).Pluck("name", &names)

// 带条件
facades.DB().Query().
    Model(&models.User{}).
    Where("status = ?", 1).
    Pluck("email", &emails)
```

---

## Scan — 扫描到任意结构

```go
type Result struct {
    Name  string
    Count int64
}

var results []Result
facades.DB().Query().
    Model(&models.User{}).
    Select("name, COUNT(*) as count").
    Group("name").
    Scan(&results)
```

---

## 链式查询（动态条件）

```go
// 构建可复用的查询基础
q := facades.DB().Query().Model(&models.User{}).Order("created_at DESC")

// 按需追加条件
if req.Name != "" {
    q = q.Where("name LIKE ?", "%"+req.Name+"%")
}
if req.Status != 0 {
    q = q.Where("status = ?", req.Status)
}

// 先统计��数
var total int64
q.Count(&total)

// 再分页查询
var users []models.User
q.Paginate(req.Page, req.Size).Find(&users)
```

> ⚠️ `Query()` 返回的实例是**不可变**的，链式调用每次返回新实例，原始 `q` 不受影响。

---

## FindInBatches — 分批处理大量数据

```go
var users []models.User

err := facades.DB().Query().
    Model(&models.User{}).
    Where("status = ?", 1).
    FindInBatches(&users, 100, func(tx contracts.Query, batch int) error {
        fmt.Printf("第 %d 批，共 %d 条\n", batch, len(users))
        for _, u := range users {
            // 处理每条记录
        }
        return nil // 返回 error 则终止
    })
```

---

## Exists — 判断是否存在

```go
exists, err := facades.DB().Query().
    Exists(&models.User{}, "email = ?", "alice@example.com")
if err != nil {
    // 处理错误
}
if exists {
    return ctx.Response().Fail(http.StatusConflict, "邮箱已存在")
}
```


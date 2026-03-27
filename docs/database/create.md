# 创建记录

## Create — 创建单条记录

```go
user := &models.User{
    Name:     "Alice",
    Email:    "alice@example.com",
    Password: "hashed_password",
}

if err := facades.DB().Query().Create(user); err != nil {
    log.Println("创建失败:", err)
    return
}

fmt.Println(user.ID)         // 自动生成的 UUID v7，如 "018f3e2a-..."
fmt.Println(user.CreatedAt)  // 自动填充的 Unix 时间戳
```

> `Create` 会调用模型的 `BeforeCreate` 钩子，**自动生成 UUID v7 主键**。

---

## 指定字段创建

使用 `Select` 只写入指定列（其余列使用数据库默认值）：

```go
if err := facades.DB().Query().
    Select("name", "email").
    Create(&user); err != nil {
    // ...
}
```

使用 `Omit` 排除不写入的列：

```go
if err := facades.DB().Query().
    Omit("password").
    Create(&user); err != nil {
    // ...
}
```

---

## 批量创建

```go
users := []models.User{
    {Name: "Alice", Email: "alice@example.com", Password: "pwd1"},
    {Name: "Bob",   Email: "bob@example.com",   Password: "pwd2"},
    {Name: "Carol", Email: "carol@example.com", Password: "pwd3"},
}

if err := facades.DB().Query().Create(&users); err != nil {
    // ...
}

for _, u := range users {
    fmt.Println(u.ID) // 每条记录都有唯一 UUID
}
```

---

## CreateInBatches — 分批创建

数据量大时分批写入，控制单次 INSERT 的行数：

```go
var users []models.User
// ... 准备 1000 条记录

// 每批 100 条
if err := facades.DB().Query().CreateInBatches(&users, 100); err != nil {
    // ...
}
```

---

## Result 变体 — 获取影响行数

```go
result := facades.DB().Query().CreateResult(&user)
if result.Error != nil {
    // 处理错误
}
fmt.Println("影响行数:", result.RowsAffected) // 1
```

---

## FirstOrCreate — 查不到则创建

查找满足条件的第一条记录，不存在则创建：

```go
user := &models.User{
    Email: "alice@example.com",
}

// 按 Email 查找，找不到则创建
if err := facades.DB().Query().FirstOrCreate(user, "email = ?", "alice@example.com"); err != nil {
    // ...
}
fmt.Println(user.ID) // 无论是已有记录还是新创建，都会填充
```

---

## FirstOrInit — 查不到则初始化（不写库）

```go
user := &models.User{}
if err := facades.DB().Query().FirstOrInit(user, "email = ?", "bob@example.com"); err != nil {
    // ...
}

if user.ID == "" {
    // 没有找到记录，user 被初始化但未写入数据库
    user.Name = "Bob"
    facades.DB().Query().Create(user)
}
```

---

## 在事务中创建

```go
err := facades.DB().Transaction(func(tx contracts.Query) error {
    user := &models.User{Name: "Alice", Email: "alice@example.com", Password: "pwd"}
    if err := tx.Create(&user); err != nil {
        return err // 自动回滚
    }

    profile := &models.Profile{UserID: user.ID, Bio: "Hello"}
    if err := tx.Create(&profile); err != nil {
        return err // 自动回滚
    }

    return nil // 自动提交
})
```

> 详见 [事务](./transactions.md)。

---

## 使用 Map 创建

```go
if err := facades.DB().Query().
    Model(&models.User{}).
    Create(map[string]any{
        "id":       uuid.Must(uuid.NewV7()).String(),
        "name":     "Dave",
        "email":    "dave@example.com",
        "password": "hashed",
    }); err != nil {
    // ...
}
```

> ⚠️ 使用 Map 创建时，`BeforeCreate` 钩子（自动生成 UUID）**不会**自动触发，需手动设置 `id`。

---

## 错误处理

```go
import (
    "errors"
    "github.com/zhoudm1743/go-fast/framework/contracts"
)

if err := facades.DB().Query().Create(&user); err != nil {
    if errors.Is(err, contracts.ErrDuplicatedKey) {
        // 唯一约束冲突（如 email 重复）
        return ctx.Response().Fail(http.StatusConflict, "邮箱已存在")
    }
    return ctx.Response().Fail(http.StatusInternalServerError, "创建失败")
}
```

> 详见 [错误处理](./error-handling.md)。


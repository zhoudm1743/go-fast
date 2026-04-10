# 钩子（Hooks）

## 钩子接口一览

GoFast 通过接口检测模型是否实现了对应钩子，各 ORM 驱动在操作前后自动调用。

方法名统一使用 `On` 前缀，与 GORM/xorm 等 ORM 框架的内置 Hook 名称（`BeforeCreate`、`AfterCreate` 等）彻底分离，避免签名冲突警告：

| 接口 | 方法 | 触发时机 |
|------|------|---------|
| `contracts.BeforeCreator` | `OnBeforeCreate(q Query) error` | 创建前 |
| `contracts.AfterCreator` | `OnAfterCreate(q Query) error` | 创建后 |
| `contracts.BeforeUpdater` | `OnBeforeUpdate(q Query) error` | 更新前 |
| `contracts.AfterUpdater` | `OnAfterUpdate(q Query) error` | 更新后 |
| `contracts.BeforeDeleter` | `OnBeforeDelete(q Query) error` | 删除前 |
| `contracts.AfterDeleter` | `OnAfterDelete(q Query) error` | 删除后 |
| `contracts.AfterFinder` | `OnAfterFind(q Query) error` | 查询后 |

---

## 内置钩子：AutoGenerateID 自动生成 UUID

`database.Model` 实现了 `contracts.IDAutoGenerator` 接口，驱动层在 Create 前自动调用，生成 UUID v7 主键：

```go
// framework/database/model.go（框架内部）
func (m *Model) AutoGenerateID() {
    if m.ID == "" {
        m.ID = uuid.Must(uuid.NewV7()).String()
    }
}
```

嵌入 `database.Model` 的所有模型都自动拥有此行为，**无需任何配置**。

---

## 自定义 OnBeforeCreate

```go
type User struct {
    database.Model
    Name     string `gorm:"size:100" json:"name"`
    Email    string `gorm:"size:200;uniqueIndex" json:"email"`
    Password string `gorm:"size:255" json:"-"`
    Slug     string `gorm:"size:200;uniqueIndex" json:"slug"`
}

// OnBeforeCreate 在创建前自动生成 Slug
// 注意：UUID 由框架自动生成，无需手动调用
func (u *User) OnBeforeCreate(_ contracts.Query) error {
    if u.Slug == "" {
        u.Slug = generateSlug(u.Name)
    }
    return nil
}

func generateSlug(name string) string {
    return strings.ToLower(strings.ReplaceAll(name, " ", "-"))
}
```

---

## OnAfterCreate — 创建后发送通知

```go
type Order struct {
    database.Model
    UserID string  `gorm:"size:36;index" json:"user_id"`
    Amount float64 `gorm:"not null"      json:"amount"`
    Status string  `gorm:"size:20"       json:"status"`
}

// OnAfterCreate 订单创建后发送通知
func (o *Order) OnAfterCreate(_ contracts.Query) error {
    go notifyUser(o.UserID, fmt.Sprintf("您的订单 %s 已创建，金额 %.2f", o.ID, o.Amount))
    return nil
}
```

---

## OnBeforeUpdate — 更新前验证

```go
func (u *User) OnBeforeUpdate(_ contracts.Query) error {
    if u.Email == "" {
        return errors.New("email 不能为空")
    }
    if u.Password != "" && !isHashed(u.Password) {
        hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
        if err != nil {
            return err
        }
        u.Password = string(hash)
    }
    return nil
}
```

---

## OnBeforeDelete — 删除前检查

```go
func (u *User) OnBeforeDelete(q contracts.Query) error {
    var count int64
    if err := q.Model(&Order{}).
        Where("user_id = ? AND status = ?", u.ID, "pending").
        Count(&count); err != nil {
        return err
    }
    if count > 0 {
        return fmt.Errorf("用户有 %d 笔未完成订单，无法删除", count)
    }
    return nil
}
```

---

## OnAfterFind — 查询后处理

```go
type Product struct {
    database.Model
    Name      string  `gorm:"size:200" json:"name"`
    Price     int64   `json:"price"`
    PriceYuan float64 `gorm:"-" json:"price_yuan"`
}

// OnAfterFind 查询后自动计算展示字段
func (p *Product) OnAfterFind(_ contracts.Query) error {
    p.PriceYuan = float64(p.Price) / 100.0
    return nil
}
```

---

## 钩子中的错误处理

钩子返回 `error` 时，会阻止操作并将错误传递给调用方：

```go
func (u *User) OnBeforeCreate(q contracts.Query) error {
    var count int64
    q.Model(&User{}).Where("email = ?", u.Email).Count(&count)
    if count > 0 {
        return contracts.ErrDuplicatedKey
    }
    return nil
}

// 调用处捕获钩子错误
if err := facades.DB().Query().Create(&user); err != nil {
    if errors.Is(err, contracts.ErrDuplicatedKey) {
        return ctx.Response().Fail(http.StatusConflict, "邮箱已存在")
    }
    return ctx.Response().Fail(http.StatusInternalServerError, "创建失败")
}
```

---

## 注意事项

| 规则 | 说明 |
|------|------|
| 钩子在事务中执行 | `Create/Update/Delete` 在事务中时，钩子也在同一事务内 |
| 避免重型操作 | 钩子不应做耗时操作（如发送 HTTP 请求），改用 `goroutine` 异步处理 |
| AfterFind 限制 | 避免在 `OnAfterFind` 中再次查询数据库，会引发 N+1 问题 |
| UUID 自动生成 | 框架自动调用 `AutoGenerateID()`，自定义 `OnBeforeCreate` 无需手动处理 UUID |
| 钩子不影响 Raw/Exec | 原生 SQL 不触发任何模型钩子 |


---

## 内置钩子：BeforeCreate 自动生成 UUID

`database.Model` 已内置 `BeforeCreate`，会在创建时自动生成 UUID v7 主键：

```go
// framework/database/model.go（框架内部）
func (m *Model) BeforeCreate(_ contracts.Query) error {
    if m.ID == "" {
        m.ID = uuid.Must(uuid.NewV7()).String()
    }
    return nil
}
```

嵌入 `database.Model` 的所有模型都自动拥有此行为，**无需任何配置**。

---

## 自定义 BeforeCreate

```go
type User struct {
    database.Model
    Name     string `gorm:"size:100" json:"name"`
    Email    string `gorm:"size:200;uniqueIndex" json:"email"`
    Password string `gorm:"size:255" json:"-"`
    Slug     string `gorm:"size:200;uniqueIndex" json:"slug"`
}

// BeforeCreate 在创建前自动生成 Slug
func (u *User) BeforeCreate(q contracts.Query) error {
    // 先调用父类（生成 UUID）
    if err := u.Model.BeforeCreate(q); err != nil {
        return err
    }
    // 生成唯一 Slug
    if u.Slug == "" {
        u.Slug = generateSlug(u.Name)
    }
    return nil
}

func generateSlug(name string) string {
    // 简单示例
    return strings.ToLower(strings.ReplaceAll(name, " ", "-"))
}
```

---

## AfterCreate — 创建后发送通知

```go
type Order struct {
    database.Model
    UserID string  `gorm:"size:36;index" json:"user_id"`
    Amount float64 `gorm:"not null"      json:"amount"`
    Status string  `gorm:"size:20"       json:"status"`
}

// AfterCreate 订单创建后发送通知
func (o *Order) AfterCreate(_ contracts.Query) error {
    // 发送站内消息、邮件等（注意：此处不能访问数据库，会产生新事务）
    go notifyUser(o.UserID, fmt.Sprintf("您的订单 %s 已创建，金额 %.2f", o.ID, o.Amount))
    return nil
}
```

---

## BeforeUpdate — 更新前验证

```go
func (u *User) BeforeUpdate(_ contracts.Query) error {
    if u.Email == "" {
        return errors.New("email 不能为空")
    }
    // 密码修改时自动加密
    if u.Password != "" && !isHashed(u.Password) {
        hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
        if err != nil {
            return err
        }
        u.Password = string(hash)
    }
    return nil
}
```

---

## BeforeDelete — 删除前检查

```go
func (u *User) BeforeDelete(q contracts.Query) error {
    // 检查是否有未完成订单
    var count int64
    if err := q.Model(&Order{}).
        Where("user_id = ? AND status = ?", u.ID, "pending").
        Count(&count); err != nil {
        return err
    }
    if count > 0 {
        return fmt.Errorf("用户有 %d 笔未完成订单，无法删除", count)
    }
    return nil
}
```

---

## AfterFind — 查询后处理

```go
type Product struct {
    database.Model
    Name      string  `gorm:"size:200" json:"name"`
    Price     int64   `json:"price"`        // 以分存储
    PriceYuan float64 `gorm:"-" json:"price_yuan"` // 忽略，不存库
}

// AfterFind 查询后自动计算展示字段
func (p *Product) AfterFind(_ contracts.Query) error {
    p.PriceYuan = float64(p.Price) / 100.0
    return nil
}
```

---

## 钩子中的错误处理

钩子返回 `error` 时，会阻止操作并将错误传递给调用方：

```go
func (u *User) BeforeCreate(q contracts.Query) error {
    // 检查邮箱是否已存在
    var count int64
    q.Model(&User{}).Where("email = ?", u.Email).Count(&count)
    if count > 0 {
        return contracts.ErrDuplicatedKey // 返回标准 Sentinel 错误
    }
    return u.Model.BeforeCreate(q)
}

// 调用处捕获钩子错误
if err := facades.DB().Query().Create(&user); err != nil {
    if errors.Is(err, contracts.ErrDuplicatedKey) {
        return ctx.Response().Fail(http.StatusConflict, "邮箱已存在")
    }
    return ctx.Response().Fail(http.StatusInternalServerError, "创建失败")
}
```

---

## 注意事项

| 规则 | 说明 |
|------|------|
| 钩子在事务中执行 | `Create/Update/Delete` 在事务中时，钩子也在同一事务内 |
| 避免重型操作 | 钩子不应做耗时操作（如发送 HTTP 请求），改用 `goroutine` 异步处理 |
| AfterFind 限制 | 避免在 `AfterFind` 中再次查询数据库，会引发 N+1 问题 |
| 调用父类钩子 | 自定义 `BeforeCreate` 时应先调用 `u.Model.BeforeCreate(q)` 确保 UUID 生成 |
| 钩子不影响 Raw/Exec | 原生 SQL 不触发任何模型钩子 |


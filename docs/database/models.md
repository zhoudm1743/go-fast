# 模型声明

## 基础模型

所有业务模型推荐嵌入 `database.Model`，自动获得：

- **UUID v7** 字符串主键（创建时自动生成）
- `created_at` 整型时间戳（Unix 毫秒，`autoCreateTime`）
- `updated_at` 整型时间戳（Unix 毫秒，`autoUpdateTime`）

```go
package models

import "github.com/zhoudm1743/go-fast/framework/database"

type User struct {
    database.Model                                          // ID + CreatedAt + UpdatedAt
    Name     string `gorm:"size:100;not null"  json:"name"`
    Email    string `gorm:"size:200;uniqueIndex;not null" json:"email"`
    Password string `gorm:"size:255;not null"  json:"-"`   // json:"-" 不输出
}
```

`database.Model` 的定义：

```go
type Model struct {
    ID        string `gorm:"primaryKey;size:36;column:id"       json:"id"`
    CreatedAt int64  `gorm:"autoCreateTime;column:created_at"   json:"created_at"`
    UpdatedAt int64  `gorm:"autoUpdateTime;column:updated_at"   json:"updated_at"`
}
```

---

## 软删除模型

嵌入 `database.ModelWithSoftDelete`，`deleted_at` 为 `int64` 整型时间戳（`0` 表示未删除）：

```go
type Article struct {
    database.ModelWithSoftDelete   // ID + CreatedAt + UpdatedAt + DeletedAt
    Title   string `gorm:"size:200;not null" json:"title"`
    Content string `gorm:"type:text"         json:"content"`
}
```

软删除行为：
- `Delete()` → 设置 `deleted_at = unix_now()`（软删除）
- `Find/First` → 自动过滤 `deleted_at != 0`
- `Unscoped().Find()` → 查询所有记录（含已删除）
- `OnlyTrashed()` → 只查询已删除记录
- `Restore()` → 恢复软删除记录（`deleted_at = 0`）
- `ForceDelete()` → 物理删除

---

## 自定义模型（不嵌入 Model）

```go
type Log struct {
    ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    Level     string    `gorm:"size:20"                 json:"level"`
    Message   string    `gorm:"type:text"               json:"message"`
    CreatedAt time.Time `gorm:"autoCreateTime"          json:"created_at"`
}
```

---

## 常用字段标签

> 以 GORM 驱动为例，标签写在 `gorm:"..."` 中。

| 标签 | 说明 | 示例 |
|------|------|------|
| `primaryKey` | 主键 | `gorm:"primaryKey"` |
| `autoIncrement` | 自增 | `gorm:"autoIncrement"` |
| `column:xxx` | 列名 | `gorm:"column:user_name"` |
| `size:n` | 字符串长度 | `gorm:"size:200"` |
| `type:xxx` | 字段类型 | `gorm:"type:text"` |
| `not null` | 非空 | `gorm:"not null"` |
| `default:v` | 默认值 | `gorm:"default:0"` |
| `unique` | 唯一约束 | `gorm:"unique"` |
| `uniqueIndex` | 唯一索引 | `gorm:"uniqueIndex"` |
| `index` | 普通索引 | `gorm:"index"` |
| `index:idx_name` | 命名索引 | `gorm:"index:idx_email"` |
| `autoCreateTime` | 创建时自动赋时间 | `gorm:"autoCreateTime"` |
| `autoUpdateTime` | 更新时自动赋时间 | `gorm:"autoUpdateTime"` |
| `-` | 忽略该字段 | `gorm:"-"` |
| `->` | 只读字段 | `gorm:"->"` |
| `<-` | 只写字段 | `gorm:"<-"` |

### 复合唯一索引

```go
type UserRole struct {
    database.Model
    UserID string `gorm:"size:36;uniqueIndex:idx_user_role" json:"user_id"`
    RoleID string `gorm:"size:36;uniqueIndex:idx_user_role" json:"role_id"`
}
```

### 外键关联

```go
type Post struct {
    database.Model
    UserID string `gorm:"size:36;index;not null" json:"user_id"`
    User   User   `gorm:"foreignKey:UserID"      json:"user,omitempty"`
    Title  string `gorm:"size:200;not null"      json:"title"`
}
```

---

## 表名约定

GORM 默认将结构体名转为蛇形复数作为表名：

| 结构体 | 表名 |
|--------|------|
| `User` | `users` |
| `Post` | `posts` |
| `UserRole` | `user_roles` |

自定义表名，实现 `TableName()` 方法：

```go
func (User) TableName() string { return "sys_user" }
```

---

## 完整模型示例

```go
package models

import "github.com/zhoudm1743/go-fast/framework/database"

// Category 文章分类
type Category struct {
    database.Model
    Name     string `gorm:"size:100;not null;uniqueIndex" json:"name"`
    Slug     string `gorm:"size:100;not null;uniqueIndex" json:"slug"`
    ParentID string `gorm:"size:36;index;default:''"     json:"parent_id"`
    Sort     int    `gorm:"default:0"                   json:"sort"`
}

// Tag 标签
type Tag struct {
    database.Model
    Name string `gorm:"size:50;not null;uniqueIndex" json:"name"`
}

// Post 文章（带软删除）
type Post struct {
    database.ModelWithSoftDelete
    Title      string     `gorm:"size:200;not null"           json:"title"`
    Slug       string     `gorm:"size:200;not null;uniqueIndex" json:"slug"`
    Content    string     `gorm:"type:longtext"               json:"content"`
    Summary    string     `gorm:"size:500"                    json:"summary"`
    Cover      string     `gorm:"size:500"                    json:"cover"`
    Status     int        `gorm:"default:0;index"             json:"status"` // 0=草稿 1=发布
    Views      int64      `gorm:"default:0"                   json:"views"`
    UserID     string     `gorm:"size:36;index;not null"      json:"user_id"`
    CategoryID string     `gorm:"size:36;index"               json:"category_id"`
    // 关联（不存入数据库）
    Author     *User      `gorm:"foreignKey:UserID"           json:"author,omitempty"`
    Category   *Category  `gorm:"foreignKey:CategoryID"       json:"category,omitempty"`
    Tags       []Tag      `gorm:"many2many:post_tags;"        json:"tags,omitempty"`
}
```

---

## 自动迁移

```go
func (p *DatabaseProvider) MigrateDB(db contracts.DB) error {
    return db.AutoMigrate(
        &models.User{},
        &models.Category{},
        &models.Tag{},
        &models.Post{},
    )
}
```

> 自动迁移只增加字段/索引，**不会删除或修改**已有列，生产环境安全。


# 快速入门

## 简介

GoFast 数据库模块提供一套统一的查询接口。业务代码只依赖 `contracts.Query`，底层驱动（GORM / xorm / torm）可随时切换，**零业务改动**。

```
facades.DB().Query()          ← 统一入口
      │
      ▼
contracts.Query               ← 接口层（无任何 ORM 依赖）
      │
      ▼
GormDriver → *gorm.DB         ← 当前默认实现
```

## 第一步：配置数据库

编辑 `config/config.yaml`：

```yaml
database:
  driver: sqlite                  # 数据库引擎
  database: database/gofast.db    # 数据库文件路径
```

> 完整配置项见 [连接数据库](./connecting.md)。

## 第二步：声明模型

```go
// app/models/post.go
package models

import "github.com/zhoudm1743/go-fast/framework/database"

type Post struct {
    database.Model                // 嵌入：UUID 主键 + CreatedAt + UpdatedAt
    Title   string `gorm:"size:200;not null" json:"title"`
    Content string `gorm:"type:text"         json:"content"`
    UserID  string `gorm:"size:36;index"     json:"user_id"`
}
```

## 第三步：自动迁移

在 `ServiceProvider` 中实现 `DBMigrator` 接口：

```go
// app/providers/database_provider.go
package providers

import (
    "github.com/zhoudm1743/go-fast/app/models"
    "github.com/zhoudm1743/go-fast/framework/contracts"
    "github.com/zhoudm1743/go-fast/framework/foundation"
)

type DatabaseProvider struct{}

func (p *DatabaseProvider) Register(app foundation.Application) {}
func (p *DatabaseProvider) Boot(app foundation.Application) error { return nil }

// MigrateDB 框架启动后自动调用
func (p *DatabaseProvider) MigrateDB(db contracts.DB) error {
    return db.AutoMigrate(&models.Post{}, &models.User{})
}
```

在 `bootstrap/app.go` 注册：

```go
app.SetProviders([]foundation.ServiceProvider{
    // ...其他 Provider
    &providers.DatabaseProvider{},
})
```

## 第四步：CRUD 操作

```go
import "github.com/zhoudm1743/go-fast/framework/facades"

// ── 创建 ──────────────────────────────────────────────────────
post := &models.Post{Title: "Hello GoFast", Content: "内容..."}
if err := facades.DB().Query().Create(post); err != nil {
    // 处理错误
}
fmt.Println(post.ID) // 自动生成的 UUID v7

// ── 查询 ──────────────────────────────────────────────────────
var post models.Post
err := facades.DB().Query().First(&post, "id = ?", "some-uuid")

// ── 列表查询 ──────────────────────────────────────────────────
var posts []models.Post
err := facades.DB().Query().
    Where("user_id = ?", userID).
    Order("created_at DESC").
    Find(&posts)

// ── 更新 ──────────────────────────────────────────────────────
err := facades.DB().Query().
    Model(&post).
    Updates(map[string]any{"title": "新标题"})

// ── 删除 ──────────────────────────────────────────────────────
err := facades.DB().Query().Delete(&post)
```

## 在控制器中使用

```go
package controllers

import (
    "errors"
    "net/http"

    "github.com/zhoudm1743/go-fast/app/models"
    "github.com/zhoudm1743/go-fast/framework/contracts"
    "github.com/zhoudm1743/go-fast/framework/facades"
)

type PostController struct{}

func (c *PostController) Index(ctx contracts.Context) error {
    var posts []models.Post
    if err := facades.DB().Query().
        Order("created_at DESC").
        Find(&posts); err != nil {
        return ctx.Response().Fail(http.StatusInternalServerError, "查询失败")
    }
    return ctx.Response().Success(posts)
}

func (c *PostController) Show(ctx contracts.Context) error {
    id := ctx.Param("id")
    var post models.Post
    if err := facades.DB().Query().First(&post, "id = ?", id); err != nil {
        if errors.Is(err, contracts.ErrRecordNotFound) {
            return ctx.Response().NotFound("文章不存在")
        }
        return ctx.Response().Fail(http.StatusInternalServerError, "查询失败")
    }
    return ctx.Response().Success(post)
}
```

## 与旧 `facades.Orm()` 对比

| 旧写法 | 新写法 |
|--------|--------|
| `facades.Orm().DB().Find(&list)` | `facades.DB().Query().Find(&list)` |
| `...DB().Where(...).Find(&list)` | `...Query().Where(...).Find(&list)` |
| `...DB().First(&u).Error` | `...Query().First(&u)` |
| `...DB().Create(&u).Error` | `...Query().Create(&u)` |
| `.Offset((p-1)*s).Limit(s)` | `.Paginate(p, s)` |

> `facades.Orm()` 已标记为 **Deprecated**，将在下一主版本移除。

## 下一步

- [连接数据库](./connecting.md) — 配置 MySQL / PostgreSQL / SQLite
- [模型声明](./models.md) — 字段标签、索引、软删除
- [查询记录](./query.md) — 条件、排序、关联预加载


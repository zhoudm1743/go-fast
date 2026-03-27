# 作用域（Scopes）

## 什么是 Scopes

Scopes 允许将常用查询条件封装为**可复用函数**，在多处查询中组合使用，避免代码重复。

函数签名：`func(contracts.Query) contracts.Query`

---

## 基础用法

```go
// 定义 Scope
func Published(q contracts.Query) contracts.Query {
    return q.Where("status = ?", 1)
}

func RecentFirst(q contracts.Query) contracts.Query {
    return q.Order("created_at DESC")
}

func WithAuthor(q contracts.Query) contracts.Query {
    return q.Preload("Author")
}

// 使用 Scopes 组合
var posts []models.Post
facades.DB().Query().
    Model(&models.Post{}).
    Scopes(Published, RecentFirst, WithAuthor).
    Find(&posts)
```

---

## 带参数的 Scope（返回函数的函数）

```go
// 按分类过滤
func InCategory(categoryID string) func(contracts.Query) contracts.Query {
    return func(q contracts.Query) contracts.Query {
        if categoryID == "" {
            return q
        }
        return q.Where("category_id = ?", categoryID)
    }
}

// 关键字搜索
func SearchKeyword(keyword string) func(contracts.Query) contracts.Query {
    return func(q contracts.Query) contracts.Query {
        if keyword == "" {
            return q
        }
        return q.Where("title LIKE ? OR content LIKE ?",
            "%"+keyword+"%", "%"+keyword+"%")
    }
}

// 日期范围
func DateRange(start, end int64) func(contracts.Query) contracts.Query {
    return func(q contracts.Query) contracts.Query {
        if start > 0 {
            q = q.Where("created_at >= ?", start)
        }
        if end > 0 {
            q = q.Where("created_at <= ?", end)
        }
        return q
    }
}

// 使用
var posts []models.Post
facades.DB().Query().
    Model(&models.Post{}).
    Scopes(
        Published,
        InCategory(req.CategoryID),
        SearchKeyword(req.Keyword),
        DateRange(req.StartAt, req.EndAt),
        RecentFirst,
    ).
    Find(&posts)
```

---

## 在文件中组织 Scopes

建议将业务 Scope 集中存放：

```go
// app/scopes/post_scopes.go
package scopes

import "github.com/zhoudm1743/go-fast/framework/contracts"

// Published 只查已发布的文章
func Published(q contracts.Query) contracts.Query {
    return q.Where("status = 1")
}

// Draft 只查草稿
func Draft(q contracts.Query) contracts.Query {
    return q.Where("status = 0")
}

// ByUser 按用户过滤
func ByUser(userID string) func(contracts.Query) contracts.Query {
    return func(q contracts.Query) contracts.Query {
        return q.Where("user_id = ?", userID)
    }
}

// Paginated 分页
func Paginated(page, size int) func(contracts.Query) contracts.Query {
    return func(q contracts.Query) contracts.Query {
        return q.Paginate(page, size)
    }
}
```

使用：

```go
import "github.com/zhoudm1743/go-fast/app/scopes"

var posts []models.Post
facades.DB().Query().
    Model(&models.Post{}).
    Scopes(
        scopes.Published,
        scopes.ByUser(userID),
        scopes.Paginated(req.Page, req.Size),
    ).
    Find(&posts)
```

---

## 与 Count 配合

Scope 函数可在 Count 和 Find 中**复用同一套条件**：

```go
func buildPostQuery(req ListPostRequest) []func(contracts.Query) contracts.Query {
    fns := []func(contracts.Query) contracts.Query{scopes.Published}
    if req.CategoryID != "" {
        fns = append(fns, scopes.InCategory(req.CategoryID))
    }
    if req.Keyword != "" {
        fns = append(fns, scopes.SearchKeyword(req.Keyword))
    }
    return fns
}

// 控制器中
conditions := buildPostQuery(req)

var total int64
facades.DB().Query().Model(&models.Post{}).Scopes(conditions...).Count(&total)

var posts []models.Post
facades.DB().Query().
    Model(&models.Post{}).
    Scopes(conditions...).
    Scopes(scopes.Paginated(req.Page, req.Size)).
    Preload("Author").
    Order("created_at DESC").
    Find(&posts)
```

---

## 全局 Scope（软删除过滤示例）

如需对某个模型自动应用条件，在每次查询时统一使用 Scope：

```go
// NotDeleted 排除软删除记录
func NotDeleted(q contracts.Query) contracts.Query {
    return q.Where("deleted_at = 0")
}

// 适用于未嵌入 ModelWithSoftDelete 但有 deleted_at 字段的场景
facades.DB().Query().
    Table("audit_logs").
    Scopes(NotDeleted).
    Find(&logs)
```


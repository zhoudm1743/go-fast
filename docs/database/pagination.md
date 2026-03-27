# 分页

## Paginate — 快捷分页

```go
q := facades.DB().Query().
    Model(&models.User{}).
    Order("created_at DESC")

// 先统计总数
var total int64
q.Count(&total)

// 再分页查询（第 1 页，每页 20 条）
var users []models.User
if err := q.Paginate(1, 20).Find(&users); err != nil {
    // ...
}
```

`Paginate(page, size)` 等价于 `.Offset((page-1)*size).Limit(size)`，同时处理边界：
- `page < 1` → 自动修正为 `1`
- `size < 1` → 自动修正为 `20`

---

## 控制器完整示例

```go
type ListRequest struct {
    Page  int    `query:"page"  binding:"omitempty,gte=1"`
    Size  int    `query:"size"  binding:"omitempty,gte=1,lte=100"`
    Email string `query:"email" binding:"omitempty,email"`
}

func (c *UserController) Index(ctx contracts.Context) error {
    var req ListRequest
    if err := ctx.Bind(&req); err != nil {
        return ctx.Response().Validation(err)
    }
    if req.Page == 0 { req.Page = 1 }
    if req.Size == 0 { req.Size = 20 }

    q := facades.DB().Query().
        Model(&models.User{}).
        Order("created_at DESC")

    if req.Email != "" {
        q = q.Where("email LIKE ?", "%"+req.Email+"%")
    }

    var total int64
    q.Count(&total)

    var users []models.User
    if err := q.Paginate(req.Page, req.Size).Find(&users); err != nil {
        facades.Log().Errorf("list users: %v", err)
        return ctx.Response().Fail(http.StatusInternalServerError, "查询失败")
    }

    return ctx.Response().Paginate(users, total, req.Page, req.Size)
}
```

---

## 响应结构

`ctx.Response().Paginate()` 自动生成标准分页响应：

```json
{
    "code": 0,
    "message": "ok",
    "data": {
        "items": [...],
        "total": 100,
        "page": 1,
        "size": 20,
        "total_pages": 5
    }
}
```

---

## 手动分页（Offset + Limit）

```go
page := 2
size := 10
offset := (page - 1) * size   // 10

var users []models.User
facades.DB().Query().
    Offset(offset).
    Limit(size).
    Find(&users)
```

---

## 游标分页（Cursor-based，大数据量推荐）

传统 `OFFSET` 分页在数据量大时性能差（数据库需扫描前 N 行）。
游标分页通过记录上次最后一条的 ID 进行范围查询：

```go
type CursorRequest struct {
    Cursor string `query:"cursor"` // 上一页最后一条的 created_at
    Size   int    `query:"size"`
}

func (c *PostController) List(ctx contracts.Context) error {
    var req CursorRequest
    ctx.Bind(&req)
    if req.Size == 0 { req.Size = 20 }

    q := facades.DB().Query().
        Model(&models.Post{}).
        Where("status = ?", 1).
        Order("created_at DESC").
        Limit(req.Size + 1) // 多取一条，判断是否有下一页

    if req.Cursor != "" {
        q = q.Where("created_at < ?", req.Cursor)
    }

    var posts []models.Post
    if err := q.Find(&posts); err != nil {
        return ctx.Response().Fail(http.StatusInternalServerError, "查询失败")
    }

    hasMore := len(posts) > req.Size
    if hasMore {
        posts = posts[:req.Size]
    }

    var nextCursor string
    if hasMore && len(posts) > 0 {
        nextCursor = fmt.Sprintf("%d", posts[len(posts)-1].CreatedAt)
    }

    return ctx.Response().Success(map[string]any{
        "items":       posts,
        "has_more":    hasMore,
        "next_cursor": nextCursor,
    })
}
```

---

## 多查询复用（避免重复 Count）

将查询条件封装成 Scope，分别用于 Count 和 Find：

```go
func postQuery(status int, keyword string) func(contracts.Query) contracts.Query {
    return func(q contracts.Query) contracts.Query {
        if status != 0 {
            q = q.Where("status = ?", status)
        }
        if keyword != "" {
            q = q.Where("title LIKE ?", "%"+keyword+"%")
        }
        return q
    }
}

// 使用
scope := postQuery(req.Status, req.Keyword)

var total int64
facades.DB().Query().
    Model(&models.Post{}).
    Scopes(scope).
    Count(&total)

var posts []models.Post
facades.DB().Query().
    Model(&models.Post{}).
    Scopes(scope).
    Preload("Author").
    Order("created_at DESC").
    Paginate(req.Page, req.Size).
    Find(&posts)
```

> 详见 [作用域](./scopes.md)。


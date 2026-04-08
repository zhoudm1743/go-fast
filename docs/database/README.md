# GoFast 数据库

<p align="center">
  <strong>与 ORM 无关 · 开箱即用 · 可插拔驱动 · 多连接管理 · PostgreSQL Schema</strong>
</p>

<p align="center">
  <a href="./getting-started.md">快速入门</a> ·
  <a href="./connecting.md">连接数据库</a> ·
  <a href="./query.md">查询</a> ·
  <a href="./transactions.md">事务</a> ·
  <a href="https://github.com/zhoudm1743/go-fast">GitHub</a>
</p>

---

## 特性

- 🔌 **可插拔驱动** — 内置 GORM，可无缝切换 xorm / torm，业务代码零改动
- 🔗 **多连接管理** — 同时管理主库、只读副本、多租户等多个命名连接，懒加载初始化
- 🗂️ **PostgreSQL Schema** — 连接级 `schema` 配置 + 链式 `.Schema(name)` 动态切换，多租户 Schema 隔离开箱即用
- ⛓️ **流畅链式 API** — `Where / Order / Paginate / Scopes` 等方法链式组合，直观易读
- 📄 **分页内置** — `Paginate(page, size)` 一行搞定，自动计算 OFFSET
- 🔄 **事务完善** — 自动事务、手动事务、SavePoint、隔离级别全支持
- 🛡️ **标准错误** — `ErrRecordNotFound / ErrDuplicatedKey` 等 Sentinel，`errors.Is` 精确判断
- 🪝 **生命周期钩子** — `BeforeCreate / AfterCreate` 等接口，UUID 自动生成内置
- 🔍 **软删除** — `ModelWithSoftDelete` 开箱即用，支持 Restore / OnlyTrashed / ForceDelete
- 🧩 **作用域复用** — `Scopes(...)` 将常用条件封装为函数，DRY 原则
- 🐛 **调试友好** — `.Debug()` 一键输出完整 SQL，慢查询自动告警

---

## 快速开始

### 安装（已内置，无需额外安装）

GoFast 数据库模块开箱即用，只需在 `config/config.yaml` 中配置连接信息即可。

### 配置

```yaml
# config/config.yaml
database:
  driver: sqlite
  database: database/app.db
```

> 支持 MySQL / PostgreSQL / SQLite / SQL Server，详见 [连接数据库](./connecting.md)。

### 声明模型

```go
// app/models/user.go
package models

import "github.com/zhoudm1743/go-fast/framework/database"

type User struct {
    database.Model                                          // UUID 主键 + 时间戳
    Name     string `gorm:"size:100;not null"  json:"name"`
    Email    string `gorm:"uniqueIndex;not null" json:"email"`
    Password string `gorm:"size:255"           json:"-"`
}
```

### 自动迁移

```go
func (p *AppProvider) MigrateDB(db contracts.DB) error {
    return db.AutoMigrate(&models.User{})
}
```

### CRUD

```go
// 创建
user := &models.User{Name: "Alice", Email: "alice@example.com"}
facades.DB().Query().Create(user)
fmt.Println(user.ID) // 018f3e2a-... 自动生成 UUID v7

// 查询
var user models.User
facades.DB().Query().First(&user, "id = ?", id)

// 分页列表
var users []models.User
var total int64
q := facades.DB().Query().Model(&models.User{}).Order("created_at DESC")
q.Count(&total)
q.Paginate(1, 20).Find(&users)

// 更新
facades.DB().Query().Model(&user).Updates(map[string]any{"name": "Bob"})

// 删除（软删除）
facades.DB().Query().Delete(&user)
```

---

## 文档导航

### 🚀 入门

| 文档 | 说明 |
|------|------|
| [快速入门](./getting-started.md) | 5 分钟完成配置、模型声明和第一个 CRUD |
| [连接数据库](./connecting.md) | MySQL / PostgreSQL / SQLite / SQL Server 配置详解、连接池调优 |
| [模型声明](./models.md) | Model 结构体、字段标签、软删除、自动迁移 |

### 📝 增删改查

| 文档 | 说明 |
|------|------|
| [创建记录](./create.md) | Create、批量插入、CreateInBatches、FirstOrCreate |
| [查询记录](./query.md) | Find、First、Where、Order、Select、Pluck、链式动态条件 |
| [更新记录](./update.md) | Save、Update、Updates、Result 变体、影响行数 |
| [删除记录](./delete.md) | Delete、软删除、Restore、ForceDelete、批量删除 |

### 🔎 高级查询

| 文档 | 说明 |
|------|------|
| [高级查询](./advanced-query.md) | Join、Preload、Group、Having、子查询、悲观锁、ScanMap |
| [分页](./pagination.md) | Paginate、游标分页、Count + Find 复用 |
| [作用域](./scopes.md) | Scopes 封装复用查询条件、与 Count 配合、全局过滤 |

### ⚙️ 核心机制

| 文档 | 说明 |
|------|------|
| [事务](./transactions.md) | 自动事务、手动事务、SavePoint、隔离级别、Service 分层 |
| [原生 SQL](./raw-sql.md) | Raw、Exec、Scan、ScanMap、Row/Rows、防 SQL 注入 |
| [钩子](./hooks.md) | BeforeCreate / AfterCreate 等生命周期钩子，UUID 自动生成原理 |
| [错误处理](./error-handling.md) | Sentinel Errors、errors.Is、Result 结构体、封装辅助函数 |

### 🏗️ 架构与扩展

| 文档 | 说明 |
|------|------|
| [多数据库连接](./multi-connection.md) | 读写分离、多租户、自定义驱动注册、健康检查 |

---

## 架构设计

```
facades.DB()                    ← 统一入口（Facade）
    │
    ▼
contracts.DB                    ← 管理器接口（多连接）
    │  ├── Query(ctx)           ← 默认连接查询构建器
    │  ├── Connection("name")   ← 切换命名连接
    │  └── Driver("name")       ← 获取底层驱动
    │
    ▼
contracts.Driver                ← 驱动适配器接口
    │  ├── GormDriver           ← 内置 GORM 实现
    │  ├── XormDriver           ← 可插拔（独立插件）
    │  └── TormDriver           ← 可插拔（独立插件）
    │
    ▼
contracts.Query                 ← 查询构建器接口（链式 API）
    └── GormQuery               ← GORM 实现（代理 *gorm.DB）
```

---

## 支持的功能矩阵

| 功能 | GORM 驱动 | xorm 驱动 | torm 驱动 |
|------|:---------:|:---------:|:---------:|
| CRUD | ✅ | 🔜 | 🔜 |
| 事务 | ✅ | 🔜 | 🔜 |
| 软删除 | ✅ | 🔜 | 🔜 |
| 钩子 | ✅ | 🔜 | 🔜 |
| 悲观锁 | ✅ | 🔜 | 🔜 |
| 分页 | ✅ | 🔜 | 🔜 |
| 原生 SQL | ✅ | 🔜 | 🔜 |
| 作用域 | ✅ | 🔜 | 🔜 |
| 多连接 | ✅ | 🔜 | 🔜 |
| AutoMigrate | ✅ | 🔜 | 🔜 |

> ✅ 已支持 · 🔜 规划中

---

## 贡献

欢迎提交 Issue 和 Pull Request：

- 📦 [GoFast GitHub](https://github.com/zhoudm1743/go-fast)
- 🐛 [反馈问题](https://github.com/zhoudm1743/go-fast/issues)
- 📖 [改进文档](https://github.com/zhoudm1743/go-fast/tree/main/docs/database)

---

## 许可证

GoFast 基于 [MIT License](../../LICENSE) 开源。

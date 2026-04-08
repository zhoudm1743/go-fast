# 多数据库连接

## 配置多连接

在 `config/config.yaml` 中声明多个命名连接：

```yaml
database:
  default: main              # 默认连接

  connections:
    main:
      driver: gormdriver
      engine: mysql
      host: 10.0.0.1
      port: 3306
      username: writer
      password: secret
      database: myapp
      max_open_conns: 100

    read_replica:
      driver: gormdriver
      engine: mysql
      host: 10.0.0.2
      port: 3306
      username: reader
      password: secret
      database: myapp
      max_open_conns: 200    # 读库可以更大连接数

    analytics:
      driver: gormdriver
      engine: postgres
      host: 10.0.0.3
      port: 5432
      username: analyst
      password: secret
      database: analytics_db
      schema: analytics        # 连接级 schema，AutoMigrate 和所有查询自动使用此 schema

    analytics_raw:
      driver: gormdriver
      engine: postgres
      host: 10.0.0.3
      port: 5432
      username: analyst
      password: secret
      database: analytics_db
      # 不设置 schema，通过 .Schema() 动态切换

    cache_db:
      driver: gormdriver
      engine: sqlite
      database: database/cache.db
```

---

## Connection — 切换连接

```go
// 使用默认连接（main）
facades.DB().Query().Find(&users)

// 切换到指定连接
facades.DB().Connection("read_replica").Find(&users)
facades.DB().Connection("analytics").Raw("SELECT ...").Scan(&stats)
facades.DB().Connection("cache_db").Find(&caches)
```

---

## 读写分离模式

```go
// service/user_service.go
type UserService struct{}

// 写操作：用主库
func (s *UserService) Create(user *models.User) error {
    return facades.DB().Query().Create(user)
}

func (s *UserService) Update(user *models.User, updates map[string]any) error {
    return facades.DB().Query().Model(user).Updates(updates)
}

func (s *UserService) Delete(user *models.User) error {
    return facades.DB().Query().Delete(user)
}

// 读操作：用只读副本
func (s *UserService) FindByID(id string) (*models.User, error) {
    var user models.User
    err := facades.DB().Connection("read_replica").
        First(&user, "id = ?", id)
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (s *UserService) List(page, size int) ([]models.User, int64, error) {
    q := facades.DB().Connection("read_replica").
        Model(&models.User{}).
        Order("created_at DESC")

    var total int64
    q.Count(&total)

    var users []models.User
    err := q.Paginate(page, size).Find(&users)
    return users, total, err
}
```

---

## PostgreSQL 多 Schema

PostgreSQL 支持在同一个数据库实例内按 schema（命名空间）隔离表，GoFast 提供两种控制方式。

### 方式一：连接级 schema（推荐用于固定 schema）

在连接配置中声明 `schema`，该连接上的所有查询和 `AutoMigrate` 均自动使用该 schema：

```yaml
database:
  default: main
  connections:
    # 业务库：使用 public schema（默认）
    main:
      driver: gormdriver
      engine: postgres
      host: 127.0.0.1
      port: 5432
      database: myapp
      username: postgres
      password: secret
      schema: public

    # 数据分析库：使用 analytics schema
    analytics:
      driver: gormdriver
      engine: postgres
      host: 127.0.0.1
      port: 5432
      database: myapp
      username: analyst
      password: secret
      schema: analytics        # AutoMigrate 自动在 analytics schema 建表

    # 多租户：每个租户独立 schema
    tenant_acme:
      driver: gormdriver
      engine: postgres
      host: 127.0.0.1
      port: 5432
      database: myapp
      username: postgres
      password: secret
      schema: acme
```

```go
// AutoMigrate 在 analytics schema 下建表
facades.DB().Driver("analytics").AutoMigrate(&analytics.Event{})
// → CREATE TABLE analytics.events (...)

// 查询时自动带上 schema 前缀
facades.DB().Connection("analytics").Model(&analytics.Event{}).Find(&events)
// → SELECT * FROM analytics.events
```

### 方式二：动态 schema（推荐用于多租户）

在默认连接（或任意连接）上调用 `.Schema(name)` 临时切换，不影响连接池配置：

```go
// 同一连接，按请求动态路由到不同 schema
tenantSchema := "tenant_" + tenantID

facades.DB().Connection("pg").Schema(tenantSchema).
    Model(&models.Order{}).
    Where("status = ?", "pending").
    Find(&orders)
// → SELECT * FROM tenant_acme.orders WHERE status = 'pending'

// 也可和 Table() 一起用（Schema() 不会重复添加前缀）
facades.DB().Connection("pg").Schema(tenantSchema).Table("logs").
    Where("level = ?", "error").Find(&logs)
// → SELECT * FROM tenant_acme.logs WHERE level = 'error'
```

> `.Schema()` 的优先级说明：
> - `Schema(name).Table("tbl")` → `name.tbl`（Schema 中不含 `.` 的表名会加前缀）
> - `Schema(name).Table("other.tbl")` → `other.tbl`（表名已含 `.`，不重复添加）
> - `Schema(name).Model(&T{})` → GORM 解析 T 的表名后加上 `name.` 前缀

### 多租户 Middleware 示例

```go
// middleware/tenant.go
func TenantMiddleware(ctx contracts.Context) error {
    tenantID := ctx.Header("X-Tenant-ID")
    if tenantID == "" {
        return ctx.Response().BadRequest("missing X-Tenant-ID")
    }
    ctx.WithValue("tenant_schema", "tenant_"+tenantID)
    return ctx.Next()
}

// controller 中使用
func (c *OrderController) Index(ctx contracts.Context) error {
    schema := ctx.Value("tenant_schema").(string)

    var orders []models.Order
    err := facades.DB().Connection("pg").Schema(schema).
        Where("status = ?", "active").
        Order("created_at DESC").
        Find(&orders)
    if err != nil {
        return ctx.Response().Error(err)
    }
    return ctx.Response().Success(orders)
}
```

### AutoMigrate 多 Schema

```go
func (p *DatabaseProvider) MigrateDB(db contracts.DB) error {
    tenants := []string{"acme", "globex", "initech"}
    for _, name := range tenants {
        schema := "tenant_" + name
        // 先确保 schema 存在
        db.Connection("pg").Exec("CREATE SCHEMA IF NOT EXISTS " + schema)
        // 在该 schema 下迁移表结构
        if err := db.Driver("pg").(interface {
            AutoMigrateInSchema(schema string, models ...any) error
        }).AutoMigrateInSchema(schema,
            &models.Order{},
            &models.Product{},
        ); err != nil {
            // 降级：用 Schema().Exec 手动建表，或使用连接级 schema 连接
            tmpDB := db.Connection("pg").Schema(schema)
            _ = tmpDB.Exec("-- 在 " + schema + " 下初始化...")
        }
    }
    return nil
}
```

> **注意**：`AutoMigrate` 绑定在驱动（`Driver`）层，目前不直接支持运行时动态切换 schema 迁移。  
> 推荐做法是为每个需要迁移的 schema 创建独立的命名连接（配置 `schema` 字段），然后调用 `db.Driver("connection_name").AutoMigrate(...)`。

---

## Driver — 获取底层驱动

```go
// 获取默认连接驱动
driver := facades.DB().Driver()

// 获取指定连接驱动
driver := facades.DB().Driver("read_replica")

// 通用操作（与驱动无关）
if err := driver.Ping(); err != nil {
    log.Println("只读库连接异常:", err)
}

driver.AutoMigrate(&models.User{})
```

### 获取原始 *gorm.DB（逃生口）

```go
import gormdriver "github.com/zhoudm1743/go-fast/framework/database/drivers/gormdriver"

driver := facades.DB().Driver("main")
if gd, ok := driver.(*gormdriver.GormDriver); ok {
    rawDB := gd.RawDB() // *gorm.DB，可使用 GORM 特有功能
    // 注意：此操作引入 GORM 依赖，降低可移植性
}
```

---

## 跨库事务（不推荐）

跨数据库实例的事务需要引入分布式事务（如 Saga、TCC），GoFast 目前不内置支持。
建议通过业务补偿机制处理跨库一致性：

```go
// 推荐：各库独立事务 + 业务补偿
err1 := facades.DB().Transaction(func(tx contracts.Query) error {
    return tx.Create(&mainRecord)
})

err2 := facades.DB().Connection("analytics").Transaction(func(tx contracts.Query) error {
    return tx.Create(&analyticsRecord)
})

// 若 err2 失败，手动补偿（删除 mainRecord 或标记异常）
if err2 != nil {
    facades.DB().Query().Delete(&mainRecord)
}
```

---

## 注册自定义驱动

为不同连接注册不同的 ORM 驱动：

```go
// bootstrap/app.go 或 ServiceProvider.Register 中
import (
    "github.com/zhoudm1743/go-fast/framework/database"
    myxorm "github.com/my-org/gofast-xorm"
)

// 注册 xorm 驱动
database.RegisterDriver("xorm", func(cfg contracts.ConnectionConfig, log contracts.Log) (contracts.Driver, error) {
    return myxorm.NewDriver(cfg, log)
})
```

配置中使用：

```yaml
connections:
  legacy_db:
    driver: xorm          # 使用自定义驱动
    engine: mysql
    host: 10.0.0.4
    database: legacy
```

---

## 健康检查

```go
// 检查所有连接状态
func checkDatabaseHealth() map[string]string {
    connections := []string{"main", "read_replica", "analytics"}
    status := make(map[string]string)

    for _, name := range connections {
        if err := facades.DB().Driver(name).Ping(); err != nil {
            status[name] = "unhealthy: " + err.Error()
        } else {
            status[name] = "ok"
        }
    }
    return status
}

// 在健康检查接口中使用
func (c *HealthController) Check(ctx contracts.Context) error {
    return ctx.Response().Success(map[string]any{
        "database": checkDatabaseHealth(),
        "time":     time.Now().Unix(),
    })
}
```

---

## 多连接 AutoMigrate

```go
func (p *DatabaseProvider) MigrateDB(db contracts.DB) error {
    // 主库迁移
    if err := db.AutoMigrate(
        &models.User{},
        &models.Post{},
        &models.Order{},
    ); err != nil {
        return err
    }

    // Analytics 库单独迁移
    if err := db.Driver("analytics").AutoMigrate(
        &analytics.DailyReport{},
        &analytics.UserEvent{},
    ); err != nil {
        return err
    }

    return nil
}
```

---

## 多租户场景

GoFast 支持两种多租户隔离模式，可按需选择。

### 模式一：独立数据库（每租户一个连接）

适合强隔离要求，各租户数据库物理分离：

```go
// middleware/tenant.go
func TenantMiddleware(ctx contracts.Context) error {
    tenantID := ctx.Header("X-Tenant-ID")
    ctx.WithValue("tenant_db", "tenant_"+tenantID)
    return ctx.Next()
}

// controller 中使用
func (c *PostController) Index(ctx contracts.Context) error {
    tenantDB := ctx.Value("tenant_db").(string) // 如 "tenant_acme"

    var posts []models.Post
    facades.DB().Connection(tenantDB).Find(&posts)

    return ctx.Response().Success(posts)
}
```

配置中预先声明各租户连接：

```yaml
connections:
  tenant_acme:
    driver: gormdriver
    engine: mysql
    database: tenant_acme
    host: 10.0.1.1

  tenant_globex:
    driver: gormdriver
    engine: mysql
    database: tenant_globex
    host: 10.0.1.2
```

### 模式二：PostgreSQL Schema 隔离（每租户一个 schema）

适合中等规模租户，共享同一个 PostgreSQL 数据库实例，成本更低：

```go
// middleware/tenant.go
func TenantMiddleware(ctx contracts.Context) error {
    tenantID := ctx.Header("X-Tenant-ID")
    if tenantID == "" {
        return ctx.Response().BadRequest("missing X-Tenant-ID")
    }
    ctx.WithValue("tenant_schema", "tenant_"+tenantID)
    return ctx.Next()
}

// controller 中使用
func (c *PostController) Index(ctx contracts.Context) error {
    schema := ctx.Value("tenant_schema").(string)

    var posts []models.Post
    facades.DB().Connection("pg").Schema(schema).Find(&posts)
    // → SELECT * FROM tenant_acme.posts

    return ctx.Response().Success(posts)
}
```

配置中只需一个共享 PostgreSQL 连接（无需为每个租户声明连接）：

```yaml
connections:
  pg:
    driver: gormdriver
    engine: postgres
    host: 127.0.0.1
    port: 5432
    database: myapp
    username: postgres
    password: secret
    # 不设置 schema，由业务层通过 .Schema() 动态注入
```

> 连接采用**懒加载**机制，首次使用时才建立，不会在启动时耗尽连接池。



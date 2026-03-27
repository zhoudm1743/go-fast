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

按租户动态路由到不同数据库：

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

> 连接采用**懒加载**机制，首次使用时才建立，不会在启动时耗尽连接池。


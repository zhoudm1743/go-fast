# 数据库 Seeding（种子数据）

## 简介

Seeding 用于向数据库填充初始数据或测试数据。每个项目的种子数据各不相同，因此 **Seeder 代码放在 `app/console/commands/seeders/` 目录下**（属于业务层，而非框架层），由开发者自行维护。

框架提供统一的 `db:seed` 命令，支持默认连接与多租户连接。

---

## 目录结构

```
app/
  console/
    commands/
      seed.go                ← db:seed 命令（通常无需修改）
      seeders/
        seeder.go            ← Seeder 接口定义
        bootstrapper.go      ← TenantBootstrapper（多租户一键初始化）
        database_seeder.go   ← 根 Seeder（统一入口，按顺序调用子 Seeder）
        user_seeder.go       ← 用户表种子（示例）
bootstrap/
  commands.go                ← 注册 SeedCommand
```

---

## 第一步：定义 Seeder 接口

`Seeder.Run` 接收 `contracts.Query`，让同一套 Seeder 可以操作**任意连接**（默认库或租户库）。

```go
// app/console/commands/seeders/seeder.go
package seeders

import "github.com/zhoudm1743/go-fast/framework/contracts"

type Seeder interface {
    Run(q contracts.Query) error
}
```

---

## 第二步：编写业务 Seeder

### 用户 Seeder

```go
// app/console/commands/seeders/user_seeder.go
package seeders

import (
    "github.com/zhoudm1743/go-fast/app/models"
    "github.com/zhoudm1743/go-fast/framework/contracts"
    "golang.org/x/crypto/bcrypt"
)

type UserSeeder struct{}

func (s *UserSeeder) Run(q contracts.Query) error {
    // 幂等：已有数据则跳过
    var count int64
    q.Model(&models.User{}).Count(&count)
    if count > 0 {
        return nil
    }

    hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
    users := []models.User{
        {Name: "Alice", Email: "alice@example.com", Password: string(hash)},
        {Name: "Bob",   Email: "bob@example.com",   Password: string(hash)},
    }
    return q.Create(&users)
}
```

### 根 Seeder（统一入口）

```go
// app/console/commands/seeders/database_seeder.go
package seeders

import "github.com/zhoudm1743/go-fast/framework/contracts"

type DatabaseSeeder struct{}

func (s *DatabaseSeeder) Run(q contracts.Query) error {
    seeders := []Seeder{
        &UserSeeder{},
        // &AdminSeeder{},
        // &ProductSeeder{},
    }
    for _, seeder := range seeders {
        if err := seeder.Run(q); err != nil {
            return err
        }
    }
    return nil
}
```

---

## 第三步：注册命令

在 `bootstrap/commands.go` 中注册 `SeedCommand`：

```go
func Commands() []contracts.ConsoleCommand {
    return []contracts.ConsoleCommand{
        &commands.ExampleCommand{},
        &commands.SeedCommand{}, // ← 添加这行
    }
}
```

---

## 运行 Seed

```bash
# 运行所有 Seeder（默认连接）
go run . fast db:seed

# 运行指定 Seeder
go run . fast db:seed --class UserSeeder

# 对指定租户连接运行所有 Seeder
go run . fast db:seed --tenant tenant_acme

# 对指定租户连接运行指定 Seeder
go run . fast db:seed --tenant tenant_acme --class UserSeeder
```

---

## 多租户：迁移与种子

在 SaaS / 多租户应用中，**每个租户拥有独立的数据库**。平台创建租户后需要：

1. 对该租户的数据库执行表结构迁移（`AutoMigrate`）
2. 写入默认种子数据（`Seeder`）

| 工具 | 说明 |
|------|------|
| `facades.DB().Register(name, cfg)` | 运行时动态注册租户数据库连接 |
| `facades.DB().Driver(name).AutoMigrate(...)` | 对指定连接执行迁移 |
| `facades.DB().Connection(name)` | 获取指定连接的 Query，传入 Seeder |
| `seeders.TenantBootstrapper` | 封装迁移 + 种子的一键初始化工具 |

### 租户连接的两种注册方式

**方式一：config.yaml 预配置**（适合租户数量固定）

```yaml
database:
  default: main
  connections:
    main:
      engine: mysql
      host: 127.0.0.1
      port: 3306
      database: platform
      username: root
      password: secret
    tenant_acme:
      engine: mysql
      host: 127.0.0.1
      port: 3306
      database: tenant_acme
      username: root
      password: secret
```

**方式二：运行时动态注册**（适合租户数量动态增长）

```go
facades.DB().Register("tenant_acme", contracts.ConnectionConfig{
    Engine:   "mysql",
    Host:     "127.0.0.1",
    Port:     3306,
    Database: "tenant_acme",
    Username: "root",
    Password: "secret",
})
```

### 使用 TenantBootstrapper（推荐）

`TenantBootstrapper` 封装了「注册连接 → 迁移 → 种子」三步，适合在平台业务代码中调用：

```go
boot := seeders.NewTenantBootstrapper(
    []any{&models.User{}, &models.Order{}}, // 需要迁移的模型
    &seeders.DatabaseSeeder{},              // 种子入口
)

// 连接已在 config.yaml 预配置时
boot.Bootstrap("tenant_acme")

// 连接配置来自主库（动态注册 + 迁移 + 种子，一步完成）
boot.RegisterAndBootstrap("tenant_acme", contracts.ConnectionConfig{
    Engine:   "mysql",
    Host:     tenant.DBHost,
    Port:     tenant.DBPort,
    Database: tenant.DBName,
    Username: tenant.DBUser,
    Password: tenant.DBPassword,
})
```

`RegisterAndBootstrap` 内部依次执行：

```
1. facades.DB().Register(connName, cfg)    // 注册连接
2. Driver(connName).AutoMigrate(models...) // 迁移表结构
3. DatabaseSeeder.Run(q)                   // 写入种子数据
```

### 不使用 Bootstrapper（单独调用）

```go
connName := "tenant_acme"

// 仅迁移
facades.DB().Driver(connName).AutoMigrate(&models.User{}, &models.Order{})

// 仅种子
q := facades.DB().Connection(connName)
(&seeders.DatabaseSeeder{}).Run(q)
```

### 租户专属 Seeder

不同租户需要不同初始数据时，可定义独立的 Seeder：

```go
type PremiumTenantSeeder struct{}

func (s *PremiumTenantSeeder) Run(q contracts.Query) error {
    features := []models.Feature{
        {Name: "advanced_analytics", Enabled: true},
        {Name: "api_access", Enabled: true},
    }
    return q.Create(&features)
}
```

根据租户类型选择 Seeder：

```go
func onTenantCreated(tenant *models.Tenant) error {
    var seeder seeders.Seeder
    if tenant.Plan == "premium" {
        seeder = &seeders.PremiumTenantSeeder{}
    } else {
        seeder = &seeders.DatabaseSeeder{}
    }
    boot := seeders.NewTenantBootstrapper(TenantModels, seeder)
    return boot.RegisterAndBootstrap("tenant_"+tenant.ID, tenant.DBConfig())
}
```

### 完整流程图

```
平台创建租户
     │
     ▼
facades.DB().Register("tenant_xxx", cfg)    ← 注册连接
     │
     ▼
Driver("tenant_xxx").AutoMigrate(models...) ← 迁移表结构
     │
     ▼
DatabaseSeeder.Run(                          ← 种子数据
    facades.DB().Connection("tenant_xxx"),
)
     │
     ├── UserSeeder.Run(q)
     ├── AdminSeeder.Run(q)
     └── ...
```

---

## 注意事项

1. **幂等性**：每个 Seeder 应检查数据是否已存在，避免重复运行时重复插入。
2. **先迁移再种子**：`AutoMigrate` 必须在 `Seeder` 之前执行，否则表不存在。
3. **密码安全**：种子数据中的密码必须使用 `bcrypt` 等哈希算法，禁止存储明文。
4. **顺序**：`DatabaseSeeder` 中各子 Seeder 的调用顺序很重要，有外键依赖时请确保先插父表数据。
5. **事务**：若需要原子性操作，可在 Seeder 中使用事务：

```go
func (s *UserSeeder) Run(q contracts.Query) error {
    return q.Transaction(func(tx contracts.Query) error {
        return tx.Create(&models.User{...})
    })
}
```

6. **连接生命周期**：动态注册的连接会在应用关闭时由框架统一关闭（`Close()`）。


# 连接数据库

## 支持的数据库引擎

| 引擎 | `engine` 值 | 说明 |
|------|-------------|------|
| SQLite | `sqlite` / `sqlite3` | 开发/测试首选，无需安装服务 |
| MySQL | `mysql` | 生产常用 |
| PostgreSQL | `postgres` | 生产常用 |
| SQL Server | `mssql` | 企业场景 |

---

## 配置格式

### 旧格式（扁平化，向后兼容）

```yaml
# config/config.yaml
database:
  driver: sqlite               # 同时代表引擎名
  database: database/gofast.db
  max_idle_conns: 10
  max_open_conns: 100
  conn_max_lifetime: 60        # 分钟
  conn_max_idle_time: 30       # 分钟
  log_level: info              # error | warn | info | silent
  slow_threshold: 200          # ms，超过此值输出慢查询日志
```

> 旧格式自动适配为 `connections.main`，**无需任何代码改动即可升级**。

### 新格式（多连接，推荐）

```yaml
database:
  default: main                # 默认连接名

  connections:
    main:
      driver: gormdriver        # ORM 驱动标识
      engine: sqlite
      database: database/gofast.db
      max_idle_conns: 10
      max_open_conns: 100
      conn_max_lifetime: 60
      conn_max_idle_time: 30
      log_level: info
      slow_threshold: 200

    read_replica:
      driver: gormdriver
      engine: mysql
      host: 10.0.0.2
      port: 3306
      username: reader
      password: secret
      database: myapp
      charset: utf8mb4
      loc: Local
```

---

## 各引擎配置详解

### SQLite

```yaml
connections:
  main:
    driver: gormdriver
    engine: sqlite
    database: database/app.db   # 文件路径（相对于项目根目录）
```

也可以使用内存数据库（测试用）：

```yaml
database: ":memory:"
```

### MySQL

```yaml
connections:
  main:
    driver: gormdriver
    engine: mysql
    host: 127.0.0.1
    port: 3306
    username: root
    password: secret
    database: myapp
    charset: utf8mb4            # 推荐 utf8mb4 支持 emoji
    loc: Local                  # 时区
```

完整 DSN 格式（也可直接填 `dsn` 字段）：
```
user:pass@tcp(host:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local
```

### PostgreSQL

```yaml
connections:
  main:
    driver: gormdriver
    engine: postgres
    host: 127.0.0.1
    port: 5432
    username: postgres
    password: secret
    database: myapp
    ssl_mode: disable           # disable | require | verify-full
    loc: Local
    schema: public              # 可选，设置默认 search_path，省略则不设置
```

> **`schema` 字段**：
> - 会将 `search_path=<schema>` 写入 DSN，所有原生 SQL（`Exec`/`Raw`）默认在该 schema 下执行
> - 同时配置 GORM `NamingStrategy`，使 `AutoMigrate` 和 GORM 结构化查询自动带上 `schema.` 前缀
> - 若同时含有 `table_prefix`，实际前缀为 `schema.table_prefix`

### SQL Server

```yaml
connections:
  main:
    driver: gormdriver
    engine: mssql
    host: 127.0.0.1
    port: 1433
    username: sa
    password: secret
    database: myapp
```

---

## 直连 DSN

所有引擎均支持直接填写 `dsn` 字段，优先级高于分项配置：

```yaml
connections:
  main:
    driver: gormdriver
    engine: mysql
    dsn: "root:secret@tcp(127.0.0.1:3306)/myapp?charset=utf8mb4&parseTime=True&loc=Local"
```

---

## 连接池配置

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `max_idle_conns` | 10 | 最大空闲连接数 |
| `max_open_conns` | 100 | 最大打开连接数 |
| `conn_max_lifetime` | 60 | 连接最大复用时间（分钟） |
| `conn_max_idle_time` | 30 | 连接最大空闲时间（分钟） |

生产环境推荐配置：

```yaml
max_idle_conns: 25
max_open_conns: 200
conn_max_lifetime: 30
conn_max_idle_time: 15
```

---

## 日志配置

| `log_level` | 说明 |
|-------------|------|
| `silent` | 不输出任何日志 |
| `error` | 仅输出错误 |
| `warn` | 输出警告（含慢查询） |
| `info` | 输出全部 SQL（开发推荐） |

`slow_threshold`（毫秒）：超过阈值的 SQL 会被标记为慢查询，无论 `log_level` 如何都会输出警告。

```yaml
log_level: warn
slow_threshold: 500   # 超过 500ms 输出慢查询警告
```

---

## 在代码中获取连接

```go
// 获取默认连接的查询构建器
q := facades.DB().Query()

// 获取指定连接的查询构建器
q := facades.DB().Connection("read_replica")

// PostgreSQL：动态切换 schema（对连接级 schema 的有益补充）
q := facades.DB().Connection("pg").Schema("analytics")

// 获取底层驱动（逃生口，不推荐常规使用）
driver := facades.DB().Driver()          // 默认连接
driver := facades.DB().Driver("main")   // 指定连接

// 若使用 GORM 驱动，可获取原始 *gorm.DB
import gormdriver "github.com/zhoudm1743/go-fast/framework/database/drivers/gormdriver"
if gd, ok := driver.(*gormdriver.GormDriver); ok {
    rawDB := gd.RawDB() // *gorm.DB
}
```

---

## 注册自定义驱动

如需接入 xorm 等第三方驱动：

```go
// 实现 contracts.Driver 和 contracts.Query 接口，然后注册：
database.RegisterDriver("xorm", func(cfg contracts.ConnectionConfig, log contracts.Log) (contracts.Driver, error) {
    return myxorm.NewDriver(cfg, log)
})
```

配置中将 `driver` 改为 `xorm` 即可，业务代码无需改动。

---

## 多连接切换示例

```go
// 写操作用主库
if err := facades.DB().Query().Create(&user); err != nil {
    return err
}

// 读操作用只读副本
var users []models.User
facades.DB().Connection("read_replica").
    Where("status = ?", 1).
    Find(&users)
```

> 详见 [多数据库连接](./multi-connection.md)。


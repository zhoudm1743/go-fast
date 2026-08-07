# GoFast 配置说明

> GoFast 支持两种配置方式：YAML 配置文件 和 Go 源代码配置文件。两者可同时使用，YAML 中的值会覆盖 Go 配置中的同名键。

---

## 一、配置方式概览

| 方式 | 文件位置 | 格式 | 适用场景 |
|------|---------|------|---------|
| YAML 配置 | `config/config.yaml` | YAML | 运行时覆盖、环境切换 |
| Go 配置文件 | `config/*.go` | Go 源码 | 默认值、类型安全、IDE 补全 |

### 优先级（低 -> 高）

```
Go 配置文件 (config/*.go init)  <  YAML 文件 (config/config.yaml)  <  运行时 Set()
```

- Go 配置文件提供**默认值**，定义在 `init()` 函数中
- YAML 文件提供**覆盖值**，用于不同环境（开发/生产）的差异化配置
- 运行时 `facades.Config().Set()` 拥有最高优先级

---

## 二、Go 配置文件

### 工作原理

1. `config/` 目录下的每个 `.go` 文件通过 `init()` 函数注册一个配置命名空间
2. `bootstrap/app.go` 中的空白导入 `_ "github.com/zhoudm1743/go-fast/config"` 触发所有 `init()` 执行
3. 框架启动时，将注册的配置写入默认值层；YAML 中的同名键自动覆盖默认值

### 现有配置文件

| 文件 | 命名空间 | 说明 |
|------|---------|------|
| `config/app.go` | `app` | 应用基本信息 |
| `config/server.go` | `server` | HTTP 服务器 |
| `config/jwt.go` | `jwt` | JWT 鉴权 |
| `config/view.go` | `view` | 模板引擎 |
| `config/session.go` | `session` | 会话管理 |
| `config/database.go` | `database` | 数据库连接 |
| `config/log.go` | `log` | 日志 |
| `config/filesystem.go` | `filesystem` | 文件存储 |
| `config/cache.go` | `cache` | 缓存 |

### 编写自己的配置文件

```go
// config/myapp.go
package config

import fwconfig "github.com/zhoudm1743/go-fast/framework/config"

func init() {
    fwconfig.Add("myapp", map[string]any{
        "api_endpoint": "https://api.example.com",
        "timeout":      30,
        "features": map[string]any{
            "new_ui":     true,
            "analytics":  false,
        },
    })
}
```

访问配置：

```go
import "github.com/zhoudm1743/go-fast/framework/facades"

endpoint := facades.Config().GetString("myapp.api_endpoint")
timeout  := facades.Config().GetInt("myapp.timeout", 10)
```

### 运行时注册

除 `init()` 外，也可在 ServiceProvider 中运行时注册：

```go
func (sp *MyProvider) Boot(app foundation.Application) error {
    app.Config().Add("myapp", map[string]any{
        "dynamic_key": computeValue(),
    })
    return nil
}
```

---

## 三、YAML 配置文件

YAML 文件路径：`config/config.yaml`，所有 Go 配置默认值可被 YAML 覆盖。

```yaml
# 覆盖 Go 默认值示例
server:
  port: 8080        # 覆盖 config/server.go 中的 3000
  mode: release     # 覆盖 debug

database:
  driver: mysql     # 覆盖 SQLite 默认值
  host: 10.0.0.5
  port: 3306
  database: gofast_prod
  username: prod_user
  password: ${DB_PASSWORD}
```

---

## 四、读取配置

### 通过 Facade

```go
import "github.com/zhoudm1743/go-fast/framework/facades"

// 通用读取
val := facades.Config().Get("app.name")
val := facades.Config().Get("cache.memory.shard_count", 32)

// 类型化读取
name    := facades.Config().GetString("app.name", "GoFast")
port    := facades.Config().GetInt("server.port", 3000)
debug   := facades.Config().GetBool("app.debug", false)
rate    := facades.Config().GetFloat64("threshold", 0.85)

// 切片和字典
origins := facades.Config().GetStringSlice("server.cors_allow_origins")
dbCfg   := facades.Config().GetStringMap("database")
```

### 通过 Application

```go
app.Config().GetString("app.name")
```

### 环境变量

```go
// 直接读取系统环境变量（不经过配置文件）
mode := facades.Config().Env("APP_MODE", "development")
```

---

## 五、各配置项参考

### app -- 应用信息

| 键 | 类型 | 默认值 | 说明 |
|----|------|--------|------|
| `app.name` | string | `GoFast` | 应用名称 |
| `app.env` | string | `production` | 运行环境 |
| `app.debug` | bool | `false` | 调试模式 |

### server -- HTTP 服务器

| 键 | 类型 | 默认值 | 说明 |
|----|------|--------|------|
| `server.name` | string | `GoFast` | 服务名称 |
| `server.driver` | string | `gin` | HTTP 驱动：gin / fiber |
| `server.host` | string | `0.0.0.0` | 监听地址 |
| `server.port` | int | `3000` | 监听端口 |
| `server.mode` | string | `debug` | 运行模式：debug / release |
| `server.read_timeout_sec` | int | `30` | 读超时（秒） |
| `server.write_timeout_sec` | int | `30` | 写超时（秒） |
| `server.idle_timeout_sec` | int | `120` | 空闲超时（秒） |
| `server.shutdown_timeout_sec` | int | `10` | 优雅关闭超时（秒） |
| `server.prefork` | bool | `false` | Prefork 模式（仅 fiber） |
| `server.body_limit_mb` | int | `10` | 请求体限制（MB） |
| `server.cors_allow_origins` | []string | `["*"]` | CORS 来源白名单 |

### jwt -- JWT 鉴权

| 键 | 类型 | 默认值 | 说明 |
|----|------|--------|------|
| `jwt.secret` | string | `change-me-to-a-random-secret` | 签名密钥 |
| `jwt.ttl` | int | `60` | Token 有效期（分钟） |
| `jwt.alg` | string | `HS256` | 签名算法 |

### view -- 模板引擎

| 键 | 类型 | 默认值 | 说明 |
|----|------|--------|------|
| `view.dir` | string | `resources/views` | 模板目录 |
| `view.extension` | string | `.html` | 文件扩展名 |
| `view.reload` | bool | `true` | 热重载 |

### session -- 会话

| 键 | 类型 | 默认值 | 说明 |
|----|------|--------|------|
| `session.lifetime` | int | `7200` | 有效期（秒） |
| `session.cookie` | string | `go_fast_session` | Cookie 名称 |

### database -- 数据库

每个数据库连接为一个独立的 map，配置在 `database.connections` 下。

| 键 | 类型 | 默认值 | 说明 |
|----|------|--------|------|
| `database.default` | string | `main` | 默认连接名称 |

**连接配置** (`database.connections.<name>.*`)：

| 键 | 类型 | 默认值 | 说明 |
|----|------|--------|------|
| `driver` | string | `gormdriver` | 驱动注册名 |
| `engine` | string | `sqlite` | 引擎：sqlite / mysql / postgres / mssql |
| `database` | string | `database/gofast.db` | 数据库名/文件路径 |
| `host` | string | `localhost` | 主机地址 |
| `port` | int | `3306` | 端口 |
| `username` | string | `root` | 用户名 |
| `password` | string | `""` | 密码 |
| `charset` | string | `utf8mb4` | 字符集 |
| `loc` | string | `Local` | 时区 |
| `ssl_mode` | string | `""` | SSL 模式（PostgreSQL） |
| `max_idle_conns` | int | `10` | 最大空闲连接数 |
| `max_open_conns` | int | `100` | 最大打开连接数 |
| `conn_max_lifetime` | int | `60` | 连接最大存活时间（分钟） |
| `conn_max_idle_time` | int | `30` | 连接最大空闲时间（分钟） |
| `log_level` | string | `info` | GORM 日志级别：silent / error / warn / info |
| `slow_threshold` | int | `200` | 慢查询阈值（毫秒），0 表示不记录 |

### log -- 日志

| 键 | 类型 | 默认值 | 说明 |
|----|------|--------|------|
| `log.mode` | string | `hybrid` | 输出模式：console / file / hybrid |
| `log.level` | string | `debug` | 日志级别 |
| `log.format` | string | `color` | 控制台格式：color / json / text |
| `log.file_format` | string | `json` | 文件格式：json / text |
| `log.output_path` | string | `storage/logs/app.log` | 日志文件路径 |
| `log.timestamp_format` | string | `2006-01-02 15:04:05` | 时间戳格式 |
| `log.max_size` | int | `100` | 单文件最大 MB |
| `log.max_backups` | int | `5` | 保留旧文件数 |
| `log.max_age` | int | `30` | 保留天数 |
| `log.compress` | bool | `false` | 是否压缩 |

### filesystem -- 文件存储

| 键 | 类型 | 默认值 | 说明 |
|----|------|--------|------|
| `filesystem.default` | string | `local` | 默认磁盘名称 |
| `filesystem.disks.<name>.driver` | string | `local` | 驱动：local / oss / cos / minio / s3 |

**local 驱动参数：**

| 键 | 类型 | 默认值 | 说明 |
|----|------|--------|------|
| `filesystem.disks.local.root` | string | `storage/app` | 存储根目录 |
| `filesystem.disks.local.url` | string | `/storage` | 访问 URL 前缀 |

**OSS 驱动参数：** `key`, `secret`, `bucket`, `url`, `endpoint`

**COS 驱动参数：** `key`, `secret`, `url`

**MinIO 驱动参数：** `key`, `secret`, `bucket`, `url`, `endpoint`, `region`, `ssl`

**S3 驱动参数：** `key`, `secret`, `region`, `bucket`, `url`, `token`, `endpoint`, `cdn`, `object_canned_acl`, `use_path_style`

### cache -- 缓存

| 键 | 类型 | 默认值 | 说明 |
|----|------|--------|------|
| `cache.driver` | string | `memory` | 缓存驱动：memory / redis |
| `cache.memory.shard_count` | int | `32` | 内存分片数 |
| `cache.memory.clean_interval` | int | `60` | 清理间隔（秒） |
| `cache.redis.host` | string | `127.0.0.1` | Redis 主机地址 |
| `cache.redis.port` | int | `6379` | Redis 端口 |
| `cache.redis.password` | string | `""` | Redis 密码 |
| `cache.redis.db` | int | `0` | Redis 数据库编号 |
| `cache.redis.prefix` | string | `""` | 缓存键前缀 |

---

## 六、插件默认配置

插件可通过实现 `foundation.ConfigProvider` 接口声明默认配置项，框架在 Boot 阶段自动调用：

```go
type MyPlugin struct{}

func (p *MyPlugin) ConfigDefaults() map[string]any {
    return map[string]any{
        "myplugin.timeout": 30,
        "myplugin.retries": 3,
    }
}
```

插件默认值优先级低于 Go 配置文件和 YAML。

---

## 七、下一步

- [快速开始](getting-started.md)
- [Facade 使用说明](facade.md)
- [编写自定义 Provider](service-provider.md)

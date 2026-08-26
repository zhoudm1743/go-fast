# AGENT.md — GoFast 项目 AI 协作指南

> 本文档供 AI Agent 在 **go-fast**（应用骨架）仓库中开发时使用。
> 框架内核为独立 Go module：[go-fast-framework](https://github.com/zhoudm1743/go-fast-framework) **v0.7.10**。
> 阅读本文后再动手，可避免目录放错、Provider 阶段误用、配置触发意外行为等问题。

---

## 0. 快速上下文


| 项         | 值                                                              |
| --------- | -------------------------------------------------------------- |
| Go 版本     | 1.25+                                                          |
| 模块名       | `github.com/zhoudm1743/go-fast`                                |
| 框架依赖      | `github.com/zhoudm1743/go-fast-framework v0.7.10`              |
| HTTP 默认引擎 | Gin（`config/server.go` → `driver: gin`）                        |
| 默认缓存驱动    | **file**（`storage/cache/`，跨重启持久化）                              |
| 默认数据库     | SQLite（`database/gofast.db`）                                   |
| 入口        | `main.go` → `bootstrap.Boot()` → `routes.Register()` → HTTP 启动 |


**两种运行模式：**

```bash
go run .              # HTTP 服务（默认 port 3000）
go run . fast <cmd>   # Fast 控制台（migrate / seed / example 等）
```

---



## 1. 仓库边界


| 仓库                    | 职责                                    | Agent 是否改代码        |
| --------------------- | ------------------------------------- | ------------------ |
| **go-fast**（本仓库）      | 业务骨架：路由、控制器、模型、配置、自定义 ServiceProvider | ✅ 是                |
| **go-fast-framework** | 框架内核：IoC、Facade、日志、DB、缓存、HTTP 等       | ❌ 否（仅 `go get` 升级） |


**原则：**

- 业务逻辑写在本仓库；框架能力通过 `github.com/zhoudm1743/go-fast-framework/*` 导入。
- 不在本仓库复刻框架代码；缺能力时升级框架版本或在本仓库写 ServiceProvider。
- 非必要不新建 `docs/` 文件；非用户要求不提交 git。

---



## 2. 目录地图

```
go-fast/
├── app/                         # 业务入口层（HTTP / CLI / 异步）
│   ├── http/
│   │   ├── app/                 # 前台：controllers / middleware / requests
│   │   └── admin/               # 后台：controllers / middleware / requests
│   ├── console/commands/        # Fast 自定义命令
│   ├── events/                  # 事件（contracts.Eventer）
│   ├── listeners/               # 监听器（contracts.EventListener）
│   ├── jobs/                    # 队列任务（contracts.QueueJob）
│   └── models/                  # GORM 模型（嵌入 database.Model）
│
├── services/                    # ★ 业务服务层
│   ├── contracts/               # 业务契约（本仓库专用接口）
│   └── <name>/                  # 服务实现 + service_provider.go + _test.go
│
├── bootstrap/
│   ├── app.go                   # Provider 列表、事件/调度、Boot 流程
│   └── commands.go              # Fast 命令注册
│
├── config/
│   ├── *.go                     # Go 默认值（init 注册）
│   └── config.yaml              # YAML 运行时覆盖
│
├── routes/
│   ├── api.go                   # Register() 统一入口
│   ├── app.go                   # 前台路由 /api/...
│   └── admin.go                 # 后台路由 /admin/...、/pages/...
│
├── resources/                   # 静态资源 & HTML 模板
├── storage/                     # 运行时（gitignore，勿提交）
├── database/migrations/           # 迁移（如有）
└── docs/                        # 人类文档
```



### 2.1 代码归属决策树

```
需要写什么？
├─ HTTP 接口 / 页面       → app/http/{app|admin}/controllers + routes/
├─ 请求参数校验 struct      → app/http/{app|admin}/requests/
├─ 鉴权 / 跨切面逻辑        → app/http/{app|admin}/middleware/
├─ 数据库表结构             → app/models/
├─ 可复用业务逻辑（多入口）  → services/<name>/ + services/contracts/
├─ 后台异步任务             → app/jobs/ + facades.Queue()
├─ 事件驱动解耦             → app/events/ + app/listeners/
├─ CLI 命令                → app/console/commands/ + bootstrap/commands.go
├─ 配置默认值               → config/<ns>.go
└─ 环境差异覆盖             → config/config.yaml
```

---



## 3. import 规范

本项目存在**两套 contracts**，不可混用：


| import 路径                                           | 用途                                         |
| --------------------------------------------------- | ------------------------------------------ |
| `github.com/zhoudm1743/go-fast-framework/contracts` | 框架契约：Context、DB、Cache、Controller、Eventer 等 |
| `github.com/zhoudm1743/go-fast/services/contracts`  | **本仓库**业务契约：Test 等自定义服务接口                  |
| `github.com/zhoudm1743/go-fast-framework/facades`   | 所有 Facade 访问（内置 + 自定义服务解析入口）               |


**推荐别名（多包并存时）：**

```go
import (
    appcontracts "github.com/zhoudm1743/go-fast/services/contracts"
    "github.com/zhoudm1743/go-fast-framework/contracts"
    "github.com/zhoudm1743/go-fast-framework/facades"
)
```

**禁止：**

- 控制器 / 中间件 import `github.com/gin-gonic/gin` 或 `github.com/gofiber/fiber`
- 在 `Register()` 阶段 `MustMake()` 其他服务
- 在项目根或 `services/` 外散落 `contracts/`、`facades/` 包

---



## 4. IoC 容器与 Facade



### 4.1 框架内置容器 Key


| Key         | Facade 访问                        | Provider                      |
| ----------- | -------------------------------- | ----------------------------- |
| `config`    | `facades.Config()`               | config.ServiceProvider        |
| `log`       | `facades.Log()`                  | log.ServiceProvider           |
| `cache`     | `facades.Cache()`                | cache.ServiceProvider         |
| `tenant`    | `facades.Tenant()`               | tenant.ServiceProvider        |
| `db`        | `facades.DB()`                   | database.ServiceProvider      |
| `storage`   | `facades.Storage()`              | filesystem.ServiceProvider    |
| `jwt`       | `facades.Http.JWT()`             | jwt.ServiceProvider           |
| `route`     | `facades.Http.Route()`           | http.ServiceProvider          |
| `validator` | `facades.Http.Validator()`       | http.ServiceProvider          |
| `session`   | `facades.Http.Session()`         | http.ServiceProvider          |
| `fast`      | `facades.Fast()`                 | fast.ServiceProvider          |
| `event`     | `facades.Event()`                | event.ServiceProvider         |
| `queue`     | `facades.Queue()`                | queue.ServiceProvider         |
| `schedule`  | `facades.Schedule()`             | schedule.ServiceProvider      |
| `test`      | `facades.App().MustMake("test")` | services/test.ServiceProvider |


> `orm` 已 Deprecated，统一用 `facades.DB()`。



### 4.2 ServiceProvider 两阶段

```
Phase 1 — Register（按 providers() 声明顺序）
  只做 Bind / Singleton / Instance
  ❌ 不可 MustMake 其他服务

Phase 2 — Boot（同上顺序）
  ✅ 可 MustMake 任意已注册服务
  ✅ 可注册 app.OnShutdown() 释放资源
```



### 4.3 自定义服务：注册与访问

**注册（容器）：**

```go
// services/<name>/service_provider.go
func (sp *ServiceProvider) Register(app foundation.Application) {
    app.Singleton("my_svc", func(app foundation.Application) (any, error) {
        cfg := app.MustMake("config").(contracts.Config)
        log := app.MustMake("log").(contracts.Log)
        return NewMyService(cfg, log)
    })
}
```

**访问（Facade）：**

```go
// 内置
facades.Log().Info("ok")

// 自定义 — 通过框架 facades.App() 解析，无需 services/facades/ 包
svc := facades.App().MustMake("my_svc").(appcontracts.MyService)
```

**为何不建** `services/facades/`**：**

Go 模块无法向 `go-fast-framework/facades` 追加函数。`facades.Xxx()` 顶级函数仅存在于框架源码中；应用层自定义服务用 `facades.App().MustMake("key")` 即为标准 Facade 模式。

---



## 5. 启动流程详解

```
main.go
│
├─ bootstrap.Boot()
│   ├─ foundation.NewApplication(".")
│   ├─ app.SetProviders(providers())
│   ├─ app.Boot()                          ← 全部 Register → Boot
│   ├─ facades.SetApp(app)                 ← ★ 此后 Facade 可用
│   ├─ goevent.RegisterEvents(...)         ← 事件映射
│   ├─ goschedule.RegisterSchedule(...)    ← 定时任务（已启动 cron）
│   └─ facades.Fast().Register(Commands())
│
├─ [fast 模式] facades.Fast().Run(args) → return
│
├─ routes.Register()
│   ├─ RegisterApp()    → /api/ping, /api/v1/...
│   └─ RegisterAdmin()  → /admin/..., /pages/..., /static/...
│
├─ go facades.Http.Route().Run()           ← 非阻塞
└─ 信号监听 → app.Shutdown()               ← 逆序 OnShutdown
```



### 5.1 Provider 顺序（bootstrap/app.go）

被依赖者在前，**不可随意调换**：

```
config → log → cache → tenant → database → filesystem
→ jwt → http → fast → event → queue → schedule
→ [自定义 Provider...]
```



### 5.2 Boot 后注册项

以下必须在 `facades.SetApp(app)` **之后**调用（已在 `bootstrap/app.go` 实现）：


| 注册  | 函数                                                  | 说明         |
| --- | --------------------------------------------------- | ---------- |
| 事件  | `goevent.RegisterEvents(app, map)`                  | 事件 → 监听器列表 |
| 调度  | `goschedule.RegisterSchedule(app, []ScheduleEvent)` | 启动 cron    |
| 命令  | `facades.Fast().Register(Commands())`               | Fast 控制台命令 |


---



## 6. HTTP 层规范



### 6.1 路由架构

- **控制器自注册**：实现 `contracts.Controller`，在 `Boot(r contracts.Route)` 内声明路由。
- **路由文件只做编排**：`routes/app.go` / `routes/admin.go` 负责 Group、中间件、Register。
- **入口**：`routes/api.go` → `Register()` 调用各模块。

```go
// routes/app.go 典型模式
r.Group("/api/v1", appMiddleware.Auth, func(v1 contracts.Route) {
    v1.Register(&controllers.UserController{})
})
// 实际路径 = Group 前缀 + Controller.Prefix() + Boot 内路径
// 例：/api/v1 + /user + /profile → GET /api/v1/user/profile
```



### 6.2 控制器模板

```go
type XxxController struct{}

func (c *XxxController) Prefix() string { return "/xxx" }  // 可选 Prefixer

func (c *XxxController) Boot(r contracts.Route) {
    r.Get("/list", c.List)
}

func (c *XxxController) List(ctx contracts.Context) error {
    var req requests.ListRequest
    if err := ctx.Bind(&req); err != nil {
        return ctx.Response().Validation(err)
    }
    // ...
    return ctx.Response().Success(data)
}
```



### 6.3 响应 helper


| 方法                                 | 场景     |
| ---------------------------------- | ------ |
| `ctx.Response().Success(data)`     | 200 成功 |
| `ctx.Response().Fail(code, msg)`   | 业务失败   |
| `ctx.Response().Validation(err)`   | 参数校验失败 |
| `ctx.Response().Unauthorized(msg)` | 401    |
| `ctx.Response().NotFound(msg)`     | 404    |
| `ctx.JSON(status, data)`           | 完全自定义  |




### 6.4 参数绑定

```go
type ListReq struct {
    Page  int    `query:"page"  default:"1"`
    Size  int    `query:"size"  default:"20"`
    Name  string `json:"name"   binding:"required,min=2"`
}
ctx.Bind(&req)  // 零值字段自动填 default；binding 校验失败 → Validation()
```



### 6.5 中间件

- 签名：`func(ctx contracts.Context) error`
- 鉴权后注入：`ctx.WithValue("user_id", id)` → 下游 `ctx.Value("user_id")`
- 放行：`return ctx.Next()`

---



## 7. 数据库与模型



### 7.1 模型

```go
type User struct {
    database.Model              // 时序 ID 主键 + CreatedAt/UpdatedAt
    Name  string `gorm:"size:100;not null" json:"name"`
}
```

- 主键 ID 由框架自动生成（UUID v7 风格，16 字符），**不要手动赋值**。
- 软删除：嵌入 `database.ModelWithSoftDelete`。



### 7.2 查询

```go
facades.DB().Query().Find(&users)                          // 默认连接
facades.DB().Connection("replica").Find(&users)            // 指定连接
facades.DB().Transaction(func(tx contracts.Query) error {   // 事务
    return tx.Create(&user)
})
```



### 7.3 配置格式

```yaml
database:
  default: main
  connections:
    main:
      driver: gormdriver
      engine: sqlite
      database: database/gofast.db
```

旧版扁平格式（`database.driver`）框架仍兼容，但示例项目已统一用 `connections` 格式。

---



## 8. 配置系统



### 8.1 优先级

```
config/*.go（Go 默认值）  <  config/config.yaml  <  运行时 facades.Config().Set()
```



### 8.2 新增配置步骤

1. 新建 `config/<namespace>.go`，`init()` 中 `fwconfig.Add(...)`
2. 可选：在 `config/config.yaml` 写覆盖值
3. 确保 `bootstrap/app.go` 有 `_ "github.com/zhoudm1743/go-fast/config"`



### 8.3 配置命名空间速查


| 命名空间         | 文件                   | 要点                               |
| ------------ | -------------------- | -------------------------------- |
| `server`     | config/server.go     | driver(gin/fiber)、port、mode、cors |
| `database`   | config/database.go   | `connections.<name>.*`           |
| `cache`      | config/cache.go      | driver 默认 **file**；见下方陷阱         |
| `log`        | config/log.go        | mode: hybrid / console / file    |
| `jwt`        | config/jwt.go        | secret、ttl、guards                |
| `filesystem` | config/filesystem.go | disks.local.root                 |
| `session`    | config/session.go    | lifetime、cookie                  |
| `view`       | config/view.go       | dir、extension、reload             |
| `queue`      | config/queue.go      | workers（0 = NumCPU）              |
| `test`       | config/test.go       | 示例服务 greeting_prefix             |




### 8.4 ⚠️ 缓存配置陷阱

框架逻辑：只要 `cache.redis.host` **非空**，就会尝试连接 Redis（与 `driver` 无关），以便 `Store("redis")` 可用。

```go
// ❌ 错误：driver=file 但预填 host，启动时会报 Redis 连接失败
"redis": map[string]any{"host": "127.0.0.1", ...}

// ✅ 正确：默认留空，需要 Redis 时在 config.yaml 显式配置
"redis": map[string]any{"host": "", ...}
```

缓存降级链：`redis → file → memory`（仅影响默认 Store，不影响 `Store("xxx")` 显式切换）。

---



## 9. 常见任务清单



### 9.1 新增业务服务

参照 `services/test/`，按序完成：


| #   | 文件                                         | 动作                 |
| --- | ------------------------------------------ | ------------------ |
| 1   | `services/contracts/<name>.go`             | 定义 interface       |
| 2   | `services/<name>/<name>.go`                | 实现 + 构造函数          |
| 3   | `services/<name>/service_provider.go`      | Register + Boot    |
| 4   | `services/<name>/service_provider_test.go` | 测试                 |
| 5   | `config/<name>.go`                         | 配置默认值              |
| 6   | `config/config.yaml`                       | 可选覆盖               |
| 7   | `bootstrap/app.go`                         | `providers()` 末尾追加 |




### 9.2 新增 HTTP 控制器

1. `app/http/{app|admin}/controllers/xxx_controller.go`
2. 可选：`app/http/{app|admin}/requests/xxx.go`
3. `routes/{app|admin}.go` 中 Group + Register
4. 仅用 `facades.*` 和 `contracts.Context`，不引 Gin/Fiber



### 9.3 新增 Fast 命令

1. `app/console/commands/xxx.go` — 实现 `contracts.ConsoleCommand`
2. `bootstrap/commands.go` — 追加到 `Commands()` 切片



### 9.4 新增事件

1. `app/events/xxx.go` — 实现 `contracts.Eventer`
2. `app/listeners/xxx.go` — 实现 `contracts.EventListener`
3. `bootstrap/app.go` — `goevent.RegisterEvents()` 追加映射



### 9.5 新增队列任务

1. `app/jobs/xxx.go` — 实现 `Signature()` + `Handle()`
2. 调度：`facades.Queue().Job(&jobs.Xxx{}, args).Dispatch()`



### 9.6 升级框架

```bash
go get github.com/zhoudm1743/go-fast-framework@vX.Y.Z
go mod tidy && go build ./...
```

同步检查：配置默认值、Provider 列表、示例代码是否与新版本 API 一致。

---



## 10. 代码风格与质量

- **最小改动**：只改任务相关文件，不顺手重构
- **复用现有模式**：服务参照 `services/test/`，控制器参照 `UserController`
- **接口优先**：业务服务先写 `services/contracts` 接口
- **Singleton 默认**：全局共享用 `Singleton()`；请求级有状态才用 `Bind()`
- **注释**：只解释非显而易见的业务逻辑
- **测试**：服务层写 `_test.go`；改完必跑 `go build ./...` + `go test ./...`
- **不提交**：`storage/`、`*.db`、含真实密钥的配置

---



## 11. 验证命令

```bash
go build ./...                         # 全项目编译
go test ./...                          # 全项目测试
go test ./services/test/...            # 单包测试

go run .                               # 启动 HTTP（:3000）
curl http://localhost:3000/health       # {"status":"ok"}
curl http://localhost:3000/api/ping      # 示例服务响应

go run . fast list                      # 控制台命令列表
go run . fast example --name GoFast
go run . fast migrate                   # 数据库迁移
go run . fast db:seed                   # 数据填充
```

---



## 12. 常见问题排查


| 现象                    | 可能原因                     | 处理                                            |
| --------------------- | ------------------------ | --------------------------------------------- |
| 启动报 Redis 连接失败        | `cache.redis.host` 非空    | 留空 host 或启动 Redis                             |
| `MustMake panic`      | Provider 顺序错误 / 未注册      | 检查 `bootstrap/app.go` providers 列表            |
| Register 阶段 panic     | Register 中调用了 MustMake   | 移到 Boot 阶段                                    |
| Facade 返回 nil / panic | `facades.SetApp()` 之前调用  | 确保在 Boot 之后                                   |
| 路由 404                | Group 前缀 + Prefix() 拼接错误 | 检查 routes 与 Controller.Prefix()               |
| 配置不生效                 | 未空白导入 config 包           | 确认 `_ "github.com/zhoudm1743/go-fast/config"` |
| YAML 未覆盖 Go 默认值       | 键路径不一致                   | 对照 `config/*.go` 命名空间与嵌套结构                    |


---



## 13. 参考实现（复制此模式）

新增自定义服务时，直接对照：

```
services/contracts/test.go              # 业务契约
services/test/test.go                   # 服务实现
services/test/service_provider.go       # Provider（Register + Boot）
services/test/service_provider_test.go  # 测试
config/test.go                          # 配置默认值
bootstrap/app.go                        # providers() 末尾注册
routes/app.go                           # facades.App().MustMake("test") 用法
```

新增 HTTP 控制器时，对照：

```
app/http/app/controllers/user_controller.go
app/http/app/middleware/auth.go
routes/app.go
```

---



## 14. 文档索引


| 场景              | 文档                                                                                                                   |
| --------------- | -------------------------------------------------------------------------------------------------------------------- |
| **文档索引 & 架构**   | [docs/README.md](docs/README.md)                                                                                     |
| 首次了解            | [docs/getting-started.md](docs/getting-started.md)                                                                   |
| 配置完整参考          | [docs/configuration.md](docs/configuration.md)                                                                       |
| ServiceProvider | [docs/service-provider.md](docs/service-provider.md)                                                                 |
| Facade API      | [docs/facade.md](docs/facade.md)                                                                                     |
| 控制器             | [docs/controller.md](docs/controller.md)                                                                             |
| 路由              | [docs/route.md](docs/route.md)                                                                                       |
| 数据库             | [docs/database/README.md](docs/database/README.md)                                                                   |
| 队列 / 事件 / 调度    | [docs/queue.md](docs/queue.md) / [docs/event.md](docs/event.md) / [docs/task-scheduling.md](docs/task-scheduling.md) |
| Fast 控制台        | [docs/fast.md](docs/fast.md)                                                                                         |
| 容器 API          | [docs/container.md](docs/container.md)                                                                               |
| 插件              | [docs/plugins.md](docs/plugins.md)                                                                                   |



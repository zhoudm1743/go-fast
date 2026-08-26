# GoFast 文档索引

> **请先读本节。** GoFast 由两个独立仓库组成，文档中的路径与 import 必须区分清楚。

---

## 架构：两个仓库

| 仓库 | Module | 职责 | 文档范围 |
|------|--------|------|---------|
| **go-fast**（本仓库） | `github.com/zhoudm1743/go-fast` | 应用骨架：业务代码、路由、配置、自定义 Provider | 本 `docs/` 目录 |
| **go-fast-framework** | `github.com/zhoudm1743/go-fast-framework` | 框架内核：IoC、Facade、日志、DB、缓存、HTTP 等 | [框架 README](https://github.com/zhoudm1743/go-fast-framework) |

```
go get github.com/zhoudm1743/go-fast-framework@v0.7.10   # 框架（go.mod 依赖）
git clone github.com/zhoudm1743/go-fast                    # 应用骨架（本仓库）
```

**import 规范：**

```go
// 框架 API — 来自 go-fast-framework module（不在本仓库源码树内）
import (
    "github.com/zhoudm1743/go-fast-framework/contracts"
    "github.com/zhoudm1743/go-fast-framework/facades"
    "github.com/zhoudm1743/go-fast-framework/foundation"
)

// 本仓库业务代码
import (
    "github.com/zhoudm1743/go-fast/app/models"
    "github.com/zhoudm1743/go-fast/services/contracts"  // 业务契约，非框架 contracts
)
```

**常见误解：**

| ❌ 错误理解 | ✅ 正确理解 |
|-----------|-----------|
| 在本仓库新建 `facades/`、`contracts/` 包 | 框架 Facade/契约在 **go-fast-framework**；业务契约放 `services/contracts/` |
| 修改 `go-fast-framework/facades/*.go` | 框架源码不在本仓库；自定义服务用 `facades.App().MustMake("key")` |
| `config/app.go` 存在 | 应用名在 `config/server.go` 的 `server.name` |
| 独立的 `validation.ServiceProvider` | 验证器由 `http.ServiceProvider` 一并注册（key: `validator`） |

---

## 本仓库目录（业务层）

```
go-fast/
├── app/          # 控制器、模型、事件、任务、命令
├── services/     # 自定义业务服务 + ServiceProvider
├── bootstrap/    # Provider 列表、Boot 流程
├── config/       # Go 默认值 + config.yaml
├── routes/       # 路由编排
└── docs/         # 本文档
```

AI 协作请参阅根目录 [AGENT.md](../AGENT.md)。

---

## 文档导航

### 入门

| 文档 | 说明 |
|------|------|
| [快速开始](getting-started.md) | 环境、配置、启动、路由 |
| [配置说明](configuration.md) | Go 配置 + YAML 完整参考 |
| [AGENT.md](../AGENT.md) | AI Agent 协作指南 |

### 核心机制

| 文档 | 说明 |
|------|------|
| [ServiceProvider](service-provider.md) | 自定义服务注册 |
| [容器 API](container.md) | Bind / Singleton / Make |
| [Facade 使用说明](facade.md) | 框架 Facade API（**module 内**，非本仓库目录） |
| [路由设计](route.md) | 控制器自注册、路由组 |
| [控制器指南](controller.md) | HTTP 控制器完整示例 |

### 基础设施

| 文档 | 说明 |
|------|------|
| [数据库](database/README.md) | 查询、事务、多连接 |
| [文件存储](storage.md) | Storage 多磁盘 |
| [队列](queue.md) | 异步任务 |
| [事件](event.md) | 事件与监听器 |
| [任务调度](task-scheduling.md) | Cron 定时任务 |
| [Fast 控制台](fast.md) | CLI 与脚手架 |

### 扩展

| 文档 | 说明 |
|------|------|
| [插件开发](plugins.md) | 独立 Go module 插件 |

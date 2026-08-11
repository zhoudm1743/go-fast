# GoFast

> GoFast -- 一个轻量、可扩展的 Go 语言快速开发框架。整合 IoC 容器、ServiceProvider 生命周期、Facade 门面、Go 配置文件、结构化日志、GORM 数据库、验证器和可插拔文件存储等企业级能力，帮助团队以最少样板代码搭建可靠的后端服务。

[![Go Version](https://img.shields.io/badge/Go-1.25+-blue.svg)](https://go.dev/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

---

## 特性

- **IoC 服务容器** -- Bind / Singleton / Instance，支持延迟加载
- **ServiceProvider 机制** -- Register + Boot 两阶段，声明式管理服务生命周期
- **Facade 门面** -- 一行代码访问任意服务
- **配置管理** -- 支持 Go 配置文件 + YAML，点号路径访问，优先级可控
- **日志系统** -- 基于 Zap，支持控制台/文件/混合输出，结构化字段，文件轮转
- **数据库** -- 基于 GORM，支持 MySQL / PostgreSQL / SQLite / SQL Server，内置时序 ID 主键，多连接管理
- **缓存系统** -- 多 Store、标签分组、原子操作、Hash
- **文件存储** -- 多磁盘管理，内置本地驱动，可扩展 OSS/COS/MinIO/S3
- **验证器** -- 基于 go-playground/validator，结构体 tag 声明式验证
- **HTTP 路由** -- 支持 Gin / Fiber 双引擎，链式注册、路由组、中间件
- **Fast 控制台** -- 内置脚手架命令（make:model / make:controller / make:provider 等），支持自定义命令、交互式输入
- **优雅关闭** -- 信号监听，按逆序释放资源
- **插件化** -- 任何 Go module 均可作为插件接入

---

## 快速开始

```bash
git clone https://github.com/zhoudm1743/go-fast.git
cd GoFast
go run main.go
```

---

## 文档

| 文档 | 说明 |
|------|------|
| [快速开始](docs/getting-started.md) | 环境要求、配置、启动、路由注册、模型定义 |
| [配置说明](docs/configuration.md) | Go 配置文件与 YAML 的完整参考，所有配置项一览 |
| [控制器开发指南](docs/controller.md) | 控制器、请求验证、数据库、中间件完整示例 |
| [Fast 控制台](docs/fast.md) | 脚手架命令、自定义命令、交互式输入 |
| [容器 API](docs/container.md) | Bind / Singleton / Instance / Make 完整接口 |
| [Facade 使用说明](docs/facade.md) | Config / Log / Cache / DB / Route / Storage / Validator |
| [编写自定义 Provider](docs/service-provider.md) | ServiceProvider 接口、延迟加载、关闭钩子 |
| [插件开发指南](docs/plugins.md) | 独立 module 插件的开发、发布与接入规范 |

---

## 项目结构

```
GoFast/
├── app/                         # 业务代码
│   ├── console/
│   │   └── commands/            # 自定义 Fast 命令
│   ├── http/
│   │   ├── admin/
│   │   │   ├── controllers/     # 后台控制器
│   │   │   ├── middleware/      # 后台中间件
│   │   │   └── requests/        # 后台请求结构体
│   │   └── app/
│   │       ├── controllers/     # 前台控制器
│   │       ├── middleware/      # 前台中间件
│   │       └── requests/        # 前台请求结构体
│   ├── models/                  # 数据模型
│   ├── providers/               # 自定义 ServiceProvider
│   └── rules/                   # 自定义验证规则
├── bootstrap/
│   ├── app.go                   # 应用引导 & Provider 列表
│   └── commands.go              # Fast 命令注册入口
├── config/
│   ├── config.yaml              # YAML 配置文件（覆盖 Go 默认值）
│   ├── app.go                   # 应用配置默认值
│   ├── cache.go                 # 缓存配置默认值
│   ├── database.go              # 数据库配置默认值
│   ├── filesystem.go            # 文件存储配置默认值
│   ├── jwt.go                   # JWT 配置默认值
│   ├── log.go                   # 日志配置默认值
│   ├── server.go                # HTTP 服务配置默认值
│   ├── session.go               # Session 配置默认值
│   └── view.go                  # 视图引擎配置默认值
├── database/
│   └── migrations/              # 数据库迁移
├── routes/
│   ├── api.go                   # 路由统一入口
│   ├── app.go                   # 前台路由注册
│   └── admin.go                 # 后台路由注册
├── framework/                   # 框架核心
│   ├── foundation/              # IoC 容器 & Application
│   ├── contracts/               # 服务接口定义
│   ├── facades/                 # 静态门面
│   ├── config/                  # 配置服务
│   ├── log/                     # 日志服务
│   ├── cache/                   # 缓存服务
│   ├── database/                # 数据库服务
│   ├── filesystem/              # 文件存储服务
│   ├── http/                    # HTTP 路由服务
│   ├── jwt/                     # JWT 鉴权服务
│   ├── event/                   # 事件系统
│   ├── queue/                   # 队列系统
│   ├── schedule/                # 任务调度
│   ├── fast/                    # Fast 控制台
│   ├── id/                      # 时序 ID 生成
│   └── utils/                   # 工具函数
├── docs/                        # 文档
├── main.go                      # 入口
└── go.mod
```

---

## License

[MIT](LICENSE)

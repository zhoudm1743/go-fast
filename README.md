# GoFast

> GoFast — 一个轻量、可扩展且面向生产的 Go 语言快速开发框架。它整合了 IoC 服务容器、ServiceProvider 生命周期、Facade 门面、配置管理、结构化日志、GORM ORM、验证器和可插拔的文件存储等常用企业级功能，帮助团队以最少样板代码快速搭建可靠、可维护的后端服务。

[![Go Version](https://img.shields.io/badge/Go-1.25+-blue.svg)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

---

## ✨ 特性

- **IoC 服务容器** — Bind / Singleton / Instance，支持延迟加载
- **ServiceProvider 机制** — Register + Boot 两阶段，声明式管理服务生命周期
- **Facade 门面** — 一行代码访问任意服务，零样板代码
- **配置管理** — 基于 Viper，支持 YAML + 环境变量 + 点号路径
- **日志系统** — 基于 Logrus，支持结构化字段、文件轮转、多输出
- **ORM 数据库** — 基于 GORM，支持 MySQL / PostgreSQL / SQLite / SQL Server，UUID v7 主键
- **缓存系统** — 多 Store、标签分组、原子操作、Hash、分布式锁
- **文件存储** — 多磁盘管理，内置本地驱动，可扩展云存储
- **验证器** — 基于 go-playground/validator，结构体 tag 声明式验证
- **HTTP 路由** — 基于 Fiber v2，链式注册、路由组、中间件
- **gRPC 服务** — 内置 gRPC Server，支持 Unary/Stream 拦截器、Keepalive、Server Reflection，与 HTTP 同进程并行运行
- **Fast 控制台** — 内置脚手架命令（make:model / make:controller / make:provider / make:validator / make:utils），支持自定义命令、交互式输入、彩色输出
- **优雅关闭** — 信号监听，按逆序释放资源（HTTP + gRPC 均覆盖）
- **插件化** — 任何 Go module 均可作为插件接入

---

## 🚀 快速开始

```bash
git clone https://github.com/zhoudm1743/go-fast.git
cd GoFast
# 创建 config.yaml（参考 docs/getting-started.md）
go run main.go
```

---

## 📖 文档

| 文档 | 说明 |
|------|------|
| [快速开始](docs/getting-started.md) | 环境要求、配置、启动、路由注册、模型定义 |
| [控制器开发指南](docs/controller.md) | 控制器、请求验证、数据库、中间件完整示例 |
| [gRPC 使用指南](docs/grpc.md) | gRPC 服务定义、代码生成、拦截器、grpcurl 调试 |
| [Fast 控制台](docs/fast.md) | 脚手架命令、自定义命令、交互式输入、注册 |
| [容器 API](docs/container.md) | Bind / Singleton / Instance / Make 完整接口 |
| [Facade 使用说明](docs/facade.md) | Config / Log / Cache / Orm / Route / Storage / Validator / GRPC |
| [编写自定义 Provider](docs/service-provider.md) | ServiceProvider 接口、延迟加载、关闭钩子 |
| [插件开发指南](docs/plugins.md) | 独立 module 插件的开发、发布与接入规范 |

---

## 🏗 项目结构

```
GoFast/
├── app/                         # 业务代码（用户编写）
│   ├── console/
│   │   └── commands/            # 自定义 Fast 命令
│   ├── http/
│   │   ├── admin/
│   │   │   ├── controllers/     # 后台管理控制器
│   │   │   ├── middleware/      # 后台鉴权中间件
│   │   │   └── requests/        # 后台请求结构体
│   │   └── app/
│   │       ├── controllers/     # 前台/用户端控制器
│   │       ├── middleware/      # 前台鉴权中间件
│   │       └── requests/        # 前台请求结构体
│   ├── models/                  # 数据模型
│   ├── providers/               # 自定义 ServiceProvider
│   └── rules/                   # 自定义验证规则
├── bootstrap/
│   ├── app.go                   # 应用引导 & Provider 列表
│   └── commands.go              # Fast 命令注册入口
├── config/
│   └── config.yaml              # 配置文件
├── database/
│   └── migrations/              # 数据库迁移
├── routes/
│   ├── api.go                   # 路由统一入口
│   ├── app.go                   # 前台路由注册
│   ├── admin.go                 # 后台路由注册
│   └── grpc.go                  # gRPC 服务注册
├── framework/                   # 框架核心（不建议修改）
│   ├── foundation/              # IoC 容器 & Application
│   ├── contracts/               # 服务接口定义
│   ├── facades/                 # 静态门面
│   ├── config/                  # 配置服务
│   ├── log/                     # 日志服务
│   ├── cache/                   # 缓存服务
│   ├── database/                # ORM 数据库服务
│   ├── filesystem/              # 文件存储服务
│   ├── http/                    # HTTP 路由服务
│   ├── gRPC/                    # gRPC 服务器（拦截器、Server 封装、ServiceProvider）
│   ├── fast/                 # Fast 控制台（内核、脚手架命令）
│   ├── validation/              # 验证服务
│   └── utils/                 # 工具函数（StringUtil / FileUtil / ToolsUtil …）
├── docs/                        # 文档
├── main.go                      # 入口
└── go.mod
```

---

## 🤝 参与贡献

1. Fork 本仓库
2. 新建 `feat/xxx` 分支
3. 提交代码
4. 新建 Pull Request

---

## 📄 License

[MIT](LICENSE)

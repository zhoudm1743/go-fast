# GoFast gRPC 使用指南

> 版本：v0.1.0  
> GoFast gRPC 模块让你可以在同一进程内同时运行 HTTP 和 gRPC 服务，遵循与 HTTP 层完全一致的 ServiceProvider / Facade / 契约设计模式。

---

## 目录

1. [快速概览](#1-快速概览)
2. [配置](#2-配置)
3. [定义 Proto 文件](#3-定义-proto-文件)
4. [生成 Go 代码](#4-生成-go-代码)
5. [实现 gRPC 服务](#5-实现-grpc-服务)
6. [注册路由](#6-注册路由)
7. [启动服务](#7-启动服务)
8. [使用 Facade 访问 gRPC Server](#8-使用-facade-访问-grpc-server)
9. [内置拦截器](#9-内置拦截器)
10. [调试：grpcurl](#10-调试grpcurl)
11. [错误处理规范](#11-错误处理规范)
12. [与 HTTP 层对比](#12-与-http-层对比)

---

## 1. 快速概览

GoFast gRPC 遵循框架核心理念 — **契约优先、配置驱动、单进程共存**：

```
请求 ──→ gRPC Server (:9000)
              │
         拦截链（Recovery → Logging）
              │
         UserServiceServer.GetUser()
              │
         facades.DB().Query()...   # 与 HTTP 控制器共用同一个数据库服务
              │
         UserReply
```

gRPC Server 与 HTTP Server 在同一进程内以独立协程运行，共享所有 GoFast 内置服务（DB、Log、Cache 等）。

---

## 2. 配置

在 `config/config.yaml` 中的 `grpc` 配置节：

```yaml
# ── gRPC ─────────────────────────────────────────────────────────────
grpc:
  host: 0.0.0.0
  port: 9000
  mode: debug                          # debug（开启 reflection）/ release
  max_recv_msg_size_mb: 4
  max_send_msg_size_mb: 4
  max_conn_age_sec: 300                # 单个连接最大存活时间（秒）
  max_conn_age_grace_sec: 5            # 超过 max_conn_age 后的宽限期（秒）
  keepalive_time_sec: 60               # Keepalive ping 间隔
  keepalive_timeout_sec: 20            # Keepalive 超时

  # TLS（可选，留空则不启用）
  tls:
    cert_file: ""
    key_file: ""
```

| 配置键 | 默认值 | 说明 |
|--------|--------|------|
| `grpc.host` | `0.0.0.0` | 监听地址 |
| `grpc.port` | `9000` | 监听端口 |
| `grpc.mode` | `debug` | `debug` 开启 gRPC Server Reflection；`release` 关闭 |
| `grpc.max_recv_msg_size_mb` | `4` | 单条请求最大字节（MB） |
| `grpc.max_send_msg_size_mb` | `4` | 单条响应最大字节（MB） |
| `grpc.max_conn_age_sec` | `300` | 连接最长存活时间（秒） |

---

## 3. 定义 Proto 文件

在 `app/grpc/proto/<service>/` 目录下创建 `.proto` 文件，以 `UserService` 为例：

**文件：`app/grpc/proto/user/user.proto`**

```protobuf
syntax = "proto3";

package user;

option go_package = "github.com/zhoudm1743/go-fast/app/grpc/proto/user;userpb";

service UserService {
    rpc GetUser(GetUserRequest) returns (UserReply);
    rpc ListUsers(ListUsersRequest) returns (ListUsersReply);
    rpc CreateUser(CreateUserRequest) returns (UserReply);
    rpc UpdateUser(UpdateUserRequest) returns (UserReply);
    rpc DeleteUser(DeleteUserRequest) returns (DeleteUserReply);
}

message GetUserRequest  { string id = 1; }
message ListUsersRequest { int32 page = 1; int32 size = 2; string email = 3; }
message CreateUserRequest { string name = 1; string email = 2; string password = 3; }
message UpdateUserRequest { string id = 1; string name = 2; string email = 3; }
message DeleteUserRequest { string id = 1; }

message UserReply {
    string id         = 1;
    string name       = 2;
    string email      = 3;
    string created_at = 4;
    string updated_at = 5;
}

message ListUsersReply {
    repeated UserReply users = 1;
    int64 total              = 2;
    int32 page               = 3;
    int32 size               = 4;
}

message DeleteUserReply { bool success = 1; }
```

> **包命名约定**：`go_package` 末尾的 `;userpb` 指定 Go 包短名，避免与业务包名冲突。

---

## 4. 生成 Go 代码

### 4.1 安装工具

```bash
# 安装 Go 代码生成插件（仅首次需要）
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

`protoc` 编译器本身需单独安装：
- **Windows**：从 [GitHub Releases](https://github.com/protocolbuffers/protobuf/releases) 下载 `protoc-xx-win64.zip`，解压后将 `bin/protoc.exe` 加入 PATH
- **macOS**：`brew install protobuf`
- **Linux**：`apt install protobuf-compiler` 或从 Releases 下载

### 4.2 生成命令

在项目根目录执行：

```bash
protoc \
  --go_out=. \
  --go_opt=paths=source_relative \
  --go-grpc_out=. \
  --go-grpc_opt=paths=source_relative \
  app/grpc/proto/user/user.proto
```

生成后目录结构：

```
app/grpc/proto/user/
├── user.proto          # 源文件（手动维护）
├── user.pb.go          # 消息类型（自动生成，勿手动修改）
└── user_grpc.pb.go     # 服务客户端/服务端接口（自动生成，勿手动修改）
```

---

## 5. 实现 gRPC 服务

在 `app/grpc/services/` 下创建服务实现文件。

**文件：`app/grpc/services/user_service.go`**

```go
package services

import (
    "context"
    "errors"
    "strconv"

    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"

    "github.com/zhoudm1743/go-fast/app/models"
    userpb "github.com/zhoudm1743/go-fast/app/grpc/proto/user"
    "github.com/zhoudm1743/go-fast/framework/contracts"
    "github.com/zhoudm1743/go-fast/framework/facades"
)

// UserServiceServer 嵌入 UnimplementedUserServiceServer 保证前向兼容。
type UserServiceServer struct {
    userpb.UnimplementedUserServiceServer
}

func (s *UserServiceServer) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.UserReply, error) {
    if req.Id == "" {
        return nil, status.Error(codes.InvalidArgument, "id 不能为空")
    }
    var user models.User
    if err := facades.DB().Query().Where("id = ?", req.Id).First(&user); err != nil {
        if errors.Is(err, contracts.ErrRecordNotFound) {
            return nil, status.Errorf(codes.NotFound, "用户 %s 不存在", req.Id)
        }
        facades.Log().Errorf("grpc get user: %v", err)
        return nil, status.Error(codes.Internal, "查询失败")
    }
    return toUserReply(&user), nil
}

func (s *UserServiceServer) ListUsers(ctx context.Context, req *userpb.ListUsersRequest) (*userpb.ListUsersReply, error) {
    page, size := int(req.Page), int(req.Size)
    if page <= 0 { page = 1 }
    if size <= 0 || size > 100 { size = 20 }

    q := facades.DB().Query().Model(&models.User{}).Order("created_at DESC")
    if req.Email != "" {
        q = q.Where("email LIKE ?", "%"+req.Email+"%")
    }
    var total int64
    q.Count(&total)

    var users []models.User
    if err := q.Paginate(page, size).Find(&users); err != nil {
        return nil, status.Error(codes.Internal, "查询失败")
    }
    replies := make([]*userpb.UserReply, 0, len(users))
    for i := range users { replies = append(replies, toUserReply(&users[i])) }
    return &userpb.ListUsersReply{Users: replies, Total: total, Page: int32(page), Size: int32(size)}, nil
}

func (s *UserServiceServer) CreateUser(ctx context.Context, req *userpb.CreateUserRequest) (*userpb.UserReply, error) {
    if req.Name == "" || req.Email == "" || req.Password == "" {
        return nil, status.Error(codes.InvalidArgument, "name、email、password 均不能为空")
    }
    user := &models.User{Name: req.Name, Email: req.Email, Password: req.Password}
    if err := facades.DB().Query().Create(user); err != nil {
        if errors.Is(err, contracts.ErrDuplicatedKey) {
            return nil, status.Error(codes.AlreadyExists, "邮箱已被注册")
        }
        return nil, status.Error(codes.Internal, "创建失败")
    }
    return toUserReply(user), nil
}

func (s *UserServiceServer) UpdateUser(ctx context.Context, req *userpb.UpdateUserRequest) (*userpb.UserReply, error) {
    if req.Id == "" {
        return nil, status.Error(codes.InvalidArgument, "id 不能为空")
    }
    var user models.User
    if err := facades.DB().Query().Where("id = ?", req.Id).First(&user); err != nil {
        if errors.Is(err, contracts.ErrRecordNotFound) {
            return nil, status.Errorf(codes.NotFound, "用户 %s 不存在", req.Id)
        }
        return nil, status.Error(codes.Internal, "查询失败")
    }
    updates := map[string]any{}
    if req.Name != "" { updates["name"] = req.Name }
    if req.Email != "" { updates["email"] = req.Email }
    if len(updates) > 0 {
        if err := facades.DB().Query().Model(&user).Updates(updates); err != nil {
            return nil, status.Error(codes.Internal, "更新失败")
        }
    }
    return toUserReply(&user), nil
}

func (s *UserServiceServer) DeleteUser(ctx context.Context, req *userpb.DeleteUserRequest) (*userpb.DeleteUserReply, error) {
    if req.Id == "" {
        return nil, status.Error(codes.InvalidArgument, "id 不能为空")
    }
    if err := facades.DB().Query().Where("id = ?", req.Id).Delete(&models.User{}); err != nil {
        if errors.Is(err, contracts.ErrRecordNotFound) {
            return nil, status.Errorf(codes.NotFound, "用户 %s 不存在", req.Id)
        }
        return nil, status.Error(codes.Internal, "删除失败")
    }
    return &userpb.DeleteUserReply{Success: true}, nil
}

func toUserReply(u *models.User) *userpb.UserReply {
    return &userpb.UserReply{
        Id:        u.ID,
        Name:      u.Name,
        Email:     u.Email,
        CreatedAt: strconv.FormatInt(u.CreatedAt, 10),
        UpdatedAt: strconv.FormatInt(u.UpdatedAt, 10),
    }
}
```

**关键规范：**

- 嵌入 `UnimplementedUserServiceServer` — 新增 RPC 方法时旧客户端不报错
- 参数校验使用 `status.Error(codes.InvalidArgument, ...)` — 对应 HTTP 400
- 记录不存在使用 `codes.NotFound` — 对应 HTTP 404
- 内部错误用 `codes.Internal` — **不要**将底层错误信息暴露给客户端
- 数据库错误用 `errors.Is(err, contracts.ErrRecordNotFound)` 判断

---

## 6. 注册路由

**文件：`routes/grpc.go`**

```go
package routes

import (
    userpb "github.com/zhoudm1743/go-fast/app/grpc/proto/user"
    "github.com/zhoudm1743/go-fast/app/grpc/services"
    "github.com/zhoudm1743/go-fast/framework/facades"
)

// RegisterGRPC 向 gRPC Server 注册所有服务。
func RegisterGRPC() {
    grpcServer := facades.GRPC()
    grpcServer.RegisterService(
        &userpb.UserService_ServiceDesc,
        &services.UserServiceServer{},
    )
    // 追加更多服务：
    // grpcServer.RegisterService(&orderpb.OrderService_ServiceDesc, &services.OrderServiceServer{})
}
```

---

## 7. 启动服务

`main.go` 中 HTTP 和 gRPC 服务器并行启动：

```go
func main() {
    app := bootstrap.Boot()

    routes.Register()      // 注册 HTTP 路由
    routes.RegisterGRPC()  // 注册 gRPC 服务

    // HTTP 服务器（非阻塞）
    go func() {
        if err := facades.Route().Run(); err != nil {
            facades.Log().Errorf("http server error: %v", err)
        }
    }()

    // gRPC 服务器（非阻塞）
    go func() {
        if err := facades.GRPC().Run(); err != nil {
            facades.Log().Errorf("grpc server error: %v", err)
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    app.Shutdown() // 同时优雅关闭 HTTP + gRPC
}
```

启动后终端将输出：

```
[GoFast] v0.1.0 booted
[GoFast/grpc] server listening on 0.0.0.0:9000
```

---

## 8. 使用 Facade 访问 gRPC Server

```go
import "github.com/zhoudm1743/go-fast/framework/facades"

// 向 gRPC Server 注册服务（通常在 routes/grpc.go 中调用）
facades.GRPC().RegisterService(&userpb.UserService_ServiceDesc, &services.UserServiceServer{})

// 启动监听（通常由 main.go 调用）
facades.GRPC().Run("0.0.0.0:9000")

// 优雅关闭（通常由 app.Shutdown 自动触发）
facades.GRPC().Shutdown()

// 获取底层 *grpc.Server（高级场景：注册健康检查、gRPC Gateway 等）
rawSrv := facades.GRPC().RawServer()
```

---

## 9. 内置拦截器

gRPC Server 自动挂载以下拦截器（Unary + Stream 各一套），**无需手动配置**：

| 拦截器 | 作用 |
|--------|------|
| `RecoveryInterceptor` | 捕获业务代码 panic，记录堆栈，返回 `codes.Internal` |
| `LoggingInterceptor` | 记录每次 RPC 调用的方法名和耗时；出错时附带错误详情 |

日志输出示例：

```
[GoFast/grpc] /user.UserService/GetUser | 1.2ms
[GoFast/grpc] /user.UserService/CreateUser | 3.5ms | rpc error: code = InvalidArgument desc = ...
[GoFast/grpc] stream /user.UserService/Watch | 2.1s | err=<nil>
```

### 自定义拦截器

在 `framework/gRPC/server.go` 的 `NewServer` 函数中追加到拦截链：

```go
grpc.ChainUnaryInterceptor(
    RecoveryInterceptor(log),
    LoggingInterceptor(log),
    AuthInterceptor(log),   // 自定义 JWT 鉴权拦截器
),
```

---

## 10. 调试：grpcurl

`debug` 模式下 gRPC Server Reflection 自动开启，可使用 [grpcurl](https://github.com/fullstorydev/grpcurl) 进行接口调试：

```bash
# 安装
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# 列出所有服务
grpcurl -plaintext localhost:9000 list

# 列出 UserService 的所有方法
grpcurl -plaintext localhost:9000 list user.UserService

# 调用 GetUser
grpcurl -plaintext -d '{"id": "your-user-id"}' \
  localhost:9000 user.UserService/GetUser

# 分页查询用户
grpcurl -plaintext -d '{"page": 1, "size": 10}' \
  localhost:9000 user.UserService/ListUsers

# 邮箱模糊搜索
grpcurl -plaintext -d '{"page": 1, "size": 10, "email": "alice"}' \
  localhost:9000 user.UserService/ListUsers

# 创建用户
grpcurl -plaintext \
  -d '{"name":"Alice","email":"alice@example.com","password":"secret"}' \
  localhost:9000 user.UserService/CreateUser

# 更新用户
grpcurl -plaintext \
  -d '{"id":"<user-id>","name":"Bob"}' \
  localhost:9000 user.UserService/UpdateUser

# 删除用户
grpcurl -plaintext -d '{"id":"<user-id>"}' \
  localhost:9000 user.UserService/DeleteUser
```

> **安全提示**：生产环境请将 `grpc.mode` 设置为 `release` 关闭 Reflection，避免接口信息泄露。

---

## 11. 错误处理规范

gRPC 使用 `status` 包的错误码替代 HTTP 状态码：

| gRPC Code | 含义 | 对应 HTTP |
|-----------|------|-----------|
| `codes.OK` | 成功 | 200 |
| `codes.InvalidArgument` | 参数错误 | 400 |
| `codes.NotFound` | 资源不存在 | 404 |
| `codes.AlreadyExists` | 资源已存在 | 409 |
| `codes.PermissionDenied` | 无权限 | 403 |
| `codes.Unauthenticated` | 未认证 | 401 |
| `codes.Internal` | 服务器内部错误 | 500 |
| `codes.Unavailable` | 服务不可用 | 503 |

返回错误的标准写法：

```go
// 参数错误
return nil, status.Error(codes.InvalidArgument, "id 不能为空")

// 带格式化信息
return nil, status.Errorf(codes.NotFound, "用户 %s 不存在", req.Id)

// 内部错误（记录详情，不暴露给客户端）
facades.Log().Errorf("db error: %v", err)
return nil, status.Error(codes.Internal, "查询失败")
```

框架数据库错误到 gRPC Code 的映射：

```go
errors.Is(err, contracts.ErrRecordNotFound) → codes.NotFound
errors.Is(err, contracts.ErrDuplicatedKey)  → codes.AlreadyExists
```

---

## 12. 与 HTTP 层对比

| 特性 | HTTP（Fiber） | gRPC |
|------|-------------|------|
| 服务契约 | `contracts/http.go → Route` | `contracts/grpc.go → GRPCServer` |
| 服务提供者 | `framework/http/service_provider.go` | `framework/gRPC/service_provider.go` |
| Facade | `facades.Route()` | `facades.GRPC()` |
| 路由注册 | `routes/admin.go` `routes/app.go` | `routes/grpc.go` |
| 中间件/拦截器 | `contracts.HandlerFunc` 链 | Unary + Stream Interceptor |
| 控制器/服务 | `Controller` 接口 + `Boot(Route)` | `XxxServiceServer`（protoc 生成接口） |
| 配置节 | `server:` | `grpc:` |
| 优雅关闭 | `app.OnShutdown → Route.Shutdown()` | `app.OnShutdown → GRPCServer.Shutdown()` |
| 端口 | `3000`（默认） | `9000`（默认） |
| 调试工具 | Browser / curl / Postman | grpcurl |

---

## 附：添加新的 gRPC 服务

以添加 `OrderService` 为例，只需四步：

**1. 创建 proto 文件**
```
app/grpc/proto/order/order.proto
```

**2. 生成代码**
```bash
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       app/grpc/proto/order/order.proto
```

**3. 实现服务**
```
app/grpc/services/order_service.go
```

**4. 注册到 `routes/grpc.go`**
```go
grpcServer.RegisterService(
    &orderpb.OrderService_ServiceDesc,
    &services.OrderServiceServer{},
)
```

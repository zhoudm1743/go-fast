# v0.4.0 Release Notes

> 发布日期：2026-08-07

## 新特性

### JWT：Guard 多守卫支持 (`feat(jwt)`)

支持声明多个命名 JWT Guard，每个 Guard 独立配置 secret / ttl / alg，适用于多端认证场景（如用户端、管理后台、开放平台）。

- 拆分 `contracts.JWT` 为 `JWTDriver`（核心加解密能力）+ `JWT`（管理器接口）
- 新增 `jwtManager` 实现，通过 `Guard(name)` 方法切换命名守卫
- 命名守卫从 `jwt.guards.<name>.*` 独立读取配置，懒加载并缓存
- 未配置 guards 时使用顶层 `jwt.*` 作为默认守卫，完全向后兼容
- 新增 9 个单元测试覆盖

```yaml
# config.yaml - 多 Guard 配置示例
jwt:
  secret: "default-secret"
  ttl: 60
  alg: HS256
  guards:
    admin:
      secret: "admin-guard-secret"
      ttl: 120
      alg: HS512
    api:
      secret: "api-guard-secret"
      ttl: 30
      alg: HS256
```

```go
// 使用命名 Guard
token, _ := facades.JWT.Guard("admin").Login(uid, payload)
claims, _ := facades.JWT.Guard("api").Parse(token)
```

### 配置：Go 配置文件支持 (`feat(config)`)

除 YAML 配置文件外，新增 Go 源码文件方式注册配置。Go 配置文件通过 `init()` 自动注册，类型安全，IDE 友好，适合复杂默认值和计算型配置。

- `contracts.Config` 新增 `Add(namespace, key, value)` 方法
- Go 配置文件放在 `config/` 目录，通过 `init()` 注册到容器
- 内置 `config/server.go`、`config/database.go`、`config/log.go`、`config/jwt.go`、`config/cache.go`、`config/filesystem.go`、`config/session.go`、`config/view.go` 等命名空间配置
- `bootstrap/app.go` 通过空导入触发 Go 配置文件 `init()` 注册
- 优先级：显式调用 > Go 文件默认值 > YAML 文件值
- 新增 19 个单元测试（含并发安全测试）

```go
// config/database.go - Go 配置文件示例
func init() {
    config.Add("database", "driver", "sqlite")
    config.Add("database", "max_idle_conns", 10)
    config.Add("database", "max_open_conns", 100)
}
```

## 重构

### 移除 gRPC 模块 (`refactor(grpc)`)

- 删除 `config/grpc.go`、`framework/contracts/grpc.go`、`framework/facades/grpc.go`
- 移除 gRPC 相关文档 `docs/grpc.md`
- 更新 README 和配置文档，清理所有 gRPC 引用
- 简化框架结构，减少不必要的依赖

### 移除 ORM 接口 (`refactor(config)`)

- 删除旧 `contracts.ORM` 接口及 `framework/database/orm.go` 实现
- 删除 `facades.ORM` 门面
- 数据库连接管理迁移至统一的 DB 接口
- 新增多数据库连接配置支持
- `framework/foundation/provider.go` 简化，移除 ORM 相关注册逻辑

## 改进

### 日志模块增强

- **调用栈追踪优化**：`callerFields` 改用 `runtime.Callers` 遍历帧栈，自动跳过 `logger.go` 内部包装帧，不再依赖硬编码的 skip 偏移量，更健壮
- **控制台格式重构**：`twoLineEncoder` 升级为 `prettyConsoleEncoder`，新增彩色输出支持（`format: color`），字段以 `key=value` 格式内联展示
- **时间格式化增强**：时间戳字段采用用户配置的自定义格式
- 新增 `unknownCallerFields()` 回退，边界情况不再 panic
- 更新单元测试覆盖新的日志格式

### 配置模块增强

- 文件系统配置新增多磁盘驱动支持（local / oss / cos / minio / s3）
- 数据库配置支持多连接（`database.connections.<name>.*`）
- JWT 配置新增 guards 段

### 文档更新

- `docs/configuration.md` 大幅扩充（304 行），覆盖 Go 配置文件和 YAML 双配置源完整说明
- `docs/getting-started.md` 更新，反映最新配置结构和用法
- `docs/facade.md` 补充 JWT Guard 用法
- 删除过时的 `docs/grpc.md`

---

**完整变更**: [v0.3.0...v0.4.0](https://github.com/zhoudm1743/go-fast/compare/v0.3.0...v0.4.0)

**统计**: 36 个文件变更，+1673 / -1011 行

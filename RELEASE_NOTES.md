# v0.3.0 Release Notes

> 发布日期：2026-07-31

## 新特性

### 日志：logrus -> zap 迁移 (`feat(log)`)

底层日志库从 logrus 迁移到 [zap](https://github.com/uber-go/zap)，性能大幅提升。

- 三种输出模式：新增 `log.mode` 配置（`console` / `file` / `hybrid`），灵活控制日志输出目标
- 可配置文件格式：新增 `log.file_format`，文件日志不再硬编码为 JSON
- 自定义时间格式：新增 `log.timestamp_format` 配置项
- 接口扩展：`contracts.Log` 新增 `WithError`、`WithContext` 方法
- 新增 20 个单元测试覆盖

```yaml
# 配置示例
log:
  mode: hybrid           # console | file | hybrid
  file_format: json      # json | text
  timestamp_format: "2006-01-02 15:04:05"
```

## 修复

### 队列：重构为 Worker 池模式 (`fix(queue)`)

- 新增有界 worker 池控制并发上限
- 新增 `Dispatch` goroutine panic 恢复
- 新增 Retry 重试机制（线性退避 100ms）
- 新增 `Boot` 优雅关闭（排空队列后退出）
- 移除死代码 `Register` / `m.jobs` / `RegisterJobs`
- 用 `contracts.Log` 替换 `fmt.Printf` 输出错误
- 新增 14 个单元测试（含 race 检测）

### ID：Crockford Base32 增强 (`fix(id)`)

- `Parse` 支持大小写不敏感、纠错字符（I/L->1, O->0）、连字符
- 错误信息改为中文
- 新增 8 个测试用例

### 配置：并发安全与修复 (`fix(config)`)

- 添加 `sync.RWMutex` 保护并发读写
- 修复 `GetStringMap` 不存在 key 时返回 nil 问题
- `GetStringMap` 补全 `defaultValue` 参数
- 新增 37 个单元测试

## 重构

### HTTP：validation 子包归入 (`refactor(http)`)

- `framework/validation` 移至 `framework/http/validation/`
- `validation.ServiceProvider` 逻辑并入 `http.ServiceProvider.Register()`
- `bootstrap/app.go` 简化 Provider 注册

---

**完整变更**: `v0.2.0...v0.3.0`

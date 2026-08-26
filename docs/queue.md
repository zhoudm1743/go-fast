# 队列系统


> 框架 API 来自独立 module [`go-fast-framework`](https://github.com/zhoudm1743/go-fast-framework)；业务代码在本仓库（go-fast）。架构说明见 [README.md](README.md)。

## 简介

GoFast 队列系统允许你将耗时任务推送到后台异步执行，提升应用响应速度。  
使用 `facades.Queue()` 操作队列。

---

## 驱动

### 同步驱动（默认）

当前默认使用同步驱动，内部采用**有界 worker 池**模式：
- `Dispatch()` 将任务投递到 worker 池异步执行，队列满时阻塞背压
- `DispatchSync()` 在当前 goroutine 内立即执行
- 默认 worker 数量为 `runtime.NumCPU()`，可通过 `config.yaml` 的 `queue.workers` 配置
- 应用关闭时自动排空队列中的任务后优雅退出

---

## 创建任务

### 任务类结构

```go
// app/jobs/process_order.go
package jobs

import "fmt"

type ProcessOrder struct{}

// Signature 任务唯一标识
func (j *ProcessOrder) Signature() string {
    return "process_order"
}

// Handle 执行任务
func (j *ProcessOrder) Handle(args ...any) error {
    fmt.Println("Processing order:", args)
    return nil
}
```

---

## 调度任务

### 基本调度

```go
import (
    "github.com/zhoudm1743/go-fast/app/jobs"
    "github.com/zhoudm1743/go-fast-framework/contracts"
    "github.com/zhoudm1743/go-fast-framework/facades"
)

err := facades.Queue().Job(&jobs.ProcessOrder{}, []contracts.QueueArg{
    {Type: "string", Value: "order-123"},
    {Type: "int", Value: 1},
}).Dispatch()
```

### 同步调度（立即执行）

```go
err := facades.Queue().Job(&jobs.ProcessOrder{}, []contracts.QueueArg{}).DispatchSync()
```

### 延迟调度

```go
err := facades.Queue().Job(&jobs.ProcessOrder{}, []contracts.QueueArg{}).
    Delay(time.Now().Add(100 * time.Second)).
    Dispatch()
```

### 指定队列

```go
err := facades.Queue().Job(&jobs.ProcessOrder{}, []contracts.QueueArg{}).
    OnQueue("emails").
    Dispatch()
```

### 指定连接

```go
err := facades.Queue().Job(&jobs.ProcessOrder{}, []contracts.QueueArg{}).
    OnConnection("redis").
    OnQueue("processing").
    Dispatch()
```

### 任务链

任务链按顺序执行，任一失败则终止后续任务：

```go
err := facades.Queue().Chain([]contracts.QueueChain{
    {
        Job:  &jobs.ProcessOrder{},
        Args: []contracts.QueueArg{{Type: "int", Value: 1}},
    },
    {
        Job:  &jobs.SendNotification{},
        Args: []contracts.QueueArg{{Type: "string", Value: "done"}},
    },
}).Dispatch()
```

---

### 重试

通过 `Retry()` 设置失败后的最大重试次数（不含首次执行）：

```go
err := facades.Queue().Job(&jobs.ProcessOrder{}, []contracts.QueueArg{}).
    Retry(3).  // 失败后最多重试 3 次
    Dispatch()
```

重试间隔采用线性退避：第 n 次重试等待 `(n-1) × 100ms`。每次重试前记录 Warn 日志，全部失败后记录 Error 日志。

### Panic 恢复

任务的 `Handle()` 方法中发生的 panic 会被自动捕获并转换为 error：
- 同步执行（`DispatchSync`）：panic 转为 error 直接返回
- 异步执行（`Dispatch`）：panic 转为 error 由 worker 记录日志

---

## 优雅关闭

队列管理器在应用关闭时会自动：
1. 停止接收新任务（`Dispatch()` 返回错误）
2. 排空缓冲区中已有的任务
3. 等待所有 worker 完成后退出

关闭顺序确保队列排空先于日志关闭，worker 的日志输出不会丢失。

---

## 配置

`config/config.yaml` 中可配置 worker 数量：

```yaml
queue:
  workers: 8  # 默认为 runtime.NumCPU()
```

---

## `QueueArg.Type` 支持的类型

```
bool, int, int8, int16, int32, int64
uint, uint8, uint16, uint32, uint64
float32, float64, string
[]bool, []int, []string, ...
```

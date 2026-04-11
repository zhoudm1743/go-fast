# 队列系统

## 简介

GoFast 队列系统允许你将耗时任务推送到后台异步执行，提升应用响应速度。  
使用 `facades.Queue()` 操作队列。

---

## 驱动

### 同步驱动（默认）

当前默认使用同步驱动：`Dispatch()` 以独立 goroutine 执行，`DispatchSync()` 在当前进程内立即执行。

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

### 注册任务

在 `bootstrap/app.go` 中注册任务类：

```go
import (
    "github.com/zhoudm1743/go-fast/app/jobs"
    "github.com/zhoudm1743/go-fast/framework/contracts"
    goqueue "github.com/zhoudm1743/go-fast/framework/queue"
)

func Boot() foundation.Application {
    app := /* 基础引导 ... */

    goqueue.RegisterJobs(app, []contracts.QueueJob{
        &jobs.ProcessOrder{},
    })

    return app
}
```

---

## 调度任务

### 基本调度

```go
import (
    "github.com/zhoudm1743/go-fast/app/jobs"
    "github.com/zhoudm1743/go-fast/framework/contracts"
    "github.com/zhoudm1743/go-fast/framework/facades"
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

## `QueueArg.Type` 支持的类型

```
bool, int, int8, int16, int32, int64
uint, uint8, uint16, uint32, uint64
float32, float64, string
[]bool, []int, []string, ...
```

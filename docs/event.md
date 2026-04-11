# 事件系统

## 简介

GoFast 事件系统提供简单的观察者模式实现，允许订阅和监听应用中发生的各种事件。  
事件类存储在 `app/events` 目录，监听器存储在 `app/listeners` 目录。

通过事件系统可以解耦系统各部分：一个事件可以有多个独立监听器。

---

## 注册事件与监听器

在 `bootstrap/app.go` 的 `Boot()` 函数中完成注册：

```go
package bootstrap

import (
    "github.com/zhoudm1743/go-fast/app/events"
    "github.com/zhoudm1743/go-fast/app/listeners"
    "github.com/zhoudm1743/go-fast/framework/contracts"
    goevent "github.com/zhoudm1743/go-fast/framework/event"
)

func Boot() foundation.Application {
    app := /* 基础引导 ... */

    // 注册事件 → 监听器映射
    goevent.RegisterEvents(app, map[contracts.Eventer][]contracts.EventListener{
        &events.OrderShipped{}: {
            &listeners.SendShipmentNotification{},
        },
    })

    return app
}
```

---

## 定义事件

事件类是数据容器，`Handle` 方法用于加工参数后传给所有监听器：

```go
// app/events/order_shipped.go
package events

import "github.com/zhoudm1743/go-fast/framework/contracts"

type OrderShipped struct{}

func (e *OrderShipped) Handle(args []contracts.EventArg) ([]contracts.EventArg, error) {
    return args, nil
}
```

---

## 定义监听器

```go
// app/listeners/send_shipment_notification.go
package listeners

import (
    "fmt"
    "github.com/zhoudm1743/go-fast/framework/contracts"
)

type SendShipmentNotification struct{}

func (l *SendShipmentNotification) Signature() string {
    return "send_shipment_notification"
}

// Queue 返回队列配置；Enable=false 时同步执行
func (l *SendShipmentNotification) Queue(args ...any) contracts.EventQueue {
    return contracts.EventQueue{
        Enable:     false,
        Connection: "",
        Queue:      "",
    }
}

func (l *SendShipmentNotification) Handle(args ...any) error {
    fmt.Println("Shipment notification sent:", args)
    return nil
}
```

### 停止事件传播

在同步监听器的 `Handle` 方法中返回 `error`，可阻止后续监听器执行：

```go
func (l *MyListener) Handle(args ...any) error {
    return errors.New("stop propagation")
}
```

---

## 事件监听器队列

若监听器执行耗时任务，可启用队列异步处理：

```go
func (l *SendShipmentNotification) Queue(args ...any) contracts.EventQueue {
    return contracts.EventQueue{
        Enable:     true,   // 启用异步队列
        Connection: "",
        Queue:      "",
    }
}
```

启用后，监听器在独立 goroutine 中异步执行，不影响主流程。

---

## 派发事件

通过 `facades.Event().Job().Dispatch()` 派发事件：

```go
package controllers

import (
    "github.com/zhoudm1743/go-fast/app/events"
    "github.com/zhoudm1743/go-fast/framework/contracts"
    "github.com/zhoudm1743/go-fast/framework/facades"
)

func (c *OrderController) Ship(ctx http.Context) {
    err := facades.Event().Job(&events.OrderShipped{}, []contracts.EventArg{
        {Type: "string", Value: "order-123"},
        {Type: "int", Value: 1},
    }).Dispatch()
    if err != nil {
        // 处理错误
    }
}
```

---

## `EventArg.Type` 支持的类型

```
bool, int, int8, int16, int32, int64
uint, uint8, uint16, uint32, uint64
float32, float64, string
[]bool, []int, []string, ...
```

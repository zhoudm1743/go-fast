# 任务调度

## 简介

GoFast 任务调度基于 `robfig/cron/v3` 实现，允许你在代码中流畅地定义定时任务，而无需在服务器上手动配置 Cron。  
使用 `facades.Schedule()` 操作调度器。

---

## 定义调度任务

在 `bootstrap/app.go` 的 `Boot()` 函数中定义并注册：

```go
import (
    "fmt"
    "github.com/zhoudm1743/go-fast/framework/contracts"
    "github.com/zhoudm1743/go-fast/framework/facades"
    goschedule "github.com/zhoudm1743/go-fast/framework/schedule"
)

func Boot() foundation.Application {
    app := /* 基础引导 ... */

    // 注册调度任务并启动调度器
    goschedule.RegisterSchedule(app, []contracts.ScheduleEvent{
        facades.Schedule().Call(func() {
            fmt.Println("每分钟执行一次")
        }).EveryMinute().Name("ping"),

        facades.Schedule().Command("send:emails").Daily(),
    })

    return app
}
```

> **注意**：`goschedule.RegisterSchedule` 必须在 `facades.SetApp(app)` 之后调用，确保 Schedule facade 可用。

---

## 闭包调度

```go
facades.Schedule().Call(func() {
    // 任意 Go 代码
}).Daily()
```

## Artisan 命令调度

```go
facades.Schedule().Command("send:emails name").Daily()
```

---

## 调度频率选项

| 方法 | 说明 |
|------|------|
| `.Cron("* * * * *")` | 自定义 Cron（分钟级，5 字段） |
| `.Cron("* * * * * *")` | 自定义 Cron（秒级，6 字段） |
| `.EverySecond()` | 每秒执行一次 |
| `.EveryTwoSeconds()` | 每 2 秒执行一次 |
| `.EveryFiveSeconds()` | 每 5 秒执行一次 |
| `.EveryTenSeconds()` | 每 10 秒执行一次 |
| `.EveryFifteenSeconds()` | 每 15 秒执行一次 |
| `.EveryTwentySeconds()` | 每 20 秒执行一次 |
| `.EveryThirtySeconds()` | 每 30 秒执行一次 |
| `.EveryMinute()` | 每分钟执行一次 |
| `.EveryTwoMinutes()` | 每 2 分钟执行一次 |
| `.EveryThreeMinutes()` | 每 3 分钟执行一次 |
| `.EveryFiveMinutes()` | 每 5 分钟执行一次 |
| `.EveryTenMinutes()` | 每 10 分钟执行一次 |
| `.EveryFifteenMinutes()` | 每 15 分钟执行一次 |
| `.EveryThirtyMinutes()` | 每 30 分钟执行一次 |
| `.Hourly()` | 每小时执行一次 |
| `.HourlyAt(17)` | 每小时第 17 分钟执行 |
| `.EveryTwoHours()` | 每 2 小时执行一次 |
| `.EveryThreeHours()` | 每 3 小时执行一次 |
| `.EveryFourHours()` | 每 4 小时执行一次 |
| `.EverySixHours()` | 每 6 小时执行一次 |
| `.Daily()` | 每天 00:00 执行 |
| `.DailyAt("13:00")` | 每天 13:00 执行 |
| `.Weekdays()` | 每周一至周五执行 |
| `.Weekends()` | 每周六、日执行 |
| `.Mondays()` ~ `.Sundays()` | 指定某天执行 |
| `.Weekly()` | 每周日 00:00 执行 |
| `.Monthly()` | 每月 1 日 00:00 执行 |
| `.Quarterly()` | 每季度第一天 00:00 执行 |
| `.Yearly()` | 每年 1 月 1 日 00:00 执行 |

---

## 避免任务重复

默认情况下，即使上次未完成，任务也会按时触发。使用以下方法控制：

```go
// 上次未完成则跳过本次
facades.Schedule().Command("send:emails").EveryMinute().SkipIfStillRunning()

// 上次未完成则等待后再执行
facades.Schedule().Command("send:emails").EveryMinute().DelayIfStillRunning()
```

---

## 任务只在一台服务器上运行

多服务器部署时，防止重复执行（需配置 Redis 缓存）：

```go
facades.Schedule().Command("report:generate").Daily().OnOneServer()

// 闭包任务必须命名
facades.Schedule().Call(func() {
    fmt.Println("单服务器任务")
}).Daily().OnOneServer().Name("generate-report")
```

> 需要应用使用 Redis 缓存驱动，确保所有服务器连接同一 Redis 实例。

---

## 任务命名

```go
facades.Schedule().Call(func() {
    // ...
}).Daily().Name("my-daily-task")
```

---

## 启动调度器

调度器通过 `goschedule.RegisterSchedule` 启动，无需额外操作。  
若需手动控制启停：

```go
// 获取调度器实例
scheduler := facades.Schedule()

// 停止调度器
scheduler.Stop()

// 查看已注册任务
for _, task := range scheduler.List() {
    fmt.Println(task)
}
```

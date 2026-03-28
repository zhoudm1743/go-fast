# GoFast Fast 控制台指南

> Fast 是 GoFast 内置的命令行工具。它提供了一系列脚手架命令，帮助你快速生成模型、控制器、服务提供者和工具类，同时支持你编写任意自定义命令。

---

## 目录

- [基本用法](#一基本用法)
- [内置命令](#二内置命令)
  - [list](#list)
  - [help](#help)
  - [make:model](#makemodel)
  - [make:controller](#makecontroller)
  - [make:provider](#makeprovider)
  - [make:validator](#makevalidator)
  - [make:utils](#makeutils)
- [编写自定义命令](#三编写自定义命令)
  - [命令结构](#命令结构)
  - [声明选项与参数](#声明选项与参数)
  - [接收用户输入](#接收用户输入)
  - [控制台输出](#控制台输出)
  - [注册命令](#注册命令)
- [以编程方式执行命令](#四以编程方式执行命令)

---

## 一、基本用法

```bash
# 查看所有可用命令
go run . fast list

# 查看某命令的帮助
go run . fast help make:model

# 命令 + --help 同效
go run . fast make:model --help
```

---

## 二、内置命令

### `list`

列出所有已注册的命令，按分类分组显示。

```bash
go run . fast list
```

---

### `help`

查看指定命令的详细说明、参数和选项。

```bash
go run . fast help <command>
```

---

### `make:model`

在 `app/models/` 下生成一个新模型，自动嵌入 `database.Model`（UUID v7 主键 + 时间戳）。

```bash
go run . fast make:model <Name> [--soft-delete]
```

| 参数 / 选项 | 说明 |
|-----------|------|
| `Name` | 模型名（PascalCase），如 `Post`、`UserProfile` |
| `--soft-delete` / `-s` | 改用 `database.ModelWithSoftDelete`（含 `deleted_at` 软删除字段） |

**示例**

```bash
# 生成 app/models/post.go
go run . fast make:model Post

# 生成 app/models/order_item.go（带软删除）
go run . fast make:model OrderItem --soft-delete
```

**生成内容预览**（`app/models/post.go`）

```go
package models

import "github.com/zhoudm1743/go-fast/framework/database"

// Post 模型。
// 嵌入 database.Model 自动获得 UUID v7 主键、CreatedAt、UpdatedAt。
type Post struct {
    database.Model
    // TODO: 在此添加字段
    // Name string `gorm:"size:100;not null" json:"name"`
}
```

---

### `make:controller`

在 `app/http/<group>/controllers/` 下生成控制器，默认同时在 `app/http/<group>/requests/` 下生成请求文件。

```bash
go run . fast make:controller <Name> [--group admin|app] [--no-request]
```

| 参数 / 选项 | 说明 |
|-----------|------|
| `Name` | 控制器名（PascalCase），如 `Post`、`PostController` |
| `--group` / `-g` | 路由分组：`admin`（后台）或 `app`（前台），默认 `app` |
| `--no-request` | 跳过请求文件生成，仅生成控制器 |

**示例**

```bash
# 前台控制器（默认 app）
go run . fast make:controller Comment

# 后台控制器
go run . fast make:controller Post --group admin

# 仅生成控制器，不生成请求文件
go run . fast make:controller Post --group admin --no-request
```

**生成文件**

```
app/http/admin/controllers/post_controller.go
app/http/admin/requests/post.go
```

**生成内容预览**（控制器）

```go
package controllers

// PostController Post 控制器。
type PostController struct{}

func (c *PostController) Prefix() string { return "/posts" }

func (c *PostController) Boot(r contracts.Route) {
    r.Get("/", c.Index)         // GET    /posts
    r.Get("/:id", c.Show)       // GET    /posts/:id
    r.Post("/", c.Store)        // POST   /posts
    r.Put("/:id", c.Update)     // PUT    /posts/:id
    r.Delete("/:id", c.Destroy) // DELETE /posts/:id
}
```

**生成内容预览**（请求）

```go
package requests

type ListPostRequest struct {
    Page int `query:"page" binding:"omitempty,gte=1"`
    Size int `query:"size" binding:"omitempty,gte=1,lte=100"`
}

type CreatePostRequest struct {
    // TODO: 添加字段
}

type UpdatePostRequest struct {
    ID string `uri:"id" binding:"required"`
    // TODO: 添加字段
}
```

> 生成后需在对应路由文件（`routes/admin.go` 或 `routes/app.go`）中用 `r.Register(&controllers.PostController{})` 注册控制器。

---

### `make:provider`

在 `app/providers/` 下生成一个空白服务提供者，遵循框架 `ServiceProvider` 接口。

```bash
go run . fast make:provider <Name>
```

| 参数 | 说明 |
|------|------|
| `Name` | 提供者名（PascalCase），自动补全 `ServiceProvider` 后缀 |

**示例**

```bash
# 生成 app/providers/redis_service_provider.go
go run . fast make:provider Redis

# 也可以带完整后缀
go run . fast make:provider RedisServiceProvider
```

**生成内容预览**

```go
package providers

import "github.com/zhoudm1743/go-fast/framework/foundation"

type RedisServiceProvider struct{}

func (sp *RedisServiceProvider) Register(app foundation.Application) {
    // app.Singleton("redis", func(app foundation.Application) (any, error) {
    //     return NewRedisClient(), nil
    // })
}

func (sp *RedisServiceProvider) Boot(_ foundation.Application) error {
    return nil
}
```

> **注意：** 生成后需在 `bootstrap/app.go` 的 `providers()` 函数中注册：
> ```go
> &providers.RedisServiceProvider{},
> ```

---

### `make:validator`

在 `app/rules/` 下生成一个自定义验证规则，基于 `go-playground/validator`。

```bash
go run . fast make:validator <Name>
```

| 参数 | 说明 |
|------|------|
| `Name` | 规则名（PascalCase），自动补全 `Rule` 后缀 |

**示例**

```bash
# 生成 app/rules/phone_rule.go，binding tag 名为 "phone"
go run . fast make:validator Phone
```

**生成内容预览**

```go
package rules

import "github.com/go-playground/validator/v10"

// PhoneRule 自定义验证规则。
// 在 binding tag 中使用：binding:"phone"
type PhoneRule struct{}

func (r *PhoneRule) Rule() string { return "phone" }

func (r *PhoneRule) Validate(fl validator.FieldLevel) bool {
    // TODO: 实现验证逻辑
    return true
}

func (r *PhoneRule) Message() string {
    return "The :attribute is not a valid phone."
}

func (r *PhoneRule) RegistrationFunc() validator.Func {
    return r.Validate
}
```

**注册规则**（在某个 ServiceProvider 的 `Boot` 方法中）：

```go
func (sp *MyServiceProvider) Boot(app foundation.Application) error {
    v := app.MustMake("validator").(contracts.Validation)
    v.RegisterRule(&rules.PhoneRule{})
    return nil
}
```

之后即可在请求结构体中使用：

```go
type CreateUserRequest struct {
    Phone string `json:"phone" binding:"required,phone"`
}
```

---

### `make:utils`

在 `framework/utils/` 下生成一个工具集文件，遵循现有 `XxxUtil = xxxUtil{}` 规范。

```bash
go run . fast make:utils <Name>
```

| 参数 | 说明 |
|------|------|
| `Name` | 工具名（PascalCase），如 `Jwt`、`Encrypt` |

**示例**

```bash
# 生成 framework/utils/jwt.go
go run . fast make:utils Jwt
```

**生成内容预览**

```go
package utils

var JwtUtil = jwtUtil{}

type jwtUtil struct{}

// Add your methods below.
//
// Example:
// func (r jwtUtil) Foo(s string) string {
//     return s
// }
```

**使用方式**（在任意业务代码中）：

```go
import "github.com/zhoudm1743/go-fast/framework/utils"

token := utils.JwtUtil.GenerateToken(userID)
```

---

## 三、编写自定义命令

### 命令结构

所有命令需实现 `contracts.ConsoleCommand` 接口：

```go
type ConsoleCommand interface {
    Signature() string
    Description() string
    Extend() CommandExtend
    Handle(ctx ConsoleContext) error
}
```

**完整示例：**

```go
// app/console/commands/send_emails.go
package commands

import (
    "github.com/zhoudm1743/go-fast/framework/fast"
    "github.com/zhoudm1743/go-fast/framework/contracts"
    "github.com/zhoudm1743/go-fast/framework/facades"
)

type SendEmailsCommand struct{}

func (c *SendEmailsCommand) Signature() string   { return "send:emails" }
func (c *SendEmailsCommand) Description() string { return "Send pending emails to users" }

func (c *SendEmailsCommand) Extend() contracts.CommandExtend {
    return contracts.CommandExtend{
        Category: "send",
        Flags: []contracts.ConsoleFlag{
            &fast.StringFlag{
                Name:    "lang",
                Value:   "zh",
                Aliases: []string{"l"},
                Usage:   "Email language",
            },
        },
        Arguments: []contracts.ConsoleArgument{
            &fast.StringArgument{
                Name:     "subject",
                Usage:    "Email subject",
                Required: true,
            },
        },
    }
}

func (c *SendEmailsCommand) Handle(ctx contracts.ConsoleContext) error {
    subject := ctx.Argument(0)
    lang := ctx.Option("lang")

    ctx.Info("Sending emails...")
    facades.Log().Infof("subject=%s lang=%s", subject, lang)

    ctx.Line("Done.")
    return nil
}
```

---

### 声明选项与参数

#### 选项（Flag）

选项以 `--key value` 或 `-alias value` 传入。

| 类型 | 说明 |
|------|------|
| `fast.StringFlag` | 字符串选项 |
| `fast.BoolFlag` | 布尔开关，声明即为 `true`，无需值 |
| `fast.IntFlag` | 整数选项 |
| `fast.StringSliceFlag` | 可多次传入的字符串切片选项 |

```go
Flags: []contracts.ConsoleFlag{
    &fast.StringFlag{
        Name:    "output",
        Value:   "stdout",   // 默认值
        Aliases: []string{"o"},
        Usage:   "Output destination",
    },
    &fast.BoolFlag{
        Name:  "verbose",
        Usage: "Enable verbose output",
    },
},
```

```bash
go run . fast my:cmd --output file.txt --verbose
go run . fast my:cmd -o file.txt
```

#### 参数（Argument）

位置参数按顺序紧跟在命令签名后面。

```go
Arguments: []contracts.ConsoleArgument{
    &fast.StringArgument{
        Name:     "name",
        Usage:    "Target name",
        Required: true,
    },
},
```

```bash
go run . fast my:cmd John
```

在 `Handle` 中读取：

```go
name := ctx.Argument(0)    // 按索引
all  := ctx.Arguments()    // 全部参数切片
lang := ctx.Option("lang") // 选项
verbose := ctx.OptionBool("verbose")
```

---

### 接收用户输入

#### 文本输入

```go
name, err := ctx.Ask("What is your name?", contracts.AskOption{
    Default: "GoFast",
})
```

#### 密码输入

```go
pwd, err := ctx.Secret("Enter password:", contracts.SecretOption{
    Validate: func(s string) error {
        if len(s) < 8 {
            return fmt.Errorf("密码至少 8 位")
        }
        return nil
    },
})
```

#### 确认操作

```go
if ctx.Confirm("确认删除所有数据？", contracts.ConfirmOption{
    Default:     false,
    Affirmative: "yes",
    Negative:    "no",
}) {
    // 执行删除
}
```

#### 单选

```go
lang, err := ctx.Choice("选择语言：", []contracts.ConsoleChoice{
    {Key: "go",  Value: "Go"},
    {Key: "py",  Value: "Python"},
    {Key: "ts",  Value: "TypeScript"},
})
```

#### 多选

```go
langs, err := ctx.MultiSelect("选择框架：", []contracts.ConsoleChoice{
    {Key: "fiber",  Value: "Fiber"},
    {Key: "gin",    Value: "Gin"},
    {Key: "echo",   Value: "Echo"},
})
```

---

### 控制台输出

```go
ctx.Line("普通消息")
ctx.Info("成功 / 信息（绿色）")
ctx.Comment("注释（青色）")
ctx.Warning("警告（黄色）")
ctx.Error("错误（红色，输出到 stderr）")
ctx.NewLine()    // 输出一个空行
ctx.NewLine(2)   // 输出两个空行
```

---

### 注册命令

在 `bootstrap/commands.go` 的 `Commands()` 函数中追加：

```go
// bootstrap/commands.go
package bootstrap

import (
    "github.com/zhoudm1743/go-fast/app/console/commands"
    "github.com/zhoudm1743/go-fast/framework/contracts"
)

func Commands() []contracts.ConsoleCommand {
    return []contracts.ConsoleCommand{
        &commands.SendEmailsCommand{},
        // 继续追加...
    }
}
```

框架在 `bootstrap.Boot()` 之后自动将这批命令注册到内核，无需其他配置。

---

## 四、以编程方式执行命令

在 HTTP Handler 或其他业务代码中调用 Fast 命令：

```go
// 无参数
facades.Fast().Call("send:emails MySubject")

// 携带选项
facades.Fast().Call("send:emails MySubject --lang en")
```

也可以传入参数切片：

```go
facades.Fast().Run([]string{"send:emails", "MySubject", "--lang", "en"})
```

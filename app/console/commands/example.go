package commands

import (
	"strings"

	"github.com/zhoudm1743/go-fast/framework/contracts"
	fast "github.com/zhoudm1743/go-fast/framework/fast"
)

// ExampleCommand 示例命令，演示如何定义并使用 Fast 命令。
// 运行方式：go run . fast example --name GoFast
// 删除本文件或将其注释掉即可移除此示例。
type ExampleCommand struct{}

// Signature 命令签名（唯一，用于调用）。
func (c *ExampleCommand) Signature() string {
	return "example"
}

// Description 命令描述，显示在 list 中。
func (c *ExampleCommand) Description() string {
	return "An example command to show Fast usage"
}

// Extend 声明选项、参数、分类。
func (c *ExampleCommand) Extend() contracts.CommandExtend {
	return contracts.CommandExtend{
		Category: "example",
		Flags: []contracts.ConsoleFlag{
			&fast.StringFlag{
				Name:    "name",
				Value:   "World",
				Aliases: []string{"n"},
				Usage:   "Name to greet",
			},
			&fast.BoolFlag{
				Name:  "shout",
				Usage: "Output in uppercase",
			},
		},
		Arguments: []contracts.ConsoleArgument{
			&fast.StringArgument{
				Name:  "message",
				Usage: "Optional extra message",
			},
		},
	}
}

// Handle 命令执行逻辑。
func (c *ExampleCommand) Handle(ctx contracts.ConsoleContext) error {
	name := ctx.Option("name")
	shout := ctx.OptionBool("shout")
	extra := ctx.Argument(0)

	greeting := "Hello, " + name + "!"
	if shout {
		greeting = strings.ToUpper(greeting)
	}

	ctx.Info(greeting)

	if extra != "" {
		ctx.Comment("Extra: " + extra)
	}

	return nil
}

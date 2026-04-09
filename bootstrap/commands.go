package bootstrap

import (
	"github.com/zhoudm1743/go-fast/app/console/commands"
	"github.com/zhoudm1743/go-fast/framework/contracts"
)

// Commands 返回所有注册的控制台命令。
// 在此处追加自定义命令，框架会在 Boot 完成后自动注册到 Fast 内核。
func Commands() []contracts.ConsoleCommand {
	return []contracts.ConsoleCommand{
		&commands.ExampleCommand{},
		&commands.SeedCommand{},
	}
}

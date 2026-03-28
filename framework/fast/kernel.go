package fast

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/zhoudm1743/go-fast/framework/contracts"
)

// fastKernel 实现 contracts.Fast。
type fastKernel struct {
	commands map[string]contracts.ConsoleCommand
}

func newKernel() *fastKernel {
	k := &fastKernel{
		commands: make(map[string]contracts.ConsoleCommand),
	}
	// 内置命令（list / help）
	k.commands["list"] = &listCommand{k: k}
	k.commands["help"] = &helpCommand{k: k}
	// 内置 make:* 脚手架命令
	k.commands["make:model"] = &MakeModelCommand{}
	k.commands["make:controller"] = &MakeControllerCommand{}
	k.commands["make:provider"] = &MakeProviderCommand{}
	k.commands["make:validator"] = &MakeValidatorCommand{}
	k.commands["make:command"] = &MakeCommandCommand{}
	k.commands["make:utils"] = &MakeUtilsCommand{}
	return k
}

// ─── contracts.Fast 实现 ────────────────────────────────────────────────────

func (k *fastKernel) Register(commands []contracts.ConsoleCommand) {
	for _, cmd := range commands {
		k.commands[cmd.Signature()] = cmd
	}
}

func (k *fastKernel) Call(command string) error {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return fmt.Errorf("fast: empty command string")
	}
	return k.Run(parts)
}

func (k *fastKernel) Run(args []string) error {
	if len(args) == 0 {
		// 无参数时显示命令列表
		return k.Run([]string{"list"})
	}

	sig := args[0]
	rest := args[1:]

	// --help / -h 转发到 help 命令
	for _, a := range rest {
		if a == "--help" || a == "-h" {
			return k.Run([]string{"help", sig})
		}
	}

	cmd, ok := k.commands[sig]
	if !ok {
		fmt.Fprintf(os.Stderr,
			"%s命令 \"%s\" 不存在。%s\n\n运行 go run . fast list 查看所有命令\n",
			colorRed, sig, colorReset,
		)
		return fmt.Errorf("command not found: %s", sig)
	}

	ext := cmd.Extend()
	options, positional := parseArgs(rest, ext.Flags)
	ctx := newConsoleContext(positional, options)

	return cmd.Handle(ctx)
}

// ─── 参数解析 ──────────────────────────────────────────────────────────────────

// parseArgs 将原始参数切片分离为选项 map 和位置参数切片。
// 支持：--key value、--key=value、-alias value、--bool-flag。
func parseArgs(rawArgs []string, flags []contracts.ConsoleFlag) (options map[string]string, positional []string) {
	options = make(map[string]string)

	// 构建 name/alias → flag 的查找表
	flagMap := make(map[string]contracts.ConsoleFlag)
	for _, f := range flags {
		flagMap[f.FlagName()] = f
		for _, alias := range f.FlagAliases() {
			flagMap[alias] = f
		}
	}

	i := 0
	for i < len(rawArgs) {
		arg := rawArgs[i]
		switch {
		case strings.HasPrefix(arg, "--"):
			key := arg[2:]
			// 支持 --key=value 格式
			if eqIdx := strings.IndexByte(key, '='); eqIdx != -1 {
				name, val := key[:eqIdx], key[eqIdx+1:]
				if f, ok := flagMap[name]; ok {
					options[f.FlagName()] = val
				} else {
					options[name] = val
				}
				break
			}
			if f, ok := flagMap[key]; ok {
				if f.IsBoolFlag() {
					options[f.FlagName()] = "true"
				} else if i+1 < len(rawArgs) && !strings.HasPrefix(rawArgs[i+1], "-") {
					options[f.FlagName()] = rawArgs[i+1]
					i++
				} else {
					options[f.FlagName()] = f.FlagDefaultStr()
				}
			} else {
				// 未声明的 flag，宽松处理
				if i+1 < len(rawArgs) && !strings.HasPrefix(rawArgs[i+1], "-") {
					options[key] = rawArgs[i+1]
					i++
				} else {
					options[key] = "true"
				}
			}

		case strings.HasPrefix(arg, "-") && len(arg) > 1:
			key := arg[1:]
			if f, ok := flagMap[key]; ok {
				if f.IsBoolFlag() {
					options[f.FlagName()] = "true"
				} else if i+1 < len(rawArgs) && !strings.HasPrefix(rawArgs[i+1], "-") {
					options[f.FlagName()] = rawArgs[i+1]
					i++
				}
			} else {
				if i+1 < len(rawArgs) && !strings.HasPrefix(rawArgs[i+1], "-") {
					options[key] = rawArgs[i+1]
					i++
				} else {
					options[key] = "true"
				}
			}

		default:
			positional = append(positional, arg)
		}
		i++
	}

	// 补全未指定的 flag 默认值
	for _, f := range flags {
		if _, exist := options[f.FlagName()]; !exist {
			options[f.FlagName()] = f.FlagDefaultStr()
		}
	}

	return options, positional
}

// ─── 内置命令：list ────────────────────────────────────────────────────────────

type listCommand struct{ k *fastKernel }

func (c *listCommand) Signature() string               { return "list" }
func (c *listCommand) Description() string             { return "列出所有可用命令" }
func (c *listCommand) Extend() contracts.CommandExtend { return contracts.CommandExtend{} }

func (c *listCommand) Handle(_ contracts.ConsoleContext) error {
	// 按 category 分组
	categories := make(map[string][]contracts.ConsoleCommand)
	for _, cmd := range c.k.commands {
		cat := cmd.Extend().Category
		categories[cat] = append(categories[cat], cmd)
	}

	// 排序分类名
	var catKeys []string
	for k := range categories {
		catKeys = append(catKeys, k)
	}
	sort.Strings(catKeys)

	fmt.Printf("%sGoFast%s 框架\n\n", colorGreen, colorReset)
	fmt.Printf("%s用法：%s\n  go run . fast [命令] [参数] [--选项]\n\n", colorYellow, colorReset)
	fmt.Printf("%s可用命令：%s\n", colorYellow, colorReset)

	// 无分类命令优先显示
	if cmds, ok := categories[""]; ok {
		sort.Slice(cmds, func(i, j int) bool { return cmds[i].Signature() < cmds[j].Signature() })
		for _, cmd := range cmds {
			fmt.Printf("  %s%-24s%s %s\n", colorGreen, cmd.Signature(), colorReset, cmd.Description())
		}
	}

	// 有分类的命令分组显示
	for _, cat := range catKeys {
		if cat == "" {
			continue
		}
		cmds := categories[cat]
		sort.Slice(cmds, func(i, j int) bool { return cmds[i].Signature() < cmds[j].Signature() })

		title := strings.ToUpper(cat[:1]) + cat[1:]
		fmt.Printf("\n %s%s%s\n", colorYellow, title, colorReset)
		for _, cmd := range cmds {
			fmt.Printf("  %s%-24s%s %s\n", colorGreen, cmd.Signature(), colorReset, cmd.Description())
		}
	}

	fmt.Printf("\n%s选项：%s\n", colorYellow, colorReset)
	fmt.Printf("  %s%-24s%s %s\n", colorGreen, "--help, -h", colorReset, "显示指定命令的帮助信息")
	fmt.Println()
	return nil
}

// ─── 内置命令：help ────────────────────────────────────────────────────────────

type helpCommand struct{ k *fastKernel }

func (c *helpCommand) Signature() string               { return "help" }
func (c *helpCommand) Description() string             { return "显示命令的帮助信息" }
func (c *helpCommand) Extend() contracts.CommandExtend { return contracts.CommandExtend{} }

func (c *helpCommand) Handle(ctx contracts.ConsoleContext) error {
	sig := ctx.Argument(0)
	if sig == "" {
		ctx.Error("用法：fast help <命令>")
		return nil
	}

	cmd, ok := c.k.commands[sig]
	if !ok {
		ctx.Error(fmt.Sprintf("命令 \"%s\" 不存在。", sig))
		return nil
	}

	ext := cmd.Extend()

	fmt.Printf("%s描述：%s\n  %s\n\n", colorYellow, colorReset, cmd.Description())
	fmt.Printf("%s用法：%s\n  %s", colorYellow, colorReset, sig)
	for _, arg := range ext.Arguments {
		if arg.IsArgRequired() {
			fmt.Printf(" <%s>", arg.ArgName())
		} else {
			fmt.Printf(" [%s]", arg.ArgName())
		}
	}
	if len(ext.Flags) > 0 {
		fmt.Print(" [--flags]")
	}
	fmt.Println()

	if len(ext.Arguments) > 0 {
		fmt.Printf("\n%s参数：%s\n", colorYellow, colorReset)
		for _, arg := range ext.Arguments {
			req := ""
			if arg.IsArgRequired() {
				req = "（必填）"
			}
			fmt.Printf("  %s%-20s%s %s%s\n", colorGreen, arg.ArgName(), colorReset, arg.ArgUsage(), req)
		}
	}

	if len(ext.Flags) > 0 {
		fmt.Printf("\n%s选项：%s\n", colorYellow, colorReset)
		for _, f := range ext.Flags {
			name := "--" + f.FlagName()
			for _, alias := range f.FlagAliases() {
				name += ", -" + alias
			}
			def := f.FlagDefaultStr()
			if def != "" && def != "false" {
				fmt.Printf("  %s%-28s%s %s [默认值: %s]\n", colorGreen, name, colorReset, f.FlagUsage(), def)
			} else {
				fmt.Printf("  %s%-28s%s %s\n", colorGreen, name, colorReset, f.FlagUsage())
			}
		}
	}

	fmt.Println()
	return nil
}

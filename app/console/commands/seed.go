package commands

import (
	"fmt"

	"github.com/zhoudm1743/go-fast/app/console/commands/seeders"
	"github.com/zhoudm1743/go-fast/framework/contracts"
	"github.com/zhoudm1743/go-fast/framework/facades"
	fast "github.com/zhoudm1743/go-fast/framework/fast"
)

// SeedCommand 数据库种子命令。
// 运行方式：
//
//	go run . fast db:seed                                    ← 运行所有 Seeder（默认连接）
//	go run . fast db:seed --class UserSeeder                 ← 运行指定 Seeder（默认连接）
//	go run . fast db:seed --tenant tenant_acme               ← 对指定租户运行所有 Seeder
//	go run . fast db:seed --tenant tenant_acme --class UserSeeder ← 对指定租户运行指定 Seeder
type SeedCommand struct{}

func (c *SeedCommand) Signature() string {
	return "db:seed"
}

func (c *SeedCommand) Description() string {
	return "Seed the database with records"
}

func (c *SeedCommand) Extend() contracts.CommandExtend {
	return contracts.CommandExtend{
		Category: "db",
		Flags: []contracts.ConsoleFlag{
			&fast.StringFlag{
				Name:    "class",
				Aliases: []string{"c"},
				Usage:   "The name of the seeder class to run (e.g. UserSeeder)",
			},
			&fast.StringFlag{
				Name:    "tenant",
				Aliases: []string{"t"},
				Usage:   "Named connection to seed (for multi-tenant, e.g. tenant_acme)",
			},
		},
	}
}

func (c *SeedCommand) Handle(ctx contracts.ConsoleContext) error {
	// 所有可用 Seeder 注册表，新增 Seeder 后在此追加。
	registry := map[string]seeders.Seeder{
		"UserSeeder":     &seeders.UserSeeder{},
		"DatabaseSeeder": &seeders.DatabaseSeeder{},
	}

	class := ctx.Option("class")
	tenant := ctx.Option("tenant")

	// 1. 确定使用哪个 Query（默认连接 or 租户连接）
	var q contracts.Query
	if tenant != "" {
		ctx.Info(fmt.Sprintf("Target connection: %s", tenant))
		q = facades.DB().Connection(tenant)
	} else {
		q = facades.DB().Query()
	}

	// 2. 确定运行哪个 Seeder
	var seeder seeders.Seeder
	if class == "" {
		seeder = &seeders.DatabaseSeeder{}
	} else {
		s, ok := registry[class]
		if !ok {
			return fmt.Errorf("seeder not found: %s", class)
		}
		seeder = s
	}

	ctx.Info(fmt.Sprintf("Running seeder: %T", seeder))

	if err := seeder.Run(q); err != nil {
		ctx.Error("Seeding failed: " + err.Error())
		return err
	}

	ctx.Info("Database seeding completed successfully.")
	return nil
}

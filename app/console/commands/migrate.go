package commands

import (
	"github.com/zhoudm1743/go-fast/app/models"
	"github.com/zhoudm1743/go-fast/framework/contracts"
	"github.com/zhoudm1743/go-fast/framework/facades"
	fast "github.com/zhoudm1743/go-fast/framework/fast"
)

// MigrateCommand 数据库迁移命令。
// 对 appModels 中注册的所有模型执行 GORM AutoMigrate（自动建表/同步字段）。
//
// 运行方式：
//
//	go run . fast db:migrate               ← 默认连接
//	go run . fast db:migrate --conn mydb   ← 指定命名连接
//
// 新增模型后，在下方 appModels 中追加对应的指针即可。
type MigrateCommand struct{}

// appModels 注册所有需要迁移的 GORM 模型（传指针）。
// ⚠ 有外键依赖时请确保父表模型在前。
var appModels = []any{
	&models.User{},
	&models.Admin{},
}

func (c *MigrateCommand) Signature() string {
	return "db:migrate"
}

func (c *MigrateCommand) Description() string {
	return "Run database migrations (AutoMigrate)"
}

func (c *MigrateCommand) Extend() contracts.CommandExtend {
	return contracts.CommandExtend{
		Category: "db",
		Flags: []contracts.ConsoleFlag{
			&fast.StringFlag{
				Name:    "conn",
				Aliases: []string{"c"},
				Usage:   "Named database connection to migrate (default: main)",
			},
		},
	}
}

func (c *MigrateCommand) Handle(ctx contracts.ConsoleContext) error {
	conn := ctx.Option("conn")

	db := facades.DB()

	var err error
	if conn != "" {
		ctx.Info("Running migrations on connection: " + conn)
		err = db.Driver(conn).AutoMigrate(appModels...)
	} else {
		ctx.Info("Running migrations on default connection...")
		err = db.AutoMigrate(appModels...)
	}

	if err != nil {
		ctx.Error("Migration failed: " + err.Error())
		return err
	}

	ctx.Info("Migration completed successfully.")
	return nil
}

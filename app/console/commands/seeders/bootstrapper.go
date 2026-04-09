package seeders

import (
	"fmt"

	"github.com/zhoudm1743/go-fast/framework/contracts"
	"github.com/zhoudm1743/go-fast/framework/facades"
)

// TenantBootstrapper 租户初始化工具，封装「迁移 + 种子」两个步骤。
// 平台在创建租户后调用 Bootstrap，即可完成该租户数据库的完整初始化。
//
// 用法示例：
//
//	boot := seeders.NewTenantBootstrapper(
//	    []any{&models.User{}, &models.Order{}},
//	    &seeders.DatabaseSeeder{},
//	)
//	if err := boot.Bootstrap("tenant_acme"); err != nil {
//	    log.Fatal(err)
//	}
type TenantBootstrapper struct {
	// Models 需要迁移的 GORM 模型列表（传指针）
	Models []any
	// Seeder 种子入口，通常为 &DatabaseSeeder{}
	Seeder Seeder
}

// NewTenantBootstrapper 创建 TenantBootstrapper。
func NewTenantBootstrapper(models []any, seeder Seeder) *TenantBootstrapper {
	return &TenantBootstrapper{Models: models, Seeder: seeder}
}

// Bootstrap 对指定命名连接执行迁移与种子。
// connName 为 facades.DB().Register() 注册时使用的连接名称。
func (b *TenantBootstrapper) Bootstrap(connName string) error {
	// 1. 自动迁移表结构
	if len(b.Models) > 0 {
		if err := facades.DB().Driver(connName).AutoMigrate(b.Models...); err != nil {
			return fmt.Errorf("migrate tenant %q: %w", connName, err)
		}
	}

	// 2. 执行种子数据
	if b.Seeder != nil {
		q := facades.DB().Connection(connName)
		if err := b.Seeder.Run(q); err != nil {
			return fmt.Errorf("seed tenant %q: %w", connName, err)
		}
	}

	return nil
}

// RegisterAndBootstrap 动态注册租户连接后立即执行迁移与种子。
// 适用于租户数据库配置来自业务数据库（非 config.yaml）的场景。
func (b *TenantBootstrapper) RegisterAndBootstrap(connName string, cfg contracts.ConnectionConfig) error {
	if err := facades.DB().Register(connName, cfg); err != nil {
		return fmt.Errorf("register tenant connection %q: %w", connName, err)
	}
	return b.Bootstrap(connName)
}

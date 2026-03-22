package contracts

import "gorm.io/gorm"

// Orm 数据库 ORM 服务契约。
type Orm interface {
	// DB 获取底层 *gorm.DB 连接实例。
	DB() *gorm.DB
	// Ping 测试数据库连接。
	Ping() error
	// Close 关闭数据库连接。
	Close() error
	// AutoMigrate 自动迁移数据库表结构，传入 GORM Model 指针列表。
	// 插件可通过实现 foundation.Migrator 接口，让框架自动调用此方法完成表结构同步。
	AutoMigrate(models ...any) error
}

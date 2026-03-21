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
}

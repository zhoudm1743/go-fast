package seeders

import "github.com/zhoudm1743/go-fast/framework/contracts"

// Seeder 种子数据接口，每个 Seeder 实现 Run 方法。
// q 为当前操作的数据库连接，支持默认连接和租户连接。
// 在 app/console/commands/seeders/ 目录下创建新文件并实现此接口，
// 然后在 DatabaseSeeder.Run() 中追加即可。
type Seeder interface {
	Run(q contracts.Query) error
}

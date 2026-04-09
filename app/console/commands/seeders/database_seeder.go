package seeders

import "github.com/zhoudm1743/go-fast/framework/contracts"

// DatabaseSeeder 根 Seeder，统一入口。
// 按顺序调用所有子 Seeder，注意有外键依赖时先插父表。
// 新增 Seeder 后在此追加即可。
type DatabaseSeeder struct{}

func (s *DatabaseSeeder) Run(q contracts.Query) error {
	seeders := []Seeder{
		&UserSeeder{},
		// &AdminSeeder{},
		// &ProductSeeder{},
	}

	for _, seeder := range seeders {
		if err := seeder.Run(q); err != nil {
			return err
		}
	}
	return nil
}

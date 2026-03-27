package facades

import "github.com/zhoudm1743/go-fast/framework/contracts"

// Orm 获取 ORM 服务实例。
func Orm() contracts.Orm {
	return App().MustMake("orm").(contracts.Orm)
}

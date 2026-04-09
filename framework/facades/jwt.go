package facades

import "github.com/zhoudm1743/go-fast/framework/contracts"

// JWT 获取 JWT 服务实例。
func JWT() contracts.JWT {
	return App().MustMake("jwt").(contracts.JWT)
}

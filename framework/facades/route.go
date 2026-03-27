package facades

import "github.com/zhoudm1743/go-fast/framework/contracts"

// Route 获取路由服务实例。
func Route() contracts.Route {
	return App().MustMake("route").(contracts.Route)
}

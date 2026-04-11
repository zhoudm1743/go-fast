package events

import "github.com/zhoudm1743/go-fast/framework/contracts"

// OrderShipped 订单已发货事件示例。
type OrderShipped struct{}

// Handle 加工事件参数，返回结果将传递给所有关联监听器。
func (e *OrderShipped) Handle(args []contracts.EventArg) ([]contracts.EventArg, error) {
	return args, nil
}

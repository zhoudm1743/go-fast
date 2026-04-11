package listeners

import (
	"fmt"

	"github.com/zhoudm1743/go-fast/framework/contracts"
)

// SendShipmentNotification 发货通知监听器示例。
type SendShipmentNotification struct{}

func (l *SendShipmentNotification) Signature() string {
	return "send_shipment_notification"
}

// Queue 返回队列配置；Enable=false 时同步执行。
func (l *SendShipmentNotification) Queue(args ...any) contracts.EventQueue {
	return contracts.EventQueue{
		Enable:     false,
		Connection: "",
		Queue:      "",
	}
}

// Handle 处理事件（args 来自 event.Handle 的返回结果）。
func (l *SendShipmentNotification) Handle(args ...any) error {
	fmt.Println("[Listener] SendShipmentNotification received:", args)
	return nil
}

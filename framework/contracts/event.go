package contracts

// Event 事件总线服务契约（预留）。
type Event interface {
	// Dispatch 派发事件。
	Dispatch(event any) error
	// Listen 注册事件监听器。
	Listen(event any, listener any)
}

package contracts

// Test 示例业务服务契约（services/test 实现）。
// 演示自定义 ServiceProvider 注册到容器并通过 Facade 访问的完整流程。
type Test interface {
	// Greet 返回问候语。
	Greet(name string) string

	// Status 返回服务运行状态信息。
	Status() map[string]any
}

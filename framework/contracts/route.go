package contracts

// Route HTTP 路由服务契约。
// 所有 Handler 使用 HandlerFunc 类型，应用代码无需引入任何底层 HTTP 框架。
type Route interface {
	// Run 启动 HTTP 服务器，addr 可选（默认读取配置）。
	Run(addr ...string) error
	// Shutdown 优雅关闭 HTTP 服务器。
	Shutdown() error

	// 路由注册方法，返回自身以支持链式调用。
	Get(path string, handler HandlerFunc) Route
	Post(path string, handler HandlerFunc) Route
	Put(path string, handler HandlerFunc) Route
	Delete(path string, handler HandlerFunc) Route
	Patch(path string, handler HandlerFunc) Route
	Head(path string, handler HandlerFunc) Route
	Options(path string, handler HandlerFunc) Route

	// Group 创建路由组（共享前缀和中间件）。
	Group(prefix string) Route
	// Use 注册中间件，按注册顺序执行。
	Use(middleware ...HandlerFunc) Route
}

package contracts

// ── 控制器接口 ────────────────────────────────────────────────────────

// Controller 控制器接口，所有控制器必须实现。
// 通过 Route.Register() 注册，框架自动处理 Prefix / Middleware 后调用 Boot。
type Controller interface {
	// Boot 声明该控制器的所有路由。
	// r 已被框架定位到正确的路由组下（根据 Prefixer / Middlewarer）。
	Boot(r Route)
}

// Prefixer 可选接口 —— 声明控制器路由前缀。
// Register() 检测到此接口时，自动为该控制器创建路由组。
type Prefixer interface {
	Prefix() string
}

// Middlewarer 可选接口 —— 声明控制器级中间件。
// Register() 检测到此接口时，自动为该控制器的路由组添加中间件。
type Middlewarer interface {
	Middleware() []HandlerFunc
}

// ── 路由服务契约 ──────────────────────────────────────────────────────

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
	// args 支持两种类型，可任意组合：
	//   - HandlerFunc  → 作为中间件
	//   - func(Route)  → 作为回调，在回调内声明该组的路由
	// 无回调时返回 group 对象供链式调用。
	Group(prefix string, args ...any) Route

	// Use 注册中间件，按注册顺序执行。
	Use(middleware ...HandlerFunc) Route

	// Register 批量注册控制器。
	// 框架自动检测控制器的 Prefix() 和 Middleware()（可选），
	// 然后调用 Boot(r) 让控制器声明自己的路由。
	Register(controllers ...Controller) Route
}

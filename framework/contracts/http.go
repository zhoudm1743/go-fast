package contracts

// HandlerFunc 是 GoFast HTTP 处理函数的统一签名。
// 应用层代码只需依赖此类型，无需引入任何底层 HTTP 框架包。
type HandlerFunc func(ctx Context) error

// Context 是传递给每个 Handler 和 Middleware 的核心对象。
// 它封装了请求读取、响应发送、上下文存储三大能力。
type Context interface {
	// ── 请求读取 ──────────────────────────────────

	// Method 返回 HTTP 方法（GET / POST / …）。
	Method() string
	// Path 返回请求路径。
	Path() string
	// Param 返回 URL 路径参数，例如路由 "/users/:id" 中的 "id"。
	Param(key string) string
	// Query 返回查询字符串参数，支持默认值。
	Query(key string, defaultValue ...string) string
	// Header 返回请求头的值。
	Header(key string) string
	// IP 返回客户端 IP 地址。
	IP() string
	// BodyRaw 返回原始请求体字节。
	BodyRaw() []byte

	// Bind 将请求体（JSON / Form）反序列化到 obj，并自动执行验证（binding tag）。
	// 验证失败时返回包含字段错误信息的 error。
	Bind(obj any) error

	// ── 响应发送 ──────────────────────────────────

	// JSON 以 JSON 格式发送响应，code 为 HTTP 状态码。
	JSON(code int, obj any) error
	// String 发送纯文本响应。
	String(code int, s string) error
	// Status 设置响应状态码，返回自身以支持链式调用。
	Status(code int) Context
	// SetHeader 设置响应头，返回自身以支持链式调用。
	SetHeader(key, value string) Context

	// ── 上下文存储（用于 Middleware 传值）────────

	// Value 读取通过 WithValue 存储的键值。
	Value(key string) any
	// WithValue 在当前请求上下文中存储键值对，返回自身。
	WithValue(key string, value any) Context

	// ── Middleware 控制 ───────────────────────────

	// Next 调用调用链中的下一个处理函数（供 Middleware 使用）。
	Next() error
}

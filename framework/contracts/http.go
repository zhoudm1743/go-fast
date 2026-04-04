package contracts

// HandlerFunc 是 GoFast HTTP 处理函数的统一签名。
// 应用层代码只需依赖此类型，无需引入任何底层 HTTP 框架包。
type HandlerFunc func(ctx Context) error

// Response 是统一响应构建器。
// 控制器可通过 ctx.Response() 获取它，快速返回标准 JSON 响应。
type Response interface {
	// Build 构建并发送完整响应。
	Build(status int, code int, message string, data any) error
	// Json 快速返回任意 JSON 响应（自定义 HTTP 状态码，业务码固定为 0）。
	Json(status int, data any, message ...string) error
	// String 快速返回纯文本响应（HTTP 200）。
	String(s string) error
	// File 直接输出存储中的文件内容，默认使用 storage 默认磁盘。
	File(file string, disk ...string) error
	// Download 以附件下载方式输出文件，可自定义下载文件名。
	Download(file string, name string, disk ...string) error
	// Success 快速返回成功响应（HTTP 200, code=0）。
	Success(data any, message ...string) error
	// Fail 快速返回失败响应（默认业务码 code=1）。
	Fail(status int, message string, code ...int) error
	// Created 快速返回创建成功响应（HTTP 201, code=0）。
	Created(data any, message ...string) error
	// Unauthorized 快速返回未授权响应（HTTP 401）。
	Unauthorized(message ...string) error
	// Forbidden 快速返回无权限响应（HTTP 403）。
	Forbidden(message ...string) error
	// NotFound 快速返回资源不存在响应（HTTP 404）。
	NotFound(message ...string) error
	// Validation 快速返回参数验证失败响应（HTTP 422）。
	Validation(err error, message ...string) error
	// Paginate 快速返回分页数据响应（HTTP 200, code=0）。
	Paginate(list any, total int64, page int, size int, message ...string) error
}

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

	// ── 文件与存储 ─────────────────────────────────

	// Storage 返回文件存储服务（供 Response.File / Response.Download 使用）。
	Storage() Storage
	// SendFile 直接从文件系统绝对路径发送文件响应（由底层驱动实现）。
	SendFile(path string) error

	// ── 响应发送 ──────────────────────────────────

	// JSON 以 JSON 格式发送响应，code 为 HTTP 状态码。
	JSON(code int, obj any) error
	// String 发送纯文本响应。
	String(code int, s string) error
	// Response 返回统一响应构建器。
	Response() Response
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
	Abort() error
	AbortWithCode(code int) error
	AbortWithJson(code int, obj any) error
}

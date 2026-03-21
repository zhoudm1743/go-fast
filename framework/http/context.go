package http

import (
	"go-fast/framework/contracts"

	"github.com/gofiber/fiber/v2"
)

// fiberContext 实现 contracts.Context，内部包装 *fiber.Ctx。
// 应用代码只见 contracts.Context，不感知 Fiber。
type fiberContext struct {
	c         *fiber.Ctx
	store     map[string]any
	validator contracts.Validation
}

func newFiberContext(c *fiber.Ctx, v contracts.Validation) *fiberContext {
	return &fiberContext{c: c, store: make(map[string]any), validator: v}
}

// ── 请求读取 ────────────────────────────────────────────────────────

func (ctx *fiberContext) Method() string { return ctx.c.Method() }
func (ctx *fiberContext) Path() string   { return ctx.c.Path() }

func (ctx *fiberContext) Param(key string) string { return ctx.c.Params(key) }

func (ctx *fiberContext) Query(key string, defaultValue ...string) string {
	val := ctx.c.Query(key)
	if val == "" && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return val
}

func (ctx *fiberContext) Header(key string) string { return ctx.c.Get(key) }
func (ctx *fiberContext) IP() string               { return ctx.c.IP() }
func (ctx *fiberContext) BodyRaw() []byte          { return ctx.c.Body() }

// Bind 解析请求体（JSON / Form）并执行验证（binding tag）。
func (ctx *fiberContext) Bind(obj any) error {
	if err := ctx.c.BodyParser(obj); err != nil {
		return err
	}
	if ctx.validator != nil {
		return ctx.validator.Validate(obj)
	}
	return nil
}

// ── 响应发送 ────────────────────────────────────────────────────────

func (ctx *fiberContext) JSON(code int, obj any) error {
	return ctx.c.Status(code).JSON(obj)
}

func (ctx *fiberContext) String(code int, s string) error {
	return ctx.c.Status(code).SendString(s)
}

func (ctx *fiberContext) Status(code int) contracts.Context {
	ctx.c.Status(code)
	return ctx
}

func (ctx *fiberContext) SetHeader(key, value string) contracts.Context {
	ctx.c.Set(key, value)
	return ctx
}

// ── 上下文存储 ──────────────────────────────────────────────────────

func (ctx *fiberContext) Value(key string) any {
	return ctx.store[key]
}

func (ctx *fiberContext) WithValue(key string, value any) contracts.Context {
	ctx.store[key] = value
	return ctx
}

// ── Middleware 控制 ─────────────────────────────────────────────────

func (ctx *fiberContext) Next() error {
	return ctx.c.Next()
}

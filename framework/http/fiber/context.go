package fiber

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/zhoudm1743/go-fast/framework/contracts"
	"github.com/zhoudm1743/go-fast/framework/http/base"
)

// Context 实现 contracts.Context，内部包装 *fiber.Ctx。
// 应用代码只见 contracts.Context，不感知 Fiber。
type Context struct {
	c         *fiber.Ctx
	store     map[string]any
	validator contracts.Validation
	storage   contracts.Storage
}

// NewContext 创建 Fiber 上下文包装器。
func NewContext(c *fiber.Ctx, v contracts.Validation, s contracts.Storage) *Context {
	return &Context{c: c, store: make(map[string]any), validator: v, storage: s}
}

// ── 请求读取 ────────────────────────────────────────────────────────

func (ctx *Context) Method() string { return ctx.c.Method() }
func (ctx *Context) Path() string   { return ctx.c.Path() }

func (ctx *Context) Param(key string) string { return ctx.c.Params(key) }

func (ctx *Context) Query(key string, defaultValue ...string) string {
	val := ctx.c.Query(key)
	if val == "" && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return val
}

func (ctx *Context) Header(key string) string { return ctx.c.Get(key) }
func (ctx *Context) IP() string               { return ctx.c.IP() }
func (ctx *Context) BodyRaw() []byte          { return ctx.c.Body() }

// Bind 将请求数据填充到 obj（URI → Query → Body），最后统一验证。
func (ctx *Context) Bind(obj any) error {
	// 1. URI 路径参数
	bindURI(obj, ctx.c.Params)

	// 2. Query String
	if err := ctx.c.QueryParser(obj); err != nil {
		return err
	}

	// 3. 请求体
	if len(ctx.c.Body()) > 0 {
		if err := ctx.c.BodyParser(obj); err != nil {
			return err
		}
	}

	// 4. 验证
	if ctx.validator != nil {
		return ctx.validator.Validate(obj)
	}
	return nil
}

// ── 文件与存储 ────────────────────────────────────────────────────────

func (ctx *Context) Storage() contracts.Storage { return ctx.storage }

func (ctx *Context) SendFile(path string) error { return ctx.c.SendFile(path) }

// ── 响应发送 ────────────────────────────────────────────────────────

func (ctx *Context) JSON(code int, obj any) error {
	return ctx.c.Status(code).JSON(obj)
}

func (ctx *Context) String(code int, s string) error {
	return ctx.c.Status(code).SendString(s)
}

func (ctx *Context) Response() contracts.Response {
	return base.NewResponse(ctx)
}

func (ctx *Context) Status(code int) contracts.Context {
	ctx.c.Status(code)
	return ctx
}

func (ctx *Context) SetHeader(key, value string) contracts.Context {
	ctx.c.Set(key, value)
	return ctx
}

// ── 上下文存储 ──────────────────────────────────────────────────────

func (ctx *Context) Value(key string) any {
	return ctx.c.Locals(key)
}

func (ctx *Context) WithValue(key string, value any) contracts.Context {
	ctx.c.Locals(key, value)
	return ctx
}

// ── Middleware 控制 ─────────────────────────────────────────────────

func (ctx *Context) Next() error {
	return ctx.c.Next()
}

func (ctx *Context) Abort() error {
	return nil
}

func (ctx *Context) AbortWithCode(code int) error {
	return ctx.c.SendStatus(code)
}

func (ctx *Context) AbortWithJson(code int, obj any) error {
	return ctx.c.Status(code).JSON(obj)
}

// ── 内部：URI 路径参数绑定 ──────────────────────────────────────────

func bindURI(obj any, params func(key string, defaultValue ...string) string) {
	rv := reflect.ValueOf(obj)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return
	}
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		tag := field.Tag.Get("uri")
		if tag == "" || tag == "-" {
			continue
		}
		name := strings.SplitN(tag, ",", 2)[0]
		val := params(name)
		if val == "" {
			continue
		}
		fv := rv.Field(i)
		if !fv.CanSet() {
			continue
		}
		setFieldFromString(fv, val)
	}
}

func setFieldFromString(fv reflect.Value, val string) {
	switch fv.Kind() {
	case reflect.String:
		fv.SetString(val)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if n, err := strconv.ParseInt(val, 10, 64); err == nil {
			fv.SetInt(n)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if n, err := strconv.ParseUint(val, 10, 64); err == nil {
			fv.SetUint(n)
		}
	case reflect.Float32, reflect.Float64:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			fv.SetFloat(f)
		}
	case reflect.Bool:
		if b, err := strconv.ParseBool(val); err == nil {
			fv.SetBool(b)
		}
	}
}

package http

import (
	"reflect"
	"strconv"
	"strings"

	"go-fast/framework/contracts"

	"github.com/gofiber/fiber/v2"
)

// fiberContext 实现 contracts.Context，内部包装 *fiber.Ctx。
// 应用代码只见 contracts.Context，不感知 Fiber。
type fiberContext struct {
	c         *fiber.Ctx
	store     map[string]any
	validator contracts.Validation
	storage   contracts.Storage
}

func newFiberContext(c *fiber.Ctx, v contracts.Validation, s contracts.Storage) *fiberContext {
	return &fiberContext{c: c, store: make(map[string]any), validator: v, storage: s}
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

// Bind 将请求数据填充到 obj，按 tag 多源解析后统一验证：
//
//  1. `uri:"xxx"`        → URL 路径参数（如 /users/:id 中的 id）
//  2. `query:"xxx"`      → Query String（如 ?page=1&size=20）
//  3. `json:"xxx"` 等    → 请求体（JSON / Form / XML，无 body 时跳过）
//
// 最后对整个 obj 执行 binding tag 验证，任意来源的字段均会被验证。
func (ctx *fiberContext) Bind(obj any) error {
	// 1. URI 路径参数（自定义反射填充，支持 `uri` tag）
	bindURI(obj, ctx.c.Params)

	// 2. Query String（Fiber 原生，使用 `query` tag）
	if err := ctx.c.QueryParser(obj); err != nil {
		return err
	}

	// 3. 请求体：仅有 body 时解析（避免 GET 等请求因无 body 报错）
	if len(ctx.c.Body()) > 0 {
		if err := ctx.c.BodyParser(obj); err != nil {
			return err
		}
	}

	// 4. 统一验证（binding tag）
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

func (ctx *fiberContext) Response() contracts.Response {
	return NewResponse(ctx)
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

func (ctx *fiberContext) Abort() error {
	return nil
}

func (ctx *fiberContext) AbortWithCode(code int) error {
	return ctx.c.SendStatus(code)
}

func (ctx *fiberContext) AbortWithJson(code int, obj any) error {
	return ctx.c.Status(code).JSON(obj)
}

// ── 内部：URI 路径参数绑定 ──────────────────────────────────────────

// bindURI 通过反射读取 struct 字段的 `uri` tag，从 URL 路径参数填充对应字段。
// 支持字段类型：string / int* / uint* / float* / bool。
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

// setFieldFromString 将字符串值转换为目标字段类型后赋值。
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

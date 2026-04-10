package gin

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zhoudm1743/go-fast/framework/contracts"
	"github.com/zhoudm1743/go-fast/framework/http/base"
)

// Context 实现 contracts.Context，内部包装 *gin.Context。
// 应用代码只见 contracts.Context，不感知 Gin。
type Context struct {
	c         *gin.Context
	store     map[string]any
	validator contracts.Validation
	storage   contracts.Storage
}

// NewContext 创建 Gin 上下文包装器。
func NewContext(c *gin.Context, v contracts.Validation, s contracts.Storage) *Context {
	return &Context{c: c, store: make(map[string]any), validator: v, storage: s}
}

// ── 请求读取 ────────────────────────────────────────────────────────

func (ctx *Context) Method() string { return ctx.c.Request.Method }
func (ctx *Context) Path() string   { return ctx.c.Request.URL.Path }

func (ctx *Context) Param(key string) string { return ctx.c.Param(key) }

func (ctx *Context) Query(key string, defaultValue ...string) string {
	val := ctx.c.Query(key)
	if val == "" && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return val
}

func (ctx *Context) Header(key string) string { return ctx.c.GetHeader(key) }
func (ctx *Context) IP() string               { return ctx.c.ClientIP() }
func (ctx *Context) BodyRaw() []byte {
	body, _ := ctx.c.GetRawData()
	return body
}

// Bind 将请求数据填充到 obj（URI → Query → Body），最后统一验证。
//
//  1. `uri:"xxx"`   → URL 路径参数
//  2. `query:"xxx"` → Query String
//  3. `json:"xxx"`  → 请求体（仅有 body 时解析）
func (ctx *Context) Bind(obj any) error {
	// 1. URI 路径参数（使用 `uri` tag，与 Fiber 保持一致）
	if err := ctx.c.ShouldBindUri(obj); err != nil {
		// ShouldBindUri 在无 uri tag 字段时也可能返回 error，此处忽略
		_ = err
	}

	// 2. Query String（自定义反射绑定，使用 `query` tag，与 Fiber 保持一致）
	bindQuery(obj, ctx.c.Query)

	// 3. 请求体（仅 POST / PUT / PATCH 等有 body 时解析）
	if ctx.c.Request.ContentLength > 0 {
		if err := ctx.c.ShouldBind(obj); err != nil {
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

// SendFile 使用 Gin 发送文件响应。
func (ctx *Context) SendFile(path string) error {
	ctx.c.File(path)
	return nil
}

// ── 响应发送 ────────────────────────────────────────────────────────

func (ctx *Context) JSON(code int, obj any) error {
	ctx.c.JSON(code, obj)
	return nil
}

func (ctx *Context) String(code int, s string) error {
	ctx.c.String(code, "%s", s)
	return nil
}

func (ctx *Context) Response() contracts.Response {
	return base.NewResponse(ctx)
}

func (ctx *Context) Status(code int) contracts.Context {
	ctx.c.Status(code)
	return ctx
}

func (ctx *Context) SetHeader(key, value string) contracts.Context {
	ctx.c.Header(key, value)
	return ctx
}

// ── 上下文存储 ──────────────────────────────────────────────────────

func (ctx *Context) Value(key string) any {
	val, _ := ctx.c.Get(key)
	return val
}

func (ctx *Context) WithValue(key string, value any) contracts.Context {
	ctx.c.Set(key, value)
	return ctx
}

// ── Middleware 控制 ─────────────────────────────────────────────────

func (ctx *Context) Next() error {
	ctx.c.Next()
	return nil
}

func (ctx *Context) Abort() error {
	ctx.c.Abort()
	return nil
}

func (ctx *Context) AbortWithCode(code int) error {
	ctx.c.AbortWithStatus(code)
	return nil
}

func (ctx *Context) AbortWithJson(code int, obj any) error {
	ctx.c.AbortWithStatusJSON(code, obj)
	return nil
}

// ── 内部：Query 参数绑定（使用 `query` tag，与 Fiber 保持一致）──────

func bindQuery(obj any, queryFn func(key string) string) {
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
		tag := field.Tag.Get("query")
		if tag == "" || tag == "-" {
			continue
		}
		name := strings.SplitN(tag, ",", 2)[0]
		val := queryFn(name)
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

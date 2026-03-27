package http

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/zhoudm1743/go-fast/framework/contracts"
)

// Response 是 GoFast 标准 JSON 响应结构。
// 它实现了 contracts.Response，可通过 ctx.Response() 获取并直接发送。
type Response struct {
	ctx     contracts.Context `json:"-"`
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Data    any               `json:"data,omitempty"`
}

// NewResponse 为当前请求上下文创建一个响应构建器。
func NewResponse(ctx contracts.Context) *Response {
	return &Response{
		ctx:     ctx,
		Code:    0,
		Message: "ok",
	}
}

// Build 构建并发送完整响应。
func (r *Response) Build(status int, code int, message string, data any) error {
	r.Code = code
	r.Message = message
	r.Data = data
	return r.ctx.JSON(status, r)
}

// Json 快速返回任意 JSON 响应（业务码固定为 0）。
func (r *Response) Json(status int, data any, message ...string) error {
	msg := "ok"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return r.Build(status, 0, msg, data)
}

// String 快速返回纯文本响应（HTTP 200）。
func (r *Response) String(s string) error {
	return r.ctx.String(http.StatusOK, s)
}

// File 直接输出存储中的文件内容，默认使用 storage 默认磁盘。
func (r *Response) File(file string, disk ...string) error {
	ctx, ok := r.ctx.(*fiberContext)
	if !ok || ctx.storage == nil {
		return r.Fail(http.StatusInternalServerError, "storage not available")
	}

	var driver contracts.StorageDriver = ctx.storage
	if len(disk) > 0 && disk[0] != "" {
		driver = ctx.storage.Disk(disk[0])
	}

	if driver.Missing(file) {
		return r.NotFound("文件不存在")
	}

	if mime, err := driver.MimeType(file); err == nil && mime != "" {
		ctx.c.Set("Content-Type", mime)
	}

	return ctx.c.SendFile(driver.Path(file))
}

// Download 以附件下载方式输出文件，可自定义下载文件名。
func (r *Response) Download(file string, name string, disk ...string) error {
	ctx, ok := r.ctx.(*fiberContext)
	if !ok || ctx.storage == nil {
		return r.Fail(http.StatusInternalServerError, "storage not available")
	}

	var driver contracts.StorageDriver = ctx.storage
	if len(disk) > 0 && disk[0] != "" {
		driver = ctx.storage.Disk(disk[0])
	}

	if driver.Missing(file) {
		return r.NotFound("文件不存在")
	}

	filename := name
	if filename == "" {
		filename = filepath.Base(file)
	}
	ctx.c.Attachment(filename)
	if mime, err := driver.MimeType(file); err == nil && mime != "" {
		ctx.c.Set("Content-Type", mime)
	}
	return ctx.c.SendFile(driver.Path(file))
}

// Success 快速返回成功响应（HTTP 200, code=0）。
func (r *Response) Success(data any, message ...string) error {
	msg := "ok"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return r.Build(http.StatusOK, 0, msg, data)
}

// Fail 快速返回失败响应（默认业务码 code=1）。
func (r *Response) Fail(status int, message string, code ...int) error {
	bizCode := 0
	if len(code) > 0 {
		bizCode = code[0]
	}
	return r.Build(status, bizCode, message, nil)
}

// Created 快速返回创建成功响应（HTTP 201, code=0）。
func (r *Response) Created(data any, message ...string) error {
	msg := "ok"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return r.Build(http.StatusCreated, 0, msg, data)
}

// Unauthorized 快速返回未授权响应（HTTP 401）。
func (r *Response) Unauthorized(message ...string) error {
	msg := "未授权"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return r.Fail(http.StatusUnauthorized, msg)
}

// Forbidden 快速返回无权限响应（HTTP 403）。
func (r *Response) Forbidden(message ...string) error {
	msg := "无权限访问"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return r.Fail(http.StatusForbidden, msg)
}

// NotFound 快速返回资源不存在响应（HTTP 404）。
func (r *Response) NotFound(message ...string) error {
	msg := "资源不存在"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return r.Fail(http.StatusNotFound, msg)
}

// Validation 快速返回参数验证失败响应（HTTP 422）。
// 默认 message 为“参数验证失败”，err.Error() 放入 data，便于前端调试/展示。
func (r *Response) Validation(err error, message ...string) error {
	msg := "参数验证失败"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	if err == nil {
		return r.Fail(http.StatusUnprocessableEntity, msg)
	}
	return r.Build(http.StatusUnprocessableEntity, 1, msg, map[string]any{
		"error": err.Error(),
	})
}

// Paginate 快速返回分页数据响应（HTTP 200, code=0）。
func (r *Response) Paginate(list any, total int64, page int, size int, message ...string) error {
	msg := "ok"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	return r.Build(http.StatusOK, 0, msg, map[string]any{
		"list":  list,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// Errorf 是一个可选辅助方法，便于把格式化错误快速转为 Validation/Fall 等上层调用。
func Errorf(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}

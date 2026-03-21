package controllers

import (
	"net/http"

	"go-fast/app/models"
	"go-fast/framework/contracts"
	"go-fast/framework/facades"
)

// UserController 用户资源控制器。
// 每个方法签名均为 func(contracts.Context) error，无 Fiber 依赖。
type UserController struct{}

// ── 请求体定义 ───────────────────────────────────────────────────────

// CreateUserRequest POST /api/v1/users 的请求体。
// binding tag 同时控制解析字段名和验证规则。
type CreateUserRequest struct {
	Name     string `json:"name"     binding:"required,min=2,max=50"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// UpdateUserRequest PUT /api/v1/users/:id 的请求体（字段均为可选）。
type UpdateUserRequest struct {
	Name  string `json:"name"  binding:"omitempty,min=2,max=50"`
	Email string `json:"email" binding:"omitempty,email"`
}

// ── 响应体定义 ───────────────────────────────────────────────────────

// apiResponse 统一 JSON 响应结构。
type apiResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func ok(ctx contracts.Context, data any) error {
	return ctx.JSON(http.StatusOK, apiResponse{Code: 0, Message: "ok", Data: data})
}

func created(ctx contracts.Context, data any) error {
	return ctx.JSON(http.StatusCreated, apiResponse{Code: 0, Message: "ok", Data: data})
}

func fail(ctx contracts.Context, code int, msg string) error {
	return ctx.JSON(code, apiResponse{Code: 1, Message: msg})
}

// ── 控制器方法 ───────────────────────────────────────────────────────

// Index GET /api/v1/users — 用户列表。
func (c UserController) Index(ctx contracts.Context) error {
	var users []models.User

	result := facades.Orm().DB().Order("created_at DESC").Find(&users)
	if result.Error != nil {
		facades.Log().Errorf("list users failed: %v", result.Error)
		return fail(ctx, http.StatusInternalServerError, "查询失败")
	}

	return ok(ctx, map[string]any{
		"list":  users,
		"total": result.RowsAffected,
	})
}

// Show GET /api/v1/users/:id — 用户详情。
func (c UserController) Show(ctx contracts.Context) error {
	id := ctx.Param("id")

	var user models.User
	if err := facades.Orm().DB().First(&user, "id = ?", id).Error; err != nil {
		return fail(ctx, http.StatusNotFound, "用户不存在")
	}

	return ok(ctx, user)
}

// Store POST /api/v1/users — 创建用户。
func (c UserController) Store(ctx contracts.Context) error {
	var req CreateUserRequest

	// Bind 同时完成：JSON 解析 + binding tag 验证
	if err := ctx.Bind(&req); err != nil {
		return fail(ctx, http.StatusUnprocessableEntity, err.Error())
	}

	// 检查邮箱唯一性
	var count int64
	facades.Orm().DB().Model(&models.User{}).Where("email = ?", req.Email).Count(&count)
	if count > 0 {
		return fail(ctx, http.StatusConflict, "邮箱已存在")
	}

	user := models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password, // 实际项目中应先哈希处理
	}

	if err := facades.Orm().DB().Create(&user).Error; err != nil {
		facades.Log().Errorf("create user failed: %v", err)
		return fail(ctx, http.StatusInternalServerError, "创建失败")
	}

	return created(ctx, user)
}

// Update PUT /api/v1/users/:id — 更新用户。
func (c UserController) Update(ctx contracts.Context) error {
	id := ctx.Param("id")

	var user models.User
	if err := facades.Orm().DB().First(&user, "id = ?", id).Error; err != nil {
		return fail(ctx, http.StatusNotFound, "用户不存在")
	}

	var req UpdateUserRequest
	if err := ctx.Bind(&req); err != nil {
		return fail(ctx, http.StatusUnprocessableEntity, err.Error())
	}

	updates := map[string]any{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}

	if len(updates) > 0 {
		if err := facades.Orm().DB().Model(&user).Updates(updates).Error; err != nil {
			return fail(ctx, http.StatusInternalServerError, "更新失败")
		}
	}

	return ok(ctx, user)
}

// Destroy DELETE /api/v1/users/:id — 删除用户。
func (c UserController) Destroy(ctx contracts.Context) error {
	id := ctx.Param("id")

	var user models.User
	if err := facades.Orm().DB().First(&user, "id = ?", id).Error; err != nil {
		return fail(ctx, http.StatusNotFound, "用户不存在")
	}

	if err := facades.Orm().DB().Delete(&user).Error; err != nil {
		return fail(ctx, http.StatusInternalServerError, "删除失败")
	}

	return ok(ctx, nil)
}

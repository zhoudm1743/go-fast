package controllers

import (
	"net/http"

	"go-fast/app/models"
	"go-fast/framework/contracts"
	"go-fast/framework/facades"
)

// UserController 前台用户自服务控制器。
// 用户只能查看/修改自己的信息，无法操作他人数据。
type UserController struct{}

// ── 请求体 ───────────────────────────────────────────────────────────

// UpdateProfileRequest 更新个人资料（仅修改自己）。
type UpdateProfileRequest struct {
	Name string `json:"name" binding:"omitempty,min=2,max=50"`
}

// ── 控制器方法 ────────────────────────────────────────────────────────

// Profile GET /api/v1/user/profile
// 读取当前登录用户的个人信息（user_id 由 Auth 中间件通过 ctx.WithValue 注入）。
func (c UserController) Profile(ctx contracts.Context) error {
	userID, ok2 := ctx.Value("user_id").(string)
	if !ok2 || userID == "" {
		return ctx.Response().Unauthorized("未登录")
	}

	var user models.User
	if err := facades.Orm().DB().First(&user, "id = ?", userID).Error; err != nil {
		return ctx.Response().NotFound("用户不存在")
	}
	return ctx.Response().Success(user)
}

// UpdateProfile PUT /api/v1/user/profile
// 更新当前登录用户的个人资料。
func (c UserController) UpdateProfile(ctx contracts.Context) error {
	userID, ok2 := ctx.Value("user_id").(string)
	if !ok2 || userID == "" {
		return ctx.Response().Unauthorized("未登录")
	}

	var req UpdateProfileRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.Response().Validation(err)
	}

	var user models.User
	if err := facades.Orm().DB().First(&user, "id = ?", userID).Error; err != nil {
		return ctx.Response().NotFound("用户不存在")
	}

	if req.Name != "" {
		if err := facades.Orm().DB().Model(&user).Update("name", req.Name).Error; err != nil {
			return ctx.Response().Fail(http.StatusInternalServerError, "更新失败")
		}
	}
	return ctx.Response().Success(user)
}

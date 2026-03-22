package controllers

import (
	"go-fast/app/http/admin/requests"
	"net/http"

	"go-fast/app/models"
	"go-fast/framework/contracts"
	"go-fast/framework/facades"
)

// UserController 前台用户自服务控制器。
// 用户只能查看/修改自己的信息，无法操作他人数据。
type UserController struct{}

// Prefix 路由前缀（实现 contracts.Prefixer）。
func (c *UserController) Prefix() string { return "/user" }

// Boot 声明路由（实现 contracts.Controller）。
func (c *UserController) Boot(r contracts.Route) {
	r.Get("/profile", c.Profile)
	r.Put("/profile", c.UpdateProfile)
}

// ── 控制器方法 ────────────────────────────────────────────────────────

// Profile GET /user/profile
// 读取当前登录用户的个人信息（user_id 由 Auth 中间件通过 ctx.WithValue 注入）。
func (c *UserController) Profile(ctx contracts.Context) error {
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

// UpdateProfile PUT /user/profile
// 更新当前登录用户的个人资料。
func (c *UserController) UpdateProfile(ctx contracts.Context) error {
	userID, ok2 := ctx.Value("user_id").(string)
	if !ok2 || userID == "" {
		return ctx.Response().Unauthorized("未登录")
	}

	var req requests.UpdateProfileRequest
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

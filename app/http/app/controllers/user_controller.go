package controllers

import (
	"errors"
	"net/http"

	"github.com/zhoudm1743/go-fast/app/http/admin/requests"
	"github.com/zhoudm1743/go-fast/app/models"
	"github.com/zhoudm1743/go-fast/framework/contracts"
	"github.com/zhoudm1743/go-fast/framework/facades"
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

// Profile GET /user/profile
func (c *UserController) Profile(ctx contracts.Context) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return ctx.Response().Unauthorized("未登录")
	}

	var user models.User
	if err := facades.DB().Query().First(&user, "id = ?", userID); err != nil {
		if errors.Is(err, contracts.ErrRecordNotFound) {
			return ctx.Response().NotFound("用户不存在")
		}
		return ctx.Response().Fail(http.StatusInternalServerError, "查询失败")
	}
	return ctx.Response().Success(user)
}

// UpdateProfile PUT /user/profile
func (c *UserController) UpdateProfile(ctx contracts.Context) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return ctx.Response().Unauthorized("未登录")
	}

	var req requests.UpdateProfileRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.Response().Validation(err)
	}

	var user models.User
	if err := facades.DB().Query().First(&user, "id = ?", userID); err != nil {
		if errors.Is(err, contracts.ErrRecordNotFound) {
			return ctx.Response().NotFound("用户不存在")
		}
		return ctx.Response().Fail(http.StatusInternalServerError, "查询失败")
	}

	if req.Name != "" {
		if err := facades.DB().Query().Model(&user).Update("name", req.Name); err != nil {
			return ctx.Response().Fail(http.StatusInternalServerError, "更新失败")
		}
	}
	return ctx.Response().Success(user)
}

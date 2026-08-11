package controllers

import (
	"errors"
	"net/http"

	"github.com/zhoudm1743/go-fast/example/app/models"
	"github.com/zhoudm1743/go-fast-framework/contracts"
	"github.com/zhoudm1743/go-fast-framework/facades"
)

// UserController 用户 CRUD 控制器示例。
type UserController struct{}

func (c *UserController) Prefix() string { return "/users" }

func (c *UserController) Boot(r contracts.Route) {
	r.Get("/", c.Index)
	r.Get("/:id", c.Show)
	r.Post("/", c.Store)
	r.Put("/:id", c.Update)
	r.Delete("/:id", c.Destroy)
}

// Index GET /users — 用户列表。
func (c *UserController) Index(ctx contracts.Context) error {
	var users []models.User
	if err := facades.DB().Query().Order("created_at DESC").Find(&users); err != nil {
		return ctx.Response().Fail(http.StatusInternalServerError, "查询失败")
	}
	return ctx.Response().Success(users)
}

// Show GET /users/:id — 用户详情。
func (c *UserController) Show(ctx contracts.Context) error {
	id := ctx.Param("id")
	var user models.User
	if err := facades.DB().Query().First(&user, "id = ?", id); err != nil {
		if errors.Is(err, contracts.ErrRecordNotFound) {
			return ctx.Response().NotFound("用户不存在")
		}
		return ctx.Response().Fail(http.StatusInternalServerError, "查询失败")
	}
	return ctx.Response().Success(user)
}

// Store POST /users — 创建用户。
func (c *UserController) Store(ctx contracts.Context) error {
	var req struct {
		Name  string `json:"name"  binding:"required,min=2,max=50"`
		Email string `json:"email" binding:"required,email"`
	}
	if err := ctx.Bind(&req); err != nil {
		return ctx.Response().Validation(err)
	}

	user := models.User{Name: req.Name, Email: req.Email}
	if err := facades.DB().Query().Create(&user); err != nil {
		if errors.Is(err, contracts.ErrDuplicatedKey) {
			return ctx.Response().Fail(http.StatusConflict, "邮箱已存在")
		}
		return ctx.Response().Fail(http.StatusInternalServerError, "创建失败")
	}
	return ctx.Response().Created(user)
}

// Update PUT /users/:id — 更新用户。
func (c *UserController) Update(ctx contracts.Context) error {
	id := ctx.Param("id")
	var req struct {
		Name  string `json:"name"  binding:"omitempty,min=2,max=50"`
		Email string `json:"email" binding:"omitempty,email"`
	}
	if err := ctx.Bind(&req); err != nil {
		return ctx.Response().Validation(err)
	}

	var user models.User
	if err := facades.DB().Query().First(&user, "id = ?", id); err != nil {
		if errors.Is(err, contracts.ErrRecordNotFound) {
			return ctx.Response().NotFound("用户不存在")
		}
		return ctx.Response().Fail(http.StatusInternalServerError, "查询失败")
	}

	updates := map[string]any{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if len(updates) > 0 {
		if err := facades.DB().Query().Model(&user).Updates(updates); err != nil {
			return ctx.Response().Fail(http.StatusInternalServerError, "更新失败")
		}
	}
	return ctx.Response().Success(user)
}

// Destroy DELETE /users/:id — 删除用户。
func (c *UserController) Destroy(ctx contracts.Context) error {
	id := ctx.Param("id")
	var user models.User
	if err := facades.DB().Query().First(&user, "id = ?", id); err != nil {
		if errors.Is(err, contracts.ErrRecordNotFound) {
			return ctx.Response().NotFound("用户不存在")
		}
		return ctx.Response().Fail(http.StatusInternalServerError, "查询失败")
	}
	if err := facades.DB().Query().Delete(&user); err != nil {
		return ctx.Response().Fail(http.StatusInternalServerError, "删除失败")
	}
	return ctx.Response().Success(nil)
}

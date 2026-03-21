package controllers

import (
	"net/http"

	"go-fast/app/models"
	"go-fast/framework/contracts"
	"go-fast/framework/facades"
)

// UserController 后台用户管理控制器（完整 CRUD）。
type UserController struct{}

// ── 请求体 ───────────────────────────────────────────────────────────

type ListUserRequest struct {
	Page  int    `query:"page"  binding:"omitempty,gte=1"`
	Size  int    `query:"size"  binding:"omitempty,gte=1,lte=100"`
	Email string `query:"email" binding:"omitempty,email"`
}

type UserIDRequest struct {
	ID string `uri:"id" binding:"required"`
}

type CreateUserRequest struct {
	Name     string `json:"name"     binding:"required,min=2,max=50"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type UpdateUserRequest struct {
	ID    string `uri:"id"     binding:"required"`
	Name  string `json:"name"  binding:"omitempty,min=2,max=50"`
	Email string `json:"email" binding:"omitempty,email"`
}

// ── 控制器方法 ────────────────────────────────────────────────────────

// Index GET /admin/api/v1/users?page=1&size=20&email=xxx
func (c UserController) Index(ctx contracts.Context) error {
	var req ListUserRequest
	if err := ctx.Bind(&req); err != nil { // query tag 自动填充
		return ctx.Response().Validation(err)
	}
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Size == 0 {
		req.Size = 20
	}

	db := facades.Orm().DB().Model(&models.User{}).Order("created_at DESC")
	if req.Email != "" {
		db = db.Where("email LIKE ?", "%"+req.Email+"%")
	}

	var total int64
	db.Count(&total)

	var users []models.User
	if err := db.Offset((req.Page - 1) * req.Size).Limit(req.Size).Find(&users).Error; err != nil {
		facades.Log().Errorf("admin list users: %v", err)
		return ctx.Response().Fail(http.StatusInternalServerError, "查询失败")
	}

	return ctx.Response().Paginate(users, total, req.Page, req.Size)
}

// Show GET /admin/api/v1/users/:id
func (c UserController) Show(ctx contracts.Context) error {
	var req UserIDRequest
	if err := ctx.Bind(&req); err != nil { // uri tag 自动填充
		return ctx.Response().Validation(err)
	}

	var user models.User
	if err := facades.Orm().DB().First(&user, "id = ?", req.ID).Error; err != nil {
		return ctx.Response().NotFound("用户不存在")
	}
	return ctx.Response().Success(user)
}

// Store POST /admin/api/v1/users
func (c UserController) Store(ctx contracts.Context) error {
	var req CreateUserRequest
	if err := ctx.Bind(&req); err != nil { // json body 自动填充
		return ctx.Response().Validation(err)
	}

	var count int64
	facades.Orm().DB().Model(&models.User{}).Where("email = ?", req.Email).Count(&count)
	if count > 0 {
		return ctx.Response().Fail(http.StatusConflict, "邮箱已存在")
	}

	user := models.User{Name: req.Name, Email: req.Email, Password: req.Password}
	if err := facades.Orm().DB().Create(&user).Error; err != nil {
		facades.Log().Errorf("admin create user: %v", err)
		return ctx.Response().Fail(http.StatusInternalServerError, "创建失败")
	}
	return ctx.Response().Created(user)
}

// Update PUT /admin/api/v1/users/:id  （uri + json body 混合绑定）
func (c UserController) Update(ctx contracts.Context) error {
	var req UpdateUserRequest
	if err := ctx.Bind(&req); err != nil { // uri 填充 ID + json 填充 Name/Email
		return ctx.Response().Validation(err)
	}

	var user models.User
	if err := facades.Orm().DB().First(&user, "id = ?", req.ID).Error; err != nil {
		return ctx.Response().NotFound("用户不存在")
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
			return ctx.Response().Fail(http.StatusInternalServerError, "更新失败")
		}
	}
	return ctx.Response().Success(user)
}

// Destroy DELETE /admin/api/v1/users/:id
func (c UserController) Destroy(ctx contracts.Context) error {
	var req UserIDRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.Response().Validation(err)
	}

	var user models.User
	if err := facades.Orm().DB().First(&user, "id = ?", req.ID).Error; err != nil {
		return ctx.Response().NotFound("用户不存在")
	}

	if err := facades.Orm().DB().Delete(&user).Error; err != nil {
		return ctx.Response().Fail(http.StatusInternalServerError, "删除失败")
	}
	return ctx.Response().Success(nil)
}

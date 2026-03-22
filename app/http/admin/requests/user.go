package requests

// ── 请求体 ───────────────────────────────────────────────────────────

// UpdateProfileRequest 更新个人资料（仅修改自己）。
type UpdateProfileRequest struct {
	Name string `json:"name" binding:"omitempty,min=2,max=50"`
}

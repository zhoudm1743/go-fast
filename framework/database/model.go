package database

// Model 基础模型，所有业务模型应嵌入此结构体。
// ID 为 UUID v7 字符串主键，由 ORM 回调自动生成。
type Model struct {
	ID        string `gorm:"primaryKey;size:36;column:id" json:"id"`
	CreatedAt int64  `gorm:"autoCreateTime;column:created_at" json:"created_at"`
	UpdatedAt int64  `gorm:"autoUpdateTime;column:updated_at" json:"updated_at"`
}

// ModelWithSoftDelete 带软删除的基础模型。
type ModelWithSoftDelete struct {
	Model
	DeletedAt int64 `gorm:"column:deleted_at;index;default:0" json:"deleted_at"`
}

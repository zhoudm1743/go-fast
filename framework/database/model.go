package database

import (
	"github.com/zhoudm1743/go-fast/framework/contracts"

	"github.com/google/uuid"
)

// Model 基础模型，所有业务模型应嵌入此结构体。
// ID 为 UUID v7 字符串主键，由框架层 BeforeCreate Hook 自动生成。
type Model struct {
	ID        string `gormdriver:"primaryKey;size:36;column:id"      xorm:"pk varchar(36) 'id'"    json:"id"`
	CreatedAt int64  `gormdriver:"autoCreateTime;column:created_at"  xorm:"created 'created_at'"   json:"created_at"`
	UpdatedAt int64  `gormdriver:"autoUpdateTime;column:updated_at"  xorm:"updated 'updated_at'"   json:"updated_at"`
}

// BeforeCreate 框架层 Hook：创建前自动生成 UUID v7 主键。
// 各 ORM 驱动在 Create 前检查目标对象是否实现 BeforeCreator，如有则调用。
func (m *Model) BeforeCreate(_ contracts.Query) error {
	if m.ID == "" {
		m.ID = uuid.Must(uuid.NewV7()).String()
	}
	return nil
}

// ModelWithSoftDelete 带软删除的基础模型。
type ModelWithSoftDelete struct {
	Model
	DeletedAt int64 `gormdriver:"column:deleted_at;index;default:0" xorm:"'deleted_at' index default(0)" json:"deleted_at"`
}

// ── 模型钩子接口（各驱动在对应阶段调用）────────────────────────────────

// BeforeCreator 创建前钩子
type BeforeCreator interface {
	BeforeCreate(q contracts.Query) error
}

// AfterCreator 创建后钩子
type AfterCreator interface {
	AfterCreate(q contracts.Query) error
}

// BeforeUpdater 更新前钩子
type BeforeUpdater interface {
	BeforeUpdate(q contracts.Query) error
}

// AfterUpdater 更新后钩子
type AfterUpdater interface {
	AfterUpdate(q contracts.Query) error
}

// BeforeDeleter 删除前钩子
type BeforeDeleter interface {
	BeforeDelete(q contracts.Query) error
}

// AfterDeleter 删除后钩子
type AfterDeleter interface {
	AfterDelete(q contracts.Query) error
}

// AfterFinder 查询后钩子
type AfterFinder interface {
	AfterFind(q contracts.Query) error
}

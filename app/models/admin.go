package models

import "github.com/zhoudm1743/go-fast/framework/database"

// Admin 模型。
// 嵌入 database.Model 自动获得 UUID v7 主键、CreatedAt、UpdatedAt。
type Admin struct {
	database.Model
	// TODO: 在此添加字段
	// Name string `gorm:"size:100;not null" json:"name"`
}

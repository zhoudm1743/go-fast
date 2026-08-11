package models

import "github.com/zhoudm1743/go-fast-framework/database"

// User 用户模型示例。
type User struct {
	database.Model
	Name  string `json:"name"  gorm:"size:100;not null"`
	Email string `json:"email" gorm:"size:200;uniqueIndex;not null"`
}

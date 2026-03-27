package models

import "github.com/zhoudm1743/go-fast/framework/database"

// User 用户模型。
// 嵌入 database.Model 自动获得 UUID v7 主键、CreatedAt、UpdatedAt。
type User struct {
	database.Model
	Name     string `gormdriver:"size:100;not null"        json:"name"`
	Email    string `gormdriver:"size:200;uniqueIndex;not null" json:"email"`
	Password string `gormdriver:"size:255;not null"        json:"-"` // 不输出密码
}

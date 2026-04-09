package seeders

import (
	"github.com/zhoudm1743/go-fast/app/models"
	"github.com/zhoudm1743/go-fast/framework/contracts"
	"golang.org/x/crypto/bcrypt"
)

// UserSeeder 用户表种子数据。
// 仅在数据库中无用户时执行，保证幂等性。
type UserSeeder struct{}

func (s *UserSeeder) Run(q contracts.Query) error {
	var count int64
	q.Model(&models.User{}).Count(&count)
	if count > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	users := []models.User{
		{Name: "Alice", Email: "alice@example.com", Password: string(hash)},
		{Name: "Bob", Email: "bob@example.com", Password: string(hash)},
	}

	return q.Create(&users)
}

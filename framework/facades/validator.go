package facades

import "go-fast/framework/contracts"

// Validator 获取验证服务实例。
func Validator() contracts.Validation {
	return App().MustMake("validator").(contracts.Validation)
}

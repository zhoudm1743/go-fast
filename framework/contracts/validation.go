package contracts

// Validation 验证服务契约。
type Validation interface {
	// Validate 验证结构体，返回验证错误。
	Validate(obj any) error
	// RegisterRule 注册自定义验证规则。
	RegisterRule(rule any) error
}

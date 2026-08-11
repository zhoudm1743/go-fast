package jobs

import (
	"github.com/zhoudm1743/go-fast-framework/facades"
)

// ProcessOrder 处理订单队列任务示例。
type ProcessOrder struct{}

func (j *ProcessOrder) Signature() string {
	return "process_order"
}

func (j *ProcessOrder) Handle(args ...any) error {
	facades.Log().Infof("[Job] ProcessOrder executed with args: %v", args)
	return nil
}

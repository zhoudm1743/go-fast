package jobs

import "fmt"

// ProcessOrder 处理订单队列任务示例。
type ProcessOrder struct{}

func (j *ProcessOrder) Signature() string {
	return "process_order"
}

func (j *ProcessOrder) Handle(args ...any) error {
	fmt.Println("[Job] ProcessOrder executed with args:", args)
	return nil
}

package queue

import (
	"github.com/zhoudm1743/go-fast/framework/contracts"
	"github.com/zhoudm1743/go-fast/framework/foundation"
)

// ServiceProvider 队列服务提供者。
type ServiceProvider struct{}

func (sp *ServiceProvider) Register(app foundation.Application) {
	app.Singleton("queue", func(app foundation.Application) (any, error) {
		return New(), nil
	})
}

func (sp *ServiceProvider) Boot(app foundation.Application) error {
	return nil
}

// RegisterJobs 在引导完成后，通过 app 注册任务类。
// 通常在 bootstrap/app.go 的 Boot() 函数中调用。
func RegisterJobs(app foundation.Application, jobs []contracts.QueueJob) {
	q := app.MustMake("queue").(contracts.Queue)
	q.Register(jobs)
}

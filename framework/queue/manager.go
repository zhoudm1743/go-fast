package queue

import (
	"fmt"
	"sync"
	"time"

	"github.com/zhoudm1743/go-fast/framework/contracts"
)

// manager 队列管理器实现。
type manager struct {
	mu   sync.RWMutex
	jobs map[string]contracts.QueueJob // signature → job
}

// New 创建队列管理器（同步驱动）。
func New() contracts.Queue {
	return &manager{
		jobs: make(map[string]contracts.QueueJob),
	}
}

func (m *manager) Register(jobs []contracts.QueueJob) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, j := range jobs {
		m.jobs[j.Signature()] = j
	}
}

func (m *manager) Job(job contracts.QueueJob, args []contracts.QueueArg) contracts.QueuePending {
	return &pendingTask{
		mgr:  m,
		jobs: []contracts.QueueChain{{Job: job, Args: args}},
	}
}

func (m *manager) Chain(jobs []contracts.QueueChain) contracts.QueuePending {
	return &pendingTask{
		mgr:  m,
		jobs: jobs,
	}
}

// executeChain 按顺序执行任务链；任一失败则终止。
func executeChain(chain []contracts.QueueChain) error {
	for _, item := range chain {
		anyArgs := argsToAny(item.Args)
		if err := item.Job.Handle(anyArgs...); err != nil {
			return fmt.Errorf("[GoFast] queue job %s failed: %w", item.Job.Signature(), err)
		}
	}
	return nil
}

func argsToAny(args []contracts.QueueArg) []any {
	result := make([]any, len(args))
	for i, a := range args {
		result[i] = a.Value
	}
	return result
}

// ── pendingTask ──────────────────────────────────────────────────────

type pendingTask struct {
	mgr        *manager
	jobs       []contracts.QueueChain
	queue      string
	connection string
	delay      time.Time
}

func (p *pendingTask) OnQueue(queue string) contracts.QueuePending {
	p.queue = queue
	return p
}

func (p *pendingTask) OnConnection(connection string) contracts.QueuePending {
	p.connection = connection
	return p
}

func (p *pendingTask) Delay(delay time.Time) contracts.QueuePending {
	p.delay = delay
	return p
}

// Dispatch 推送到后台队列（当前：同步驱动直接 goroutine 执行；支持 Delay）。
func (p *pendingTask) Dispatch() error {
	go func() {
		if !p.delay.IsZero() {
			wait := time.Until(p.delay)
			if wait > 0 {
				time.Sleep(wait)
			}
		}
		if err := executeChain(p.jobs); err != nil {
			fmt.Printf("[GoFast] queue dispatch error: %v\n", err)
		}
	}()
	return nil
}

// DispatchSync 同步执行，不走队列，忽略 Delay。
func (p *pendingTask) DispatchSync() error {
	return executeChain(p.jobs)
}

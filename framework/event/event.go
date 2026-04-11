package event

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/zhoudm1743/go-fast/framework/contracts"
)

// manager 事件总线实现。
type manager struct {
	mu       sync.RWMutex
	handlers map[string][]contracts.EventListener // event type key → listeners
	events   map[string]contracts.Eventer         // event type key → event
}

// New 创建事件管理器。
func New() contracts.Event {
	return &manager{
		handlers: make(map[string][]contracts.EventListener),
		events:   make(map[string]contracts.Eventer),
	}
}

func eventKey(e contracts.Eventer) string {
	t := reflect.TypeOf(e)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.PkgPath() + "." + t.Name()
}

func (m *manager) Register(events map[contracts.Eventer][]contracts.EventListener) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for e, listeners := range events {
		key := eventKey(e)
		m.events[key] = e
		m.handlers[key] = append(m.handlers[key], listeners...)
	}
}

func (m *manager) Job(event contracts.Eventer, args []contracts.EventArg) contracts.EventPending {
	return &pendingEvent{mgr: m, event: event, args: args}
}

// dispatch 执行事件派发（由 pendingEvent 调用）。
func (m *manager) dispatch(event contracts.Eventer, args []contracts.EventArg) error {
	key := eventKey(event)

	// 执行事件 Handle（数据加工）
	processed, err := event.Handle(args)
	if err != nil {
		return fmt.Errorf("[GoFast] event %s Handle error: %w", key, err)
	}

	m.mu.RLock()
	listeners := m.handlers[key]
	m.mu.RUnlock()

	// 将 []EventArg 转换为 []any 传给监听器
	anyArgs := make([]any, len(processed))
	for i, a := range processed {
		anyArgs[i] = a.Value
	}

	for _, l := range listeners {
		queue := l.Queue(anyArgs...)
		if queue.Enable {
			// 异步队列处理（简化实现：goroutine；生产环境可接入 Queue 系统）
			go func(listener contracts.EventListener, a []any) {
				if err := listener.Handle(a...); err != nil {
					// 仅记录错误，不中断其他监听器
					fmt.Printf("[GoFast] event listener %s error: %v\n", listener.Signature(), err)
				}
			}(l, anyArgs)
		} else {
			if err := l.Handle(anyArgs...); err != nil {
				// 同步模式：有错误则停止向后传播
				return fmt.Errorf("[GoFast] event listener %s error: %w", l.Signature(), err)
			}
		}
	}
	return nil
}

// ── pendingEvent ─────────────────────────────────────────────────────

type pendingEvent struct {
	mgr   *manager
	event contracts.Eventer
	args  []contracts.EventArg
}

func (p *pendingEvent) Dispatch() error {
	return p.mgr.dispatch(p.event, p.args)
}

package foundation

// ServiceProvider 服务提供者接口。
// Register 阶段绑定服务到容器（不可使用其他服务）；
// Boot 阶段所有 Provider 已 Register 完成，可安全使用其他服务。
type ServiceProvider interface {
	// Register 将服务绑定到容器。此时其他服务可能尚未就绪，不可调用 MustMake。
	Register(app Application)
	// Boot 引导服务。所有 Provider 的 Register 均已执行完毕，可放心使用容器中的服务。
	Boot(app Application) error
}

// DeferredProvider 延迟服务提供者。
// 实现此接口的 Provider 在 Boot 阶段不会立即执行，而是等到
// 首次 Make 其声明的 key 时才自动触发 Register + Boot。
type DeferredProvider interface {
	ServiceProvider
	// DeferredServices 返回该 Provider 提供的服务 key 列表。
	DeferredServices() []string
}

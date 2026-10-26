package middleware

import (
	"sync"

	logincontroller "github.com/zhoudm1743/go-fast/app/window/login/controller"
	gofastfyne "github.com/zhoudm1743/gofast-fyne"
)

// 会话状态：示例用进程内变量，真实场景应换框架 cache / jwt。
var (
	mu      sync.RWMutex
	current string
)

// Login 记录登录用户。
func Login(user string) {
	mu.Lock()
	defer mu.Unlock()
	current = user
}

// Logout 退出登录。
func Logout() {
	mu.Lock()
	defer mu.Unlock()
	current = ""
}

// CurrentUser 返回当前登录用户；空串表示未登录。
func CurrentUser() string {
	mu.RLock()
	defer mu.RUnlock()
	return current
}

// Auth 窗口级中间件：已登录直接执行 next；
// 未登录打开登录窗口，登录成功后自动继续。
func Auth(g gofastfyne.Gui, next func()) func() {
	return func() {
		if CurrentUser() != "" {
			next()
			return
		}
		g.OpenWindow("login", "登录", logincontroller.New(func(user string) {
			Login(user)
			_ = g.Emit("app.logged_in", map[string]any{"user": user})
			next()
		}).View(g))
	}
}

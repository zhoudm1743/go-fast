package middleware

import (
	"time"

	"github.com/zhoudm1743/go-fast-framework/facades"
	logincontroller "github.com/zhoudm1743/go-fast/app/window/login/controller"
	gofastfyne "github.com/zhoudm1743/gofast-fyne"
)

// sessionKey 桌面会话在框架缓存中的键。
const sessionKey = "window:session:user"

// sessionTTL 会话有效期，默认对齐框架 HTTP 会话的 session.lifetime（秒）。
func sessionTTL() time.Duration {
	sec := facades.Config().GetInt("session.lifetime", 7200)
	return time.Duration(sec) * time.Second
}

// Login 将会话写入框架缓存（默认 file 驱动；改配置即可切 memory/redis）。
// 写入失败返回错误，调用方不应视为已登录。
func Login(user string) error {
	return facades.Cache().Put(sessionKey, user, sessionTTL())
}

// Logout 注销会话（删除缓存键）。
func Logout() error {
	return facades.Cache().Forget(sessionKey)
}

// CurrentUser 返回当前登录用户；未登录或会话已过期返回空串。
func CurrentUser() string {
	return facades.Cache().GetString(sessionKey, "")
}

// Auth 窗口级中间件：已登录直接执行 next；
// 未登录打开登录窗口，登录成功后自动继续。
// 会话写缓存失败时记录错误且不继续（保持未登录态，避免半登录状态）。
func Auth(g gofastfyne.Gui, next func()) func() {
	return func() {
		if CurrentUser() != "" {
			next()
			return
		}
		g.OpenWindow("login", "登录", logincontroller.New(func(user string) {
			if err := Login(user); err != nil {
				facades.Log().Errorf("[window] 会话写入缓存失败: %v", err)
				return
			}
			_ = g.Emit("app.logged_in", map[string]any{"user": user})
			next()
		}).View(g))
	}
}

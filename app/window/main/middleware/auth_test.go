package middleware

import (
	"testing"

	"github.com/zhoudm1743/go-fast-framework/cache"
	"github.com/zhoudm1743/go-fast-framework/config"
	"github.com/zhoudm1743/go-fast-framework/facades"
	"github.com/zhoudm1743/go-fast-framework/foundation"
	"github.com/zhoudm1743/go-fast-framework/log"
)

// newTestApp 引导含缓存的最小应用（file 驱动写入临时目录，不污染仓库）。
func newTestApp(t *testing.T) foundation.Application {
	t.Helper()
	app := foundation.NewApplication(t.TempDir())
	app.SetProviders([]foundation.ServiceProvider{
		&config.ServiceProvider{},
		&log.ServiceProvider{},
		&cache.ServiceProvider{},
	})
	app.Boot()
	facades.SetApp(app)
	return app
}

func TestSession_LoginCurrentLogout(t *testing.T) {
	newTestApp(t)

	if got := CurrentUser(); got != "" {
		t.Fatalf("初始应未登录，实际 %q", got)
	}
	if err := Login("alice"); err != nil {
		t.Fatalf("Login 失败: %v", err)
	}
	if got := CurrentUser(); got != "alice" {
		t.Fatalf("CurrentUser = %q, 期望 alice", got)
	}
	if err := Logout(); err != nil {
		t.Fatalf("Logout 失败: %v", err)
	}
	if got := CurrentUser(); got != "" {
		t.Fatalf("注销后应为空，实际 %q", got)
	}
}

func TestSession_Overwrite(t *testing.T) {
	newTestApp(t)

	_ = Login("alice")
	if err := Login("bob"); err != nil {
		t.Fatal(err)
	}
	if got := CurrentUser(); got != "bob" {
		t.Fatalf("重复登录应覆盖会话，实际 %q", got)
	}
}

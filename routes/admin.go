package routes

import (
	"html/template"
	"strings"
	"time"

	adminControllers "github.com/zhoudm1743/go-fast/app/http/admin/controllers"
	adminMiddleware "github.com/zhoudm1743/go-fast/app/http/admin/middleware"
	"github.com/zhoudm1743/go-fast/framework/contracts"
	"github.com/zhoudm1743/go-fast/framework/facades"
)

// RegisterAdmin 注册后台管理路由。
func RegisterAdmin() {
	r := facades.Http.Route()

	// ── 静态资源（无需认证） ─────────────────────────
	// 访问示例: GET /static/admin/style.css
	r.Static("/static", "resources/static")

	// ── 向视图引擎注入自定义模板函数 ───────────
	if ve, err := facades.App().Make("view"); err == nil && ve != nil {
		ve.(contracts.ViewEngine).AddFuncMap(template.FuncMap{
			// inc 实现 1-based 序号：{{inc $i}}
			"inc": func(i int) int { return i + 1 },
			// upper 转大写
			"upper": strings.ToUpper,
			// now 返回当前时间
			"now": func(format ...string) string {
				f := "2006-01-02 15:04:05"
				if len(format) > 0 {
					f = format[0]
				}
				return time.Now().Format(f)
			},
		})
	}

	// ── HTML 演示页面（无需认证，直接浏览器开开） ────────
	// 访问示例: GET /pages/dashboard
	r.Group("/pages", func(pages contracts.Route) {
		pages.Register(&adminControllers.PageController{})
	})

	// ── JSON API（需要管理员认证） ────────────────────
	r.Group("/admin", adminMiddleware.AdminAuth, func(admin contracts.Route) {
		admin.Register(
			&adminControllers.UserController{},
		)
	})
}

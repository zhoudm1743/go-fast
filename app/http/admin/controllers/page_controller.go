package controllers

import (
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/zhoudm1743/go-fast/app/models"
	"github.com/zhoudm1743/go-fast/framework/contracts"
	"github.com/zhoudm1743/go-fast/framework/facades"
)

// PageController 演示 HTML 模板渲染功能。
//
// 访问地址（默认端口 3000）：
//
//	GET http://localhost:3000/pages/dashboard  控制面板
//	GET http://localhost:3000/pages/users      用户列表（HTML 版）
//
// ────────────────────────────────────────────────────────────────────
// go:embed 嵌入打包场景
// ────────────────────────────────────────────────────────────────────
// 生产环境可将模板 + 静态资源一起打进二进制，无需部署额外文件夹：
//
//  1. 在本文件（或 main.go）顶部声明：
//
//     //go:embed ../../../resources/views
//     var viewFS embed.FS
//
//     //go:embed ../../../resources/static
//     var staticFS embed.FS
//
//  2. 在自定义 ServiceProvider 中覆盖默认引擎：
//
//     import (
//         "io/fs"
//         "net/http"
//         goview "github.com/zhoudm1743/go-fast/framework/http/view"
//     )
//
//     func (sp *EmbedProvider) Register(app foundation.Application) {
//         sub, _ := fs.Sub(viewFS, "resources/views")
//         app.Instance("view", goview.New(".", goview.WithFS(sub)))
//     }
//
//     func (sp *EmbedProvider) Boot(app foundation.Application) error {
//         sub, _ := fs.Sub(staticFS, "resources/static")
//         facades.Http.Route().StaticFS("/static", http.FS(sub))
//         return nil
//     }
//
// ────────────────────────────────────────────────────────────────────

// PageController 实现 contracts.Controller。
type PageController struct{}

// Prefix 路由前缀（实现 contracts.Prefixer）。
func (c *PageController) Prefix() string { return "" }

// Boot 声明路由（实现 contracts.Controller）。
func (c *PageController) Boot(r contracts.Route) {
	r.Get("/dashboard", c.Dashboard)
	r.Get("/users", c.Users)
}

// ─── 共享模板函数 ────────────────────────────────────────────────────

// viewFuncMap 注册模板辅助函数，在引擎初始化后通过 ServiceProvider 注入。
// 此处暂提供示例，实际注入见 bootstrap/app.go 注释。
func viewFuncMap() template.FuncMap {
	return template.FuncMap{
		// inc 在模板中实现序号（{{inc $i}} 得到 1-based 索引）
		"inc": func(i int) int { return i + 1 },
		// upper 转大写
		"upper": strings.ToUpper,
		// now 返回当前时间，格式可选
		"now": func(format ...string) string {
			f := "2006-01-02 15:04:05"
			if len(format) > 0 {
				f = format[0]
			}
			return time.Now().Format(f)
		},
	}
}

// ─── 控制器方法 ──────────────────────────────────────────────────────

// statsData 统计面板的数据结构。
type statsData struct {
	TotalUsers  int64
	ActiveUsers int64
	TodayNew    int64
	Disabled    int64
}

// recentUserRow 模板中的用户行数据。
type recentUserRow struct {
	Name      string
	Email     string
	CreatedAt string
	Active    bool
}

// Dashboard GET /pages/dashboard
func (c *PageController) Dashboard(ctx contracts.Context) error {
	var total, todayNew int64
	facades.DB().Query().Model(&models.User{}).Count(&total)
	facades.DB().Query().Model(&models.User{}).
		Where("created_at >= ?", time.Now().Truncate(24*time.Hour)).
		Count(&todayNew)

	// 最近 5 条用户记录
	var users []models.User
	_ = facades.DB().Query().Model(&models.User{}).Order("created_at DESC").Limit(5).Find(&users)

	rows := make([]recentUserRow, len(users))
	for i, u := range users {
		rows[i] = recentUserRow{
			Name:      u.Name,
			Email:     u.Email,
			CreatedAt: time.Unix(u.CreatedAt, 0).Format("2006-01-02 15:04"),
			Active:    true,
		}
	}

	return ctx.HTML(http.StatusOK, "admin/dashboard.html", map[string]any{
		"Title":  "控制面板",
		"Active": "dashboard",
		"Stats": statsData{
			TotalUsers:  total,
			ActiveUsers: total,
			TodayNew:    todayNew,
			Disabled:    0,
		},
		"RecentUsers": rows,
	})
}

// Users GET /pages/users
func (c *PageController) Users(ctx contracts.Context) error {
	var users []models.User
	_ = facades.DB().Query().Model(&models.User{}).Order("created_at DESC").Find(&users)

	rows := make([]recentUserRow, len(users))
	for i, u := range users {
		rows[i] = recentUserRow{
			Name:      u.Name,
			Email:     u.Email,
			CreatedAt: time.Unix(u.CreatedAt, 0).Format("2006-01-02 15:04"),
			Active:    true,
		}
	}

	return ctx.HTML(http.StatusOK, "admin/users.html", map[string]any{
		"Title":  "用户管理",
		"Active": "users",
		"Users":  rows,
	})
}

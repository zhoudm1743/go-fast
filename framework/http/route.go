package http

import (
	"context"
	"fmt"
	"time"

	"go-fast/framework/contracts"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

// route 实现 contracts.Route，封装 Fiber。
// 应用代码通过 contracts.HandlerFunc / contracts.Context 与框架交互，
// 不直接依赖任何 Fiber 类型。
type route struct {
	app       *fiber.App
	cfg       contracts.Config
	router    fiber.Router
	validator contracts.Validation
}

// NewRoute 创建 HTTP 路由服务实例。
func NewRoute(cfg contracts.Config, validator contracts.Validation) (contracts.Route, error) {
	readTimeout := time.Duration(cfg.GetInt("server.read_timeout_sec", 30)) * time.Second
	writeTimeout := time.Duration(cfg.GetInt("server.write_timeout_sec", 30)) * time.Second
	idleTimeout := time.Duration(cfg.GetInt("server.idle_timeout_sec", 120)) * time.Second
	name := cfg.GetString("server.name", "GoFast")

	fiberCfg := fiber.Config{
		AppName:               name,
		ServerHeader:          name,
		DisableStartupMessage: false,
		ReadTimeout:           readTimeout,
		WriteTimeout:          writeTimeout,
		IdleTimeout:           idleTimeout,
		Prefork:               cfg.GetBool("server.prefork"),
	}
	if limit := cfg.GetInt("server.body_limit_mb"); limit > 0 {
		fiberCfg.BodyLimit = limit * 1024 * 1024
	}

	app := fiber.New(fiberCfg)

	mode := cfg.GetString("server.mode", "debug")
	app.Use(recover.New(recover.Config{EnableStackTrace: mode != "release"}))
	app.Use(requestid.New())

	allowOrigins := "*"
	if originsRaw := cfg.Get("server.cors_allow_origins"); originsRaw != nil {
		if origins, ok := originsRaw.([]any); ok && len(origins) > 0 {
			allowOrigins = ""
			for i, o := range origins {
				if i > 0 {
					allowOrigins += ","
				}
				allowOrigins += fmt.Sprintf("%v", o)
			}
		}
	}
	app.Use(cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: allowOrigins != "*",
		MaxAge:           86400,
	}))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	return &route{app: app, cfg: cfg, router: app, validator: validator}, nil
}

func (r *route) Run(addr ...string) error {
	address := fmt.Sprintf("%s:%d",
		r.cfg.GetString("server.host", "0.0.0.0"),
		r.cfg.GetInt("server.port", 3000))
	if len(addr) > 0 && addr[0] != "" {
		address = addr[0]
	}
	return r.app.Listen(address)
}

func (r *route) Shutdown() error {
	timeout := time.Duration(r.cfg.GetInt("server.shutdown_timeout_sec", 10)) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return r.app.ShutdownWithContext(ctx)
}

func (r *route) Get(path string, h contracts.HandlerFunc) contracts.Route {
	r.router.Get(path, r.wrap(h))
	return r
}
func (r *route) Post(path string, h contracts.HandlerFunc) contracts.Route {
	r.router.Post(path, r.wrap(h))
	return r
}
func (r *route) Put(path string, h contracts.HandlerFunc) contracts.Route {
	r.router.Put(path, r.wrap(h))
	return r
}
func (r *route) Delete(path string, h contracts.HandlerFunc) contracts.Route {
	r.router.Delete(path, r.wrap(h))
	return r
}
func (r *route) Patch(path string, h contracts.HandlerFunc) contracts.Route {
	r.router.Patch(path, r.wrap(h))
	return r
}
func (r *route) Head(path string, h contracts.HandlerFunc) contracts.Route {
	r.router.Head(path, r.wrap(h))
	return r
}
func (r *route) Options(path string, h contracts.HandlerFunc) contracts.Route {
	r.router.Options(path, r.wrap(h))
	return r
}

func (r *route) Group(prefix string) contracts.Route {
	return &route{app: r.app, cfg: r.cfg, router: r.router.Group(prefix), validator: r.validator}
}

func (r *route) Use(middleware ...contracts.HandlerFunc) contracts.Route {
	for _, m := range middleware {
		r.router.Use(r.wrap(m))
	}
	return r
}

// wrap 将 contracts.HandlerFunc 转为 Fiber handler —— 唯一知道 Fiber 的地方。
func (r *route) wrap(h contracts.HandlerFunc) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return h(newFiberContext(c, r.validator))
	}
}

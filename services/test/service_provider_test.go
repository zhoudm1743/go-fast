package test

import (
	"testing"

	_ "github.com/zhoudm1743/go-fast/config"

	appcontracts "github.com/zhoudm1743/go-fast/services/contracts"
	"github.com/zhoudm1743/go-fast-framework/cache"
	"github.com/zhoudm1743/go-fast-framework/config"
	"github.com/zhoudm1743/go-fast-framework/facades"
	"github.com/zhoudm1743/go-fast-framework/foundation"
	"github.com/zhoudm1743/go-fast-framework/log"
)

func TestServiceProvider_RegisterAndBoot(t *testing.T) {
	app := foundation.NewApplication(".")
	app.SetProviders([]foundation.ServiceProvider{
		&config.ServiceProvider{},
		&log.ServiceProvider{},
		&cache.ServiceProvider{},
		&ServiceProvider{},
	})
	app.Boot()
	facades.SetApp(app)

	svc, err := app.Make("test")
	if err != nil {
		t.Fatalf("make test service: %v", err)
	}

	testSvc, ok := svc.(appcontracts.Test)
	if !ok {
		t.Fatalf("expected contracts.Test, got %T", svc)
	}

	if got := testSvc.Greet("GoFast"); got != "Hello, GoFast!" {
		t.Fatalf("unexpected greet: %q", got)
	}

	status := testSvc.Status()
	if status["service"] != "test" {
		t.Fatalf("unexpected status: %v", status)
	}
}

func TestNewTestService(t *testing.T) {
	cfg, err := config.NewConfig("/nonexistent/config.yaml")
	if err != nil {
		t.Fatalf("new config: %v", err)
	}
	cfg.Set("test.greeting_prefix", "Hi")

	svc, err := NewTestService(cfg, nil, nil)
	if err != nil {
		t.Fatalf("new test service: %v", err)
	}
	if got := svc.Greet(""); got != "Hi, World!" {
		t.Fatalf("unexpected greet: %q", got)
	}
}

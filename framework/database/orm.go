package database

import (
	"fmt"

	"github.com/zhoudm1743/go-fast/framework/contracts"
	gormdriver "github.com/zhoudm1743/go-fast/framework/database/drivers/gormdriver"
	"gorm.io/gorm"
)

// orm 包装 GormDriver，实现 contracts.Orm（向后兼容）。
// Deprecated: 请使用 contracts.DB 接口。
type orm struct {
	drv *gormdriver.GormDriver
}

// NewOrm 根据旧版扁平配置创建 ORM 实例。
// Deprecated: 请使用 NewDBManager 创建数据库管理器。
func NewOrm(cfg contracts.Config, log contracts.Log) (contracts.Orm, error) {
	cc := legacyConfigToConnectionConfig(cfg)
	cc.ApplyDefaults()
	drv, err := gormdriver.NewGormDriver(cc, log)
	if err != nil {
		return nil, fmt.Errorf("[GoFast] 创建 ORM 实例失败: %w", err)
	}
	return &orm{drv: drv}, nil
}

func (o *orm) DB() *gorm.DB {
	return o.drv.RawDB()
}

func (o *orm) Ping() error {
	return o.drv.Ping()
}

func (o *orm) Close() error {
	return o.drv.Close()
}

func (o *orm) AutoMigrate(models ...any) error {
	return o.drv.AutoMigrate(models...)
}

// legacyConfigToConnectionConfig 将旧版扁平配置转换为 ConnectionConfig。
// 保留旧版行为：log_level 为空时，根据 server.mode 决定默认日志级别。
func legacyConfigToConnectionConfig(cfg contracts.Config) contracts.ConnectionConfig {
	logLevel := cfg.GetString("database.log_level")
	if logLevel == "" {
		if cfg.GetString("server.mode") == "release" {
			logLevel = "warn"
		} else {
			logLevel = "info"
		}
	}

	return contracts.ConnectionConfig{
		Driver:          "gormdriver",
		Engine:          cfg.GetString("database.driver", "sqlite"),
		Host:            cfg.GetString("database.host", "localhost"),
		Port:            cfg.GetInt("database.port", 0),
		Username:        cfg.GetString("database.username"),
		Password:        cfg.GetString("database.password"),
		Database:        cfg.GetString("database.database"),
		Charset:         cfg.GetString("database.charset"),
		Loc:             cfg.GetString("database.loc"),
		SSLMode:         cfg.GetString("database.ssl_mode"),
		MaxIdleConns:    cfg.GetInt("database.max_idle_conns", 0),
		MaxOpenConns:    cfg.GetInt("database.max_open_conns", 0),
		ConnMaxLifetime: cfg.GetInt("database.conn_max_lifetime", 0),
		ConnMaxIdleTime: cfg.GetInt("database.conn_max_idle_time", 0),
		LogLevel:        logLevel,
		SlowThreshold:   cfg.GetInt("database.slow_threshold", 0),
	}
}

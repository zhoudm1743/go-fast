package gormdriver

import (
	"context"
	"fmt"
	"time"

	"github.com/zhoudm1743/go-fast/framework/contracts"
	"github.com/zhoudm1743/go-fast/framework/database"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// GormDriver 实现 contracts.Driver
type GormDriver struct {
	db *gorm.DB
}

var _ contracts.Driver = (*GormDriver)(nil)

// NewGormDriver 根据连接配置创建 GORM 驱动实例。
func NewGormDriver(cfg database.ConnectionConfig, log contracts.Log) (*GormDriver, error) {
	cfg.ApplyDefaults()

	dsn := cfg.BuildDSN()
	if dsn == "" {
		return nil, fmt.Errorf("[GoFast] gorm driver: unsupported engine %q", cfg.Engine)
	}

	gormCfg := buildGormConfig(cfg, log)

	var db *gorm.DB
	var err error

	switch cfg.Engine {
	case "postgres":
		db, err = gorm.Open(postgres.Open(dsn), gormCfg)
	case "sqlite", "sqlite3":
		db, err = gorm.Open(sqlite.Open(dsn), gormCfg)
	case "mysql":
		db, err = gorm.Open(mysql.Open(dsn), gormCfg)
	case "mssql":
		db, err = gorm.Open(sqlserver.Open(dsn), gormCfg)
	default:
		return nil, fmt.Errorf("[GoFast] gorm driver: unsupported engine %q", cfg.Engine)
	}

	if err != nil {
		return nil, fmt.Errorf("[GoFast] gorm driver: connection failed: %w", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("[GoFast] gorm driver: get sql.DB failed: %w", err)
	}
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Minute)
	sqlDB.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTime) * time.Minute)

	// 表前缀
	if cfg.TablePrefix != "" {
		db = db.Session(&gorm.Session{})
		// GORM 的 NamingStrategy 在 Config 中设置
	}

	return &GormDriver{db: db}, nil
}

func (d *GormDriver) Query(ctx ...context.Context) contracts.Query {
	db := d.db.Session(&gorm.Session{NewDB: true})
	if len(ctx) > 0 {
		db = db.WithContext(ctx[0])
	}
	return &GormQuery{db: db}
}

func (d *GormDriver) DriverName() string { return "gorm" }

func (d *GormDriver) Ping() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return sqlDB.PingContext(ctx)
}

func (d *GormDriver) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (d *GormDriver) AutoMigrate(models ...any) error {
	return d.db.AutoMigrate(models...)
}

// RawDB 逃生口：允许高级用户直接获取 *gorm.DB（不推荐常规使用）。
func (d *GormDriver) RawDB() *gorm.DB { return d.db }

// ── 内部：构建 GORM 配置 ────────────────────────────────────────────

type logWriter struct {
	log contracts.Log
}

func (w *logWriter) Printf(format string, args ...interface{}) {
	w.log.Infof(format, args...)
}

func buildGormConfig(cfg database.ConnectionConfig, log contracts.Log) *gorm.Config {
	var gormLogLevel gormLogger.LogLevel
	switch cfg.LogLevel {
	case "error":
		gormLogLevel = gormLogger.Error
	case "warn":
		gormLogLevel = gormLogger.Warn
	case "info":
		gormLogLevel = gormLogger.Info
	case "silent":
		gormLogLevel = gormLogger.Silent
	default:
		gormLogLevel = gormLogger.Info
	}

	customLogger := gormLogger.New(
		&logWriter{log: log},
		gormLogger.Config{
			SlowThreshold:             time.Duration(cfg.SlowThreshold) * time.Millisecond,
			LogLevel:                  gormLogLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	return &gorm.Config{
		Logger: customLogger,
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
		DisableForeignKeyConstraintWhenMigrating: true,
		PrepareStmt:                              true,
	}
}

package gormdriver

import (
	"context"
	"fmt"
	"time"

	"github.com/zhoudm1743/go-fast/framework/contracts"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	gormSchema "gorm.io/gorm/schema"
)

// GormDriver 实现 contracts.Driver
type GormDriver struct {
	db *gorm.DB
}

var _ contracts.Driver = (*GormDriver)(nil)

// NewGormDriver 根据连接配置创建 GORM 驱动实例。
func NewGormDriver(cfg contracts.ConnectionConfig, log contracts.Log) (*GormDriver, error) {
	cfg.ApplyDefaults()

	dsn := cfg.BuildDSN()
	if dsn == "" {
		return nil, fmt.Errorf("[GoFast] gormdriver driver: unsupported engine %q", cfg.Engine)
	}

	gormCfg := buildGormConfig(cfg, log)

	// 当配置了 Schema（主要用于 PostgreSQL）时，设置 GORM NamingStrategy，
	// 使 AutoMigrate 和所有 GORM 生成的 SQL 都自动携带 "schema." 前缀。
	// search_path 已在 DSN 中设置，两者并行，互不冲突：
	//   - search_path：保证原始 SQL（Exec/Raw）及 PostgreSQL 内部（外键等）正确路由
	//   - NamingStrategy.TablePrefix：保证 GORM 结构化查询和 AutoMigrate 在正确 schema 建表
	if cfg.Schema != "" {
		tablePrefix := cfg.Schema + "."
		if cfg.TablePrefix != "" {
			tablePrefix = cfg.Schema + "." + cfg.TablePrefix
		}
		gormCfg.NamingStrategy = gormSchema.NamingStrategy{
			TablePrefix: tablePrefix,
		}
	} else if cfg.TablePrefix != "" {
		gormCfg.NamingStrategy = gormSchema.NamingStrategy{
			TablePrefix: cfg.TablePrefix,
		}
}

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
		return nil, fmt.Errorf("[GoFast] gormdriver driver: unsupported engine %q", cfg.Engine)
	}

	if err != nil {
		return nil, fmt.Errorf("[GoFast] gormdriver driver: connection failed: %w", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("[GoFast] gormdriver driver: get sql.DB failed: %w", err)
	}
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Minute)
	sqlDB.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTime) * time.Minute)

	return &GormDriver{db: db}, nil
}

func (d *GormDriver) Query(ctx ...context.Context) contracts.Query {
	db := d.db.Session(&gorm.Session{NewDB: true})
	if len(ctx) > 0 {
		db = db.WithContext(ctx[0])
	}
	return &GormQuery{db: db}
}

func (d *GormDriver) DriverName() string { return "gormdriver" }

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

func buildGormConfig(cfg contracts.ConnectionConfig, log contracts.Log) *gorm.Config {
	var level gormLogger.LogLevel
	switch cfg.LogLevel {
	case "error":
		level = gormLogger.Error
	case "warn":
		level = gormLogger.Warn
	case "info":
		level = gormLogger.Info
	case "silent":
		level = gormLogger.Silent
	default:
		level = gormLogger.Info
	}

	customLogger := gormLogger.New(
		&logWriter{log: log},
		gormLogger.Config{
			SlowThreshold:             time.Duration(cfg.SlowThreshold) * time.Millisecond,
			LogLevel:                  level,
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

package database

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"go-fast/framework/contracts"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

type orm struct {
	db *gorm.DB
}

// NewOrm 根据配置创建 ORM 实例。
func NewOrm(cfg contracts.Config, log contracts.Log) (contracts.Orm, error) {
	driver := cfg.GetString("database.driver", "sqlite")
	dsn := buildDSN(cfg, driver)
	if dsn == "" {
		return nil, fmt.Errorf("[GoFast] unsupported database driver: %s", driver)
	}

	gormCfg := buildGormConfig(cfg, log)

	var db *gorm.DB
	var err error

	switch driver {
	case "postgres":
		db, err = gorm.Open(postgres.Open(dsn), gormCfg)
	case "sqlite", "sqlite3":
		db, err = gorm.Open(sqlite.Open(dsn), gormCfg)
	case "mysql":
		db, err = gorm.Open(mysql.Open(dsn), gormCfg)
	case "mssql":
		db, err = gorm.Open(sqlserver.Open(dsn), gormCfg)
	default:
		return nil, fmt.Errorf("[GoFast] unsupported database driver: %s", driver)
	}

	if err != nil {
		return nil, fmt.Errorf("[GoFast] database connection failed: %w", err)
	}

	registerUUIDPrimaryKey(db)

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("[GoFast] get sql.DB failed: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.GetInt("database.max_idle_conns", 10))
	sqlDB.SetMaxOpenConns(cfg.GetInt("database.max_open_conns", 100))
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.GetInt("database.conn_max_lifetime", 60)) * time.Minute)
	sqlDB.SetConnMaxIdleTime(time.Duration(cfg.GetInt("database.conn_max_idle_time", 30)) * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("[GoFast] database ping failed: %w", err)
	}

	return &orm{db: db}, nil
}

func (o *orm) DB() *gorm.DB { return o.db }

func (o *orm) AutoMigrate(models ...any) error {
	return o.db.AutoMigrate(models...)
}

func (o *orm) Ping() error {
	sqlDB, err := o.db.DB()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return sqlDB.PingContext(ctx)
}

func (o *orm) Close() error {
	sqlDB, err := o.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func buildDSN(cfg contracts.Config, driver string) string {
	switch driver {
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=%s",
			cfg.GetString("database.host", "localhost"),
			cfg.GetInt("database.port", 5432),
			cfg.GetString("database.username"),
			cfg.GetString("database.password"),
			cfg.GetString("database.database"),
			cfg.GetString("database.loc", "Local"))
	case "sqlite", "sqlite3":
		return fmt.Sprintf("file:%s?cache=shared&_journal_mode=WAL&_foreign_keys=1&_busy_timeout=5000",
			cfg.GetString("database.database", "database.db"))
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=%s",
			cfg.GetString("database.username"),
			cfg.GetString("database.password"),
			cfg.GetString("database.host", "localhost"),
			cfg.GetInt("database.port", 3306),
			cfg.GetString("database.database"),
			cfg.GetString("database.loc", "Local"))
	case "mssql":
		return fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
			cfg.GetString("database.username"),
			cfg.GetString("database.password"),
			cfg.GetString("database.host", "localhost"),
			cfg.GetInt("database.port", 1433),
			cfg.GetString("database.database"))
	default:
		return ""
	}
}

type logWriter struct {
	log contracts.Log
}

func (w *logWriter) Printf(format string, args ...interface{}) {
	w.log.Infof(format, args...)
}

func buildGormConfig(cfg contracts.Config, log contracts.Log) *gorm.Config {
	var gormLogLevel gormLogger.LogLevel
	switch cfg.GetString("database.log_level") {
	case "error":
		gormLogLevel = gormLogger.Error
	case "warn":
		gormLogLevel = gormLogger.Warn
	case "info":
		gormLogLevel = gormLogger.Info
	case "silent":
		gormLogLevel = gormLogger.Silent
	default:
		if cfg.GetString("server.mode") == "release" {
			gormLogLevel = gormLogger.Warn
		} else {
			gormLogLevel = gormLogger.Info
		}
	}

	slowThreshold := cfg.GetInt("database.slow_threshold", 200)
	customLogger := gormLogger.New(
		&logWriter{log: log},
		gormLogger.Config{
			SlowThreshold:             time.Duration(slowThreshold) * time.Millisecond,
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

func registerUUIDPrimaryKey(db *gorm.DB) {
	_ = db.Callback().Create().Before("gorm:before_create").Register("gofast:uuid_primary_key", func(db *gorm.DB) {
		if db.Statement.Dest == nil {
			return
		}
		val := reflect.ValueOf(db.Statement.Dest)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		switch val.Kind() {
		case reflect.Slice:
			for i := 0; i < val.Len(); i++ {
				setUUIDIfEmpty(val.Index(i))
			}
		case reflect.Struct:
			setUUIDIfEmpty(val)
		}
	})
}

func setUUIDIfEmpty(v reflect.Value) {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return
	}
	f := v.FieldByName("ID")
	if !f.IsValid() || f.Kind() != reflect.String || f.String() != "" || !f.CanSet() {
		return
	}
	f.SetString(uuid.Must(uuid.NewV7()).String())
}

package database

import "fmt"

// ConnectionConfig 单个数据库连接的完整配置。
type ConnectionConfig struct {
	// ── 驱动与引擎 ──────────────────────────────────────────
	Driver string `mapstructure:"driver"` // ORM 驱动：gorm | xorm | torm
	Engine string `mapstructure:"engine"` // 数据库引擎：mysql | postgres | sqlite | mssql

	// ── 直连 DSN（优先级高于以下分项配置）──────────────────
	DSN string `mapstructure:"dsn"`

	// ── 分项连接配置 ────────────────────────────────────────
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	Charset  string `mapstructure:"charset"`  // 默认 utf8mb4（MySQL）
	Loc      string `mapstructure:"loc"`      // 时区，默认 Local
	SSLMode  string `mapstructure:"ssl_mode"` // postgres: disable|require|verify-full

	// ── 表前缀 ──────────────────────────────────────────────
	TablePrefix string `mapstructure:"table_prefix"`

	// ── 连接池 ──────────────────────────────────────────────
	MaxIdleConns    int `mapstructure:"max_idle_conns"`     // 默认 10
	MaxOpenConns    int `mapstructure:"max_open_conns"`     // 默认 100
	ConnMaxLifetime int `mapstructure:"conn_max_lifetime"`  // 分钟，默认 60
	ConnMaxIdleTime int `mapstructure:"conn_max_idle_time"` // 分钟，默认 30

	// ── 日志与监控 ──────────────────────────────────────────
	LogLevel      string `mapstructure:"log_level"`      // error|warn|info|silent
	SlowThreshold int    `mapstructure:"slow_threshold"` // ms，默认 200
}

// BuildDSN 根据分项配置生成 DSN 字符串（DSN 字段非空时直接返回）。
func (c *ConnectionConfig) BuildDSN() string {
	if c.DSN != "" {
		return c.DSN
	}

	charset := c.Charset
	if charset == "" {
		charset = "utf8mb4"
	}
	loc := c.Loc
	if loc == "" {
		loc = "Local"
	}

	switch c.Engine {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=%s",
			c.Username, c.Password, c.Host, c.Port, c.Database, charset, loc)
	case "postgres":
		sslMode := c.SSLMode
		if sslMode == "" {
			sslMode = "disable"
		}
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
			c.Host, c.Port, c.Username, c.Password, c.Database, sslMode, loc)
	case "sqlite", "sqlite3":
		return fmt.Sprintf("file:%s?cache=shared&_journal_mode=WAL&_foreign_keys=1&_busy_timeout=5000",
			c.Database)
	case "mssql":
		return fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
			c.Username, c.Password, c.Host, c.Port, c.Database)
	default:
		return ""
	}
}

// ApplyDefaults 为未设置的配置项填充默认值。
func (c *ConnectionConfig) ApplyDefaults() {
	if c.Driver == "" {
		c.Driver = "gorm"
	}
	if c.MaxIdleConns == 0 {
		c.MaxIdleConns = 10
	}
	if c.MaxOpenConns == 0 {
		c.MaxOpenConns = 100
	}
	if c.ConnMaxLifetime == 0 {
		c.ConnMaxLifetime = 60
	}
	if c.ConnMaxIdleTime == 0 {
		c.ConnMaxIdleTime = 30
	}
	if c.SlowThreshold == 0 {
		c.SlowThreshold = 200
	}
}

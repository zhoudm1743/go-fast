package database

import (
	"context"
	"fmt"
	"sync"

	"github.com/zhoudm1743/go-fast/framework/contracts"
)

// ── 驱动工厂注册表（全局�?──────────────────────────────────────────

// DriverFactory 根据连接配置创建 Driver�?type DriverFactory func(cfg ConnectionConfig, log contracts.Log) (contracts.Driver, error)

var (
	driverFactoriesMu sync.RWMutex
	driverFactories   = map[string]DriverFactory{}
)

// RegisterDriver 由插件的 ServiceProvider 在启动时调用，注�?ORM 驱动工厂�?func RegisterDriver(name string, f DriverFactory) {
	driverFactoriesMu.Lock()
	defer driverFactoriesMu.Unlock()
	driverFactories[name] = f
}

func getDriverFactory(name string) (DriverFactory, bool) {
	driverFactoriesMu.RLock()
	defer driverFactoriesMu.RUnlock()
	f, ok := driverFactories[name]
	return f, ok
}

// ── dbManager —�?contracts.DB 实现 ──────────────────────────────────

type dbManager struct {
	cfg         contracts.Config
	log         contracts.Log
	defaultConn string
	connConfigs map[string]ConnectionConfig // 配置缓存
	connections map[string]contracts.Driver // 懒加载驱动缓�?	mu          sync.RWMutex
}

var _ contracts.DB = (*dbManager)(nil)

// NewDBManager 创建数据库管理器实例�?// 自动检测新/旧配置格式，向后兼容旧的扁平化配置�?func NewDBManager(cfg contracts.Config, log contracts.Log) (contracts.DB, error) {
	m := &dbManager{
		cfg:         cfg,
		log:         log,
		connConfigs: make(map[string]ConnectionConfig),
		connections: make(map[string]contracts.Driver),
	}

	// 解析配置
	if err := m.parseConfig(); err != nil {
		return nil, err
	}

	return m, nil
}

// parseConfig 解析数据库配置。支持新格式（connections 节点）和旧格式（扁平化）�?func (m *dbManager) parseConfig() error {
	connMap := m.cfg.GetStringMap("database.connections")

	if len(connMap) > 0 {
		// ── 新配置格�?──────────────────────────────────────
		m.defaultConn = m.cfg.GetString("database.default", "main")
		for name := range connMap {
			cc := m.readConnectionConfig("database.connections." + name)
			cc.ApplyDefaults()
			m.connConfigs[name] = cc
		}
	} else {
		// ── 旧配置格式（向后兼容）────────────────────────────
		// 将扁平化 database.* 适配�?connections.main
		m.defaultConn = "main"
		cc := m.readLegacyConfig()
		cc.ApplyDefaults()
		m.connConfigs["main"] = cc
	}

	return nil
}

// readConnectionConfig 从配置中读取指定前缀的连接配置�?func (m *dbManager) readConnectionConfig(prefix string) ConnectionConfig {
	return ConnectionConfig{
		Driver:          m.cfg.GetString(prefix+".driver", "gorm"),
		Engine:          m.cfg.GetString(prefix+".engine", "sqlite"),
		DSN:             m.cfg.GetString(prefix + ".dsn"),
		Host:            m.cfg.GetString(prefix+".host", "localhost"),
		Port:            m.cfg.GetInt(prefix+".port", 0),
		Username:        m.cfg.GetString(prefix + ".username"),
		Password:        m.cfg.GetString(prefix + ".password"),
		Database:        m.cfg.GetString(prefix + ".database"),
		Charset:         m.cfg.GetString(prefix + ".charset"),
		Loc:             m.cfg.GetString(prefix + ".loc"),
		SSLMode:         m.cfg.GetString(prefix + ".ssl_mode"),
		TablePrefix:     m.cfg.GetString(prefix + ".table_prefix"),
		MaxIdleConns:    m.cfg.GetInt(prefix+".max_idle_conns", 0),
		MaxOpenConns:    m.cfg.GetInt(prefix+".max_open_conns", 0),
		ConnMaxLifetime: m.cfg.GetInt(prefix+".conn_max_lifetime", 0),
		ConnMaxIdleTime: m.cfg.GetInt(prefix+".conn_max_idle_time", 0),
		LogLevel:        m.cfg.GetString(prefix + ".log_level"),
		SlowThreshold:   m.cfg.GetInt(prefix+".slow_threshold", 0),
	}
}

// readLegacyConfig 读取旧的扁平�?database.* 配置�?func (m *dbManager) readLegacyConfig() ConnectionConfig {
	// 旧配置的 driver 同时代表 engine（sqlite/mysql/postgres/mssql�?	engine := m.cfg.GetString("database.driver", "sqlite")
	return ConnectionConfig{
		Driver:          "gorm", // 旧配置默认使�?GORM
		Engine:          engine,
		Host:            m.cfg.GetString("database.host", "localhost"),
		Port:            m.cfg.GetInt("database.port", 0),
		Username:        m.cfg.GetString("database.username"),
		Password:        m.cfg.GetString("database.password"),
		Database:        m.cfg.GetString("database.database"),
		Loc:             m.cfg.GetString("database.loc"),
		MaxIdleConns:    m.cfg.GetInt("database.max_idle_conns", 0),
		MaxOpenConns:    m.cfg.GetInt("database.max_open_conns", 0),
		ConnMaxLifetime: m.cfg.GetInt("database.conn_max_lifetime", 0),
		ConnMaxIdleTime: m.cfg.GetInt("database.conn_max_idle_time", 0),
		LogLevel:        m.cfg.GetString("database.log_level"),
		SlowThreshold:   m.cfg.GetInt("database.slow_threshold", 0),
	}
}

// getOrCreateDriver 懒加载并缓存指定连接�?Driver�?func (m *dbManager) getOrCreateDriver(name string) contracts.Driver {
	// 快路径：读锁检查缓�?	m.mu.RLock()
	drv, ok := m.connections[name]
	m.mu.RUnlock()
	if ok {
		return drv
	}

	// 慢路径：写锁创建
	m.mu.Lock()
	defer m.mu.Unlock()

	// 二次检�?	if drv, ok = m.connections[name]; ok {
		return drv
	}

	cc, exists := m.connConfigs[name]
	if !exists {
		panic(fmt.Sprintf("[GoFast] database connection %q not configured", name))
	}

	factory, ok := getDriverFactory(cc.Driver)
	if !ok {
		panic(fmt.Sprintf("[GoFast] database driver %q not registered (connection %q). Did you register the driver plugin?", cc.Driver, name))
	}

	drv, err := factory(cc, m.log)
	if err != nil {
		panic(fmt.Sprintf("[GoFast] database connection %q init failed: %v", name, err))
	}

	m.connections[name] = drv
	return drv
}

// ── contracts.DB 接口实现 ────────────────────────────────────────────

func (m *dbManager) Query(ctx ...context.Context) contracts.Query {
	drv := m.getOrCreateDriver(m.defaultConn)
	return drv.Query(ctx...)
}

func (m *dbManager) Connection(name string) contracts.Query {
	drv := m.getOrCreateDriver(name)
	return drv.Query()
}

func (m *dbManager) Driver(name ...string) contracts.Driver {
	connName := m.defaultConn
	if len(name) > 0 {
		connName = name[0]
	}
	return m.getOrCreateDriver(connName)
}

func (m *dbManager) Transaction(fc func(tx contracts.Query) error, opts ...contracts.TxOption) error {
	return m.Query().Transaction(fc, opts...)
}

func (m *dbManager) AutoMigrate(models ...any) error {
	return m.getOrCreateDriver(m.defaultConn).AutoMigrate(models...)
}

func (m *dbManager) Ping() error {
	return m.getOrCreateDriver(m.defaultConn).Ping()
}

func (m *dbManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var lastErr error
	for name, drv := range m.connections {
		if err := drv.Close(); err != nil {
			m.log.Errorf("[GoFast] close database connection %q: %v", name, err)
			lastErr = err
		}
	}
	m.connections = make(map[string]contracts.Driver)
	return lastErr
}


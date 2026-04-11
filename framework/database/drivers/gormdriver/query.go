package gormdriver

import (
	"context"
	"database/sql"
	"reflect"
	"strings"

	"github.com/zhoudm1743/go-fast/framework/contracts"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GormQuery 将 contracts.Query 的每个方法代理到 *gorm.DB。
// 所有链式方法返回新的 GormQuery 实例，保持不可变。
type GormQuery struct {
	db     *gorm.DB
	schema string // 动态 schema 前缀，主要用于 PostgreSQL 多 schema 场景
}

var _ contracts.Query = (*GormQuery)(nil)

// wrap 创建新的 GormQuery，传入新的 *gorm.DB，并保留当前 schema。
func (q *GormQuery) wrap(db *gorm.DB) *GormQuery {
	return &GormQuery{db: db, schema: q.schema}
}

// schemaTable 在 schema 非空且 name 中不含 "." 时自动加上 "schema." 前缀。
func (q *GormQuery) schemaTable(name string) string {
	if q.schema != "" && !strings.Contains(name, ".") {
		return q.schema + "." + name
	}
	return name
}

// ── 调试／Schema ─────────────────────────────────────────────────────

// Schema 在当前查询链上设置动态 schema（主要用于 PostgreSQL）。
// 后续的 Model()/Table() 调用将自动在表名前加上 "schema." 前缀。
// 示例：facades.DB().Connection("pg").Schema("analytics").Model(&Event{}).Find(&events)
func (q *GormQuery) Schema(name string) contracts.Query {
	return &GormQuery{db: q.db, schema: name}
}

// applySchema 在终结方法（First/Find/Create 等）执行前，
// 如果设置了 schema 且 dest 是可解析的 GORM model，
// 则自动在 db 上调用 Table(schema.tableName)。
func (q *GormQuery) applySchema(dest any) *gorm.DB {
	if q.schema != "" && dest != nil {
		stmt := &gorm.Statement{DB: q.db}
		if err := stmt.Parse(dest); err == nil && stmt.Table != "" && !strings.Contains(stmt.Table, ".") {
			return q.db.Table(q.schema + "." + stmt.Table)
		}
	}
	return q.db
}

// ── 构建条件 ─────────────────────────────────────────────────────────

func (q *GormQuery) Table(name string) contracts.Query {
	return q.wrap(q.db.Table(q.schemaTable(name)))
}

func (q *GormQuery) Model(value any) contracts.Query {
	if q.schema != "" {
		// 通过 gorm.Statement.Parse 解析 model 对应的裸表名，再加上 schema 前缀。
		stmt := &gorm.Statement{DB: q.db}
		if err := stmt.Parse(value); err == nil && stmt.Table != "" && !strings.Contains(stmt.Table, ".") {
			return q.wrap(q.db.Table(q.schema + "." + stmt.Table).Model(value))
		}
	}
	return q.wrap(q.db.Model(value))
}

func (q *GormQuery) Select(query any, args ...any) contracts.Query {
	return q.wrap(q.db.Select(query, args...))
}

func (q *GormQuery) Omit(columns ...string) contracts.Query {
	return q.wrap(q.db.Omit(columns...))
}

func (q *GormQuery) Where(query any, args ...any) contracts.Query {
	return q.wrap(q.db.Where(query, args...))
}

func (q *GormQuery) OrWhere(query any, args ...any) contracts.Query {
	return q.wrap(q.db.Or(query, args...))
}

func (q *GormQuery) Not(query any, args ...any) contracts.Query {
	return q.wrap(q.db.Not(query, args...))
}

func (q *GormQuery) Order(value any) contracts.Query {
	return q.wrap(q.db.Order(value))
}

func (q *GormQuery) Limit(limit int) contracts.Query {
	return q.wrap(q.db.Limit(limit))
}

func (q *GormQuery) Offset(offset int) contracts.Query {
	return q.wrap(q.db.Offset(offset))
}

func (q *GormQuery) Group(name string) contracts.Query {
	return q.wrap(q.db.Group(name))
}

func (q *GormQuery) Having(query any, args ...any) contracts.Query {
	return q.wrap(q.db.Having(query, args...))
}

func (q *GormQuery) Distinct(args ...any) contracts.Query {
	return q.wrap(q.db.Distinct(args...))
}

// ── 关联 ─────────────────────────────────────────────────────────────

func (q *GormQuery) Joins(query string, args ...any) contracts.Query {
	return q.wrap(q.db.Joins(query, args...))
}

func (q *GormQuery) Preload(query string, args ...any) contracts.Query {
	return q.wrap(q.db.Preload(query, args...))
}

// ── 终结方法 ─────────────────────────────────────────────────────────

func (q *GormQuery) Find(dest any, conds ...any) error {
	return wrapError(q.applySchema(dest).Find(dest, conds...).Error)
}

func (q *GormQuery) First(dest any, conds ...any) error {
	return wrapError(q.applySchema(dest).First(dest, conds...).Error)
}

func (q *GormQuery) Last(dest any, conds ...any) error {
	return wrapError(q.applySchema(dest).Last(dest, conds...).Error)
}

func (q *GormQuery) Take(dest any, conds ...any) error {
	return wrapError(q.applySchema(dest).Take(dest, conds...).Error)
}

func (q *GormQuery) Create(value any) error {
	if err := invokeBeforeCreate(q, value); err != nil {
		return err
	}
	return wrapError(q.applySchema(value).Create(value).Error)
}

func (q *GormQuery) CreateInBatches(value any, batchSize int) error {
	return wrapError(q.applySchema(value).CreateInBatches(value, batchSize).Error)
}

func (q *GormQuery) Save(value any) error {
	return wrapError(q.applySchema(value).Save(value).Error)
}

func (q *GormQuery) Update(column string, value any) error {
	return wrapError(q.db.Update(column, value).Error)
}

func (q *GormQuery) Updates(values any) error {
	return wrapError(q.db.Updates(values).Error)
}

func (q *GormQuery) Delete(value any, conds ...any) error {
	return wrapError(q.applySchema(value).Delete(value, conds...).Error)
}

func (q *GormQuery) Count(count *int64) error {
	return wrapError(q.db.Count(count).Error)
}

func (q *GormQuery) Scan(dest any) error {
	return wrapError(q.db.Scan(dest).Error)
}

func (q *GormQuery) Pluck(column string, dest any) error {
	return wrapError(q.db.Pluck(column, dest).Error)
}

func (q *GormQuery) Row() contracts.Row {
	return q.db.Row()
}

func (q *GormQuery) Rows() (contracts.Rows, error) {
	rows, err := q.db.Rows()
	if err != nil {
		return nil, wrapError(err)
	}
	return rows, nil
}

// ── 写操作 Result 变体 ──────────────────────────────────────────────

func (q *GormQuery) CreateResult(value any) contracts.Result {
	if err := invokeBeforeCreate(q, value); err != nil {
		return contracts.Result{Error: err}
	}
	tx := q.applySchema(value).Create(value)
	return contracts.Result{RowsAffected: tx.RowsAffected, Error: wrapError(tx.Error)}
}

func (q *GormQuery) UpdateResult(column string, value any) contracts.Result {
	tx := q.db.Update(column, value)
	return contracts.Result{RowsAffected: tx.RowsAffected, Error: wrapError(tx.Error)}
}

func (q *GormQuery) UpdatesResult(values any) contracts.Result {
	tx := q.db.Updates(values)
	return contracts.Result{RowsAffected: tx.RowsAffected, Error: wrapError(tx.Error)}
}

func (q *GormQuery) DeleteResult(value any, conds ...any) contracts.Result {
	tx := q.applySchema(value).Delete(value, conds...)
	return contracts.Result{RowsAffected: tx.RowsAffected, Error: wrapError(tx.Error)}
}

func (q *GormQuery) SaveResult(value any) contracts.Result {
	tx := q.applySchema(value).Save(value)
	return contracts.Result{RowsAffected: tx.RowsAffected, Error: wrapError(tx.Error)}
}

// ── 原生 SQL ─────────────────────────────────────────────────────────

func (q *GormQuery) Raw(sql string, values ...any) contracts.Query {
	return q.wrap(q.db.Raw(sql, values...))
}

func (q *GormQuery) Exec(sql string, values ...any) error {
	return wrapError(q.db.Exec(sql, values...).Error)
}

// ── 事务 ─────────────────────────────────────────────────────────────

func (q *GormQuery) Transaction(fc func(tx contracts.Query) error, opts ...contracts.TxOption) error {
	txOpts := parseTxOptions(opts...)
	return wrapError(q.db.Transaction(func(tx *gorm.DB) error {
		return fc(q.wrap(tx))
	}, txOpts))
}

func (q *GormQuery) Begin(opts ...contracts.TxOption) contracts.Query {
	txOpts := parseTxOptions(opts...)
	var tx *gorm.DB
	if txOpts != nil {
		tx = q.db.Begin(txOpts)
	} else {
		tx = q.db.Begin()
	}
	return q.wrap(tx)
}

func (q *GormQuery) Commit() error {
	return wrapError(q.db.Commit().Error)
}

func (q *GormQuery) Rollback() error {
	return wrapError(q.db.Rollback().Error)
}

func (q *GormQuery) SavePoint(name string) error {
	return wrapError(q.db.SavePoint(name).Error)
}

func (q *GormQuery) RollbackTo(name string) error {
	return wrapError(q.db.RollbackTo(name).Error)
}

// ── 分页 ─────────────────────────────────────────────────────────────

func (q *GormQuery) Paginate(page, size int) contracts.Query {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	return q.wrap(q.db.Offset((page - 1) * size).Limit(size))
}

// ── 作用域 ───────────────────────────────────────────────────────────

func (q *GormQuery) Scopes(funcs ...func(contracts.Query) contracts.Query) contracts.Query {
	gormScopes := make([]func(*gorm.DB) *gorm.DB, 0, len(funcs))
	for _, fn := range funcs {
		fn := fn
		gormScopes = append(gormScopes, func(db *gorm.DB) *gorm.DB {
			result := fn(q.wrap(db))
			if gq, ok := result.(*GormQuery); ok {
				return gq.db
			}
			return db
		})
	}
	return q.wrap(q.db.Scopes(gormScopes...))
}

// ── 上下文 ───────────────────────────────────────────────────────────

func (q *GormQuery) WithContext(ctx context.Context) contracts.Query {
	return q.wrap(q.db.WithContext(ctx))
}

// ── 调试 ─────────────────────────────────────────────────────

func (q *GormQuery) Debug() contracts.Query {
	return q.wrap(q.db.Debug())
}

// ── 悲观锁 ───────────────────────────────────────────────────────────

func (q *GormQuery) Lock(mode contracts.LockMode) contracts.Query {
	switch mode {
	case contracts.LockForUpdate:
		return q.wrap(q.db.Clauses(clause.Locking{Strength: "UPDATE"}))
	case contracts.LockShareMode:
		return q.wrap(q.db.Clauses(clause.Locking{Strength: "SHARE"}))
	default:
		return q
	}
}

// ── 软删除扩展 ───────────────────────────────────────────────────

func (q *GormQuery) Unscoped() contracts.Query {
	return q.wrap(q.db.Unscoped())
}

func (q *GormQuery) OnlyTrashed() contracts.Query {
	return q.wrap(q.db.Unscoped().Where("deleted_at != 0"))
}

func (q *GormQuery) Restore() error {
	return wrapError(q.db.Unscoped().Update("deleted_at", 0).Error)
}

func (q *GormQuery) ForceDelete(value any, conds ...any) error {
	return wrapError(q.db.Unscoped().Delete(value, conds...).Error)
}

// ── 高级查询 ─────────────────────────────────────────────────────────

func (q *GormQuery) FirstOrCreate(dest any, conds ...any) error {
	return wrapError(q.db.FirstOrCreate(dest, conds...).Error)
}

func (q *GormQuery) FirstOrInit(dest any, conds ...any) error {
	return wrapError(q.db.FirstOrInit(dest, conds...).Error)
}

func (q *GormQuery) FindInBatches(dest any, batchSize int, fc func(tx contracts.Query, batch int) error) error {
	return wrapError(q.db.FindInBatches(dest, batchSize, func(tx *gorm.DB, batch int) error {
		return fc(q.wrap(tx), batch)
	}).Error)
}

func (q *GormQuery) ScanMap(dest *[]map[string]any) error {
	rows, err := q.db.Rows()
	if err != nil {
		return wrapError(err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return wrapError(err)
	}

	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			return wrapError(err)
		}
		row := make(map[string]any, len(columns))
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		*dest = append(*dest, row)
	}
	return nil
}

func (q *GormQuery) Exists(dest any, conds ...any) (bool, error) {
	var count int64
	err := q.db.Model(dest).Where(conds[0], conds[1:]...).Limit(1).Count(&count).Error
	if err != nil {
		return false, wrapError(err)
	}
	return count > 0, nil
}

// ── 辅助 ─────────────────────────────────────────────────────────────

func parseTxOptions(opts ...contracts.TxOption) *sql.TxOptions {
	for _, opt := range opts {
		if std, ok := opt.(*contracts.StandardTxOptions); ok {
			return &sql.TxOptions{
				Isolation: std.Isolation,
				ReadOnly:  std.ReadOnly,
			}
		}
	}
	return nil
}

// invokeBeforeCreate 对 value 调用 BeforeCreate Hook。
// 支持以下传入形式：
//   - *T（单条记录指针）
//   - []T 或 *[]T（批量记录，逐条调用）
//   - []*T 或 *[]*T（批量记录指针，逐条调用）
//
// 调用顺序：
//  1. contracts.IDAutoGenerator.AutoGenerateID() —— 自动生成主键，方法名不与任何 ORM Hook 冲突
//  2. contracts.BeforeCreator.BeforeCreate(q)    —— 业务自定义创建前逻辑
func invokeBeforeCreate(q contracts.Query, value any) error {
	// 直接实现接口：最常见的 *T 路径，快速返回
	if handled := callBeforeCreateOnValue(q, value); handled != nil {
		return handled
	}

	// 用 reflect 处理切片/切片指针场景
	rv := reflect.ValueOf(value)
	// 解引用外层指针（*[]T → []T）
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Slice {
		return nil
	}
	for i := 0; i < rv.Len(); i++ {
		elem := rv.Index(i)
		// 获取可寻址的指针以便接口匹配（[]T → &T）
		var iface any
		if elem.Kind() == reflect.Ptr {
			iface = elem.Interface()
		} else if elem.CanAddr() {
			iface = elem.Addr().Interface()
		} else {
			continue
		}
		if err := callBeforeCreateOnValue(q, iface); err != nil {
			return err
		}
	}
	return nil
}

// callBeforeCreateOnValue 对单个对象调用 IDAutoGenerator 和 BeforeCreator。
// 返回 nil 表示没有错误（包含"未实现接口"的情况）。
func callBeforeCreateOnValue(q contracts.Query, iface any) error {
	if ag, ok := iface.(contracts.IDAutoGenerator); ok {
		ag.AutoGenerateID()
	}
	if bc, ok := iface.(contracts.BeforeCreator); ok {
		return bc.OnBeforeCreate(q)
	}
	return nil
}

package log

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/zhoudm1743/go-fast/framework/contracts"
	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// logger 实现 contracts.Log 接口，包装 zap。
type logger struct {
	base    *zap.Logger
	sugar   *zap.SugaredLogger
	closers []io.Closer // lumberjack 等需要 Close 的 writer
}

// logEntry 包装 zap.SugaredLogger 以实现 contracts.Log 链式调用。
type logEntry struct {
	sugar *zap.SugaredLogger
}

// callerFields 通过 runtime.Caller 获取调用者信息，返回 zap 字段。
// skip=3：0=callerFields，1=write/withCaller，2=公开方法（如 Info），3=业务代码。
func callerFields() []zapcore.Field {
	_, file, line, ok := runtime.Caller(3)
	f := shorten(file, ok)
	return []zapcore.Field{
		zap.String("caller_file", f),
		zap.Int("caller_line", line),
		zap.String("caller", fmt.Sprintf("%s:%d", f, line)),
	}
}

func shorten(file string, ok bool) string {
	if !ok {
		return "unknown"
	}
	return filepath.ToSlash(filepath.Clean(file))
}

// =============================================================================
// 级别解析
// =============================================================================

var levelMap = map[string]zapcore.Level{
	"debug": zapcore.DebugLevel,
	"info":  zapcore.InfoLevel,
	"warn":  zapcore.WarnLevel,
	"error": zapcore.ErrorLevel,
	"fatal": zapcore.FatalLevel,
	"panic": zapcore.PanicLevel,
}

func parseLevel(s string) (zapcore.Level, error) {
	if lv, ok := levelMap[strings.ToLower(s)]; ok {
		return lv, nil
	}
	return zapcore.InfoLevel, fmt.Errorf("[GoFast] 无效的日志级别 %q，合法值: debug, info, warn, error, fatal, panic", s)
}

// =============================================================================
// NewLogger
// =============================================================================

// NewLogger 根据配置创建 Logger 实例。
func NewLogger(cfg contracts.Config) (contracts.Log, error) {
	// 1. 解析级别
	level, err := parseLevel(cfg.GetString("log.level", "info"))
	if err != nil {
		return nil, err
	}

	// 2. release 模式下提升最低级别到 Info（过滤掉 Debug），保留高级别
	if cfg.GetString("server.mode", "debug") == "release" && level < zapcore.InfoLevel {
		level = zapcore.InfoLevel
	}
	atomicLevel := zap.NewAtomicLevelAt(level)

	// 3. 按 mode 构建 Core
	mode := cfg.GetString("log.mode", "hybrid")
	core, closers, err := buildCores(cfg, atomicLevel, mode)
	if err != nil {
		return nil, err
	}

	z := zap.New(core,
		zap.AddCaller(),           // 兜底 caller
		zap.AddCallerSkip(1),      // 跳过 zap 自身包装
		zap.AddStacktrace(zapcore.FatalLevel), // Fatal 及以上记录调用栈
	)

	return &logger{
		base:    z,
		sugar:   z.Sugar(),
		closers: closers,
	}, nil
}

func buildCores(cfg contracts.Config, lvl zap.AtomicLevel, mode string) (zapcore.Core, []io.Closer, error) {
	var cores []zapcore.Core
	var closers []io.Closer

	switch mode {
	case "console":
		cores = append(cores, newConsoleCore(cfg, lvl))
	case "file":
		path := cfg.GetString("log.output_path")
		if path == "" {
			return nil, nil, fmt.Errorf("[GoFast] log.mode=file 时 log.output_path 不能为空")
		}
		fileCore, fc, err := newFileCore(cfg, lvl, path)
		if err != nil {
			return nil, nil, err
		}
		cores = append(cores, fileCore)
		closers = append(closers, fc)
	case "hybrid":
		cores = append(cores, newConsoleCore(cfg, lvl))
		path := cfg.GetString("log.output_path")
		if path != "" {
			fileCore, fc, err := newFileCore(cfg, lvl, path)
			if err != nil {
				return nil, nil, err
			}
			cores = append(cores, fileCore)
			closers = append(closers, fc)
		}
	default:
		return nil, nil, fmt.Errorf("[GoFast] 无效的 log.mode %q，合法值: console, file, hybrid", mode)
	}

	return zapcore.NewTee(cores...), closers, nil
}

// =============================================================================
// Core 构造
// =============================================================================

func newConsoleCore(cfg contracts.Config, lvl zap.AtomicLevel) zapcore.Core {
	format := cfg.GetString("log.format", "color")
	tsFmt := cfg.GetString("log.timestamp_format", "2006-01-02 15:04:05")

	var enc zapcore.Encoder
	encCfg := zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     "", // caller 由 twoLineEncoder 手动处理
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.CapitalLevelEncoder,
		EncodeTime:    zapcore.TimeEncoderOfLayout(tsFmt),
		EncodeDuration: zapcore.SecondsDurationEncoder,
	}

	switch format {
	case "json":
		enc = zapcore.NewJSONEncoder(encCfg)
	case "color":
		encCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		fallthrough
	default:
		base := zapcore.NewConsoleEncoder(encCfg)
		// 仅在非 JSON 模式下使用两行格式（caller 单独一行）
		if format != "json" {
			enc = &twoLineEncoder{Encoder: base}
		} else {
			enc = base
		}
	}

	return zapcore.NewCore(enc, zapcore.Lock(os.Stdout), lvl)
}

func newFileCore(cfg contracts.Config, lvl zap.AtomicLevel, path string) (zapcore.Core, io.Closer, error) {
	logDir := filepath.Dir(path)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, nil, fmt.Errorf("[GoFast] 创建日志目录失败: %w", err)
	}

	lj := &lumberjack.Logger{
		Filename:   path,
		MaxSize:    cfg.GetInt("log.max_size", 100),
		MaxBackups: cfg.GetInt("log.max_backups", 5),
		MaxAge:     cfg.GetInt("log.max_age", 30),
		Compress:   cfg.GetBool("log.compress"),
		LocalTime:  true,
	}

	fileFormat := cfg.GetString("log.file_format", "json")
	tsFmt := cfg.GetString("log.timestamp_format", "2006-01-02 15:04:05")

	encCfg := zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     "caller",
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.LowercaseLevelEncoder,
		EncodeTime:    zapcore.TimeEncoderOfLayout(tsFmt),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var enc zapcore.Encoder
	switch fileFormat {
	case "text":
		enc = zapcore.NewConsoleEncoder(encCfg)
	default:
		enc = zapcore.NewJSONEncoder(encCfg)
	}

	return zapcore.NewCore(enc, zapcore.AddSync(lj), lvl), lj, nil
}

// =============================================================================
// twoLineEncoder：控制台两行格式（caller 单独一行）
// =============================================================================

type twoLineEncoder struct {
	zapcore.Encoder
}

func (e *twoLineEncoder) EncodeEntry(ent zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	// 提取 caller 字段，过滤掉 caller_file / caller_line
	var callerLine string
	filtered := make([]zapcore.Field, 0, len(fields))
	for _, f := range fields {
		switch f.Key {
		case "caller":
			if f.Type == zapcore.StringType {
				callerLine = f.String
			}
		case "caller_file", "caller_line":
			// 跳过，由 caller 统一展示
		default:
			filtered = append(filtered, f)
		}
	}

	buf, err := e.Encoder.EncodeEntry(ent, filtered)
	if err != nil {
		return nil, err
	}
	if callerLine == "" {
		return buf, nil
	}

	// 在首部插入 caller 行
	existing := buf.Bytes()
	buf.Reset()
	buf.AppendString(callerLine + "\n")
	buf.AppendBytes(existing)
	return buf, nil
}

func (e *twoLineEncoder) Clone() zapcore.Encoder {
	return &twoLineEncoder{Encoder: e.Encoder.Clone()}
}

// =============================================================================
// logger 写方法（核心内部方法）
// =============================================================================

func (l *logger) write(level zapcore.Level, args ...any) {
	fields := callerFields()
	l.sugar.With(fieldsToAny(fields)...).Log(level, args...)
}

func (l *logger) writef(level zapcore.Level, format string, args ...any) {
	fields := callerFields()
	l.sugar.With(fieldsToAny(fields)...).Logf(level, format, args...)
}

// fieldsToAny 将 []zapcore.Field 转为 ...any 供 Sugar 使用。
func fieldsToAny(fields []zapcore.Field) []any {
	result := make([]any, len(fields))
	for i, f := range fields {
		result[i] = f
	}
	return result
}

// =============================================================================
// logger 公开方法（实现 contracts.Log）
// =============================================================================

func (l *logger) Debug(args ...any)                 { l.write(zapcore.DebugLevel, args...) }
func (l *logger) Debugf(f string, a ...any)         { l.writef(zapcore.DebugLevel, f, a...) }
func (l *logger) Info(args ...any)                  { l.write(zapcore.InfoLevel, args...) }
func (l *logger) Infof(f string, a ...any)          { l.writef(zapcore.InfoLevel, f, a...) }
func (l *logger) Warn(args ...any)                  { l.write(zapcore.WarnLevel, args...) }
func (l *logger) Warnf(f string, a ...any)          { l.writef(zapcore.WarnLevel, f, a...) }
func (l *logger) Error(args ...any)                 { l.write(zapcore.ErrorLevel, args...) }
func (l *logger) Errorf(f string, a ...any)         { l.writef(zapcore.ErrorLevel, f, a...) }

// Fatal 写日志后调用 os.Exit(1)，兼容 logrus 语义。
func (l *logger) Fatal(args ...any) {
	l.write(zapcore.FatalLevel, args...)
	os.Exit(1)
}

// Fatalf 写日志后调用 os.Exit(1)，兼容 logrus 语义。
func (l *logger) Fatalf(f string, a ...any) {
	l.writef(zapcore.FatalLevel, f, a...)
	os.Exit(1)
}

// Panic 写日志后 panic，兼容 logrus 语义。
func (l *logger) Panic(args ...any) {
	l.write(zapcore.PanicLevel, args...)
	panic(fmt.Sprint(args...))
}

// Panicf 写日志后 panic，兼容 logrus 语义。
func (l *logger) Panicf(f string, a ...any) {
	l.writef(zapcore.PanicLevel, f, a...)
	panic(fmt.Sprintf(f, a...))
}

func (l *logger) WithField(key string, value any) contracts.Log {
	return &logEntry{sugar: l.sugar.With(key, value)}
}

func (l *logger) WithFields(fields map[string]any) contracts.Log {
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return &logEntry{sugar: l.sugar.With(args...)}
}

func (l *logger) WithError(err error) contracts.Log {
	return &logEntry{sugar: l.sugar.With("error", err)}
}

func (l *logger) WithContext(ctx context.Context) contracts.Log {
	// zap 的 SugaredLogger 不直接支持 Context
	// 基础实现：不做任何处理，子类可覆盖
	return &logEntry{sugar: l.sugar}
}

// Close 刷新缓冲区并关闭底层 writer（如 lumberjack）。
func (l *logger) Close() error {
	_ = l.base.Sync()
	var errs []error
	for _, c := range l.closers {
		if e := c.Close(); e != nil {
			errs = append(errs, e)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("[GoFast] 关闭日志失败: %v", errs)
	}
	return nil
}

// Printf 实现 GORM logger.Writer 接口。
func (l *logger) Printf(format string, args ...any) {
	if len(args) == 0 {
		// GORM 传已格式化字符串、零参数时原样 Info 输出
		l.Info(strings.TrimSuffix(format, "\n"))
	} else {
		l.Infof(format, args...)
	}
}

// =============================================================================
// logEntry 方法（链式调用）
// =============================================================================

func (e *logEntry) withCaller(level zapcore.Level, args ...any) {
	fields := callerFields()
	e.sugar.With(fieldsToAny(fields)...).Log(level, args...)
}

func (e *logEntry) withCallerf(level zapcore.Level, format string, args ...any) {
	fields := callerFields()
	e.sugar.With(fieldsToAny(fields)...).Logf(level, format, args...)
}

func (e *logEntry) Debug(args ...any)                 { e.withCaller(zapcore.DebugLevel, args...) }
func (e *logEntry) Debugf(f string, a ...any)         { e.withCallerf(zapcore.DebugLevel, f, a...) }
func (e *logEntry) Info(args ...any)                  { e.withCaller(zapcore.InfoLevel, args...) }
func (e *logEntry) Infof(f string, a ...any)          { e.withCallerf(zapcore.InfoLevel, f, a...) }
func (e *logEntry) Warn(args ...any)                  { e.withCaller(zapcore.WarnLevel, args...) }
func (e *logEntry) Warnf(f string, a ...any)          { e.withCallerf(zapcore.WarnLevel, f, a...) }
func (e *logEntry) Error(args ...any)                 { e.withCaller(zapcore.ErrorLevel, args...) }
func (e *logEntry) Errorf(f string, a ...any)         { e.withCallerf(zapcore.ErrorLevel, f, a...) }

// Fatal 写日志后 os.Exit(1)，兼容 logrus 语义（logEntry 与 logger 行为一致）。
func (e *logEntry) Fatal(args ...any) {
	e.withCaller(zapcore.FatalLevel, args...)
	os.Exit(1)
}

func (e *logEntry) Fatalf(f string, a ...any) {
	e.withCallerf(zapcore.FatalLevel, f, a...)
	os.Exit(1)
}

// Panic 写日志后 panic，兼容 logrus 语义。
func (e *logEntry) Panic(args ...any) {
	e.withCaller(zapcore.PanicLevel, args...)
	panic(fmt.Sprint(args...))
}

func (e *logEntry) Panicf(f string, a ...any) {
	e.withCallerf(zapcore.PanicLevel, f, a...)
	panic(fmt.Sprintf(f, a...))
}

func (e *logEntry) WithField(key string, value any) contracts.Log {
	return &logEntry{sugar: e.sugar.With(key, value)}
}

func (e *logEntry) WithFields(fields map[string]any) contracts.Log {
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return &logEntry{sugar: e.sugar.With(args...)}
}

func (e *logEntry) WithError(err error) contracts.Log {
	return &logEntry{sugar: e.sugar.With("error", err)}
}

func (e *logEntry) WithContext(ctx context.Context) contracts.Log {
	return &logEntry{sugar: e.sugar}
}

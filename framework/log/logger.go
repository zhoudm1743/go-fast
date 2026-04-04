package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/zhoudm1743/go-fast/framework/contracts"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// logger 实现 contracts.Log 接口，包装 logrus。
type logger struct {
	instance *logrus.Logger
}

// logEntry 包装 logrus.Entry 以实现 contracts.Log 链式调用。
type logEntry struct {
	entry *logrus.Entry
}

func callerFields(file string, line int, ok bool) logrus.Fields {
	callerFile := shorten(file, ok)
	return logrus.Fields{
		"caller_file": callerFile,
		"caller_line": line,
		"caller":      fmt.Sprintf("%s:%d", callerFile, line),
	}
}

func (e *logEntry) withCaller(level logrus.Level, args ...any) {
	_, file, line, ok := runtime.Caller(2)
	entry := e.entry.WithFields(callerFields(file, line, ok))
	entry.Log(level, args...)
}

func (e *logEntry) withCallerf(level logrus.Level, format string, args ...any) {
	_, file, line, ok := runtime.Caller(2)
	entry := e.entry.WithFields(callerFields(file, line, ok))
	entry.Logf(level, format, args...)
}

// NewLogger 根据配置创建 Logger 实例。
func NewLogger(cfg contracts.Config) (contracts.Log, error) {
	l := logrus.New()

	levelStr := cfg.GetString("log.level", "info")
	level, err := logrus.ParseLevel(levelStr)
	if err != nil {
		level = logrus.InfoLevel
	}

	mode := cfg.GetString("server.mode", "debug")
	if mode == "release" && level > logrus.InfoLevel {
		level = logrus.InfoLevel
	}
	l.SetLevel(level)
	// Caller is recorded by withCaller/withCallerf to avoid wrapper stack offset.
	l.SetReportCaller(false)

	format := cfg.GetString("log.format", "color")
	switch format {
	case "json":
		l.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	default:
		textFormatter := &logrus.TextFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			ForceColors:     format == "color",
			FullTimestamp:   true,
		}
		l.SetFormatter(&twoLineConsoleFormatter{base: textFormatter})
	}

	l.SetOutput(os.Stdout)

	outputPath := cfg.GetString("log.output_path")
	if outputPath != "" {
		logDir := filepath.Dir(outputPath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, fmt.Errorf("[GoFast] create log directory failed: %w", err)
		}

		fileWriter := &lumberjack.Logger{
			Filename:   outputPath,
			MaxSize:    cfg.GetInt("log.max_size", 100),
			MaxBackups: cfg.GetInt("log.max_backups", 5),
			MaxAge:     cfg.GetInt("log.max_age", 30),
			Compress:   cfg.GetBool("log.compress"),
			LocalTime:  true,
		}

		l.AddHook(&fileHook{
			writer: fileWriter,
			formatter: &logrus.JSONFormatter{
				TimestampFormat: "2006-01-02 15:04:05",
			},
		})
	}

	return &logger{instance: l}, nil
}

// fileHook 文件输出 Hook。
type fileHook struct {
	writer    io.Writer
	formatter logrus.Formatter
}

// twoLineConsoleFormatter prints caller on a dedicated line for easier click-to-open.
type twoLineConsoleFormatter struct {
	base logrus.Formatter
}

func (f *twoLineConsoleFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	caller, ok := entry.Data["caller"].(string)
	if !ok || caller == "" {
		return f.base.Format(entry)
	}

	data := make(logrus.Fields, len(entry.Data))
	for k, v := range entry.Data {
		switch k {
		case "caller", "caller_file", "caller_line":
			continue
		default:
			data[k] = v
		}
	}

	clone := &logrus.Entry{
		Logger:  entry.Logger,
		Data:    data,
		Time:    entry.Time,
		Level:   entry.Level,
		Message: entry.Message,
		Context: entry.Context,
		Caller:  entry.Caller,
	}

	body, err := f.base.Format(clone)
	if err != nil {
		return nil, err
	}

	return append([]byte(caller+"\n"), body...), nil
}

func (h *fileHook) Levels() []logrus.Level { return logrus.AllLevels }
func (h *fileHook) Fire(entry *logrus.Entry) error {
	b, err := h.formatter.Format(entry)
	if err != nil {
		return err
	}
	_, err = h.writer.Write(b)
	return err
}

func (l *logger) withCaller(level logrus.Level, args ...any) {
	_, file, line, ok := runtime.Caller(2)
	entry := l.instance.WithFields(callerFields(file, line, ok))
	entry.Log(level, args...)
}

func (l *logger) withCallerf(level logrus.Level, format string, args ...any) {
	_, file, line, ok := runtime.Caller(2)
	entry := l.instance.WithFields(callerFields(file, line, ok))
	entry.Logf(level, format, args...)
}

func shorten(file string, ok bool) string {
	if !ok {
		return "unknown"
	}
	return filepath.ToSlash(filepath.Clean(file))
}

func (l *logger) Debug(args ...any) { l.withCaller(logrus.DebugLevel, args...) }
func (l *logger) Debugf(format string, args ...any) {
	l.withCallerf(logrus.DebugLevel, format, args...)
}
func (l *logger) Info(args ...any)                 { l.withCaller(logrus.InfoLevel, args...) }
func (l *logger) Infof(format string, args ...any) { l.withCallerf(logrus.InfoLevel, format, args...) }
func (l *logger) Warn(args ...any)                 { l.withCaller(logrus.WarnLevel, args...) }
func (l *logger) Warnf(format string, args ...any) { l.withCallerf(logrus.WarnLevel, format, args...) }
func (l *logger) Error(args ...any)                { l.withCaller(logrus.ErrorLevel, args...) }
func (l *logger) Errorf(format string, args ...any) {
	l.withCallerf(logrus.ErrorLevel, format, args...)
}
func (l *logger) Fatal(args ...any)                 { l.instance.Fatal(args...) }
func (l *logger) Fatalf(format string, args ...any) { l.instance.Fatalf(format, args...) }
func (l *logger) Panic(args ...any)                 { l.instance.Panic(args...) }
func (l *logger) Panicf(format string, args ...any) { l.instance.Panicf(format, args...) }

func (l *logger) WithField(key string, value any) contracts.Log {
	return &logEntry{entry: l.instance.WithField(key, value)}
}
func (l *logger) WithFields(fields map[string]any) contracts.Log {
	return &logEntry{entry: l.instance.WithFields(fields)}
}

func (e *logEntry) Debug(args ...any) { e.withCaller(logrus.DebugLevel, args...) }
func (e *logEntry) Debugf(format string, args ...any) {
	e.withCallerf(logrus.DebugLevel, format, args...)
}
func (e *logEntry) Info(args ...any) { e.withCaller(logrus.InfoLevel, args...) }
func (e *logEntry) Infof(format string, args ...any) {
	e.withCallerf(logrus.InfoLevel, format, args...)
}
func (e *logEntry) Warn(args ...any) { e.withCaller(logrus.WarnLevel, args...) }
func (e *logEntry) Warnf(format string, args ...any) {
	e.withCallerf(logrus.WarnLevel, format, args...)
}
func (e *logEntry) Error(args ...any) { e.withCaller(logrus.ErrorLevel, args...) }
func (e *logEntry) Errorf(format string, args ...any) {
	e.withCallerf(logrus.ErrorLevel, format, args...)
}
func (e *logEntry) Fatal(args ...any)                 { e.entry.Fatal(args...) }
func (e *logEntry) Fatalf(format string, args ...any) { e.entry.Fatalf(format, args...) }
func (e *logEntry) Panic(args ...any)                 { e.entry.Panic(args...) }
func (e *logEntry) Panicf(format string, args ...any) { e.entry.Panicf(format, args...) }

func (e *logEntry) WithField(key string, value any) contracts.Log {
	return &logEntry{entry: e.entry.WithField(key, value)}
}
func (e *logEntry) WithFields(fields map[string]any) contracts.Log {
	return &logEntry{entry: e.entry.WithFields(fields)}
}

// Printf 实现 GORM logger.Writer 接口。
func (l *logger) Printf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	msg = strings.TrimSuffix(msg, "%!s(MISSING)")
	l.instance.Info(msg)
}

package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"go-fast/framework/contracts"

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
	l.SetReportCaller(true)

	format := cfg.GetString("log.format", "color")
	switch format {
	case "json":
		l.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	default:
		l.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			ForceColors:     format == "color",
			FullTimestamp:   true,
		})
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
	entry := l.instance.WithField("caller_file", shorten(file, ok)).WithField("caller_line", line)
	entry.Log(level, args...)
}

func (l *logger) withCallerf(level logrus.Level, format string, args ...any) {
	_, file, line, ok := runtime.Caller(2)
	entry := l.instance.WithField("caller_file", shorten(file, ok)).WithField("caller_line", line)
	entry.Logf(level, format, args...)
}

func shorten(file string, ok bool) string {
	if !ok {
		return "unknown"
	}
	return filepath.Base(file)
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

func (e *logEntry) Debug(args ...any)                 { e.entry.Debug(args...) }
func (e *logEntry) Debugf(format string, args ...any) { e.entry.Debugf(format, args...) }
func (e *logEntry) Info(args ...any)                  { e.entry.Info(args...) }
func (e *logEntry) Infof(format string, args ...any)  { e.entry.Infof(format, args...) }
func (e *logEntry) Warn(args ...any)                  { e.entry.Warn(args...) }
func (e *logEntry) Warnf(format string, args ...any)  { e.entry.Warnf(format, args...) }
func (e *logEntry) Error(args ...any)                 { e.entry.Error(args...) }
func (e *logEntry) Errorf(format string, args ...any) { e.entry.Errorf(format, args...) }
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

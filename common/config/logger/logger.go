package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Log is the global logger instance.
// It is safe for concurrent use.
var Log *logrus.Logger

var (
	logFile *os.File
	mu      sync.Mutex // protects logFile during Init/Shutdown
)

func init() {
	// Ensure a working default logger before Init() is called
	Log = logrus.New()
	Log.SetLevel(logrus.InfoLevel)
	Log.SetFormatter(newFormatter())
	Log.SetOutput(os.Stdout)
}

// ---------------------------------------------------------------------------
// Formatter
// ---------------------------------------------------------------------------

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
	colorGray   = "\033[90m"
)

// compactFormatter produces clean, single-line log output.
type compactFormatter struct {
	enableColor bool
}

func newFormatter() *compactFormatter {
	return &compactFormatter{
		enableColor: isTerminal(),
	}
}

// isTerminal checks if stdout is a terminal (not piped/redirected)
func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// levelTag returns a fixed-width 5-char uppercase level tag, e.g. "INFO ", "DEBUG".
func levelTag(l logrus.Level) string {
	s := strings.ToUpper(l.String())
	if len(s) >= 5 {
		return s[:5]
	}
	return s + strings.Repeat(" ", 5-len(s))
}

// levelColor returns the ANSI color for a given log level
func levelColor(l logrus.Level) string {
	switch l {
	case logrus.TraceLevel:
		return colorGray
	case logrus.DebugLevel:
		return colorCyan
	case logrus.InfoLevel:
		return colorGreen
	case logrus.WarnLevel:
		return colorYellow
	case logrus.ErrorLevel:
		return colorRed
	case logrus.FatalLevel, logrus.PanicLevel:
		return colorPurple
	default:
		return colorWhite
	}
}

func (f *compactFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := entry.Time.Format("2006-01-02 15:04:05")
	tag := levelTag(entry.Level)

	// Resolve caller
	caller := "unknown"
	for i := 1; i < 15; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		if isInternalFrame(file) {
			continue
		}
		caller = formatCaller(file, line)
		break
	}

	var b strings.Builder
	b.Grow(256)

	if f.enableColor {
		c := levelColor(entry.Level)
		b.WriteString(colorGray)
		b.WriteString(timestamp)
		b.WriteString(colorReset)

		b.WriteString(" ")
		b.WriteString(c)
		b.WriteString("[")
		b.WriteString(tag)
		b.WriteString("]")
		b.WriteString(colorReset)

		b.WriteString(" ")
		b.WriteString(colorGray)
		b.WriteString(caller)
		b.WriteString(colorReset)

		b.WriteString(" ")
		b.WriteString(entry.Message)
	} else {
		b.WriteString(timestamp)
		b.WriteString(" [")
		b.WriteString(tag)
		b.WriteString("] ")
		b.WriteString(caller)
		b.WriteByte(' ')
		b.WriteString(entry.Message)
	}

	// Append structured fields
	if len(entry.Data) > 0 {
		if f.enableColor {
			b.WriteString(" ")
			b.WriteString(colorCyan)
			b.WriteString("| ")
			b.WriteString(colorReset)
		} else {
			b.WriteString(" | ")
		}
		i := 0
		for k, v := range entry.Data {
			if i > 0 {
				b.WriteString(", ")
			}
			if f.enableColor {
				fmt.Fprintf(&b, "%s%s%s=%v", colorCyan, k, colorReset, v)
			} else {
				fmt.Fprintf(&b, "%s=%v", k, v)
			}
			i++
		}
	}
	b.WriteByte('\n')

	return []byte(b.String()), nil
}

// isInternalFrame returns true for stack frames that belong to logrus or this package.
func isInternalFrame(file string) bool {
	// logrus internals
	if strings.Contains(file, "sirupsen/logrus") {
		return true
	}
	// This package itself – match both "logger/logger.go" and "logger/wrapper.go" etc.
	// We use the directory name to be robust.
	dir := filepath.Base(filepath.Dir(file))
	if dir == "logger" {
		return true
	}
	return false
}

// formatCaller builds a short caller string like "service/handler.go:42".
func formatCaller(file string, line int) string {
	dir := filepath.Base(filepath.Dir(file))
	base := filepath.Base(file)
	return fmt.Sprintf("%s/%s:%d", dir, base, line)
}

// ---------------------------------------------------------------------------
// Initialization
// ---------------------------------------------------------------------------

// Init initializes the global logger.
// Pass nil for default config (console, info level, no file).
func Init(cfg *Config) error {
	mu.Lock()
	defer mu.Unlock()

	if cfg == nil {
		cfg = &Config{}
	}
	cfg.SetDefaults()

	if logFile != nil {
		logFile.Close()
		logFile = nil
	}

	Log = logrus.New()

	level, _ := logrus.ParseLevel(cfg.Level)
	Log.SetLevel(level)

	Log.SetFormatter(newFormatter())

	writers, file, err := buildWriters(cfg)
	if err != nil {
		return err
	}
	logFile = file
	Log.SetOutput(io.MultiWriter(writers...))

	if cfg.ServiceName != "" {
		Log.AddHook(&defaultFieldHook{fields: logrus.Fields{"service": cfg.ServiceName}})
	}
	return nil
}

func buildWriters(cfg *Config) ([]io.Writer, *os.File, error) {
	var writers []io.Writer
	if cfg.ConsoleOutput == nil || *cfg.ConsoleOutput {
		writers = append(writers, os.Stdout)
	}

	fileWriter, err := createFileWriter(cfg.LogDir, cfg.LogFileName)
	if err != nil {
		return nil, nil, err
	}
	if fileWriter != nil {
		writers = append(writers, fileWriter)
	}

	if len(writers) == 0 {
		writers = append(writers, os.Stdout)
	}
	return writers, fileWriter.(*os.File), nil
}

func createFileWriter(logDir, fileName string) (io.Writer, error) {
	if logDir == "" {
		return nil, nil
	}
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}
	if fileName == "" {
		fileName = fmt.Sprintf("axis_%s.log", time.Now().Format("2006-01-02"))
	}
	fp := filepath.Join(logDir, fileName)
	f, err := os.OpenFile(fp, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// defaultFieldHook injects fields into every log entry.
type defaultFieldHook struct {
	fields logrus.Fields
}

func (h *defaultFieldHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *defaultFieldHook) Fire(entry *logrus.Entry) error {
	for k, v := range h.fields {
		// Don't overwrite if already set
		if _, exists := entry.Data[k]; !exists {
			entry.Data[k] = v
		}
	}
	return nil
}

// InitWithSimpleConfig is a convenience wrapper.
func InitWithSimpleConfig(level string) error {
	return Init(&Config{Level: level})
}

// Shutdown closes the log file gracefully.
func Shutdown() {
	mu.Lock()
	defer mu.Unlock()
	if logFile != nil {
		logFile.Close()
		logFile = nil
	}
}

// ---------------------------------------------------------------------------
// Package-level convenience functions
// ---------------------------------------------------------------------------

// WithFields returns a new Entry with structured fields.
func WithFields(fields logrus.Fields) *logrus.Entry {
	return Log.WithFields(fields)
}

// WithField returns a new Entry with a single field.
func WithField(key string, value interface{}) *logrus.Entry {
	return Log.WithField(key, value)
}

// ---------------------------Basic functions---------------------------

// --- Trace ---
func Trace(args ...interface{})                 { Log.Trace(args...) }
func Tracef(format string, args ...interface{}) { Log.Tracef(format, args...) }

// --- Debug ---
func Debug(args ...interface{})                 { Log.Debug(args...) }
func Debugf(format string, args ...interface{}) { Log.Debugf(format, args...) }

// --- Info ---
func Info(args ...interface{})                 { Log.Info(args...) }
func Infof(format string, args ...interface{}) { Log.Infof(format, args...) }

// --- Warn ---
func Warn(args ...interface{})                 { Log.Warn(args...) }
func Warnf(format string, args ...interface{}) { Log.Warnf(format, args...) }

// --- Error ---
func Error(args ...interface{})                 { Log.Error(args...) }
func Errorf(format string, args ...interface{}) { Log.Errorf(format, args...) }

// --- Fatal ---
func Fatal(args ...interface{})                 { Log.Fatal(args...) }
func Fatalf(format string, args ...interface{}) { Log.Fatalf(format, args...) }

// --- Panic ---
func Panic(args ...interface{})                 { Log.Panic(args...) }
func Panicf(format string, args ...interface{}) { Log.Panicf(format, args...) }

// ---------------------------Error-aware convenience functions---------------------------

// WithError creates a log entry with an error field.
// Usage: logger.WithError(err).Fatal("failed")
func WithError(err error) *logrus.Entry {
	return Log.WithError(err)
}

// ---------------------------Println-aware convenience functions---------------------------

func Println(args interface{}) {
	Log.Println(args)
}

func Entry() *logrus.Entry {
	return logrus.NewEntry(Log) // Log 是你的全局 *logrus.Logger
}

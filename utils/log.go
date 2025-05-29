package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Logger interface {
	Debug(msg string)
	Info(msg string)
	Error(msg string)
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	DebugCtx(ctx context.Context, msg string)
	InfoCtx(ctx context.Context, msg string)
	ErrorCtx(ctx context.Context, msg string)
	WithFields(fields map[string]interface{}) Logger
	WithField(key string, value interface{}) Logger
	SetLevel(level Level)
	SetFormat(format Format)
	SetOutput(w io.Writer)
}

type cilLogger struct {
	name   string
	fields map[string]interface{}
	level  Level
	format Format
	output io.Writer
	lock   sync.RWMutex
}

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelError
)

type Format int

const (
	FormatText Format = iota
	FormatJSON
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

func (l Level) Color() string {
	switch l {
	case LevelDebug:
		return "\033[36m" // 青色
	case LevelInfo:
		return "\033[32m" // 绿色
	case LevelError:
		return "\033[31m" // 红色
	default:
		return "\033[0m"
	}
}

func New(name string) Logger {
	format := FormatText
	if os.Getenv("ENV") == "prod" {
		format = FormatJSON
	}
	return &cilLogger{
		name:   name,
		fields: make(map[string]interface{}),
		level:  LevelDebug,
		output: os.Stdout,
		format: format,
	}
}

func (l *cilLogger) clone() *cilLogger {
	l.lock.RLock()
	defer l.lock.RUnlock()

	newFields := make(map[string]interface{}, len(l.fields))
	for k, v := range l.fields {
		newFields[k] = v
	}
	return &cilLogger{
		name:   l.name,
		fields: newFields,
		level:  l.level,
		output: l.output,
		format: l.format,
	}
}

func (l *cilLogger) WithFields(fields map[string]interface{}) Logger {
	clone := l.clone()
	for k, v := range fields {
		clone.fields[k] = v
	}
	return clone
}

func (l *cilLogger) WithField(key string, value interface{}) Logger {
	return l.WithFields(map[string]interface{}{key: value})
}

func (l *cilLogger) SetLevel(level Level) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.level = level
}

func (l *cilLogger) SetFormat(format Format) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.format = format
}

func (l *cilLogger) SetOutput(w io.Writer) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.output = w
}

func (l *cilLogger) shouldLog(level Level) bool {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return level >= l.level
}

func (l *cilLogger) formatLog(level Level, msg string) string {
	l.lock.RLock()
	defer l.lock.RUnlock()
	timeStr := time.Now().Format("2006-01-02 15:04:05")

	if l.format == FormatJSON {
		data := map[string]interface{}{
			"level":   level.String(),
			"time":    timeStr,
			"name":    l.name,
			"message": msg,
			"fields":  l.fields,
		}
		bytes, _ := json.Marshal(data)
		return string(bytes) + "\n"
	}

	color := level.Color()
	reset := "\033[0m"

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s[%s] [%s] [%s]", color, level.String(), timeStr, l.name))

	keys := make([]string, 0, len(l.fields))
	for k := range l.fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		sb.WriteString(fmt.Sprintf(" %s=%v", k, l.fields[k]))
	}

	sb.WriteString(fmt.Sprintf(" | %s%s\n", msg, reset))
	return sb.String()
}

func (l *cilLogger) log(level Level, msg string) {
	if !l.shouldLog(level) {
		return
	}
	l.lock.RLock()
	output := l.output
	l.lock.RUnlock()
	_, _ = output.Write([]byte(l.formatLog(level, msg)))
}

func (l *cilLogger) Debug(msg string) { l.log(LevelDebug, msg) }
func (l *cilLogger) Info(msg string)  { l.log(LevelInfo, msg) }
func (l *cilLogger) Error(msg string) { l.log(LevelError, msg) }
func (l *cilLogger) Debugf(format string, args ...interface{}) {
	l.Debug(fmt.Sprintf(format, args...))
}

func (l *cilLogger) Infof(format string, args ...interface{}) {
	l.Info(fmt.Sprintf(format, args...))
}

func (l *cilLogger) Errorf(format string, args ...interface{}) {
	l.Error(fmt.Sprintf(format, args...))
}

type ctxKey string

const traceIDKey ctxKey = "trace_id"

func NewTraceID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}

func TraceIDFromContext(ctx context.Context) string {
	if v := ctx.Value(traceIDKey); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return NewTraceID()
}

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

func (l *cilLogger) DebugCtx(ctx context.Context, msg string) {
	l.logCtx(ctx, LevelDebug, msg)
}

func (l *cilLogger) InfoCtx(ctx context.Context, msg string) {
	l.logCtx(ctx, LevelInfo, msg)
}

func (l *cilLogger) ErrorCtx(ctx context.Context, msg string) {
	l.logCtx(ctx, LevelError, msg)
}

func (l *cilLogger) logCtx(ctx context.Context, level Level, msg string) {
	traceID := TraceIDFromContext(ctx)
	loggerWithTrace := l.WithField("trace_id", traceID)
	loggerWithTrace.(*cilLogger).log(level, msg)
}

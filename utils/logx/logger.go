package logx

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/moscososirenita-design/tk-common/utils/ctxx"
)

// LogLevel 定义日志等级。
type LogLevel int

// 声明日志等级常量。
const (
	// LevelDebug 记录调试日志。
	LevelDebug LogLevel = iota
	// LevelInfo 记录信息日志。
	LevelInfo
	// LevelWarn 记录警告日志。
	LevelWarn
	// LevelError 记录错误日志。
	LevelError
)

// Logger 定义基础日志记录器。
type Logger struct {
	// Level 保存当前日志等级。
	Level LogLevel
	// debugLog 处理 debug 级别输出。
	debugLog *log.Logger
	// infoLog 处理 info 级别输出。
	infoLog *log.Logger
	// warnLog 处理 warn 级别输出。
	warnLog *log.Logger
	// errorLog 处理 error 级别输出。
	errorLog *log.Logger
	// file 记录文件句柄，便于优雅关闭。
	file *os.File
}

// Config 定义日志配置。
type Config struct {
	Level      LogLevel  // 日志等级
	Output     io.Writer // 输出目标
	FilePath   string    // 文件路径
	MaxSize    int64     // 最大文件大小（预留）
	MaxBackups int       // 最大备份数（预留）
	MaxAge     int       // 最大保存天数（预留）
}

// LogConfig 与历史命名保持兼容。
type LogConfig = Config

// DefaultConfig 返回默认日志配置。
func DefaultConfig() Config {
	// 返回当前处理结果。
	return Config{
		// 设置默认等级为 info。
		Level: LevelInfo,
		// 默认输出到标准输出。
		Output: os.Stdout,
		// 默认不写入文件。
		FilePath: "",
	}
}

// DefaultLogConfig 与历史命名保持兼容。
func DefaultLogConfig() Config {
	// 返回当前处理结果。
	return DefaultConfig()
}

// NewLogger 创建日志记录器。
func NewLogger(cfg Config) (*Logger, error) {
	// 声明当前变量。
	var output io.Writer = cfg.Output
	// 声明当前变量。
	var file *os.File

	// 判断条件并进入对应分支逻辑。
	if output == nil {
		// 未设置输出时，回退到标准输出。
		output = os.Stdout
	}

	// 判断条件并进入对应分支逻辑。
	if cfg.FilePath != "" {
		// 确保日志目录存在。
		dir := filepath.Dir(cfg.FilePath)
		// 判断条件并进入对应分支逻辑。
		if err := os.MkdirAll(dir, 0755); err != nil {
			// 返回当前处理结果。
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		// 以追加模式打开日志文件。
		f, err := os.OpenFile(cfg.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		// 判断条件并进入对应分支逻辑。
		if err != nil {
			// 返回当前处理结果。
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		// 更新当前变量或字段值。
		file = f
		// 更新当前变量或字段值。
		output = f

		// 若原输出是终端，则同时写终端和文件。
		if cfg.Output == os.Stdout || cfg.Output == os.Stderr {
			// 更新当前变量或字段值。
			output = io.MultiWriter(cfg.Output, f)
		}
	}

	// 返回当前处理结果。
	return &Logger{
		// 写入配置的日志等级。
		Level: cfg.Level,
		// 初始化 debug 日志器。
		debugLog: log.New(output, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile),
		// 初始化 info 日志器。
		infoLog: log.New(output, "[INFO]  ", log.Ldate|log.Ltime|log.Lshortfile),
		// 初始化 warn 日志器。
		warnLog: log.New(output, "[WARN]  ", log.Ldate|log.Ltime|log.Lshortfile),
		// 初始化 error 日志器。
		errorLog: log.New(output, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile),
		// 保存文件句柄用于关闭。
		file: file,
	}, nil
}

// ContextLogger 定义带上下文的日志记录器。
type ContextLogger struct {
	*Logger
	// ctx 保存请求上下文，便于输出 request_id。
	ctx context.Context
}

// NewContextLogger 创建带上下文日志记录器。
func NewContextLogger(ctx context.Context, logger *Logger) *ContextLogger {
	// 返回当前处理结果。
	return &ContextLogger{
		// 注入底层 Logger。
		Logger: logger,
		// 注入上下文。
		ctx: ctx,
	}
}

// WithContext 绑定上下文。
func (l *Logger) WithContext(ctx context.Context) *ContextLogger {
	// 返回当前处理结果。
	return NewContextLogger(ctx, l)
}

// getCallerInfo 获取调用位置信息。
func getCallerInfo() string {
	// 跳过包装层栈帧，返回真实调用点。
	_, file, line, ok := runtime.Caller(3)
	// 判断条件并进入对应分支逻辑。
	if !ok {
		// 返回当前处理结果。
		return ""
	}
	// 返回当前处理结果。
	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
}

// buildContextPrefix 构建 request_id 前缀。
func (cl *ContextLogger) buildContextPrefix() string {
	// 判断条件并进入对应分支逻辑。
	if cl == nil || cl.ctx == nil {
		// 返回当前处理结果。
		return ""
	}
	// 定义并初始化当前变量。
	reqID := ctxx.RequestIDFromContext(cl.ctx)
	// 判断条件并进入对应分支逻辑。
	if reqID == "" {
		// 返回当前处理结果。
		return ""
	}
	// 返回当前处理结果。
	return fmt.Sprintf("[req:%s] ", reqID)
}

// Debug 记录调试日志。
func (l *Logger) Debug(format string, v ...interface{}) {
	// 判断条件并进入对应分支逻辑。
	if l.Level > LevelDebug {
		// 返回当前处理结果。
		return
	}
	// 定义并初始化当前变量。
	caller := getCallerInfo()
	// 判断条件并进入对应分支逻辑。
	if caller != "" {
		// 更新当前变量或字段值。
		format = fmt.Sprintf("%s %s", caller, format)
	}
	// 输出 debug 日志。
	l.debugLog.Printf(format, v...)
}

// Info 记录信息日志。
func (l *Logger) Info(format string, v ...interface{}) {
	// 判断条件并进入对应分支逻辑。
	if l.Level > LevelInfo {
		// 返回当前处理结果。
		return
	}
	// 定义并初始化当前变量。
	caller := getCallerInfo()
	// 判断条件并进入对应分支逻辑。
	if caller != "" {
		// 更新当前变量或字段值。
		format = fmt.Sprintf("%s %s", caller, format)
	}
	// 输出 info 日志。
	l.infoLog.Printf(format, v...)
}

// Warn 记录警告日志。
func (l *Logger) Warn(format string, v ...interface{}) {
	// 判断条件并进入对应分支逻辑。
	if l.Level > LevelWarn {
		// 返回当前处理结果。
		return
	}
	// 定义并初始化当前变量。
	caller := getCallerInfo()
	// 判断条件并进入对应分支逻辑。
	if caller != "" {
		// 更新当前变量或字段值。
		format = fmt.Sprintf("%s %s", caller, format)
	}
	// 输出 warn 日志。
	l.warnLog.Printf(format, v...)
}

// Error 记录错误日志。
func (l *Logger) Error(format string, v ...interface{}) {
	// 判断条件并进入对应分支逻辑。
	if l.Level > LevelError {
		// 返回当前处理结果。
		return
	}
	// 定义并初始化当前变量。
	caller := getCallerInfo()
	// 判断条件并进入对应分支逻辑。
	if caller != "" {
		// 更新当前变量或字段值。
		format = fmt.Sprintf("%s %s", caller, format)
	}
	// 输出 error 日志。
	l.errorLog.Printf(format, v...)
}

// Fatal 输出致命日志并退出进程。
func (l *Logger) Fatal(format string, v ...interface{}) {
	// 定义并初始化当前变量。
	caller := getCallerInfo()
	// 判断条件并进入对应分支逻辑。
	if caller != "" {
		// 更新当前变量或字段值。
		format = fmt.Sprintf("%s %s", caller, format)
	}
	// 输出 fatal 日志并终止进程。
	l.errorLog.Fatalf(format, v...)
}

// Debug 记录带上下文的调试日志。
func (cl *ContextLogger) Debug(format string, v ...interface{}) {
	// 判断条件并进入对应分支逻辑。
	if cl.Level > LevelDebug {
		// 返回当前处理结果。
		return
	}
	// 定义并初始化当前变量。
	prefix := cl.buildContextPrefix()
	// 定义并初始化当前变量。
	caller := getCallerInfo()
	// 判断条件并进入对应分支逻辑。
	if caller != "" {
		// 更新当前变量或字段值。
		format = fmt.Sprintf("%s%s %s", prefix, caller, format)
	} else {
		// 更新当前变量或字段值。
		format = fmt.Sprintf("%s%s", prefix, format)
	}
	// 输出 debug 日志。
	cl.debugLog.Printf(format, v...)
}

// Info 记录带上下文的信息日志。
func (cl *ContextLogger) Info(format string, v ...interface{}) {
	// 判断条件并进入对应分支逻辑。
	if cl.Level > LevelInfo {
		// 返回当前处理结果。
		return
	}
	// 定义并初始化当前变量。
	prefix := cl.buildContextPrefix()
	// 定义并初始化当前变量。
	caller := getCallerInfo()
	// 判断条件并进入对应分支逻辑。
	if caller != "" {
		// 更新当前变量或字段值。
		format = fmt.Sprintf("%s%s %s", prefix, caller, format)
	} else {
		// 更新当前变量或字段值。
		format = fmt.Sprintf("%s%s", prefix, format)
	}
	// 输出 info 日志。
	cl.infoLog.Printf(format, v...)
}

// Warn 记录带上下文的警告日志。
func (cl *ContextLogger) Warn(format string, v ...interface{}) {
	// 判断条件并进入对应分支逻辑。
	if cl.Level > LevelWarn {
		// 返回当前处理结果。
		return
	}
	// 定义并初始化当前变量。
	prefix := cl.buildContextPrefix()
	// 定义并初始化当前变量。
	caller := getCallerInfo()
	// 判断条件并进入对应分支逻辑。
	if caller != "" {
		// 更新当前变量或字段值。
		format = fmt.Sprintf("%s%s %s", prefix, caller, format)
	} else {
		// 更新当前变量或字段值。
		format = fmt.Sprintf("%s%s", prefix, format)
	}
	// 输出 warn 日志。
	cl.warnLog.Printf(format, v...)
}

// Error 记录带上下文的错误日志。
func (cl *ContextLogger) Error(format string, v ...interface{}) {
	// 判断条件并进入对应分支逻辑。
	if cl.Level > LevelError {
		// 返回当前处理结果。
		return
	}
	// 定义并初始化当前变量。
	prefix := cl.buildContextPrefix()
	// 定义并初始化当前变量。
	caller := getCallerInfo()
	// 判断条件并进入对应分支逻辑。
	if caller != "" {
		// 更新当前变量或字段值。
		format = fmt.Sprintf("%s%s %s", prefix, caller, format)
	} else {
		// 更新当前变量或字段值。
		format = fmt.Sprintf("%s%s", prefix, format)
	}
	// 输出 error 日志。
	cl.errorLog.Printf(format, v...)
}

// Close 关闭日志文件句柄。
func (l *Logger) Close() error {
	// 判断条件并进入对应分支逻辑。
	if l.file != nil {
		// 返回当前处理结果。
		return l.file.Close()
	}
	// 返回当前处理结果。
	return nil
}

// 声明全局日志实例与锁，避免并发读写冲突。
var (
	// globalMu 保护 globalLogger。
	globalMu sync.RWMutex
	// globalLogger 保存进程级 logger。
	globalLogger *Logger
)

// InitGlobalLogger 初始化全局日志器。
func InitGlobalLogger(cfg Config) error {
	// 定义并初始化当前变量。
	logger, err := NewLogger(cfg)
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return err
	}
	// 加锁写入全局实例。
	globalMu.Lock()
	// 更新当前变量或字段值。
	globalLogger = logger
	// 调用globalMu.Unlock完成当前处理。
	globalMu.Unlock()
	// 返回当前处理结果。
	return nil
}

// GetLogger 获取全局日志器；若未初始化则使用默认配置。
func GetLogger() *Logger {
	// 先尝试读锁快速路径。
	globalMu.RLock()
	// 定义并初始化当前变量。
	current := globalLogger
	// 调用globalMu.RUnlock完成当前处理。
	globalMu.RUnlock()
	// 判断条件并进入对应分支逻辑。
	if current != nil {
		// 返回当前处理结果。
		return current
	}

	// 慢路径：初始化默认 logger。
	globalMu.Lock()
	// 注册延迟执行逻辑。
	defer globalMu.Unlock()
	// 判断条件并进入对应分支逻辑。
	if globalLogger == nil {
		// 定义并初始化当前变量。
		logger, _ := NewLogger(DefaultConfig())
		// 更新当前变量或字段值。
		globalLogger = logger
	}
	// 返回当前处理结果。
	return globalLogger
}

// LoggerFromContext 从上下文提取日志器。
func LoggerFromContext(ctx context.Context) *ContextLogger {
	// 判断条件并进入对应分支逻辑。
	if ctxLogger, ok := ctxx.Get[*ContextLogger](ctx, ctxx.LoggerKey); ok && ctxLogger != nil {
		// 返回当前处理结果。
		return ctxLogger
	}
	// 兼容历史字符串键读取。
	if ctx != nil {
		// 判断条件并进入对应分支逻辑。
		if logger, ok := ctx.Value("logger").(*ContextLogger); ok {
			// 返回当前处理结果。
			return logger
		}
	}
	// 返回当前处理结果。
	return GetLogger().WithContext(ctx)
}

// WithContextLogger 写入上下文日志器。
func WithContextLogger(ctx context.Context, logger *ContextLogger) context.Context {
	// 返回当前处理结果。
	return ctxx.With(ctx, ctxx.LoggerKey, logger)
}

// WithRequestID 写入请求 ID。
func WithRequestID(ctx context.Context, requestID string) context.Context {
	// 返回当前处理结果。
	return ctxx.With(ctx, ctxx.RequestIDKey, requestID)
}

// LogLevelFromString 将字符串转为日志级别。
func LogLevelFromString(level string) LogLevel {
	// 根据表达式进入多分支处理。
	switch level {
	case "debug":
		// 返回当前处理结果。
		return LevelDebug
	case "info":
		// 返回当前处理结果。
		return LevelInfo
	case "warn":
		// 返回当前处理结果。
		return LevelWarn
	case "error":
		// 返回当前处理结果。
		return LevelError
	default:
		// 返回当前处理结果。
		return LevelInfo
	}
}

// StringFromLogLevel 将日志级别转为字符串。
func StringFromLogLevel(level LogLevel) string {
	// 根据表达式进入多分支处理。
	switch level {
	case LevelDebug:
		// 返回当前处理结果。
		return "debug"
	case LevelInfo:
		// 返回当前处理结果。
		return "info"
	case LevelWarn:
		// 返回当前处理结果。
		return "warn"
	case LevelError:
		// 返回当前处理结果。
		return "error"
	default:
		// 返回当前处理结果。
		return "info"
	}
}

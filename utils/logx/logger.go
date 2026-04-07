package logx

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/moscososirenita-design/tk-common/utils/ctxx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LogLevel 定义日志等级（与 zapcore.Level 数值对齐：Debug=0 Info=1 Warn=2 Error=3）。
type LogLevel int

// 声明日志等级常量。
const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

// Config 定义日志配置。
type Config struct {
	Level             LogLevel // 日志等级
	FilePath          string   // 预留：文件输出路径
	MaxSize           int64    // 预留：最大文件大小
	MaxBackups        int      // 预留：最大备份数
	MaxAge            int      // 预留：最大保存天数
	Development       bool     // 开发模式：彩色输出 + panic 级别 stacktrace
	ServerName        string   // 服务名，写入每条日志的 appName 字段
	Version           string   // 服务版本
	DisableStacktrace bool     // 禁用自动 stacktrace（配合 StacktraceLevel 使用）
	StacktraceLevel   *int     // 指定触发 stacktrace 的最低级别（zapcore.Level）
}

// LogConfig 兼容历史命名。
type LogConfig = Config

// DefaultConfig 返回默认日志配置。
func DefaultConfig() Config {
	return Config{Level: LevelInfo}
}

// DefaultLogConfig 兼容历史命名。
func DefaultLogConfig() Config { return DefaultConfig() }

// ── 环境检测 ─────────────────────────────────────────────────────────────────

// isDevMode 检查 APP_ENV 环境变量，local/dev/development 视为本地开发环境。
func isDevMode() bool {
	env := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	return env == "local" || env == "dev" || env == "development"
}

// ── 彩色输出支持 ──────────────────────────────────────────────────────────────

type color string

func (c color) add(s string) string {
	return fmt.Sprintf(string(c), 0x1B, s, 0x1B)
}

const (
	colorMagenta color = "%c[0;42;35m%v%c[0m"
	colorBlue    color = "%c[0;42;30m%v%c[0m"
	colorYellow  color = "%c[0;43;30m%v%c[0m"
	colorRed     color = "%c[0;41;30m%v%c[0m"
)

var levelColors = map[zapcore.Level]color{
	zapcore.DebugLevel:  colorMagenta,
	zapcore.InfoLevel:   colorBlue,
	zapcore.WarnLevel:   colorYellow,
	zapcore.ErrorLevel:  colorRed,
	zapcore.DPanicLevel: colorRed,
	zapcore.PanicLevel:  colorRed,
	zapcore.FatalLevel:  colorRed,
}

func colorLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	c, ok := levelColors[l]
	if !ok {
		c = colorRed
	}
	enc.AppendString(c.add(l.CapitalString()))
}

// ── 内部 zap.Logger 构造 ──────────────────────────────────────────────────────

func buildZapLogger(cfg Config) *zap.Logger {
	encCfg := zap.NewProductionEncoderConfig()
	encCfg.EncodeTime = zapcore.TimeEncoder(func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(fmt.Sprintf("%v %v", t.Format("2006-01-02 15:04:05.000Z07"), t.UnixMilli()))
	})
	dev := cfg.Development || isDevMode()
	if dev {
		encCfg.EncodeLevel = colorLevelEncoder
	} else {
		encCfg.EncodeLevel = zapcore.CapitalLevelEncoder
	}
	encCfg.EncodeCaller = zapcore.ShortCallerEncoder
	encCfg.EncodeDuration = zapcore.StringDurationEncoder

	// LogLevel 与 zapcore.Level 数值完全对齐，可直接转换
	zapLevel := zapcore.Level(cfg.Level)
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encCfg),
		zapcore.AddSync(os.Stdout),
		zap.NewAtomicLevelAt(zapLevel),
	)

	opts := []zap.Option{
		zap.AddCaller(),
		zap.AddCallerSkip(1), // 跳过 Logger/ContextLogger 包装层，指向真实调用点
	}
	if cfg.ServerName != "" {
		opts = append(opts, zap.Fields(zap.String("appName", cfg.ServerName)))
		if cfg.Version != "" {
			opts = append(opts, zap.Fields(zap.String("version", cfg.Version)))
		}
	}
	if dev {
		opts = append(opts, zap.Development())
	}
	if cfg.DisableStacktrace {
		level := zapcore.DPanicLevel
		if cfg.StacktraceLevel != nil {
			level = zapcore.Level(*cfg.StacktraceLevel)
		}
		opts = append(opts, zap.AddStacktrace(level))
	}
	return zap.New(core, opts...)
}

// ── Logger ────────────────────────────────────────────────────────────────────

// Logger 定义基础日志记录器，底层使用 zap.SugaredLogger 实现 printf 风格调用。
type Logger struct {
	Level LogLevel
	zl    *zap.Logger
	sugar *zap.SugaredLogger
}

// NewLogger 创建日志记录器。
func NewLogger(cfg Config) (*Logger, error) {
	zl := buildZapLogger(cfg)
	return &Logger{
		Level: cfg.Level,
		zl:    zl,
		sugar: zl.Sugar(),
	}, nil
}

// GetZapLogger 返回底层 *zap.Logger，供框架适配器（gorm、gin 等）使用。
func (l *Logger) GetZapLogger() *zap.Logger { return l.zl }

func (l *Logger) Debug(format string, v ...interface{}) { l.sugar.Debugf(format, v...) }
func (l *Logger) Info(format string, v ...interface{})  { l.sugar.Infof(format, v...) }
func (l *Logger) Warn(format string, v ...interface{})  { l.sugar.Warnf(format, v...) }
func (l *Logger) Error(format string, v ...interface{}) { l.sugar.Errorf(format, v...) }
func (l *Logger) Fatal(format string, v ...interface{}) { l.sugar.Fatalf(format, v...) }

// Close 刷新 zap 缓冲并关闭底层资源。
func (l *Logger) Close() error { _ = l.zl.Sync(); return nil }

// WithContext 绑定上下文，返回 ContextLogger。
func (l *Logger) WithContext(ctx context.Context) *ContextLogger {
	return NewContextLogger(ctx, l)
}

// ── ContextLogger ─────────────────────────────────────────────────────────────

// ContextLogger 定义带请求上下文的日志记录器，自动附加 request_id 字段。
type ContextLogger struct {
	*Logger
	ctx context.Context
}

// NewContextLogger 创建带上下文日志记录器。
func NewContextLogger(ctx context.Context, logger *Logger) *ContextLogger {
	return &ContextLogger{Logger: logger, ctx: ctx}
}

// contextSugar 从 ctx 提取 request_id，返回携带该字段的 SugaredLogger。
func (cl *ContextLogger) contextSugar() *zap.SugaredLogger {
	if cl.ctx == nil {
		return cl.sugar
	}
	reqID := ctxx.RequestIDFromContext(cl.ctx)
	if reqID == "" {
		return cl.sugar
	}
	return cl.sugar.With("request_id", reqID)
}

func (cl *ContextLogger) Debug(format string, v ...interface{}) {
	cl.contextSugar().Debugf(format, v...)
}
func (cl *ContextLogger) Info(format string, v ...interface{}) {
	cl.contextSugar().Infof(format, v...)
}
func (cl *ContextLogger) Warn(format string, v ...interface{}) {
	cl.contextSugar().Warnf(format, v...)
}
func (cl *ContextLogger) Error(format string, v ...interface{}) {
	cl.contextSugar().Errorf(format, v...)
}
func (cl *ContextLogger) Fatal(format string, v ...interface{}) {
	cl.contextSugar().Fatalf(format, v...)
}

// ── 全局 Logger 单例 ──────────────────────────────────────────────────────────

var (
	globalMu     sync.RWMutex
	globalLogger *Logger
	// ginZapLogger 专用于 gin 访问日志（不携带调用者信息）
	ginZapLogger *zap.Logger
)

func init() {
	// 保证进程启动即有可用 logger，各服务启动时调用 InitGlobalLogger 覆盖配置
	_ = InitGlobalLogger(DefaultConfig())
}

// InitGlobalLogger 使用指定配置初始化全局日志器，同时更新框架适配器。
func InitGlobalLogger(cfg Config) error {
	l, err := NewLogger(cfg)
	if err != nil {
		return err
	}
	globalMu.Lock()
	globalLogger = l
	ginZapLogger = l.zl.WithOptions(zap.WithCaller(false))
	globalMu.Unlock()
	return nil
}

// InitDefaultLogger 使用服务名和版本号初始化全局日志器（便捷入口）。
func InitDefaultLogger(serverName, version string) {
	cfg := DefaultConfig()
	cfg.ServerName = serverName
	cfg.Version = version
	_ = InitGlobalLogger(cfg)
}

// GetLogger 获取全局日志器；若未初始化则自动使用默认配置。
func GetLogger() *Logger {
	globalMu.RLock()
	l := globalLogger
	globalMu.RUnlock()
	if l != nil {
		return l
	}
	globalMu.Lock()
	defer globalMu.Unlock()
	if globalLogger == nil {
		l, _ := NewLogger(DefaultConfig())
		globalLogger = l
		ginZapLogger = l.zl.WithOptions(zap.WithCaller(false))
	}
	return globalLogger
}

// GetZapLogger 返回全局底层 *zap.Logger，供框架适配器使用。
func GetZapLogger() *zap.Logger {
	return GetLogger().GetZapLogger()
}

// LoggerFromContext 从上下文提取 ContextLogger；若未注入则返回绑定了 ctx 的全局 logger。
func LoggerFromContext(ctx context.Context) *ContextLogger {
	if ctxLogger, ok := ctxx.Get[*ContextLogger](ctx, ctxx.LoggerKey); ok && ctxLogger != nil {
		return ctxLogger
	}
	// 兼容历史字符串键
	if ctx != nil {
		if logger, ok := ctx.Value("logger").(*ContextLogger); ok {
			return logger
		}
	}
	return GetLogger().WithContext(ctx)
}

// WithContextLogger 将 ContextLogger 写入 context，便于在调用链中传递。
func WithContextLogger(ctx context.Context, logger *ContextLogger) context.Context {
	return ctxx.With(ctx, ctxx.LoggerKey, logger)
}

// WithRequestID 将请求 ID 写入 context。
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return ctxx.With(ctx, ctxx.RequestIDKey, requestID)
}

// ── 工具函数 ──────────────────────────────────────────────────────────────────

// LogLevelFromString 将字符串解析为日志级别。
func LogLevelFromString(level string) LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

// StringFromLogLevel 将日志级别转为字符串。
func StringFromLogLevel(level LogLevel) string {
	switch level {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	default:
		return "info"
	}
}

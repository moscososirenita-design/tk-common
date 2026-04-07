package logx

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/moscososirenita-design/tk-common/utils/ctxx"
	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// GormLogger 实现 gorm.io/gorm/logger.Interface，将 GORM 日志接入 zap。
type GormLogger struct {
	logLevel gormlogger.LogLevel
	logger   *zap.Logger
}

// NewGormLogger 创建 GormLogger。level 控制 SQL 日志详细程度：
//
//	gormlogger.Silent — 不记录任何 SQL
//	gormlogger.Error  — 只记录错误
//	gormlogger.Warn   — 记录慢查询 + 错误
//	gormlogger.Info   — 记录所有 SQL（调试用）
func NewGormLogger(level gormlogger.LogLevel) *GormLogger {
	// AddCallerSkip(3)：跳过 GormLogger 方法 + gorm 内部两层包装，指向业务调用点
	return &GormLogger{
		logLevel: level,
		logger:   GetZapLogger().WithOptions(zap.AddCallerSkip(2)),
	}
}

// LogMode 实现 gormlogger.Interface，返回新的 GormLogger 并设置日志级别。
func (g *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return &GormLogger{logLevel: level, logger: g.logger}
}

func (g *GormLogger) Info(ctx context.Context, s string, i ...interface{}) {
	if g.logLevel < gormlogger.Info {
		return
	}
	g.logger.Info(fmt.Sprintf(s, i...), requestIDField(ctx))
}

func (g *GormLogger) Warn(ctx context.Context, s string, i ...interface{}) {
	if g.logLevel < gormlogger.Warn {
		return
	}
	g.logger.Warn(fmt.Sprintf(s, i...), requestIDField(ctx))
}

func (g *GormLogger) Error(ctx context.Context, s string, i ...interface{}) {
	if g.logLevel < gormlogger.Error {
		return
	}
	g.logger.Error(fmt.Sprintf(s, i...), requestIDField(ctx))
}

// Trace 记录 SQL 执行明细：慢查询 → Warn，错误 → Error，正常 → Info。
func (g *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if g.logLevel <= gormlogger.Silent {
		return
	}
	elapsed := time.Since(begin)
	sql, rows := fc()
	reqID := requestIDField(ctx)

	fields := []zap.Field{
		zap.Duration("elapsed", elapsed),
		zap.String("sql", sql),
		zap.Int64("rows", rows),
		reqID,
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 记录未找到：Warn 级别，通常属于正常业务逻辑
			g.logger.Warn("[MySQL] Exec", fields...)
		} else {
			g.logger.Error("[MySQL] Exec", append(fields, zap.Error(err))...)
		}
		return
	}
	if g.logLevel >= gormlogger.Info {
		g.logger.Info("[MySQL] Exec", fields...)
	}
}

// requestIDField 从 context 提取 request_id，返回 zap.Field；无则返回空字段。
func requestIDField(ctx context.Context) zap.Field {
	if ctx == nil {
		return zap.Skip()
	}
	reqID := ctxx.RequestIDFromContext(ctx)
	if reqID == "" {
		return zap.Skip()
	}
	return zap.String("request_id", reqID)
}

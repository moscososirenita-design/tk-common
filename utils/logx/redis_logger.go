package logx

import (
	"context"
	"fmt"
)

// RedisLogger 实现 redis 客户端日志钩子接口（go-redis v8/v9 的 SetLogger）。
// 使用示例：
//
//	rdb.SetLogger(&logx.RedisLogger{})
type RedisLogger struct{}

// Printf 接收 redis 客户端内部日志，转发到全局 logger（Info 级别）。
func (r *RedisLogger) Printf(ctx context.Context, format string, v ...interface{}) {
	LoggerFromContext(ctx).Info(fmt.Sprintf(format, v...))
}

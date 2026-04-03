package ctxx

import "context"

// Key 定义上下文键类型，避免字符串键冲突。
type Key struct {
	// name 保存键名，仅用于调试展示。
	name string
}

// String 返回键名字符串，便于日志排查。
func (k Key) String() string {
	// 返回当前处理结果。
	return "ctxx." + k.name
}

// 声明常用上下文键，服务间共享一致的上下文协议。
var (
	// DBKey 保存数据库连接。
	DBKey = Key{name: "db"}
	// RedisKey 保存 Redis 客户端。
	RedisKey = Key{name: "redis"}
	// LoggerKey 保存上下文日志记录器。
	LoggerKey = Key{name: "logger"}
	// RequestIDKey 保存请求追踪 ID。
	RequestIDKey = Key{name: "request_id"}
)

// With 向上下文写入键值；若 ctx 为空则回退到 Background。
func With(ctx context.Context, key Key, value any) context.Context {
	// 判断条件并进入对应分支逻辑。
	if ctx == nil {
		// 更新当前变量或字段值。
		ctx = context.Background()
	}
	// 返回当前处理结果。
	return context.WithValue(ctx, key, value)
}

// Get 从上下文读取指定类型的值。
func Get[T any](ctx context.Context, key Key) (T, bool) {
	// 声明当前变量。
	var zero T
	// 判断条件并进入对应分支逻辑。
	if ctx == nil {
		// 返回当前处理结果。
		return zero, false
	}
	// 判断条件并进入对应分支逻辑。
	if raw := ctx.Value(key); raw != nil {
		// 判断条件并进入对应分支逻辑。
		if val, ok := raw.(T); ok {
			// 返回当前处理结果。
			return val, true
		}
	}
	// 返回当前处理结果。
	return zero, false
}

// RequestIDFromContext 从上下文读取请求 ID。
func RequestIDFromContext(ctx context.Context) string {
	// 判断条件并进入对应分支逻辑。
	if ctx == nil {
		// 返回当前处理结果。
		return ""
	}
	// 判断条件并进入对应分支逻辑。
	if requestID, ok := Get[string](ctx, RequestIDKey); ok {
		// 返回当前处理结果。
		return requestID
	}
	// 兼容历史字符串键读取，降低迁移风险。
	if requestID, ok := ctx.Value("request_id").(string); ok {
		// 返回当前处理结果。
		return requestID
	}
	// 返回当前处理结果。
	return ""
}

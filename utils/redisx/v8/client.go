// Deprecated: Use github.com/moscososirenita-design/tk-common/utils/redisx/v9 instead.
// This package wraps github.com/go-redis/redis/v8 which is no longer maintained.
// It is kept only for backward compatibility with external consumers.
package v8

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	redis "github.com/go-redis/redis/v8"
)

// Config Redis 客户端配置。
type Config struct {
	// 处理当前语句逻辑。
	Addr string
	// 处理当前语句逻辑。
	Password string
	// 处理当前语句逻辑。
	DB int
	// 处理当前语句逻辑。
	PoolSize int
	// 处理当前语句逻辑。
	MinIdleConns int
	// 处理当前语句逻辑。
	DialTimeout time.Duration
	// 处理当前语句逻辑。
	ReadTimeout time.Duration
	// 处理当前语句逻辑。
	WriteTimeout time.Duration
}

// DefaultConfig 返回默认 Redis 配置。
func DefaultConfig() Config {
	// 返回当前处理结果。
	return Config{
		// 处理当前语句逻辑。
		PoolSize: 10,
		// 处理当前语句逻辑。
		MinIdleConns: 5,
		// 处理当前语句逻辑。
		DialTimeout: 5 * time.Second,
		// 处理当前语句逻辑。
		ReadTimeout: 3 * time.Second,
		// 处理当前语句逻辑。
		WriteTimeout: 3 * time.Second,
	}
}

// NewClient 创建 Redis v8 客户端并执行连通性校验。
func NewClient(ctx context.Context, cfg Config) (*redis.Client, error) {
	// 判断条件并进入对应分支逻辑。
	if cfg.Addr == "" {
		// 返回当前处理结果。
		return nil, fmt.Errorf("redis address is empty")
	}
	// 定义并初始化当前变量。
	client := redis.NewClient(&redis.Options{
		// 处理当前语句逻辑。
		Addr: cfg.Addr,
		// 处理当前语句逻辑。
		Password: cfg.Password,
		// 处理当前语句逻辑。
		DB: cfg.DB,
		// 处理当前语句逻辑。
		PoolSize: cfg.PoolSize,
		// 处理当前语句逻辑。
		MinIdleConns: cfg.MinIdleConns,
		// 处理当前语句逻辑。
		DialTimeout: cfg.DialTimeout,
		// 处理当前语句逻辑。
		ReadTimeout: cfg.ReadTimeout,
		// 处理当前语句逻辑。
		WriteTimeout: cfg.WriteTimeout,
	})
	// 定义并初始化当前变量。
	pingCtx := ctx
	// 判断条件并进入对应分支逻辑。
	if pingCtx == nil {
		// 定义并初始化当前变量。
		tmp, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		// 注册延迟执行逻辑。
		defer cancel()
		// 更新当前变量或字段值。
		pingCtx = tmp
	}
	// 判断条件并进入对应分支逻辑。
	if err := client.Ping(pingCtx).Err(); err != nil {
		// 返回当前处理结果。
		return nil, fmt.Errorf("failed to connect redis: %w", err)
	}
	// 返回当前处理结果。
	return client, nil
}

// SetString 写字符串缓存。
func SetString(ctx context.Context, cli *redis.Client, key string, val string, ttl time.Duration) error {
	// 判断条件并进入对应分支逻辑。
	if cli == nil {
		// 返回当前处理结果。
		return nil
	}
	// 返回当前处理结果。
	return cli.Set(ctx, key, val, ttl).Err()
}

// GetString 读取字符串缓存。
func GetString(ctx context.Context, cli *redis.Client, key string) (string, bool, error) {
	// 判断条件并进入对应分支逻辑。
	if cli == nil {
		// 返回当前处理结果。
		return "", false, nil
	}
	// 定义并初始化当前变量。
	raw, err := cli.Get(ctx, key).Result()
	// 判断条件并进入对应分支逻辑。
	if errors.Is(err, redis.Nil) {
		// 返回当前处理结果。
		return "", false, nil
	}
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return "", false, err
	}
	// 返回当前处理结果。
	return raw, true, nil
}

// Del 删除缓存键。
func Del(ctx context.Context, cli *redis.Client, keys ...string) error {
	// 判断条件并进入对应分支逻辑。
	if cli == nil || len(keys) == 0 {
		// 返回当前处理结果。
		return nil
	}
	// 返回当前处理结果。
	return cli.Del(ctx, keys...).Err()
}

// IncrWithExpire 自增计数并在首次创建时设置过期时间。
func IncrWithExpire(ctx context.Context, cli *redis.Client, key string, ttl time.Duration) (int64, error) {
	// 判断条件并进入对应分支逻辑。
	if cli == nil {
		// 返回当前处理结果。
		return 0, nil
	}
	// 定义并初始化当前变量。
	n, err := cli.Incr(ctx, key).Result()
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return 0, err
	}
	// 判断条件并进入对应分支逻辑。
	if n == 1 && ttl > 0 {
		// 更新当前变量或字段值。
		_ = cli.Expire(ctx, key, ttl).Err()
	}
	// 返回当前处理结果。
	return n, nil
}

// SetJSON 序列化并写入 JSON 缓存。
func SetJSON(ctx context.Context, cli *redis.Client, key string, val any, ttl time.Duration) error {
	// 判断条件并进入对应分支逻辑。
	if cli == nil {
		// 返回当前处理结果。
		return nil
	}
	// 定义并初始化当前变量。
	raw, err := json.Marshal(val)
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return err
	}
	// 返回当前处理结果。
	return cli.Set(ctx, key, raw, ttl).Err()
}

// GetJSON 读取并反序列化 JSON 缓存。
func GetJSON(ctx context.Context, cli *redis.Client, key string, out any) (bool, error) {
	// 判断条件并进入对应分支逻辑。
	if cli == nil {
		// 返回当前处理结果。
		return false, nil
	}
	// 定义并初始化当前变量。
	raw, err := cli.Get(ctx, key).Bytes()
	// 判断条件并进入对应分支逻辑。
	if errors.Is(err, redis.Nil) {
		// 返回当前处理结果。
		return false, nil
	}
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return false, err
	}
	// 判断条件并进入对应分支逻辑。
	if len(raw) == 0 {
		// 返回当前处理结果。
		return false, nil
	}
	// 判断条件并进入对应分支逻辑。
	if err := json.Unmarshal(raw, out); err != nil {
		// 返回当前处理结果。
		return false, err
	}
	// 返回当前处理结果。
	return true, nil
}

// RedisFromContext 从上下文按 key 提取 Redis 客户端。
func RedisFromContext(ctx context.Context, key any) *redis.Client {
	// 判断条件并进入对应分支逻辑。
	if ctx == nil {
		// 返回当前处理结果。
		return nil
	}
	// 判断条件并进入对应分支逻辑。
	if cli, ok := ctx.Value(key).(*redis.Client); ok {
		// 返回当前处理结果。
		return cli
	}
	// 返回当前处理结果。
	return nil
}

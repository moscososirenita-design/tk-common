// Package cmdx 提供统一的 Redis 操作指令集。
//
// 设计原则：
//   - 所有操作均接收 context.Context，支持超时传递与链路追踪；
//   - 每个操作均集成日志记录，错误时自动打印 key / err / 耗时；
//   - client 为 nil 时不 panic，直接返回"不可用"语义结果；
//   - key 前缀约定：{service}:{domain}:{id}，由调用方负责构造；
//   - 所有写操作必须携带 TTL，禁止裸 Set 不设过期时间（0 = 永久需明确传入）。
//
// 文件职责说明（按文件拆分，方便按需引入）：
//
//	base.go      — String / TTL / Del / Exists / Incr 等基础指令
//	json.go      — SetJSON / GetJSON / MGetJSON 等 JSON 序列化指令
//	lock.go      — SetNX 分布式锁 / 续约 / 释放（Lua 原子删除）
//	rwlock.go    — 读写锁（多读单写，基于 Redis Lua 脚本模拟）
//	rate.go      — 滑动窗口计数限流 / 令牌桶辅助操作
//	pipeline.go  — Pipeline 批量写入 / MGet 批量读取封装
package cmdx

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/moscososirenita-design/tk-common/utils/logx"
	"github.com/redis/go-redis/v9"
)

// ErrNilClient 表示 Redis 客户端未初始化。
var ErrNilClient = errors.New("redis: client is nil")

// ErrKeyNotFound 表示 Redis key 不存在（与 redis.Nil 对齐）。
var ErrKeyNotFound = errors.New("redis: key not found")

// ─────────────────────────────────────────────
// Set — 写字符串（必须设置 TTL，ttl=0 表示永久）
// ─────────────────────────────────────────────

// Set 将字符串值写入 Redis，ttl 为 0 时表示不过期（永久存储）。
// 生产环境中大多数 key 应设置合理 TTL；
// 仅开奖结果等"写一次永久有效"的场景才允许 ttl=0。
func Set(ctx context.Context, cli *redis.Client, key, val string, ttl time.Duration) error {
	if cli == nil {
		// 客户端未初始化时记录警告，业务层可决定是否降级
		logx.LoggerFromContext(ctx).Warn("cmdx.Set: nil client, key=%s", key)
		return ErrNilClient
	}

	start := time.Now()
	err := cli.Set(ctx, key, val, ttl).Err()
	elapsed := time.Since(start)

	if err != nil {
		logx.LoggerFromContext(ctx).Error("cmdx.Set: key=%s ttl=%s elapsed=%s err=%v",
			key, ttl, elapsed, err)
		return fmt.Errorf("cmdx.Set %q: %w", key, err)
	}

	logx.LoggerFromContext(ctx).Debug("cmdx.Set: key=%s ttl=%s elapsed=%s ok", key, ttl, elapsed)
	return nil
}

// SetEX 写字符串并强制要求 ttl > 0（语义更明确，防止意外永久存储）。
// 若 ttl <= 0 直接返回错误，强制调用方明确传入有效期。
func SetEX(ctx context.Context, cli *redis.Client, key, val string, ttl time.Duration) error {
	if ttl <= 0 {
		return fmt.Errorf("cmdx.SetEX %q: ttl must be > 0", key)
	}
	return Set(ctx, cli, key, val, ttl)
}

// ─────────────────────────────────────────────
// Get — 读字符串
// ─────────────────────────────────────────────

// Get 读取字符串值。
// 返回 (value, true, nil) 表示命中；
// 返回 ("", false, nil) 表示 key 不存在；
// 返回 ("", false, err) 表示 Redis 报错。
func Get(ctx context.Context, cli *redis.Client, key string) (string, bool, error) {
	if cli == nil {
		logx.LoggerFromContext(ctx).Warn("cmdx.Get: nil client, key=%s", key)
		return "", false, ErrNilClient
	}

	start := time.Now()
	val, err := cli.Get(ctx, key).Result()
	elapsed := time.Since(start)

	if errors.Is(err, redis.Nil) {
		// key 不存在是正常业务语义，不记录 error
		logx.LoggerFromContext(ctx).Debug("cmdx.Get: key=%s miss elapsed=%s", key, elapsed)
		return "", false, nil
	}
	if err != nil {
		logx.LoggerFromContext(ctx).Error("cmdx.Get: key=%s elapsed=%s err=%v", key, elapsed, err)
		return "", false, fmt.Errorf("cmdx.Get %q: %w", key, err)
	}

	logx.LoggerFromContext(ctx).Debug("cmdx.Get: key=%s hit elapsed=%s", key, elapsed)
	return val, true, nil
}

// MGet 批量读取多个 key 的字符串值（保持与 key 列表等长的结果切片）。
// 结果切片中 nil 表示对应 key 不存在。
func MGet(ctx context.Context, cli *redis.Client, keys ...string) ([]interface{}, error) {
	if cli == nil {
		return nil, ErrNilClient
	}
	if len(keys) == 0 {
		return nil, nil
	}

	start := time.Now()
	vals, err := cli.MGet(ctx, keys...).Result()
	elapsed := time.Since(start)

	if err != nil {
		logx.LoggerFromContext(ctx).Error("cmdx.MGet: keys=%v elapsed=%s err=%v", keys, elapsed, err)
		return nil, fmt.Errorf("cmdx.MGet: %w", err)
	}

	logx.LoggerFromContext(ctx).Debug("cmdx.MGet: count=%d elapsed=%s", len(keys), elapsed)
	return vals, nil
}

// ─────────────────────────────────────────────
// Del — 删除 key
// ─────────────────────────────────────────────

// Del 删除一个或多个 key，返回实际删除的数量。
// 若 key 不存在不报错，返回 0。
func Del(ctx context.Context, cli *redis.Client, keys ...string) (int64, error) {
	if cli == nil || len(keys) == 0 {
		return 0, nil
	}

	start := time.Now()
	n, err := cli.Del(ctx, keys...).Result()
	elapsed := time.Since(start)

	if err != nil {
		logx.LoggerFromContext(ctx).Error("cmdx.Del: keys=%v elapsed=%s err=%v", keys, elapsed, err)
		return 0, fmt.Errorf("cmdx.Del: %w", err)
	}

	logx.LoggerFromContext(ctx).Debug("cmdx.Del: keys=%v deleted=%d elapsed=%s", keys, n, elapsed)
	return n, nil
}

// ─────────────────────────────────────────────
// Exists — 检查 key 是否存在
// ─────────────────────────────────────────────

// Exists 检查 key 是否存在。
// 返回 true 表示存在，false 表示不存在或客户端 nil。
func Exists(ctx context.Context, cli *redis.Client, key string) (bool, error) {
	if cli == nil {
		return false, ErrNilClient
	}

	start := time.Now()
	n, err := cli.Exists(ctx, key).Result()
	elapsed := time.Since(start)

	if err != nil {
		logx.LoggerFromContext(ctx).Error("cmdx.Exists: key=%s elapsed=%s err=%v", key, elapsed, err)
		return false, fmt.Errorf("cmdx.Exists %q: %w", key, err)
	}

	logx.LoggerFromContext(ctx).Debug("cmdx.Exists: key=%s exists=%v elapsed=%s", key, n > 0, elapsed)
	return n > 0, nil
}

// ─────────────────────────────────────────────
// TTL — 查询过期时间 / 续约
// ─────────────────────────────────────────────

// TTL 查询 key 的剩余 TTL。
// 返回 -1 表示 key 存在但未设置过期；
// 返回 -2 表示 key 不存在；
// 返回 > 0 为剩余有效时间。
func TTL(ctx context.Context, cli *redis.Client, key string) (time.Duration, error) {
	if cli == nil {
		return 0, ErrNilClient
	}

	start := time.Now()
	ttl, err := cli.TTL(ctx, key).Result()
	elapsed := time.Since(start)

	if err != nil {
		logx.LoggerFromContext(ctx).Error("cmdx.TTL: key=%s elapsed=%s err=%v", key, elapsed, err)
		return 0, fmt.Errorf("cmdx.TTL %q: %w", key, err)
	}

	logx.LoggerFromContext(ctx).Debug("cmdx.TTL: key=%s ttl=%s elapsed=%s", key, ttl, elapsed)
	return ttl, nil
}

// Expire 为已存在的 key 设置或更新 TTL（续约常用）。
// 返回 true 表示设置成功，false 表示 key 不存在。
func Expire(ctx context.Context, cli *redis.Client, key string, ttl time.Duration) (bool, error) {
	if cli == nil {
		return false, ErrNilClient
	}
	if ttl <= 0 {
		return false, fmt.Errorf("cmdx.Expire %q: ttl must be > 0", key)
	}

	start := time.Now()
	ok, err := cli.Expire(ctx, key, ttl).Result()
	elapsed := time.Since(start)

	if err != nil {
		logx.LoggerFromContext(ctx).Error("cmdx.Expire: key=%s ttl=%s elapsed=%s err=%v",
			key, ttl, elapsed, err)
		return false, fmt.Errorf("cmdx.Expire %q: %w", key, err)
	}

	logx.LoggerFromContext(ctx).Debug("cmdx.Expire: key=%s ttl=%s ok=%v elapsed=%s",
		key, ttl, ok, elapsed)
	return ok, nil
}

// ─────────────────────────────────────────────
// Incr / Decr — 计数器操作
// ─────────────────────────────────────────────

// Incr 对 key 进行自增，首次创建时自动设置 TTL（原子操作保证计数精度）。
// 使用场景：短信发送次数统计、登录失败计数、限流计数器。
// 注意：INCR 和 EXPIRE 不是原子操作，极端情况下 EXPIRE 可能失败；
//
//	若对计数精度要求极高，应改用 Lua 脚本（见 IncrWithTTLLua）。
func Incr(ctx context.Context, cli *redis.Client, key string, ttl time.Duration) (int64, error) {
	if cli == nil {
		return 0, ErrNilClient
	}

	start := time.Now()
	n, err := cli.Incr(ctx, key).Result()
	elapsed := time.Since(start)

	if err != nil {
		logx.LoggerFromContext(ctx).Error("cmdx.Incr: key=%s elapsed=%s err=%v", key, elapsed, err)
		return 0, fmt.Errorf("cmdx.Incr %q: %w", key, err)
	}

	// 首次创建时设置过期，防止计数 key 永久存在
	if n == 1 && ttl > 0 {
		if expErr := cli.Expire(ctx, key, ttl).Err(); expErr != nil {
			// Expire 失败记录 warn，不影响主流程计数结果
			logx.LoggerFromContext(ctx).Warn("cmdx.Incr: Expire key=%s err=%v", key, expErr)
		}
	}

	logx.LoggerFromContext(ctx).Debug("cmdx.Incr: key=%s count=%d elapsed=%s", key, n, elapsed)
	return n, nil
}

// IncrWithTTLLua 使用 Lua 脚本原子地执行 INCR + 首次 EXPIRE，
// 避免极端并发下 INCR 和 EXPIRE 之间的 key 被删除导致 TTL 未设置的问题。
//
// Lua 脚本逻辑：
//  1. INCR key  → 得到新值 n
//  2. 若 n == 1（首次创建），则 EXPIRE key ttl
//  3. 返回 n
func IncrWithTTLLua(ctx context.Context, cli *redis.Client, key string, ttl time.Duration) (int64, error) {
	if cli == nil {
		return 0, ErrNilClient
	}

	// Lua 脚本：原子 INCR + 首次 EXPIRE
	const script = `
local n = redis.call('INCR', KEYS[1])
if n == 1 then
  redis.call('EXPIRE', KEYS[1], ARGV[1])
end
return n`

	start := time.Now()
	result, err := cli.Eval(ctx, script, []string{key}, int64(ttl.Seconds())).Result()
	elapsed := time.Since(start)

	if err != nil {
		logx.LoggerFromContext(ctx).Error("cmdx.IncrWithTTLLua: key=%s elapsed=%s err=%v",
			key, elapsed, err)
		return 0, fmt.Errorf("cmdx.IncrWithTTLLua %q: %w", key, err)
	}

	n, ok := result.(int64)
	if !ok {
		return 0, fmt.Errorf("cmdx.IncrWithTTLLua %q: unexpected result type %T", key, result)
	}

	logx.LoggerFromContext(ctx).Debug("cmdx.IncrWithTTLLua: key=%s count=%d elapsed=%s",
		key, n, elapsed)
	return n, nil
}

// Decr 对 key 进行自减，返回自减后的值。
func Decr(ctx context.Context, cli *redis.Client, key string) (int64, error) {
	if cli == nil {
		return 0, ErrNilClient
	}

	start := time.Now()
	n, err := cli.Decr(ctx, key).Result()
	elapsed := time.Since(start)

	if err != nil {
		logx.LoggerFromContext(ctx).Error("cmdx.Decr: key=%s elapsed=%s err=%v", key, elapsed, err)
		return 0, fmt.Errorf("cmdx.Decr %q: %w", key, err)
	}

	logx.LoggerFromContext(ctx).Debug("cmdx.Decr: key=%s count=%d elapsed=%s", key, n, elapsed)
	return n, nil
}

// ─────────────────────────────────────────────
// SetNX — 仅当 key 不存在时写入（幂等写入 / 分布式锁基础原语）
// ─────────────────────────────────────────────

// SetNX 仅当 key 不存在时写入 val，成功返回 true。
// 常用于：幂等请求去重、分布式锁的 try-acquire。
// 注意：若需要完整的锁语义（含 token 续约和 Lua 原子释放），请使用 lock.go 中的函数。
func SetNX(ctx context.Context, cli *redis.Client, key, val string, ttl time.Duration) (bool, error) {
	if cli == nil {
		return false, ErrNilClient
	}
	if ttl <= 0 {
		return false, fmt.Errorf("cmdx.SetNX %q: ttl must be > 0", key)
	}

	start := time.Now()
	ok, err := cli.SetNX(ctx, key, val, ttl).Result()
	elapsed := time.Since(start)

	if err != nil {
		logx.LoggerFromContext(ctx).Error("cmdx.SetNX: key=%s elapsed=%s err=%v", key, elapsed, err)
		return false, fmt.Errorf("cmdx.SetNX %q: %w", key, err)
	}

	logx.LoggerFromContext(ctx).Debug("cmdx.SetNX: key=%s acquired=%v elapsed=%s", key, ok, elapsed)
	return ok, nil
}

// ─────────────────────────────────────────────
// GetSet — 原子读取旧值并写入新值
// ─────────────────────────────────────────────

// GetSet 原子地将 key 设置为 newVal 并返回旧值（key 不存在时返回 ErrKeyNotFound）。
// 常用于：Token 轮换、状态机原子转换。
func GetSet(ctx context.Context, cli *redis.Client, key, newVal string) (string, error) {
	if cli == nil {
		return "", ErrNilClient
	}

	start := time.Now()
	old, err := cli.GetSet(ctx, key, newVal).Result()
	elapsed := time.Since(start)

	if errors.Is(err, redis.Nil) {
		logx.LoggerFromContext(ctx).Debug("cmdx.GetSet: key=%s no old value elapsed=%s", key, elapsed)
		return "", ErrKeyNotFound
	}
	if err != nil {
		logx.LoggerFromContext(ctx).Error("cmdx.GetSet: key=%s elapsed=%s err=%v", key, elapsed, err)
		return "", fmt.Errorf("cmdx.GetSet %q: %w", key, err)
	}

	logx.LoggerFromContext(ctx).Debug("cmdx.GetSet: key=%s elapsed=%s ok", key, elapsed)
	return old, nil
}

// ─────────────────────────────────────────────
// Ping — 连通性探测
// ─────────────────────────────────────────────

// Ping 探测 Redis 连通性，生产健康检查接口使用。
func Ping(ctx context.Context, cli *redis.Client) error {
	if cli == nil {
		return ErrNilClient
	}
	return cli.Ping(ctx).Err()
}

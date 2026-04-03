package cmdx

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/moscososirenita-design/tk-common/utils/logx"
	"github.com/redis/go-redis/v9"

	"github.com/google/uuid"
)

// ─────────────────────────────────────────────
// 分布式互斥锁
//
// 实现方式：SET key token NX EX ttl
// 释放方式：Lua 脚本原子比较 + 删除（防误删其他持有者的锁）
//
// 使用模式（推荐）：
//
//	lock, err := cmdx.AcquireLock(ctx, cli, "lock:order:123", 10*time.Second)
//	if err != nil { /* 获取锁失败，另一个实例正在处理 */ }
//	defer lock.Release(ctx, cli)
//	// ... 业务逻辑 ...
//	lock.Renew(ctx, cli, 10*time.Second)   // 长任务中途续约
// ─────────────────────────────────────────────

// ErrLockNotAcquired 表示锁被其他持有者占用，当前获取失败。
var ErrLockNotAcquired = errors.New("redis lock: not acquired")

// ErrLockExpired 表示锁在操作期间已超时自动释放（TTL 耗尽）。
var ErrLockExpired = errors.New("redis lock: expired before release")

// Lock 代表一把已持有的分布式锁，通过 token 标识所有权。
type Lock struct {
	// key 是锁在 Redis 中的存储键
	key string
	// token 是随机生成的唯一标识，用于 Lua 脚本比对防止误删
	token string
	// ttl 是锁的初始 TTL，续约时复用
	ttl time.Duration
}

// Key 返回锁的 Redis key（便于日志打印）。
func (l *Lock) Key() string { return l.key }

// AcquireLock 尝试获取分布式互斥锁。
//
// 参数：
//   - key:  锁键，建议格式 lock:{domain}:{id}，如 lock:draw:123
//   - ttl:  锁的最大持有时间，超时自动释放（防止死锁）
//
// 返回：
//   - *Lock:               成功持有时返回锁句柄
//   - ErrLockNotAcquired:  锁已被占用（调用方可选择重试或放弃）
//   - 其他 err:             Redis 通信错误
func AcquireLock(ctx context.Context, cli *redis.Client, key string, ttl time.Duration) (*Lock, error) {
	if cli == nil {
		return nil, ErrNilClient
	}
	if ttl <= 0 {
		return nil, fmt.Errorf("cmdx.AcquireLock %q: ttl must be > 0", key)
	}

	// 生成唯一 token 标识当前锁持有者
	token := uuid.New().String()

	start := time.Now()
	ok, err := cli.SetNX(ctx, key, token, ttl).Result()
	elapsed := time.Since(start)

	if err != nil {
		logx.LoggerFromContext(ctx).Error("cmdx.AcquireLock: key=%s elapsed=%s err=%v",
			key, elapsed, err)
		return nil, fmt.Errorf("cmdx.AcquireLock %q: %w", key, err)
	}

	if !ok {
		// 锁已被其他实例持有
		logx.LoggerFromContext(ctx).Debug("cmdx.AcquireLock: key=%s already locked elapsed=%s",
			key, elapsed)
		return nil, ErrLockNotAcquired
	}

	logx.LoggerFromContext(ctx).Debug("cmdx.AcquireLock: key=%s ttl=%s acquired elapsed=%s",
		key, ttl, elapsed)
	return &Lock{key: key, token: token, ttl: ttl}, nil
}

// AcquireLockWithRetry 带重试的锁获取，适合可以稍等片刻的场景（如并发注册）。
//
// 参数：
//   - maxRetry:   最多重试次数
//   - retryDelay: 每次重试前等待时长
func AcquireLockWithRetry(
	ctx context.Context,
	cli *redis.Client,
	key string,
	ttl time.Duration,
	maxRetry int,
	retryDelay time.Duration,
) (*Lock, error) {
	for i := 0; i <= maxRetry; i++ {
		lock, err := AcquireLock(ctx, cli, key, ttl)
		if err == nil {
			// 获取成功
			return lock, nil
		}
		if !errors.Is(err, ErrLockNotAcquired) {
			// Redis 通信错误，不再重试
			return nil, err
		}
		if i < maxRetry {
			// 等待后重试（检查 ctx 是否已取消）
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(retryDelay):
			}
		}
	}
	logx.LoggerFromContext(ctx).Warn("cmdx.AcquireLockWithRetry: key=%s maxRetry=%d exhausted", key, maxRetry)
	return nil, ErrLockNotAcquired
}

// Release 使用 Lua 脚本原子释放锁（只释放自己持有的锁，防止误删）。
//
// Lua 脚本逻辑：
//  1. GET key → 若不等于自己的 token，直接返回 0（已过期或被他人持有）
//  2. DEL key → 删除锁，返回 1
//
// 若返回 ErrLockExpired，说明锁在业务执行期间 TTL 耗尽（需要在业务层中检查数据一致性）。
func (l *Lock) Release(ctx context.Context, cli *redis.Client) error {
	if cli == nil {
		return ErrNilClient
	}

	// Lua 原子比较 + 删除，防止释放他人持有的锁
	const script = `
if redis.call('GET', KEYS[1]) == ARGV[1] then
  return redis.call('DEL', KEYS[1])
else
  return 0
end`

	start := time.Now()
	result, err := cli.Eval(ctx, script, []string{l.key}, l.token).Result()
	elapsed := time.Since(start)

	if err != nil && !errors.Is(err, redis.Nil) {
		logx.LoggerFromContext(ctx).Error("cmdx.Lock.Release: key=%s elapsed=%s err=%v",
			l.key, elapsed, err)
		return fmt.Errorf("cmdx.Lock.Release %q: %w", l.key, err)
	}

	n, _ := result.(int64)
	if n == 0 {
		// 锁不存在或 token 不匹配（已过期被自动清理）
		logx.LoggerFromContext(ctx).Warn("cmdx.Lock.Release: key=%s token mismatch or expired elapsed=%s",
			l.key, elapsed)
		return ErrLockExpired
	}

	logx.LoggerFromContext(ctx).Debug("cmdx.Lock.Release: key=%s elapsed=%s released", l.key, elapsed)
	return nil
}

// Renew 为锁续约，重置 TTL（适用于执行时间较长的任务）。
// 仅当锁仍由自己持有（token 匹配）时续约才会生效。
// 建议在长任务中每隔 ttl/2 调用一次 Renew。
func (l *Lock) Renew(ctx context.Context, cli *redis.Client, ttl time.Duration) (bool, error) {
	if cli == nil {
		return false, ErrNilClient
	}
	if ttl <= 0 {
		return false, fmt.Errorf("cmdx.Lock.Renew %q: ttl must be > 0", l.key)
	}

	// Lua 脚本：比对 token + 更新 TTL，原子执行
	const script = `
if redis.call('GET', KEYS[1]) == ARGV[1] then
  return redis.call('EXPIRE', KEYS[1], ARGV[2])
else
  return 0
end`

	start := time.Now()
	result, err := cli.Eval(ctx, script, []string{l.key}, l.token, int64(ttl.Seconds())).Result()
	elapsed := time.Since(start)

	if err != nil {
		logx.LoggerFromContext(ctx).Error("cmdx.Lock.Renew: key=%s elapsed=%s err=%v",
			l.key, elapsed, err)
		return false, fmt.Errorf("cmdx.Lock.Renew %q: %w", l.key, err)
	}

	n, _ := result.(int64)
	ok := n == 1
	if !ok {
		logx.LoggerFromContext(ctx).Warn("cmdx.Lock.Renew: key=%s token mismatch or expired elapsed=%s",
			l.key, elapsed)
	} else {
		// 更新内部记录的 TTL，供下次续约参考
		l.ttl = ttl
		logx.LoggerFromContext(ctx).Debug("cmdx.Lock.Renew: key=%s ttl=%s elapsed=%s ok",
			l.key, ttl, elapsed)
	}
	return ok, nil
}

// ─────────────────────────────────────────────
// 辅助：TryLockFunc — 获取锁 + 执行 + 自动释放（便捷封装）
// ─────────────────────────────────────────────

// TryLockFunc 尝试获取锁后执行 fn，执行结束后自动释放。
// 若获取锁失败（ErrLockNotAcquired），直接返回该错误，不执行 fn。
//
// 示例：
//
//	err := cmdx.TryLockFunc(ctx, cli, "lock:draw:123", 5*time.Second, func() error {
//	    return doDrawRecord()
//	})
func TryLockFunc(ctx context.Context, cli *redis.Client, key string, ttl time.Duration, fn func() error) error {
	lock, err := AcquireLock(ctx, cli, key, ttl)
	if err != nil {
		return err
	}
	defer func() {
		if releaseErr := lock.Release(ctx, cli); releaseErr != nil && !errors.Is(releaseErr, ErrLockExpired) {
			logx.LoggerFromContext(ctx).Warn("cmdx.TryLockFunc: release key=%s err=%v", key, releaseErr)
		}
	}()
	return fn()
}

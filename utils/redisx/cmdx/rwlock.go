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
// 读写锁（RWLock）
//
// 实现方式：基于 Redis Hash + Lua 脚本模拟读写锁语义
//   - 读锁 (RLock)：多个读锁可以并行持有；有写锁时不允许读
//   - 写锁 (WLock)：独占，有读锁或其他写锁时不允许获取
//
// Key 结构（以 rwlock:<name> 为例）：
//   rwlock:<name>:writers  — 写者计数（INCR/DECR）
//   rwlock:<name>:readers  — 读者计数（INCR/DECR）
//   rwlock:<name>:wtoken   — 当前写者 token（写锁标识）
//
// 适用场景：
//   - 业务配置的并发读（多人同时读 Banner、广播配置）
//   - 配置写入时独占（后台管理保存配置）
//   - 开奖数据写入时独占，读取时并发（读远多于写）
//
// 注意：Redis 读写锁实现复杂且有一定局限性，
//       若业务场景简单（仅需互斥），优先使用 lock.go 中的 AcquireLock。
// ─────────────────────────────────────────────

// ErrRLockConflict 表示当前有写锁，读锁获取失败。
var ErrRLockConflict = errors.New("redis rwlock: write lock held, read lock refused")

// ErrWLockConflict 表示当前有读锁或写锁，写锁获取失败。
var ErrWLockConflict = errors.New("redis rwlock: lock conflict, write lock refused")

// RLock 代表一把已持有的读锁。
type RLock struct {
	// base 是锁的命名空间前缀（不含 :readers 后缀）
	base  string
	token string
}

// WLock 代表一把已持有的写锁。
type WLock struct {
	base  string
	token string
	ttl   time.Duration
}

// rwKey 构造读写锁各分量 key。
func rwReadersKey(base string) string { return base + ":readers" }
func rwWritersKey(base string) string { return base + ":writers" }
func rwWTokenKey(base string) string  { return base + ":wtoken" }

// AcquireRLock 尝试获取读锁。
// 若当前存在写锁（writers > 0），返回 ErrRLockConflict。
//
// 参数：
//   - base: 锁命名空间，如 rwlock:biz:banner
//   - ttl:  读锁最大持有时间（防止持有者崩溃导致死锁）
func AcquireRLock(ctx context.Context, cli *redis.Client, base string, ttl time.Duration) (*RLock, error) {
	if cli == nil {
		return nil, ErrNilClient
	}
	if ttl <= 0 {
		return nil, fmt.Errorf("cmdx.AcquireRLock %q: ttl must be > 0", base)
	}

	// Lua 脚本：原子检查写者计数 + 递增读者计数
	const script = `
local writers = tonumber(redis.call('GET', KEYS[1]) or '0')
if writers > 0 then
  return 0
end
local n = redis.call('INCR', KEYS[2])
if n == 1 then
  redis.call('EXPIRE', KEYS[2], ARGV[1])
end
return n`

	start := time.Now()
	result, err := cli.Eval(ctx, script,
		[]string{rwWritersKey(base), rwReadersKey(base)},
		int64(ttl.Seconds()),
	).Result()
	elapsed := time.Since(start)

	if err != nil {
		logx.LoggerFromContext(ctx).Error("cmdx.AcquireRLock: base=%s elapsed=%s err=%v",
			base, elapsed, err)
		return nil, fmt.Errorf("cmdx.AcquireRLock %q: %w", base, err)
	}

	n, _ := result.(int64)
	if n == 0 {
		logx.LoggerFromContext(ctx).Debug("cmdx.AcquireRLock: base=%s write lock exists, refused elapsed=%s",
			base, elapsed)
		return nil, ErrRLockConflict
	}

	token := uuid.New().String()
	logx.LoggerFromContext(ctx).Debug("cmdx.AcquireRLock: base=%s readers=%d acquired elapsed=%s",
		base, n, elapsed)
	return &RLock{base: base, token: token}, nil
}

// Release 释放读锁（递减读者计数）。
func (l *RLock) Release(ctx context.Context, cli *redis.Client) error {
	if cli == nil {
		return ErrNilClient
	}

	// 读者计数 -1，最小不低于 0
	const script = `
local n = tonumber(redis.call('GET', KEYS[1]) or '0')
if n > 0 then
  n = redis.call('DECR', KEYS[1])
end
return n`

	start := time.Now()
	_, err := cli.Eval(ctx, script, []string{rwReadersKey(l.base)}).Result()
	elapsed := time.Since(start)

	if err != nil && !errors.Is(err, redis.Nil) {
		logx.LoggerFromContext(ctx).Error("cmdx.RLock.Release: base=%s elapsed=%s err=%v",
			l.base, elapsed, err)
		return fmt.Errorf("cmdx.RLock.Release %q: %w", l.base, err)
	}

	logx.LoggerFromContext(ctx).Debug("cmdx.RLock.Release: base=%s elapsed=%s ok", l.base, elapsed)
	return nil
}

// AcquireWLock 尝试获取写锁。
// 若当前存在读者或其他写者，返回 ErrWLockConflict。
//
// 参数：
//   - base: 锁命名空间，如 rwlock:biz:banner
//   - ttl:  写锁最大持有时间
func AcquireWLock(ctx context.Context, cli *redis.Client, base string, ttl time.Duration) (*WLock, error) {
	if cli == nil {
		return nil, ErrNilClient
	}
	if ttl <= 0 {
		return nil, fmt.Errorf("cmdx.AcquireWLock %q: ttl must be > 0", base)
	}

	token := uuid.New().String()

	// Lua 脚本：原子检查读者计数 + 写者计数 + 设置写锁
	const script = `
local readers = tonumber(redis.call('GET', KEYS[1]) or '0')
local writers = tonumber(redis.call('GET', KEYS[2]) or '0')
if readers > 0 or writers > 0 then
  return 0
end
redis.call('SET', KEYS[2], '1', 'EX', ARGV[1])
redis.call('SET', KEYS[3], ARGV[2], 'EX', ARGV[1])
return 1`

	start := time.Now()
	result, err := cli.Eval(ctx, script,
		[]string{rwReadersKey(base), rwWritersKey(base), rwWTokenKey(base)},
		int64(ttl.Seconds()), token,
	).Result()
	elapsed := time.Since(start)

	if err != nil {
		logx.LoggerFromContext(ctx).Error("cmdx.AcquireWLock: base=%s elapsed=%s err=%v",
			base, elapsed, err)
		return nil, fmt.Errorf("cmdx.AcquireWLock %q: %w", base, err)
	}

	n, _ := result.(int64)
	if n == 0 {
		logx.LoggerFromContext(ctx).Debug("cmdx.AcquireWLock: base=%s conflict elapsed=%s",
			base, elapsed)
		return nil, ErrWLockConflict
	}

	logx.LoggerFromContext(ctx).Debug("cmdx.AcquireWLock: base=%s ttl=%s acquired elapsed=%s",
		base, ttl, elapsed)
	return &WLock{base: base, token: token, ttl: ttl}, nil
}

// Release 释放写锁（原子清除 writers 和 wtoken）。
func (l *WLock) Release(ctx context.Context, cli *redis.Client) error {
	if cli == nil {
		return ErrNilClient
	}

	// Lua 脚本：验证 token 后删除写者标记
	const script = `
if redis.call('GET', KEYS[2]) == ARGV[1] then
  redis.call('DEL', KEYS[1])
  redis.call('DEL', KEYS[2])
  return 1
end
return 0`

	start := time.Now()
	result, err := cli.Eval(ctx, script,
		[]string{rwWritersKey(l.base), rwWTokenKey(l.base)},
		l.token,
	).Result()
	elapsed := time.Since(start)

	if err != nil && !errors.Is(err, redis.Nil) {
		logx.LoggerFromContext(ctx).Error("cmdx.WLock.Release: base=%s elapsed=%s err=%v",
			l.base, elapsed, err)
		return fmt.Errorf("cmdx.WLock.Release %q: %w", l.base, err)
	}

	n, _ := result.(int64)
	if n == 0 {
		logx.LoggerFromContext(ctx).Warn("cmdx.WLock.Release: base=%s token mismatch or expired elapsed=%s",
			l.base, elapsed)
		return ErrLockExpired
	}

	logx.LoggerFromContext(ctx).Debug("cmdx.WLock.Release: base=%s elapsed=%s released", l.base, elapsed)
	return nil
}

// Renew 续约写锁（重置 writers / wtoken 的 TTL）。
func (l *WLock) Renew(ctx context.Context, cli *redis.Client, ttl time.Duration) (bool, error) {
	if cli == nil {
		return false, ErrNilClient
	}
	if ttl <= 0 {
		return false, fmt.Errorf("cmdx.WLock.Renew %q: ttl must be > 0", l.base)
	}

	const script = `
if redis.call('GET', KEYS[2]) == ARGV[1] then
  redis.call('EXPIRE', KEYS[1], ARGV[2])
  redis.call('EXPIRE', KEYS[2], ARGV[2])
  return 1
end
return 0`

	start := time.Now()
	result, err := cli.Eval(ctx, script,
		[]string{rwWritersKey(l.base), rwWTokenKey(l.base)},
		l.token, int64(ttl.Seconds()),
	).Result()
	elapsed := time.Since(start)

	if err != nil {
		logx.LoggerFromContext(ctx).Error("cmdx.WLock.Renew: base=%s elapsed=%s err=%v",
			l.base, elapsed, err)
		return false, fmt.Errorf("cmdx.WLock.Renew %q: %w", l.base, err)
	}

	n, _ := result.(int64)
	ok := n == 1
	if ok {
		l.ttl = ttl
	}
	logx.LoggerFromContext(ctx).Debug("cmdx.WLock.Renew: base=%s ok=%v elapsed=%s", l.base, ok, elapsed)
	return ok, nil
}

// ─────────────────────────────────────────────
// 便捷封装
// ─────────────────────────────────────────────

// WithRLock 获取读锁后执行 fn，自动释放。
func WithRLock(ctx context.Context, cli *redis.Client, base string, ttl time.Duration, fn func() error) error {
	lock, err := AcquireRLock(ctx, cli, base, ttl)
	if err != nil {
		return err
	}
	defer func() {
		if releaseErr := lock.Release(ctx, cli); releaseErr != nil {
			logx.LoggerFromContext(ctx).Warn("cmdx.WithRLock: release base=%s err=%v", base, releaseErr)
		}
	}()
	return fn()
}

// WithWLock 获取写锁后执行 fn，自动释放。
func WithWLock(ctx context.Context, cli *redis.Client, base string, ttl time.Duration, fn func() error) error {
	lock, err := AcquireWLock(ctx, cli, base, ttl)
	if err != nil {
		return err
	}
	defer func() {
		if releaseErr := lock.Release(ctx, cli); releaseErr != nil && !errors.Is(releaseErr, ErrLockExpired) {
			logx.LoggerFromContext(ctx).Warn("cmdx.WithWLock: release base=%s err=%v", base, releaseErr)
		}
	}()
	return fn()
}

package cmdx

import (
	"context"
	"fmt"
	"time"

	"github.com/moscososirenita-design/tk-common/utils/logx"
	"github.com/redis/go-redis/v9"
)

// ─────────────────────────────────────────────
// 限流操作
//
// 两种策略：
//  1. 固定窗口计数（FixedWindowAllow）：简单高效，适合按分钟/天统计
//  2. 滑动窗口（SlidingWindowAllow）：精度高，基于 ZSet，适合精确 QPS 控制
//
// 适用场景：
//   - SMS 验证码每分钟发送次数限制
//   - 登录失败锁定计数
//   - 接口 QPS 限流
// ─────────────────────────────────────────────

// RateLimitResult 描述一次限流检查的结果。
type RateLimitResult struct {
	// Allowed 为 true 表示本次请求被放行
	Allowed bool
	// Current 是当前窗口已消耗的请求数
	Current int64
	// Limit 是窗口内最大允许数
	Limit int64
}

// ─────────────────────────────────────────────
// 固定窗口计数限流（推荐用于短信/登录限频）
// ─────────────────────────────────────────────

// FixedWindowAllow 使用固定窗口计数判断是否允许本次请求通过。
//
// 参数：
//   - key:   限流 key，建议格式：rate:{domain}:{id}:{窗口时间}
//   - limit: 窗口内最大允许次数
//   - window: 窗口时长（如 time.Minute）
//
// 原子保证：使用 Lua 脚本确保 INCR + 首次 EXPIRE 原子执行。
func FixedWindowAllow(ctx context.Context, cli *redis.Client, key string, limit int64, window time.Duration) (RateLimitResult, error) {
	if cli == nil {
		// 客户端不可用时，降级放行（fail-open）并记录警告
		// 如需 fail-close，在调用方检查返回的 err
		logx.LoggerFromContext(ctx).Warn("cmdx.FixedWindowAllow: nil client, key=%s fail-open", key)
		return RateLimitResult{Allowed: true, Limit: limit}, ErrNilClient
	}

	// Lua 脚本：原子 INCR + 首次 EXPIRE
	const script = `
local n = redis.call('INCR', KEYS[1])
if n == 1 then
  redis.call('EXPIRE', KEYS[1], ARGV[2])
end
return n`

	start := time.Now()
	result, err := cli.Eval(ctx, script,
		[]string{key},
		limit,                   // ARGV[1]（未使用，占位便于扩展）
		int64(window.Seconds()), // ARGV[2] 窗口秒数
	).Result()
	elapsed := time.Since(start)

	if err != nil {
		logx.LoggerFromContext(ctx).Error("cmdx.FixedWindowAllow: key=%s elapsed=%s err=%v",
			key, elapsed, err)
		// Redis 报错时 fail-open，避免影响正常用户；调用方可根据 err 决定是否降级
		return RateLimitResult{Allowed: true, Limit: limit}, fmt.Errorf("cmdx.FixedWindowAllow %q: %w", key, err)
	}

	current, _ := result.(int64)
	allowed := current <= limit

	logx.LoggerFromContext(ctx).Debug("cmdx.FixedWindowAllow: key=%s current=%d limit=%d allowed=%v elapsed=%s",
		key, current, limit, allowed, elapsed)

	return RateLimitResult{
		Allowed: allowed,
		Current: current,
		Limit:   limit,
	}, nil
}

// ─────────────────────────────────────────────
// 多维度限流（手机号 + IP 联合检查）
// ─────────────────────────────────────────────

// MultiWindowAllow 对多个限流 key 逐一检查，任意一个触发限制则返回不允许。
// 适用于 SMS 发送的"手机号分钟限 + 手机号日限 + IP 分钟限"联合频控。
//
// 示例：
//
//	rules := []cmdx.RateLimitRule{
//	    {Key: "rate:sms:phone:minute:13800138000:202401011200", Limit: 1, Window: time.Minute},
//	    {Key: "rate:sms:phone:daily:13800138000:20240101",      Limit: 5, Window: 24*time.Hour},
//	    {Key: "rate:sms:ip:minute:1.2.3.4:202401011200",        Limit: 10, Window: time.Minute},
//	}
//	res, err := cmdx.MultiWindowAllow(ctx, cli, rules)
type RateLimitRule struct {
	// Key 是此条规则对应的 Redis key（由调用方按约定拼接）
	Key string
	// Limit 是窗口内最大允许次数
	Limit int64
	// Window 是窗口时长
	Window time.Duration
}

// MultiWindowAllowResult 描述多维度限流的综合结果。
type MultiWindowAllowResult struct {
	// Allowed 为 true 表示所有维度均通过
	Allowed bool
	// BlockedRule 是首个触发限制的规则（Allowed=false 时有效）
	BlockedRule *RateLimitRule
	// Results 是各规则的独立结果
	Results []RateLimitResult
}

// MultiWindowAllow 对多条限流规则逐一检查。
// 注意：若某条规则触发，后续规则的计数也会自增（无法回滚）；
// 若需精确回滚，请改用事务或 Lua 脚本一次性检查所有规则。
func MultiWindowAllow(ctx context.Context, cli *redis.Client, rules []RateLimitRule) (MultiWindowAllowResult, error) {
	out := MultiWindowAllowResult{Allowed: true, Results: make([]RateLimitResult, len(rules))}

	for i, rule := range rules {
		res, err := FixedWindowAllow(ctx, cli, rule.Key, rule.Limit, rule.Window)
		out.Results[i] = res
		if err != nil {
			// Redis 通信错误时整批 fail-open，记录日志
			logx.LoggerFromContext(ctx).Warn("cmdx.MultiWindowAllow: rule[%d] key=%s err=%v (fail-open)",
				i, rule.Key, err)
			continue
		}
		if !res.Allowed && out.Allowed {
			// 首个触发限制的规则
			out.Allowed = false
			r := rules[i]
			out.BlockedRule = &r
		}
	}
	return out, nil
}

// ─────────────────────────────────────────────
// 滑动窗口限流（精确 QPS 控制，基于 ZSet）
// ─────────────────────────────────────────────

// SlidingWindowAllow 使用 Redis ZSet 实现精确滑动窗口限流。
//
// 原理：以当前时间戳为 score，每次请求往 ZSet 写入一个成员；
//
//	读取 (now-window, now] 区间内的成员数作为当前请求计数；
//	超过 limit 则拒绝，并清理过期成员（惰性 GC）。
//
// 适用场景：API 网关精确 QPS、高频接口防刷
func SlidingWindowAllow(ctx context.Context, cli *redis.Client, key string, limit int64, window time.Duration) (RateLimitResult, error) {
	if cli == nil {
		return RateLimitResult{Allowed: true, Limit: limit}, ErrNilClient
	}

	now := time.Now().UnixMilli() // 毫秒时间戳，精度更高
	windowMs := window.Milliseconds()
	windowStart := now - windowMs    // 窗口左边界
	member := fmt.Sprintf("%d", now) // 每次请求用时间戳作成员（允许同毫秒多请求）

	// Lua 脚本：
	// 1. ZREMRANGEBYSCORE 清除窗口外的旧成员
	// 2. ZADD 写入当前请求
	// 3. ZCARD 统计窗口内请求数
	// 4. EXPIRE 保证 key 自动清理
	const script = `
redis.call('ZREMRANGEBYSCORE', KEYS[1], '-inf', ARGV[1])
redis.call('ZADD', KEYS[1], ARGV[2], ARGV[3])
local count = redis.call('ZCARD', KEYS[1])
redis.call('EXPIRE', KEYS[1], ARGV[4])
return count`

	start := time.Now()
	result, err := cli.Eval(ctx, script,
		[]string{key},
		windowStart,               // ARGV[1] 窗口左边界（毫秒）
		now,                       // ARGV[2] 当前时间戳（作为 score）
		member,                    // ARGV[3] 成员值
		int64(window.Seconds())+1, // ARGV[4] key TTL（稍大于窗口）
	).Result()
	elapsed := time.Since(start)

	if err != nil {
		logx.LoggerFromContext(ctx).Error("cmdx.SlidingWindowAllow: key=%s elapsed=%s err=%v",
			key, elapsed, err)
		return RateLimitResult{Allowed: true, Limit: limit}, fmt.Errorf("cmdx.SlidingWindowAllow %q: %w", key, err)
	}

	current, _ := result.(int64)
	allowed := current <= limit

	logx.LoggerFromContext(ctx).Debug("cmdx.SlidingWindowAllow: key=%s current=%d limit=%d allowed=%v elapsed=%s",
		key, current, limit, allowed, elapsed)

	return RateLimitResult{
		Allowed: allowed,
		Current: current,
		Limit:   limit,
	}, nil
}

// ─────────────────────────────────────────────
// 幂等去重（基于 X-Request-ID）
// ─────────────────────────────────────────────

// IdempotentCheck 检查请求是否已处理过（防重复提交）。
//
// 工作机制：
//   - 首次调用：写入 key（SETNX），返回 true（允许处理）
//   - 重复调用：key 已存在，返回 false（拒绝重复处理）
//
// 参数：
//   - key: 幂等 key，建议使用 idempotent:{service}:{requestID}
//   - ttl: key 保留时长（建议 60s~5min，根据业务操作时长决定）
func IdempotentCheck(ctx context.Context, cli *redis.Client, key string, ttl time.Duration) (bool, error) {
	if cli == nil {
		// 客户端不可用时 fail-open（不影响业务，但丢失幂等保证）
		logx.LoggerFromContext(ctx).Warn("cmdx.IdempotentCheck: nil client, key=%s fail-open", key)
		return true, ErrNilClient
	}
	if ttl <= 0 {
		ttl = 60 * time.Second
	}

	start := time.Now()
	ok, err := cli.SetNX(ctx, key, "1", ttl).Result()
	elapsed := time.Since(start)

	if err != nil {
		logx.LoggerFromContext(ctx).Error("cmdx.IdempotentCheck: key=%s elapsed=%s err=%v",
			key, elapsed, err)
		return true, fmt.Errorf("cmdx.IdempotentCheck %q: %w", key, err)
	}

	logx.LoggerFromContext(ctx).Debug("cmdx.IdempotentCheck: key=%s isNew=%v elapsed=%s",
		key, ok, elapsed)
	// ok=true 表示首次，允许处理；ok=false 表示重复，拒绝
	return ok, nil
}

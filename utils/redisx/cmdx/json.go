package cmdx

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/moscososirenita-design/tk-common/utils/logx"
	"github.com/redis/go-redis/v9"
)

// ─────────────────────────────────────────────
// SetJSON / GetJSON — 结构体序列化读写
// ─────────────────────────────────────────────

// SetJSON 将任意 Go 值序列化为 JSON 后写入 Redis。
// ttl=0 表示永久存储（仅限明确不需要过期的场景，如开奖最终结果）。
// 推荐优先使用 SetJSONEX（强制要求 ttl>0）。
func SetJSON(ctx context.Context, cli *redis.Client, key string, val any, ttl time.Duration) error {
	if cli == nil {
		logx.LoggerFromContext(ctx).Warn("cmdx.SetJSON: nil client, key=%s", key)
		return ErrNilClient
	}

	// 序列化为 JSON
	raw, err := json.Marshal(val)
	if err != nil {
		logx.LoggerFromContext(ctx).Error("cmdx.SetJSON: marshal key=%s err=%v", key, err)
		return fmt.Errorf("cmdx.SetJSON marshal %q: %w", key, err)
	}

	start := time.Now()
	err = cli.Set(ctx, key, raw, ttl).Err()
	elapsed := time.Since(start)

	if err != nil {
		logx.LoggerFromContext(ctx).Error("cmdx.SetJSON: key=%s ttl=%s elapsed=%s err=%v",
			key, ttl, elapsed, err)
		return fmt.Errorf("cmdx.SetJSON %q: %w", key, err)
	}

	logx.LoggerFromContext(ctx).Debug("cmdx.SetJSON: key=%s ttl=%s elapsed=%s ok", key, ttl, elapsed)
	return nil
}

// SetJSONEX 强制要求 ttl>0 的 JSON 写入，防止意外永久存储。
func SetJSONEX(ctx context.Context, cli *redis.Client, key string, val any, ttl time.Duration) error {
	if ttl <= 0 {
		return fmt.Errorf("cmdx.SetJSONEX %q: ttl must be > 0", key)
	}
	return SetJSON(ctx, cli, key, val, ttl)
}

// GetJSON 从 Redis 读取并反序列化 JSON 到 out 指针。
// 返回 (true, nil)  — 命中且解析成功；
// 返回 (false, nil) — key 不存在（缓存 miss，正常业务分支）；
// 返回 (false, err) — Redis 错误或 JSON 解析失败。
func GetJSON(ctx context.Context, cli *redis.Client, key string, out any) (bool, error) {
	if cli == nil {
		return false, ErrNilClient
	}

	start := time.Now()
	raw, err := cli.Get(ctx, key).Bytes()
	elapsed := time.Since(start)

	if isRedisNil(err) {
		// key 不存在：正常 miss，调用方走穿透路径
		logx.LoggerFromContext(ctx).Debug("cmdx.GetJSON: key=%s miss elapsed=%s", key, elapsed)
		return false, nil
	}
	if err != nil {
		logx.LoggerFromContext(ctx).Error("cmdx.GetJSON: key=%s elapsed=%s err=%v", key, elapsed, err)
		return false, fmt.Errorf("cmdx.GetJSON %q: %w", key, err)
	}
	if len(raw) == 0 {
		// 存储的是空字节（防缓存穿透的空值占位）
		logx.LoggerFromContext(ctx).Debug("cmdx.GetJSON: key=%s empty value elapsed=%s", key, elapsed)
		return false, nil
	}

	// 反序列化
	if jsonErr := json.Unmarshal(raw, out); jsonErr != nil {
		logx.LoggerFromContext(ctx).Error("cmdx.GetJSON: unmarshal key=%s err=%v", key, jsonErr)
		return false, fmt.Errorf("cmdx.GetJSON unmarshal %q: %w", key, jsonErr)
	}

	logx.LoggerFromContext(ctx).Debug("cmdx.GetJSON: key=%s hit elapsed=%s", key, elapsed)
	return true, nil
}

// ─────────────────────────────────────────────
// SetNullGuard — 防缓存穿透：写空值占位
// ─────────────────────────────────────────────

// nullPlaceholder 是写入 Redis 的空值占位符，区分"未查询"和"查过了但无数据"。
const nullPlaceholder = "\x00null"

// SetNullGuard 在查询结果为空时写入占位符，防止缓存穿透导致 DB 被击穿。
// ttl 建议设置较短（如 60s），避免占位符长期屏蔽真实数据。
func SetNullGuard(ctx context.Context, cli *redis.Client, key string, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = 60 * time.Second // 空值占位默认 60s
	}
	return Set(ctx, cli, key, nullPlaceholder, ttl)
}

// IsNullGuard 判断读取到的字符串值是否为空值占位符。
// 调用方在 Get 之后先调用此函数，若为 true 则直接返回空结果，不穿透 DB。
func IsNullGuard(val string) bool {
	return val == nullPlaceholder
}

// ─────────────────────────────────────────────
// MSetJSON / MGetJSON — 批量 JSON 读写
// ─────────────────────────────────────────────

// MSetJSONItem 是批量写入的单个条目。
type MSetJSONItem struct {
	Key string
	Val any
	TTL time.Duration
}

// MSetJSON 使用 Pipeline 批量写入多个 JSON key，减少网络往返。
// 各 key 可以有不同 TTL；任意一个序列化失败会中止整批写入。
func MSetJSON(ctx context.Context, cli *redis.Client, items []MSetJSONItem) error {
	if cli == nil {
		return ErrNilClient
	}
	if len(items) == 0 {
		return nil
	}

	// 预先序列化，失败则整批中止（避免部分写入）
	type kv struct {
		key string
		raw []byte
		ttl time.Duration
	}
	pairs := make([]kv, 0, len(items))
	for _, item := range items {
		raw, err := json.Marshal(item.Val)
		if err != nil {
			return fmt.Errorf("cmdx.MSetJSON marshal key=%q: %w", item.Key, err)
		}
		pairs = append(pairs, kv{key: item.Key, raw: raw, ttl: item.TTL})
	}

	start := time.Now()
	pipe := cli.Pipeline()
	for _, p := range pairs {
		pipe.Set(ctx, p.key, p.raw, p.ttl)
	}
	_, err := pipe.Exec(ctx)
	elapsed := time.Since(start)

	if err != nil {
		logx.LoggerFromContext(ctx).Error("cmdx.MSetJSON: count=%d elapsed=%s err=%v",
			len(items), elapsed, err)
		return fmt.Errorf("cmdx.MSetJSON: %w", err)
	}

	logx.LoggerFromContext(ctx).Debug("cmdx.MSetJSON: count=%d elapsed=%s ok", len(items), elapsed)
	return nil
}

// MGetJSONResult 是批量读取的单个结果。
type MGetJSONResult struct {
	Key string
	Hit bool  // true=命中，false=miss
	Err error // 仅该 key 的解析错误
}

// MGetJSON 批量读取并反序列化多个 JSON key。
// out 是与 keys 等长的目标指针切片，由调用方提前分配好类型。
// 任意 key miss 或报错不影响其余 key 的读取，通过 results 逐一报告。
func MGetJSON(ctx context.Context, cli *redis.Client, keys []string, out []any) ([]MGetJSONResult, error) {
	if cli == nil {
		return nil, ErrNilClient
	}
	if len(keys) == 0 {
		return nil, nil
	}
	if len(keys) != len(out) {
		return nil, fmt.Errorf("cmdx.MGetJSON: keys len(%d) != out len(%d)", len(keys), len(out))
	}

	start := time.Now()
	vals, err := cli.MGet(ctx, keys...).Result()
	elapsed := time.Since(start)

	if err != nil {
		logx.LoggerFromContext(ctx).Error("cmdx.MGetJSON: count=%d elapsed=%s err=%v",
			len(keys), elapsed, err)
		return nil, fmt.Errorf("cmdx.MGetJSON: %w", err)
	}

	logx.LoggerFromContext(ctx).Debug("cmdx.MGetJSON: count=%d elapsed=%s", len(keys), elapsed)

	results := make([]MGetJSONResult, len(keys))
	for i, raw := range vals {
		results[i].Key = keys[i]
		if raw == nil {
			// key 不存在
			results[i].Hit = false
			continue
		}
		str, ok := raw.(string)
		if !ok || len(str) == 0 {
			results[i].Hit = false
			continue
		}
		if jsonErr := json.Unmarshal([]byte(str), out[i]); jsonErr != nil {
			results[i].Hit = false
			results[i].Err = fmt.Errorf("unmarshal key=%q: %w", keys[i], jsonErr)
			logx.LoggerFromContext(ctx).Warn("cmdx.MGetJSON: %v", results[i].Err)
			continue
		}
		results[i].Hit = true
	}

	return results, nil
}

// ─────────────────────────────────────────────
// 内部工具
// ─────────────────────────────────────────────

// isRedisNil 判断 error 是否为 redis.Nil（key 不存在语义）。
func isRedisNil(err error) bool {
	return err != nil && err.Error() == redis.Nil.Error()
}

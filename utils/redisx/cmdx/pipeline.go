package cmdx

import (
	"context"
	"fmt"
	"time"

	"github.com/moscososirenita-design/tk-common/utils/logx"
	"github.com/redis/go-redis/v9"
)

// ─────────────────────────────────────────────
// Pipeline 批量操作
//
// 适用场景：
//   - 批量写入配置缓存（Banner/广播/外链等）
//   - 批量 DEL 失效一组缓存
//   - 批量 GET 减少网络往返
// ─────────────────────────────────────────────

// PipeSetItem 是 Pipeline 批量写入的单个条目。
type PipeSetItem struct {
	Key string
	Val string
	TTL time.Duration
}

// PipeSet 使用 Pipeline 批量写入多个字符串 key，减少网络往返。
// 所有命令在一次网络往返中提交；单个命令失败不会影响其他命令。
// 若需事务语义（全部成功或全部失败），请改用 TxPipeSet。
func PipeSet(ctx context.Context, cli *redis.Client, items []PipeSetItem) error {
	if cli == nil {
		return ErrNilClient
	}
	if len(items) == 0 {
		return nil
	}

	start := time.Now()
	pipe := cli.Pipeline()
	for _, item := range items {
		pipe.Set(ctx, item.Key, item.Val, item.TTL)
	}
	cmds, err := pipe.Exec(ctx)
	elapsed := time.Since(start)

	if err != nil {
		logx.LoggerFromContext(ctx).Error("cmdx.PipeSet: count=%d elapsed=%s err=%v",
			len(items), elapsed, err)
		return fmt.Errorf("cmdx.PipeSet: %w", err)
	}

	// 收集各命令的执行错误（Pipeline 不会因单个失败而中止）
	var firstErr error
	for i, cmd := range cmds {
		if cmd.Err() != nil {
			logx.LoggerFromContext(ctx).Warn("cmdx.PipeSet: key=%s err=%v", items[i].Key, cmd.Err())
			if firstErr == nil {
				firstErr = cmd.Err()
			}
		}
	}

	logx.LoggerFromContext(ctx).Debug("cmdx.PipeSet: count=%d elapsed=%s ok", len(items), elapsed)
	return firstErr
}

// PipeDel 使用 Pipeline 批量删除多个 key。
// 返回实际被删除的 key 总数（不存在的 key 不计入）。
func PipeDel(ctx context.Context, cli *redis.Client, keys []string) (int64, error) {
	if cli == nil {
		return 0, ErrNilClient
	}
	if len(keys) == 0 {
		return 0, nil
	}

	start := time.Now()
	// Redis DEL 支持一次传多个 key，直接调用更高效
	n, err := cli.Del(ctx, keys...).Result()
	elapsed := time.Since(start)

	if err != nil {
		logx.LoggerFromContext(ctx).Error("cmdx.PipeDel: count=%d elapsed=%s err=%v",
			len(keys), elapsed, err)
		return 0, fmt.Errorf("cmdx.PipeDel: %w", err)
	}

	logx.LoggerFromContext(ctx).Debug("cmdx.PipeDel: total=%d deleted=%d elapsed=%s",
		len(keys), n, elapsed)
	return n, nil
}

// PipeGet 使用 MGet 批量读取字符串，返回与 keys 等长的结果切片。
// 结果中 nil 表示对应 key 不存在（miss）。
// 返回的 []string 保证与 keys 一一对应，不存在的 key 位置值为 ""，hit[i]=false。
func PipeGet(ctx context.Context, cli *redis.Client, keys []string) (vals []string, hits []bool, err error) {
	if cli == nil {
		return nil, nil, ErrNilClient
	}
	if len(keys) == 0 {
		return nil, nil, nil
	}

	start := time.Now()
	rawVals, mgetErr := cli.MGet(ctx, keys...).Result()
	elapsed := time.Since(start)

	if mgetErr != nil {
		logx.LoggerFromContext(ctx).Error("cmdx.PipeGet: count=%d elapsed=%s err=%v",
			len(keys), elapsed, mgetErr)
		return nil, nil, fmt.Errorf("cmdx.PipeGet: %w", mgetErr)
	}

	vals = make([]string, len(keys))
	hits = make([]bool, len(keys))
	for i, raw := range rawVals {
		if raw != nil {
			if s, ok := raw.(string); ok {
				vals[i] = s
				hits[i] = true
			}
		}
	}

	hitCount := 0
	for _, h := range hits {
		if h {
			hitCount++
		}
	}
	logx.LoggerFromContext(ctx).Debug("cmdx.PipeGet: total=%d hits=%d elapsed=%s",
		len(keys), hitCount, elapsed)
	return vals, hits, nil
}

// ─────────────────────────────────────────────
// TxPipeSet — 事务 Pipeline（MULTI/EXEC）
// ─────────────────────────────────────────────

// TxPipeSet 使用 MULTI/EXEC 事务批量写入多个 key，保证原子性（全部成功或全部失败）。
// 注意：Redis 事务不支持回滚，EXEC 只保证命令被顺序执行，不保证业务层的逻辑一致性。
func TxPipeSet(ctx context.Context, cli *redis.Client, items []PipeSetItem) error {
	if cli == nil {
		return ErrNilClient
	}
	if len(items) == 0 {
		return nil
	}

	start := time.Now()
	_, err := cli.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		for _, item := range items {
			pipe.Set(ctx, item.Key, item.Val, item.TTL)
		}
		return nil
	})
	elapsed := time.Since(start)

	if err != nil {
		logx.LoggerFromContext(ctx).Error("cmdx.TxPipeSet: count=%d elapsed=%s err=%v",
			len(items), elapsed, err)
		return fmt.Errorf("cmdx.TxPipeSet: %w", err)
	}

	logx.LoggerFromContext(ctx).Debug("cmdx.TxPipeSet: count=%d elapsed=%s ok", len(items), elapsed)
	return nil
}

// ─────────────────────────────────────────────
// 缓存失效：按前缀批量删除（SCAN + DEL，不阻塞主线程）
// ─────────────────────────────────────────────

// DelByPattern 使用 SCAN 迭代匹配 pattern 的 key 并批量删除。
// 注意：SCAN 是非阻塞的，但在 key 数量极大时可能耗时较长；
//
//	生产环境建议在低峰期或异步任务中调用。
//
// pattern 示例："tk:home:*"、"rbac:perms:*"
func DelByPattern(ctx context.Context, cli *redis.Client, pattern string) (int64, error) {
	if cli == nil {
		return 0, ErrNilClient
	}

	var totalDeleted int64
	var cursor uint64
	batch := make([]string, 0, 100)

	start := time.Now()
	for {
		// 每批扫描 100 个 key
		keys, nextCursor, err := cli.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			logx.LoggerFromContext(ctx).Error("cmdx.DelByPattern: pattern=%s cursor=%d err=%v",
				pattern, cursor, err)
			return totalDeleted, fmt.Errorf("cmdx.DelByPattern %q: %w", pattern, err)
		}

		batch = append(batch, keys...)

		// 积累 100 个以上时批量删除，减少 DEL 调用次数
		if len(batch) >= 100 {
			n, delErr := cli.Del(ctx, batch...).Result()
			if delErr != nil {
				logx.LoggerFromContext(ctx).Warn("cmdx.DelByPattern: batch del err=%v", delErr)
			}
			totalDeleted += n
			batch = batch[:0] // 清空复用 slice
		}

		cursor = nextCursor
		if cursor == 0 {
			break // SCAN 迭代完毕
		}
	}

	// 删除最后一批不足 100 个的 key
	if len(batch) > 0 {
		n, _ := cli.Del(ctx, batch...).Result()
		totalDeleted += n
	}

	elapsed := time.Since(start)
	logx.LoggerFromContext(ctx).Debug("cmdx.DelByPattern: pattern=%s deleted=%d elapsed=%s",
		pattern, totalDeleted, elapsed)
	return totalDeleted, nil
}

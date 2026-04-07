// Package cachex 提供缓存相关工具，包括 singleflight 请求合并与缓存加载封装。
package cachex

import (
	"context"
	"sync"
	"time"

	"github.com/moscososirenita-design/tk-common/utils/logx"
	redisx "github.com/moscososirenita-design/tk-common/utils/redisx/v9"
	"github.com/redis/go-redis/v9"
)

// call 表示一次正在进行的函数调用。
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

// Group 提供 singleflight 能力：对同一 key 的并发请求只执行一次实际调用，
// 其他请求共享结果。相比分布式锁 + sleep 等待，singleflight 无需轮询，延迟更低。
type Group struct {
	mu sync.Mutex
	m  map[string]*call
}

// Do 对同一 key 的并发调用，只有第一个调用会执行 fn，其他调用等待并共享结果。
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}

// cacheWriteSem 限制异步写缓存的并发 goroutine 数量，防止 Redis 慢时 goroutine 堆积。
var cacheWriteSem = make(chan struct{}, 32)

// CacheLoader 封装 "读缓存 -> singleflight 合并 -> 查 DB -> 异步写缓存" 完整流程。
type CacheLoader struct {
	redis *redis.Client
	group Group
}

// NewCacheLoader 创建缓存加载器，redisClient 为 nil 时退化为无缓存模式。
func NewCacheLoader(redisClient *redis.Client) *CacheLoader {
	return &CacheLoader{
		redis: redisClient,
	}
}

// LoadMap 加载 map 类型缓存数据。
// cacheKey: Redis 缓存键
// ttl: 缓存过期时间
// loader: 缓存 miss 时的数据加载函数（查 DB）
func (cl *CacheLoader) LoadMap(ctx context.Context, cacheKey string, ttl time.Duration, loader func() (map[string]interface{}, error)) (map[string]interface{}, error) {
	logger := logx.LoggerFromContext(ctx)

	// 1. 尝试读缓存
	if cl.redis != nil {
		var cached map[string]interface{}
		hit, err := redisx.GetJSON(ctx, cl.redis, cacheKey, &cached)
		if err != nil {
			logger.Warn("cachex.LoadMap: cache get err=%v key=%s, fallback", err, cacheKey)
		} else if hit && len(cached) > 0 {
			logger.Debug("cachex.LoadMap: cache hit key=%s", cacheKey)
			return cached, nil
		}
	}

	// 2. singleflight 合并同 key 并发请求
	logger.Info("cachex.LoadMap: cache miss key=%s, loading via singleflight", cacheKey)
	val, err := cl.group.Do(cacheKey, func() (interface{}, error) {
		// double-check：获得执行权后再读一次缓存
		if cl.redis != nil {
			var cached map[string]interface{}
			if hit, _ := redisx.GetJSON(ctx, cl.redis, cacheKey, &cached); hit && len(cached) > 0 {
				logger.Debug("cachex.LoadMap: singleflight double-check hit key=%s", cacheKey)
				return cached, nil
			}
		}

		// 执行实际加载
		data, loadErr := loader()
		if loadErr != nil {
			return nil, loadErr
		}

		// 异步写缓存（受 semaphore 限制并发数）
		if cl.redis != nil && len(data) > 0 {
			go func() {
				select {
				case cacheWriteSem <- struct{}{}:
					defer func() { <-cacheWriteSem }()
					writeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
					defer cancel()
					if writeErr := redisx.SetJSON(writeCtx, cl.redis, cacheKey, data, ttl); writeErr != nil {
						logger.Warn("cachex.LoadMap: async cache set failed key=%s err=%v", cacheKey, writeErr)
					} else {
						logger.Info("cachex.LoadMap: cache set ok key=%s ttl=%s", cacheKey, ttl)
					}
				default:
					logger.Warn("cachex.LoadMap: async write skipped (concurrency limit) key=%s", cacheKey)
				}
			}()
		}

		return data, nil
	})
	if err != nil {
		return nil, err
	}
	return val.(map[string]interface{}), nil
}

// LoadSlice 加载 slice 类型缓存数据。
func (cl *CacheLoader) LoadSlice(ctx context.Context, cacheKey string, ttl time.Duration, loader func() ([]map[string]interface{}, error)) ([]map[string]interface{}, error) {
	logger := logx.LoggerFromContext(ctx)

	// 1. 尝试读缓存
	if cl.redis != nil {
		var cached []map[string]interface{}
		hit, err := redisx.GetJSON(ctx, cl.redis, cacheKey, &cached)
		if err != nil {
			logger.Warn("cachex.LoadSlice: cache get err=%v key=%s, fallback", err, cacheKey)
		} else if hit {
			logger.Debug("cachex.LoadSlice: cache hit key=%s count=%d", cacheKey, len(cached))
			return cached, nil
		}
	}

	// 2. singleflight 合并
	logger.Info("cachex.LoadSlice: cache miss key=%s, loading via singleflight", cacheKey)
	val, err := cl.group.Do(cacheKey, func() (interface{}, error) {
		if cl.redis != nil {
			var cached []map[string]interface{}
			if hit, _ := redisx.GetJSON(ctx, cl.redis, cacheKey, &cached); hit {
				return cached, nil
			}
		}

		data, loadErr := loader()
		if loadErr != nil {
			return nil, loadErr
		}

		if cl.redis != nil && len(data) > 0 {
			go func() {
				select {
				case cacheWriteSem <- struct{}{}:
					defer func() { <-cacheWriteSem }()
					writeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
					defer cancel()
					if writeErr := redisx.SetJSON(writeCtx, cl.redis, cacheKey, data, ttl); writeErr != nil {
						logger.Warn("cachex.LoadSlice: async cache set failed key=%s err=%v", cacheKey, writeErr)
					} else {
						logger.Info("cachex.LoadSlice: cache set ok key=%s ttl=%s", cacheKey, ttl)
					}
				default:
					logger.Warn("cachex.LoadSlice: async write skipped (concurrency limit) key=%s", cacheKey)
				}
			}()
		}

		return data, nil
	})
	if err != nil {
		return nil, err
	}
	return val.([]map[string]interface{}), nil
}

// Invalidate 主动清除缓存键。
func (cl *CacheLoader) Invalidate(ctx context.Context, key string) error {
	if cl.redis == nil {
		return nil
	}
	logger := logx.LoggerFromContext(ctx)
	logger.Info("cachex.Invalidate: key=%s", key)
	return cl.redis.Del(ctx, key).Err()
}

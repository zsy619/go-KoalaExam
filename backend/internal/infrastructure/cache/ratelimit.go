// Package cache 基础设施 - 缓存与限流。
package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisLoginLimiter 基于 Redis 的登录限流器。
//
// 策略：滑动窗口
//   - 5 分钟内最多 5 次失败
//   - 超过后锁定 5 分钟
type RedisLoginLimiter struct {
	rdb *redis.Client
	max int           // 最大尝试次数
	window time.Duration // 时间窗口
}

// NewRedisLoginLimiter 构造登录限流器。
func NewRedisLoginLimiter(rdb *redis.Client, max int, window time.Duration) *RedisLoginLimiter {
	if max <= 0 {
		max = 5
	}
	if window <= 0 {
		window = 5 * time.Minute
	}
	return &RedisLoginLimiter{rdb: rdb, max: max, window: window}
}

// Allow 检查是否允许登录。
func (l *RedisLoginLimiter) Allow(ctx context.Context, key string) (bool, time.Duration, error) {
	if l.rdb == nil {
		return true, 0, nil
	}
	redisKey := fmt.Sprintf("ratelimit:login:%s", key)

	// 获取当前计数
	count, err := l.rdb.Get(ctx, redisKey).Int()
	if err == redis.Nil {
		count = 0
	} else if err != nil {
		return true, 0, err
	}

	if count >= l.max {
		// 已锁定
		ttl, _ := l.rdb.TTL(ctx, redisKey).Result()
		return false, ttl, nil
	}

	// 增加计数
	pipe := l.rdb.TxPipeline()
	pipe.Incr(ctx, redisKey)
	pipe.Expire(ctx, redisKey, l.window)
	if _, err := pipe.Exec(ctx); err != nil {
		return true, 0, err
	}

	return true, 0, nil
}

// Reset 重置限流。
func (l *RedisLoginLimiter) Reset(ctx context.Context, key string) error {
	if l.rdb == nil {
		return nil
	}
	redisKey := fmt.Sprintf("ratelimit:login:%s", key)
	return l.rdb.Del(ctx, redisKey).Err()
}

// RedisTokenBlacklist 基于 Redis 的 Token 黑名单。
type RedisTokenBlacklist struct {
	rdb *redis.Client
}

// NewRedisTokenBlacklist 构造 Token 黑名单。
func NewRedisTokenBlacklist(rdb *redis.Client) *RedisTokenBlacklist {
	return &RedisTokenBlacklist{rdb: rdb}
}

// Add 添加 token 到黑名单。
func (b *RedisTokenBlacklist) Add(ctx context.Context, token string, ttl time.Duration) error {
	if b.rdb == nil {
		return nil
	}
	key := fmt.Sprintf("blacklist:token:%x", hashToken(token))
	return b.rdb.Set(ctx, key, "1", ttl).Err()
}

// IsBlacklisted 检查 token 是否在黑名单。
func (b *RedisTokenBlacklist) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	if b.rdb == nil {
		return false, nil
	}
	key := fmt.Sprintf("blacklist:token:%x", hashToken(token))
	n, err := b.rdb.Exists(ctx, key).Result()
	return n > 0, err
}

// hashToken 计算 token 哈希（避免完整 token 存 Redis）。
func hashToken(token string) string {
	h := sha256.New()
	h.Write([]byte(token))
	return hex.EncodeToString(h.Sum(nil))
}

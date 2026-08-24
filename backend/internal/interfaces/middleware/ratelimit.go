package middleware

import (
	"context"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/redis/go-redis/v9"

	"github.com/your-team/koala-exam-backend/internal/domain/errcode"
	"github.com/your-team/koala-exam-backend/pkg/response"
)

// RateLimit 基于 Redis 的令牌桶限流
func RateLimit(rdb *redis.Client, qps, burst int) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		ip := c.ClientIP()
		path := string(c.Path())
		key := fmt.Sprintf("koala:ratelimit:%s:%s", ip, path)
		cnt, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			c.Next(ctx)
			return
		}
		if cnt == 1 {
			rdb.Expire(ctx, key, 60)
		}
		if cnt > int64(burst) {
			response.Fail(c, 429, errcode.CodeTooManyRequest, "请求过于频繁")
			c.Abort()
			return
		}
		c.Next(ctx)
	}
}

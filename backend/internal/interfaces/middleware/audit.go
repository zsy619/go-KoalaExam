package middleware

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// Audit 行为审计日志中间件
func Audit() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		start := time.Now()
		c.Next(ctx)
		cost := time.Since(start)
		hlog.Infof("%s %s | ip=%s | code=%d | cost=%s",
			c.Method(), c.Path(), c.ClientIP(), c.Response.StatusCode(), cost)
	}
}

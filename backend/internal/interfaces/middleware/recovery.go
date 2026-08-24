package middleware

import (
	"context"
	"runtime/debug"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"

	"github.com/your-team/koala-exam-backend/pkg/response"
)

// Recovery panic 恢复中间件
func Recovery() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		defer func() {
			if r := recover(); r != nil {
				hlog.Errorf("panic: %v\n%s", r, debug.Stack())
				response.Fail(c, 500, 100005, "服务器内部错误")
				c.Abort()
			}
		}()
		c.Next(ctx)
	}
}

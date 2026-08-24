package middleware

import (
	"context"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/your-team/koala-exam-backend/internal/domain/consts"
	"github.com/your-team/koala-exam-backend/internal/domain/errcode"
	"github.com/your-team/koala-exam-backend/pkg/jwt"
	"github.com/your-team/koala-exam-backend/pkg/response"
)

// Auth JWT 鉴权中间件
func Auth(helper *jwt.Helper) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		token := c.GetHeader("Authorization")
		if len(token) == 0 {
			response.Fail(c, 401, errcode.CodeUnauthorized, "未登录或 token 缺失")
			c.Abort()
			return
		}
		tokenStr := strings.TrimPrefix(string(token), "Bearer ")
		claims, err := helper.Parse(tokenStr)
		if err != nil {
			response.Fail(c, 401, errcode.CodeTokenInvalid, "token 无效")
			c.Abort()
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next(ctx)
	}
}

// RequireRole 角色权限校验
func RequireRole(roles ...int8) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		roleI, ok := c.Get("role")
		if !ok {
			response.Fail(c, 401, errcode.CodeUnauthorized, "请先登录")
			c.Abort()
			return
		}
		role := roleI.(int8)
		allow := false
		for _, r := range roles {
			if r == role {
				allow = true
				break
			}
		}
		if !allow {
			response.Fail(c, 403, errcode.CodePermissionDenied, "权限不足")
			c.Abort()
			return
		}
		c.Next(ctx)
	}
}

// AdminOnly 仅超管
func AdminOnly() app.HandlerFunc {
	return RequireRole(consts.RoleSuperAdmin)
}

// TeacherOrAdmin 教师或超管
func TeacherOrAdmin() app.HandlerFunc {
	return RequireRole(consts.RoleSuperAdmin, consts.RoleTeacher)
}

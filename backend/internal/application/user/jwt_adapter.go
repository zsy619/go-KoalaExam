// Package user - JWT adapter for application layer.
package user

import (
	"time"

	"github.com/your-team/koala-exam-backend/pkg/jwt"
)

// JwtTokenAdapter 将 jwt.Helper 适配为 TokenService 接口。
type JwtTokenAdapter struct {
	helper *jwt.Helper
}

// NewJwtTokenAdapter 构造适配器。
func NewJwtTokenAdapter(h *jwt.Helper) *JwtTokenAdapter {
	return &JwtTokenAdapter{helper: h}
}

// Generate 生成 token。
func (a *JwtTokenAdapter) Generate(uid int64, username string, role int8, tokenType string) (string, time.Time, error) {
	return a.helper.Generate(uid, username, role, tokenType)
}

// Parse 解析 token 返回 (uid, role, error)。
func (a *JwtTokenAdapter) Parse(token string) (int64, int8, error) {
	claims, err := a.helper.Parse(token)
	if err != nil {
		return 0, 0, err
	}
	return claims.UserID, claims.Role, nil
}

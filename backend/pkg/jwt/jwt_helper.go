// Package jwt JWT 工具。
//
// 遵循 Google Go 风格：
//   - 通过 Helper 结构体持有配置
//   - 显式错误返回
//   - 不导出内部状态
package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Token 类型常量
const (
	AccessToken  = "access"
	RefreshToken = "refresh"
)

// AccessTTL 是默认的 access token 有效期（2 小时）。
const AccessTTL = 2 * time.Hour

// Claims 自定义 JWT 声明。
type Claims struct {
	UserID   int64  `json:"uid"`
	Username string `json:"uname"`
	Role     int8   `json:"role"`
	Type     string `json:"type"` // access | refresh
	jwt.RegisteredClaims
}

// Helper JWT 工具。
type Helper struct {
	secret        []byte
	issuer        string
	accessExpire  int // 秒
	refreshExpire int // 秒
}

// New 构造 Helper。
func New(secret, issuer string, accessExp, refreshExp int) *Helper {
	return &Helper{
		secret:        []byte(secret),
		issuer:        issuer,
		accessExpire:  accessExp,
		refreshExpire: refreshExp,
	}
}

// Generate 生成 token（返回 token、过期时间、error）。
func (h *Helper) Generate(userID int64, username string, role int8, tokenType string) (string, time.Time, error) {
	expSec := h.accessExpire
	if tokenType == RefreshToken {
		expSec = h.refreshExpire
	}
	exp := time.Now().Add(time.Duration(expSec) * time.Second)
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		Type:     tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    h.issuer,
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(h.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, exp, nil
}

// Parse 解析 token 并返回 Claims。
func (h *Helper) Parse(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return h.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// Refresh 用 refresh token 换 access token。
func (h *Helper) Refresh(refreshToken string) (string, time.Time, error) {
	claims, err := h.Parse(refreshToken)
	if err != nil {
		return "", time.Time{}, err
	}
	if claims.Type != RefreshToken {
		return "", time.Time{}, errors.New("not a refresh token")
	}
	return h.Generate(claims.UserID, claims.Username, claims.Role, AccessToken)
}

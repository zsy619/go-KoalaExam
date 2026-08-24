package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims 自定义 JWT 声明
type Claims struct {
	UserID   int64  `json:"uid"`
	Username string `json:"uname"`
	Role     int8   `json:"role"`
	Type     string `json:"type"` // access | refresh
	jwt.RegisteredClaims
}

// Helper JWT 工具
type Helper struct {
	secret        []byte
	issuer        string
	accessExpire  int
	refreshExpire int
}

func New(secret, issuer string, accessExp, refreshExp int) *Helper {
	return &Helper{
		secret:        []byte(secret),
		issuer:        issuer,
		accessExpire:  accessExp,
		refreshExpire: refreshExp,
	}
}

// Generate 生成 token
func (h *Helper) Generate(userID int64, username string, role int8, tokenType string) (string, time.Time, error) {
	var expSec int
	if tokenType == "refresh" {
		expSec = h.refreshExpire
	} else {
		expSec = h.accessExpire
	}
	exp := time.Now().Add(time.Duration(expSec) * time.Second)
	claims := Claims{
		UserID: userID, Username: username, Role: role, Type: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    h.issuer,
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := t.SignedString(h.secret)
	return token, exp, err
}

// Parse 解析 token
func (h *Helper) Parse(tokenStr string) (*Claims, error) {
	t, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return h.secret, nil
	})
	if err != nil {
		return nil, err
	}
	c, ok := t.Claims.(*Claims)
	if !ok || !t.Valid {
		return nil, errors.New("invalid token")
	}
	return c, nil
}

// Refresh 刷新 token
func (h *Helper) Refresh(refreshToken string) (string, time.Time, error) {
	c, err := h.Parse(refreshToken)
	if err != nil {
		return "", time.Time{}, err
	}
	if c.Type != "refresh" {
		return "", time.Time{}, errors.New("not a refresh token")
	}
	return h.Generate(c.UserID, c.Username, c.Role, "access")
}

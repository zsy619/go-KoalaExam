package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// UUID 生成 UUID v4
func UUID() string { return uuid.NewString() }

// RandString 随机字符串
func RandString(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// NowUnix 当前 unix 时间戳（秒）
func NowUnix() int64 { return time.Now().Unix() }

// FormatTime YYYY-MM-DD HH:mm:ss
func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// ParseInt64 安全转换
func ParseInt64(s string, def int64) int64 {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return def
	}
	return v
}

// ParseInt 安全转换
func ParseInt(s string, def int) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}

// GetClientIP 获取客户端 IP
func GetClientIP(remoteAddr string, xForwardedFor string) string {
	if xForwardedFor != "" {
		return xForwardedFor
	}
	return remoteAddr
}

// SecToDuration 秒转 mm:ss
func SecToDuration(sec int64) string {
	m := sec / 60
	s := sec % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}

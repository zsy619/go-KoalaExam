package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/redis/go-redis/v9"

	"github.com/your-team/koala-exam-backend/pkg/config"
)

// InitRedis 初始化 Redis
func InitRedis(cfg config.RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  time.Duration(cfg.DialTimeout) * time.Second,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}
	hlog.Infof("Redis connected: %s:%d", cfg.Host, cfg.Port)
	return rdb, nil
}

// RedisKey 集中管理 key 前缀
type RedisKey string

const (
	KeyExamProgress   RedisKey = "koala:exam:progress:%d:%d" // userID, examID
	KeyExamPaperCache RedisKey = "koala:exam:paper:%d"       // examID
	KeyRateLimit      RedisKey = "koala:ratelimit:%s:%s"     // ip, path
	KeyWrongStats     RedisKey = "koala:wrong:stats:%d"      // userID
	KeyUserSession    RedisKey = "koala:session:%d"          // userID
	KeyQuestionCache  RedisKey = "koala:question:%d"         // questionID
)

// Build 构造 key
func (k RedisKey) Build(args ...interface{}) string {
	return fmt.Sprintf(string(k), args...)
}

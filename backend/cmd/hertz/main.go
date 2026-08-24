package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"

	"github.com/your-team/koala-exam-backend/internal/infrastructure/cache"
	"github.com/your-team/koala-exam-backend/internal/infrastructure/database"
	"github.com/your-team/koala-exam-backend/internal/interfaces/router"
	"github.com/your-team/koala-exam-backend/pkg/config"
	"github.com/your-team/koala-exam-backend/pkg/jwt"
	"github.com/your-team/koala-exam-backend/pkg/logger"
)

// @title KoalaExam API
// //  @version 1.0
// //  @description 基于 Hertz + Vue3 的在线考试系统 API
// //  @host localhost:8080
// //  @BasePath /api/v1
func main() {
	// 加载配置
	env := os.Getenv("APP_ENV")
	if env == "" { env = "dev" }
	cfg, err := config.LoadConfigByEnv(env)
	if err != nil { log.Fatalf("load config failed: %v", err) }

	// 初始化日志
	_, err = logger.New(cfg.Log.Level, cfg.Log.Path, cfg.Log.MaxSize, cfg.Log.MaxAge, cfg.Log.MaxBackups, cfg.Log.Compress)
	if err != nil { log.Fatalf("init logger failed: %v", err) }
	// hlog.SetLogger(zl) - using hlog default

	hlog.Infof("🦥 KoalaExam starting, env=%s, mode=%s", env, cfg.App.Mode)

	// 初始化 MySQL
	db, err := database.InitMySQL(cfg.MySQL)
	if err != nil { hlog.Fatalf("init mysql failed: %v", err) }
	if env == "dev" {
		if err := database.AutoMigrate(db); err != nil {
			hlog.Warnf("auto migrate failed: %v", err)
		} else {
		hlog.Info("database auto-migrated")
	}
	}

	// 初始化 Redis
	rdb, err := cache.InitRedis(cfg.Redis)
	if err != nil { hlog.Fatalf("init redis failed: %v", err) }

	// 初始化 JWT
	jwtHelper := jwt.New(cfg.JWT.Secret, cfg.JWT.Issuer, cfg.JWT.AccessExpire, cfg.JWT.RefreshExpire)

	// 创建 Hertz Server（可启用 Netpoll 高性能网络库）
	addr := cfg.App.Host + ":" + strconv.Itoa(cfg.App.Port)
	_ = addr // Use addr when creating server below
	if cfg.App.EnableNetpoll {
		// Hertz 默认使用标准库；如需 Netpoll：opts = append(opts, server.WithTransport(transport.NewNetpollTransport()))
	}
	h := server.Default(server.WithHostPorts(addr), server.WithReadTimeout(time.Duration(cfg.App.ReadTimeout)*time.Second), server.WithWriteTimeout(time.Duration(cfg.App.WriteTimeout)*time.Second))
	hlog.SetLevel(hlog.LevelInfo)

	// 注册路由
	router.RegisterRoutes(h, db, rdb, jwtHelper)

	// 优雅退出
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		hlog.Info("shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		h.Shutdown(ctx)
	}()

	hlog.Infof("KoalaExam server started at %s:%d", cfg.App.Host, cfg.App.Port)
	h.Spin()
}


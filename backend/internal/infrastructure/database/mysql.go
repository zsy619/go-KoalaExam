package database

import (
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/your-team/koala-exam-backend/internal/domain/entity"
	"github.com/your-team/koala-exam-backend/pkg/config"
)

// InitMySQL 初始化 MySQL
func InitMySQL(cfg config.MySQLConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database,
		cfg.Charset, cfg.ParseTime, cfg.Loc,
	)

	var dialector gorm.Dialector
	if cfg.Driver == "postgres" {
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
			cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database)
		dialector = postgres.Open(dsn)
	} else {
		dialector = mysql.Open(dsn)
	}

	var lvl logger.LogLevel
	if cfg.LogLevel == "debug" {
		lvl = logger.Info
	} else {
		lvl = logger.Warn
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(lvl),
	})
	if err != nil {
		return nil, fmt.Errorf("gorm open failed: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)
	sqlDB.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTime) * time.Second)

	hlog.Infof("MySQL connected: %s:%d/%s", cfg.Host, cfg.Port, cfg.Database)
	return db, nil
}

// AutoMigrate 自动迁移表结构（开发环境）
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&entity.Department{},
		&entity.Class{},
		&entity.User{},
		&entity.QuestionCategory{},
		&entity.Question{},
		&entity.Paper{},
		&entity.PaperQuestion{},
		&entity.Exam{},
		&entity.ExamRecord{},
		&entity.FavoriteFolder{},
		&entity.Favorite{},
		&entity.WrongAnswerLog{},
	)
}

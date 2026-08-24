package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"github.com/your-team/koala-exam-backend/internal/infrastructure/database"
	"github.com/your-team/koala-exam-backend/pkg/config"
)

// 数据库迁移 CLI：up / down / reset / seed
func main() {
	op := flag.String("op", "up", "up | down | reset | seed | fresh")
	env := flag.String("env", "dev", "dev | prod")
	flag.Parse()

	cfg, err := config.LoadConfigByEnv(*env)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// 自动创建数据库（如果不存在）
	if err := ensureDatabase(cfg.MySQL); err != nil {
		log.Fatalf("ensure db: %v", err)
	}

	db, err := database.InitMySQL(cfg.MySQL)
	if err != nil {
		log.Fatalf("init mysql: %v", err)
	}

	switch *op {
	case "up":
		if err := database.AutoMigrate(db); err != nil {
			log.Fatalf("migrate up: %v", err)
		}
		fmt.Println("✓ migrate up done")
	case "down":
		fmt.Println("⚠ drop all tables")
		dropAll(db)
		fmt.Println("✓ migrate down done")
	case "reset":
		dropAll(db)
		if err := database.AutoMigrate(db); err != nil {
			log.Fatal(err)
		}
		fmt.Println("✓ migrate reset done")
	case "seed":
		if err := database.SeedDev(db); err != nil {
			log.Fatal(err)
		}
		fmt.Println("✓ seed done")
	case "fresh":
		dropAll(db)
		if err := database.AutoMigrate(db); err != nil {
			log.Fatal(err)
		}
		if err := database.SeedDev(db); err != nil {
			log.Fatal(err)
		}
		fmt.Println("✓ fresh done (drop + migrate + seed)")
	default:
		fmt.Println("unknown op:", *op)
		os.Exit(1)
	}
}

// ensureDatabase 自动创建数据库
func ensureDatabase(cfg config.MySQLConfig) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8mb4&parseTime=true&loc=Local",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port)
	sqldb, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	defer sqldb.Close()
	sqldb.SetConnMaxLifetime(time.Minute * 3)
	sqldb.SetMaxOpenConns(10)
	if err := sqldb.Ping(); err != nil {
		return err
	}
	_, err = sqldb.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci", cfg.Database))
	return err
}

// dropAll 删除所有 ke_ 前缀表
func dropAll(db *gorm.DB) {
	_ = db.Exec("SET FOREIGN_KEY_CHECKS = 0").Error
	tables := []string{
		"ke_wrong_log", "ke_favorite", "ke_favorite_folder",
		"ke_exam_record", "ke_exam", "ke_paper_question", "ke_paper",
		"ke_question", "ke_question_category", "ke_user",
		"ke_class", "ke_department",
	}
	for _, t := range tables {
		if err := db.Exec("DROP TABLE IF EXISTS " + t).Error; err != nil {
			log.Printf("drop %s failed: %v", t, err)
		}
	}
	_ = db.Exec("SET FOREIGN_KEY_CHECKS = 1").Error
}

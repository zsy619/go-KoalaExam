package entity

import (
	"time"

	"gorm.io/gorm"
)

// Favorite 收藏主表（支持题目/试卷/知识点多态）
type Favorite struct {
	ID         int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     int64          `gorm:"index:idx_user_target,unique;not null" json:"user_id"`
	TargetType int8           `gorm:"index:idx_user_target,unique;not null" json:"target_type"` // 1:题目 2:试卷 3:知识点
	TargetID   int64          `gorm:"index:idx_user_target,unique;not null" json:"target_id"`
	FolderID   *int64         `gorm:"index" json:"folder_id"`
	SourceType int8           `gorm:"default:1" json:"source_type"` // 1:主动收藏 2:错题自动 3:智能推荐
	Difficulty int8           `gorm:"default:2" json:"difficulty"`  // 收藏时的难度
	Note       string         `gorm:"size:500" json:"note"`         // 用户备注
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (f *Favorite) TableName() string { return "ke_favorite" }

// FavoriteFolder 收藏夹/错题本分组
type FavoriteFolder struct {
	ID          int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      int64          `gorm:"index;not null" json:"user_id"`
	Name        string         `gorm:"size:64;not null" json:"name"`
	IsSystem    bool           `gorm:"default:false" json:"is_system"` // true:系统自动生成的错题本
	Color       string         `gorm:"size:16" json:"color"`
	Icon        string         `gorm:"size:64" json:"icon"`
	QuestionCnt int            `gorm:"default:0" json:"question_cnt"` // 缓存题目数量
	Sort        int            `gorm:"default:0" json:"sort"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ff *FavoriteFolder) TableName() string { return "ke_favorite_folder" }

// WrongAnswerLog 错题追踪（每一次错误）
type WrongAnswerLog struct {
	ID            int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID        int64          `gorm:"index:idx_user_q;not null" json:"user_id"`
	QuestionID    int64          `gorm:"index:idx_user_q;not null" json:"question_id"`
	ExamID        int64          `gorm:"index" json:"exam_id"`
	ExamRecordID  int64          `gorm:"index" json:"exam_record_id"`
	UserAnswer    string         `gorm:"type:text" json:"user_answer"`
	CorrectAnswer string         `gorm:"type:text" json:"correct_answer"`
	WrongCount    int            `gorm:"default:1" json:"wrong_count"` // 累计错次
	LastWrongAt   time.Time      `gorm:"index" json:"last_wrong_at"`
	IsReviewed    bool           `gorm:"default:false" json:"is_reviewed"`
	MasteryLevel  int8           `gorm:"default:1;index" json:"mastery_level"` // 1-5 掌握度
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (w *WrongAnswerLog) TableName() string { return "ke_wrong_log" }

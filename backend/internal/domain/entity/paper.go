package entity

import (
	"time"

	"gorm.io/gorm"
)

// Paper 试卷
type Paper struct {
	ID          int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Title       string         `gorm:"size:128;not null" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	Strategy    int8           `gorm:"default:1" json:"strategy"` // 1:固定 2:随机 3:遗传算法
	TotalScore  float64        `gorm:"default:100" json:"total_score"`
	Duration    int            `gorm:"default:60" json:"duration"` // 考试时长（分钟）
	PassScore   float64        `gorm:"default:60" json:"pass_score"`
	Status      int8           `gorm:"default:1;index" json:"status"` // 0:草稿 1:已发布 2:归档
	CreatorID   int64          `gorm:"index" json:"creator_id"`
	ConfigRule  string         `gorm:"type:json" json:"config_rule"`  // 随机/遗传算法配置（JSON）
	QuestionIDs string         `gorm:"type:json" json:"question_ids"` // 固定策略的题目ID列表
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (p *Paper) TableName() string { return "ke_paper" }

// PaperQuestion 试卷-题目关联（含小题分值、排序）
type PaperQuestion struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	PaperID    int64     `gorm:"index:idx_paper_q,unique;not null" json:"paper_id"`
	QuestionID int64     `gorm:"index:idx_paper_q,unique;not null" json:"question_id"`
	Type       int8      `gorm:"index" json:"type"`      // 题型（冗余便于查询）
	Score      float64   `gorm:"default:1" json:"score"` // 小题分值
	Sort       int       `gorm:"default:0" json:"sort"`
	Section    string    `gorm:"size:32" json:"section"` // 大题标题（如 一、选择题）
	CreatedAt  time.Time `json:"created_at"`
}

func (pq *PaperQuestion) TableName() string { return "ke_paper_question" }

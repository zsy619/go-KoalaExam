package entity

import (
	"time"

	"gorm.io/gorm"
)

// Question 试题实体（支持 6 种题型）
type Question struct {
	ID           int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	CategoryID   int64          `gorm:"index;not null" json:"category_id"` // 分类ID
	Type         int8           `gorm:"index;not null" json:"type"`        // 1:单选 2:多选 3:判断 4:填空 5:不定项 6:编程
	Difficulty   int8           `gorm:"default:2;index" json:"difficulty"` // 1:易 2:中 3:难
	Title        string         `gorm:"type:text;not null" json:"title"`   // 题干（富文本/Markdown）
	Options      string         `gorm:"type:json" json:"options"`          // 选项 JSON: [{"key":"A","text":"..."}]
	Answer       string         `gorm:"type:text;not null" json:"-"`       // 正确答案（JSON 数组）
	Analysis     string         `gorm:"type:text" json:"analysis"`         // 解析
	Tags         string         `gorm:"size:255;index" json:"tags"`        // 标签逗号分隔
	Score        float64        `gorm:"default:1" json:"score"`            // 默认分值
	CreatorID    int64          `gorm:"index" json:"creator_id"`
	Status       int8           `gorm:"default:1;index" json:"status"` // 0:草稿 1:已发布 2:归档
	UseCount     int64          `gorm:"default:0" json:"use_count"`    // 被引用次数
	CorrectCount int64          `gorm:"default:0" json:"correct_count"`
	WrongCount   int64          `gorm:"default:0" json:"wrong_count"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (q *Question) TableName() string { return "ke_question" }

// QuestionCategory 题库分类
type QuestionCategory struct {
	ID        int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	ParentID  int64          `gorm:"default:0;index" json:"parent_id"`
	Name      string         `gorm:"size:64;not null" json:"name"`
	Code      string         `gorm:"size:64" json:"code"`
	Sort      int            `gorm:"default:0" json:"sort"`
	CreatorID int64          `gorm:"index" json:"creator_id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (qc *QuestionCategory) TableName() string { return "ke_question_category" }

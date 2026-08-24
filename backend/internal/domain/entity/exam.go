package entity

import (
	"time"

	"gorm.io/gorm"
)

// Exam 考试（一次开考）
type Exam struct {
	ID            int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Title         string         `gorm:"size:128;not null" json:"title"`
	Description   string         `gorm:"type:text" json:"description"`
	PaperID       int64          `gorm:"index;not null" json:"paper_id"`
	StartTime     time.Time      `gorm:"index" json:"start_time"`
	EndTime       time.Time      `gorm:"index" json:"end_time"`
	Duration      int            `gorm:"default:60" json:"duration"`    // 考试时长（分钟）
	MaxAttempts   int            `gorm:"default:1" json:"max_attempts"` // 允许重考次数
	ShuffleQ      bool           `gorm:"default:true" json:"shuffle_q"`
	ShuffleOpt    bool           `gorm:"default:true" json:"shuffle_opt"`
	AntiCheat     bool           `gorm:"default:true" json:"anti_cheat"`
	Status        int8           `gorm:"default:1;index" json:"status"` // 0:未发布 1:进行中 2:已结束
	CreatorID     int64          `gorm:"index" json:"creator_id"`
	TargetUsers   string         `gorm:"type:text" json:"target_users"`   // 目标用户JSON
	TargetClasses string         `gorm:"type:text" json:"target_classes"` // 目标班级JSON
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (e *Exam) TableName() string { return "ke_exam" }

// ExamRecord 考试记录（一次考生×考试）
type ExamRecord struct {
	ID              int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	ExamID          int64          `gorm:"index:idx_exam_user,unique;not null" json:"exam_id"`
	UserID          int64          `gorm:"index:idx_exam_user,unique;not null" json:"user_id"`
	PaperSnapshot   string         `gorm:"type:json" json:"paper_snapshot"` // 试卷快照（防止题目被删影响）
	Answers         string         `gorm:"type:json" json:"answers"`        // 答题内容JSON
	AuditSummary    string         `gorm:"type:json" json:"audit_summary"`   // 防作弊审计汇总
	Status          int8           `gorm:"default:0;index" json:"status"`   // 0:进行中 1:已交卷 2:已批改 3:异常
	StartTime       time.Time      `json:"start_time"`
	SubmitTime      *time.Time     `json:"submit_time"`
	Duration        int            `gorm:"default:0" json:"duration"` // 实际用时（秒）
	TotalScore      float64        `gorm:"default:0" json:"total_score"`
	ObjectiveScore  float64        `gorm:"default:0" json:"objective_score"`
	SubjectiveScore float64        `gorm:"default:0" json:"subjective_score"`
	Passed          bool           `gorm:"default:false" json:"passed"`
	ScoreHash       string         `gorm:"size:128" json:"score_hash"` // SHA-256 签名防篡改
	TabSwitchCnt    int            `gorm:"default:0" json:"tab_switch_cnt"`
	AuditLog        string         `gorm:"type:text" json:"audit_log"`       // 行为审计 JSON
	SubjectiveDetail string         `gorm:"type:json" json:"subjective_detail"` // 主观题批改详情 JSON: [{qid, score, comment, grader, graded_at}]
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (er *ExamRecord) TableName() string { return "ke_exam_record" }

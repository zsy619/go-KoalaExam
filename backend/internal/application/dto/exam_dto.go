package dto

import (
	"errors"
	"time"
)

// CreateExamReq 创建/更新考试请求。
//
// 字段使用 RFC3339 时间字符串，便于跨平台序列化；
// 校验在 Validate() 方法中完成（避免 Hertz binding tag 与业务校验耦合）。
type CreateExamReq struct {
	Title         string  `json:"title"`
	Description   string  `json:"description"`
	PaperID       int64   `json:"paper_id"`
	StartTime     string  `json:"start_time"`
	EndTime       string  `json:"end_time"`
	Duration      int     `json:"duration"`
	MaxAttempts   int     `json:"max_attempts"`
	ShuffleQ      bool    `json:"shuffle_q"`
	ShuffleOpt    bool    `json:"shuffle_opt"`
	AntiCheat     bool    `json:"anti_cheat"`
	TargetUsers   []int64 `json:"target_users"`
	TargetClasses []int64 `json:"target_classes"`
}

// StartEndTime 解析后的时间。
type StartEndTime struct {
	Start time.Time
	End   time.Time
}

// Parse 解析时间字符串，返回 time.Time。
func (r *CreateExamReq) Parse() (StartEndTime, error) {
	start, err1 := time.Parse(time.RFC3339, r.StartTime)
	end, err2 := time.Parse(time.RFC3339, r.EndTime)
	if err1 != nil || err2 != nil {
		return StartEndTime{}, errors.New("时间格式错误（需要 RFC3339）")
	}
	if !end.After(start) {
		return StartEndTime{}, errors.New("结束时间必须晚于开始时间")
	}
	if r.Duration <= 0 {
		return StartEndTime{}, errors.New("考试时长必须 > 0")
	}
	if r.PaperID <= 0 {
		return StartEndTime{}, errors.New("必须选择试卷")
	}
	return StartEndTime{Start: start, End: end}, nil
}

// Validate 校验请求。
func (r *CreateExamReq) Validate() error {
	if _, err := r.Parse(); err != nil {
		return err
	}
	return nil
}

// StartExamResp 开始考试响应（包含试卷内容）。
type StartExamResp struct {
	ExamID     int64          `json:"exam_id"`
	RecordID   int64          `json:"record_id"`
	Title      string         `json:"title"`
	Duration   int            `json:"duration"`
	Questions  []QuestionResp `json:"questions"`
	StartTime  string         `json:"start_time"`
	EndTime    string         `json:"end_time"`
	ShuffleQ   bool           `json:"shuffle_q"`
	ShuffleOpt bool           `json:"shuffle_opt"`
}

// SaveAnswerReq 保存单题答案（每 10s 同步一次）。
type SaveAnswerReq struct {
	RecordID   int64       `json:"record_id"`
	QuestionID int64       `json:"question_id"`
	Answer     interface{} `json:"answer"`
	Elapsed    int         `json:"elapsed"`
 // 已用秒
}

// Validate 校验保存答案请求。
func (r *SaveAnswerReq) Validate() error {
	if r.RecordID <= 0 || r.QuestionID <= 0 {
		return errors.New("record_id 和 question_id 必须 > 0")
	}
	return nil
}

// SubmitExamReq 交卷请求。
type SubmitExamReq struct {
	RecordID int64 `json:"record_id"`
	Force    bool  `json:"force"`
}

// AuditReq 行为审计请求（切屏等）。
type AuditReq struct {
	RecordID int64                  `json:"record_id"`
	Events   map[string]interface{} `json:"events"`
}

// GradeSubjectiveReq 主观题评分请求。
type GradeSubjectiveReq struct {
	RecordID   int64   `json:"record_id"`
	QuestionID int64   `json:"question_id"`
	Score      float64 `json:"score"`
	Comment    string  `json:"comment"`
}

// Validate 校验主观题评分。
func (r *GradeSubjectiveReq) Validate() error {
	if r.RecordID <= 0 || r.QuestionID <= 0 {
		return errors.New("record_id 和 question_id 必须 > 0")
	}
	if r.Score < 0 {
		return errors.New("分数不能为负")
	}
	return nil
}

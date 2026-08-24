// Package valueobject 领域值对象。
//
// 值对象是不可变的、通过值相等性比较的对象。
// 用于封装业务规则（如答题答案比较、分数计算、防作弊阈值）。
package valueobject

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// ExamPaper 试卷快照值对象（不可变）。
type ExamPaper struct {
	Questions  []ExamQuestion `json:"questions"`
	TotalScore float64        `json:"total_score"`
}

// ExamQuestion 试卷中题目值对象。
type ExamQuestion struct {
	ID         int64       `json:"id"`
	Type       int8        `json:"type"`
	Difficulty int8        `json:"difficulty"`
	Title      string      `json:"title"`
	Options    interface{} `json:"options,omitempty"`
	Score      float64     `json:"score"`
}

// NewExamPaper 构造试卷值对象。
func NewExamPaper(questions []ExamQuestion) ExamPaper {
	total := 0.0
	for _, q := range questions {
		total += q.Score
	}
	return ExamPaper{Questions: questions, TotalScore: total}
}

// Marshal 序列化。
func (p ExamPaper) Marshal() string {
	data, _ := json.Marshal(p)
	return string(data)
}

// UnmarshalExamPaper 反序列化。
func UnmarshalExamPaper(s string) (ExamPaper, error) {
	var p ExamPaper
	if s == "" {
		return p, nil
	}
	err := json.Unmarshal([]byte(s), &p)
	return p, err
}

// QuestionIDs 返回所有题目 ID。
func (p ExamPaper) QuestionIDs() []int64 {
	ids := make([]int64, 0, len(p.Questions))
	for _, q := range p.Questions {
		ids = append(ids, q.ID)
	}
	return ids
}

// QuestionCount 题目数。
func (p ExamPaper) QuestionCount() int {
	return len(p.Questions)
}

// UserAnswers 答案值对象（map[questionID]answer）。
type UserAnswers map[string]interface{}

// NewUserAnswers 构造答案值对象。
func NewUserAnswers() UserAnswers {
	return make(UserAnswers)
}

// Set 设置单题答案。
func (a UserAnswers) Set(questionID int64, answer interface{}) {
	a[fmt.Sprintf("%d", questionID)] = answer
}

// Get 获取单题答案。
func (a UserAnswers) Get(questionID int64) (interface{}, bool) {
	v, ok := a[fmt.Sprintf("%d", questionID)]
	return v, ok
}

// Marshal 序列化。
func (a UserAnswers) Marshal() string {
	if a == nil || len(a) == 0 {
		return "{}"
	}
	data, err := json.Marshal(a)
	if err != nil {
		return "{}"
	}
	return string(data)
}

// UnmarshalUserAnswers 反序列化。
func UnmarshalUserAnswers(s string) (UserAnswers, error) {
	a := NewUserAnswers()
	if s == "" {
		return a, nil
	}
	if err := json.Unmarshal([]byte(s), &a); err != nil {
		return nil, err
	}
	return a, nil
}

// ExamScore 考试分值值对象（不可变）。
type ExamScore struct {
	ObjectiveScore  float64
	SubjectiveScore float64
	PassScore       float64
}

// NewExamScore 构造分值。
func NewExamScore(objective, subjective, pass float64) ExamScore {
	return ExamScore{ObjectiveScore: objective, SubjectiveScore: subjective, PassScore: pass}
}

// Total 总分。
func (s ExamScore) Total() float64 {
	return s.ObjectiveScore + s.SubjectiveScore
}

// Passed 是否通过。
func (s ExamScore) Passed() bool {
	return s.Total() >= s.PassScore
}

// AuditSummary 防作弊审计汇总。
type AuditSummary struct {
	SwitchTabCount int       `json:"switch_tab_count"`
	FullscreenExit int       `json:"fullscreen_exit"`
	CopyCount      int       `json:"copy_count"`
	PasteCount     int       `json:"paste_count"`
	DevtoolsOpen   int       `json:"devtools_open"`
	LastAt         time.Time `json:"last_at,omitempty"`
}

// IsCheating 判断是否作弊（任一指标超阈值）。
func (a AuditSummary) IsCheating(t AntiCheatThresholds) bool {
	return a.SwitchTabCount >= t.SwitchTabMax ||
		a.FullscreenExit >= t.FullscreenExitMax ||
		a.DevtoolsOpen >= t.DevtoolsOpenMax
}

// AntiCheatThresholds 防作弊阈值。
type AntiCheatThresholds struct {
	SwitchTabMax      int
	FullscreenExitMax int
	DevtoolsOpenMax   int
}

// DefaultAntiCheatThresholds 默认阈值。
func DefaultAntiCheatThresholds() AntiCheatThresholds {
	return AntiCheatThresholds{
		SwitchTabMax:      3,
		FullscreenExitMax: 2,
		DevtoolsOpenMax:   1,
	}
}

// TimeWindow 时间窗口。
type TimeWindow struct {
	Start time.Time
	End   time.Time
}

// NewTimeWindow 构造时间窗口。
func NewTimeWindow(start, end time.Time) (TimeWindow, error) {
	if !end.After(start) {
		return TimeWindow{}, errors.New("end must be after start")
	}
	return TimeWindow{Start: start, End: end}, nil
}

// Contains 是否包含给定时间。
func (w TimeWindow) Contains(t time.Time) bool {
	return !t.Before(w.Start) && !t.After(w.End)
}

// IsUpcoming 是否尚未开始。
func (w TimeWindow) IsUpcoming() bool {
	return time.Now().Before(w.Start)
}

// IsClosed 是否已结束。
func (w TimeWindow) IsClosed() bool {
	return time.Now().After(w.End)
}

// Duration 持续时间。
func (w TimeWindow) Duration() time.Duration {
	return w.End.Sub(w.Start)
}

// Password 密码值对象。
type Password string

// NewPassword 构造密码，校验强度。
func NewPassword(raw string) (Password, error) {
	if len(raw) < 8 {
		return "", errors.New("password must be at least 8 characters")
	}
	if len(raw) > 64 {
		return "", errors.New("password too long")
	}
	matched := 0
	if regexp.MustCompile(`[a-z]`).MatchString(raw) {
		matched++
	}
	if regexp.MustCompile(`[A-Z]`).MatchString(raw) {
		matched++
	}
	if regexp.MustCompile(`[0-9]`).MatchString(raw) {
		matched++
	}
	if regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>/?]`).MatchString(raw) {
		matched++
	}
	if matched < 2 {
		return "", errors.New("password must contain at least 2 of: lowercase, uppercase, digit, special")
	}
	return Password(raw), nil
}

// String 实现 Stringer。
func (p Password) String() string { return string(p) }

// Strength 密码强度（弱/中/强）。
func (p Password) Strength() string {
	s := string(p)
	if len(s) >= 12 && strings.ContainsAny(s, "!@#$%^&*") {
		return "strong"
	}
	if len(s) >= 8 {
		return "medium"
	}
	return "weak"
}

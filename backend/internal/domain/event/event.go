// Package event 领域事件。
//
// 领域事件代表已经发生的事实，用于解耦应用层副作用（如通知、统计、审计）。
//
// 命名约定：动名词过去式（StartedExam / SubmittedExam / RecordedWrongAnswer）。
package event

import (
	"context"
	"time"
)

// Event 领域事件基础接口。
type Event interface {
	EventName() string
	OccurredAt() time.Time
}

// Handler 事件处理器。
type Handler interface {
	Handle(ctx context.Context, e Event) error
}

// HandlerFunc 函数式处理器。
type HandlerFunc func(ctx context.Context, e Event) error

// Handle 实现 Handler 接口。
func (f HandlerFunc) Handle(ctx context.Context, e Event) error { return f(ctx, e) }

// Bus 事件总线（进程内 pub-sub）。
type Bus struct {
	handlers map[string][]Handler
}

// NewBus 构造事件总线。
func NewBus() *Bus {
	return &Bus{handlers: make(map[string][]Handler)}
}

// Subscribe 订阅事件。
func (b *Bus) Subscribe(eventName string, h Handler) {
	b.handlers[eventName] = append(b.handlers[eventName], h)
}

// Publish 同步发布事件（错误聚合返回）。
func (b *Bus) Publish(ctx context.Context, events ...Event) error {
	for _, e := range events {
		for _, h := range b.handlers[e.EventName()] {
			if err := h.Handle(ctx, e); err != nil {
				return err
			}
		}
	}
	return nil
}

// ============================================================
// 考试相关事件
// ============================================================

// ExamStartedEvent 考试开始事件。
type ExamStartedEvent struct {
	UserID    int64
	ExamID    int64
	RecordID  int64
	StartedAt time.Time
}

// EventName 实现 Event 接口。
func (e *ExamStartedEvent) EventName() string { return "exam.started" }

// OccurredAt 实现 Event 接口。
func (e *ExamStartedEvent) OccurredAt() time.Time { return e.StartedAt }

// ExamSubmittedEvent 考试提交事件。
type ExamSubmittedEvent struct {
	UserID     int64
	ExamID     int64
	RecordID   int64
	Score      float64
	IsAbnormal bool
	SubmittedAt time.Time
}

// EventName 实现 Event 接口。
func (e *ExamSubmittedEvent) EventName() string { return "exam.submitted" }

// OccurredAt 实现 Event 接口。
func (e *ExamSubmittedEvent) OccurredAt() time.Time { return e.SubmittedAt }

// ExamCheatedEvent 考试作弊事件（审计告警）。
type ExamCheatedEvent struct {
	UserID    int64
	ExamID    int64
	RecordID  int64
	EventType string
	Count     int
	DetectedAt time.Time
}

// EventName 实现 Event 接口。
func (e *ExamCheatedEvent) EventName() string { return "exam.cheated" }

// OccurredAt 实现 Event 接口。
func (e *ExamCheatedEvent) OccurredAt() time.Time { return e.DetectedAt }

// ============================================================
// 收藏/错题相关事件
// ============================================================

// FavoriteToggledEvent 收藏切换事件。
type FavoriteToggledEvent struct {
	UserID     int64
	TargetType int8
	TargetID   int64
	Added      bool
	ToggledAt  time.Time
}

// EventName 实现 Event 接口。
func (e *FavoriteToggledEvent) EventName() string { return "favorite.toggled" }

// OccurredAt 实现 Event 接口。
func (e *FavoriteToggledEvent) OccurredAt() time.Time { return e.ToggledAt }

// WrongAnswerRecordedEvent 错题记录事件。
type WrongAnswerRecordedEvent struct {
	UserID     int64
	QuestionID int64
	WrongCount int
	RecordedAt time.Time
}

// EventName 实现 Event 接口。
func (e *WrongAnswerRecordedEvent) EventName() string { return "wrong.recorded" }

// OccurredAt 实现 Event 接口。
func (e *WrongAnswerRecordedEvent) OccurredAt() time.Time { return e.RecordedAt }

// WrongBookReviewedEvent 错题已复习事件。
type WrongBookReviewedEvent struct {
	UserID       int64
	LogID        int64
	MasteryLevel int8
	ReviewedAt   time.Time
}

// EventName 实现 Event 接口。
func (e *WrongBookReviewedEvent) EventName() string { return "wrong.reviewed" }

// OccurredAt 实现 Event 接口。
func (e *WrongBookReviewedEvent) OccurredAt() time.Time { return e.ReviewedAt }


// ============================================================
// 用户认证事件
// ============================================================

// UserLoggedOutEvent 用户登出事件。
type UserLoggedOutEvent struct {
	UserID   int64
	Token    string
	LogOutAt time.Time
}

// EventName 实现 Event 接口。
func (e *UserLoggedOutEvent) EventName() string { return "user.logged_out" }

// OccurredAt 实现 Event 接口。
func (e *UserLoggedOutEvent) OccurredAt() time.Time { return e.LogOutAt }

// UserLoginFailedEvent 用户登录失败事件（用于风控统计）。
type UserLoginFailedEvent struct {
	Username  string
	IP        string
	Reason    string
	FailedAt  time.Time
}

// EventName 实现 Event 接口。
func (e *UserLoginFailedEvent) EventName() string { return "user.login_failed" }

// OccurredAt 实现 Event 接口。
func (e *UserLoginFailedEvent) OccurredAt() time.Time { return e.FailedAt }

// UserCreatedEvent 用户创建事件。
type UserCreatedEvent struct {
	UserID    int64
	Username  string
	Role      int8
	CreatedAt time.Time
}

// EventName 实现 Event 接口。
func (e *UserCreatedEvent) EventName() string { return "user.created" }

// OccurredAt 实现 Event 接口。
func (e *UserCreatedEvent) OccurredAt() time.Time { return e.CreatedAt }

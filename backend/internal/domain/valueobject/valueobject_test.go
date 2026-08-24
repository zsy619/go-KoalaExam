package valueobject

import (
	"testing"
	"time"
)

func TestNewUserAnswers(t *testing.T) {
	a := NewUserAnswers()
	a.Set(1, "A")
	a.Set(2, []string{"B", "C"})

	if v, ok := a.Get(1); !ok || v != "A" {
		t.Errorf("Get(1) = %v, %v", v, ok)
	}
	if v, ok := a.Get(2); !ok || v == nil {
		t.Errorf("Get(2) = %v, %v", v, ok)
	}
	if _, ok := a.Get(99); ok {
		t.Error("Get(99) should not exist")
	}
}

func TestUserAnswers_Marshal(t *testing.T) {
	empty := NewUserAnswers()
	if empty.Marshal() != "{}" {
		t.Errorf("expected empty to marshal as {}, got %s", empty.Marshal())
	}

	a := NewUserAnswers()
	a.Set(1, "A")
	if a.Marshal() == "" || a.Marshal() == "{}" {
		t.Error("non-empty answers should marshal to non-empty JSON")
	}
}

func TestNewExamPaper(t *testing.T) {
	paper := NewExamPaper([]ExamQuestion{
		{ID: 1, Score: 10},
		{ID: 2, Score: 20},
		{ID: 3, Score: 30},
	})
	if paper.TotalScore != 60 {
		t.Errorf("expected total 60, got %v", paper.TotalScore)
	}
	if paper.QuestionCount() != 3 {
		t.Errorf("expected 3 questions, got %d", paper.QuestionCount())
	}
	ids := paper.QuestionIDs()
	if len(ids) != 3 || ids[0] != 1 || ids[2] != 3 {
		t.Errorf("unexpected IDs: %v", ids)
	}
}

func TestExamScore(t *testing.T) {
	score := NewExamScore(40, 30, 60)
	if score.Total() != 70 {
		t.Errorf("expected 70, got %v", score.Total())
	}
	if !score.Passed() {
		t.Error("70 >= 60, should be passed")
	}

	failed := NewExamScore(20, 20, 60)
	if failed.Passed() {
		t.Error("40 < 60, should be failed")
	}
}

func TestAntiCheatThresholds(t *testing.T) {
	t1 := DefaultAntiCheatThresholds()

	a := AuditSummary{SwitchTabCount: 3}
	if !a.IsCheating(t1) {
		t.Error("3 switches should trigger cheating")
	}

	a2 := AuditSummary{SwitchTabCount: 2}
	if a2.IsCheating(t1) {
		t.Error("2 switches should NOT trigger (default threshold is 3)")
	}

	a3 := AuditSummary{DevtoolsOpen: 1}
	if !a3.IsCheating(t1) {
		t.Error("1 devtools open should trigger")
	}
}

func TestTimeWindow(t *testing.T) {
	start := time.Now()
	end := start.Add(time.Hour)
	w, err := NewTimeWindow(start, end)
	if err != nil {
		t.Fatalf("NewTimeWindow error: %v", err)
	}

	if !w.Contains(start.Add(30 * time.Minute)) {
		t.Error("window should contain middle time")
	}
	if w.Contains(start.Add(-time.Minute)) {
		t.Error("window should not contain before-start time")
	}
	if w.Contains(end.Add(time.Minute)) {
		t.Error("window should not contain after-end time")
	}

	// 倒序时间应报错
	_, err = NewTimeWindow(end, start)
	if err == nil {
		t.Error("expected error for inverted window")
	}
}

func TestPassword(t *testing.T) {
	tests := []struct {
		raw     string
		wantErr bool
	}{
		{"abc", true},           // 太短
		{"abcd1234", false},     // 字母+数字
		{"Abcd1234", false},     // 大小写+数字
		{"Abcd1234!", false},    // 包含特殊字符
		{"abcdefgh", true},      // 只有小写
		{"12345678", true},      // 只有数字
		{string(make([]byte, 65)), true}, // 太长
	}
	for _, tt := range tests {
		_, err := NewPassword(tt.raw)
		if (err != nil) != tt.wantErr {
			t.Errorf("NewPassword(%q) error = %v, wantErr %v", tt.raw, err, tt.wantErr)
		}
	}
}

func TestPassword_Strength(t *testing.T) {
	tests := []struct {
		raw  string
		want string
	}{
		{"Abcd1234", "medium"},
		{"Abcd1234!@#$", "strong"},
		{"abcd123", "weak"},
	}
	for _, tt := range tests {
		p, _ := NewPassword(tt.raw)
		if p == "" {
			continue
		}
		if got := p.Strength(); got != tt.want {
			t.Errorf("Password(%q).Strength() = %v, want %v", tt.raw, got, tt.want)
		}
	}
}

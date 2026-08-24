package service

import (
	"encoding/json"
	"testing"

	"github.com/your-team/koala-exam-backend/internal/domain/entity"
)

func TestAnswerComparator_Strings(t *testing.T) {
	c := NewAnswerComparator()
	tests := []struct {
		name    string
		correct interface{}
		user    interface{}
		want    bool
	}{
		{"exact match", "A", "A", true},
		{"case insensitive", "A", "a", true},
		{"with spaces", " A ", "a", true},
		{"different", "A", "B", false},
		{"empty", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.Compare(tt.correct, tt.user)
			if got != tt.want {
				t.Errorf("Compare(%v, %v) = %v, want %v", tt.correct, tt.user, got, tt.want)
			}
		})
	}
}

func TestAnswerComparator_Slices(t *testing.T) {
	c := NewAnswerComparator()
	correct := []interface{}{"A", "B"}
	user1 := []interface{}{"B", "A"}   // 顺序不同
	user2 := []interface{}{"A", "C"}   // 内容不同
	user3 := []interface{}{"A", "B", "C"} // 长度不同

	if !c.Compare(correct, user1) {
		t.Error("Compare with different order should still match")
	}
	if c.Compare(correct, user2) {
		t.Error("Compare with different content should not match")
	}
	if c.Compare(correct, user3) {
		t.Error("Compare with different length should not match")
	}
}

func TestGradingStrategy_GradeObjective(t *testing.T) {
	g := NewGradingStrategy()
	ansJSON, _ := json.Marshal("A")
	q := entity.Question{
		ID:    1,
		Type:  1, // 单选
		Score: 5,
		Answer: string(ansJSON),
	}

	// 正确答案
	score, ok := g.GradeObjective(q, "A")
	if !ok || score != 5 {
		t.Errorf("GradeObjective(correct) = (%v, %v), want (5, true)", score, ok)
	}

	// 错误答案
	score, ok = g.GradeObjective(q, "B")
	if ok || score != 0 {
		t.Errorf("GradeObjective(wrong) = (%v, %v), want (0, false)", score, ok)
	}
}

func TestQuestionSelector_SelectFixed(t *testing.T) {
	s := NewQuestionSelector()
	all := []entity.Question{
		{ID: 1, Title: "Q1"},
		{ID: 2, Title: "Q2"},
		{ID: 3, Title: "Q3"},
	}
	picked, err := s.SelectFixed(all, []int64{3, 1})
	if err != nil {
		t.Fatalf("SelectFixed error: %v", err)
	}
	if len(picked) != 2 {
		t.Fatalf("expected 2 questions, got %d", len(picked))
	}
	if picked[0].ID != 3 || picked[1].ID != 1 {
		t.Error("order is not preserved")
	}
}

func TestQuestionSelector_SelectFixed_NotFound(t *testing.T) {
	s := NewQuestionSelector()
	all := []entity.Question{{ID: 1}}
	_, err := s.SelectFixed(all, []int64{99})
	if err == nil {
		t.Error("expected error for missing question")
	}
}

func TestScoreSigner(t *testing.T) {
	s := NewScoreSigner("test-secret")
	sig1 := s.Sign(1, 2, 3, 80.5)
	sig2 := s.Sign(1, 2, 3, 80.5)
	if sig1 == sig2 {
		t.Error("signatures should differ (includes random component)")
	}
	if len(sig1) != 64 {
		t.Errorf("expected 64-char hex SHA-256, got %d", len(sig1))
	}
}

func TestScoreCalculator(t *testing.T) {
	calc := &ScoreCalculator{}
	score := calc.Calculate(40, 30, 60)
	if score.Total() != 70 {
		t.Errorf("expected total 70, got %v", score.Total())
	}
	if !score.Passed() {
		t.Error("expected passed (70 >= 60)")
	}

	failed := calc.Calculate(20, 20, 60)
	if failed.Passed() {
		t.Error("expected failed (40 < 60)")
	}
}

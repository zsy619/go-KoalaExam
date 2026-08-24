package entity_test

import (
	"testing"

	"github.com/your-team/koala-exam-backend/internal/domain/entity"
)

func TestUserTableName(t *testing.T) {
	u := &entity.User{}
	if u.TableName() != "ke_user" { t.Fatalf("expected ke_user, got %s", u.TableName()) }
}

func TestAllTableNames(t *testing.T) {
	cases := []struct {
		name string
		got  string
		want string
	}{
		{"user", (&entity.User{}).TableName(), "ke_user"},
		{"class", (&entity.Class{}).TableName(), "ke_class"},
		{"department", (&entity.Department{}).TableName(), "ke_department"},
		{"question_category", (&entity.QuestionCategory{}).TableName(), "ke_question_category"},
		{"question", (&entity.Question{}).TableName(), "ke_question"},
		{"paper", (&entity.Paper{}).TableName(), "ke_paper"},
		{"paper_question", (&entity.PaperQuestion{}).TableName(), "ke_paper_question"},
		{"exam", (&entity.Exam{}).TableName(), "ke_exam"},
		{"exam_record", (&entity.ExamRecord{}).TableName(), "ke_exam_record"},
		{"favorite_folder", (&entity.FavoriteFolder{}).TableName(), "ke_favorite_folder"},
		{"favorite", (&entity.Favorite{}).TableName(), "ke_favorite"},
		{"wrong_log", (&entity.WrongAnswerLog{}).TableName(), "ke_wrong_log"},
	}
	for _, c := range cases {
		if c.got != c.want { t.Errorf("%s: got %s, want %s", c.name, c.got, c.want) }
	}
}

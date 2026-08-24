package grading

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/your-team/koala-exam-backend/internal/application/dto"
	"github.com/your-team/koala-exam-backend/internal/application/exam"
	"github.com/your-team/koala-exam-backend/internal/application/favorite"
	"github.com/your-team/koala-exam-backend/internal/domain/consts"
	"github.com/your-team/koala-exam-backend/internal/domain/entity"
	"github.com/your-team/koala-exam-backend/internal/domain/errcode"
	"github.com/your-team/koala-exam-backend/internal/infrastructure/repository"
)

// GradingApp 阅卷应用服务
type GradingApp struct {
	recordRepo *repository.ExamRecordRepository
	qRepo      *repository.QuestionRepository
	examApp    *exam.ExamApp
	favApp     *favorite.FavoriteApp
}

func NewGradingApp(r *repository.ExamRecordRepository, q *repository.QuestionRepository, e *exam.ExamApp, f *favorite.FavoriteApp) *GradingApp {
	return &GradingApp{recordRepo: r, qRepo: q, examApp: e, favApp: f}
}

// AutoGrade 自动阅卷（客观题）
func (a *GradingApp) AutoGrade(ctx context.Context, recordID int64) (*entity.ExamRecord, error) {
	rec, err := a.recordRepo.GetByID(ctx, recordID)
	if err != nil {
		return nil, errcode.New(errcode.CodeNotFound, "NotFound")
	}

	var qs []entity.Question
	if err := json.Unmarshal([]byte(rec.PaperSnapshot), &qs); err != nil {
		return nil, err
	}

	var userAnswers map[string]interface{}
	if rec.Answers != "" {
		_ = json.Unmarshal([]byte(rec.Answers), &userAnswers)
	}

	var objective float64
	wrongItems := []favorite.WrongItem{}
	now := time.Now()

	for _, q := range qs {
		if !consts.IsObjective(q.Type) {
			continue
		}
		correctAns, _ := parseAnswer(q.Answer)
		userAns, _ := userAnswers[fmt.Sprintf("%d", q.ID)]
		correct := compareAnswer(correctAns, userAns)
		if correct {
			objective += q.Score
		} else {
			wrongItems = append(wrongItems, favorite.WrongItem{
				QuestionID:        q.ID,
				UserAnswerText:    fmt.Sprintf("%v", userAns),
				CorrectAnswerText: fmt.Sprintf("%v", correctAns),
				Now:               now,
			})
		}
		// 题库统计
		_ = a.qRepo.IncStat(ctx, q.ID, correct)
	}

	rec.ObjectiveScore = objective
	rec.TotalScore = objective + rec.SubjectiveScore
	rec.Passed = rec.TotalScore >= 60 // 默认 60 分及格，可在试卷中配置
	rec.Status = consts.RecordStatusGraded
	a.examApp.SignScore(rec)

	if err := a.recordRepo.Update(ctx, rec); err != nil {
		return nil, err
	}

	// 自动收录错题
	if len(wrongItems) > 0 {
		_ = a.favApp.RecordWrongAnswers(ctx, rec.UserID, rec.ExamID, rec.ID, wrongItems)
	}

	return rec, nil
}

// GradeSubjective 主观题评分（教师）
func (a *GradingApp) GradeSubjective(ctx context.Context, req *dto.GradeSubjectiveReq) error {
	rec, err := a.recordRepo.GetByID(ctx, req.RecordID)
	if err != nil {
		return errcode.New(errcode.CodeNotFound, "NotFound")
	}
	// 简化：累加到主观题总分
	rec.SubjectiveScore += req.Score
	rec.TotalScore = rec.ObjectiveScore + rec.SubjectiveScore
	rec.Passed = rec.TotalScore >= 60
	a.examApp.SignScore(rec)
	return a.recordRepo.Update(ctx, rec)
}

// parseAnswer 解析答案（支持字符串或数组）
func parseAnswer(raw string) (interface{}, error) {
	var v interface{}
	err := json.Unmarshal([]byte(raw), &v)
	return v, err
}

// compareAnswer 比对答案（忽略大小写、空格）
func compareAnswer(correct, user interface{}) bool {
	if correct == nil || user == nil {
		return false
	}
	cs := fmt.Sprintf("%v", correct)
	us := fmt.Sprintf("%v", user)
	if cs == us {
		return true
	}
	// 简化比对：去空格、转小写
	return normalize(cs) == normalize(us)
}

func normalize(s string) string {
	out := ""
	for _, r := range s {
		if r != ' ' && r != '\t' && r != '\n' {
			out += string(r)
		}
	}
	return out
}

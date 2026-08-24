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
	wrongItems := []int64{}
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
			wrongItems = append(wrongItems, q.ID)
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
		_ = a.favApp.RecordWrongAnswers(ctx, rec.UserID, wrongItems)
	}

	return rec, nil
}

// SubjectiveGradeItem 单题批改条目
type SubjectiveGradeItem struct {
	QuestionID int64   `json:"question_id"`
	Score      float64 `json:"score"`
	Comment    string  `json:"comment"`
}

// SubjectiveGradeReq 批量批改请求
type SubjectiveGradeReq struct {
	RecordID int64                `json:"record_id"`
	GraderID int64                `json:"grader_id"`
	Items    []SubjectiveGradeItem `json:"items"`
}

// GradeSubjectiveBatch 主观题批量评分（教师）
func (a *GradingApp) GradeSubjectiveBatch(ctx context.Context, req *SubjectiveGradeReq) error {
	rec, err := a.recordRepo.GetByID(ctx, req.RecordID)
	if err != nil {
		return errcode.New(errcode.CodeNotFound, "NotFound")
	}

	// 读取已有的主观题批改详情（覆盖式更新）
	var details []map[string]interface{}
	if rec.SubjectiveDetail != "" {
		_ = json.Unmarshal([]byte(rec.SubjectiveDetail), &details)
	}
	existing := make(map[int64]int) // qid -> index in details
	for i, d := range details {
		if qid, ok := d["question_id"].(float64); ok {
			existing[int64(qid)] = i
		}
	}

	now := time.Now()
	totalSubjective := 0.0
	for _, it := range req.Items {
		entry := map[string]interface{}{
			"question_id": it.QuestionID,
			"score":       it.Score,
			"comment":     it.Comment,
			"grader_id":   req.GraderID,
			"graded_at":   now,
		}
		if idx, ok := existing[it.QuestionID]; ok {
			details[idx] = entry
		} else {
			details = append(details, entry)
		}
		totalSubjective += it.Score
	}

	// 写入详情
	newJSON, _ := json.Marshal(details)
	rec.SubjectiveDetail = string(newJSON)
	rec.SubjectiveScore = totalSubjective
	rec.TotalScore = rec.ObjectiveScore + rec.SubjectiveScore
	rec.Passed = rec.TotalScore >= 60
	rec.Status = consts.RecordStatusGraded
	a.examApp.SignScore(rec)
	return a.recordRepo.Update(ctx, rec)
}

// GradeSubjective 单题批改（兼容旧版接口）
func (a *GradingApp) GradeSubjective(ctx context.Context, req *dto.GradeSubjectiveReq) error {
	return a.GradeSubjectiveBatch(ctx, &SubjectiveGradeReq{
		RecordID: req.RecordID,
		GraderID: 0,
		Items: []SubjectiveGradeItem{{
			QuestionID: req.QuestionID,
			Score:      req.Score,
			Comment:    req.Comment,
		}},
	})
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

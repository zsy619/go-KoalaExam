// Package exam 考试应用服务。
//
// 采用 Google Go 风格：
//   - 简洁的命名（StartExam、SaveAnswer 等动词+名词）
//   - 显式错误返回（error 作为最后一个返回值）
//   - context 作为第一个参数传递
//   - 通过 interface 注入依赖，便于单元测试
package exam

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/your-team/koala-exam-backend/internal/application/dto"
	"github.com/your-team/koala-exam-backend/internal/domain/consts"
	"github.com/your-team/koala-exam-backend/internal/domain/entity"
	"github.com/your-team/koala-exam-backend/internal/domain/errcode"
	"github.com/your-team/koala-exam-backend/internal/domain/repository"
	"github.com/your-team/koala-exam-backend/internal/infrastructure/cache"
	"github.com/your-team/koala-exam-backend/pkg/encrypt"
)

// PaperAssembler 试卷组装器（领域服务）。
// 由 application/question 包实现，本包通过接口依赖，避免循环依赖。
type PaperAssembler interface {
	Assemble(ctx context.Context, paperID int64) ([]entity.Question, error)
}

// ExamApp 考试应用服务。
type ExamApp struct {
	exams   repository.ExamRepository
	records repository.ExamRecordRepository
	papers  PaperAssembler
	rdb     *redis.Client
}

// NewExamApp 构造考试应用服务（依赖注入）。
func NewExamApp(
	exams repository.ExamRepository,
	records repository.ExamRecordRepository,
	papers PaperAssembler,
	rdb *redis.Client,
) *ExamApp {
	return &ExamApp{exams: exams, records: records, papers: papers, rdb: rdb}
}

// CreateExam 创建考试。
func (a *ExamApp) CreateExam(ctx context.Context, req *dto.CreateExamReq, creatorID int64) (int64, error) {
	if err := req.Validate(); err != nil {
		return 0, err
	}
	exam := buildExam(req, creatorID)
	if err := a.exams.Create(ctx, exam); err != nil {
		return 0, err
	}
	return exam.ID, nil
}

// ListExams 分页查询考试。
func (a *ExamApp) ListExams(ctx context.Context, filter repository.ExamListFilter) ([]entity.Exam, int64, error) {
	return a.exams.List(ctx, filter)
}

// ListAvailableExams 学员可参加的考试列表。
func (a *ExamApp) ListAvailableExams(ctx context.Context, uid int64) ([]entity.Exam, error) {
	return a.exams.ListAvailableForUser(ctx, uid)
}

// GetExam 查询单个考试。
func (a *ExamApp) GetExam(ctx context.Context, id int64) (*entity.Exam, error) {
	exam, err := a.exams.GetByID(ctx, id)
	if err != nil {
		return nil, errcode.New(errcode.CodeExamNotExist, "ExamNotExist")
	}
	return exam, nil
}

// UpdateExam 更新考试。
func (a *ExamApp) UpdateExam(ctx context.Context, id int64, req *dto.CreateExamReq) error {
	if err := req.Validate(); err != nil {
		return err
	}
	exam := buildExam(req, 0)
	exam.ID = id
	return a.exams.Update(ctx, exam)
}

// DeleteExam 软删除考试。
func (a *ExamApp) DeleteExam(ctx context.Context, id int64) error {
	return a.exams.Delete(ctx, id)
}

// ArchiveExam 归档考试（状态改为已结束）。
func (a *ExamApp) ArchiveExam(ctx context.Context, id int64) error {
	exam := &entity.Exam{ID: id, Status: consts.ExamStatusArchived}
	return a.exams.Update(ctx, exam)
}

// StartExam 开始考试，支持断线续考。
func (a *ExamApp) StartExam(ctx context.Context, examID, userID int64) (*dto.StartExamResp, error) {
	exam, err := a.getRunningExam(ctx, examID)
	if err != nil {
		return nil, err
	}

	record, err := a.getOrCreateRecord(ctx, exam, userID)
	if err != nil {
		return nil, err
	}

	questions, err := a.loadQuestionsForRecord(record, exam)
	if err != nil {
		return nil, err
	}

	return buildStartResp(exam, record, questions), nil
}

// SaveAnswer 实时保存单题答案到 Redis（4 小时 TTL）。
func (a *ExamApp) SaveAnswer(ctx context.Context, req *dto.SaveAnswerReq) error {
	if err := req.Validate(); err != nil {
		return err
	}
	key := cache.KeyExamProgress.Build(req.RecordID)
	data, _ := json.Marshal(map[string]interface{}{
		"qid":     req.QuestionID,
		"ans":     req.Answer,
		"elapsed": req.Elapsed,
		"ts":      time.Now().Unix(),
	})
	pipe := a.rdb.TxPipeline()
	pipe.HSet(ctx, key, fmt.Sprintf("%d", req.QuestionID), data)
	pipe.Expire(ctx, key, 4*time.Hour)
	_, err := pipe.Exec(ctx)
	return err
}

// AuditEvent 记录考试行为（切屏/复制粘贴）。
func (a *ExamApp) AuditEvent(ctx context.Context, req *dto.AuditReq) error {
	rec, err := a.records.GetByID(ctx, req.RecordID)
	if err != nil {
		return errcode.New(errcode.CodeNotFound, "RecordNotFound")
	}
	record := rec.(*entity.ExamRecord)

	if ev, ok := req.Events["type"].(string); ok && ev == "tab_switch" {
		record.TabSwitchCnt++
	}
	appendAuditLog(record, req.Events)
	return a.records.Update(ctx, record)
}

// SubmitExam 交卷：合并 Redis 数据 → 更新状态 → 触发自动批改（由 handler 编排）。
func (a *ExamApp) SubmitExam(ctx context.Context, recordID int64) (*entity.ExamRecord, error) {
	rec, err := a.records.GetByID(ctx, recordID)
	if err != nil {
		return nil, errcode.New(errcode.CodeNotFound, "RecordNotFound")
	}
	record := rec.(*entity.ExamRecord)

	if record.Status != consts.RecordStatusOngoing {
		return nil, errcode.New(errcode.CodeExamSubmitted, "ExamSubmitted")
	}

	answers, err := a.mergeRedisAnswers(ctx, recordID)
	if err != nil {
		return nil, err
	}

	record.Answers = encodeJSON(answers)
	record.SubmitTime = ptrTime(time.Now())
	record.Status = consts.RecordStatusSubmitted
	record.Duration = int(time.Since(record.StartTime).Seconds())

	if err := a.records.Update(ctx, record); err != nil {
		return nil, err
	}
	return record, nil
}

// SignScore 计算 SHA-256 分数签名（防篡改）。
func (a *ExamApp) SignScore(rec *entity.ExamRecord) {
	rec.ScoreHash = encrypt.SHA256Hex(
		fmt.Sprintf("%d|%.2f|%.2f|%.2f", rec.ID, rec.TotalScore, rec.ObjectiveScore, rec.SubjectiveScore),
		"koala-exam-salt",
	)
}

// ListRecordsByExam 查询某场考试的所有记录（分页）。
func (a *ExamApp) ListRecordsByExam(ctx context.Context, examID int64, page, size int) ([]entity.ExamRecord, int64, error) {
	return a.records.ListByExam(ctx, examID, page, size)
}

// ListRecordsByUser 查询学员的所有记录（分页）。
func (a *ExamApp) ListRecordsByUser(ctx context.Context, uid int64, page, size int) ([]entity.ExamRecord, int64, error) {
	filter := repository.ExamRecordListFilter{
		PageQuery: repository.PageQuery{Page: page, Size: size},
		UserID:    uid,
	}
	return a.records.ListByUser(ctx, filter)
}

// GetRecord 查询单条考试记录。
func (a *ExamApp) GetRecord(ctx context.Context, id int64) (*entity.ExamRecord, error) {
	rec, err := a.records.GetByID(ctx, id)
	if err != nil {
		return nil, errcode.New(errcode.CodeNotFound, "RecordNotFound")
	}
	return rec.(*entity.ExamRecord), nil
}

// ============================================================
// 私有辅助方法
// ============================================================

// getRunningExam 查询当前正在进行的考试。
func (a *ExamApp) getRunningExam(ctx context.Context, examID int64) (*entity.Exam, error) {
	exam, err := a.exams.GetByID(ctx, examID)
	if err != nil {
		return nil, errcode.New(errcode.CodeExamNotExist, "ExamNotExist")
	}
	now := time.Now()
	if now.Before(exam.StartTime) || now.After(exam.EndTime) {
		return nil, errcode.New(errcode.CodeExamNotRunning, "ExamNotRunning")
	}
	return exam, nil
}

// getOrCreateRecord 获取已有记录（断线续考）或创建新记录。
func (a *ExamApp) getOrCreateRecord(ctx context.Context, exam *entity.Exam, userID int64) (*entity.ExamRecord, error) {
	existing, err := a.records.GetByExamAndUser(ctx, exam.ID, userID)
	if err == nil && existing != nil {
		record := existing.(*entity.ExamRecord)
		if record.Status != consts.RecordStatusOngoing {
			return nil, errcode.New(errcode.CodeExamSubmitted, "ExamSubmitted")
		}
		return record, nil
	}

	questions, err := a.papers.Assemble(ctx, exam.PaperID)
	if err != nil {
		return nil, err
	}
	snapshot, _ := json.Marshal(questions)

	record := &entity.ExamRecord{
		ExamID:        exam.ID,
		UserID:        userID,
		PaperSnapshot: encodeJSON(snapshot),
		Answers:       "{}",
		Status:        consts.RecordStatusOngoing,
		StartTime:     time.Now(),
	}
	if err := a.records.Create(ctx, record); err != nil {
		return nil, err
	}
	return record, nil
}

// loadQuestionsForRecord 从快照中加载题目并打乱。
func (a *ExamApp) loadQuestionsForRecord(record *entity.ExamRecord, exam *entity.Exam) ([]entity.Question, error) {
	var questions []entity.Question
	if err := json.Unmarshal([]byte(record.PaperSnapshot), &questions); err != nil {
		return nil, err
	}
	if exam.ShuffleQ {
		rand.Shuffle(len(questions), func(i, j int) { questions[i], questions[j] = questions[j], questions[i] })
	}
	return questions, nil
}

// mergeRedisAnswers 从 Redis 读取所有已保存答案。
func (a *ExamApp) mergeRedisAnswers(ctx context.Context, recordID int64) (map[string]interface{}, error) {
	key := cache.KeyExamProgress.Build(recordID)
	hash, err := a.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	answers := make(map[string]interface{}, len(hash))
	for k, v := range hash {
		var entry map[string]interface{}
		if err := json.Unmarshal([]byte(v), &entry); err == nil {
			answers[k] = entry["ans"]
		}
	}
	return answers, nil
}

// ============================================================
// 包级辅助函数
// ============================================================

// buildExam 构造 Exam 实体（已通过 Validate）。
func buildExam(req *dto.CreateExamReq, creatorID int64) *entity.Exam {
	tu, _ := json.Marshal(req.TargetUsers)
	tc, _ := json.Marshal(req.TargetClasses)
	parsed, _ := req.Parse()
	return &entity.Exam{
		Title:         req.Title,
		Description:   req.Description,
		PaperID:       req.PaperID,
		StartTime:     parsed.Start,
		EndTime:       parsed.End,
		Duration:      req.Duration,
		MaxAttempts:   req.MaxAttempts,
		ShuffleQ:      req.ShuffleQ,
		ShuffleOpt:    req.ShuffleOpt,
		AntiCheat:     req.AntiCheat,
		Status:        consts.ExamStatusRunning,
		CreatorID:     creatorID,
		TargetUsers:   string(tu),
		TargetClasses: string(tc),
	}
}

// buildStartResp 构造开始考试响应。
func buildStartResp(exam *entity.Exam, record *entity.ExamRecord, questions []entity.Question) *dto.StartExamResp {
	return &dto.StartExamResp{
		ExamID:    exam.ID,
		RecordID:  record.ID,
		Title:     exam.Title,
		Duration:  exam.Duration,
		Questions: questions,
		StartTime: exam.StartTime.Format(time.RFC3339),
		EndTime:   exam.EndTime.Format(time.RFC3339),
		ShuffleQ:  exam.ShuffleQ,
		ShuffleOpt: exam.ShuffleOpt,
	}
}

// appendAuditLog 追加行为审计日志。
func appendAuditLog(record *entity.ExamRecord, events map[string]interface{}) {
	var audit []map[string]interface{}
	if record.AuditLog != "" {
		_ = json.Unmarshal([]byte(record.AuditLog), &audit)
	}
	audit = append(audit, events)
	data, _ := json.Marshal(audit)
	record.AuditLog = string(data)
}

// encodeJSON 将任意类型转为 JSON 字符串，空时返回 "{}"。
func encodeJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(data)
}

// ptrTime 返回 *time.Time。
func ptrTime(t time.Time) *time.Time { return &t }

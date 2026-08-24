package handler

import (
	"context"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"gorm.io/gorm"

	"github.com/your-team/koala-exam-backend/internal/application/dto"
	"github.com/your-team/koala-exam-backend/internal/application/exam"
	"github.com/your-team/koala-exam-backend/internal/application/grading"
	"github.com/your-team/koala-exam-backend/pkg/response"
)

// ExamHandler 考试 / 阅卷 handler
type ExamHandler struct {
	examApp    *exam.ExamApp
	gradingApp *grading.GradingApp
	db         *gorm.DB
}

func NewExamHandler(e *exam.ExamApp, g *grading.GradingApp, db *gorm.DB) *ExamHandler {
	return &ExamHandler{examApp: e, gradingApp: g, db: db}
}

// Create 创建考试
func (h *ExamHandler) Create(ctx context.Context, c *app.RequestContext) {
	var req dto.CreateExamReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, 400, 100001, "参数错误")
		return
	}
	id, err := h.examApp.CreateExam(ctx, &req, 0)
	if err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, map[string]any{"id": id})
}

// List 列出考试（管理员）
func (h *ExamHandler) List(ctx context.Context, c *app.RequestContext) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))
	keyword := c.Query("keyword")
	list, total, err := h.examApp.ListExams(ctx, page, size, int8(status), keyword)
	if err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Page(c, list, total, page, size)
}

// Get 获取单个考试
func (h *ExamHandler) Get(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	exam, err := h.examApp.GetExam(ctx, id)
	if err != nil {
		response.Fail(c, 404, 100004, "NotFound")
		return
	}
	response.Success(c, exam)
}

// Update 更新考试
func (h *ExamHandler) Update(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req dto.CreateExamReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, 400, 100001, "参数错误")
		return
	}
	if err := h.examApp.UpdateExam(ctx, id, &req); err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, nil)
}

// Delete 删除考试（软删）
func (h *ExamHandler) Delete(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.examApp.DeleteExam(ctx, id); err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, nil)
}

// AdminListRecords 管理员查看所有考试记录（带用户名/考试名 JOIN）
func (h *ExamHandler) AdminListRecords(ctx context.Context, c *app.RequestContext) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	examID, _ := strconv.ParseInt(c.DefaultQuery("exam_id", "0"), 10, 64)
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))
	keyword := c.Query("keyword")

	type Row struct {
		ID              int64      `json:"id"`
		ExamID          int64      `json:"exam_id"`
		UserID          int64      `json:"user_id"`
		UserName        string     `json:"user_name"`
		UserAccount      string     `json:"user_account"`
		ExamTitle       string     `json:"exam_title"`
		Status          int8       `json:"status"`
		TotalScore      float64    `json:"total_score"`
		ObjectiveScore  float64    `json:"objective_score"`
		SubjectiveScore float64    `json:"subjective_score"`
		Passed          bool       `json:"passed"`
		SubmitTime      *time.Time `json:"submit_time"`
		StartTime       *time.Time `json:"start_time"`
		Duration        int        `json:"duration"`
		TabSwitchCnt    int        `json:"tab_switch_cnt"`
		Answers         string     `json:"answers"`
	}

	var rows []Row
	var total int64

	q := h.db.WithContext(ctx).Table("ke_exam_record r").
		Select("r.id, r.exam_id, r.user_id, COALESCE(u.nickname,'') as user_name, COALESCE(u.username,'') as user_account, " +
			"COALESCE(e.title,'') as exam_title, r.status, r.total_score, r.objective_score, r.subjective_score, " +
			"r.passed, r.submit_time, r.start_time, r.duration, r.tab_switch_cnt, r.answers").
		Joins("LEFT JOIN ke_user u ON u.id = r.user_id").
		Joins("LEFT JOIN ke_exam e ON e.id = r.exam_id").
		Where("r.deleted_at IS NULL")

	if examID > 0 {
		q = q.Where("r.exam_id = ?", examID)
	}
	if status >= 0 {
		q = q.Where("r.status = ?", status)
	}
	if keyword != "" {
		q = q.Where("(u.username LIKE ? OR u.nickname LIKE ?)", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := q.Count(&total).Error; err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	if err := q.Order("r.id DESC").Offset((page - 1) * size).Limit(size).Scan(&rows).Error; err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Page(c, rows, total, page, size)
}

// Records 某场考试的所有记录
func (h *ExamHandler) Records(ctx context.Context, c *app.RequestContext) {
	examID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	list, total, err := h.examApp.ListRecordsByExam(ctx, examID, page, size)
	if err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Page(c, list, total, page, size)
}

// MyRecords 我的考试记录（学员视角）
func (h *ExamHandler) MyRecords(ctx context.Context, c *app.RequestContext) {
	uid := c.GetInt64("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	list, total, err := h.examApp.ListRecordsByUser(ctx, uid, page, size)
	if err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Page(c, list, total, page, size)
}

// GetRecord 获取单条考试记录
func (h *ExamHandler) GetRecord(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	rec, err := h.examApp.GetRecord(ctx, id)
	if err != nil {
		response.Fail(c, 404, 100004, "NotFound")
		return
	}
	response.Success(c, rec)
}

// Available 学员可参加的考试
func (h *ExamHandler) Available(ctx context.Context, c *app.RequestContext) {
	uid := c.GetInt64("user_id")
	list, err := h.examApp.ListAvailableExams(ctx, uid)
	if err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, list)
}

// Start 开始考试
func (h *ExamHandler) Start(ctx context.Context, c *app.RequestContext) {
	examID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	uid := c.GetInt64("user_id")
	resp, err := h.examApp.StartExam(ctx, examID, uid)
	if err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, resp)
}

// SaveAnswer 保存答案
func (h *ExamHandler) SaveAnswer(ctx context.Context, c *app.RequestContext) {
	var req dto.SaveAnswerReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, 400, 100001, "参数错误")
		return
	}
	if err := h.examApp.SaveAnswer(ctx, &req); err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, nil)
}

// Audit 上报行为审计
func (h *ExamHandler) Audit(ctx context.Context, c *app.RequestContext) {
	var req dto.AuditReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, 400, 100001, "参数错误")
		return
	}
	if err := h.examApp.AuditEvent(ctx, &req); err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, nil)
}

// Submit 交卷
func (h *ExamHandler) Submit(ctx context.Context, c *app.RequestContext) {
	var req dto.SubmitExamReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, 400, 100001, "参数错误")
		return
	}
	rec, err := h.examApp.SubmitExam(ctx, req.RecordID)
	if err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, rec)
}

// GradeSubjective 教师批改主观题（兼容旧版）
func (h *ExamHandler) GradeSubjective(ctx context.Context, c *app.RequestContext) {
	var req dto.GradeSubjectiveReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, 400, 100001, "参数错误")
		return
	}
	if err := h.gradingApp.GradeSubjective(ctx, &req); err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, nil)
}

// GradeSubjectiveBatch 教师批量批改主观题
func (h *ExamHandler) GradeSubjectiveBatch(ctx context.Context, c *app.RequestContext) {
	var req grading.SubjectiveGradeReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, 400, 100001, "参数错误")
		return
	}
	// 自动获取当前登录用户 ID 作为批改人
	if req.GraderID == 0 {
		req.GraderID = c.GetInt64("user_id")
	}
	if req.RecordID <= 0 {
		response.Fail(c, 400, 100001, "记录ID无效")
		return
	}
	if len(req.Items) == 0 {
		response.Fail(c, 400, 100001, "批改项为空")
		return
	}
	if err := h.gradingApp.GradeSubjectiveBatch(ctx, &req); err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, nil)
}

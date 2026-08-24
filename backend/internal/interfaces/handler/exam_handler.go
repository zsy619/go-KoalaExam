package handler

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/your-team/koala-exam-backend/internal/application/dto"
	"github.com/your-team/koala-exam-backend/internal/application/exam"
	"github.com/your-team/koala-exam-backend/internal/application/grading"
	"github.com/your-team/koala-exam-backend/pkg/response"
)

type ExamHandler struct {
	examApp    *exam.ExamApp
	gradingApp *grading.GradingApp
}

func NewExamHandler(e *exam.ExamApp, g *grading.GradingApp) *ExamHandler {
	return &ExamHandler{examApp: e, gradingApp: g}
}

func (h *ExamHandler) Create(ctx context.Context, c *app.RequestContext) {
	var req dto.CreateExamReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, 400, 100001, "参数错误")
		return
	}
	uid := c.GetInt64("user_id")
	id, err := h.examApp.CreateExam(ctx, &req, uid)
	if err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"id": id})
}

func (h *ExamHandler) List(ctx context.Context, c *app.RequestContext) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status, _ := strconv.Atoi(c.Query("status"))
	list, total, err := h.examApp.ListExams(ctx, page, size, int8(status), c.Query("keyword"))
	if err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Page(c, list, total, page, size)
}

func (h *ExamHandler) Available(ctx context.Context, c *app.RequestContext) {
	uid := c.GetInt64("user_id")
	list, err := h.examApp.ListAvailableExams(ctx, uid)
	if err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, list)
}

func (h *ExamHandler) Start(ctx context.Context, c *app.RequestContext) {
	examID := (func() int64 { v, _ := strconv.ParseInt(c.Param("id"), 10, 64); return v })()
	uid := c.GetInt64("user_id")
	resp, err := h.examApp.StartExam(ctx, examID, uid)
	if err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, resp)
}

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
	// 自动阅卷
	_, _ = h.gradingApp.AutoGrade(ctx, rec.ID)
	response.Success(c, rec)
}

func (h *ExamHandler) Records(ctx context.Context, c *app.RequestContext) {
	examID := (func() int64 { v, _ := strconv.ParseInt(c.Param("id"), 10, 64); return v })()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	list, total, err := h.examApp.ListRecordsByExam(ctx, examID, page, size)
	if err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Page(c, list, total, page, size)
}

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
func (h *ExamHandler) Get(ctx context.Context, c *app.RequestContext) {
	id := (func() int64 { v, _ := strconv.ParseInt(c.Param("id"), 10, 64); return v })()
	ex, err := h.examApp.GetExam(ctx, id)
	if err != nil { response.Fail(c, 404, 100004, "NotFound"); return }
	response.Success(c, ex)
}

func (h *ExamHandler) Update(ctx context.Context, c *app.RequestContext) {
	id := (func() int64 { v, _ := strconv.ParseInt(c.Param("id"), 10, 64); return v })()
	var req dto.CreateExamReq
	if err := c.BindAndValidate(&req); err != nil { response.Fail(c, 400, 100001, "参数错误"); return }
	if err := h.examApp.UpdateExam(ctx, id, &req); err != nil { response.Fail(c, 500, 100005, err.Error()); return }
	response.Success(c, nil)
}

func (h *ExamHandler) Delete(ctx context.Context, c *app.RequestContext) {
	id := (func() int64 { v, _ := strconv.ParseInt(c.Param("id"), 10, 64); return v })()
	if err := h.examApp.DeleteExam(ctx, id); err != nil { response.Fail(c, 500, 100005, err.Error()); return }
	response.Success(c, nil)
}

func (h *ExamHandler) GetRecord(ctx context.Context, c *app.RequestContext) {
	id := (func() int64 { v, _ := strconv.ParseInt(c.Param("id"), 10, 64); return v })()
	rec, err := h.examApp.GetRecord(ctx, id)
	if err != nil { response.Fail(c, 404, 100004, "NotFound"); return }
	response.Success(c, rec)
}

package handler

import (
	"context"

	"strconv"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/your-team/koala-exam-backend/internal/application/dto"
	"github.com/your-team/koala-exam-backend/internal/application/question"
	"github.com/your-team/koala-exam-backend/pkg/response"
)

type PaperHandler struct{ app *question.PaperApp }

func NewPaperHandler(a *question.PaperApp) *PaperHandler { return &PaperHandler{app: a} }

func (h *PaperHandler) Create(ctx context.Context, c *app.RequestContext) {
	var req dto.CreatePaperReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, 400, 100001, "参数错误")
		return
	}
	uid := c.GetInt64("user_id")
	id, err := h.app.Create(ctx, &req, uid)
	if err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"id": id})
}


func (h *PaperHandler) Update(ctx context.Context, c *app.RequestContext) {
	id := (func() int64 { v, _ := strconv.ParseInt(c.Param("id"), 10, 64); return v })()
	var req dto.CreatePaperReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, 400, 100001, "参数错误")
		return
	}
	if err := h.app.Update(ctx, id, &req); err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, nil)
}
func (h *PaperHandler) Get(ctx context.Context, c *app.RequestContext) {
	id := (func() int64 { v, _ := strconv.ParseInt(c.Param("id"), 10, 64); return v })()
	p, qs, err := h.app.GetDetail(ctx, id)
	if err != nil {
		response.Fail(c, 404, 100004, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"paper": p, "questions": qs})
}

func (h *PaperHandler) List(ctx context.Context, c *app.RequestContext) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	list, total, err := h.app.List(ctx, page, size, c.Query("keyword"))
	if err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Page(c, list, total, page, size)
}

func (h *PaperHandler) Delete(ctx context.Context, c *app.RequestContext) {
	id := (func() int64 { v, _ := strconv.ParseInt(c.Param("id"), 10, 64); return v })()
	if err := h.app.Delete(ctx, id); err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, nil)
}

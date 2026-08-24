package handler

import (
	"context"

	"strconv"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/your-team/koala-exam-backend/internal/application/dto"
	"github.com/your-team/koala-exam-backend/internal/application/question"
	"github.com/your-team/koala-exam-backend/internal/domain/entity"
	"github.com/your-team/koala-exam-backend/pkg/response"
)

type QuestionHandler struct{ app *question.QuestionApp }

func NewQuestionHandler(a *question.QuestionApp) *QuestionHandler { return &QuestionHandler{app: a} }

func (h *QuestionHandler) Create(ctx context.Context, c *app.RequestContext) {
	var req dto.CreateQuestionReq
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

func (h *QuestionHandler) Update(ctx context.Context, c *app.RequestContext) {
	id := (func() int64 { v, _ := strconv.ParseInt(c.Param("id"), 10, 64); return v })()
	var req dto.CreateQuestionReq
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

func (h *QuestionHandler) Delete(ctx context.Context, c *app.RequestContext) {
	id := (func() int64 { v, _ := strconv.ParseInt(c.Param("id"), 10, 64); return v })()
	if err := h.app.Delete(ctx, id); err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *QuestionHandler) Get(ctx context.Context, c *app.RequestContext) {
	id := (func() int64 { v, _ := strconv.ParseInt(c.Param("id"), 10, 64); return v })()
	q, err := h.app.Get(ctx, id)
	if err != nil {
		response.Fail(c, 404, 100004, "题目不存在")
		return
	}
	response.Success(c, q)
}

func (h *QuestionHandler) List(ctx context.Context, c *app.RequestContext) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	cat, _ := strconv.ParseInt(c.Query("category_id"), 10, 64)
	qt, _ := strconv.ParseInt(c.Query("type"), 10, 64)
	diff, _ := strconv.ParseInt(c.Query("difficulty"), 10, 64)
	keyword := c.Query("keyword")
	list, total, err := h.app.List(ctx, page, size, cat, qt, diff, keyword)
	if err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Page(c, list, total, page, size)
}

func (h *QuestionHandler) BatchImport(ctx context.Context, c *app.RequestContext) {
	uid := c.GetInt64("user_id")
	var req dto.BatchImportReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, 400, 100001, "参数错误")
		return
	}
	n, err := h.app.BatchImport(ctx, &req, uid)
	if err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, map[string]int{"imported": n})
}

func (h *QuestionHandler) ListCategories(ctx context.Context, c *app.RequestContext) {
	list, err := h.app.ListCategories(ctx)
	if err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, list)
}

func (h *QuestionHandler) CreateCategory(ctx context.Context, c *app.RequestContext) {
	var cat entity.QuestionCategory
	if err := c.BindAndValidate(&cat); err != nil {
		response.Fail(c, 400, 100001, "参数错误")
		return
	}
	uid := c.GetInt64("user_id")
	cat.CreatorID = uid
	if err := h.app.CreateCategory(ctx, &cat); err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, cat)
}


// UpdateCategory 更新分类。
func (h *QuestionHandler) UpdateCategory(ctx context.Context, c *app.RequestContext) {
	id := (func() int64 { v, _ := strconv.ParseInt(c.Param("id"), 10, 64); return v })()
	var cat entity.QuestionCategory
	if err := c.BindAndValidate(&cat); err != nil {
		response.Fail(c, 400, 100001, "参数错误")
		return
	}
	cat.ID = id
	if err := h.app.UpdateCategory(ctx, &cat); err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, cat)
}

// DeleteCategory 删除分类。
func (h *QuestionHandler) DeleteCategory(ctx context.Context, c *app.RequestContext) {
	id := (func() int64 { v, _ := strconv.ParseInt(c.Param("id"), 10, 64); return v })()
	if err := h.app.DeleteCategory(ctx, id); err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, nil)
}

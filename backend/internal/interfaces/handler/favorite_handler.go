package handler

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/your-team/koala-exam-backend/internal/application/dto"
	"github.com/your-team/koala-exam-backend/internal/application/favorite"
	"github.com/your-team/koala-exam-backend/pkg/response"
)

type FavoriteHandler struct{ app *favorite.FavoriteApp }

func NewFavoriteHandler(a *favorite.FavoriteApp) *FavoriteHandler { return &FavoriteHandler{app: a} }

// Toggle 单个收藏/取消收藏
func (h *FavoriteHandler) Toggle(ctx context.Context, c *app.RequestContext) {
	var req dto.FavoriteReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, 400, 100001, "参数错误")
		return
	}
	uid := c.GetInt64("user_id")
	favorited, err := h.app.Toggle(ctx, uid, &req)
	if err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, map[string]bool{"favorited": favorited})
}

// BatchAdd 批量收藏（错题自动入库）
func (h *FavoriteHandler) BatchAdd(ctx context.Context, c *app.RequestContext) {
	var req dto.BatchFavoriteReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, 400, 100001, "参数错误")
		return
	}
	uid := c.GetInt64("user_id")
	n, err := h.app.BatchAdd(ctx, uid, &req)
	if err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, map[string]int{"count": n})
}

// IsFavorited
func (h *FavoriteHandler) IsFavorited(ctx context.Context, c *app.RequestContext) {
	uid := c.GetInt64("user_id")
	tt, _ := strconv.Atoi(c.Query("target_type"))
	tid, _ := strconv.ParseInt(c.Query("target_id"), 10, 64)
	res, _ := h.app.IsFavorited(ctx, uid, int8(tt), tid)
	response.Success(c, map[string]bool{"favorited": res})
}

// ListFavorites
func (h *FavoriteHandler) List(ctx context.Context, c *app.RequestContext) {
	uid := c.GetInt64("user_id")
	tt, _ := strconv.Atoi(c.Query("target_type"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	list, total, err := h.app.ListFavorites(ctx, uid, int8(tt), page, size)
	if err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Page(c, list, total, page, size)
}

// ListFolders 列出收藏夹
func (h *FavoriteHandler) ListFolders(ctx context.Context, c *app.RequestContext) {
	uid := c.GetInt64("user_id")
	list, err := h.app.ListFolders(ctx, uid)
	if err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, list)
}

// CreateFolder
func (h *FavoriteHandler) CreateFolder(ctx context.Context, c *app.RequestContext) {
	uid := c.GetInt64("user_id")
	var req dto.CreateFolderReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, 400, 100001, "参数错误")
		return
	}
	id, err := h.app.CreateFolder(ctx, uid, &req)
	if err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"id": id})
}

// DeleteFolder
func (h *FavoriteHandler) DeleteFolder(ctx context.Context, c *app.RequestContext) {
	uid := c.GetInt64("user_id")
	id := (func() int64 { v, _ := strconv.ParseInt(c.Param("id"), 10, 64); return v })()
	if err := h.app.DeleteFolder(ctx, uid, id); err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, nil)
}

// GetWrongBook 错题本（带掌握度筛选）
func (h *FavoriteHandler) GetWrongBook(ctx context.Context, c *app.RequestContext) {
	uid := c.GetInt64("user_id")
	mastery, _ := strconv.Atoi(c.Query("mastery_level"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	list, total, err := h.app.GetWrongBook(ctx, uid, int8(mastery), page, size)
	if err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Page(c, list, total, page, size)
}

// MarkReviewed
func (h *FavoriteHandler) MarkReviewed(ctx context.Context, c *app.RequestContext) {
	id := (func() int64 { v, _ := strconv.ParseInt(c.Param("id"), 10, 64); return v })()
	mastery, _ := strconv.Atoi(c.Query("mastery_level"))
	if err := h.app.MarkReviewed(ctx, id, int8(mastery)); err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, nil)
}

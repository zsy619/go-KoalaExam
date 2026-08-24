package handler

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/your-team/koala-exam-backend/internal/application/dto"
	"github.com/your-team/koala-exam-backend/internal/application/favorite"
	"github.com/your-team/koala-exam-backend/internal/domain/entity"
	"github.com/your-team/koala-exam-backend/pkg/response"
)

// FavoriteHandler 收藏 HTTP 处理器。
type FavoriteHandler struct {
	app *favorite.FavoriteApp
}

// NewFavoriteHandler 构造收藏处理器。
func NewFavoriteHandler(a *favorite.FavoriteApp) *FavoriteHandler { return &FavoriteHandler{app: a} }

// Toggle 切换收藏。
func (h *FavoriteHandler) Toggle(ctx context.Context, c *app.RequestContext) {
	var req dto.FavoriteReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, 400, 100001, "参数错误")
		return
	}
	uid := c.GetInt64("user_id")
	toggleReq := &favorite.ToggleReq{
		UserID:     uid,
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
		FolderID:   derefInt64(req.FolderID),
	}
	favorited, err := h.app.ToggleFavorite(ctx, toggleReq)
	if err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, map[string]bool{"favorited": favorited})
}

// BatchAdd 批量收藏。
func (h *FavoriteHandler) BatchAdd(ctx context.Context, c *app.RequestContext) {
	var req dto.BatchFavoriteReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, 400, 100001, "参数错误")
		return
	}
	uid := c.GetInt64("user_id")
	batchReq := &favorite.BatchAddReq{
		UserID:      uid,
		QuestionIDs: req.TargetIDs,
		FolderID:    derefInt64(req.FolderID),
	}
	result, err := h.app.BatchAdd(ctx, batchReq)
	if err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{
		"added_count":   len(result.AddedIDs),
		"skipped_count": len(result.SkippedIDs),
		"added_ids":     result.AddedIDs,
		"skipped_ids":   result.SkippedIDs,
	})
}

// IsFavorited 判断是否已收藏。
func (h *FavoriteHandler) IsFavorited(ctx context.Context, c *app.RequestContext) {
	uid := c.GetInt64("user_id")
	tt, _ := strconv.Atoi(c.Query("target_type"))
	tid, _ := strconv.ParseInt(c.Query("target_id"), 10, 64)
	res, _ := h.app.IsFavorited(ctx, uid, tid, int8(tt))
	response.Success(c, map[string]bool{"favorited": res})
}

// List 列出收藏。
func (h *FavoriteHandler) List(ctx context.Context, c *app.RequestContext) {
	uid := c.GetInt64("user_id")
	tt, _ := strconv.Atoi(c.Query("target_type"))
	list, err := h.app.ListFavorites(ctx, uid, int8(tt), 0)
	if err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"list": list, "total": len(list)})
}

// ListFolders 列出收藏夹。
func (h *FavoriteHandler) ListFolders(ctx context.Context, c *app.RequestContext) {
	uid := c.GetInt64("user_id")
	list, err := h.app.ListFolders(ctx, uid)
	if err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, list)
}

// CreateFolder 创建收藏夹。
// derefInt64 安全解引用 *int64。
func derefInt64(p *int64) int64 {
	if p == nil {
		return 0
	}
	return *p
}

func (h *FavoriteHandler) CreateFolder(ctx context.Context, c *app.RequestContext) {
	uid := c.GetInt64("user_id")
	var req dto.CreateFolderReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, 400, 100001, "参数错误")
		return
	}
	folder := &entity.FavoriteFolder{
		UserID: uid,
		Name:   req.Name,
		Color:  req.Color,
		Icon:   req.Icon,
		
	}
	if err := h.app.CreateFolder(ctx, folder); err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"id": folder.ID})
}

// DeleteFolder 删除收藏夹。
func (h *FavoriteHandler) DeleteFolder(ctx context.Context, c *app.RequestContext) {
	id := (func() int64 { v, _ := strconv.ParseInt(c.Param("id"), 10, 64); return v })()
	if err := h.app.DeleteFolder(ctx, id); err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, nil)
}

// GetWrongBook 错题本。
func (h *FavoriteHandler) GetWrongBook(ctx context.Context, c *app.RequestContext) {
	uid := c.GetInt64("user_id")
	reviewed, _ := strconv.ParseBool(c.DefaultQuery("is_reviewed", ""))
	var ptr *bool
	if c.Query("is_reviewed") != "" {
		ptr = &reviewed
	}
	list, _, err := h.app.GetWrongBook(ctx, uid, favorite.WrongBookQuery{
		IsReviewed: ptr,
	})
	if err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"list": list, "total": len(list)})
}


// GetStats 收藏统计（按类型/文件夹聚合）。
func (h *FavoriteHandler) GetStats(ctx context.Context, c *app.RequestContext) {
	uid := c.GetInt64("user_id")
	stats, err := h.app.GetStats(ctx, uid)
	if err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, stats)
}

// MasteryDistribution 错题掌握度分布。
func (h *FavoriteHandler) MasteryDistribution(ctx context.Context, c *app.RequestContext) {
	uid := c.GetInt64("user_id")
	dist, err := h.app.GetMasteryDistribution(ctx, uid)
	if err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, dist)
}

// MarkReviewed 标记错题为已复习。
func (h *FavoriteHandler) MarkReviewed(ctx context.Context, c *app.RequestContext) {
	id := (func() int64 { v, _ := strconv.ParseInt(c.Param("id"), 10, 64); return v })()
	mastery, _ := strconv.Atoi(c.DefaultQuery("mastery_level", "3"))
	if err := h.app.MarkReviewed(ctx, id, int8(mastery)); err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, nil)
}

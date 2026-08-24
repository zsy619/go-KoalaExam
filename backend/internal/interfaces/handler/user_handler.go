package handler

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/your-team/koala-exam-backend/internal/application/dto"
	"github.com/your-team/koala-exam-backend/internal/application/user"
	"github.com/your-team/koala-exam-backend/internal/domain/entity"
	"github.com/your-team/koala-exam-backend/pkg/response"
)

type UserHandler struct{ app *user.UserApp }

func NewUserHandler(a *user.UserApp) *UserHandler { return &UserHandler{app: a} }

func (h *UserHandler) Login(ctx context.Context, c *app.RequestContext) {
	var req dto.LoginReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, 400, 100001, "参数错误")
		return
	}
	ip := c.ClientIP()
	resp, err := h.app.Login(ctx, &req, ip)
	if err != nil {
		response.Fail(c, 401, 200002, err.Error())
		return
	}
	response.Success(c, resp)
}

func (h *UserHandler) RefreshToken(ctx context.Context, c *app.RequestContext) {
	var req dto.RefreshReq
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, 400, 100001, "参数错误")
		return
	}
	token, exp, err := h.app.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		response.Fail(c, 401, 200005, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"access_token": token, "expires_in": exp})
}

func (h *UserHandler) Profile(ctx context.Context, c *app.RequestContext) {
	uid := c.GetInt64("user_id")
	u, err := h.app.GetProfile(ctx, uid)
	if err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, u)
}

func (h *UserHandler) List(ctx context.Context, c *app.RequestContext) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	role, _ := strconv.Atoi(c.Query("role"))
	keyword := c.Query("keyword")
	list, total, err := h.app.ListUsers(ctx, page, size, int8(role), keyword)
	if err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Page(c, list, total, page, size)
}

func (h *UserHandler) Create(ctx context.Context, c *app.RequestContext) {
	var u entity.User
	if err := c.BindAndValidate(&u); err != nil {
		response.Fail(c, 400, 100001, "参数错误")
		return
	}
	if err := h.app.CreateUser(ctx, &u); err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, u)
}

func (h *UserHandler) ResetPassword(ctx context.Context, c *app.RequestContext) {
	id := (func() int64 { v, _ := strconv.ParseInt(c.Param("id"), 10, 64); return v })()
	newPwd, err := h.app.ResetPassword(ctx, id)
	if err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, map[string]string{"new_password": newPwd})
}
func (h *UserHandler) ChangePassword(ctx context.Context, c *app.RequestContext) {
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, 400, 100001, "参数错误")
		return
	}
	uid := c.GetInt64("user_id")
	if err := h.app.ChangePassword(ctx, uid, req.OldPassword, req.NewPassword); err != nil {
		response.Fail(c, 500, 200003, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *UserHandler) UpdateProfile(ctx context.Context, c *app.RequestContext) {
	var req struct {
		Nickname string `json:"nickname"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Avatar   string `json:"avatar"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, 400, 100001, "参数错误")
		return
	}
	uid := c.GetInt64("user_id")
	if err := h.app.UpdateProfile(ctx, uid, req.Nickname, req.Email, req.Phone, req.Avatar); err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *UserHandler) Delete(ctx context.Context, c *app.RequestContext) {
	id := (func() int64 { v, _ := strconv.ParseInt(c.Param("id"), 10, 64); return v })()
	if err := h.app.DeleteUser(ctx, id); err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *UserHandler) ToggleStatus(ctx context.Context, c *app.RequestContext) {
	id := (func() int64 { v, _ := strconv.ParseInt(c.Param("id"), 10, 64); return v })()
	var req struct { Status int8 `json:"status" binding:"required"` }
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, 400, 100001, "参数错误")
		return
	}
	if err := h.app.ToggleStatus(ctx, id, req.Status); err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *UserHandler) GetByID(ctx context.Context, c *app.RequestContext) {
	id := (func() int64 { v, _ := strconv.ParseInt(c.Param("id"), 10, 64); return v })()
	u, err := h.app.GetByID(ctx, id)
	if err != nil {
		response.Fail(c, 404, 200001, err.Error())
		return
	}
	response.Success(c, u)
}

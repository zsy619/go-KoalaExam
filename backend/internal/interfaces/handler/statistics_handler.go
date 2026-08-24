package handler

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/your-team/koala-exam-backend/internal/application/statistics"
	"github.com/your-team/koala-exam-backend/pkg/response"
)

// StatisticsHandler 统计接口
type StatisticsHandler struct {
	app *statistics.StatisticsApp
}

func NewStatisticsHandler(a *statistics.StatisticsApp) *StatisticsHandler {
	return &StatisticsHandler{app: a}
}

// Dashboard 仪表盘汇总（超管首页）
func (h *StatisticsHandler) Dashboard(ctx context.Context, c *app.RequestContext) {
	d, err := h.app.DashboardSummary(ctx)
	if err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, d)
}

// ExamOverview 考试概览（教师/超管）
func (h *StatisticsHandler) ExamOverview(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	overview, err := h.app.ExamOverview(ctx, id)
	if err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, overview)
}

// MyLearning 我的学习统计（学员）
func (h *StatisticsHandler) MyLearning(ctx context.Context, c *app.RequestContext) {
	uid := c.GetInt64("user_id")
	stats, err := h.app.UserLearningStats(ctx, uid)
	if err != nil {
		response.Fail(c, 500, 100005, err.Error())
		return
	}
	response.Success(c, stats)
}

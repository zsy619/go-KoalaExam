package response

import (
	"github.com/cloudwego/hertz/pkg/app"

	"github.com/your-team/koala-exam-backend/internal/domain/errcode"
)

// Resp 统一返回结构
type Resp struct {
	Code    errcode.Code `json:"code"`
	Message string       `json:"message"`
	Data    interface{}  `json:"data,omitempty"`
	TraceID string       `json:"trace_id,omitempty"`
}

// PageData 分页数据
type PageData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

func Success(c *app.RequestContext, data interface{}) {
	c.JSON(200, Resp{Code: errcode.CodeSuccess, Message: "success", Data: data})
}

func SuccessWithCode(c *app.RequestContext, code int, data interface{}) {
	c.JSON(code, Resp{Code: errcode.CodeSuccess, Message: "success", Data: data})
}

func Fail(c *app.RequestContext, code int, ec errcode.Code, msg string) {
	if msg == "" {
		msg = errcode.Message(ec)
	}
	c.JSON(code, Resp{Code: ec, Message: msg})
}

func FailWithErr(c *app.RequestContext, code int, e *errcode.AppError) {
	c.JSON(code, Resp{Code: e.Code, Message: e.Error()})
}

func Page(c *app.RequestContext, list interface{}, total int64, page, size int) {
	Success(c, PageData{List: list, Total: total, Page: page, PageSize: size})
}

package response_test

import (
	"encoding/json"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"github.com/your-team/koala-exam-backend/internal/domain/errcode"
	"github.com/your-team/koala-exam-backend/pkg/response"
)

func TestSuccess(t *testing.T) {
	c := &app.RequestContext{}
	response.Success(c, map[string]string{"hello": "world"})
	if c.Response.StatusCode() != consts.StatusOK { t.Fatal("status not 200") }
	var resp response.Resp
	if err := json.Unmarshal(c.Response.Body(), &resp); err != nil { t.Fatal(err) }
	if resp.Code != 0 { t.Fatalf("expected code 0, got %d", resp.Code) }
}

func TestFail(t *testing.T) {
	c := &app.RequestContext{}
	response.Fail(c, 400, errcode.CodeBadRequest, "bad")
	if c.Response.StatusCode() != consts.StatusBadRequest { t.Fatal("status not 400") }
}

package errcode

// 业务错误码（与 HTTP 状态解耦）
// 格式：业务域(2位) + 业务编号(4位)
// 1xxxxx: 通用  2xxxxx: 用户  3xxxxx: 题目  4xxxxx: 试卷/考试  5xxxxx: 阅卷  6xxxxx: 收藏  7xxxxx: 系统
type Code int

const (
	CodeSuccess        Code = 0
	CodeBadRequest     Code = 100001
	CodeUnauthorized   Code = 100002
	CodeForbidden      Code = 100003
	CodeNotFound       Code = 100004
	CodeInternal       Code = 100005
	CodeTooManyRequest Code = 100006

	CodeUserNotExist      Code = 200001
	CodeUserPasswordWrong Code = 200002
	CodeUserDisabled      Code = 200003
	CodeTokenExpired      Code = 200004
	CodeTokenInvalid      Code = 200005
	CodePermissionDenied  Code = 200006

	CodeQuestionNotExist Code = 300001
	CodeQuestionEmpty    Code = 300002

	CodeExamNotExist   Code = 400001
	CodeExamNotRunning Code = 400002
	CodeExamExpired    Code = 400003
	CodeExamSubmitted  Code = 400004
	CodePaperNotExist  Code = 400005

	CodeFavoriteExist    Code = 600001
	CodeFavoriteNotExist Code = 600002

	CodeSystemError Code = 700001
)

// Message 错误码文案
func Message(c Code) string {
	switch c {
	case CodeSuccess:
		return "成功"
	case CodeBadRequest:
		return "请求参数错误"
	case CodeUnauthorized:
		return "未登录或登录已过期"
	case CodeForbidden:
		return "无权限访问"
	case CodeNotFound:
		return "资源不存在"
	case CodeInternal:
		return "服务器内部错误"
	case CodeTooManyRequest:
		return "请求过于频繁，请稍后再试"
	case CodeUserNotExist:
		return "用户不存在"
	case CodeUserPasswordWrong:
		return "账号或密码错误"
	case CodeUserDisabled:
		return "账号已被禁用"
	case CodeTokenExpired:
		return "Token 已过期"
	case CodeTokenInvalid:
		return "Token 无效"
	case CodePermissionDenied:
		return "权限不足"
	case CodeQuestionNotExist:
		return "题目不存在"
	case CodeQuestionEmpty:
		return "题库为空"
	case CodeExamNotExist:
		return "考试不存在"
	case CodeExamNotRunning:
		return "考试不在进行中"
	case CodeExamExpired:
		return "考试已过期"
	case CodeExamSubmitted:
		return "已交卷，请勿重复提交"
	case CodePaperNotExist:
		return "试卷不存在"
	case CodeFavoriteExist:
		return "已收藏"
	case CodeFavoriteNotExist:
		return "收藏记录不存在"
	case CodeSystemError:
		return "系统异常"
	}
	return "未知错误"
}

// AppError 业务错误（实现 error 接口）
type AppError struct {
	Code    Code
	Message string
	Cause   error
}

func (e *AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return Message(e.Code)
}

func (e *AppError) Unwrap() error { return e.Cause }

func New(code Code, msg string) *AppError {
	if msg == "" {
		msg = Message(code)
	}
	return &AppError{Code: code, Message: msg}
}

func Wrap(code Code, cause error) *AppError {
	return &AppError{Code: code, Message: Message(code), Cause: cause}
}

package dto

// LoginReq 登录请求
type LoginReq struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
	Password string `json:"password" binding:"required,min=6,max=64"`
}

// LoginResp 登录响应
type LoginResp struct {
	User         interface{} `json:"user"`
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresIn    int64       `json:"expires_in"`
}

// RefreshReq 刷新 token
type RefreshReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Package user 用户应用服务。
//
// 遵循 Google Go 风格：
//   - 命名简洁（Login/CreateUser/ResetPassword）
//   - context 作为第一个参数
//   - error 作为最后一个返回值
//   - 通过 interface 注入，便于测试
package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/your-team/koala-exam-backend/internal/domain/consts"
	"github.com/your-team/koala-exam-backend/internal/domain/entity"
	"github.com/your-team/koala-exam-backend/internal/domain/errcode"
	"github.com/your-team/koala-exam-backend/internal/domain/repository"
	"github.com/your-team/koala-exam-backend/pkg/encrypt"
	"github.com/your-team/koala-exam-backend/pkg/jwt"
	"github.com/your-team/koala-exam-backend/pkg/utils"
)

// TokenService Token 服务接口。
type TokenService interface {
	Generate(userID int64, username string, role int8, tokenType string) (string, error)
}

// UserApp 用户应用服务。
type UserApp struct {
	users repository.UserRepository
	tokens TokenService
}

// NewUserApp 构造用户应用服务。
func NewUserApp(users repository.UserRepository, tokens TokenService) *UserApp {
	return &UserApp{users: users, tokens: tokens}
}

// Login 校验用户名/密码，返回 JWT。
//
// 采用 Google Go 风格：返回 *LoginResult 与 error。
func (a *UserApp) Login(ctx context.Context, username, password, ip string) (*LoginResult, error) {
	user, err := a.users.GetByUsername(ctx, username)
	if err != nil {
		return nil, errcode.New(errcode.CodeUserNotExist, "UserNotExist")
	}
	if user.Status == consts.UserStatusDisabled {
		return nil, errcode.New(errcode.CodeUserDisabled, "UserDisabled")
	}
	if !encrypt.BcryptCheck(user.Password, password) {
		return nil, errcode.New(errcode.CodeUserPasswordWrong, "UserPasswordWrong")
	}

	access, _, err := a.tokens.Generate(user.ID, user.Username, user.Role, jwt.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}
	refresh, _, err := a.tokens.Generate(user.ID, user.Username, user.Role, jwt.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	// 异步更新最后登录信息（可改为异步队列）
	_ = a.users.UpdateLastLogin(ctx, user.ID, ip)

	return &LoginResult{
		User:         user,
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    jwt.AccessTTL.Seconds(),
	}, nil
}

// RefreshToken 用 refresh token 换取新的 access token。
func (a *UserApp) RefreshToken(ctx context.Context, refreshToken string) (string, float64, error) {
	claims, err := jwt.Parse(refreshToken)
	if err != nil {
		return "", 0, errcode.New(errcode.CodeTokenInvalid, "TokenInvalid")
	}
	if claims.Type != jwt.RefreshToken {
		return "", 0, errcode.New(errcode.CodeTokenInvalid, "NotRefreshToken")
	}
	access, _, err := a.tokens.Generate(claims.UserID, claims.Username, claims.Role, jwt.AccessToken)
	if err != nil {
		return "", 0, err
	}
	return access, jwt.AccessTTL.Seconds(), nil
}

// GetProfile 获取个人资料。
func (a *UserApp) GetProfile(ctx context.Context, uid int64) (*entity.User, error) {
	user, err := a.users.GetByID(ctx, uid)
	if err != nil {
		return nil, errcode.New(errcode.CodeUserNotExist, "UserNotExist")
	}
	return user, nil
}

// UpdateProfile 更新个人资料（昵称/手机/邮箱）。
func (a *UserApp) UpdateProfile(ctx context.Context, uid int64, nickname, newPhone, email string) error {
	user := &entity.User{
		ID:       uid,
		Nickname: nickname,
		Phone:    newPhone,
		Email:    email,
	}
	return a.users.UpdateProfile(ctx, user)
}

// ChangePassword 修改密码（需验证旧密码）。
func (a *UserApp) ChangePassword(ctx context.Context, uid int64, oldPwd, newPwd string) error {
	user, err := a.users.GetByID(ctx, uid)
	if err != nil {
		return errcode.New(errcode.CodeUserNotExist, "UserNotExist")
	}
	if !encrypt.BcryptCheck(user.Password, oldPwd) {
		return errcode.New(errcode.CodeUserPasswordWrong, "UserPasswordWrong")
	}
	hashed, err := encrypt.BcryptPassword(newPwd)
	if err != nil {
		return err
	}
	return a.users.UpdatePassword(ctx, uid, hashed)
}

// CreateUser 创建用户。
func (a *UserApp) CreateUser(ctx context.Context, u *entity.User, plainPassword string) (int64, error) {
	if u.Username == "" {
		return 0, errors.New("用户名不能为空")
	}
	if _, err := a.users.GetByUsername(ctx, u.Username); err == nil {
		return 0, errcode.New(errcode.CodeBadRequest, "UsernameExists")
	}
	if plainPassword == "" {
		plainPassword = "koala123"
	}
	hashed, err := encrypt.BcryptPassword(plainPassword)
	if err != nil {
		return 0, err
	}
	u.Password = hashed
	u.LastLoginAt = nil
	u.LastLoginIP = ""
	if err := a.users.Create(ctx, u); err != nil {
		return 0, err
	}
	return u.ID, nil
}

// ListUsers 分页查询用户。
func (a *UserApp) ListUsers(ctx context.Context, filter repository.UserListFilter) ([]entity.User, int64, error) {
	return a.users.List(ctx, filter)
}

// GetByID 根据 ID 查询。
func (a *UserApp) GetByID(ctx context.Context, id int64) (*entity.User, error) {
	return a.users.GetByID(ctx, id)
}

// ResetPassword 重置密码为默认值，返回明文密码。
func (a *UserApp) ResetPassword(ctx context.Context, uid int64) (string, error) {
	newPwd := utils.RandString(8)
	hashed, err := encrypt.BcryptPassword(newPwd)
	if err != nil {
		return "", err
	}
	if err := a.users.UpdatePassword(ctx, uid, hashed); err != nil {
		return "", err
	}
	return newPwd, nil
}

// ChangeStatus 切换启用/禁用。
func (a *UserApp) ChangeStatus(ctx context.Context, uid int64, status int8) error {
	user, err := a.users.GetByID(ctx, uid)
	if err != nil {
		return errcode.New(errcode.CodeUserNotExist, "UserNotExist")
	}
	user.Status = status
	return a.users.Update(ctx, user)
}

// DeleteUser 软删除。
func (a *UserApp) DeleteUser(ctx context.Context, uid int64) error {
	return a.users.Delete(ctx, uid)
}

// LoginResult 登录结果。
type LoginResult struct {
	User         *entity.User `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    float64      `json:"expires_in"`
}

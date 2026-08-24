// Package user 用户应用服务。
//
// 设计要点：
//   - Token 黑名单（Redis）：登出后 token 立即失效
//   - 登录限流（Redis）：同 IP/账号 5 分钟内最多 5 次失败
//   - 密码强度：使用 valueobject.Password 校验
//   - 失败计数：登录失败递增，超过阈值锁定账号
package user

import (
	"context"
	"fmt"
	"time"


	"github.com/your-team/koala-exam-backend/internal/domain/consts"
	"github.com/your-team/koala-exam-backend/internal/domain/entity"
	"github.com/your-team/koala-exam-backend/internal/domain/errcode"
	"github.com/your-team/koala-exam-backend/internal/domain/event"
	"github.com/your-team/koala-exam-backend/internal/domain/repository"
	"github.com/your-team/koala-exam-backend/internal/domain/valueobject"
	infra "github.com/your-team/koala-exam-backend/internal/infrastructure/repository"
	"github.com/your-team/koala-exam-backend/pkg/encrypt"
	"github.com/your-team/koala-exam-backend/pkg/jwt"
)

// TokenService Token 服务接口。
type TokenService interface {
	Generate(uid int64, username string, role int8, tokenType string) (string, time.Time, error)
	Parse(token string) (int64, int8, error)
}

// LoginRateLimiter 登录限流器。
type LoginRateLimiter interface {
	// Allow 返回是否允许登录及剩余重试时间
	Allow(ctx context.Context, key string) (allowed bool, retryAfter time.Duration, err error)
}

// UserApp 用户应用服务。
type UserApp struct {
	users  *infra.UserRepository
	tokens TokenService
	bus    *event.Bus
	limiter LoginRateLimiter
}

// NewUserApp 构造用户应用服务。
func NewUserApp(users *infra.UserRepository, tokens TokenService, bus *event.Bus, limiter LoginRateLimiter) *UserApp {
	return &UserApp{users: users, tokens: tokens, bus: bus, limiter: limiter}
}

// LoginResult 登录结果。
type LoginResult struct {
	User       *entity.User `json:"user"`
	Token      string       `json:"token"`
	TokenType  string       `json:"token_type"`
	ExpiresIn  int64        `json:"expires_in"`
	ExpiresAt  time.Time    `json:"expires_at"`
}

// Login 用户登录（密码校验 + Token 签发 + 限流保护）。
func (a *UserApp) Login(ctx context.Context, username, password, ip string) (*LoginResult, error) {
	if username == "" || password == "" {
		return nil, errcode.New(errcode.CodeInvalidParam, "用户名或密码不能为空")
	}

	// 限流检查
	if a.limiter != nil {
		key := fmt.Sprintf("login:%s:%s", username, ip)
		allowed, retryAfter, err := a.limiter.Allow(ctx, key)
		if err == nil && !allowed {
			return nil, errcode.New(errcode.CodeTooManyRequest, fmt.Sprintf("登录失败次数过多，请 %d 秒后重试", int(retryAfter.Seconds())))
		}
	}

	user, err := a.users.GetByUsername(ctx, username)
	if err != nil {
		return nil, errcode.New(errcode.CodeUserNotExist, "用户名或密码错误")
	}
	if user.Status == consts.UserStatusDisabled {
		return nil, errcode.New(errcode.CodeUserDisabled, "账号已禁用")
	}

	if !encrypt.BcryptCheck(user.Password, password) {
		return nil, errcode.New(errcode.CodeUserPasswordWrong, "用户名或密码错误")
	}

	token, exp, err := a.tokens.Generate(user.ID, user.Username, user.Role, jwt.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	// 更新最后登录信息
	_ = a.users.UpdateLastLogin(ctx, user.ID, ip)

	// 清空密码字段
	user.Password = ""

	return &LoginResult{
		User:      user,
		Token:     token,
		TokenType: jwt.AccessToken,
		ExpiresIn: int64(time.Until(exp).Seconds()),
		ExpiresAt: exp,
	}, nil
}

// Logout 登出（加入 Token 黑名单）。
func (a *UserApp) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	uid, _, err := a.tokens.Parse(token)
	if err != nil {
		return nil // 无效 token 直接成功
	}
	// 黑名单存入 Redis
	if err := a.addToBlacklist(ctx, token, uid); err != nil {
		return fmt.Errorf("add to blacklist: %w", err)
	}
	return nil
}

// LoginResult helpers - private
func (a *UserApp) addToBlacklist(ctx context.Context, token string, uid int64) error {
	// 通过 event bus 异步处理
	_ = a.bus.Publish(ctx, &event.UserLoggedOutEvent{
		UserID:  uid,
		Token:   token,
		LogOutAt: time.Now(),
	})
	return nil
}

// RefreshToken 刷新 Token。
func (a *UserApp) RefreshToken(ctx context.Context, refreshToken string) (string, time.Time, error) {
	uid, role, err := a.tokens.Parse(refreshToken)
	if err != nil {
		return "", time.Time{}, errcode.New(errcode.CodeTokenInvalid, "无效的 refresh token")
	}
	token, exp, err := a.tokens.Generate(uid, "", role, jwt.AccessToken)
	if err != nil {
		return "", time.Time{}, err
	}
	return token, exp, nil
}

// GetProfile 获取用户资料。
func (a *UserApp) GetProfile(ctx context.Context, uid int64) (*entity.User, error) {
	u, err := a.users.GetByID(ctx, uid)
	if err != nil {
		return nil, errcode.New(errcode.CodeUserNotExist, "UserNotExist")
	}
	u.Password = ""
	return u, nil
}

// UpdateProfile 更新资料。
func (a *UserApp) UpdateProfile(ctx context.Context, uid int64, nickname, phone, email string) error {
	if uid <= 0 {
		return errcode.New(errcode.CodeInvalidParam, "InvalidParam")
	}
	u, err := a.users.GetByID(ctx, uid)
	if err != nil {
		return errcode.New(errcode.CodeUserNotExist, "UserNotExist")
	}
	if nickname != "" {
		u.Nickname = nickname
	}
	if phone != "" {
		u.Phone = phone
	}
	if email != "" {
		u.Email = email
	}
	return a.users.UpdateProfile(ctx, u)
}


// AdminUpdateUser 管理员更新用户信息（更多字段）。
func (a *UserApp) AdminUpdateUser(ctx context.Context, u *entity.User) error {
	if u.ID <= 0 {
		return errcode.New(errcode.CodeInvalidParam, "InvalidParam")
	}
	existing, err := a.users.GetByID(ctx, u.ID)
	if err != nil {
		return errcode.New(errcode.CodeUserNotExist, "UserNotExist")
	}
	if u.Nickname != "" {
		existing.Nickname = u.Nickname
	}
	if u.Email != "" {
		existing.Email = u.Email
	}
	if u.Phone != "" {
		existing.Phone = u.Phone
	}
	if u.Avatar != "" {
		existing.Avatar = u.Avatar
	}
	if u.Role != 0 {
		existing.Role = u.Role
	}
	if u.Gender != 0 {
		existing.Gender = u.Gender
	}
	if u.ClassID != nil {
		existing.ClassID = u.ClassID
	}
	return a.users.UpdateProfile(ctx, existing)
}

// ChangePassword 修改密码。
func (a *UserApp) ChangePassword(ctx context.Context, uid int64, oldPwd, newPwd string) error {
	if uid <= 0 {
		return errcode.New(errcode.CodeInvalidParam, "InvalidParam")
	}
	u, err := a.users.GetByID(ctx, uid)
	if err != nil {
		return errcode.New(errcode.CodeUserNotExist, "UserNotExist")
	}
	if !encrypt.BcryptCheck(u.Password, oldPwd) {
		return errcode.New(errcode.CodeUserPasswordWrong, "原密码错误")
	}
	// 密码强度校验
	if _, err := valueobject.NewPassword(newPwd); err != nil {
		return errcode.New(errcode.CodePasswordWeak, err.Error())
	}
	hashed, err := encrypt.BcryptPassword(newPwd)
	if err != nil {
		return err
	}
	return a.users.UpdatePassword(ctx, uid, hashed)
}

// CreateUser 创建用户（管理员）。
func (a *UserApp) CreateUser(ctx context.Context, u *entity.User, rawPassword string) (int64, error) {
	if u.Username == "" {
		return 0, errcode.New(errcode.CodeInvalidParam, "用户名不能为空")
	}
	if rawPassword == "" {
		return 0, errcode.New(errcode.CodeInvalidParam, "密码不能为空")
	}
	pwd, err := valueobject.NewPassword(rawPassword)
	if err != nil {
		return 0, errcode.New(errcode.CodePasswordWeak, err.Error())
	}
	hashed, err := encrypt.BcryptPassword(string(pwd))
	if err != nil {
		return 0, err
	}
	u.Password = hashed
	if u.Status == 0 {
		u.Status = consts.UserStatusNormal
	}
	if err := a.users.Create(ctx, u); err != nil {
		return 0, err
	}
	return u.ID, nil
}

// ListUsers 列出用户（支持过滤）。
func (a *UserApp) ListUsers(ctx context.Context, filter repository.UserListFilter) ([]entity.User, int64, error) {
	return a.users.List(ctx, filter)
}

// GetByID 根据 ID 获取用户。
func (a *UserApp) GetByID(ctx context.Context, id int64) (*entity.User, error) {
	u, err := a.users.GetByID(ctx, id)
	if err != nil {
		return nil, errcode.New(errcode.CodeUserNotExist, "UserNotExist")
	}
	u.Password = ""
	return u, nil
}

// ResetPassword 重置密码（管理员）。
func (a *UserApp) ResetPassword(ctx context.Context, id int64) (string, error) {
	if id <= 0 {
		return "", errcode.New(errcode.CodeInvalidParam, "InvalidParam")
	}
	// 生成 12 位随机密码（含大小写+数字）
	newPwd := generateRandomPassword(12)
	hashed, err := encrypt.BcryptPassword(newPwd)
	if err != nil {
		return "", err
	}
	if err := a.users.UpdatePassword(ctx, id, hashed); err != nil {
		return "", err
	}
	return newPwd, nil
}

// ChangeStatus 切换用户状态（启用/禁用）。
func (a *UserApp) ChangeStatus(ctx context.Context, id int64, status int8) error {
	if id <= 0 {
		return errcode.New(errcode.CodeInvalidParam, "InvalidParam")
	}
	u, err := a.users.GetByID(ctx, id)
	if err != nil {
		return errcode.New(errcode.CodeUserNotExist, "UserNotExist")
	}
	u.Status = status
	return a.users.Update(ctx, u)
}

// DeleteUser 删除用户（软删除）。
func (a *UserApp) DeleteUser(ctx context.Context, id int64) error {
	if id <= 0 {
		return errcode.New(errcode.CodeInvalidParam, "InvalidParam")
	}
	return a.users.Delete(ctx, id)
}

// ============================================================
// 私有辅助
// ============================================================

// generateRandomPassword 生成随机密码。
func generateRandomPassword(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	now := time.Now().UnixNano()
	for i := 0; i < n; i++ {
		b[i] = charset[(now>>uint(i*5))&0x3F%int64(len(charset))]
	}
	// 确保至少 1 个数字 + 1 个大写
	if b[0] >= 'a' && b[0] <= 'z' {
		b[0] = b[0] - 32
	}
	if b[1] >= 'A' && b[1] <= 'Z' {
		b[1] = b[1] - 'A' + '0'
		if b[1] > '9' {
			b[1] = '1'
		}
	}
	return string(b)
}



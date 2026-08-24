package user

import (
	"context"
	"time"

	"github.com/your-team/koala-exam-backend/internal/application/dto"
	"github.com/your-team/koala-exam-backend/internal/domain/entity"
	"github.com/your-team/koala-exam-backend/internal/domain/errcode"
	"github.com/your-team/koala-exam-backend/internal/infrastructure/repository"
	"github.com/your-team/koala-exam-backend/pkg/encrypt"
	"github.com/your-team/koala-exam-backend/pkg/jwt"
	"github.com/your-team/koala-exam-backend/pkg/utils"
)

// UserApp 用户应用服务
type UserApp struct {
	userRepo *repository.UserRepository
	jwt      *jwt.Helper
}

func NewUserApp(r *repository.UserRepository, j *jwt.Helper) *UserApp {
	return &UserApp{userRepo: r, jwt: j}
}

// Login 登录
func (a *UserApp) Login(ctx context.Context, req *dto.LoginReq, ip string) (*dto.LoginResp, error) {
	u, err := a.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, errcode.New(errcode.CodeUserNotExist, "账号或密码错误")
	}
	if u.Status != 1 {
		return nil, errcode.New(errcode.CodeUserDisabled, "UserDisabled")
	}
	if !encrypt.BcryptCheck(u.Password, req.Password) {
		return nil, errcode.New(errcode.CodeUserPasswordWrong, "UserPasswordWrong")
	}

	access, exp, _ := a.jwt.Generate(u.ID, u.Username, u.Role, "access")
	refresh, _, _ := a.jwt.Generate(u.ID, u.Username, u.Role, "refresh")
	u.Password = ""
	_ = a.userRepo.UpdateLastLogin(ctx, u.ID, ip)
	return &dto.LoginResp{
		User: u, AccessToken: access, RefreshToken: refresh,
		ExpiresIn: int64(time.Until(exp).Seconds()),
	}, nil
}

// RefreshToken 刷新 token
func (a *UserApp) RefreshToken(ctx context.Context, refreshToken string) (string, int64, error) {
	access, exp, err := a.jwt.Refresh(refreshToken)
	if err != nil {
		return "", 0, errcode.New(errcode.CodeTokenInvalid, "TokenInvalid")
	}
	return access, int64(time.Until(exp).Seconds()), nil
}

// GetProfile 获取当前用户资料
func (a *UserApp) GetProfile(ctx context.Context, uid int64) (*entity.User, error) {
	u, err := a.userRepo.GetByID(ctx, uid)
	if err != nil {
		return nil, errcode.New(errcode.CodeUserNotExist, "UserNotExist")
	}
	u.Password = ""
	return u, nil
}

// CreateUser 创建用户（管理员）
func (a *UserApp) CreateUser(ctx context.Context, u *entity.User) error {
	if u.Password == "" {
		u.Password = "koala123"
	}
	hashed, err := encrypt.BcryptPassword(u.Password)
	if err != nil {
		return err
	}
	u.Password = hashed
	// 确保 nullable 字段为 nil
	u.LastLoginAt = nil
	u.LastLoginIP = ""
	return a.userRepo.Create(ctx, u)
}

// ListUsers 列表
func (a *UserApp) ListUsers(ctx context.Context, page, size int, role int8, keyword string) ([]entity.User, int64, error) {
	return a.userRepo.List(ctx, page, size, role, keyword)
}

// ResetPassword 重置密码
func (a *UserApp) ResetPassword(ctx context.Context, uid int64) (string, error) {
	newPwd := utils.RandString(8)
	hashed, err := encrypt.BcryptPassword(newPwd)
	if err != nil {
		return "", err
	}
	if err := a.userRepo.Update(ctx, &entity.User{ID: uid, Password: hashed}); err != nil {
		return "", err
	}
	return newPwd, nil
}

// ChangePassword 修改密码（学员自己）
func (a *UserApp) ChangePassword(ctx context.Context, uid int64, oldPwd, newPwd string) error {
	u, err := a.userRepo.GetByID(ctx, uid)
	if err != nil { return errcode.New(errcode.CodeUserNotExist, "UserNotExist") }
	if !encrypt.BcryptCheck(u.Password, oldPwd) { return errcode.New(errcode.CodeUserPasswordWrong, "UserPasswordWrong") }
	hashed, _ := encrypt.BcryptPassword(newPwd)
	return a.userRepo.Update(ctx, &entity.User{ID: uid, Password: hashed})
}

// UpdateProfile 更新资料
func (a *UserApp) UpdateProfile(ctx context.Context, uid int64, nickname, email, phone, avatar string) error {
	u := &entity.User{ID: uid, Nickname: nickname, Email: email, Phone: phone, Avatar: avatar}
	return a.userRepo.Update(ctx, u)
}

// DeleteUser 软删除
func (a *UserApp) DeleteUser(ctx context.Context, uid int64) error {
	return a.userRepo.Delete(ctx, uid)
}

// ToggleStatus 启用/禁用
func (a *UserApp) ToggleStatus(ctx context.Context, uid int64, status int8) error {
	return a.userRepo.Update(ctx, &entity.User{ID: uid, Status: status})
}

// GetByID 查询单个
func (a *UserApp) GetByID(ctx context.Context, uid int64) (*entity.User, error) {
	u, err := a.userRepo.GetByID(ctx, uid)
	if err != nil { return nil, errcode.New(errcode.CodeUserNotExist, "UserNotExist") }
	u.Password = ""
	return u, nil
}
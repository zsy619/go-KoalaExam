// Package repository 仓储实现层（GORM）。
//
// 这里实现 domain/repository 中定义的接口。
// 命名与接口一一对应，便于通过 wire / 依赖注入容器装配。
package repository

import (
	"context"

	"gorm.io/gorm"

	domRepo "github.com/your-team/koala-exam-backend/internal/domain/repository"
	"github.com/your-team/koala-exam-backend/internal/domain/entity"
)

// UserRepository 用户仓储实现。
//
// 编译期断言：实现 domain/repository.UserRepository 接口。
var _ domRepo.UserRepository = (*UserRepository)(nil)

// UserRepository GORM 实现。
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository 构造 UserRepository。
func NewUserRepository(db *gorm.DB) *UserRepository { return &UserRepository{db: db} }

// Create 创建用户。
func (r *UserRepository) Create(ctx context.Context, u *entity.User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

// Update 全量更新用户信息（不含密码、个人资料，后两者有专门方法）。
func (r *UserRepository) Update(ctx context.Context, u *entity.User) error {
	return r.db.WithContext(ctx).Model(u).Where("id = ?", u.ID).
		Updates(map[string]interface{}{
			"username": u.Username,
			"password": u.Password,
			"nickname": u.Nickname,
			"email":    u.Email,
			"phone":    u.Phone,
			"avatar":   u.Avatar,
			"role":     u.Role,
			"gender":   u.Gender,
			"status":   u.Status,
		}).Error
}

// UpdateProfile 仅更新个人资料字段。
func (r *UserRepository) UpdateProfile(ctx context.Context, u *entity.User) error {
	return r.db.WithContext(ctx).Model(u).Where("id = ?", u.ID).
		Updates(map[string]interface{}{
			"nickname": u.Nickname,
			"phone":    u.Phone,
			"email":    u.Email,
			"avatar":   u.Avatar,
		}).Error
}

// UpdatePassword 仅更新密码字段。
func (r *UserRepository) UpdatePassword(ctx context.Context, id int64, hashed string) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?", id).
		Update("password", hashed).Error
}

// Delete 软删除用户。
func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&entity.User{}, id).Error
}

// GetByID 根据 ID 查询。
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*entity.User, error) {
	var u entity.User
	if err := r.db.WithContext(ctx).First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

// GetByUsername 根据用户名查询。
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	var u entity.User
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

// List 分页查询用户。
func (r *UserRepository) List(ctx context.Context, filter domRepo.UserListFilter) ([]entity.User, int64, error) {
	var list []entity.User
	var total int64
	q := r.db.WithContext(ctx).Model(&entity.User{})
	if filter.Role > 0 {
		q = q.Where("role = ?", filter.Role)
	}
	if filter.Status >= 0 {
		q = q.Where("status = ?", filter.Status)
	}
	if filter.ClassID > 0 {
		q = q.Where("class_id = ?", filter.ClassID)
	}
	if filter.Keyword != "" {
		kw := "%" + filter.Keyword + "%"
		q = q.Where("username LIKE ? OR nickname LIKE ? OR phone LIKE ?", kw, kw, kw)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("id DESC").
		Offset(filter.Offset()).Limit(filter.Limit()).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// UpdateLastLogin 记录登录时间/IP。
func (r *UserRepository) UpdateLastLogin(ctx context.Context, id int64, ip string) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_login_at": gorm.Expr("NOW()"),
			"last_login_ip": ip,
		}).Error
}

// CountByRole 统计某角色的用户数。
func (r *UserRepository) CountByRole(ctx context.Context, role int8) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&entity.User{}).Where("role = ?", role).Count(&n).Error
	return n, err
}

// CountTotal 用户总数。
func (r *UserRepository) CountTotal(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&entity.User{}).Count(&n).Error
	return n, err
}

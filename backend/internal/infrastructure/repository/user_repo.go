package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/your-team/koala-exam-backend/internal/domain/entity"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository { return &UserRepository{db: db} }

func (r *UserRepository) Create(ctx context.Context, u *entity.User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *UserRepository) Update(ctx context.Context, u *entity.User) error {
	// 使用 Updates 避免零值时间字段
	return r.db.WithContext(ctx).Model(u).Where("id = ?", u.ID).Updates(map[string]interface{}{
		"username": u.Username, "password": u.Password, "nickname": u.Nickname,
		"email": u.Email, "phone": u.Phone, "avatar": u.Avatar,
		"role": u.Role, "gender": u.Gender, "status": u.Status,
	}).Error
}

func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&entity.User{}, id).Error
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*entity.User, error) {
	var u entity.User
	if err := r.db.WithContext(ctx).First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	var u entity.User
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) List(ctx context.Context, page, size int, role int8, keyword string) ([]entity.User, int64, error) {
	var list []entity.User
	var total int64
	q := r.db.WithContext(ctx).Model(&entity.User{})
	if role > 0 {
		q = q.Where("role = ?", role)
	}
	if keyword != "" {
		q = q.Where("username LIKE ? OR nickname LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Offset((page - 1) * size).Limit(size).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id int64, ip string) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?", id).
		Updates(map[string]interface{}{"last_login_at": gorm.Expr("NOW()"), "last_login_ip": ip}).Error
}

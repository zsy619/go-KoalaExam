// Package repository 定义领域仓储接口
// 遵循 DDD 模式：repository 接口属于 domain 层，实现在 infrastructure 层
package repository

import (
	"context"

	"github.com/your-team/koala-exam-backend/internal/domain/entity"
)

// PageQuery 通用分页参数
type PageQuery struct {
	Page    int    // 页码（从 1 开始）
	Size    int    // 每页大小（默认 10，最大 100）
	Keyword string // 关键字（按 username/nickname 模糊搜索）
}

// Offset 计算偏移量
func (q PageQuery) Offset() int {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.Size <= 0 {
		q.Size = 10
	}
	if q.Size > 100 {
		q.Size = 100
	}
	return (q.Page - 1) * q.Size
}

// Limit 计算限制数
func (q PageQuery) Limit() int {
	if q.Size <= 0 {
		return 10
	}
	if q.Size > 100 {
		return 100
	}
	return q.Size
}

// UserListFilter 用户列表过滤条件
type UserListFilter struct {
	PageQuery
	Role    int8 // 1=超管 2=教师 3=学生；0=全部
	Status  int8 // 0=禁用 1=正常；-1=全部
	ClassID int64
}

// UserRepository 用户仓储接口
// 领域层只依赖接口，具体实现在 infrastructure 层
type UserRepository interface {
	Create(ctx context.Context, u *entity.User) error
	Update(ctx context.Context, u *entity.User) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*entity.User, error)
	GetByUsername(ctx context.Context, username string) (*entity.User, error)

	// List 返回分页用户列表
	List(ctx context.Context, filter UserListFilter) ([]entity.User, int64, error)

	// UpdatePassword 仅更新密码字段（避免其他字段被覆盖）
	UpdatePassword(ctx context.Context, id int64, hashed string) error

	// UpdateProfile 更新个人资料（昵称/邮箱/手机/头像）
	UpdateProfile(ctx context.Context, u *entity.User) error

	// UpdateLastLogin 记录登录时间/IP
	UpdateLastLogin(ctx context.Context, id int64, ip string) error

	// CountByRole 统计某角色的用户数
	CountByRole(ctx context.Context, role int8) (int64, error)

	// CountTotal 统计总用户数
	CountTotal(ctx context.Context) (int64, error)
}

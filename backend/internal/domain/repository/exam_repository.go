package repository

import (
	"context"
)

// ExamListFilter 考试过滤
type ExamListFilter struct {
	PageQuery
	Status    int8 // 0=草稿 1=进行中 2=已结束
	CreatorID int64
}

// ExamRepository 考试仓储接口
type ExamRepository interface {
	Create(ctx context.Context, e any) error
	Update(ctx context.Context, e any) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (any, error)
	List(ctx context.Context, filter ExamListFilter) (any, int64, error)
	ListAvailableForUser(ctx context.Context, userID int64) (any, error)
	CountTotal(ctx context.Context) (int64, error)
}

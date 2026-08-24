package repository

import (
	"context"

	"github.com/your-team/koala-exam-backend/internal/domain/entity"
)

// PaperListFilter 试卷过滤
type PaperListFilter struct {
	PageQuery
	Strategy int8 // 1=固定 2=随机 3=GA；0=全部
	Status   int8 // 0=草稿 1=发布 2=归档；-1=全部
	CreatorID int64
}

// PaperRepository 试卷仓储接口
type IPaperRepository interface {
	Create(ctx context.Context, p *entity.Paper) error
	Update(ctx context.Context, p *entity.Paper) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*entity.Paper, error)
	List(ctx context.Context, filter PaperListFilter) ([]entity.Paper, int64, error)

	// BindQuestions 绑定题目（覆盖式）
	BindQuestions(ctx context.Context, paperID int64, qids []int64) error

	// GetQuestions 获取试卷的题目
	GetQuestions(ctx context.Context, paperID int64) ([]entity.PaperQuestion, error)

	// CountTotal 试卷总数
	CountTotal(ctx context.Context) (int64, error)
}

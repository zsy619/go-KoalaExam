package repository

import (
	"context"

	"github.com/your-team/koala-exam-backend/internal/domain/entity"
)

// QuestionListFilter 题目列表过滤
type QuestionListFilter struct {
	PageQuery
	CategoryID int64  // 分类 ID；0=全部
	Type       int8   // 题型 1-6；0=全部
	Difficulty int8   // 难度 1-5；0=全部
	Status     int8   // 0=禁用 1=正常 2=审核中；-1=全部
	Tag        string // 按 tag 精确匹配
}

// QuestionRepository 题目仓储接口
type QuestionRepository interface {
	Create(ctx context.Context, q *entity.Question) error
	Update(ctx context.Context, q *entity.Question) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*entity.Question, error)
	List(ctx context.Context, filter QuestionListFilter) ([]entity.Question, int64, error)

	// BatchImport 批量导入
	BatchImport(ctx context.Context, qs []entity.Question) error

	// ListByIDs 批量查询（用于试卷组装）
	ListByIDs(ctx context.Context, ids []int64) ([]entity.Question, error)

	// RandomByTypeAndDifficulty 随机抽题（用于随机组卷）
	RandomByTypeAndDifficulty(ctx context.Context, qType int8, difficulty int8, n int) ([]entity.Question, error)

	// CountByCategory 统计某分类下的题目数
	CountByCategory(ctx context.Context, categoryID int64) (int64, error)

	// CountTotal 题目总数
	CountTotal(ctx context.Context) (int64, error)

	// IncStat 累加题目统计（use_count/correct_count/wrong_count）
	IncStat(ctx context.Context, id int64, field string, delta int) error
}

// CategoryRepository 题库分类仓储接口
type CategoryRepository interface {
	Create(ctx context.Context, c *entity.QuestionCategory) error
	Update(ctx context.Context, c *entity.QuestionCategory) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*entity.QuestionCategory, error)
	List(ctx context.Context) ([]entity.QuestionCategory, error)
}
